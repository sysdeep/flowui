package navigation

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

type paginationBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newPaginationBenchmarkFrame(maximum image.Point) *paginationBenchmarkFrame {
	active := new(theme.Light())
	frame := &paginationBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *paginationBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkPaginationLayoutSteadyState(b *testing.B) {
	frame := newPaginationBenchmarkFrame(image.Pt(640, 80))
	value := NewPagination("benchmark-pagination", 6, 128).Size(PaginationSizeMedium)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkPaginationComponentTreeSteadyState(b *testing.B) {
	frame := newPaginationBenchmarkFrame(image.Pt(640, 1200))
	values := make([]PaginationWidget, 128)
	for index := range values {
		value := NewPagination(fmt.Sprintf("pagination-%d", index), index%64+1, 128)
		if index%3 == 0 {
			value = value.Size(PaginationSizeSmall)
		} else if index%3 == 1 {
			value = value.Size(PaginationSizeLarge).ShowControls(false)
		}
		values[index] = value
	}
	root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, len(values))
		for _, value := range values {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return value.Layout(ctx, gtx)
			}))
		}
		return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx, children...)
	})
	frame.layout(root)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(root)
	}
}

func BenchmarkPaginationVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	value := NewPagination("pagination-cache", 6, 128)
	value.visualWithCache(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx)
	}
	b.StopTimer()
	stats := value.visualCache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "pagination-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "pagination-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "pagination-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "pagination-visual-cache-key-invalidations/op")
}

func BenchmarkPaginationVisualCacheMiss(b *testing.B) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	b.ReportAllocs()
	for b.Loop() {
		value := NewPagination("pagination-cache", 6, 128)
		value.visualWithCache(ctx)
	}
	b.StopTimer()
	b.ReportMetric(0, "pagination-visual-cache-hits/op")
	b.ReportMetric(1, "pagination-visual-cache-misses/op")
	b.ReportMetric(0, "pagination-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "pagination-visual-cache-key-invalidations/op")
}

func BenchmarkPaginationVisualCacheThemeInvalidation(b *testing.B) {
	first := new(theme.Light())
	second := new(theme.Light())
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := NewPagination("pagination-cache", 6, 128)
	value.visualWithCache(firstContext)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			value.visualWithCache(firstContext)
		} else {
			value.visualWithCache(secondContext)
		}
	}
	b.StopTimer()
	stats := value.visualCache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "pagination-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "pagination-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "pagination-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "pagination-visual-cache-key-invalidations/op")
}
