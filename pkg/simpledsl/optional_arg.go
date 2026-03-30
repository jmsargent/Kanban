package simpledsl

// OptionalArg defines an optional parameter.
type OptionalArg struct {
	*SimpleDslArg
}

// NewOptionalArg creates a new optional argument definition.
func NewOptionalArg(name string) *OptionalArg {
	return &OptionalArg{
		SimpleDslArg: &SimpleDslArg{Name: name, Required: false},
	}
}

// SetDefault sets the default value for the argument if it's not provided.
func (a *OptionalArg) SetDefault(val string) *OptionalArg {
	a.DefaultValue = &val
	return a
}

// SetAllowedValues sets the list of allowed values for the argument.
func (a *OptionalArg) SetAllowedValues(vals ...string) *OptionalArg {
	a.AllowedValues = append([]string{}, vals...)
	return a
}

// SetAllowMultipleValues configures whether the argument can accept multiple values.
func (a *OptionalArg) SetAllowMultipleValues(multiple bool) *OptionalArg {
	a.AllowMultiple = multiple
	return a
}
