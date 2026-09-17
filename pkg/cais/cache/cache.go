package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type entry[V any] struct {
	val       V
	expiresAt time.Time
}

// Cache is a thread-safe in-memory key-value store with per-entry TTL.
type Cache[V any] struct {
	mu         sync.RWMutex
	ttl        time.Duration
	data       map[string]entry[V]
	maxEntries int
}

// New returns a cache where each Set expires after ttl.
func New[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{
		ttl:        ttl,
		data:       make(map[string]entry[V]),
		maxEntries: defaultMaxEntries,
	}
}

// SetMaxEntries caps the number of live entries (<= 0 restores the default).
// When the cap is hit, expired entries are swept first and then the entries
// closest to expiry are evicted in one batch (#124), so a high-cardinality
// route (query strings via Key) cannot grow the heap without bound.
func (c *Cache[V]) SetMaxEntries(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n <= 0 {
		n = defaultMaxEntries
	}
	c.maxEntries = n
	if len(c.data) > c.maxEntries {
		c.evictLocked(time.Now())
	}
}

// Get returns the value for key when present and not expired.
func (c *Cache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.val, true
}

// sweepThreshold bounds memory: expired entries are only removed lazily on
// overwrite, so high-cardinality keys would otherwise leak forever (#174).
const sweepThreshold = 1024

// defaultMaxEntries is the hard cap on live entries (see SetMaxEntries).
const defaultMaxEntries = 4096

// Set stores val under key; it expires after the cache TTL.
// When the map grows past sweepThreshold, expired entries are swept inline;
// past maxEntries, the oldest entries are evicted in one batch (#124).
func (c *Cache[V]) Set(key string, val V) {
	c.mu.Lock()
	c.data[key] = entry[V]{
		val:       val,
		expiresAt: time.Now().Add(c.ttl),
	}
	if len(c.data) > sweepThreshold {
		c.sweepExpiredLocked(time.Now())
	}
	if len(c.data) > c.maxEntries {
		c.evictLocked(time.Now())
	}
	c.mu.Unlock()
}

func (c *Cache[V]) sweepExpiredLocked(now time.Time) {
	for k, e := range c.data {
		if now.After(e.expiresAt) {
			delete(c.data, k)
		}
	}
}

// evictLocked drops the entries closest to expiry down to 3/4 of the cap in a
// single pass, amortizing the O(n log n) sort over many Sets (#124).
func (c *Cache[V]) evictLocked(now time.Time) {
	c.sweepExpiredLocked(now)
	if len(c.data) <= c.maxEntries {
		return
	}
	target := c.maxEntries * 3 / 4
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return c.data[keys[i]].expiresAt.Before(c.data[keys[j]].expiresAt)
	})
	for i := 0; i < len(keys) && len(c.data) > target; i++ {
		delete(c.data, keys[i])
	}
}

// Len reports the number of entries currently stored (including not-yet-expired).
func (c *Cache[V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Delete removes key from the cache.
func (c *Cache[V]) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Key builds a deterministic cache key by joining the parts with "|".
// Recommended for list pages: combine stable identifiers + a Hash() of a version struct
// (e.g. count + max updated time) instead of serializing the entire list into the key.
func Key(parts ...any) string {
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteByte('|')
		}
		fmt.Fprintf(&b, "%v", p)
	}
	return b.String()
}

// Hash returns a short (16 hex chars) stable hash of v.
// Use it to create cache keys / ETags that change only when the important data changes.
func Hash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(fmt.Sprintf("%#v", v))
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
