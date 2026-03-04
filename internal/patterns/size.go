package patterns

import (
	"regexp"
	"strings"
)

func registerSizePatterns(m *Matcher) {
	reSpace := regexp.MustCompile(`\s+`)
	AddPatternSpec(m, "size", `(?i)\b\d+(?:\.\d+)?[ ]?(?:kb|mb|gb|tb)\b`, PatternSpec{
		Formatter: func(raw string) any {
			raw = strings.ToUpper(strings.TrimSpace(raw))
			raw = reSpace.ReplaceAllString(raw, "")
			return raw
		},
	})
}
