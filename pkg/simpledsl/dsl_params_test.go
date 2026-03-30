package simpledsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsingASingleRequiredValue(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"))
	values, err := params.Parse([]string{"a value"})
	require.NoError(t, err)
	assert.Equal(t, "a value", values.Value("arg1"))
}

func TestParsingASingleRequiredValueWithWhitespace(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"))
	values, err := params.Parse([]string{" a value "})
	require.NoError(t, err)
	assert.Equal(t, "a value", values.Value("arg1"))
}

func TestShouldNotAllowAPositionalArgumentToFollowANamedArgument(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"), NewRequiredArg("arg2"))
	_, err := params.Parse([]string{"arg1:a value", "another value"})
	require.Error(t, err)
	assert.Equal(t, "unexpected positional argument: \"another value\"", err.Error())
}

func TestParsingTwoRequiredValues(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"), NewRequiredArg("arg2"))
	values, err := params.Parse([]string{"value1", "value2"})
	require.NoError(t, err)
	assert.Equal(t, "value1", values.Value("arg1"))
	assert.Equal(t, "value2", values.Value("arg2"))
}

func TestParsingTwoRequiredValuesProvidedInDifferentOrder(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"), NewRequiredArg("arg2"))
	values, err := params.Parse([]string{"arg2:value2", "arg1:value1"})
	require.NoError(t, err)
	assert.Equal(t, "value1", values.Value("arg1"))
	assert.Equal(t, "value2", values.Value("arg2"))
}

func TestParsingOneOfTwoRequiredValues(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"), NewRequiredArg("arg2"))
	_, err := params.Parse([]string{"value1"})
	require.Error(t, err)
	assert.Equal(t, "missing required parameter: \"arg2\"", err.Error())
}

func TestParsingOneOfTwoRequiredValuesAsNamed(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"), NewRequiredArg("arg2"))
	_, err := params.Parse([]string{"arg2:value2"})
	require.Error(t, err)
	assert.Equal(t, "missing required parameter: \"arg1\"", err.Error())
}

func TestParsingAnOptionalValue(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1"))
	values, err := params.Parse([]string{"a value"})
	require.NoError(t, err)
	assert.Equal(t, "a value", values.Value("arg1"))
}

func TestParsingAnOptionalValueProvidedAsANamedArgument(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1"))
	values, err := params.Parse([]string{"arg1:a value"})
	require.NoError(t, err)
	assert.Equal(t, "a value", values.Value("arg1"))
}

func TestItIsAnErrorToProvideAnOptionalArgumentThatIsNotKnown(t *testing.T) {
	params := NewDslParams(NewOptionalArg("arg1"))
	_, err := params.Parse([]string{"arg2:a value"})
	require.Error(t, err)
	assert.Equal(t, "unknown parameter: \"arg2\"", err.Error())
}

func TestItIsAnErrorToProvideAnUnknownArgument(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"))
	_, err := params.Parse([]string{"arg1:a value", "arg2:another value"})
	require.Error(t, err)
	assert.Equal(t, "unknown parameter: \"arg2\"", err.Error())
}

func TestShouldBeAbleToParseBooleanValues(t *testing.T) {
	params := NewDslParams(NewOptionalArg("is-set"))
	values, err := params.Parse([]string{"is-set:true"})
	require.NoError(t, err)
	assert.True(t, values.ValueAsBool("is-set"))

	values, err = params.Parse([]string{"is-set:false"})
	require.NoError(t, err)
	assert.False(t, values.ValueAsBool("is-set"))
}

func TestShouldBeAbleToParseMultipleValuesFromACommaSeparatedList(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1").SetAllowMultipleValues(true))
	values, err := params.Parse([]string{"a,b,c"})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, values.Values("arg1"))
}

func TestShouldBeAbleToParseMultipleValuesFromACommaSeparatedListWithSpaces(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1").SetAllowMultipleValues(true))
	values, err := params.Parse([]string{"a, b, c"})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, values.Values("arg1"))
}

func TestItIsAnErrorToProvideMultipleValuesToAnArgumentThatDoesNotSupportIt(t *testing.T) {
	params := NewDslParams(NewRequiredArg("arg1"))
	_, err := params.Parse([]string{"arg1:a,b"})
	require.Error(t, err)
	assert.Equal(t, "parameter \"arg1\" does not allow multiple values", err.Error())
}
