package patterns

import "strconv"

func registerFilmPatterns(m *Matcher) {
	AddPatternSpec(m, "film", `(?i)\bf(\d{1,2})\b`, PatternSpec{
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
