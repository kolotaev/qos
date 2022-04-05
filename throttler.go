package qos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Throttler object that limits bandwidth for a particular server and connection.
// Throttler uses 1 second resolution and allows to set bandwidth limits in bytes.
// Thus minimum bandwidth value is `1 b/s` which is a fair minimum for a practical usage.
type Throttler struct {
	enabled       bool
	totalLimit    int64
	freeLimitPool int64 // used for optimization
	db            *Database
	limiter       *rate.Limiter
	mu            *sync.RWMutex
	listener      net.Listener
}

// NewThrottler Throttler ctor.
func NewThrottler(totalLimit int64, enabled bool) *Throttler {
	return &Throttler{
		enabled:       enabled,
		totalLimit:    totalLimit,
		freeLimitPool: totalLimit,
		db:            NewDatabase(),
		limiter:       rate.NewLimiter(rate.Every(time.Duration(1)*time.Second), 1),
		mu:            new(sync.RWMutex),
	}
}

// Listen start listening to incoming connections.
func (t *Throttler) Listen(network, address string) error {
	if t.listener != nil {
		return errors.New("listening was started previously, it can be started only once")
	}
	l, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("failed to listen with Throttler: %s", err)
	}
	t.listener = l
	return nil
}

// Accept waits for and returns the next connection to the listener.
func (t *Throttler) Accept() (net.Conn, error) {
	if t.listener == nil {
		return nil, errors.New("please start listening first")
	}
	return t.listener.Accept()
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (t *Throttler) Close() error {
	if t.listener == nil {
		return errors.New("please start listening first")
	}
	return t.listener.Close()
}

// Addr returns the listener's network address.
func (t *Throttler) Addr() net.Addr {
	if t.listener == nil {
		return nil
	}
	return t.listener.Addr()
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

// Write write data from the input source to an output writer.
func (t *Throttler) Write(ctx context.Context, dest io.Writer,
	destKey string, src io.Reader) (servedBytes int64, err error) {

	servedBytes = int64(0)

	t.RegisterConnection(destKey)

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

// SetBandwidthLimit set bandwidth limitting value for a server.
func (t *Throttler) SetBandwidthLimit(limit int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If we increase global limit - just update and increase free pool
	if limit >= t.totalLimit {
		t.freeLimitPool += limit - t.totalLimit
		t.totalLimit = limit
		return
	}

	// If we decrease global limit - make sure existing individual limits sum
	// doesn't exceed new max total limit capacity.
	individualLimitsSum := t.totalLimit - t.freeLimitPool
	if individualLimitsSum <= limit {
		t.freeLimitPool = limit - individualLimitsSum
		t.totalLimit = limit
		return
	}
	// If so, just decrease existing individual limits by a proportional amount.
	minAllowed := int64(math.Floor(float64(limit) / float64(t.db.CountConnectionsWithIndividualLimit())))
	t.db.UpdateIndividualLimits(minAllowed)
	t.freeLimitPool = 0
	t.totalLimit = limit
}

// SetBandwidthLimitForConnection set bandwidth limitting value for a connection.
func (t *Throttler) SetBandwidthLimitForConnection(limit int64, connectionKey string) {
	t.RegisterConnection(connectionKey)

	t.mu.Lock()
	defer t.mu.Unlock()

	// We can't allow to use more than we have in free allowed bandwidth per pool
	if limit > t.freeLimitPool {
		limit = t.freeLimitPool
	}

	t.freeLimitPool -= limit
	t.db.SetLimit(limit, connectionKey)
}

// GetLimitForConnection get bandwidth limitting value for a connection.
func (t *Throttler) GetBandwidthLimitForConnection(connectionKey string) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	c := t.db.Get(connectionKey)
	if !c.Active {
		return 0
	}
	if c.HasIndividualLimit {
		return c.Limit
	}

	countWithoutIndividualLimits := t.db.CountActiveConnections() - t.db.CountConnectionsWithIndividualLimit()
	return int64(math.Floor(float64(t.freeLimitPool) / float64(countWithoutIndividualLimits)))
}

// RegisterConnection register a connection.
func (t *Throttler) RegisterConnection(connectionKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.db.Activate(connectionKey)
}

// UnregisterConnection unregister connection.
func (t *Throttler) UnregisterConnection(connectionKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.db.Deactivate(connectionKey)

	c := t.db.Get(connectionKey)
	if c.HasIndividualLimit {
		t.freeLimitPool += c.Limit
	}
}
