package parser

// Parser decodes a data file into a generic Go value.
//
// Implementing a parser and registering it with a Registry is all that
// is required to add support for a new file format; the HTTP server
// requires no changes.
type Parser interface {
	// Extensions returns the file extensions this parser supports,
	// with or without a leading dot (e.g. ".json").
	Extensions() []string

	// Parse reads the file at path and returns its root value.
	Parse(path string) (any, error)
}
