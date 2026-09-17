package middleware

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

type RateLimiter struct {
	mu         sync.Mutex
	cfg        cais.Config
	limit      int
	window     time.Duration
	buckets    map[string][]time.Time
	maxBuckets int
}

// bucketCleanupThreshold is when stale buckets are swept inline.
const bucketCleanupThreshold = 1000

// defaultMaxBuckets bounds memory under a flood of distinct IP+path keys.
const defaultMaxBuckets = 4096

func NewRateLimiter(limit int, cfg cais.Config) *RateLimiter {
	return &RateLimiter{
		cfg:        cfg,
		limit:      limit,
		window:     time.Minute,
		buckets:    make(map[string][]time.Time),
		maxBuckets: defaultMaxBuckets,
	}
}

// SetMaxBuckets caps concurrent buckets (<= 0 restores the default). When the
// cap is hit, stale buckets are swept and then the least recently used ones
// are evicted in one batch (#124).
func (rl *RateLimiter) SetMaxBuckets(n int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if n <= 0 {
		n = defaultMaxBuckets
	}
	rl.maxBuckets = n
	if len(rl.buckets) > rl.maxBuckets {
		rl.evictLocked(time.Now())
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := ClientIP(r, rl.cfg) + ":" + r.URL.Path
		if !rl.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	if len(rl.buckets) > bucketCleanupThreshold {
		rl.cleanupBuckets(cutoff)
	}
	if len(rl.buckets) > rl.maxBuckets {
		rl.evictLocked(now)
	}
	times := rl.buckets[key]
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= rl.limit {
		rl.setBucket(key, filtered)
		return false
	}
	rl.setBucket(key, append(filtered, now))
	return true
}

func (rl *RateLimiter) setBucket(key string, times []time.Time) {
	if len(times) == 0 {
		delete(rl.buckets, key)
		return
	}
	rl.buckets[key] = times
}

func (rl *RateLimiter) cleanupBuckets(cutoff time.Time) {
	for key, times := range rl.buckets {
		inWindow := false
		for _, ts := range times {
			if ts.After(cutoff) {
				inWindow = true
				break
			}
		}
		if !inWindow {
			delete(rl.buckets, key)
		}
	}
}

// evictLocked drops the least recently used buckets down to 3/4 of the cap,
// amortizing the O(n log n) sort over many requests (#124).
func (rl *RateLimiter) evictLocked(now time.Time) {
	rl.cleanupBuckets(now.Add(-rl.window))
	if len(rl.buckets) <= rl.maxBuckets {
		return
	}
	target := rl.maxBuckets * 3 / 4
	type bucketAge struct {
		key  string
		last time.Time
	}
	ages := make([]bucketAge, 0, len(rl.buckets))
	for key, times := range rl.buckets {
		last := time.Time{}
		for _, ts := range times {
			if ts.After(last) {
				last = ts
			}
		}
		ages = append(ages, bucketAge{key: key, last: last})
	}
	sort.Slice(ages, func(i, j int) bool { return ages[i].last.Before(ages[j].last) })
	for i := 0; i < len(ages) && len(rl.buckets) > target; i++ {
		delete(rl.buckets, ages[i].key)
	}
}
