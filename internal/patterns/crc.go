package patterns

import "strings"

func registerCRC32Patterns(m *Matcher) {
	AddPatternSpec(m, "crc32", `(?i)[\[\(]([0-9a-f]{8})[\]\)]`, PatternSpec{
		Formatter: func(raw string) any {
			raw = strings.TrimPrefix(raw, "[")
			raw = strings.TrimPrefix(raw, "(")
			raw = strings.TrimSuffix(raw, "]")
			raw = strings.TrimSuffix(raw, ")")
			return strings.ToUpper(raw)
		},
	})
}
