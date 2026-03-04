package patterns

import "strconv"

func registerBonusPatterns(m *Matcher) {
	AddPatternSpec(m, "bonus", `(?i)\bx(\d{1,2})\b`, PatternSpec{
		Formatter: func(raw string) any {
			if len(raw) < 2 {
				return raw
			}
			v, convErr := strconv.Atoi(raw[1:])
			if convErr != nil {
				return raw
			}
			return v
		},
	})
}
