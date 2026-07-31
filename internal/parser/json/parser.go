package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Parser decodes JSON files using json.Decoder.UseNumber so that
// numeric values round-trip without precision loss.
type Parser struct{}

// New returns a new JSON parser.
func New() *Parser { return &Parser{} }

// Extensions returns the file extensions handled by this parser.
func (*Parser) Extensions() []string { return []string{".json"} }

// Parse decodes the JSON document at path into a generic value.
// The root value may be an array, object, or any JSON primitive.
func (*Parser) Parse(path string) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("decode %s: unexpected content after top-level JSON value", path)
	}
	return v, nil
}
