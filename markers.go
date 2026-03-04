package guessit

import "strings"

// GroupMarker is a bracketed span marker.
type GroupMarker struct {
	Start int
	End   int
	Open  rune
	Close rune
}

// SplitPathParts splits a path-like input into non-empty parts.
func SplitPathParts(input string) []string {
	f := strings.FieldsFunc(input, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	out := make([]string, 0, len(f))
	for _, p := range f {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// FindGroups finds bracketed spans in input.
func FindGroups(input string) []GroupMarker {
	type openItem struct {
		pos   int
		open  rune
		close rune
	}
	var out []GroupMarker
	var stack []openItem

	for i, r := range input {
		switch r {
		case '(':
			stack = append(stack, openItem{pos: i, open: '(', close: ')'})
		case '[':
			stack = append(stack, openItem{pos: i, open: '[', close: ']'})
		case '{':
			stack = append(stack, openItem{pos: i, open: '{', close: '}'})
		case ')', ']', '}':
			if len(stack) == 0 {
				continue
			}
			top := stack[len(stack)-1]
			if top.close != r {
				continue
			}
			stack = stack[:len(stack)-1]
			out = append(out, GroupMarker{
				Start: top.pos,
				End:   i + 1,
				Open:  top.open,
				Close: top.close,
			})
		}
	}
	return out
}
