package chart

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
)

func TestPointerAndClicksRespectOffsetBoundsAndDisabledContext(t *testing.T) {
	h := uitest.New(image.Pt(160, 100), theme.Light())
	var tag struct{ _ byte }
	var click gesture.Click
	var position f32.Point
	hovered, disabled, clicks := false, false, 0
	bounds := image.Rect(40, 20, 120, 80)
	view := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		UpdateHoverPointer(gtx, true, &tag, &hovered, &position)
		if activated, _ := UpdateClicks(gtx, true, &click); activated {
			clicks++
		}
		AddPointerInputWithCursor(gtx, bounds, true, pointer.CursorPointer, &tag)
		AddClickInputRect(gtx, bounds, true, &click)
		return layout.Dimensions{Size: image.Pt(160, 100)}
	})
	h.Frame(view)
	h.Click(f32.Pt(10, 10))
	h.Frame(view)
	if hovered || clicks != 0 {
		t.Fatalf("outside hit: hovered=%v clicks=%d", hovered, clicks)
	}
	h.Click(f32.Pt(60, 40))
	h.Frame(view)
	if !hovered || position != f32.Pt(60, 40) || clicks != 1 {
		t.Fatalf("inside hit: hovered=%v position=%v clicks=%d", hovered, position, clicks)
	}
	// Queue input against last frame's enabled regions, then disable the context.
	h.Click(f32.Pt(60, 40))
	disabled = true
	h.Frame(view)
	if hovered || clicks != 1 {
		t.Fatalf("disabled hit: hovered=%v clicks=%d", hovered, clicks)
	}
}

func TestWindowPointerProcessesQueuedZoomAndCancelsPan(t *testing.T) {
	h := uitest.New(image.Pt(100, 100), theme.Light())
	var tag struct{ _ byte }
	var windowGesture DataWindowGesture
	var position f32.Point
	window, hovered, enabled := FullDataWindow(), false, true
	view := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		plot := image.Rect(0, 0, 100, 100)
		UpdatePointer(gtx, enabled, plot, window, false, func(next DataWindow) {
			window = next
		}, &tag, &hovered, &position, &windowGesture)
		AddPointerInput(gtx, plot, enabled, &tag)
		return layout.Dimensions{Size: plot.Size()}
	})
	h.Frame(view)
	for range 2 {
		h.Router().Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse,
			Position: f32.Pt(50, 50), Scroll: f32.Pt(0, -1)})
	}
	h.Frame(view)
	if !closeWindow(window, DataWindow{Start: .18, End: .82}) {
		t.Fatalf("queued zoom window = %+v, want [.18, .82]", window)
	}
	h.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse,
		PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(50, 50)})
	h.Frame(view)
	h.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse,
		PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(60, 50)})
	h.Frame(view)
	if !windowGesture.Dragging() {
		t.Fatal("pointer movement did not start panning")
	}
	enabled = false
	h.Frame(view)
	if hovered || windowGesture.Dragging() {
		t.Fatal("disabled chart retained hover or active panning")
	}
}
