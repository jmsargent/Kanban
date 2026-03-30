package simpledsl

// SimpleDslArg contains common properties for scalar arguments.
type SimpleDslArg struct {
	Name          string
	Required      bool
	DefaultValue  *string
	AllowedValues []string
	AllowMultiple bool
}

func (a *SimpleDslArg) GetName() string  { return a.Name }
func (a *SimpleDslArg) IsRequired() bool { return a.Required }
