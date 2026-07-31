package tsv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensions(t *testing.T) {
	assert.Equal(t, []string{".tsv"}, New().Extensions())
}

func TestParse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.tsv")
	require.NoError(t, os.WriteFile(path, []byte(
		"code\tname\nPH\tPhilippines\nUS\tUnited States\n"), 0o644))

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{
		map[string]any{"code": "PH", "name": "Philippines"},
		map[string]any{"code": "US", "name": "United States"},
	}, got)
}

func TestParseIgnoresCommas(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.tsv")
	require.NoError(t, os.WriteFile(path, []byte(
		"name\tcount\nApples, Oranges\t2\n"), 0o644))

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{map[string]any{"name": "Apples, Oranges", "count": json.Number("2")}}, got)
}
