package whatsthis

import (
	"regexp"
	"strings"
)

var reHyphenatedEnglishSubs = regexp.MustCompile(`(?i)e-subs`)

// normalizeSubtitleAliases keeps E-Subs together before metadata tokenization,
// so its E is not mistaken for a language code or discarded. Plain ESubs is
// already handled by lookupReleaseLanguage. Only complete tags are rewritten.
func normalizeSubtitleAliases(meta string) string {
	var normalized strings.Builder
	last := 0
	for _, match := range reHyphenatedEnglishSubs.FindAllStringIndex(meta, -1) {
		start, end := match[0], match[1]
		if !metadataAliasBoundary(meta, start-1) || !metadataAliasBoundary(meta, end) {
			continue
		}
		normalized.WriteString(meta[last:start])
		normalized.WriteString("ESubs")
		last = end
	}
	if last == 0 {
		return meta
	}
	normalized.WriteString(meta[last:])
	return normalized.String()
}

func metadataAliasBoundary(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return true
	}
	c := s[i]
	return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9')
}
