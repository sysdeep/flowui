package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestDragCapturesOnePointerAndStopsWhenDisabled(t *testing.T) {
	h := uitest.New(image.Pt(100, 40), theme.Light())
	var drag base.DragInteraction
	var positions []f32.Point
	starts, enabled := 0, true
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		drag.Update(ctx, gtx, &drag, enabled, func(position f32.Point, started bool) {
			positions = append(positions, position)
			if started {
				starts++
			}
		})
		drag.AddInput(gtx, &drag, image.Rect(0, 0, 100, 40), enabled, pointer.CursorPointer, "Axis", "")
		return layout.Dimensions{Size: image.Pt(100, 40)}
	})
	h.Frame(view)
	h.Touch(1, pointer.Press, f32.Pt(10, 10))
	h.Frame(view)
	h.Touch(2, pointer.Press, f32.Pt(80, 10))
	h.Frame(view)
	h.Touch(2, pointer.Release, f32.Pt(90, 10))
	h.Frame(view)
	h.Touch(1, pointer.Move, f32.Pt(130, 10))
	h.Frame(view)
	if starts != 1 || !drag.Dragging || positions[len(positions)-1].X != 130 {
		t.Fatalf("starts=%d dragging=%v positions=%v", starts, drag.Dragging, positions)
	}
	count := len(positions)
	enabled = false
	h.Frame(view)
	h.Touch(1, pointer.Release, f32.Pt(150, 10))
	h.Frame(view)
	if drag.Dragging || len(positions) != count {
		t.Fatalf("disabled drag continued: dragging=%v positions=%v", drag.Dragging, positions)
	}
}
