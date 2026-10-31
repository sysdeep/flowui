package titlebar

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkTitleBarLayoutSteadyState(b *testing.B) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	router := new(input.Router)
	value := New("workspace", "FlowUI", nil)
	layoutTitleBarWithConstraints(ctx, router, value, time.Unix(0, 0), layout.Exact(image.Pt(800, 35)))
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutTitleBarWithConstraints(ctx, router, value, time.Unix(int64(index), 0), layout.Exact(image.Pt(800, 35)))
	}
	b.StopTimer()
	reportTitleBarVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkTitleBarVisualCacheHit(b *testing.B) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("workspace", "FlowUI", nil)
	value.visualWithCache(ctx)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx)
	}
	b.StopTimer()
	reportTitleBarVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func reportTitleBarVisualCacheStats(b *testing.B, after, before computed.VisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "titlebar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "titlebar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "titlebar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "titlebar-visual-cache-key-invalidations/op")
}

func BenchmarkTitleBarVisualCacheInvalidation(b *testing.B) {
	first := new(theme.Light())
	second := new(theme.Light())
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("workspace", "FlowUI", nil)
	value.visualWithCache(firstContext)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 7 {
		case 0:
			value.visualWithCache(firstContext)
		case 1:
			value.visualWithCache(firstContext)
		case 2:
			value.visualWithCache(secondContext)
		case 3:
			value.visualWithCache(secondContext)
		case 4:
			value.clientDecorations = false
			value.visualWithCache(secondContext)
		case 5:
			value.visualWithCache(secondContext)
		case 6:
			value.clientDecorations = true
			value.visualWithCache(secondContext)
		case 7:
			value.visualWithCache(secondContext)
		}
	}
	b.StopTimer()
	reportTitleBarVisualCacheStats(b, value.visualCache.Stats(), computed.VisualCacheStats{})
}

func BenchmarkTitleBarStyledVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	value := New("workspace", "FlowUI", nil).Style(style.Style{}.Height(40))
	value.visualWithCache(ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx)
	}
}
