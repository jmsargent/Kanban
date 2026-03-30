package simpledsl

// RequiredArg defines a mandatory parameter.
type RequiredArg struct {
	*SimpleDslArg
}

// NewRequiredArg creates a new required argument definition.
func NewRequiredArg(name string) *RequiredArg {
	return &RequiredArg{
		SimpleDslArg: &SimpleDslArg{Name: name, Required: true},
	}
}

// SetAllowedValues sets the list of allowed values for the argument.
func (a *RequiredArg) SetAllowedValues(vals ...string) *RequiredArg {
	a.AllowedValues = append([]string{}, vals...)
	return a
}

// SetAllowMultipleValues configures whether the argument can accept multiple values.
func (a *RequiredArg) SetAllowMultipleValues(multiple bool) *RequiredArg {
	a.AllowMultiple = multiple
	return a
}
