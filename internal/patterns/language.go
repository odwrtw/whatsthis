package patterns

import "strings"

func registerLanguagePatterns(m *Matcher) {
	AddConstPattern(m, "language", `(?i)\b(english|eng|en)\b`, "English")
	AddConstPattern(m, "language", `(?i)\b(french|francais|fra|fr)\b`, "French")
	AddConstPattern(m, "language", `(?i)\b(spanish|esp|es)\b`, "Spanish")
	AddConstPattern(m, "language", `(?i)\b(german|deu|de)\b`, "German")
	AddConstPattern(m, "language", `(?i)\bcatalan\b`, "Catalan")

	AddPatternSpec(m, "subtitle_language", `(?i)\b(?:sub|subs|subtitle|vostfr)[ ._-]?(en|eng|english|fr|fre?nch|spanish|german|it|ita)\b`, PatternSpec{
		Formatter: func(raw string) any {
			switch {
			case containsI(raw, "eng"), containsI(raw, " en"):
				return "English"
			case containsI(raw, "fre"), containsI(raw, "fr"), containsI(raw, "vostfr"):
				return "French"
			case containsI(raw, "span"):
				return "Spanish"
			case containsI(raw, "ger"):
				return "German"
			case containsI(raw, "ita"), containsI(raw, "it"):
				return "it"
			default:
				return raw
			}
		},
	})
}

func containsI(s, part string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(part))
}
