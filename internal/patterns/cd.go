package patterns

import (
	"regexp"
	"strconv"
	"strings"
)

func registerCDPatterns(m *Matcher) {
	reFirstNumber := regexp.MustCompile(`\d+`)
	parseFirstInt := func(raw string) any {
		n := reFirstNumber.FindString(raw)
		if n == "" {
			return raw
		}
		v, err := strconv.Atoi(n)
		if err != nil {
			return raw
		}
		return v
	}

	reOf := regexp.MustCompile(`(?i)(\d+)[ ._-]?of[ ._-]?(\d+)`)

	AddPatternSpec(m, "cd", `(?i)\b(?:cd|disc)[ ._-]?\d{1,2}\b`, PatternSpec{
		Formatter: parseFirstInt,
	})

	AddPatternSpec(m, "cd", `(?i)\bcd[ ._-]?\d{1,2}[ ._-]?of[ ._-]?\d{1,2}\b`, PatternSpec{
		Formatter: parseFirstInt,
	})

	AddPatternSpec(m, "cd_count", `(?i)\b(?:cd[ ._-]?\d{1,2}[ ._-]?of[ ._-]?\d{1,2}|\d{1,2}[ ._-]?cds?)\b`, PatternSpec{
		Formatter: func(raw string) any {
			raw = strings.ToLower(raw)
			if strings.Contains(raw, "of") {
				matches := reOf.FindStringSubmatch(raw)
				if len(matches) == 3 {
					if v, convErr := strconv.Atoi(matches[2]); convErr == nil {
						return v
					}
				}
			}
			n := reFirstNumber.FindString(raw)
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
