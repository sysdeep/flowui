package toast

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

func BenchmarkToastComponentTree(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	host := New([]Item{
		NewItem("one", "Saved").Description("A persisted notification").Timeout(0),
		NewItem("two", "Uploading").Loading(true).Timeout(0),
		NewItem("three", "Warning").Variant(Warning).Timeout(0),
	}).Placement(BottomEnd)
	now := time.Unix(1, 0)
	layoutToastFrame(ctx, router, host, now, image.Pt(900, 600))
	baseline := host.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutToastFrame(ctx, router, host, now.Add(time.Duration(i)*time.Millisecond), image.Pt(900, 600))
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, host.visualCache.stats(), baseline)
}

func BenchmarkToastVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	host := New(nil).Key("bench")
	host.visualWithCache(ctx, Success, button.Ghost)
	baseline := host.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		host.visualWithCache(ctx, Success, button.Ghost)
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, host.visualCache.stats(), baseline)
}

func reportToastVisualCacheStats(b *testing.B, after, before computed.VisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "toast-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "toast-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "toast-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "toast-visual-cache-key-invalidations/op")
}

func BenchmarkToastVisualCacheInvalidation(b *testing.B) {
	first := new(theme.Light())
	second := new(theme.Light())
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	host := New(nil).Key("bench")
	host.visualWithCache(firstContext, Success, button.Ghost)
	baseline := host.visualCache.stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			host.visualWithCache(firstContext, Success, button.Ghost)
		case 1:
			host.visualWithCache(firstContext, Warning, button.Ghost)
		case 2:
			host.visualWithCache(secondContext, Warning, button.Ghost)
		default:
			host.visualWithCache(secondContext, Success, button.Ghost)
		}
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, host.visualCache.stats(), baseline)
}
