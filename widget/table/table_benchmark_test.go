package table

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

type tableCheckboxBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newTableCheckboxBenchmarkFrame(maximum image.Point) *tableCheckboxBenchmarkFrame {
	activeTheme := theme.Light()
	frame := &tableCheckboxBenchmarkFrame{
		ctx: core.New(nil, &activeTheme, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *tableCheckboxBenchmarkFrame) layout(value Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkTableCheckboxCompositeLayout(b *testing.B) {
	frame := newTableCheckboxBenchmarkFrame(image.Pt(720, 360))
	rows := make([]Row, 128)
	for index := range rows {
		rows[index] = Row{
			Key: "row-" + strconv.Itoa(index),
			Cells: []Cell{
				{Text: "Item " + strconv.Itoa(index)},
				{Text: "Group" + strconv.Itoa(index%8)},
			},
		}
	}
	value := New("checkbox-table", []Column{
		{Key: "name", Label: "Name", Width: 280},
		{Key: "group", Label: "Group", Width: 180},
	}, rows).
		SelectionMode(SelectionMultiple).
		ShowSelectionIndicator().
		SelectedKeys([]string{"row-3", "row-17", "row-41"}).
		MaxHeight(320)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	state := tablePeekState(frame.ctx, value.key)
	if state == nil {
		b.Fatal("table state was not initialized")
	}
	stats := state.checkboxVisual.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "checkbox-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "checkbox-cache-misses/op")
}

func BenchmarkTableDefaultVisualCacheHit(b *testing.B) {
	activeTheme := theme.Light()
	var cache tableDefaultsCache
	key := tableDefaultsKey{variant: VariantPrimary}
	cache.resolve(&activeTheme, key)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.resolve(&activeTheme, key)
	}
	stats := cache.statsSnapshot()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "table-defaults-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "table-defaults-cache-misses/op")
}

func tableBenchmarkVirtualValue() Widget {
	return NewVirtual("virtual-benchmark", []Column{
		{Key: "name", Label: "Name", Width: 280, RowHeader: true},
		{Key: "group", Label: "Group", Width: 180},
	}, 10000, func(index int) Row {
		return Row{
			Key: "row-" + strconv.Itoa(index),
			Cells: []Cell{
				{Text: "Item " + strconv.Itoa(index)},
				{Text: "Group" + strconv.Itoa(index%8)},
			},
		}
	}).
		SelectionMode(SelectionSingle).
		SelectedKey("row-3").
		MaxHeight(320)
}

func tableDefaultsStats(frame *tableCheckboxBenchmarkFrame, key string) tableDefaultsCacheStats {
	state := tablePeekState(frame.ctx, key)
	if state == nil {
		return tableDefaultsCacheStats{}
	}
	return state.defaultsCache.statsSnapshot()
}

func reportTableDefaultsStats(b *testing.B, after, before tableDefaultsCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "table-defaults-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "table-defaults-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "table-defaults-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "table-defaults-cache-key-invalidations/op")
}

func BenchmarkTableVirtualComponentTreeSteadyState(b *testing.B) {
	frame := newTableCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := tableBenchmarkVirtualValue()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkTableVirtualComponentTreeVisualCacheHit(b *testing.B) {
	frame := newTableCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := tableBenchmarkVirtualValue()
	frame.layout(value)
	before := tableDefaultsStats(frame, value.key)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportTableDefaultsStats(b, tableDefaultsStats(frame, value.key), before)
}

func BenchmarkTableVirtualComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newTableCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := tableBenchmarkVirtualValue()
	frame.layout(value)
	before := tableDefaultsStats(frame, value.key)
	light := theme.Light()
	dark := theme.Light()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.layout(value)
	}
	b.StopTimer()
	reportTableDefaultsStats(b, tableDefaultsStats(frame, value.key), before)
}

func BenchmarkTableVirtualComponentTreeVisualCacheKeyInvalidation(b *testing.B) {
	frame := newTableCheckboxBenchmarkFrame(image.Pt(720, 360))
	primary := tableBenchmarkVirtualValue()
	secondary := primary.Variant(VariantSecondary)
	frame.layout(primary)
	before := tableDefaultsStats(frame, primary.key)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			frame.layout(primary)
		} else {
			frame.layout(secondary)
		}
	}
	b.StopTimer()
	reportTableDefaultsStats(b, tableDefaultsStats(frame, primary.key), before)
}
