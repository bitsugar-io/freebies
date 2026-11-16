// Package ttlcache is a small bounded cache with per-key in-flight
// de-duplication, for the HTTP-backed game sources. Different keys fetch
// concurrently; concurrent callers of the same key share one fetch.
package ttlcache

import (
	"context"
	"sync"
	"time"
)

// Cache holds up to max values for ttl each.
type Cache[V any] struct {
	ttl time.Duration
	max int
	now func() time.Time

	mu      sync.Mutex
	entries map[string]entry[V]
	flight  map[string]*call[V]
}

type entry[V any] struct {
	val     V
	fetched time.Time
}

type call[V any] struct {
	done chan struct{}
	val  V
	err  error
}

// New returns a cache whose values live for ttl and which never holds more
// than max (the oldest are evicted first).
func New[V any](ttl time.Duration, max int) *Cache[V] {
	return &Cache[V]{
		ttl: ttl, max: max, now: time.Now,
		entries: map[string]entry[V]{}, flight: map[string]*call[V]{},
	}
}

// Len reports how many values are cached.
func (c *Cache[V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// Get returns the cached value for key, or calls fetch once (however many
// goroutines ask at the same time) and caches its result. Errors are not
// cached. A waiting caller returns early with its own ctx error if ctx ends.
func (c *Cache[V]) Get(ctx context.Context, key string, fetch func(context.Context) (V, error)) (V, error) {
	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Sub(e.fetched) < c.ttl {
		c.mu.Unlock()
		return e.val, nil
	}
	if f, ok := c.flight[key]; ok {
		c.mu.Unlock()
		select {
		case <-f.done:
			return f.val, f.err
		case <-ctx.Done():
			var zero V
			return zero, ctx.Err()
		}
	}
	f := &call[V]{done: make(chan struct{})}
	c.flight[key] = f
	c.mu.Unlock()

	f.val, f.err = fetch(ctx)

	c.mu.Lock()
	delete(c.flight, key)
	if f.err == nil {
		c.store(key, f.val)
	}
	c.mu.Unlock()
	close(f.done)
	return f.val, f.err
}

// store inserts under c.mu, evicting expired then oldest entries to stay <= max.
func (c *Cache[V]) store(key string, v V) {
	now := c.now()
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.max {
		for k, e := range c.entries {
			if now.Sub(e.fetched) >= c.ttl {
				delete(c.entries, k)
			}
		}
		for len(c.entries) >= c.max {
			var oldest string
			var at time.Time
			for k, e := range c.entries {
				if oldest == "" || e.fetched.Before(at) {
					oldest, at = k, e.fetched
				}
			}
			delete(c.entries, oldest)
		}
	}
	c.entries[key] = entry[V]{val: v, fetched: now}
}
