package patterns

import (
	"regexp"
	"strings"
)

var reWebsiteEpisodeLike = regexp.MustCompile(`(?i)\bs\d{1,3}(?:[ ._-]?e|x(?:e)?)[ ._-]?\d{1,3}\b|\b\d{1,2}x\d{1,3}\b`)

func registerWebsitePatterns(m *Matcher) {
	AddPatternSpec(m, "website", `(?i)\b(?:www\.)?[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:com|org|net|io|ru|to)\b`, PatternSpec{
		Validator: func(input string, start, end int) bool {
			// Avoid swallowing episode-like spans ending in ".to" inside titles.
			return !reWebsiteEpisodeLike.MatchString(input[start:end])
		},
		Formatter: func(raw string) any { return strings.ToLower(raw) },
	})
}
