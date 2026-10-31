package toggle

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

type switchBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newSwitchBenchmarkFrame(maximum image.Point) *switchBenchmarkFrame {
	activeTheme := new(theme.Light())
	frame := &switchBenchmarkFrame{
		ctx: core.New(nil, activeTheme, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *switchBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkSwitchLayoutSteadyState(b *testing.B) {
	frame := newSwitchBenchmarkFrame(image.Pt(320, 80))
	value := NewSwitch("Notifications").Key("notifications").Checked(false).Description("Supporting text")
	frame.layout(value)
	state, ok := core.PeekState[switchState](frame.ctx, core.FullKey(frame.ctx, value.key), switchStateSlot)
	if !ok {
		b.Fatal("switch state was not retained after initial layout")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	stats := state.styleCache.stats()
	if b.N > 0 {
		b.ReportMetric(float64(stats.Hits)/float64(b.N), "switch-visual-cache-hits/op")
		b.ReportMetric(float64(stats.Misses)/float64(b.N), "switch-visual-cache-misses/op")
		b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "switch-visual-cache-theme-invalidations/op")
		b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "switch-visual-cache-key-invalidations/op")
	}
}

func BenchmarkSwitchComponentTreeSteadyState(b *testing.B) {
	frame := newSwitchBenchmarkFrame(image.Pt(640, 900))
	children := make([]layout.FlexChild, 128)
	for index := range children {
		value := NewSwitch("Item").Key(fmt.Sprintf("switch-%d", index)).Checked(index%3 == 0)
		if index%7 == 0 {
			value = value.Description("Supporting text")
		}
		if index%11 == 0 {
			value = value.Disabled(true)
		}
		children[index] = layout.Rigid(func(value SwitchWidget) func(layout.Context) layout.Dimensions {
			return func(gtx layout.Context) layout.Dimensions {
				return value.Layout(frame.ctx, gtx)
			}
		}(value))
	}
	root := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx, children...)
	})
	frame.layout(root)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(root)
	}
}

func BenchmarkSwitchVisualCacheHit(b *testing.B) {
	activeTheme := new(theme.Light())
	var cache switchVisualCache
	cache.resolve(activeTheme, SwitchMedium)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.resolve(activeTheme, SwitchMedium)
	}
	b.ReportMetric(1, "switch-visual-cache-hits/op")
	b.ReportMetric(0, "switch-visual-cache-misses/op")
	b.ReportMetric(0, "switch-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "switch-visual-cache-key-invalidations/op")
}

func BenchmarkSwitchVisualCacheMiss(b *testing.B) {
	activeTheme := new(theme.Light())
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache switchVisualCache
		size := SwitchMedium
		if index&1 != 0 {
			size = SwitchLarge
		}
		_ = cache.resolve(activeTheme, size)
	}
	b.ReportMetric(0, "switch-visual-cache-hits/op")
	b.ReportMetric(1, "switch-visual-cache-misses/op")
	b.ReportMetric(0, "switch-visual-cache-theme-invalidations/op")
	b.ReportMetric(1, "switch-visual-cache-key-invalidations/op")
}
