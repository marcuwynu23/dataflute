package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyQuery_FilterMixedTypes(t *testing.T) {
	// Data built with float64 (not json.Number) and nil values exercises
	// the fallback branches of scalarEqual.
	data := []any{
		map[string]any{"score": 12.5, "tag": nil},
		map[string]any{"score": 9.0, "tag": nil},
		map[string]any{"score": 12.5, "tag": "x"},
	}

	out, err := ApplyQuery(data, values("score=12.5"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 2)

	out, err = ApplyQuery(data, values("tag=null"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 2)
}

func TestApplyQuery_SortMissingField(t *testing.T) {
	data := []any{
		map[string]any{"name": "b"},
		map[string]any{"name": "a"},
		map[string]any{"other": 1},
	}
	out, err := ApplyQuery(data, values("sort=name"))
	require.NoError(t, err)
	items := out.([]any)

	// Elements without the sort key sort before those with it.
	assert.Equal(t, 1, items[0].(map[string]any)["other"])
}

func TestApplyQuery_SortMixedTypes(t *testing.T) {
	data := []any{
		map[string]any{"v": 1},
		map[string]any{"v": "x"},
	}
	out, err := ApplyQuery(data, values("sort=v"))
	require.NoError(t, err)
	items := out.([]any)
	assert.Equal(t, 1, items[0].(map[string]any)["v"])
}

func TestApplyQuery_SortInvalidKey(t *testing.T) {
	_, err := ApplyQuery(users(t), values("sort=-"))
	require.Error(t, err)
}

func TestApplyQuery_MultipleValuesForFilterKey(t *testing.T) {
	// ?country=PH&country=US uses the first value (PH).
	out, err := ApplyQuery(users(t), map[string][]string{"country": {"PH", "US"}})
	require.NoError(t, err)
	assert.Len(t, out.([]any), 2)
}

func TestContainsTermNumber(t *testing.T) {
	assert.True(t, containsTerm(map[string]any{"id": json.Number("42")}, "42"))
	assert.True(t, containsTerm(map[string]any{"nested": []any{map[string]any{"k": "Hello"}}}, "hello"))
	assert.False(t, containsTerm(map[string]any{"n": 1.5}, "hello"))
}
