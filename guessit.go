package guessit

import (
	"sync"

	internalpatterns "github.com/odwrtw/go-guessit/internal/patterns"
	internalpostprocess "github.com/odwrtw/go-guessit/internal/postprocess"
)

var (
	defaultMatcher     Matcher
	defaultMatcherOnce sync.Once
)

// GuessIt parses a media filename and returns extracted metadata.
func GuessIt(input string, opts ...Option) Result {
	options := NewOptions(opts...)
	_ = SplitPathParts(input)
	_ = FindGroups(input)
	defaultMatcherOnce.Do(func() {
		registerDefaultPatterns(&defaultMatcher)
	})
	matches := defaultMatcher.Match(input)
	result := Result{}
	for _, m := range matches {
		if !propertyEnabled(m.Name, options) {
			continue
		}
		result.Set(m.Name, m.Value)
	}
	internalpostprocess.PostProcessResult(input, matches, internalpostprocess.Result(result), &internalpostprocess.Options{
		TypeHint:      options.TypeHint,
		ExpectedTitle: options.ExpectedTitle,
		Includes:      options.Includes,
		Excludes:      options.Excludes,
	})
	return result
}

func registerDefaultPatterns(m *Matcher) {
	internalpatterns.RegisterDefault(m)
}

func propertyEnabled(name string, opts *Options) bool {
	if len(opts.Includes) > 0 {
		if _, ok := opts.Includes[name]; !ok {
			return false
		}
	}
	if _, denied := opts.Excludes[name]; denied {
		return false
	}
	return true
}
