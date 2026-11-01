package menubar

import (
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkMenubarLayout(b *testing.B) {
	benchmarkMenubarLayout(b)
}

func BenchmarkMenubarLayoutSteadyState(b *testing.B) {
	benchmarkMenubarLayout(b)
}

func benchmarkMenubarLayout(b *testing.B) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(1, 0)
	layoutMenubarFrame(ctx, router, widget, now)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		layoutMenubarFrame(ctx, router, widget, now)
	}
}

func BenchmarkMenubarVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	cache := &menubarDefaultsCache{}
	widget := Widget{}
	widget.defaultsWithCache(ctx, cache)
	before := cache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		widget.defaultsWithCache(ctx, cache)
	}
	b.StopTimer()
	reportMenubarVisualCacheStats(b, cache.Stats(), before)
}

func BenchmarkMenubarVisualCacheInvalidation(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	cache := &menubarDefaultsCache{}
	widget := Widget{}
	widget.defaultsWithCache(ctx, cache)
	first := new(theme.Light())
	second := new(theme.Light())
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index % 4 {
		case 0, 1:
			core.ReplaceTheme(ctx, *first)
			widget.compact = false
		case 2, 3:
			core.ReplaceTheme(ctx, *second)
			widget.compact = true
		}
		widget.defaultsWithCache(ctx, cache)
	}
	b.StopTimer()
	reportMenubarVisualCacheStats(b, cache.Stats(), computed.VisualCacheStats{})
}

func reportMenubarVisualCacheStats(b *testing.B, after, before computed.VisualCacheStats) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "menubar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "menubar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "menubar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "menubar-visual-cache-key-invalidations/op")
}

func BenchmarkMenubarComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	items := make([]Item, 0, 32)
	for index := 0; index < cap(items); index++ {
		items = append(items, NewMenu("item-"+strconv.Itoa(index), "Item "+strconv.Itoa(index), nil))
	}
	widget := New(items).Key("benchmark")
	now := time.Unix(0, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutMenubarFrame(ctx, router, widget, now.Add(time.Duration(index+1)*time.Millisecond))
	}
}
