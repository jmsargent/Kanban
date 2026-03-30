package simpledsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalArgumentIsPopulatedByItsDefaultValueIfItIsNotSupplied(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1").SetDefault("default value"))
	values, err := params.Parse([]string{})
	require.NoError(t, err)
	assert.Equal(t, "default value", values.Value("arg1"))
}

func TestDefaultValueCanBeOverridden(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1").SetDefault("default value"))
	values, err := params.Parse([]string{"someOtherValue"})
	require.NoError(t, err)
	assert.Equal(t, "someOtherValue", values.Value("arg1"))
}

func TestDefaultValueIsOverriddenByNamedParameter(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1").SetDefault("default value"))
	values, err := params.Parse([]string{"arg1:someOtherValue"})
	require.NoError(t, err)
	assert.Equal(t, "someOtherValue", values.Value("arg1"))
}

func TestHasValueReturnsFalseForAnOptionalArgumentThatHasNotBeenProvided(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1"))
	values, err := params.Parse([]string{})
	require.NoError(t, err)
	assert.False(t, values.HasValue("arg1"))
}

func TestHasValueReturnsTrueForAnOptionalArgumentThatHasBeenProvided(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1"))
	values, err := params.Parse([]string{"a value"})
	require.NoError(t, err)
	assert.True(t, values.HasValue("arg1"))
}

func TestHasValueReturnsTrueForAnOptionalArgumentThatHasADefaultValue(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1").SetDefault("default"))
	values, err := params.Parse([]string{})
	require.NoError(t, err)
	assert.True(t, values.HasValue("arg1"))
}

func TestItIsAnErrorToProvideAValueThatIsNotInTheAllowedValuesForAnOptionalArgument(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1").SetAllowedValues("a", "b", "c"))
	_, err := params.Parse([]string{"d"})
	require.Error(t, err)
	assert.Equal(t, "value \"d\" not allowed for parameter \"arg1\"", err.Error())
}
