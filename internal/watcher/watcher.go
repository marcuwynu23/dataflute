package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/marcuwynu23/dataserve/internal/config"
	"github.com/marcuwynu23/dataserve/internal/parser"
)

// Handler is the subset of Service operations the watcher needs.
type Handler interface {
	// ReloadFile reparses a file and replaces its dataset.
	ReloadFile(path string) error

	// RemoveFile removes the dataset backed by path.
	RemoveFile(path string)
}

// debounceDelay groups bursts of filesystem events before reloading.
const debounceDelay = 200 * time.Millisecond

// Watcher monitors data files for changes and reloads affected datasets
// without requiring a server restart.
type Watcher struct {
	cfg      *config.Config
	handler  Handler
	registry *parser.Registry
	logger   *slog.Logger
}

// New returns a Watcher. The target path is not validated here; that is
// done by config validation.
func New(cfg *config.Config, handler Handler, registry *parser.Registry, logger *slog.Logger) (*Watcher, error) {
	w := &Watcher{cfg: cfg, handler: handler, registry: registry, logger: logger}
	if _, err := w.watchTarget(); err != nil {
		return nil, err
	}
	return w, nil
}

// Start begins watching the configured path and blocks until ctx is
// cancelled or the watcher fails.
func (w *Watcher) Start(ctx context.Context) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create file watcher: %w", err)
	}
	defer fsw.Close()

	target, err := w.watchTarget()
	if err != nil {
		return err
	}
	if err := fsw.Add(target); err != nil {
		return fmt.Errorf("watch %s: %w", target, err)
	}
	w.logger.Info("watching for changes", "path", target)

	var (
		mu      sync.Mutex
		pending = make(map[string]struct{})
		timer   *time.Timer
	)

	flush := func() {
		mu.Lock()
		paths := make([]string, 0, len(pending))
		for p := range pending {
			paths = append(paths, p)
		}
		pending = make(map[string]struct{})
		timer = nil
		mu.Unlock()

		for _, p := range paths {
			if err := w.handler.ReloadFile(p); err != nil {
				w.logger.Error("reload failed", "path", p, "error", err)
			}
		}
	}

	schedule := func(p string) {
		mu.Lock()
		pending[p] = struct{}{}
		if timer == nil {
			timer = time.AfterFunc(debounceDelay, flush)
		}
		mu.Unlock()
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-fsw.Errors:
			if !ok {
				return nil
			}
			w.logger.Error("file watcher error", "error", err)
		case ev, ok := <-fsw.Events:
			if !ok {
				return nil
			}
			if !w.filterEvent(ev) {
				continue
			}
			switch {
			case ev.Op&fsnotify.Remove != 0:
				w.handler.RemoveFile(ev.Name)
			case ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0:
				schedule(ev.Name)
			}
		}
	}
}

// watchTarget returns the directory to watch. Single file mode watches
// the file's parent directory and filters events to the file itself,
// which is robust against files being atomically replaced.
func (w *Watcher) watchTarget() (string, error) {
	if w.cfg.Mode() == config.ModeSingle {
		return filepath.Dir(w.cfg.FilePath), nil
	}
	return w.cfg.DataDir, nil
}

// filterEvent reports whether an event should be processed.
func (w *Watcher) filterEvent(ev fsnotify.Event) bool {
	if w.cfg.Mode() == config.ModeSingle {
		return pathsEqual(ev.Name, w.cfg.FilePath)
	}
	base := filepath.Base(ev.Name)
	if strings.HasPrefix(base, ".") {
		return false
	}
	_, ok := w.registry.ParserFor(ev.Name)
	return ok
}

// pathsEqual compares paths case-insensitively so that watchers on
// case-insensitive filesystems still match.
func pathsEqual(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}
