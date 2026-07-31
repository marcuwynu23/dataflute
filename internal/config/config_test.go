package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Mode(t *testing.T) {
	assert.Equal(t, ModeDirectory, (&Config{DataDir: "./data"}).Mode())
	assert.Equal(t, ModeSingle, (&Config{FilePath: "./users.json"}).Mode())
}

func TestConfig_Validate_RequiresDataOrFile(t *testing.T) {
	err := (&Config{Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--data or --file")
}

func TestConfig_Validate_RejectsBoth(t *testing.T) {
	dir := t.TempDir()
	err := (&Config{DataDir: dir, FilePath: "x.json", Endpoint: "/x", Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be used together")
}

func TestConfig_Validate_DirectoryMode(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, (&Config{DataDir: dir, Host: "0.0.0.0", Port: 8080}).Validate())
}

func TestConfig_Validate_DataDirMustExist(t *testing.T) {
	err := (&Config{DataDir: filepath.Join(t.TempDir(), "missing"), Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
}

func TestConfig_Validate_DataPathIsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "x.json")
	require.NoError(t, os.WriteFile(f, []byte("[]"), 0o644))
	err := (&Config{DataDir: f, Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
}

func TestConfig_Validate_SingleFileMode(t *testing.T) {
	f := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(f, []byte("[]"), 0o644))

	cfg := &Config{FilePath: f, Endpoint: "/users", Host: "0.0.0.0", Port: 8080}
	require.NoError(t, cfg.Validate())
}

func TestConfig_Validate_EndpointRequired(t *testing.T) {
	f := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(f, []byte("[]"), 0o644))

	err := (&Config{FilePath: f, Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--endpoint")
}

func TestConfig_Validate_EndpointMustStartWithSlash(t *testing.T) {
	f := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(f, []byte("[]"), 0o644))

	err := (&Config{FilePath: f, Endpoint: "users", Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "begin with '/'")
}

func TestConfig_Validate_EndpointReserved(t *testing.T) {
	f := filepath.Join(t.TempDir(), "users.json")
	require.NoError(t, os.WriteFile(f, []byte("[]"), 0o644))

	err := (&Config{FilePath: f, Endpoint: "/health", Host: "0.0.0.0", Port: 8080}).Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestConfig_Validate_PortRange(t *testing.T) {
	dir := t.TempDir()
	err := (&Config{DataDir: dir, Host: "0.0.0.0", Port: 0}).Validate()
	require.Error(t, err)

	err = (&Config{DataDir: dir, Host: "0.0.0.0", Port: 70000}).Validate()
	require.Error(t, err)
}

func TestConfig_Addr(t *testing.T) {
	cfg := &Config{Host: "0.0.0.0", Port: 8080}
	assert.Equal(t, "0.0.0.0:8080", cfg.Addr())
}
