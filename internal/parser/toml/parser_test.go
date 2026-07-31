package toml

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
	assert.Equal(t, []string{".toml"}, New().Extensions())
}

func TestParse(t *testing.T) {
	path := writeFile(t, "app.toml", `
name = "dataflute"
version = "1.0.0"

[server]
host = "0.0.0.0"
port = 8080
pretty = true
`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"name":    "dataflute",
		"version": "1.0.0",
		"server": map[string]any{
			"host":   "0.0.0.0",
			"port":   int64(8080),
			"pretty": true,
		},
	}, got)
}

func TestParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.toml"))
	require.Error(t, err)
}
