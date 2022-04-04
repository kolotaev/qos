package qos

import (
	"context"
	"io"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Throttler object that limits bandwidth for a particular server and connection.
// Throttler uses 1 second resolution and allows to set bandwidth limits in bytes.
// Thus minimum bandwidth value is `1 b/s` which is a fair minimum for a practical usage.
type Throttler struct {
	enabled           bool
	defaultLimit      int64
	connectionsLimits map[string]int64
	limiter           *rate.Limiter
	mu                *sync.RWMutex
}

// NewThrottler Throttler ctor.
func NewThrottler(defaultLimit int64, enabled bool) *Throttler {
	return &Throttler{
		enabled:           enabled,
		defaultLimit:      defaultLimit,
		connectionsLimits: make(map[string]int64),
		limiter:           rate.NewLimiter(rate.Every(time.Duration(1)*time.Second), 1),
		mu:                new(sync.RWMutex),
	}
}

// Enable bandwidth limitting.
func (t *Throttler) Enable() {
	t.enabled = true
}

// Disable bandwidth limitting.
func (t *Throttler) Disable() {
	t.enabled = false
}

// IsEnabled is bandwidth limitting enabled or not?
func (t *Throttler) IsEnabled() bool {
	return t.enabled
}

// SetBandwidthLimit set bandwidth limitting value for a server.
func (t *Throttler) SetBandwidthLimit(limit int64) {
	t.defaultLimit = limit
	t.mu.Lock()
	defer t.mu.Unlock()
	for key := range t.connectionsLimits {
		t.connectionsLimits[key] = t.defaultLimit
	}
}

// SetBandwidthLimitForConnection set bandwidth limitting value for a connection.
func (t *Throttler) SetBandwidthLimitForConnection(limit int64, connectionKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connectionsLimits[connectionKey] = limit
}

// Write write data to a connection.
func (t *Throttler) Write(ctx context.Context, dest io.Writer,
	destKey string, src io.Reader) (servedBytes int64, err error) {

	servedBytes = int64(0)

	for {
		var n int64
		connectionLimit := t.GetBandwidthLimitForConnection(destKey)
		err = t.limiter.Wait(ctx)
		if err != nil {
			return
		}
		if !t.enabled {
			n, err = io.Copy(dest, src)
			servedBytes += n
			return
		}
		n, err = io.CopyN(dest, src, connectionLimit)
		servedBytes += n
		if err != nil {
			return
		}
	}
}

// GetLimitForConnection get bandwidth limitting value for a connection.
func (t *Throttler) GetBandwidthLimitForConnection(connectionKey string) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.connectionsLimits[connectionKey]; !ok {
		return t.defaultLimit
	}
	return t.connectionsLimits[connectionKey]
}

// GetLimitForConnection get bandwidth limitting value for a connection.
func (t *Throttler) UnregisterConnection(connectionKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.connectionsLimits, connectionKey)
}
