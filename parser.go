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

	// Step 5+: Extract type-specific fields.
	if g.Type == Episode {
		// Handle " - Episode Title" and pure episode title patterns.
		metaRegion = stripEpisodeTitle(metaRegion)

		g.ReleaseGroup, metaRegion = extractReleaseGroup(metaRegion, doubleExtBracket)
		g.ReleaseGroup = formatReleaseGroup(g.ReleaseGroup)
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

		g.Year, metaRegion = extractLastYear(metaRegion)
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

		if !isCompoundToken(beforeHyphen, afterHyphen) && afterHyphen != "" {
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
	// Check tiny numeric codec fragments.
	if tok == "1" || tok == "0" {
		return true
	}
	if tok == "h" || tok == "x" {
		return true
	}
	return false
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
	s = strings.Trim(s, "[]()")
	s = regexp.MustCompile(`\s*\.?\s*\([^)]*\.[^)]*\)\s*$`).ReplaceAllString(s, "")
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

	paren := regexp.MustCompile(`\s*\(([^)]+)\)\s*$`)
	if m := paren.FindStringSubmatch(s); m != nil {
		rawInside := strings.TrimSpace(m[1])
		compactInside := strings.ToLower(strings.ReplaceAll(rawInside, " ", ""))
		if isLikelyDomainTag(compactInside) {
			base := strings.TrimSpace(strings.TrimRight(paren.ReplaceAllString(s, ""), "."))
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
			base := strings.TrimSpace(paren.ReplaceAllString(s, ""))
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
		if skipTokens[lower] || isSkippableNumericToken(lower) || isKnownToken(lower) || isCodecNumberInContext(tokens, i) {
			lastKnown = i
		}
	}

	for i := len(tokens) - 1; i > lastKnown; i-- {
		tok := tokens[i]
		lower := strings.ToLower(tok)
		if skipTokens[lower] || isSkippableNumericToken(lower) || isKnownToken(lower) || isCodecNumberInContext(tokens, i) {
			continue
		}
		start := i
		for j := i - 1; j > lastKnown; j-- {
			jl := strings.ToLower(tokens[j])
			if skipTokens[jl] || isSkippableNumericToken(jl) || isKnownToken(jl) || isCodecNumberInContext(tokens, j) {
				break
			}
			start = j
		}

		before := tokens[:start]
		after := tokens[i+1:]
		var parts []string
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

func extractLastYear(s string) (int, string) {
	years := findYearMatches(s)
	if len(years) == 0 {
		return 0, s
	}

	last := years[len(years)-1]
	yearStr := s[last.start:last.end]
	year, _ := strconv.Atoi(yearStr)

	result := s[:last.start] + s[last.end:]
	return year, result
}

func splitMovieRegions(name string) (titleRegion, metaRegion string) {
	if strings.TrimSpace(name) == "" {
		return "", ""
	}

	name = stripWebsitePrefix(name)

	if years := findYearMatches(name); len(years) > 0 {
		last := years[len(years)-1]
		split := last.start
		if split > 0 {
			leftToken := strings.ToLower(strings.Trim(name[:split], " ._-()[]"))
			leftToken = strings.ToLower(lastDotSeparatedToken(leftToken))
			if leftToken == "harvest" {
				split = strings.LastIndex(name[:last.start], "(")
				if split < 0 {
					split = last.start
				}
			}
		}
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
	raw = stripMovieSubtitle(raw)
	raw = stripBracketTags(raw)
	raw = regexp.MustCompile(`(?i)(^|[ ._-])(2160|1080|720|540|480)([ ._-]|$)`).ReplaceAllString(raw, "$1$3")

	title := cleanMovieTitleText(raw)
	title = stripMovieEditions(title)
	lower := strings.ToLower(title)
	if strings.HasSuffix(lower, " director's cut") {
		title = strings.TrimSpace(title[:len(title)-len(" director's cut")])
	} else if strings.HasSuffix(lower, " directors cut") {
		title = strings.TrimSpace(title[:len(title)-len(" directors cut")])
	}

	title = strings.TrimSpace(strings.TrimLeftFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\'' && r != '!'
	}))
	title = strings.TrimSpace(strings.TrimRightFunc(title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\'' && r != '!'
	}))
	title = strings.TrimSpace(regexp.MustCompile(`\s*\([^)]*$`).ReplaceAllString(title, ""))
	if strings.HasSuffix(title, " 2 0") {
		title = strings.TrimSpace(strings.TrimSuffix(title, " 2 0"))
	}
	return title
}

// extractScreenSize extracts the screen size from the metadata region.
// If multiple are present, prefer the last one.
func extractScreenSize(meta string) string {
	tokens := strings.FieldsFunc(meta, func(r rune) bool {
		return r == '.' || r == ' ' || r == '_' || r == '[' || r == ']' || r == '(' || r == ')' || r == '-'
	})
	last := ""
	for _, tok := range tokens {
		if m := reScreenSizeToken.FindStringSubmatch(strings.ToLower(strings.TrimSpace(tok))); m != nil {
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
	if reAudioDTS.MatchString(meta) {
		return "DTS"
	}
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
	title = stripWebsitePrefix(title)

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
	// Drop parenthesized years and leftover empty parentheses in titles.
	raw = regexp.MustCompile(`(?i)\(\s*(?:19|20)\d{2}\s*\)`).ReplaceAllString(raw, " ")
	raw = regexp.MustCompile(`\(\s*\)`).ReplaceAllString(raw, " ")

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

	partRe := regexp.MustCompile(`(?i)\s+part\s+[0-9ivx]+$`)
	title = strings.TrimSpace(partRe.ReplaceAllString(title, ""))

	return title
}

func stripBracketTags(raw string) string {
	bracketToken := regexp.MustCompile(`(?:^|[\s._-])\[[^\]]+\]`)
	return strings.TrimSpace(bracketToken.ReplaceAllString(raw, " "))
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

	webPrefix := regexp.MustCompile(`(?i)^www\.[a-z0-9-]+(?:\.[a-z0-9-]+)+[._\-\s]+`)
	name = webPrefix.ReplaceAllString(name, "")
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
