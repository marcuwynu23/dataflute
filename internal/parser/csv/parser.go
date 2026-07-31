package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/marcuwynu23/dataflute/internal/parser/scalar"
)

// Parser decodes delimiter-separated values into an array of objects,
// using the first row as field names.
type Parser struct {
	delimiter  rune
	extensions []string
}

// New returns a comma-delimited CSV parser.
func New() *Parser {
	return Delimited(',', ".csv")
}

// Delimited returns a parser for a custom delimiter and extensions.
func Delimited(delimiter rune, extensions ...string) *Parser {
	return &Parser{delimiter: delimiter, extensions: extensions}
}

// Extensions returns the file extensions handled by this parser.
func (p *Parser) Extensions() []string { return p.extensions }

// Parse decodes the delimited file at path into an array of objects.
func (p *Parser) Parse(path string) (any, error) {
	return ParseDelimited(path, p.delimiter)
}

// ParseDelimited decodes a delimited file into an array of objects.
func ParseDelimited(path string, delimiter rune) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = delimiter
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(records) == 0 {
		return []any{}, nil
	}

	header := records[0]
	rows := make([]any, 0, len(records)-1)
	for _, rec := range records[1:] {
		row := make(map[string]any, len(header))
		for i, name := range header {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if i < len(rec) {
				if strings.TrimSpace(rec[i]) == "" {
					row[name] = nil
				} else {
					row[name] = scalar.Value(rec[i])
				}
			} else {
				row[name] = nil
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
