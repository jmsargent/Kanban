package simpledsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBeAbleToParseARepeatingGroupOfArguments(t *testing.T) {
	params := NewDslParams(NewRepeatingGroup("group", NewRequiredArg("param")))
	values, err := params.Parse([]string{"group:value1", "group:value2"})
	require.NoError(t, err)

	groups := values.Group("group")
	require.Len(t, groups, 2)
	assert.Equal(t, "value1", groups[0].Value("param"))
	assert.Equal(t, "value2", groups[1].Value("param"))
}

func TestShouldBeAbleToParseARepeatingGroupOfArgumentsWithNamedParams(t *testing.T) {
	params := NewDslParams(
		NewRepeatingGroup("group",
			NewRequiredArg("param1"),
			NewOptionalArg("param2"),
		),
	)
	values, err := params.Parse([]string{
		"group:value1", "param2:value2",
		"group:value3", "param2:value4",
	})
	require.NoError(t, err)

	groups := values.Group("group")
	require.Len(t, groups, 2)
	assert.Equal(t, "value1", groups[0].Value("param1"))
	assert.Equal(t, "value2", groups[0].Value("param2"))
	assert.Equal(t, "value3", groups[1].Value("param1"))
	assert.Equal(t, "value4", groups[1].Value("param2"))
}

func TestItIsAnErrorToProvideAnUnknownParameterToAGroup(t *testing.T) {
	params := NewDslParams(NewRepeatingGroup("group", NewRequiredArg("param")))
	_, err := params.Parse([]string{"group:value1", "unknown:value"})
	require.Error(t, err)
	assert.Equal(t, "unknown parameter: \"unknown\"", err.Error())
}
