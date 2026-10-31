package base_test

import (
	"image"
	"testing"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestSplitPaneInteractionKeyboardLimits(t *testing.T) {
	for _, tc := range []struct {
		name string
		axis layout.Axis
		key  key.Name
		want float32
	}{
		{name: "horizontal right", axis: layout.Horizontal, key: key.NameRightArrow, want: .52},
		{name: "vertical down", axis: layout.Vertical, key: key.NameDownArrow, want: .52},
		{name: "home minimum", axis: layout.Horizontal, key: key.NameHome, want: .2},
		{name: "end maximum", axis: layout.Horizontal, key: key.NameEnd, want: .8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := uitest.New(image.Pt(100, 60), theme.Light())
			var interaction base.SplitPaneInteraction
			var ratio float32
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				ratio, _ = interaction.Update(ctx, gtx, base.SplitPaneGeometry{
					Axis: tc.axis, Available: 100, First: 50, Divider: 1, Minimum: 20, Maximum: 80,
				}, false)
				interaction.AddHandle(gtx, image.Rect(0, 0, 100, 60), "Resize", false, pointer.CursorDefault)
				return layout.Dimensions{Size: image.Pt(100, 60)}
			})
			h.Frame(view)
			h.Router().Source().Execute(key.FocusCmd{Tag: &interaction})
			h.Router().Queue(key.Event{Name: tc.key, State: key.Press})
			h.Frame(view)
			if ratio != tc.want {
				t.Fatalf("ratio = %v, want %v", ratio, tc.want)
			}
		})
	}
}

func TestSplitPaneInteractionCancelsUnavailableDrag(t *testing.T) {
	for _, tc := range []struct {
		name     string
		disabled bool
		geometry base.SplitPaneGeometry
	}{
		{name: "disabled", disabled: true, geometry: base.SplitPaneGeometry{Available: 100, First: 50, Maximum: 100}},
		{name: "no space", geometry: base.SplitPaneGeometry{}},
		{name: "fixed bounds", geometry: base.SplitPaneGeometry{Available: 100, First: 50, Minimum: 50, Maximum: 50}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interaction := base.SplitPaneInteraction{Hovered: true, Dragging: true}
			_, changed := interaction.Update(nil, layout.Context{}, tc.geometry, tc.disabled)
			if changed || interaction.Hovered || interaction.Dragging {
				t.Fatalf("unavailable handle retained input state: %+v", interaction)
			}
		})
	}
}
