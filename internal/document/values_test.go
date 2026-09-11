package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValueReturnsOnlyStrings(t *testing.T) {
	assert.Equal(t, "name", StringValue("name"))
	assert.Empty(t, StringValue(42))
}

func TestNumberValueAndOptionalNumberPreserveCoercionContract(t *testing.T) {
	assert.Equal(t, 42.5, NumberValue(float64(42.5)))
	assert.Zero(t, NumberValue("42.5"))

	value := OptionalNumber(float64(42.5))
	if assert.NotNil(t, value) {
		assert.Equal(t, 42.5, *value)
	}
	assert.Nil(t, OptionalNumber("42.5"))
}

func TestNumberSliceAndMapValueRejectWrongShapes(t *testing.T) {
	assert.Equal(t, []float64{1, 0, 3}, NumberSlice([]any{float64(1), "two", float64(3)}))
	assert.Nil(t, NumberSlice("not a slice"))

	object := map[string]any{"id": "node"}
	assert.Equal(t, object, MapValue(object))
	assert.Nil(t, MapValue("not a map"))
}
