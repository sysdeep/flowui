package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestHoverPassesClicksAndClearsWhenDisabled(t *testing.T) {
	h := uitest.New(image.Pt(100, 60), theme.Light())
	var hover base.HoverInteraction
	var child widget.Clickable
	clicks, enabled := 0, true
	view := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		hover.Update(gtx, &hover, enabled)
		for child.Clicked(gtx) {
			clicks++
		}
		dims := child.Layout(gtx, func(layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(80, 40)}
		})
		hover.AddInput(gtx, &hover, image.Rectangle{Max: dims.Size}, enabled, nil)
		return dims
	})
	h.Frame(view)
	h.Click(f32.Pt(20, 20))
	h.Frame(view)
	if !hover.Hovered || clicks != 1 {
		t.Fatalf("hovered=%v clicks=%d; want true, 1", hover.Hovered, clicks)
	}
	enabled = false
	h.Frame(view)
	if hover.Hovered {
		t.Fatal("disabled observer kept its hover state")
	}
	enabled = true
	h.Frame(view)
	h.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(90, 50)})
	h.Frame(view)
	if hover.Hovered {
		t.Fatal("observer remained hovered outside its bounds")
	}
}
