package simpledsl

// DslArg represents an argument definition in the DSL.
type DslArg interface {
	GetName() string
	IsRequired() bool
}
