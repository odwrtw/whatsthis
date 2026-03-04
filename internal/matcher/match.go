package matcher

// Match is a detected metadata span in the original input.
type Match struct {
	Name  string
	Value any
	Start int
	End   int
	Raw   string
	Tags  []string
}
