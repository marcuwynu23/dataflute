package loader

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcuwynu23/dataserve/internal/parser"
	jsonparser "github.com/marcuwynu23/dataserve/internal/parser/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLoader() *Loader {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())
	return New(reg, logger)
}

func TestLoader_LoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	d, err := newTestLoader().LoadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "users", d.Name)
	assert.Equal(t, "json", d.Format)
	assert.NotEmpty(t, d.Data)
}

func TestLoader_LoadFileUnsupported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.xyz")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))

	_, err := newTestLoader().LoadFile(path)
	require.Error(t, err)
}

func TestLoader_LoadFileParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(path, []byte("{"), 0o644))

	_, err := newTestLoader().LoadFile(path)
	require.Error(t, err)
}

func TestLoader_LoadDirectory(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.json"), []byte(`[]`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "products.json"), []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{`), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o755))

	datasets, errs := newTestLoader().LoadDirectory(dir)
	require.Len(t, datasets, 2)
	require.Len(t, errs, 1)

	names := []string{datasets[0].Name, datasets[1].Name}
	assert.ElementsMatch(t, []string{"users", "products"}, names)
}

func TestLoader_LoadDirectoryMissing(t *testing.T) {
	datasets, errs := newTestLoader().LoadDirectory(filepath.Join(t.TempDir(), "missing"))
	assert.Empty(t, datasets)
	require.NotEmpty(t, errs)
}
