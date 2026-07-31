package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser decodes YAML documents into generic Go values.
type Parser struct{}

// New returns a new YAML parser.
func New() *Parser { return &Parser{} }

// Extensions returns the file extensions handled by this parser.
func (*Parser) Extensions() []string { return []string{".yaml", ".yml"} }

// Parse decodes the YAML document at path.
func (*Parser) Parse(path string) (any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return v, nil
}
