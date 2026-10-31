package sidebar

import (
	"fmt"
	"image"
	"runtime"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type sidebarBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newSidebarBenchmarkFrame(maximum image.Point) *sidebarBenchmarkFrame {
	frame := &sidebarBenchmarkFrame{ctx: core.New(nil, new(theme.Light()), locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(), Ops: &frame.ops, Now: time.Unix(1, 0),
	}
	return frame
}

func (f *sidebarBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func sidebarBenchmarkTree() Widget {
	sections := make([]Section, 4)
	for sectionIndex := range sections {
		sections[sectionIndex].Title = fmt.Sprintf("Section %d", sectionIndex)
		sections[sectionIndex].Items = make([]Item, 8)
		for itemIndex := range sections[sectionIndex].Items {
			sections[sectionIndex].Items[itemIndex] = Item{
				Key: fmt.Sprintf("item-%d-%d", sectionIndex, itemIndex), Label: "Navigation item",
			}
		}
	}
	return NewSections("benchmark-tree", "item-0-0", sections).DataVersion(1)
}

func reportSidebarDefaultsStats(b *testing.B, after, before computed.VisualCacheStats) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "sidebar-defaults-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "sidebar-defaults-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "sidebar-defaults-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "sidebar-defaults-cache-key-invalidations/op")
}

func BenchmarkSidebarDefaultsCacheHit(b *testing.B) {
	active := new(theme.Light())
	metrics := (Widget{}).metrics(active)
	cache := sidebarDefaultsCache{}
	cache.resolve(active, metrics)
	before := cache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.resolve(active, metrics)
	}
	b.StopTimer()
	reportSidebarDefaultsStats(b, cache.stats(), before)
}

func BenchmarkSidebarComponentTreeSteadyState(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSidebarComponentTreeDefaultsCacheHit(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	state, ok := core.PeekState[sidebarState](frame.ctx, "benchmark-tree", stateSlotSidebar)
	if !ok {
		b.Fatal("sidebar state was not initialized")
	}
	before := state.defaults.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportSidebarDefaultsStats(b, state.defaults.stats(), before)
}

func BenchmarkSidebarComponentTreeDefaultsCacheMiss(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	state, ok := core.PeekState[sidebarState](frame.ctx, "benchmark-tree", stateSlotSidebar)
	if !ok {
		b.Fatal("sidebar state was not initialized")
	}
	first := theme.Light()
	second := theme.Light()
	before := state.defaults.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, first)
		} else {
			core.ReplaceTheme(frame.ctx, second)
		}
		frame.layout(value)
	}
	b.StopTimer()
	reportSidebarDefaultsStats(b, state.defaults.stats(), before)
}

func BenchmarkSidebarInstanceStyle(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := (sidebarBenchmarkTree()).Style(style.Style{}.Padding(2))
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSidebarCloneSections(b *testing.B) {
	sections := make([]Section, 8)
	for sectionIndex := range sections {
		sections[sectionIndex].Title = fmt.Sprintf("Section %d", sectionIndex)
		sections[sectionIndex].Items = make([]Item, 8)
		for itemIndex := range sections[sectionIndex].Items {
			sections[sectionIndex].Items[itemIndex] = Item{
				Key:   fmt.Sprintf("item-%d-%d", sectionIndex, itemIndex),
				Label: "Item",
			}
		}
	}

	b.ReportAllocs()
	for b.Loop() {
		cloned := cloneSections(sections)
		runtime.KeepAlive(cloned)
	}
}
