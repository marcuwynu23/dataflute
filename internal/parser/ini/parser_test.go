package ini

import (
	"encoding/json"
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
	assert.Equal(t, []string{".ini"}, New().Extensions())
}

func TestParse(t *testing.T) {
	path := writeFile(t, "settings.ini", `
; comment
[server]
host = 0.0.0.0
port = 8080
pretty = true

[database]
engine = memory
max_connections = 100
`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"server": map[string]any{
			"host":   "0.0.0.0",
			"port":   json.Number("8080"),
			"pretty": true,
		},
		"database": map[string]any{
			"engine":          "memory",
			"max_connections": json.Number("100"),
		},
	}, got)
}

func TestParseTopLevelKeys(t *testing.T) {
	path := writeFile(t, "global.ini", "name = dataserve\nversion = 1.0.0\n")

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"name":    "dataserve",
		"version": "1.0.0",
	}, got)
}

func TestParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.ini"))
	require.Error(t, err)
}
