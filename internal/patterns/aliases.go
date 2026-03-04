package patterns

import internalmatcher "github.com/odwrtw/go-guessit/internal/matcher"

type (
	Matcher     = internalmatcher.Matcher
	PatternSpec = internalmatcher.PatternSpec
)

var (
	AddPatternSpec  = internalmatcher.AddPatternSpec
	AddConstPattern = internalmatcher.AddConstPattern
)
