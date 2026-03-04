package guessit

import (
	"strings"
	"unicode"
)

// Cleanup normalizes a token by trimming separators and spaces.
func Cleanup(input string) string {
	trimmed := strings.TrimFunc(input, IsSeparator)
	trimmed = strings.ReplaceAll(trimmed, "_", " ")
	trimmed = strings.ReplaceAll(trimmed, ".", " ")
	return strings.Join(strings.Fields(trimmed), " ")
}

// IsSeparator reports whether a rune is a filename separator.
func IsSeparator(r rune) bool {
	return r == '.' || r == '_' || r == '-' || unicode.IsSpace(r)
}

// SepsBefore checks whether previous rune is a separator (or string start).
func SepsBefore(input string, start int) bool {
	if start <= 0 {
		return true
	}
	r, _ := runeAt(input, start-1)
	return IsSeparator(r)
}

// SepsAfter checks whether next rune is a separator (or string end).
func SepsAfter(input string, end int) bool {
	if end >= len(input) {
		return true
	}
	r, _ := runeAt(input, end)
	return IsSeparator(r)
}

func runeAt(s string, idx int) (rune, int) {
	if idx < 0 || idx >= len(s) {
		return 0, 0
	}
	r := rune(s[idx])
	return r, 1
}
