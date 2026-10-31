package tabs

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkTabsLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{
		{Key: "tab-0", Label: "Overview"}, {Key: "tab-1", Label: "Activity"},
		{Key: "tab-2", Label: "Settings"},
	}).Fit().Separators(true)
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(480, 160)}}
	var router input.Router
	value.Layout(ctx, gtx)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		gtx.Source = router.Source()
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		value.Layout(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(gtx.Ops)
	}
	b.StopTimer()
	reportTabsDefaultsCacheStats(b, value.visualCache.Stats(), baseline)
}

func reportTabsDefaultsCacheStats(b *testing.B, after, before computed.VisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "tabs-default-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "tabs-default-cache-misses/op")
	b.ReportMetric(
		float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N),
		"tabs-default-cache-theme-invalidations/op",
	)
	b.ReportMetric(
		float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N),
		"tabs-default-cache-key-invalidations/op",
	)
}

func BenchmarkTabsDefaultsCacheHit(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{{Key: "tab-0", Label: "Overview"}})
	value.defaultsWithCache(ctx)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value.defaultsWithCache(ctx)
	}
	b.StopTimer()
	reportTabsDefaultsCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkTabsDefaultsCacheInvalidation(b *testing.B) {
	first := new(theme.Light())
	second := new(theme.Light())
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{{Key: "tab-0", Label: "Overview"}})
	value.defaultsWithCache(firstContext)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			value.defaultsWithCache(firstContext)
		case 1:
			value.defaultsWithCache(secondContext)
		case 2:
			value.Size(Small).defaultsWithCache(secondContext)
		default:
			value.defaultsWithCache(secondContext)
		}
	}
	b.StopTimer()
	reportTabsDefaultsCacheStats(b, value.visualCache.Stats(), computed.VisualCacheStats{})
}
