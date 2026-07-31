package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Mode represents the serving mode of the application.
type Mode string

const (
	// ModeDirectory serves every supported file inside a directory.
	ModeDirectory Mode = "directory"

	// ModeSingle serves a single file at a user-defined endpoint.
	ModeSingle Mode = "single"
)

// Config holds the validated runtime configuration of the server.
type Config struct {
	// DataDir is the directory containing data files (directory mode).
	DataDir string

	// FilePath is the path to a single data file (single file mode).
	FilePath string

	// Endpoint is the user-defined endpoint for single file mode.
	Endpoint string

	// Host is the HTTP bind host.
	Host string

	// Port is the HTTP bind port.
	Port int

	// Pretty pretty-prints JSON responses.
	Pretty bool
}

var endpointPattern = regexp.MustCompile(`^/[a-zA-Z0-9_-]+$`)

// Mode returns the serving mode derived from the configuration.
func (c *Config) Mode() Mode {
	if c.FilePath != "" {
		return ModeSingle
	}
	return ModeDirectory
}

// Validate checks that the configuration is consistent and usable.
func (c *Config) Validate() error {
	switch {
	case c.DataDir == "" && c.FilePath == "":
		return errors.New("either --data or --file is required")
	case c.DataDir != "" && c.FilePath != "":
		return errors.New("--data and --file cannot be used together")
	}

	if c.DataDir != "" {
		info, err := os.Stat(c.DataDir)
		if err != nil {
			return fmt.Errorf("data directory: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("data directory %q is not a directory", c.DataDir)
		}
	}

	if c.FilePath != "" {
		if c.Endpoint == "" {
			return errors.New("--endpoint is required when using --file")
		}
		if !strings.HasPrefix(c.Endpoint, "/") {
			return errors.New("--endpoint must begin with '/'")
		}
		if !endpointPattern.MatchString(c.Endpoint) {
			return fmt.Errorf("--endpoint %q must match /[a-zA-Z0-9_-]+", c.Endpoint)
		}
		if c.Endpoint == "/health" || c.Endpoint == "/metrics" {
			return fmt.Errorf("--endpoint %q is reserved", c.Endpoint)
		}
		if _, err := os.Stat(c.FilePath); err != nil {
			return fmt.Errorf("data file: %w", err)
		}
	}

	if c.Host == "" {
		return errors.New("--host cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("--port must be between 1 and 65535, got %d", c.Port)
	}
	return nil
}

// Addr returns the host:port address for the HTTP server.
func (c *Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
