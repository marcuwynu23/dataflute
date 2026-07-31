package metrics

import (
	"sync/atomic"
	"time"
)

// Metrics collects runtime statistics for the /metrics endpoint.
type Metrics struct {
	start    time.Time
	requests atomic.Int64
	reloads  atomic.Int64
}

// Snapshot is an immutable copy of the runtime statistics.
type Snapshot struct {
	// Uptime is how long the server has been running.
	Uptime time.Duration

	// Requests is the number of HTTP requests served.
	Requests int64

	// Reloads is the number of dataset reloads performed.
	Reloads int64

	// Datasets is the number of currently registered datasets.
	Datasets int
}

// New returns a Metrics instance with its start time set to now.
func New() *Metrics {
	return &Metrics{start: time.Now()}
}

// IncRequests records one incoming request.
func (m *Metrics) IncRequests() { m.requests.Add(1) }

// IncReloads records one successful dataset reload.
func (m *Metrics) IncReloads() { m.reloads.Add(1) }

// Snapshot returns a point-in-time copy of the statistics.
func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		Uptime:   time.Since(m.start),
		Requests: m.requests.Load(),
		Reloads:  m.reloads.Load(),
	}
}
