package whatsthis

import (
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// Separator predicates for tokenization.
// isTokenSep splits on basic word separators (dots, spaces, underscores).
func isTokenSep(r rune) bool {
	return r == '.' || r == ' ' || r == '_'
}

// isMetaSep splits on all metadata-region separators including brackets and hyphens.
func isMetaSep(r rune) bool {
	return r == '.' || r == ' ' || r == '_' || r == '[' || r == ']' || r == '(' || r == ')' || r == '-' || r == '+' || r == '&'
}

// parse extracts metadata from a video filename.
func parse(input string) Info {
	var g Info
	g.Type = Episode

	// Step 1: Extract container and handle double extensions.
	name, container, doubleExtBracket := extractContainer(input)
	g.Container = container
	g.MimeType = containerMIME[container]

	// Step 2: Extract season and episode — primary anchor.
	season, episode, seStart, seEnd := extractSeasonEpisode(name)
	if seStart < 0 {
		season, seStart, seEnd = findSeasonMarker(name)
	}
	g.Season = season
	g.Episode = episode

	// Step 3: Split into title region (before SE) and metadata region (after SE).
	var titleRegion, metaRegion string
	if seStart >= 0 {
		g.Type = Episode
		titleRegion = name[:seStart]
		metaRegion = name[seEnd:]
	} else {
		g.Type = Movie
		titleRegion, metaRegion = splitMovieRegions(name)
	}

	// Step 4: Extract type-specific fields.
	if g.Type == Episode {
		titleRegion = reEpisodeSourceSuffix.ReplaceAllString(titleRegion, "")
		// Handle " - Episode Title" and pure episode title patterns.
		metaRegion = stripEpisodeTitle(metaRegion)
		g.Languages, g.Multilingual = extractLanguages(leadingBracketTags(name) + " " + metaRegion)

		g.ReleaseGroup, metaRegion = extractReleaseGroup(metaRegion, doubleExtBracket)
		g.ReleaseGroup = formatReleaseGroup(g.ReleaseGroup)
		g.Year, titleRegion = extractTitleYear(titleRegion)

		// If no year was found in the title region, allow year immediately after SxxExx.
		if g.Year == 0 {
			g.Year, metaRegion = extractYear(metaRegion)
		}

		g.ScreenSize = extractScreenSize(metaRegion)
		g.VideoCodec = extractVideoCodec(metaRegion)
		g.AudioCodec = extractAudioCodec(metaRegion)
		g.Title = extractTitle(titleRegion, seStart, seEnd, name)
	} else {
		prefixGroup, cleanedTitleRegion := extractMoviePrefixGroup(titleRegion)
		titleRegion = cleanedTitleRegion

		g.Year, metaRegion = extractYear(metaRegion)
		g.Languages, g.Multilingual = extractLanguages(leadingBracketTags(name) + " " + metaRegion)
		g.ReleaseGroup, metaRegion = extractReleaseGroup(metaRegion, doubleExtBracket)
		g.ReleaseGroup = formatReleaseGroup(g.ReleaseGroup)
		if prefixGroup != "" {
			g.ReleaseGroup = prefixGroup
		}

		g.ScreenSize = extractScreenSize(metaRegion)
		g.VideoCodec = extractVideoCodec(metaRegion)
		g.AudioCodec = extractAudioCodec(metaRegion)
		g.Title = cleanMovieTitle(titleRegion)
	}

	return g
}

// parseSeasonName parses an extensionless whole-season release name.
func parseSeasonName(name string) (Info, bool) {
	season, markerStart, markerEnd, ok := extractSeasonName(name)
	typ := ShowSeason
	seasons, seriesStart, seriesEnd := extractSeriesName(name)
	if len(seasons) > 1 {
		season = 0
		markerStart, markerEnd = seriesStart, seriesEnd
		typ = Show
		ok = true
	} else {
		seasons = nil
	}
	if !ok {
		return Info{}, false
	}
	// A season followed by an episode is not a pack. Keep the existing
	// explicit episode-range and multi-season collection policies.
	wholeEpisodeRange := false
	for _, m := range reSeasonEpisodeRange.FindAllStringSubmatch(strings.ReplaceAll(name, "_", " "), -1) {
		first, _ := strconv.Atoi(m[2])
		last, _ := strconv.Atoi(m[3])
		wholeEpisodeRange = wholeEpisodeRange || first == 1 && last > first
	}
	if typ != Show && !wholeEpisodeRange {
		if _, _, start, _ := extractSeasonEpisode(name); start >= 0 {
			return Info{}, false
		}
	}

	g := Info{
		Type:    typ,
		Season:  season,
		Seasons: seasons,
	}

	titleRegion := reComplete.ReplaceAllString(name[:markerStart], " ")
	metaRegion := name[markerEnd:]
	metaRegion = stripSeasonNameMarkers(metaRegion)
	metaRegion = reComplete.ReplaceAllString(metaRegion, " ")
	metaRegion = reFileSizeBracket.ReplaceAllString(metaRegion, " ")
	metaRegion = reContainerBracket.ReplaceAllString(metaRegion, " ")

	explicitGroup, metaRegion := normalizeSeasonMeta(metaRegion)

	g.Year, titleRegion = extractTitleYear(titleRegion)
	if g.Year == 0 {
		g.Year, metaRegion = extractYear(metaRegion)
	}
	g.Languages, g.Multilingual = extractLanguages(leadingBracketTags(name) + " " + metaRegion)

	if explicitGroup != "" {
		g.ReleaseGroup = explicitGroup
	} else {
		g.ReleaseGroup, metaRegion = extractSeasonReleaseGroup(metaRegion)
	}
	g.ReleaseGroup = formatReleaseGroup(g.ReleaseGroup)
	g.ScreenSize = extractScreenSize(metaRegion)
	g.VideoCodec = extractVideoCodec(metaRegion)
	g.AudioCodec = extractAudioCodec(metaRegion)
	g.Title = extractTitle(titleRegion, markerStart, markerEnd, name)

	return g, true
}

type seasonNameMarker struct {
	season int
	start  int
	end    int
}

// extractSeriesName collects ascending explicit ranges, merging repeated and
// overlapping ranges without inventing seasons between disjoint ranges.
func extractSeriesName(name string) (seasons []int, start, end int) {
	name = strings.ReplaceAll(name, "_", " ")
	start, end = -1, -1
	seen := make(map[int]bool)
	for _, re := range []*regexp.Regexp{reSeasonRange, reSeasonWordRange} {
		for _, match := range re.FindAllStringSubmatchIndex(name, -1) {
			firstSeason, _ := strconv.Atoi(name[match[2]:match[3]])
			lastSeason, _ := strconv.Atoi(name[match[4]:match[5]])
			if lastSeason < firstSeason || hasPrecedingMetadata(name[:match[0]]) {
				continue
			}
			if start < 0 || match[0] < start {
				start, end = match[0], match[1]
			}
			for season := firstSeason; season <= lastSeason; season++ {
				if !seen[season] {
					seasons = append(seasons, season)
					seen[season] = true
				}
			}
		}
	}
	slices.Sort(seasons)
	return seasons, start, end
}

// extractSeasonName finds markers that identify a whole-season release. When
// multiple markers are present, they must all name the same season.
func extractSeasonName(name string) (season, start, end int, ok bool) {
	name = strings.ReplaceAll(name, "_", " ")
	var markers []seasonNameMarker

	for _, m := range reSeasonEpisodeRange.FindAllStringSubmatchIndex(name, -1) {
		seasonNumber, _ := strconv.Atoi(name[m[2]:m[3]])
		firstEpisode, _ := strconv.Atoi(name[m[4]:m[5]])
		lastEpisode, _ := strconv.Atoi(name[m[6]:m[7]])
		if firstEpisode == 1 && lastEpisode > firstEpisode {
			markers = append(markers, seasonNameMarker{seasonNumber, m[0], m[1]})
		}
	}

	for _, re := range []*regexp.Regexp{reSeasonOnly, reSeasonWordOnly, reSeasonOrdinal} {
		for _, m := range re.FindAllStringSubmatchIndex(name, -1) {
			seasonNumber, _ := strconv.Atoi(name[m[2]:m[3]])
			markers = append(markers, seasonNameMarker{seasonNumber, m[0], m[1]})
		}
	}

	if len(markers) == 0 {
		return 0, 0, 0, false
	}

	first := markers[0]
	for _, marker := range markers[1:] {
		if marker.season != first.season {
			return 0, 0, 0, false
		}
		if marker.start < first.start {
			first = marker
		}
	}

	// A trailing group such as -S56 must not turn a movie into a season.
	prefix := name[:first.start]
	if hasPrecedingMetadata(prefix) {
		return 0, 0, 0, false
	}
	return first.season, first.start, first.end, true
}

func stripSeasonNameMarkers(meta string) string {
	for _, re := range []*regexp.Regexp{reSeasonDescriptorBracket, reSeasonEpisodeRange, reSeasonRange, reSeasonWordRange, reSeasonOnly, reSeasonOrdinal, reSeasonWordOnly} {
		// Normalize only for matching: preserve underscores in release groups.
		matches := re.FindAllStringIndex(strings.ReplaceAll(meta, "_", " "), -1)
		for i := len(matches) - 1; i >= 0; i-- {
			m := matches[i]
			meta = meta[:m[0]] + " " + meta[m[1]:]
		}
	}
	return meta
}

// normalizeSeasonMeta removes punctuation around a season marker. A leading
// hyphen with no metadata after it denotes an explicit release group.
func normalizeSeasonMeta(meta string) (explicitGroup, remaining string) {
	trimmed := strings.TrimSpace(meta)
	if strings.HasPrefix(trimmed, "-") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		if !hasMetadataTokens(strings.NewReplacer("[", " ", "]", " ").Replace(trimmed)) &&
			!isSeasonMetadataBracket(strings.Trim(trimmed, "[]")) {
			return cleanReleaseGroup(trimmed), ""
		}
	}

	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}

	return "", trimmed
}

// isSeasonMetadataBracket reports whether bracket content contains metadata
// that should remain available to the screen-size and codec extractors.
func isSeasonMetadataBracket(content string) bool {
	return hasMetadataTokens(content) || extractAudioCodec(content) != ""
}

// extractSeasonReleaseGroup treats a final bracket as the release group unless
// it contains known metadata (e.g., [1080p] or [x265]).
func extractSeasonReleaseGroup(meta string) (group, remaining string) {
	trimmed := strings.TrimSpace(meta)
	if strings.HasSuffix(trimmed, "]") {
		if start := strings.LastIndex(trimmed, "["); start >= 0 {
			group = strings.TrimSpace(trimmed[start+1 : len(trimmed)-1])
			if isSeasonMetadataBracket(group) {
				priorGroup, before := extractSeasonReleaseGroup(trimmed[:start])
				return priorGroup, strings.TrimSpace(before + " " + trimmed[start:])
			}
			if group != "" {
				return group, strings.TrimSpace(trimmed[:start])
			}
		}
	}
	return extractReleaseGroup(meta, "")
}

// extractContainer extracts the file extension/container from the filename.
// Returns the name without extension, the container, and any bracket content
// from a double-extension pattern (e.g., "file.mkv[novarelay].mkv").
func extractContainer(input string) (name, container, doubleExtBracket string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(input), "."))
	if _, ok := containerMIME[ext]; !ok {
		return input, "", ""
	}

	name = strings.TrimSuffix(input, filepath.Ext(input))

	// Check for double extension: "file.mkv[tag].mkv"
	if m := reDoubleExt.FindStringSubmatch(input); m != nil {
		innerExt := strings.ToLower(m[1])
		if _, ok := containerMIME[innerExt]; ok {
			doubleExtBracket = m[2]
			// Strip the inner ".ext[bracket]" from name.
			// name is currently "file.mkv[tag]", remove the "[tag]" and ".mkv".
			bracketStart := strings.LastIndex(name, "[")
			if bracketStart >= 0 {
				name = name[:bracketStart]
				innerExtWithDot := filepath.Ext(name)
				if strings.ToLower(strings.TrimPrefix(innerExtWithDot, ".")) == innerExt {
					name = strings.TrimSuffix(name, innerExtWithDot)
				}
			}
		}
	}

	return name, ext, doubleExtBracket
}

// extractSeasonEpisode extracts season and episode numbers.
// Returns season, episode, start index, and end index of the match in name.
func extractSeasonEpisode(name string) (season, episode, start, end int) {
	// Replacing ASCII separators preserves byte offsets into the original name.
	name = strings.ReplaceAll(name, "_", " ")
	// Try SxxExx pattern first (most common).
	if m := reSeasonEpisode.FindStringSubmatchIndex(name); m != nil {
		s, _ := strconv.Atoi(name[m[2]:m[3]])
		e, _ := strconv.Atoi(name[m[4]:m[5]])
		// The marker starts one byte before the season digits, at S,
		// excluding the boundary character consumed by the regexp.
		return s, e, m[2] - 1, m[5]
	}

	for _, re := range []*regexp.Regexp{reCrossEpisode, reWordSeasonEpisode, reLooseSeasonEpisode} {
		for _, m := range re.FindAllStringSubmatchIndex(name, -1) {
			// Do not read a suffix of pixel dimensions as season/episode.
			if hasPrecedingMetadata(name[:m[0]]) ||
				m[0] > 0 && name[m[0]-1] >= '0' && name[m[0]-1] <= '9' ||
				m[5] < len(name) && name[m[5]] >= '0' && name[m[5]] <= '9' {
				continue
			}
			// Without an episode label, a resolution after S01 is metadata.
			if re == reLooseSeasonEpisode && reScreenSizeNumberToken.MatchString(name[m[4]:m[5]]) {
				continue
			}
			s, _ := strconv.Atoi(name[m[2]:m[3]])
			e, _ := strconv.Atoi(name[m[4]:m[5]])
			return s, e, m[0], m[5]
		}
	}

	season, seasonStart, seasonEnd := findSeasonMarker(name)
	if seasonStart >= 0 {
		if m := reEpisodeAfterSeason.FindStringSubmatchIndex(name[seasonEnd:]); m != nil {
			number := name[seasonEnd+m[2] : seasonEnd+m[3]]
			if !reScreenSizeNumberToken.MatchString(number) {
				e, _ := strconv.Atoi(number)
				return season, e, seasonStart, seasonEnd + m[3]
			}
		}
	}
	for _, m := range reEpisodeOnly.FindAllStringSubmatchIndex(name, -1) {
		// In a year-labelled movie, "Episode 1" can be part of the title.
		if !hasPrecedingMetadata(name[:m[0]]) && (seasonStart >= 0 || reYear.FindStringIndex(name[m[3]:]) == nil) {
			e, _ := strconv.Atoi(name[m[2]:m[3]])
			start := m[0]
			if seasonStart >= 0 && seasonStart < start {
				start = seasonStart
			}
			return season, e, start, m[3]
		}
	}

	var absolute []int
	for _, m := range reAbsoluteEpisode.FindAllStringSubmatchIndex(name, -1) {
		e, _ := strconv.Atoi(name[m[2]:m[3]])
		if e >= 1900 && e <= 2099 || hasPrecedingMetadata(name[:m[0]]) ||
			seasonStart >= 0 && reScreenSizeNumberToken.MatchString(name[m[2]:m[3]]) {
			continue
		}
		if strings.HasPrefix(name, "[") || reYear.FindStringIndex(name[m[3]:]) == nil {
			if absolute == nil {
				absolute = m
			}
			// Prefer a number immediately followed by a metadata block over a
			// number inside the title (Fairy Tail - 100 Years Quest - 05).
			tail := strings.TrimLeft(name[m[3]:], " ._-")
			if len(tail) > 0 && (tail[0] == '(' || tail[0] == '[') {
				absolute = m
				break
			}
		}
	}
	if absolute != nil {
		e, _ := strconv.Atoi(name[absolute[2]:absolute[3]])
		start := absolute[0]
		if seasonStart >= 0 && seasonStart < start {
			start = seasonStart
		}
		return season, e, start, absolute[3]
	}
	if seasonStart < 0 {
		if m := reLeadingEpisode.FindStringSubmatchIndex(name); m != nil && reYear.FindStringIndex(name[m[3]:]) == nil {
			number := name[m[2]:m[3]]
			if number[0] == '0' || len(number) == 3 && strings.Contains(name[m[3]:m[1]], "-") {
				e, _ := strconv.Atoi(number)
				return 0, e, m[2], m[3]
			}
		}
	}
	// A leading fansub tag makes bare numbering more credible than a compact
	// season/episode split (One Piece 1098 is episode 1098, not S10E98).
	if strings.HasPrefix(name, "[") {
		for _, m := range reAnimeEpisode.FindAllStringSubmatchIndex(name, -1) {
			if hasPrecedingMetadata(name[:m[2]]) || seasonStart >= 0 && reScreenSizeNumberToken.MatchString(name[m[2]:m[3]]) {
				continue
			}
			e, _ := strconv.Atoi(name[m[2]:m[3]])
			if e < 1900 || e > 2099 {
				return season, e, m[2], m[3]
			}
		}
	}

	// Try compact episode pattern (e.g., 217, 1316).
	if m := reCompactEpisode.FindStringSubmatchIndex(name); m != nil {
		numStr := name[m[2]:m[3]]
		n, _ := strconv.Atoi(numStr)
		if hasPrecedingMetadata(name[:m[2]]) || seasonStart >= 0 && reScreenSizeNumberToken.MatchString(numStr) {
			return 0, 0, -1, -1
		}

		// Skip if it looks like a year.
		if n >= 1900 && n <= 2099 {
			return 0, 0, -1, -1
		}

		// If a year appears later in the filename, this is likely a movie.
		if m[3] < len(name) {
			if reYear.FindStringSubmatchIndex(name[m[3]:]) != nil {
				return 0, 0, -1, -1
			}
		}

		ep := n % 100
		se := n / 100
		if se > 0 && ep > 0 {
			return se, ep, m[2], m[3]
		}
	}

	return 0, 0, -1, -1
}

// findSeasonMarker locates the earliest standalone season annotation.
func findSeasonMarker(name string) (season, start, end int) {
	start, end = -1, -1
	name = strings.ReplaceAll(name, "_", " ")
	for _, re := range []*regexp.Regexp{reSeasonOnly, reSeasonWordOnly, reSeasonOrdinal} {
		if m := re.FindStringSubmatchIndex(name); m != nil && (start < 0 || m[0] < start) {
			if hasPrecedingMetadata(name[:m[0]]) {
				continue
			}
			season, _ = strconv.Atoi(name[m[2]:m[3]])
			start, end = m[0], m[1]
		}
	}
	return season, start, end
}

// stripEpisodeTitle handles episode title patterns after SxxExx.
//
// Pattern 1: " - Episode Title" (e.g., "Show-Name SxxExx - Abandon All Hope.mkv").
// Pattern 2: Space-separated episode title with no metadata tokens
// (e.g., "Show SxxExx Where the Road Goes.mkv").
//
// Returns the metadata region with episode title text removed.
func stripEpisodeTitle(meta string) string {
	// Pattern 1: " - Episode Title" separator.
	if reEpisodeTitleSep.MatchString(meta) {
		afterSep := reEpisodeTitleSep.ReplaceAllString(meta, "")

		// If there's parenthetical metadata, extract that.
		// e.g., " - Where I Come From (1080p WEB-DL x265 SAMPA)"
		if parenIdx := strings.Index(afterSep, "("); parenIdx >= 0 {
			parenContent := afterSep[parenIdx+1:]
			if closeIdx := strings.LastIndex(parenContent, ")"); closeIdx >= 0 {
				parenContent = parenContent[:closeIdx]
			}
			if hasMetadataTokens(parenContent) {
				return " " + parenContent
			}
		}

		// Pure episode title (no metadata tokens) -> strip entirely.
		if !hasMetadataTokens(afterSep) {
			return ""
		}

		return meta
	}

	// Pattern 2: Space-separated episode title with no metadata tokens.
	// e.g., " Where the Road Goes" or " Socalyalcon VI"
	// Only applies when there are NO dots in the meta region (space-separated filenames).
	trimmed := strings.TrimSpace(meta)
	if trimmed == "" {
		return meta
	}

	// Check for hashtag directly after SxxExx (e.g., "S03E02-#slipperyslope").
	if strings.HasPrefix(trimmed, "-#") || strings.HasPrefix(trimmed, "#") {
		return ""
	}

	// Only trigger for space-separated content without dots.
	if !strings.Contains(trimmed, ".") && !hasMetadataTokens(meta) {
		return ""
	}

	return meta
}

// hasMetadataTokens checks if a string contains any known metadata tokens.
func hasMetadataTokens(s string) bool {
	if extractVideoCodec(s) != "" || extractAudioCodec(s) != "" {
		return true
	}
	lower := strings.ToLower(s)
	tokens := strings.FieldsFunc(lower, isMetaSep)
	for _, tok := range tokens {
		if knownMetadataTokens[tok] {
			return true
		}
		if releaseLanguageName(tok) != "" || reScreenSizeToken.MatchString(tok) || reScreenSizeNumberToken.MatchString(tok) {
			return true
		}
		if _, ok := screenSizeDimensions[tok]; ok {
			return true
		}
	}
	return false
}

func hasTechnicalMetadata(s string) bool {
	return extractScreenSize(s) != "" || extractVideoCodec(s) != "" || extractAudioCodec(s) != ""
}

// Weak numbering after technical tags is usually a size, channel count, or
// release-group suffix. Leading bracket tags do not establish that boundary.
func hasPrecedingMetadata(s string) bool {
	s = strings.TrimSpace(s)
	for strings.HasPrefix(s, "[") {
		end := strings.IndexByte(s, ']')
		if end < 0 {
			break
		}
		s = strings.TrimLeft(s[end+1:], " ._-")
	}
	return hasTechnicalMetadata(s)
}

// extractReleaseGroup extracts the release group from the metadata region.
func extractReleaseGroup(meta, doubleExtBracket string) (group, remaining string) {
	if meta == "" && doubleExtBracket == "" {
		return "", meta
	}

	// If we have a double-extension bracket (e.g., from "file.mkv[novarelay].mkv"),
	// that bracket IS the release group.
	if doubleExtBracket != "" {
		return strings.TrimSpace(strings.Trim(doubleExtBracket, "[]")), meta
	}

	if meta == "" {
		return "", meta
	}

	// Strategy 1: Look for bracket release group [GROUP].
	if idx := strings.LastIndex(meta, "["); idx >= 0 {
		endIdx := strings.Index(meta[idx:], "]")
		if endIdx >= 0 {
			bracketContent := strings.TrimSpace(meta[idx+1 : idx+endIdx])
			hadDotBeforeBracket := idx > 0 && meta[idx-1] == '.'
			beforeBracket := strings.TrimRight(meta[:idx], ". ")

			if groupSep := strings.LastIndex(bracketContent, "-"); groupSep > 0 {
				bracketMeta := strings.TrimSpace(bracketContent[:groupSep])
				bracketGroup := strings.TrimSpace(bracketContent[groupSep+1:])
				if bracketMeta != "" && bracketGroup != "" && hasMetadataTokens(bracketMeta) && !hasMetadataTokens(bracketGroup) {
					remaining := strings.TrimSpace(beforeBracket + " " + bracketMeta)
					return cleanReleaseGroup(bracketGroup), remaining
				}
			}
			if hasMetadataTokens(bracketContent) {
				// Metadata brackets can precede a conventional trailing group.
				// Parse that suffix without losing the bracket's metadata or order.
				if tail := meta[idx+endIdx+1:]; strings.TrimSpace(tail) != "" {
					if suffixGroup, suffixMeta := extractReleaseGroup(tail, ""); suffixGroup != "" && !strings.ContainsAny(suffixGroup, " ()") {
						return suffixGroup, meta[:idx+endIdx+1] + suffixMeta
					}
				}
				priorGroup, before := extractReleaseGroup(beforeBracket, "")
				return priorGroup, strings.TrimSpace(before + " " + meta[idx:])
			}

			if strings.HasSuffix(beforeBracket, "-") {
				return strings.Trim(bracketContent, "[]"), strings.TrimRight(strings.TrimSuffix(beforeBracket, "-"), ". ")
			}

			// Check if the token immediately before the bracket is an unknown token.
			// If so, include it: "Will1869[novarelay.re]" -> "Will1869[novarelay.re]"
			// But NOT if it's a known token: "x264[NOVARELAYx.to]" -> "NOVARELAYx.to"
			lastTok := lastDotSeparatedToken(beforeBracket)
			lastTokLower := strings.ToLower(lastTok)
			includeLastToken := lastTok != "" && !isKnownToken(lastTokLower)

			// Check if there's a hyphen-separated release group before the bracket.
			if hyphenIdx := strings.LastIndex(beforeBracket, "-"); hyphenIdx >= 0 {
				afterHyphen := strings.TrimSpace(beforeBracket[hyphenIdx+1:])
				beforeHyphen := beforeBracket[:hyphenIdx]

				if afterHyphen != "" && !isCompoundToken(beforeHyphen, afterHyphen) && !hasTechnicalMetadata(afterHyphen) {
					// The part after the hyphen + bracket is the release group.
					// e.g., "-RARBG[novarelay.re]" -> "RARBG[novarelay.re]"
					group = strings.TrimSpace(strings.TrimRight(afterHyphen, ". ") + "[" + bracketContent + "]")
					return group, strings.TrimRight(beforeHyphen, ". ")
				}
			}

			// If bracket looks like a domain/source tag, prefer bracket only.
			if isLikelyDomainTag(bracketContent) && (!containsDigit(lastTok) || isCodecNumberToken(lastTokLower)) {
				return strings.Trim(bracketContent, "[]"), beforeBracket
			}

			if isCodecNumberToken(lastTokLower) {
				return strings.Trim(bracketContent, "[]"), beforeBracket
			}

			if includeLastToken {
				// Include the unknown token before the bracket.
				// "noobless[UTR]" -> "noobless[UTR]"
				tokenStart := strings.LastIndexAny(beforeBracket, ". ") + 1
				tokenPart := beforeBracket[tokenStart:]
				beforeToken := strings.TrimRight(beforeBracket[:tokenStart], ". ")
				sep := "["
				if hadDotBeforeBracket {
					sep = ".["
				}
				group = strings.TrimSpace(tokenPart + sep + bracketContent + "]")
				return group, beforeToken
			}

			// Just the bracket content.
			return strings.Trim(bracketContent, "[]"), beforeBracket
		}
	}

	// Strategy 2: Hyphen-separated release group.
	if hyphenIdx := strings.LastIndex(meta, "-"); hyphenIdx >= 0 {
		afterHyphen := strings.TrimRight(meta[hyphenIdx+1:], ". ")
		beforeHyphen := meta[:hyphenIdx]

		if !isCompoundToken(beforeHyphen, afterHyphen) && afterHyphen != "" && !hasTechnicalMetadata(afterHyphen) {
			// Handle short numeric tail groups like "GCJM-0".
			if len(afterHyphen) <= 2 {
				prevHyphen := strings.LastIndex(beforeHyphen, "-")
				if prevHyphen >= 0 {
					prefix := strings.TrimRight(beforeHyphen[prevHyphen+1:], ". ")
					if prefix != "" {
						group = cleanReleaseGroup(prefix + "-" + afterHyphen)
						return group, strings.TrimRight(beforeHyphen[:prevHyphen], ". ")
					}
				}
			}

			// Check if afterHyphen is empty or just a bracket.
			if strings.HasPrefix(afterHyphen, "[") {
				// "-[novarelay]" pattern.
				content := strings.Trim(afterHyphen, "[]")
				return content, strings.TrimRight(beforeHyphen, ". ")
			}
			group = cleanReleaseGroup(afterHyphen)
			return group, strings.TrimRight(beforeHyphen, ". ")
		}
	}

	// Strategy 3: Last unknown dot-separated token.
	// Only for dot-separated metadata (NOT space-only episode titles).
	if strings.Contains(meta, ".") {
		group, remaining = extractTrailingUnknownToken(meta)
		return group, remaining
	}

	// Strategy 4: Space-separated metadata ending with an unknown token.
	if hasMetadataTokens(meta) {
		group, remaining = extractTrailingUnknownToken(meta)
		return group, remaining
	}

	return "", meta
}

// lastDotSeparatedToken returns the last token in a dot/space-separated string.
func lastDotSeparatedToken(s string) string {
	s = strings.TrimRight(s, ". ")
	tokens := strings.FieldsFunc(s, isTokenSep)
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

// leadingBracketTags makes leading language tags available without scanning
// title words for languages. Website tags must not become languages (.pl).
func leadingBracketTags(name string) string {
	name = strings.TrimSpace(name)
	var tags []string
	for strings.HasPrefix(name, "[") {
		end := strings.IndexByte(name, ']')
		if end < 0 {
			break
		}
		content := name[1:end]
		if !isWebsiteLike(content) {
			// Explicit bracket tags establish context even without technical tags.
			tags = append(tags, strings.ToUpper(content))
		}
		name = strings.TrimLeft(name[end+1:], " ._-")
	}
	return strings.Join(tags, " ")
}

// releaseLanguageName handles explicit language tags, including subtitle
// affixes (VOSTFR, SweSub) and the TRUEFRENCH audio-language marker.
func releaseLanguageName(token string) string {
	if language := releaseLanguageNames[token]; language != "" {
		return language
	}
	for _, prefix := range []string{"vost", "true"} {
		if rest, ok := strings.CutPrefix(token, prefix); ok {
			if language := releaseLanguageNames[rest]; language != "" {
				return language
			}
		}
	}
	for _, suffix := range []string{"subtitles", "subtitle", "subs", "sub", "audio"} {
		if rest, ok := strings.CutSuffix(token, suffix); ok {
			if language := releaseLanguageNames[rest]; language != "" {
				return language
			}
		}
	}
	return ""
}

// extractLanguages identifies generic multilingual markers and explicit
// audio or subtitle language tags from the metadata region.
func extractLanguages(meta string) ([]string, bool) {
	var languages []string
	seen := make(map[string]bool)
	multilingual := false

	tokens := strings.FieldsFunc(meta, isMetaSep)
	for i, token := range tokens {
		lower := strings.ToLower(token)
		upper := token == strings.ToUpper(token)
		metadataNeighbor := i > 0 && isKnownToken(tokens[i-1]) || i+1 < len(tokens) && isKnownToken(tokens[i+1])
		switch lower {
		case "multi", "multiple", "dual", "dl":
			// DL means dual language, except in the source tag WEB-DL.
			if lower == "dl" && i > 0 && strings.EqualFold(tokens[i-1], "web") {
				continue
			}
			multilingual = multilingual || upper || metadataNeighbor
			continue
		}
		language := releaseLanguageName(lower)
		// Weak words and short codes can also be episode-title text (It Was
		// the Change, Cinco de Mayo, Greek Life). Require release-tag context.
		if len(lower) == 2 && !upper || strings.EqualFold(lower, language) && !upper && !metadataNeighbor {
			continue
		}
		if language != "" && !seen[language] {
			languages = append(languages, language)
			seen[language] = true
		}
	}

	return languages, multilingual || len(languages) > 1
}

// isKnownToken checks if a lowercase token is a known metadata keyword,
// screen size, codec, or language.
func isKnownToken(tok string) bool {
	// Parenthesis fragments in a credited group name must not become a
	// metadata anchor merely because trimming leaves an empty string.
	if strings.TrimSpace(tok) != "" && strings.Trim(tok, "() ") == "" {
		return false
	}
	tok = strings.ToLower(strings.Trim(tok, "[]()-_. "))
	if tok == "" {
		return true
	}
	if knownMetadataTokens[tok] {
		return true
	}
	if releaseLanguageName(tok) != "" {
		return true
	}
	if _, ok := screenSizeDimensions[tok]; ok {
		return true
	}
	if parts := strings.Split(tok, "-"); len(parts) > 1 {
		allLanguages := true
		for _, part := range parts {
			if releaseLanguageName(part) == "" {
				allLanguages = false
				break
			}
		}
		if allLanguages {
			return true
		}
	}
	compact := strings.ReplaceAll(tok, "-", "")
	if compact != tok && knownMetadataTokens[compact] {
		return true
	}
	if reScreenSizeToken.MatchString(tok) || reScreenSizeNumberToken.MatchString(tok) {
		return true
	}
	if extractVideoCodec(tok) != "" || extractAudioCodec(tok) != "" {
		return true
	}
	// Check tiny numeric codec fragments (from split channel configs like 5.1).
	if tok == "1" || tok == "0" {
		return true
	}
	return false
}

// isCodecLetterInContext checks if tokens[i] is a codec letter ("h" or "x")
// followed by a codec number ("264" or "265").
func isCodecLetterInContext(tokens []string, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}
	tok := strings.ToLower(strings.TrimSpace(tokens[i]))
	if tok != "h" && tok != "x" {
		return false
	}
	if i+1 >= len(tokens) {
		return false
	}
	next := strings.ToLower(strings.TrimSpace(tokens[i+1]))
	return next == "264" || next == "265"
}

func isCodecNumberToken(tok string) bool {
	tok = strings.TrimSpace(strings.ToLower(tok))
	return tok == "264" || tok == "265"
}

// isCompoundToken checks if a hyphen connects two parts of a compound token.
func isCompoundToken(before, after string) bool {
	lastToken := strings.ToLower(lastDotSeparatedToken(before))
	afterLower := strings.ToLower(strings.TrimRight(after, ". "))

	// Get just the first word of afterLower for compound check.
	afterFirstWord := afterLower
	if idx := strings.IndexAny(afterLower, ". "); idx >= 0 {
		afterFirstWord = afterLower[:idx]
	}

	compound := lastToken + "-" + afterFirstWord
	return compoundTokens[compound]
}

// cleanReleaseGroup cleans up a release group name.
func cleanReleaseGroup(s string) string {
	s = strings.TrimSpace(strings.TrimRight(s, ". "))
	s = strings.Trim(s, "[]()")
	s = reParenDomainSuffix.ReplaceAllString(s, "")
	s = strings.TrimSpace(strings.TrimRight(s, ". "))
	return s
}

func formatReleaseGroup(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	if strings.Contains(s, "[") {
		s = strings.ReplaceAll(s, "][", "] [")
		s = strings.TrimSpace(s)
	}

	if m := reTrailingParen.FindStringSubmatch(s); m != nil {
		rawInside := strings.TrimSpace(m[1])
		compactInside := strings.ToLower(strings.ReplaceAll(rawInside, " ", ""))
		if isLikelyDomainTag(compactInside) {
			base := strings.TrimSpace(strings.TrimRight(reTrailingParen.ReplaceAllString(s, ""), "."))
			return base
		}

		hadLeadingSpace := strings.HasPrefix(m[1], " ") || strings.HasPrefix(m[1], ".")
		inside := m[1]
		inside = strings.ReplaceAll(inside, ".", " ")
		inside = reSpaces.ReplaceAllString(inside, " ")
		inside = strings.TrimSpace(inside)
		if hadLeadingSpace {
			inside = " " + inside
		}
		if inside != "" {
			base := strings.TrimSpace(reTrailingParen.ReplaceAllString(s, ""))
			base = strings.TrimSpace(strings.TrimRight(base, "."))
			if base != "" {
				return strings.TrimSpace(base + " (" + inside + ")")
			}
		}
	}

	return s
}

// extractTrailingUnknownToken finds an unknown token only if it appears after
// the last known metadata token, preventing episode-title words from being
// treated as release groups.
func extractTrailingUnknownToken(meta string) (group, remaining string) {
	sep := "."
	if !strings.Contains(meta, ".") {
		sep = " "
		if strings.Contains(meta, "_") {
			sep = "_"
		}
	}

	tokens := strings.FieldsFunc(meta, isTokenSep)
	if len(tokens) == 0 {
		return "", meta
	}

	skipTokens := map[string]bool{"com": true, "esubs": true, "e": true}
	if len(tokens) >= 2 && strings.EqualFold(tokens[len(tokens)-1], "com") {
		skipTokens[strings.ToLower(tokens[len(tokens)-2])] = true
	}

	lastKnown := -1
	for i, tok := range tokens {
		lower := strings.ToLower(tok)
		if skipTokens[lower] || isSkippableNumericToken(lower) || isKnownToken(lower) || isCodecNumberInContext(tokens, i) || isCodecLetterInContext(tokens, i) {
			lastKnown = i
		}
	}

	for i := len(tokens) - 1; i > lastKnown; i-- {
		tok := tokens[i]
		lower := strings.ToLower(tok)
		if skipTokens[lower] || isSkippableNumericToken(lower) || isKnownToken(lower) || isCodecNumberInContext(tokens, i) || isCodecLetterInContext(tokens, i) {
			continue
		}
		start := i
		for j := i - 1; j > lastKnown; j-- {
			jl := strings.ToLower(tokens[j])
			if skipTokens[jl] || isSkippableNumericToken(jl) || isKnownToken(jl) || isCodecNumberInContext(tokens, j) || isCodecLetterInContext(tokens, j) {
				break
			}
			start = j
		}

		before := tokens[:start]
		after := tokens[i+1:]
		parts := make([]string, 0, len(before)+len(after))
		parts = append(parts, before...)
		parts = append(parts, after...)
		groupTokens := strings.Join(tokens[start:i+1], sep)
		if isCodecNumberToken(strings.ToLower(strings.TrimSpace(groupTokens))) {
			continue
		}
		return strings.TrimSpace(groupTokens), strings.Join(parts, sep)
	}

	return "", meta
}

// containsDigit reports whether s contains at least one digit.
func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func isLikelyDomainTag(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	if strings.Contains(lower, " ") || !strings.Contains(lower, ".") {
		return false
	}
	parts := strings.Split(lower, ".")
	if len(parts) < 2 {
		return false
	}
	tld := parts[len(parts)-1]
	switch tld {
	case "com", "net", "org", "to", "re", "ws", "cc", "info", "biz", "io", "tv", "me", "co", "pl", "my":
		return true
	default:
		return false
	}
}

// isNumericLike checks if a token is purely numeric or a numeric pattern like "700MB".
func isNumericLike(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	// Patterns like "700mb", "6ch".
	return reNumericSuffix.MatchString(s)
}

// A numeric series title such as 1923 is not an otherwise empty release year.
func extractTitleYear(title string) (int, string) {
	year, remaining := extractYear(title)
	if strings.Trim(stripBracketTags(remaining), " ._-()[]") == "" {
		return 0, title
	}
	return year, remaining
}

// extractYear extracts a year from the given string.
func extractYear(s string) (int, string) {
	years := findYearMatches(s)
	if len(years) == 0 {
		return 0, s
	}
	m := years[len(years)-1]

	yearStr := s[m.start:m.end]
	year, _ := strconv.Atoi(yearStr)

	result := s[:m.start] + s[m.end:]
	return year, result
}

func splitMovieRegions(name string) (titleRegion, metaRegion string) {
	if strings.TrimSpace(name) == "" {
		return "", ""
	}

	name = stripWebsitePrefix(name)

	if years := findYearMatches(name); len(years) > 0 {
		last := years[len(years)-1]
		if last.start == 0 {
			after := strings.TrimLeft(name[last.end:], " ._-")
			boundary := firstMetadataTokenIndex(after)
			if after == "" || boundary == 0 {
				return name[:last.end], after
			}
			if boundary < 0 {
				return after, name[:last.end]
			}
			return strings.TrimRight(after[:boundary], " ._-"), name[:last.end] + " " + after[boundary:]
		}
		split := last.start
		// If the year is inside parentheses (e.g., "(Label 1957)"),
		// split at the opening paren so the label stays in meta, not title.
		if split > 0 {
			if openParen := strings.LastIndex(name[:last.start], "("); openParen >= 0 {
				between := name[openParen:last.start]
				if !strings.Contains(between, ")") {
					split = openParen
				}
			}
		}
		for _, span := range reBracketTag.FindAllStringIndex(name[:split], -1) {
			// Leading tags do not end the title before it has begun.
			if strings.Trim(stripBracketTags(name[:span[0]]), " ._-") == "" {
				continue
			}
			if hasMetadataTokens(name[span[0]:span[1]]) {
				split = span[0]
				break
			}
		}
		title := strings.TrimRight(name[:split], " ._-([")
		meta := strings.TrimLeft(name[split:], " ._-")
		return title, meta
	}

	if split := firstMetadataTokenIndex(name); split >= 0 {
		title := strings.TrimRight(name[:split], " ._-")
		meta := strings.TrimLeft(name[split:], " ._-")
		return title, meta
	}

	return name, ""
}

func firstMetadataTokenIndex(s string) int {
	type tokenPos struct {
		tok   string
		start int
	}

	var tokens []tokenPos
	start := -1
	for i, r := range s {
		sep := r == '.' || r == ' ' || r == '_' || r == '-' || r == '(' || r == ')' || r == '[' || r == ']'
		if sep {
			if start >= 0 {
				tok := strings.Trim(strings.ToLower(s[start:i]), "[]()")
				if tok != "" {
					tokens = append(tokens, tokenPos{tok, start})
				}
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tok := strings.Trim(strings.ToLower(s[start:]), "[]()")
		if tok != "" {
			tokens = append(tokens, tokenPos{tok, start})
		}
	}

	// Build string slice for context-aware checks.
	toks := make([]string, len(tokens))
	for i, t := range tokens {
		toks[i] = t.tok
	}

	for i, t := range tokens {
		if t.tok == "0" || t.tok == "1" {
			continue
		}
		language := releaseLanguageName(t.tok)
		weakLanguage := language != "" && (len(t.tok) == 2 || strings.EqualFold(t.tok, language))
		if weakLanguage {
			// A language word alone can be the title (French), and short codes
			// need uppercase release-tag context (not This Is It).
			original := s[t.start : t.start+len(t.tok)]
			if i == 0 || len(t.tok) == 2 && original != strings.ToUpper(original) ||
				i+1 == len(tokens) || !isKnownToken(tokens[i+1].tok) {
				continue
			}
		}
		switch t.tok {
		case "audio", "sub", "subs", "subtitle", "subtitles", "multiple":
			if i == 0 || i+1 == len(tokens) || !isKnownToken(tokens[i+1].tok) {
				continue
			}
		case "truefrench", "vostfr", "dl", "uncut", "unrated", "dc":
			if i+1 == len(tokens) || !isKnownToken(tokens[i+1].tok) {
				continue
			}
		}
		if isKnownToken(t.tok) || isCodecLetterInContext(toks, i) || isCodecNumberInContext(toks, i) {
			return t.start
		}
	}
	return -1
}

func extractMoviePrefixGroup(titleRegion string) (group, cleanedTitle string) {
	cleaned := strings.TrimSpace(strings.TrimRight(titleRegion, "- "))

	if hyphen := strings.Index(cleaned, "-"); hyphen > 0 {
		prefix := strings.TrimSpace(cleaned[:hyphen])
		rest := strings.TrimSpace(cleaned[hyphen+1:])
		if len(prefix) > 1 && isLikelyPrefixGroup(prefix) && rest != "" {
			return prefix, rest
		}
	}

	if open := strings.LastIndex(cleaned, "("); open >= 0 && strings.HasSuffix(cleaned, ")") {
		inside := strings.TrimSpace(cleaned[open+1 : len(cleaned)-1])
		before := strings.TrimSpace(cleaned[:open])
		if inside != "" && before != "" {
			if isNumericLike(strings.ToLower(inside)) {
				return "", before
			}
			return inside, before
		}
	}

	return "", cleaned
}

func isLikelyPrefixGroup(s string) bool {
	if s == "" || len(s) > 16 || strings.Contains(s, " ") {
		return false
	}
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
			continue
		}
		if unicode.IsDigit(r) || r == '.' || r == '_' {
			continue
		}
		return false
	}
	return hasLetter
}

func cleanMovieTitle(raw string) string {
	raw = stripMovieSubtitle(raw)
	raw = stripBracketTags(raw)
	raw = reScreenSizeInTitle.ReplaceAllString(raw, "$1$3")

	title := cleanMovieTitleText(raw)
	title = stripMovieEditions(title)

	title = strings.TrimSpace(strings.TrimLeftFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\'' && r != '!'
	}))
	title = strings.TrimSpace(strings.TrimRightFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\'' && r != '!'
	}))
	title = strings.TrimSpace(reUnclosedParen.ReplaceAllString(title, ""))
	// Strip trailing channel configs that leaked from meta (e.g., "2.0" → "2 0" after normalization).
	for _, suffix := range []string{" 2 0", " 5 1", " 7 1"} {
		if strings.HasSuffix(title, suffix) {
			title = strings.TrimSpace(strings.TrimSuffix(title, suffix))
			break
		}
	}
	return title
}

// extractScreenSize extracts the screen size from the metadata region.
// If multiple are present, prefer the last one.
func extractScreenSize(meta string) string {
	tokens := strings.FieldsFunc(meta, func(r rune) bool {
		return isMetaSep(r)
	})
	last := ""
	for _, tok := range tokens {
		tok = strings.ToLower(strings.TrimSpace(tok))
		if size, ok := screenSizeDimensions[tok]; ok {
			last = size
		} else if tok == "4k" {
			last = "2160p"
		} else if m := reScreenSizeToken.FindStringSubmatch(tok); m != nil {
			last = m[1] + m[2]
		} else if m := reScreenSizeNumberToken.FindStringSubmatch(tok); m != nil {
			last = m[1] + "p"
		}
	}
	return last
}

// extractVideoCodec extracts the video codec from the metadata region.
func extractVideoCodec(meta string) string {
	if reVideoCodecH265.MatchString(meta) || reVideoCodecH265Spaced.MatchString(meta) {
		return "H.265"
	}
	if reVideoCodecH264.MatchString(meta) || reVideoCodecH264Spaced.MatchString(meta) {
		return "H.264"
	}
	if reVideoCodecXvid.MatchString(meta) {
		return "Xvid"
	}
	if reVideoCodecAV1.MatchString(meta) {
		return "AV1"
	}
	if reVideoCodecVP9.MatchString(meta) {
		return "VP9"
	}
	if reVideoCodecVC1.MatchString(meta) {
		return "VC-1"
	}
	return ""
}

// extractAudioCodec extracts the audio codec from the metadata region.
func extractAudioCodec(meta string) string {
	meta = strings.ReplaceAll(meta, "_", " ")
	if reAudioTrueHD.MatchString(meta) {
		return "Dolby TrueHD"
	}
	if reAudioDTS.MatchString(meta) {
		return "DTS"
	}
	if reAudioFLAC.MatchString(meta) {
		return "FLAC"
	}
	if reAudioOPUS.MatchString(meta) {
		return "Opus"
	}
	if reAudioDDP.MatchString(meta) {
		return "Dolby Digital Plus"
	}
	if reAudioEAC3.MatchString(meta) {
		return "Dolby Digital Plus"
	}
	if reAudioDD.MatchString(meta) {
		return "Dolby Digital"
	}
	if reAudioAAC.MatchString(meta) {
		return "AAC"
	}
	if reAudioMP3.MatchString(meta) {
		return "MP3"
	}
	return ""
}

// extractTitle extracts and cleans the title from the title region.
func extractTitle(titleRegion string, seStart, seEnd int, fullName string) string {
	title := titleRegion
	title = stripBracketTags(stripWebsitePrefix(title))
	title = strings.Trim(title, " ._-([")

	// If the title region is empty but we have a season/episode marker,
	// the title might be after the marker (e.g., "s01e01 Neon Forge.mp4").
	if strings.TrimSpace(title) == "" && seStart >= 0 && seEnd >= 0 && seEnd < len(fullName) {
		afterSE := strings.TrimLeft(fullName[seEnd:], ". _-])")
		title = extractTitleBeforeMetadata(afterSE)
	}

	return cleanTitle(title)
}

// extractTitleBeforeMetadata gets the title portion before metadata tokens start.
func extractTitleBeforeMetadata(s string) string {
	s = stripBracketTags(s)
	if boundary := firstMetadataTokenIndex(s); boundary >= 0 {
		s = s[:boundary]
	}
	return strings.Join(strings.FieldsFunc(s, isTokenSep), " ")
}

// cleanTitle converts a raw title region into a clean title string.
func cleanTitle(raw string) string {
	// Drop parenthesized years and leftover empty parentheses in titles.
	raw = reParenYear.ReplaceAllString(raw, " ")
	raw = reEmptyParen.ReplaceAllString(raw, " ")

	r := strings.NewReplacer(
		".", " ",
		"_", " ",
		"-", " ",
	)
	title := r.Replace(raw)
	title = reSpaces.ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)
	return title
}

func cleanMovieTitleText(raw string) string {
	title := normalizeTitleSeparators(raw)
	title = reSpaces.ReplaceAllString(title, " ")
	return strings.TrimSpace(title)
}

func normalizeTitleSeparators(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))

	runes := []rune(raw)
	for i, r := range runes {
		switch r {
		case '.', '_':
			b.WriteRune(' ')
		case '-':
			prevWord := i > 0 && (unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1]))
			nextWord := i+1 < len(runes) && (unicode.IsLetter(runes[i+1]) || unicode.IsDigit(runes[i+1]))
			if prevWord && nextWord {
				b.WriteRune('-')
			} else {
				b.WriteRune(' ')
			}
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func stripMovieSubtitle(raw string) string {
	if idx := strings.Index(raw, ".-."); idx >= 0 {
		return strings.TrimSpace(raw[:idx])
	}
	if idx := strings.Index(raw, " - "); idx >= 0 {
		return strings.TrimSpace(raw[:idx])
	}
	return raw
}

func stripMovieEditions(title string) string {
	suffixes := []string{
		"director's cut",
		"directors cut",
		"extended cut",
		"extended",
		"remastered",
	}

	lower := strings.ToLower(strings.TrimSpace(title))
	for _, s := range suffixes {
		if strings.HasSuffix(lower, " "+s) || lower == s {
			title = strings.TrimSpace(title[:len(title)-len(s)])
			lower = strings.ToLower(strings.TrimSpace(title))
		}
	}

	title = strings.TrimSpace(rePartSuffix.ReplaceAllString(title, ""))

	return title
}

func stripBracketTags(raw string) string {
	return strings.TrimSpace(reBracketTag.ReplaceAllString(raw, " "))
}

func stripWebsitePrefix(name string) string {
	name = strings.TrimSpace(name)

	for strings.HasPrefix(name, "[") {
		end := strings.Index(name, "]")
		if end < 0 {
			break
		}
		inside := strings.TrimSpace(name[1:end])
		if !isWebsiteLike(inside) {
			break
		}
		name = strings.TrimLeft(name[end+1:], " ._-")
	}

	name = reWebsitePrefix.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

func isWebsiteLike(s string) bool {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	if len(parts) == 0 {
		return false
	}
	joined := strings.Join(parts, "")
	if strings.HasPrefix(joined, "www.") {
		return true
	}
	return isLikelyDomainTag(joined)
}

type yearMatch struct {
	start int
	end   int
}

func findYearMatches(s string) []yearMatch {
	var out []yearMatch
	for i := 0; i+4 <= len(s); i++ {
		chunk := s[i : i+4]
		if !isAllDigits(chunk) {
			continue
		}
		y, _ := strconv.Atoi(chunk)
		if y < 1900 || y > 2099 {
			continue
		}
		leftOK := i == 0 || isYearBoundary(s[i-1])
		rightOK := i+4 == len(s) || isYearBoundary(s[i+4])
		if leftOK && rightOK {
			out = append(out, yearMatch{start: i, end: i + 4})
		}
	}
	return out
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isYearBoundary(b byte) bool {
	switch b {
	case '.', ' ', '_', '-', '(', ')', '[', ']':
		return true
	default:
		return false
	}
}

func isSkippableNumericToken(s string) bool {
	if s == "264" || s == "265" {
		return false
	}
	return isNumericLike(s)
}

func isCodecNumberInContext(tokens []string, i int) bool {
	if i < 0 || i >= len(tokens) {
		return false
	}
	tok := strings.ToLower(strings.TrimSpace(tokens[i]))
	if tok != "264" && tok != "265" {
		return false
	}
	if i == 0 {
		return false
	}
	prev := strings.ToLower(strings.TrimSpace(tokens[i-1]))
	return prev == "x" || prev == "h"
}
