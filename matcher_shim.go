package guessit

import internalmatcher "github.com/odwrtw/go-guessit/internal/matcher"

type (
	Match       = internalmatcher.Match
	Matcher     = internalmatcher.Matcher
	Pattern     = internalmatcher.Pattern
	PatternSpec = internalmatcher.PatternSpec
	ValueFunc   = internalmatcher.ValueFunc
	Validator   = internalmatcher.Validator
)

func NewPattern(name, expr string) (Pattern, error) {
	return internalmatcher.NewPattern(name, expr)
}

func AddPatternSpec(m *Matcher, name, expr string, spec PatternSpec) bool {
	return internalmatcher.AddPatternSpec(m, name, expr, spec)
}

func AddConstPattern(m *Matcher, name, expr string, value any) bool {
	return internalmatcher.AddConstPattern(m, name, expr, value)
}
