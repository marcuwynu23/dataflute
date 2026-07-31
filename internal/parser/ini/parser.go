package ini

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/marcuwynu23/dataserve/internal/parser/scalar"
)

// Parser decodes INI files into a generic object.
type Parser struct{}

// New returns a new INI parser.
func New() *Parser { return &Parser{} }

// Extensions returns the file extensions handled by this parser.
func (*Parser) Extensions() []string { return []string{".ini"} }

// Parse decodes the INI file at path. Keys before any section become
// top-level fields; each [section] becomes a nested object.
func (*Parser) Parse(path string) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	root := make(map[string]any)
	var section map[string]any

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#"):
			continue
		case strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]"):
			name := strings.TrimSpace(line[1 : len(line)-1])
			section = make(map[string]any)
			root[name] = section
		default:
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key := strings.TrimSpace(k)
			if section != nil {
				section[key] = scalar.Value(v)
			} else {
				root[key] = scalar.Value(v)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return root, nil
}
