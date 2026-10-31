package text

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkLabelLayout(b *testing.B) {
	themeValue := new(theme.Light())
	ctx := core.New(nil, themeValue, locale.LanguageEnglish)
	value := NewLabel("Field name").Required(true)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Ops:         &ops,
	}
	value.Layout(ctx, gtx)
	ops.Reset()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ops.Reset()
		value.Layout(ctx, gtx)
	}
}

func BenchmarkLabelVisualCacheHit(b *testing.B) {
	themeValue := new(theme.Light())
	ctx := core.New(nil, themeValue, locale.LanguageEnglish)
	value := NewLabel("Field name")
	value.visualWithCache(ctx)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx)
	}
	b.StopTimer()
	stats := value.visualCache.Stats()
	b.ReportMetric(float64(stats.Hits-baseline.Hits)/float64(b.N), "label-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses-baseline.Misses)/float64(b.N), "label-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations-baseline.ThemeInvalidations)/float64(b.N), "label-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations-baseline.KeyInvalidations)/float64(b.N), "label-visual-cache-key-invalidations/op")
}

func BenchmarkLabelVisualCacheMiss(b *testing.B) {
	active := new(theme.Light())
	var cache labelVisualCache
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		foreground := colorForLabelBenchmark(index)
		key := labelVisualKey{Foreground: foreground}
		cache.Resolve(active, key, func() labelVisual {
			return labelResolveLabelVisual(active, foreground)
		})
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "label-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "label-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "label-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "label-visual-cache-key-invalidations/op")
}

func BenchmarkLabelVisualCacheThemeInvalidation(b *testing.B) {
	first := new(theme.Light())
	second := new(theme.Light())
	var cache labelVisualCache
	foreground := first.Semantic().Foreground
	key := labelVisualKey{Foreground: foreground}
	cache.Resolve(first, key, func() labelVisual {
		return labelResolveLabelVisual(first, foreground)
	})
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		active := first
		if index&1 != 0 {
			active = second
		}
		cache.Resolve(active, key, func() labelVisual {
			return labelResolveLabelVisual(active, foreground)
		})
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "label-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "label-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "label-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "label-visual-cache-key-invalidations/op")
}

func colorForLabelBenchmark(index int) color.NRGBA {
	return color.NRGBA{R: uint8(index), G: uint8(index >> 8), B: uint8(index >> 16), A: 0xff}
}
