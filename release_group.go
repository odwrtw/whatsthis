package whatsthis

import (
	"regexp"
	"strings"
)

var (
	reReleaseChecksum = regexp.MustCompile(`(?i)^[a-f0-9]{8}$`)
	reReleaseSize     = regexp.MustCompile(`(?i)^\d+(?:\.\d+)?(?:kb|mb|gb|tb|ch|bit|bits)$`)
	reReleaseDomain   = regexp.MustCompile(`(?i)(?:^|[.\s_\[(-])(?:www\.)?[a-z0-9-]+\.(?:com|net|org|to|re|ws|cc|info|biz|io|tv|me|co|pl|my)\b`)
)

// credibleReleaseGroup classifies a delimited attribution, not arbitrary title
// text. Numeric/checksum, language, source and presentation tags are not groups.
func credibleReleaseGroup(s string) bool {
	s = cleanReleaseGroup(s)
	if s == "" || strings.Trim(s, "0123456789._- ") == "" || reReleaseSize.MatchString(s) || reReleaseChecksum.MatchString(strings.Trim(s, "._- ")) || strings.ContainsAny(s, "/\\:#") || isWebsiteLike(s) || hasMetadataTokens(s) {
		return false
	}
	for _, token := range strings.FieldsFunc(strings.ToLower(s), isMetaSep) {
		if isPresentationToken(token) {
			return false
		}
		switch token {
		case "bd", "dvd", "mp4", "mkv", "avi", "webm", "mov", "m4v", "ts", "wmv", "flv", "movie", "hi10", "hi10p", "ma10p", "hdlight", "brip", "3d", "hsbs", "fsbs", "hou", "uhd", "commentary":
			return false
		}
	}
	return true
}

// Leading technical brackets are metadata too, not just things to remove from
// a title. Prepend them so explicit tags after the title still win conflicts.
func leadingTechnicalTags(name string) string {
	var tags []string
	name = strings.TrimSpace(name)
	for strings.HasPrefix(name, "[") {
		end := strings.IndexByte(name, ']')
		if end < 0 {
			break
		}
		content := name[1:end]
		if !isWebsiteLike(content) && hasMetadataTokens(content) {
			tags = append(tags, "["+content+"]")
		}
		name = strings.TrimLeft(name[end+1:], " ._-")
	}
	return strings.Join(tags, " ")
}

func leadingReleaseGroup(name string) string {
	name = strings.TrimSpace(name)
	for strings.HasPrefix(name, "[") {
		end := strings.IndexByte(name, ']')
		if end < 0 {
			break
		}
		content := strings.TrimSpace(name[1:end])
		if credibleReleaseGroup(content) && !reSeasonOnly.MatchString(content) && !reCrossEpisode.MatchString(content) {
			return content
		}
		name = strings.TrimLeft(name[end+1:], " ._-")
	}
	return ""
}

// primaryReleaseGroup applies the same precedence to movies, episodes and packs:
// explicit -GROUP, leading [GROUP], then a metadata-anchored trailing credit.
// Bracketed annotations never become part of an adjacent encoder's name.
func primaryReleaseGroup(name, meta string) (string, string) {
	meta = stripReleaseAnnotations(meta)
	cleaned := []byte(meta)
	bracketGroup, explicitBracket := "", ""
	for start := 0; start < len(meta); start++ {
		if meta[start] != '[' {
			continue
		}
		end := strings.IndexByte(meta[start:], ']')
		if end < 0 {
			break
		}
		end += start
		content := strings.TrimSpace(meta[start+1 : end])
		if hasMetadataTokens(content) {
			// Keep technical tags, including a possible x264-GROUP inside them.
			cleaned[start], cleaned[end] = ' ', ' '
		} else {
			before := strings.TrimRight(meta[:start], ". _")
			if credibleReleaseGroup(content) {
				if strings.HasSuffix(before, "-") {
					if explicitBracket == "" {
						explicitBracket = content
					}
				} else if bracketGroup == "" && (start == 0 || meta[start-1] == ' ' || meta[start-1] == ')' || isKnownToken(lastDotSeparatedToken(before))) {
					bracketGroup = content
				}
			}
			for i := start; i <= end; i++ {
				cleaned[i] = ' '
			}
		}
		start = end
	}
	meta = string(cleaned)
	// A codec followed by a hyphen is not part of the domain label, even
	// though DNS labels themselves may contain hyphens (AC3-GROUP.TV).
	for _, match := range reReleaseDomain.FindAllStringIndex(meta, -1) {
		start := match[0]
		if dash := strings.LastIndexByte(meta[start:match[1]], '-'); dash >= 0 {
			if hasTechnicalMetadata(meta[start : start+dash]) {
				start += dash + 1
			}
		}
		for i := start; i < match[1]; i++ {
			cleaned[i] = ' '
		}
	}
	meta = string(cleaned)
	metadataStart := firstMetadataTokenIndex(meta, false)

	// Scan from the metadata boundary, not backwards from the last hyphen:
	// Example-Raws and GCJM-0 are indivisible credited names.
	for i := 0; i < len(meta); i++ {
		if meta[i] != '-' {
			continue
		}
		before := meta[:i]
		after := strings.TrimLeft(meta[i+1:], " \t")
		if after == "" || isCompoundToken(before, after) || strings.TrimSpace(before) != "" && (metadataStart < 0 || metadataStart >= i) {
			continue
		}
		end := strings.IndexAny(after, " \t()[]")
		if end < 0 {
			end = len(after)
		}
		end = releaseCreditEnd(after[:end])
		candidate := strings.Trim(after[:end], ". _")
		if credibleReleaseGroup(candidate) {
			start := len(meta) - len(after)
			return candidate, meta[:i] + " " + meta[start+end:]
		}
	}
	if explicitBracket != "" {
		return explicitBracket, meta
	}
	if leading := leadingReleaseGroup(name); leading != "" {
		return leading, meta
	}
	if bracketGroup != "" {
		return bracketGroup, meta
	}
	if hasMetadataTokens(meta) {
		if group, remaining := trailingReleaseGroup(meta); credibleReleaseGroup(group) {
			return cleanReleaseGroup(group), remaining
		}
	}
	return "", meta
}

// Presentation/source labels share classification with metadata boundary and
// fallback-token detection. They must never become trailing group names.
func isPresentationToken(token string) bool {
	switch strings.ToLower(token) {
	case "hdr10", "hdr10plus", "hdr10+", "dolbyvision", "dv", "atmos", "remux", "brip", "hdlight", "bd", "dvd", "uhd", "3d", "hsbs", "fsbs", "hou", "hi10", "hi10p", "ma10p", "commentary":
		return true
	}
	return false
}

// A delimited credit may precede more technical tags (GROUP.AAC). Keep the
// original punctuation within the credit, stopping only at recognized tags.
func releaseCreditEnd(s string) int {
	for start := 0; start < len(s); {
		if s[start] == '.' || s[start] == '_' {
			start++
			continue
		}
		end := start
		for end < len(s) && s[end] != '.' && s[end] != '_' {
			end++
		}
		token := s[start:end]
		if token != "0" && token != "1" && (isKnownToken(token) || startsWithCodec(s[start:])) {
			return start
		}
		start = end
	}
	return len(s)
}

// Locate the unknown suffix using offsets: joining tokenized words would turn
// my_group into my.group or collapse punctuation in a dotted identifier.
func trailingReleaseGroup(meta string) (string, string) {
	groupStart, groupEnd := -1, -1
	for start := 0; start < len(meta); {
		if isTokenSep(rune(meta[start])) || strings.ContainsRune("()", rune(meta[start])) {
			start++
			continue
		}
		end := start
		for end < len(meta) && !isTokenSep(rune(meta[end])) && !strings.ContainsRune("()", rune(meta[end])) {
			end++
		}
		token := meta[start:end]
		if isKnownToken(token) || isNumericLike(strings.ToLower(token)) {
			groupStart, groupEnd = -1, -1
		} else {
			if groupStart < 0 {
				groupStart = start
			}
			groupEnd = end
		}
		start = end
	}
	if groupStart < 0 {
		return "", meta
	}
	return meta[groupStart:groupEnd], meta[:groupStart] + " " + meta[groupEnd:]
}

// Match bounded extension fragments in the original bytes. Unicode case
// folding can change byte lengths, so offsets into ToLower(s) are not safe.
func stripProcessingSuffix(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != '.' {
			continue
		}
		for ext := range containerMIME {
			end := i + len(ext) + 2
			if end <= len(s) && strings.EqualFold(s[i:end], "."+ext+"-") {
				return s[:i]
			}
		}
	}
	return s
}

// Annotations after an established metadata boundary are not alternate group
// names. Preserve parenthesized technical metadata for all field extractors.
func stripReleaseAnnotations(meta string) string {
	meta = stripProcessingSuffix(meta)
	cleaned := []byte(meta)
	for start := 0; start < len(meta); start++ {
		if meta[start] != '(' {
			continue
		}
		end := strings.IndexByte(meta[start:], ')')
		if end < 0 {
			break
		}
		end += start
		if start > 0 && meta[start-1] == '-' && credibleReleaseGroup(meta[start+1:end]) && !strings.ContainsAny(meta[start+1:end], " \t") {
			// x264-(GROUP) is an explicit credit, unlike GROUP (annotation).
			cleaned[start], cleaned[end] = ' ', ' '
		} else if !hasMetadataTokens(meta[start+1 : end]) {
			for i := start; i <= end; i++ {
				cleaned[i] = ' '
			}
		}
		start = end
	}
	return string(cleaned)
}
