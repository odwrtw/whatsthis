package patterns

import (
	"regexp"
	"strings"
)

func registerBitRatePatterns(m *Matcher) {
	reSpace := regexp.MustCompile(`\s+`)
	AddPatternSpec(m, "bit_rate", `(?i)\b\d+(?:\.\d+)?[ ]?(?:k|m)(?:b(?:it)?s?)?(?:ps)?\b`, PatternSpec{
		Formatter: func(raw string) any {
			raw = strings.ToLower(strings.TrimSpace(raw))
			raw = reSpace.ReplaceAllString(raw, "")
			raw = strings.ReplaceAll(raw, "bits", "b")
			raw = strings.ReplaceAll(raw, "bit", "b")
			if strings.HasSuffix(raw, "b") {
				raw += "ps"
			}
			return strings.ToUpper(raw)
		},
	})
}
