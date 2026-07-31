package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Parser decodes TOML documents into generic Go values.
type Parser struct{}

// New returns a new TOML parser.
func New() *Parser { return &Parser{} }

// Extensions returns the file extensions handled by this parser.
func (*Parser) Extensions() []string { return []string{".toml"} }

// Parse decodes the TOML document at path.
func (*Parser) Parse(path string) (any, error) {
	var v map[string]any
	if _, err := toml.DecodeFile(path, &v); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return v, nil
}
