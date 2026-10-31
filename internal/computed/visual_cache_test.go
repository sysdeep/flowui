package computed

import (
	"github.com/qianniancn/flowui/theme"
	"testing"
)

type testVisualCache interface {
	Resolve(*theme.Theme, int, func() int) int
}

func visualCacheCases() []struct {
	name string
	new  func() testVisualCache
} {
	return []struct {
		name string
		new  func() testVisualCache
	}{
		{name: "single", new: func() testVisualCache { return &VisualCache[int, int]{} }},
		{name: "map", new: func() testVisualCache { return &VisualCacheMap[int, int]{} }},
		{name: "bounded map", new: func() testVisualCache {
			cache := &VisualCacheMap[int, int]{}
			cache.SetCapacity(1)
			return cache
		}},
	}
}

func TestVisualCachesResolveAndInvalidate(t *testing.T) {
	for _, tc := range visualCacheCases() {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.new()
			first, second := theme.Light(), theme.Dark()
			calls := 0
			load := func() int { calls++; return calls }
			for _, request := range []struct {
				active *theme.Theme
				key    int
				want   int
			}{
				{&first, 1, 1}, {&first, 1, 1},
				{&first, 2, 2}, {&second, 2, 3},
			} {
				if got := cache.Resolve(request.active, request.key, load); got != request.want {
					t.Fatalf("key %d resolved to %d, want %d", request.key, got, request.want)
				}
			}
			if calls != 3 {
				t.Fatalf("loader calls = %d, want 3", calls)
			}
		})
	}
}

func TestVisualCachesRetainPreviousValueAfterFailedLoad(t *testing.T) {
	for _, tc := range visualCacheCases() {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.new()
			first, second := theme.Light(), theme.Dark()
			cache.Resolve(&first, 1, func() int { return 7 })
			for _, request := range []struct {
				name   string
				active *theme.Theme
				key    int
			}{
				{"new key", &first, 2}, {"new theme", &second, 1},
			} {
				t.Run(request.name, func(t *testing.T) {
					func() {
						defer func() {
							if recover() == nil {
								t.Fatal("expected loader panic")
							}
						}()
						cache.Resolve(request.active, request.key, func() int { panic("load failed") })
					}()
					if got := cache.Resolve(&first, 1, func() int { return 9 }); got != 7 {
						t.Fatalf("value after failed load = %d, want 7", got)
					}
				})
			}
		})
	}
}

func TestVisualCacheMapRetainsMultipleKeys(t *testing.T) {
	first, second := theme.Light(), theme.Dark()
	var cache VisualCacheMap[int, int]
	cache.Resolve(&first, 1, func() int { return 10 })
	cache.Resolve(&first, 2, func() int { return 20 })
	if got := cache.Resolve(&first, 1, func() int { return 99 }); got != 10 {
		t.Fatalf("previous key = %d, want 10", got)
	}
	cache.Resolve(&second, 2, func() int { return 30 })
	if got := cache.Resolve(&second, 1, func() int { return 40 }); got != 40 {
		t.Fatalf("new theme reused old key: %d, want 40", got)
	}
}

func TestVisualCacheMapResolveWithStatusReportsHits(t *testing.T) {
	active := theme.Light()
	cache := VisualCacheMap[int, string]{}
	loads := 0
	load := func() string {
		loads++
		return "value"
	}

	if value, hit := cache.ResolveWithStatus(&active, 1, load); hit || value != "value" {
		t.Fatalf("first resolve = %q, hit=%v; want miss with value", value, hit)
	}
	if value, hit := cache.ResolveWithStatus(&active, 1, load); !hit || value != "value" {
		t.Fatalf("second resolve = %q, hit=%v; want hit with value", value, hit)
	}
	if loads != 1 {
		t.Fatalf("loader calls = %d, want 1", loads)
	}
}

func TestVisualCacheMapBoundedLRU(t *testing.T) {
	active := theme.Light()
	var cache VisualCacheMap[int, int]
	cache.SetCapacity(2)
	cache.Resolve(&active, 1, func() int { return 10 })
	cache.Resolve(&active, 2, func() int { return 20 })
	cache.Resolve(&active, 1, func() int { return 99 })
	cache.Resolve(&active, 3, func() int { return 30 })
	if got := cache.Resolve(&active, 1, func() int { return 99 }); got != 10 {
		t.Fatalf("recently used key was evicted: %d, want 10", got)
	}
	if got := cache.Resolve(&active, 2, func() int { return 40 }); got != 40 {
		t.Fatalf("least recently used key was retained: %d, want 40", got)
	}
}
