package guessit

// Options holds runtime options for parsing.
type Options struct {
	TypeHint      string
	ExpectedTitle string
	ExpectedGroup string
	Includes      map[string]struct{}
	Excludes      map[string]struct{}
}

// Option mutates Options.
type Option func(*Options)

// NewOptions builds options with defaults and applies overrides.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Includes: make(map[string]struct{}),
		Excludes: make(map[string]struct{}),
	}
	for _, apply := range opts {
		if apply != nil {
			apply(o)
		}
	}
	return o
}

// WithTypeHint sets a type hint (for example: movie or episode).
func WithTypeHint(v string) Option {
	return func(o *Options) { o.TypeHint = v }
}

// WithExpectedTitle sets expected title.
func WithExpectedTitle(v string) Option {
	return func(o *Options) { o.ExpectedTitle = v }
}

// WithExpectedGroup sets expected release group.
func WithExpectedGroup(v string) Option {
	return func(o *Options) { o.ExpectedGroup = v }
}

// WithIncludes sets included properties.
func WithIncludes(values ...string) Option {
	return func(o *Options) {
		for _, v := range values {
			o.Includes[v] = struct{}{}
		}
	}
}

// WithExcludes sets excluded properties.
func WithExcludes(values ...string) Option {
	return func(o *Options) {
		for _, v := range values {
			o.Excludes[v] = struct{}{}
		}
	}
}
