package simpledsl

import (
	"fmt"
	"strings"
)

// DslParams is the parser mechanism for evaluating string inputs.
type DslParams struct {
	args []DslArg
}

// NewDslParams creates a new parser instance with the defined argument structure.
func NewDslParams(args ...DslArg) *DslParams {
	return &DslParams{args: args}
}

// Parse interprets incoming string arguments against the defined structure.
func (p *DslParams) Parse(input []string) (*DslValues, error) {
	root := newDslValues()

	var allArgs []DslArg
	topLevelMap := make(map[string]DslArg)
	groupMap := make(map[string]*RepeatingGroup)

	for _, arg := range p.args {
		topLevelMap[arg.GetName()] = arg
		if _, ok := arg.(*RepeatingGroup); !ok {
			allArgs = append(allArgs, arg)
		}
		if rg, ok := arg.(*RepeatingGroup); ok {
			groupMap[rg.GetName()] = rg
		}
	}

	argIdx := 0
	var activeGroup *RepeatingGroup
	var activeGroupValues *DslValues
	var activeGroupMap map[string]DslArg
	namedArgsEncountered := false

	for _, s := range input {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		var name, val string
		if sepIdx := strings.IndexAny(s, ":="); sepIdx != -1 {
			name = strings.TrimSpace(s[:sepIdx])
			val = strings.TrimSpace(s[sepIdx+1:])
			namedArgsEncountered = true
		} else {
			if namedArgsEncountered {
				return nil, fmt.Errorf("unexpected positional argument: %q", s)
			}
			val = s
		}

		if name == "" {
			// Positional argument
			if argIdx < len(allArgs) {
				arg := allArgs[argIdx]
				if err := addValueForDslArg(root, arg, val); err != nil {
					return nil, err
				}
				// Only advance the positional index if the arg is not multi-valued
				simpleArg, isSimple := getSimpleDslArg(arg)
				if !isSimple || !simpleArg.AllowMultiple {
					argIdx++
				}
			} else {
				return nil, fmt.Errorf("unexpected positional argument: %q", s)
			}
			continue
		}

		// Is it a group start marker?
		if rg, ok := groupMap[name]; ok {
			activeGroup = rg
			activeGroupValues = newDslValues()
			root.groups[name] = append(root.groups[name], activeGroupValues)

			activeGroupMap = make(map[string]DslArg)
			for _, a := range rg.Args {
				activeGroupMap[a.GetName()] = a
			}

			if val != "" && len(rg.Args) > 0 {
				if err := addValueForDslArg(activeGroupValues, rg.Args[0], val); err != nil {
					return nil, err
				}
			}
			continue
		}

		// Is it a parameter for the currently active repeating group?
		if activeGroup != nil {
			if ga, ok := activeGroupMap[name]; ok {
				if err := addValueForDslArg(activeGroupValues, ga, val); err != nil {
					return nil, err
				}
				continue
			}
		}

		// Is it a standard top-level parameter?
		if ta, ok := topLevelMap[name]; ok {
			activeGroup = nil
			activeGroupValues = nil
			activeGroupMap = nil

			if err := addValueForDslArg(root, ta, val); err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("unknown parameter: %q", name)
	}

	if err := validateValues(root, topLevelMap); err != nil {
		return nil, err
	}

	for groupName, rg := range groupMap {
		groupInstances := root.groups[groupName]
		groupArgMap := make(map[string]DslArg)
		for _, a := range rg.Args {
			groupArgMap[a.GetName()] = a
		}
		for _, instanceValues := range groupInstances {
			if err := validateValues(instanceValues, groupArgMap); err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

func getSimpleDslArg(arg DslArg) (*SimpleDslArg, bool) {
	switch a := arg.(type) {
	case *RequiredArg:
		return a.SimpleDslArg, true
	case *OptionalArg:
		return a.SimpleDslArg, true
	}
	return nil, false
}

// addValueForDslArg identifies the underlying SimpleDslArg to assign the raw parsed string
func addValueForDslArg(values *DslValues, arg DslArg, val string) error {
	switch a := arg.(type) {
	case *RequiredArg:
		return addValue(values, a.SimpleDslArg, val)
	case *OptionalArg:
		return addValue(values, a.SimpleDslArg, val)
	}
	return fmt.Errorf("unsupported argument type for %q", arg.GetName())
}

func addValue(values *DslValues, arg *SimpleDslArg, val string) error {
	if !arg.AllowMultiple {
		if strings.Contains(val, ",") {
			return fmt.Errorf("parameter %q does not allow multiple values", arg.Name)
		}
		if _, exists := values.values[arg.Name]; exists {
			return fmt.Errorf("parameter %q does not allow multiple values", arg.Name)
		}
		values.values[arg.Name] = []string{val}
		return nil
	}

	for _, p := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			values.values[arg.Name] = append(values.values[arg.Name], trimmed)
		}
	}
	return nil
}

func validateValues(values *DslValues, argMap map[string]DslArg) error {
	for _, arg := range argMap {
		var sda *SimpleDslArg
		switch a := arg.(type) {
		case *RequiredArg:
			sda = a.SimpleDslArg
		case *OptionalArg:
			sda = a.SimpleDslArg
		case *RepeatingGroup:
			// No validation needed for the group container itself at this level
		}

		if sda == nil {
			continue
		}

		vals := values.values[sda.Name]
		if len(vals) == 0 {
			if sda.Required {
				return fmt.Errorf("missing required parameter: %q", sda.Name)
			}
			if sda.DefaultValue != nil {
				values.values[sda.Name] = []string{*sda.DefaultValue}
				vals = values.values[sda.Name]
			}
		}

		if len(vals) > 0 && len(sda.AllowedValues) > 0 {
			allowedMap := make(map[string]bool)
			for _, a := range sda.AllowedValues {
				allowedMap[a] = true
			}
			for _, v := range vals {
				if !allowedMap[v] {
					return fmt.Errorf("value %q not allowed for parameter %q", v, sda.Name)
				}
			}
		}
	}
	return nil
}
