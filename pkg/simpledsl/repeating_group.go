package simpledsl

// RepeatingGroup defines a repeating set of parameters.
type RepeatingGroup struct {
	Name string
	Args []DslArg
}

// NewRepeatingGroup creates a new repeating group definition.
func NewRepeatingGroup(name string, args ...DslArg) *RepeatingGroup {
	return &RepeatingGroup{Name: name, Args: args}
}

// GetName returns the name of the group.
func (g *RepeatingGroup) GetName() string { return g.Name }

// IsRequired indicates that a group itself is not a required parameter in the same way a scalar is.
// Required-ness is handled for the arguments *within* the group.
func (g *RepeatingGroup) IsRequired() bool { return false }
