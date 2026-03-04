package patterns

import "strings"

func registerContainerPatterns(m *Matcher) {
	AddPatternSpec(m, "container", `(?i)\.(mkv|avi|mp4|m4v|mov|wmv|flv|mpg|mpeg|ts|m2ts|ogm|webm)\b`, PatternSpec{
		Formatter: func(raw string) any {
			return strings.ToLower(strings.TrimPrefix(raw, "."))
		},
	})
}
