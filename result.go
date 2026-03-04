package guessit

// Result is the parsed metadata.
type Result map[string]any

// Get returns a value and whether it exists.
func (r Result) Get(name string) (any, bool) {
	v, ok := r[name]
	return v, ok
}

// Set stores a value.
func (r Result) Set(name string, value any) {
	r[name] = value
}
