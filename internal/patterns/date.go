package patterns

import (
	"regexp"
	"strconv"
)

func registerDatePatterns(m *Matcher) {
	AddPatternSpec(m, "year", `\b(19\d{2}|20[0-3]\d)\b`, PatternSpec{
		Formatter: func(raw string) any {
			v, convErr := strconv.Atoi(raw)
			if convErr != nil {
				return raw
			}
			return v
		},
	})
	re := regexp.MustCompile(`\.`)
	AddPatternSpec(m, "date", `\b(20\d{2})[.\-](\d{2})[.\-](\d{2})\b`, PatternSpec{
		Formatter: func(raw string) any { return re.ReplaceAllString(raw, "-") },
	})
	AddPatternSpec(m, "date", `\b(\d{2})[.\-](\d{2})[.\-](20\d{2})\b`, PatternSpec{
		Formatter: func(raw string) any {
			// DD-MM-YYYY -> YYYY-MM-DD
			if len(raw) != 10 {
				return raw
			}
			return raw[6:10] + "-" + raw[3:5] + "-" + raw[0:2]
		},
	})
}
