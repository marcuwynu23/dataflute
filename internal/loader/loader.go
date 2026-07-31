package loader

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcuwynu23/dataflute/internal/models"
	"github.com/marcuwynu23/dataflute/internal/parser"
)

// Loader loads data files into in-memory datasets using a parser registry.
type Loader struct {
	registry *parser.Registry
	logger   *slog.Logger
}

// New returns a Loader backed by the given parser registry.
func New(registry *parser.Registry, logger *slog.Logger) *Loader {
	return &Loader{registry: registry, logger: logger}
}

// LoadFile parses a single data file into a Dataset.
func (l *Loader) LoadFile(path string) (*models.Dataset, error) {
	p, ok := l.registry.ParserFor(path)
	if !ok {
		return nil, fmt.Errorf("no parser registered for %s", path)
	}
	data, err := p.Parse(path)
	if err != nil {
		return nil, err
	}
	return &models.Dataset{
		Name:   datasetName(path),
		Path:   path,
		Data:   data,
		Format: strings.TrimPrefix(filepath.Ext(path), "."),
	}, nil
}

// LoadDirectory loads every supported file in dir. Unsupported files are
// ignored; files that fail to parse are returned as errors so callers can
// decide whether to skip them.
func (l *Loader) LoadDirectory(dir string) ([]*models.Dataset, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []error{fmt.Errorf("read directory %s: %w", dir, err)}
	}

	var datasets []*models.Dataset
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(dir, name)
		if _, ok := l.registry.ParserFor(full); !ok {
			l.logger.Debug("unsupported file ignored", "path", full)
			continue
		}
		d, err := l.LoadFile(full)
		if err != nil {
			errs = append(errs, fmt.Errorf("skip %s: %w", full, err))
			continue
		}
		datasets = append(datasets, d)
	}
	return datasets, errs
}

func datasetName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
