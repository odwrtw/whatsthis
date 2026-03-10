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
	parsed := internalpostprocess.Result{}
	for _, m := range matches {
		if !propertyEnabled(m.Name, options) {
			continue
		}
		setParsedField(parsed, m.Name, m.Value)
	}
	internalpostprocess.PostProcessResult(input, matches, parsed, &internalpostprocess.Options{
		TypeHint:      options.TypeHint,
		ExpectedTitle: options.ExpectedTitle,
		Includes:      options.Includes,
		Excludes:      options.Excludes,
	})
	return resultFromParsed(parsed)
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

func setParsedField(parsed internalpostprocess.Result, name string, value any) {
	switch name {
	case "type", "title", "screen_size", "release_group", "audio_codec", "video_codec", "container", "mimetype":
		if s, ok := value.(string); ok {
			parsed[name] = s
		}
	case "episode", "season", "year":
		if n, ok := asInt(value); ok {
			parsed[name] = n
		}
	default:
		parsed[name] = value
	}
}

func resultFromParsed(parsed internalpostprocess.Result) Result {
	return Result{
		Type:         stringValue(parsed["type"]),
		Title:        stringValue(parsed["title"]),
		Episode:      intValue(parsed["episode"]),
		Season:       intValue(parsed["season"]),
		Year:         intValue(parsed["year"]),
		ScreenSize:   stringValue(parsed["screen_size"]),
		ReleaseGroup: stringValue(parsed["release_group"]),
		AudioCodec:   stringValue(parsed["audio_codec"]),
		VideoCodec:   stringValue(parsed["video_codec"]),
		Container:    stringValue(parsed["container"]),
		MIMEType:     stringValue(parsed["mimetype"]),
	}
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func intValue(v any) int {
	n, _ := asInt(v)
	return n
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}
