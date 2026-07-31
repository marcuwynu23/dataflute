package json

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestParser_Extensions(t *testing.T) {
	p := New()
	assert.Equal(t, []string{".json"}, p.Extensions())
}

func TestParser_ParseArray(t *testing.T) {
	path := writeTempFile(t, `[{"id":1},{"id":2}]`)
	v, err := New().Parse(path)
	require.NoError(t, err)

	arr, ok := v.([]any)
	require.True(t, ok)
	require.Len(t, arr, 2)

	obj, ok := arr[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, json.Number("1"), obj["id"])
}

func TestParser_ParsePrimitives(t *testing.T) {
	cases := map[string]any{
		`"hello"`: "hello",
		`42`:      json.Number("42"),
		`true`:    true,
		`null`:    nil,
	}
	for input, want := range cases {
		path := writeTempFile(t, input)
		v, err := New().Parse(path)
		require.NoError(t, err)
		assert.Equal(t, want, v)
	}
}

func TestParser_ParseInvalid(t *testing.T) {
	path := writeTempFile(t, `{"broken":`)
	_, err := New().Parse(path)
	require.Error(t, err)
}

func TestParser_ParseTrailingData(t *testing.T) {
	path := writeTempFile(t, `{"a":1} {"b":2}`)
	_, err := New().Parse(path)
	require.Error(t, err)
}

func TestParser_ParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.json"))
	require.Error(t, err)
}
