package matcher

import (
	"regexp"
	"sort"
)

// ValueFunc maps a raw regex hit to a property value.
type ValueFunc func(raw string) any

// Validator decides if a match is valid in context.
type Validator func(input string, start, end int) bool

// Pattern defines a named regex detector.
type Pattern struct {
	Name      string
	Expr      string
	Regex     *regexp.Regexp
	Formatter ValueFunc
	Validator Validator
	Tags      []string
	Priority  int
}

// PatternSpec carries optional pattern behavior knobs for AddPatternSpec.
type PatternSpec struct {
	Formatter ValueFunc
	Validator Validator
	Tags      []string
	Priority  int
}

// NewPattern compiles a regex pattern.
func NewPattern(name, expr string) (Pattern, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return Pattern{}, err
	}
	return Pattern{Name: name, Expr: expr, Regex: re}, nil
}

// AddPatternSpec compiles and registers a pattern with optional behavior.
func AddPatternSpec(m *Matcher, name, expr string, spec PatternSpec) bool {
	p, err := NewPattern(name, expr)
	if err != nil {
		return false
	}
	p.Formatter = spec.Formatter
	p.Validator = spec.Validator
	p.Tags = append([]string(nil), spec.Tags...)
	p.Priority = spec.Priority
	m.AddPattern(p)
	return true
}

// AddConstPattern compiles and registers a pattern with a constant value.
func AddConstPattern(m *Matcher, name, expr string, value any) bool {
	return AddPatternSpec(m, name, expr, PatternSpec{
		Formatter: func(string) any { return value },
	})
}

// Matcher stores property patterns and executes them.
type Matcher struct {
	patterns []Pattern
}

// AddPattern registers a pattern.
func (m *Matcher) AddPattern(p Pattern) {
	m.patterns = append(m.patterns, p)
}

// Match finds matches and applies simple overlap conflict resolution.
func (m *Matcher) Match(input string) []Match {
	var all []Match
	for _, p := range m.patterns {
		for _, span := range p.Regex.FindAllStringIndex(input, -1) {
			start, end := span[0], span[1]
			if p.Validator != nil && !p.Validator(input, start, end) {
				continue
			}
			raw := input[start:end]
			value := any(raw)
			if p.Formatter != nil {
				value = p.Formatter(raw)
			}
			all = append(all, Match{
				Name:  p.Name,
				Value: value,
				Start: start,
				End:   end,
				Raw:   raw,
				Tags:  append([]string(nil), p.Tags...),
			})
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Start != all[j].Start {
			return all[i].Start < all[j].Start
		}
		li := all[i].End - all[i].Start
		lj := all[j].End - all[j].Start
		return li > lj
	})
	return resolveConflicts(all)
}

func resolveConflicts(matches []Match) []Match {
	var out []Match
	for _, m := range matches {
		overlap := false
		for _, kept := range out {
			if m.Start < kept.End && kept.Start < m.End {
				overlap = true
				break
			}
		}
		if !overlap {
			out = append(out, m)
		}
	}
	return out
}
