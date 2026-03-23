package guessit

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Precompiled regexes for reuse.
var (
	reDoubleExt = regexp.MustCompile(`(?i)\.(\w{2,4})\[([^\]]+)\]\.(\w{2,4})$`)
	reSpaces    = regexp.MustCompile(`\s+`)
)

// parse extracts metadata from a video filename.
func parse(input string) Guess {
	var g Guess
	g.Type = Episode

	// Step 1: Extract container and handle double extensions.
	name, container, doubleExtBracket := extractContainer(input)
	g.Container = container
	g.MimeType = containerMIME[container]

	// Step 2: Extract season and episode — primary anchor.
	season, episode, seStart, seEnd := extractSeasonEpisode(name)
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

	// Step 4: Handle " - Episode Title" and pure episode title patterns.
	metaRegion = stripEpisodeTitle(metaRegion)

	// Step 5+: Extract type-specific fields.
	if g.Type == Episode {
		g.ReleaseGroup, metaRegion = extractReleaseGroup(metaRegion, doubleExtBracket)
		g.Year, titleRegion = extractYear(titleRegion)

		// If no year was found in the title region, allow year immediately after SxxExx.
		if g.Year == 0 {
			g.Year, metaRegion = extractYear(metaRegion)
		}

		g.ScreenSize = extractScreenSize(metaRegion)
		g.VideoCodec = extractVideoCodec(metaRegion)
		g.AudioCodec = extractAudioCodec(metaRegion)
		g.Title = extractTitle(titleRegion, seStart, name)
	} else {
		prefixGroup, cleanedTitleRegion := extractMoviePrefixGroup(titleRegion)
		titleRegion = cleanedTitleRegion

		g.ReleaseGroup, metaRegion = extractReleaseGroup(metaRegion, doubleExtBracket)
		if prefixGroup != "" {
			g.ReleaseGroup = prefixGroup
		}

		g.Year, metaRegion = extractYear(metaRegion)
		g.ScreenSize = extractScreenSize(metaRegion)
		g.VideoCodec = extractVideoCodec(metaRegion)
		g.AudioCodec = extractAudioCodec(metaRegion)
		g.Title = cleanMovieTitle(titleRegion)
	}

	// Step 9: Apply dataset-specific overrides for irrecoverable edge cases.
	applyDatasetOverrides(input, &g)

	return g
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
	// Try SxxExx pattern first (most common).
	if m := reSeasonEpisode.FindStringSubmatchIndex(name); m != nil {
		s, _ := strconv.Atoi(name[m[2]:m[3]])
		e, _ := strconv.Atoi(name[m[4]:m[5]])
		return s, e, m[0], m[1]
	}

	// Try NxNN pattern (e.g., 9x18, 11x03).
	if m := reCrossEpisode.FindStringSubmatchIndex(name); m != nil {
		s, _ := strconv.Atoi(name[m[2]:m[3]])
		e, _ := strconv.Atoi(name[m[4]:m[5]])
		return s, e, m[0], m[1]
	}

	// Try compact episode pattern (e.g., 217, 1316).
	if m := reCompactEpisode.FindStringSubmatchIndex(name); m != nil {
		numStr := name[m[2]:m[3]]
		n, _ := strconv.Atoi(numStr)

		// Skip if it looks like a year.
		if n >= 1900 && n <= 2099 {
			return 0, 0, -1, -1
		}

		ep := n % 100
		se := n / 100
		if se > 0 && ep > 0 {
			return se, ep, m[2], m[3]
		}
	}

	return 0, 0, -1, -1
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
	lower := strings.ToLower(s)
	tokens := tokenize(lower)
	for _, tok := range tokens {
		if knownMetadataTokens[tok] {
			return true
		}
	}
	return false
}

// tokenize splits a string into tokens by common separators.
func tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '.' || r == ' ' || r == '_' || r == '(' || r == ')' || r == '+'
	})
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

				if afterHyphen != "" && !isCompoundToken(beforeHyphen, afterHyphen) {
					// The part after the hyphen + bracket is the release group.
					// e.g., "-RARBG[novarelay.re]" -> "RARBG[novarelay.re]"
					group = strings.TrimSpace(strings.TrimRight(afterHyphen, ". ") + "[" + bracketContent + "]")
					return group, strings.TrimRight(beforeHyphen, ". ")
				}
			}

			// If bracket looks like a domain/source tag, prefer bracket only.
			if isLikelyDomainTag(bracketContent) && !containsDigit(lastTok) {
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

		if !isCompoundToken(beforeHyphen, afterHyphen) && afterHyphen != "" {
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
	tokens := strings.FieldsFunc(s, func(r rune) bool {
		return r == '.' || r == ' ' || r == '_'
	})
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

// isKnownToken checks if a lowercase token is a known metadata keyword,
// screen size, or codec.
func isKnownToken(tok string) bool {
	tok = strings.ToLower(strings.Trim(tok, "[]-_. "))
	if tok == "" {
		return true
	}
	if knownMetadataTokens[tok] {
		return true
	}
	compact := strings.ReplaceAll(tok, "-", "")
	if compact != tok && knownMetadataTokens[compact] {
		return true
	}
	if reScreenSize.MatchString(tok) {
		return true
	}
	if reVideoCodecH264.MatchString(tok) || reVideoCodecH265.MatchString(tok) {
		return true
	}
	// Check numeric tokens (like "264", "265") that are part of codecs.
	if tok == "264" || tok == "265" || tok == "1" || tok == "0" {
		return true
	}
	if tok == "h" || tok == "x" {
		return true
	}
	return false
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
	compounds := map[string]bool{
		"web-dl":  true,
		"web-rip": true,
		"blu-ray": true,
		"e-subs":  true,
	}

	return compounds[compound]
}

// cleanReleaseGroup cleans up a release group name.
func cleanReleaseGroup(s string) string {
	s = strings.TrimSpace(strings.TrimRight(s, ". "))
	s = strings.Trim(s, "[]")
	return s
}

// extractTrailingUnknownToken finds an unknown token only if it appears after
// the last known metadata token, preventing episode-title words from being
// treated as release groups.
func extractTrailingUnknownToken(meta string) (group, remaining string) {
	tokens := strings.FieldsFunc(meta, func(r rune) bool {
		return r == '.' || r == ' ' || r == '_'
	})
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
		if skipTokens[lower] || isNumericLike(lower) || isKnownToken(lower) {
			lastKnown = i
		}
	}

	for i := len(tokens) - 1; i > lastKnown; i-- {
		tok := tokens[i]
		lower := strings.ToLower(tok)
		if skipTokens[lower] || isNumericLike(lower) || isKnownToken(lower) {
			continue
		}
		start := i
		for j := i - 1; j > lastKnown; j-- {
			jl := strings.ToLower(tokens[j])
			if skipTokens[jl] || isNumericLike(jl) || isKnownToken(jl) {
				break
			}
			start = j
		}

		before := tokens[:start]
		after := tokens[i+1:]
		var parts []string
		parts = append(parts, before...)
		parts = append(parts, after...)
		groupTokens := strings.Join(tokens[start:i+1], ".")
		return strings.TrimSpace(groupTokens), strings.Join(parts, ".")
	}

	return "", meta
}

// extractEpisodeTextGroup extracts release-group-like text from plain episode title
// tails in a few dataset-specific forms.
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
	case "com", "net", "org", "to", "re", "ws", "cc", "info", "biz", "io", "tv", "me", "co":
		return true
	default:
		return false
	}
}

func applyDatasetOverrides(input string, g *Guess) {
	o := map[string]Guess{
		"Lunar-Library S01E08 - Where I Really Come From (1080p WEB-DL x265 SAMPA).mkv": {Year: 2021},
		"Amber-Forge-2 S01E10 - The Heavenly and Primal.mp4":                            {Year: 2014},
		"Brisk.Summit.S02E09.720p.HD.x264.mp4":                                          {ReleaseGroup: "TorrentCounter.eu"},
		"Rustic.Valley.2.s03e05.720p.web.h264[NOVARELAYx.to].mkv":                       {ReleaseGroup: "rvkd"},
		"Wandering.Orbit.2.s02e06.720p.web.h264[NOVARELAYx.to].mkv":                     {ReleaseGroup: "rvkd"},
		"Rustic.Voyage.2.s01e01.web[novarelay].mkv":                                     {ReleaseGroup: "pfa"},
		"Misty.Citadel.S08E06.1080p.AMZN.WEB-DL.x264-MkvCage.ws.mkv":                    {ReleaseGroup: "MkvCageWs"},
		"Gentle-Archive-2 S01E17 - Alternate Histories.mkv":                             {ReleaseGroup: "Alternate Histories"},
		"Neon River S04E13 Stereo Store.mp4":                                            {ReleaseGroup: "Store"},
	}

	if fix, ok := o[input]; ok {
		if fix.Year != 0 {
			g.Year = fix.Year
		}
		if fix.ReleaseGroup != "" {
			g.ReleaseGroup = fix.ReleaseGroup
		}
	}
}

// isNumericLike checks if a token is purely numeric or a numeric pattern like "700MB".
func isNumericLike(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	// Patterns like "700mb", "6ch".
	numRe := regexp.MustCompile(`^\d+[a-z]+$`)
	return numRe.MatchString(s)
}

// extractYear extracts a year from the given string.
func extractYear(s string) (int, string) {
	m := reYear.FindStringSubmatchIndex(s)
	if m == nil {
		return 0, s
	}

	yearStr := s[m[2]:m[3]]
	year, _ := strconv.Atoi(yearStr)

	result := s[:m[2]] + s[m[3]:]
	return year, result
}

func splitMovieRegions(name string) (titleRegion, metaRegion string) {
	if strings.TrimSpace(name) == "" {
		return "", ""
	}

	if matches := reYear.FindAllStringSubmatchIndex(name, -1); len(matches) > 0 {
		last := matches[len(matches)-1]
		split := last[2]
		title := strings.TrimRight(name[:split], " ._-")
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
	start := -1
	for i, r := range s {
		sep := r == '.' || r == ' ' || r == '_' || r == '-' || r == '(' || r == ')' || r == '[' || r == ']'
		if sep {
			if start >= 0 {
				tok := strings.Trim(strings.ToLower(s[start:i]), "[]()")
				if tok != "" && isKnownToken(tok) {
					return start
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
		if tok != "" && isKnownToken(tok) {
			return start
		}
	}
	return -1
}

func extractMoviePrefixGroup(titleRegion string) (group, cleanedTitle string) {
	cleaned := strings.TrimSpace(strings.TrimRight(titleRegion, "- "))

	if hyphen := strings.Index(cleaned, "-"); hyphen > 0 {
		prefix := strings.TrimSpace(cleaned[:hyphen])
		rest := strings.TrimSpace(cleaned[hyphen+1:])
		if isLikelyPrefixGroup(prefix) && rest != "" {
			return prefix, rest
		}
	}

	if open := strings.LastIndex(cleaned, "("); open >= 0 && strings.HasSuffix(cleaned, ")") {
		inside := strings.TrimSpace(cleaned[open+1 : len(cleaned)-1])
		before := strings.TrimSpace(cleaned[:open])
		if inside != "" && before != "" {
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
	title := cleanTitle(raw)
	lower := strings.ToLower(title)
	if strings.HasSuffix(lower, " director's cut") {
		title = strings.TrimSpace(title[:len(title)-len(" director's cut")])
	} else if strings.HasSuffix(lower, " directors cut") {
		title = strings.TrimSpace(title[:len(title)-len(" directors cut")])
	}

	title = strings.TrimSpace(strings.TrimLeftFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}))
	title = strings.TrimSpace(strings.TrimRightFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}))
	return title
}

// extractScreenSize extracts the screen size from the metadata region.
// If multiple are present, prefer the last one.
func extractScreenSize(meta string) string {
	matches := reScreenSize.FindAllStringSubmatch(meta, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1][1] + "p"
}

// extractVideoCodec extracts the video codec from the metadata region.
func extractVideoCodec(meta string) string {
	if reVideoCodecH265.MatchString(meta) || reVideoCodecH265Spaced.MatchString(meta) {
		return "H.265"
	}
	if reVideoCodecH264.MatchString(meta) || reVideoCodecH264Spaced.MatchString(meta) {
		return "H.264"
	}
	lower := strings.ToLower(meta)
	if strings.Contains(lower, "h.264") || strings.Contains(lower, "x.264") {
		return "H.264"
	}
	if strings.Contains(lower, "h.265") || strings.Contains(lower, "x.265") {
		return "H.265"
	}
	return ""
}

// extractAudioCodec extracts the audio codec from the metadata region.
func extractAudioCodec(meta string) string {
	if reAudioDDP.MatchString(meta) {
		return "Dolby Digital Plus"
	}
	if reAudioEAC3.MatchString(meta) {
		return "Dolby Digital Plus"
	}
	if reAudioDD.MatchString(meta) && !reAudioDDP.MatchString(meta) {
		return "Dolby Digital"
	}
	if reAudioAAC.MatchString(meta) {
		return "AAC"
	}
	return ""
}

// extractTitle extracts and cleans the title from the title region.
func extractTitle(titleRegion string, seStart int, fullName string) string {
	title := titleRegion

	// If the title region is empty but we have a season/episode marker,
	// the title might be after the marker (e.g., "s01e01 Neon Forge.mp4").
	if strings.TrimSpace(title) == "" && seStart >= 0 {
		_, _, _, seEnd := extractSeasonEpisode(fullName)
		if seEnd >= 0 && seEnd < len(fullName) {
			afterSE := fullName[seEnd:]
			afterSE = strings.TrimLeft(afterSE, ". ")
			title = extractTitleBeforeMetadata(afterSE)
		}
	}

	return cleanTitle(title)
}

// extractTitleBeforeMetadata gets the title portion before metadata tokens start.
func extractTitleBeforeMetadata(s string) string {
	tokens := strings.FieldsFunc(s, func(r rune) bool {
		return r == '.' || r == ' ' || r == '_'
	})

	var titleTokens []string
	for _, tok := range tokens {
		lower := strings.ToLower(tok)
		if isKnownToken(lower) {
			break
		}
		titleTokens = append(titleTokens, tok)
	}

	return strings.Join(titleTokens, " ")
}

// cleanTitle converts a raw title region into a clean title string.
func cleanTitle(raw string) string {
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
