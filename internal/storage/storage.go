package storage

import (
	"sort"
	"strings"
	"sync"

	"github.com/marcuwynu23/dataserve/internal/models"
)

// Storage holds datasets in memory and provides thread-safe access.
//
// Reads use a read lock so many concurrent requests can be served while
// datasets are reloaded in the background.
type Storage struct {
	mu       sync.RWMutex
	datasets map[string]*models.Dataset
}

// New returns an empty Storage.
func New() *Storage {
	return &Storage{datasets: make(map[string]*models.Dataset)}
}

// Upsert stores or replaces a dataset under the given endpoint name.
func (s *Storage) Upsert(name string, d *models.Dataset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d.Name = normalize(name)
	s.datasets[d.Name] = d
}

// Get returns the dataset registered under name.
func (s *Storage) Get(name string) (*models.Dataset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.datasets[normalize(name)]
	return d, ok
}

// Delete removes the dataset registered under name.
func (s *Storage) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.datasets, normalize(name))
}

// Names returns the sorted list of registered endpoint names.
func (s *Storage) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.datasets))
	for name := range s.datasets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered datasets.
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.datasets)
}

func normalize(name string) string {
	return strings.TrimPrefix(strings.TrimSpace(name), "/")
}
