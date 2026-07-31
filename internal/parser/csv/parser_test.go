package csv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensions(t *testing.T) {
	assert.Equal(t, []string{".csv"}, New().Extensions())
}

func TestParse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	require.NoError(t, os.WriteFile(path, []byte(
		"id,name,active,salary\n1,Alice,true,95000\n2,Bob,false,\n"), 0o644))

	got, err := New().Parse(path)
	require.NoError(t, err)

	want := []any{
		map[string]any{"id": json.Number("1"), "name": "Alice", "active": true, "salary": json.Number("95000")},
		map[string]any{"id": json.Number("2"), "name": "Bob", "active": false, "salary": nil},
	}
	assert.Equal(t, want, got)
}

func TestParseEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{}, got)
}

func TestParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.csv"))
	require.Error(t, err)
}

func TestParseSkipsEmptyHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	require.NoError(t, os.WriteFile(path, []byte(
		"id,,name\n1,x,Alice\n"), 0o644))

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{map[string]any{"id": json.Number("1"), "name": "Alice"}}, got)
}
