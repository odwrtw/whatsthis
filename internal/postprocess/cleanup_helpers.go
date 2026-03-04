package postprocess

import (
	"strings"
	"unicode"
)

func Cleanup(input string) string {
	trimmed := strings.TrimFunc(input, IsSeparator)
	trimmed = strings.ReplaceAll(trimmed, "_", " ")
	trimmed = strings.ReplaceAll(trimmed, ".", " ")
	return strings.Join(strings.Fields(trimmed), " ")
}

func IsSeparator(r rune) bool {
	return r == '.' || r == '_' || r == '-' || unicode.IsSpace(r)
}
