package xml

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
	assert.Equal(t, []string{".xml"}, New().Extensions())
}

func TestParseRepeatedChildren(t *testing.T) {
	path := writeFile(t, "report.xml", `
		<report>
		  <item><id>1</id><name>Q1</name><amount>12000.5</amount></item>
		  <item><id>2</id><name>Q2</name><amount>9800</amount></item>
		</report>`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{
		map[string]any{"id": json.Number("1"), "name": "Q1", "amount": json.Number("12000.5")},
		map[string]any{"id": json.Number("2"), "name": "Q2", "amount": json.Number("9800")},
	}, got)
}

func TestParseAttributesAndText(t *testing.T) {
	path := writeFile(t, "user.xml", `<user id="42">Mark</user>`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"@id": "42", "#text": "Mark"}, got)
}

func TestParseObjectDocument(t *testing.T) {
	path := writeFile(t, "config.xml", `<config><version>1.0.0</version><debug>true</debug></config>`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"version": "1.0.0", "debug": true}, got)
}

func TestParseRepeatedKeysAsArray(t *testing.T) {
	path := writeFile(t, "tags.xml", `<tags><tag>go</tag><tag>rest</tag></tags>`)

	got, err := New().Parse(path)
	require.NoError(t, err)
	assert.Equal(t, []any{"go", "rest"}, got)
}

func TestParseMalformed(t *testing.T) {
	path := writeFile(t, "bad.xml", `<config><version>1.0.0</config>`)

	_, err := New().Parse(path)
	require.Error(t, err)
}

func TestParseMissingFile(t *testing.T) {
	_, err := New().Parse(filepath.Join(t.TempDir(), "nope.xml"))
	require.Error(t, err)
}
