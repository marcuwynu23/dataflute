package tsv

import "github.com/marcuwynu23/dataserve/internal/parser/csv"

// Parser decodes tab-separated values into an array of objects.
type Parser struct {
	*csv.Parser
}

// New returns a tab-delimited TSV parser.
func New() *Parser {
	return &Parser{csv.Delimited('\t', ".tsv")}
}
