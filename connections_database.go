package qos

import (
	"sync"
)

// ConnectionRecord is a record for a Connection
type ConnectionRecord struct {
	Limit              int64
	Active             bool
	HasIndividualLimit bool
}

// Database is a simple in-memory database for keeping record of the connections.
// It has a convenient DAO interface for interaction with the connections record set.
type Database struct {
	connections          map[string]*ConnectionRecord
	individualLimitCount int // used for optimization
	activeConnCount      int // used for optimization
	mu                   *sync.RWMutex
}

// NewDatabase Database ctor
func NewDatabase() *Database {
	return &Database{
		connections:          make(map[string]*ConnectionRecord),
		individualLimitCount: 0,
		activeConnCount:      0,
		mu:                   new(sync.RWMutex),
	}
}

// Activate upsert a connection in an active state
func (d *Database) Activate(connectionKey string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, exists := d.connections[connectionKey]

	if !exists {
		d.connections[connectionKey] = &ConnectionRecord{Active: true}
		d.activeConnCount++
		return
	}

	if !d.connections[connectionKey].Active {
		d.connections[connectionKey].Active = true
		d.activeConnCount++
	}
}

// Deactivate deactivates connection, but keeps it in the records store.
func (d *Database) Deactivate(connectionKey string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Already unregistered
	if !d.connections[connectionKey].Active {
		return
	}

	if d.connections[connectionKey].HasIndividualLimit {
		d.individualLimitCount--
	}
	d.connections[connectionKey].Active = false
	d.activeConnCount--
}

// Get get a connection by its key.
func (d *Database) Get(connectionKey string) *ConnectionRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connections[connectionKey]
}

// UpdateIndividualLimits sets individual limits for all connection that already have individual limits.
func (d *Database) UpdateIndividualLimits(limit int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for k, v := range d.connections {
		if v.HasIndividualLimit {
			d.connections[k].Limit = limit
		}
	}
}

// SetLimit set individual limit for a connection.
func (d *Database) SetLimit(limit int64, connectionKey string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.connections[connectionKey].Limit = limit
	d.connections[connectionKey].HasIndividualLimit = true
	d.individualLimitCount++
}

// CountActiveConnections get a number of active connections.
func (d *Database) CountActiveConnections() int {
	return d.activeConnCount
}

// CountConnectionsWithIndividualLimit get a number of connections that have individual limits.
func (d *Database) CountConnectionsWithIndividualLimit() int {
	return d.individualLimitCount
}
