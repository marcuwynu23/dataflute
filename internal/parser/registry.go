package parser

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Registry maps file extensions to their parsers.
//
// A Registry is safe for concurrent use. Parsers are keyed by their
// lowercase, dot-prefixed extension, so a parser is selected
// automatically based on a file's extension.
type Registry struct {
	mu      sync.RWMutex
	parsers map[string]Parser
}

// NewRegistry returns an empty parser registry.
func NewRegistry() *Registry {
	return &Registry{parsers: make(map[string]Parser)}
}

// Register adds a parser to the registry under each of its extensions.
// Registering an extension twice overwrites the previous parser.
func (r *Registry) Register(p Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ext := range p.Extensions() {
		r.parsers[normalizeExtension(ext)] = p
	}
}

// ParserFor returns the parser registered for the given file path,
// selected by file extension. The second return value reports whether a
// parser was found.
func (r *Registry) ParserFor(path string) (Parser, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.parsers[normalizeExtension(filepath.Ext(path))]
	return p, ok
}

// Extensions returns the sorted list of supported extensions.
func (r *Registry) Extensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exts := make([]string, 0, len(r.parsers))
	for ext := range r.parsers {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

func normalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
