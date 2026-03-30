package simpledsl

import (
	"strings"
)

// DslValues holds the parsed results.
type DslValues struct {
	values map[string][]string
	groups map[string][]*DslValues
}

func newDslValues() *DslValues {
	return &DslValues{
		values: make(map[string][]string),
		groups: make(map[string][]*DslValues),
	}
}

// HasValue returns true if the parameter was explicitly provided or has a default.
func (v *DslValues) HasValue(name string) bool {
	_, ok := v.values[name]
	return ok
}

// Value returns the first parsed value for a parameter, or an empty string.
func (v *DslValues) Value(name string) string {
	if vals, ok := v.values[name]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// ValueAsBool returns the parameter's value parsed as a boolean.
func (v *DslValues) ValueAsBool(name string) bool {
	return strings.ToLower(v.Value(name)) == "true"
}

// Values returns all parsed values for a parameter (useful for comma-separated or repeated params).
func (v *DslValues) Values(name string) []string {
	return v.values[name]
}

// Group returns the slice of grouped DslValues instances for a repeating group.
func (v *DslValues) Group(name string) []*DslValues {
	return v.groups[name]
}
