package patterns

import (
	"regexp"
	"strconv"
)

func registerPartPatterns(m *Matcher) {
	reNum := regexp.MustCompile(`\d+`)
	AddPatternSpec(m, "part", `(?i)\b(?:part|pt)[ ._-]?\d{1,2}\b`, PatternSpec{
		Formatter: func(raw string) any {
			n := reNum.FindString(raw)
			if n == "" {
				return raw
			}
			v, convErr := strconv.Atoi(n)
			if convErr != nil {
				return raw
			}
			return v
		},
	})
}
