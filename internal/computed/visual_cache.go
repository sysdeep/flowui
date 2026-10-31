package computed

import (
	"github.com/qianniancn/flowui/theme"
)

// VisualCache memoizes one visual result for a theme and a comparable request key.
// The cache is intentionally instance-owned: widgets decide the
// request key and retain control of visual validation and request creation.
//
// Resolve invokes load only after a miss. A panic from load therefore leaves
// the previous entry intact and does not turn an unsuccessful request into a
// valid cache value.
type VisualCache[K comparable, V any] struct {
	theme *theme.Theme
	key   K
	value V
	valid bool

	hits               uint64
	misses             uint64
	themeInvalidations uint64
	keyInvalidations   uint64
}

// VisualCacheMap memoizes visual results for multiple comparable request
// keys while sharing the same theme and invalidation lifecycle as
// VisualCache. It is useful for compound widgets that reuse more than one
// variant in a frame or across frames.
type VisualCacheMap[K comparable, V any] struct {
	theme  *theme.Theme
	values map[K]V
	valid  bool

	// bounded is opt-in so existing map caches keep their zero-value, unbounded
	// behavior without paying for LRU bookkeeping.
	bounded  bool
	capacity int
	clock    uint64
	lastUsed map[K]uint64

	hits               uint64
	misses             uint64
	themeInvalidations uint64
	keyInvalidations   uint64
	evictions          uint64
}

// SetCapacity enables a bounded LRU policy for this cache. A positive
// capacity is required because zero-value VisualCacheMap instances are
// intentionally unbounded for existing callers. Changing the capacity clears
// any value already held by the cache.
func (c *VisualCacheMap[K, V]) SetCapacity(capacity int) {
	if c == nil {
		return
	}
	if capacity <= 0 {
		panic("computed: visual cache capacity must be positive")
	}
	if c.bounded && c.capacity == capacity {
		return
	}
	if c.valid {
		c.theme = nil
		c.values = nil
		c.valid = false
	}
	c.bounded = true
	c.capacity = capacity
	c.clock = 0
	c.lastUsed = nil
}

// Resolve returns the cached value for active and key, or loads and stores a
// fresh value on a miss. Loading happens before a new theme map is committed,
// so a panic cannot turn an unsuccessful request into a valid cache value.
func (c *VisualCacheMap[K, V]) Resolve(active *theme.Theme, key K, load func() V) V {
	value, _ := c.ResolveWithStatus(active, key, load)
	return value
}

// ResolveWithStatus is Resolve with a hit flag for callers that expose cache
// evidence. The loading and invalidation semantics are identical to Resolve.
func (c *VisualCacheMap[K, V]) ResolveWithStatus(active *theme.Theme, key K, load func() V) (V, bool) {
	if c == nil {
		return load(), false
	}

	themeChanged := c.valid && c.theme != active
	if themeChanged {
		c.themeInvalidations++
	} else if c.valid {
		if value, ok := c.values[key]; ok {
			c.hits++
			c.touch(key)
			return value, true
		}
		if len(c.values) > 0 {
			c.keyInvalidations++
		}
	}

	c.misses++
	value := load()
	if themeChanged || !c.valid {
		capacity := 2
		if c.bounded {
			capacity = c.capacity
		}
		c.values = make(map[K]V, capacity)
		c.lastUsed = nil
		c.clock = 0
		c.theme = active
		c.valid = true
	}
	if c.bounded {
		if len(c.values) >= c.capacity {
			var oldestKey K
			var oldestUse uint64
			first := true
			for candidate, lastUsed := range c.lastUsed {
				if first || lastUsed < oldestUse {
					oldestKey = candidate
					oldestUse = lastUsed
					first = false
				}
			}
			if !first {
				delete(c.values, oldestKey)
				delete(c.lastUsed, oldestKey)
				c.evictions++
			}
		}
		if c.lastUsed == nil {
			c.lastUsed = make(map[K]uint64, c.capacity)
		}
	}
	c.values[key] = value
	c.touch(key)
	return value, false
}

func (c *VisualCacheMap[K, V]) touch(key K) {
	if c == nil || !c.bounded {
		return
	}
	c.clock++
	c.lastUsed[key] = c.clock
}

// Stats returns a snapshot of cache counters without exposing cached values.
func (c *VisualCacheMap[K, V]) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	return VisualCacheStats{
		Hits:               c.hits,
		Misses:             c.misses,
		ThemeInvalidations: c.themeInvalidations,
		KeyInvalidations:   c.keyInvalidations,
		Evictions:          c.evictions,
		Size:               len(c.values),
	}
}

// Resolve returns the cached value for active and key, or loads and stores a
// fresh value on a miss. A nil receiver deliberately behaves as an uncached
// resolver, which lets callers keep optional test or benchmark caches simple.
func (c *VisualCache[K, V]) Resolve(active *theme.Theme, key K, load func() V) V {
	if c == nil {
		return load()
	}
	themeChanged := c.valid && c.theme != active
	if c.valid {
		if themeChanged {
			c.themeInvalidations++
		}
		if c.key != key {
			c.keyInvalidations++
		}
	}
	if c.valid && !themeChanged && c.key == key {
		c.hits++
		return c.value
	}

	value := load()
	c.theme = active
	c.key = key
	c.value = value
	c.valid = true
	c.misses++
	return value
}

// VisualCacheStats is a snapshot of one VisualCache instance.
type VisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64 // Theme pointer changes.
	KeyInvalidations   uint64
	Evictions          uint64
	Size               int
}

// Stats returns cache counters without exposing the cached value or key.
func (c *VisualCache[K, V]) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	return VisualCacheStats{
		Hits:               c.hits,
		Misses:             c.misses,
		ThemeInvalidations: c.themeInvalidations,
		KeyInvalidations:   c.keyInvalidations,
	}
}
