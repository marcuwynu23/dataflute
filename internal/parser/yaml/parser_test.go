package yaml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestExtensions(t *testing.T) {
	assert.Equal(t, []string{".yaml", ".yml"}, New().Extensions())
}

func TestParseObject(t *testing.T) {
	path := writeFile(t, "config.yaml", "service: dataserve\nversion: 1.0.0\n")

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"service": "dataserve",
		"version": "1.0.0",
	}, got)
}

func TestParseNestedAndTyped(t *testing.T) {
	path := writeFile(t, "server.yaml", `
server:
  host: 0.0.0.0
  port: 8080
debug: true
nullish: null
`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"server":  map[string]any{"host": "0.0.0.0", "port": 8080},
		"debug":   true,
		"nullish": nil,
	}, got)
}

func TestParseList(t *testing.T) {
	path := writeFile(t, "list.yaml", "- a\n- b\n")

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, got)
}

func TestParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.yaml"))
	require.Error(t, err)
}
