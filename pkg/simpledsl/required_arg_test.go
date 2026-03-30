package simpledsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItIsAnErrorToProvideAValueThatIsNotInTheAllowedValues(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1").SetAllowedValues("a", "b", "c"))
	_, err := params.Parse([]string{"d"})
	require.Error(t, err)
	assert.Equal(t, "value \"d\" not allowed for parameter \"arg1\"", err.Error())
}

func TestItIsAnErrorToProvideAValueThatIsNotInTheAllowedValuesForANamedArgument(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1").SetAllowedValues("a", "b", "c"))
	_, err := params.Parse([]string{"arg1:d"})
	require.Error(t, err)
	assert.Equal(t, "value \"d\" not allowed for parameter \"arg1\"", err.Error())
}
