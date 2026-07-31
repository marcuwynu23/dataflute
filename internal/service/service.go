package service

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/marcuwynu23/dataflute/internal/config"
	"github.com/marcuwynu23/dataflute/internal/loader"
	"github.com/marcuwynu23/dataflute/internal/metrics"
	"github.com/marcuwynu23/dataflute/internal/models"
	"github.com/marcuwynu23/dataflute/internal/parser"
	"github.com/marcuwynu23/dataflute/internal/storage"
)

// Service is the application layer: it owns the datasets and exposes the
// operations needed by the HTTP handlers and the file watcher.
type Service struct {
	cfg      *config.Config
	storage  *storage.Storage
	loader   *loader.Loader
	metrics  *metrics.Metrics
	registry *parser.Registry
	logger   *slog.Logger
}

// New returns a Service wired with its dependencies.
func New(cfg *config.Config, st *storage.Storage, ld *loader.Loader, met *metrics.Metrics, reg *parser.Registry, logger *slog.Logger) *Service {
	return &Service{
		cfg:      cfg,
		storage:  st,
		loader:   ld,
		metrics:  met,
		registry: reg,
		logger:   logger,
	}
}

// LoadInitial loads all datasets at startup. In directory mode, files
// that fail to parse are skipped with a warning. In single file mode, a
// parse failure is returned as an error.
func (s *Service) LoadInitial() error {
	if s.cfg.Mode() == config.ModeDirectory {
		datasets, errs := s.loader.LoadDirectory(s.cfg.DataDir)
		for _, err := range errs {
			s.logger.Error("failed to load dataset", "error", err)
		}
		for _, d := range datasets {
			s.register(d)
		}
		return nil
	}

	d, err := s.loader.LoadFile(s.cfg.FilePath)
	if err != nil {
		return err
	}
	s.setEndpointName(d)
	s.storage.Upsert(d.Name, d)
	s.logger.Info("loaded dataset", "name", d.Name, "path", d.Path, "format", d.Format)
	return nil
}

// ReloadFile reparses a data file and replaces its dataset in place.
func (s *Service) ReloadFile(path string) error {
	d, err := s.loader.LoadFile(path)
	if err != nil {
		return err
	}
	s.setEndpointName(d)
	s.storage.Upsert(d.Name, d)
	s.metrics.IncReloads()
	s.logger.Info("reloaded dataset", "name", d.Name, "path", d.Path, "format", d.Format)
	return nil
}

// RemoveFile removes the dataset backed by path in directory mode.
func (s *Service) RemoveFile(path string) {
	if s.cfg.Mode() == config.ModeSingle {
		return
	}
	name := datasetNameFromPath(path)
	s.storage.Delete(name)
	s.logger.Info("removed dataset", "name", name, "path", path)
}

// Dataset returns the dataset registered under the given endpoint name.
func (s *Service) Dataset(name string) (*models.Dataset, bool) {
	return s.storage.Get(name)
}

// AvailableEndpoints returns the sorted list of "/"-prefixed endpoints.
func (s *Service) AvailableEndpoints() []string {
	names := s.storage.Names()
	endpoints := make([]string, 0, len(names))
	for _, n := range names {
		endpoints = append(endpoints, "/"+n)
	}
	return endpoints
}

// Formats returns the sorted list of supported file extensions.
func (s *Service) Formats() []string {
	return s.registry.Extensions()
}

// Mode returns the serving mode as a string.
func (s *Service) Mode() string {
	return string(s.cfg.Mode())
}

// Pretty reports whether responses should be pretty-printed.
func (s *Service) Pretty() bool {
	return s.cfg.Pretty
}

// Metrics returns the metrics collector used by request logging.
func (s *Service) Metrics() *metrics.Metrics {
	return s.metrics
}

// MetricsSnapshot returns the current statistics with the dataset count.
func (s *Service) MetricsSnapshot() metrics.Snapshot {
	snap := s.metrics.Snapshot()
	snap.Datasets = s.storage.Count()
	return snap
}

func (s *Service) register(d *models.Dataset) {
	s.setEndpointName(d)
	if isReserved(d.Name) {
		s.logger.Warn("dataset name is reserved, skipping",
			"name", d.Name, "path", d.Path)
		return
	}
	s.storage.Upsert(d.Name, d)
	s.logger.Info("loaded dataset", "name", d.Name, "path", d.Path, "format", d.Format)
}

func (s *Service) setEndpointName(d *models.Dataset) {
	if s.cfg.Mode() == config.ModeSingle {
		d.Name = strings.TrimPrefix(s.cfg.Endpoint, "/")
		return
	}
	d.Name = datasetNameFromPath(d.Path)
}

func datasetNameFromPath(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func isReserved(name string) bool {
	switch name {
	case "", "health", "metrics":
		return true
	}
	return false
}
