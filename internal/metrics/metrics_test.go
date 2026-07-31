package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Increments(t *testing.T) {
	m := New()
	m.IncRequests()
	m.IncRequests()
	m.IncRequests()
	m.IncReloads()

	snap := m.Snapshot()
	assert.Equal(t, int64(3), snap.Requests)
	assert.Equal(t, int64(1), snap.Reloads)

	first := m.Snapshot().Uptime
	time.Sleep(2 * time.Millisecond)
	assert.Greater(t, m.Snapshot().Uptime, first)
}

func TestMetrics_SnapshotIsolation(t *testing.T) {
	m := New()
	snap := m.Snapshot()
	m.IncRequests()
	assert.Equal(t, int64(0), snap.Requests)
}
