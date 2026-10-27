package toggle

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type checkboxBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newCheckboxBenchmarkFrame(maximum image.Point) *checkboxBenchmarkFrame {
	activeTheme := testtheme.Behavior()
	frame := &checkboxBenchmarkFrame{
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

func (f *checkboxBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkCheckboxLayoutSteadyState(b *testing.B) {
	frame := newCheckboxBenchmarkFrame(image.Pt(320, 80))
	value := NewCheckbox("Accept terms").Key("terms").Checked(false).Description("Required").Required(true)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkCheckboxComponentTreeSteadyState(b *testing.B) {
	frame := newCheckboxBenchmarkFrame(image.Pt(640, 900))
	children := make([]layout.FlexChild, 128)
	for index := range children {
		value := NewCheckbox("Item").Key(fmt.Sprintf("item-%d", index)).Checked(index%3 == 0).Variant(CheckboxPrimary)
		if index%7 == 0 {
			value = value.Description("Supporting text")
		}
		if index%11 == 0 {
			value = value.Disabled(true)
		}
		children[index] = layout.Rigid(func(value CheckboxWidget) func(layout.Context) layout.Dimensions {
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

func BenchmarkCheckboxVisualCacheHit(b *testing.B) {
	activeTheme := testtheme.Behavior()
	var cache CheckboxVisualCache
	cache.Resolve(activeTheme, CheckboxPrimary)
	baseline := cache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.Resolve(activeTheme, CheckboxPrimary)
	}
	b.StopTimer()
	reportCheckboxVisualCacheStats(b, cache.Stats(), baseline)
}

func BenchmarkCheckboxVisualCacheMiss(b *testing.B) {
	activeTheme := testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache CheckboxVisualCache
		variant := CheckboxPrimary
		if index&1 != 0 {
			variant = CheckboxSecondary
		}
		_ = cache.Resolve(activeTheme, variant)
	}
}

func reportCheckboxVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "checkbox-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "checkbox-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "checkbox-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "checkbox-visual-cache-key-invalidations/op")
}

func BenchmarkCheckboxVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	var cache CheckboxVisualCache
	cache.Resolve(first, CheckboxPrimary)
	baseline := cache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			cache.Resolve(first, CheckboxPrimary)
		case 1:
			cache.Resolve(first, CheckboxSecondary)
		case 2:
			cache.Resolve(second, CheckboxSecondary)
		default:
			cache.Resolve(second, CheckboxPrimary)
		}
	}
	b.StopTimer()
	reportCheckboxVisualCacheStats(b, cache.Stats(), baseline)
}
