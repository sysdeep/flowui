package layout

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
)

func TestViewportInsetsAndClipsOverlayAnchors(t *testing.T) {
	ctx := newContext(nil)
	core.BeginFrameWithViewport(ctx, image.Pt(200, 100))
	defer core.EndFrame(ctx)
	gtx := layout.Context{Constraints: layout.Exact(image.Pt(100, 60)), Ops: new(op.Ops)}
	visual := VisualOutsetState{outset: core.VisualOutset{Top: 3, Right: 7, Bottom: 9, Left: 5}}
	var got image.Rectangle
	var hiddenCalled bool
	var constraints layout.Constraints
	var calls int
	child := WidgetFunc(func(ctx *Context, gtx layout.Context) layout.Dimensions {
		calls++
		constraints = gtx.Constraints
		core.ReportVisualOverflow(ctx, visual.outset)
		(&overlayProbeWidget{
			key: "hidden-viewport-child", size: gtx.Constraints.Min,
			anchor:  image.Rect(150, 0, 170, 10),
			capture: func(image.Rectangle) { hiddenCalled = true },
		}).Layout(ctx, gtx)
		return (&overlayProbeWidget{
			key: "viewport-child", size: gtx.Constraints.Min,
			anchor: image.Rect(0, 0, 200, 100), got: &got,
		}).Layout(ctx, gtx)
	})
	dims := Viewport(ctx, gtx, &visual, child)
	core.LayoutOverlays(ctx, gtx)
	if dims.Size != image.Pt(100, 60) || constraints != layout.Exact(image.Pt(88, 48)) || calls != 1 {
		t.Fatalf("viewport size/constraints/calls = %v/%v/%d", dims.Size, constraints, calls)
	}
	// Partial visibility retains the full anchor for positioning; only fully
	// clipped anchors are suppressed by the shared overlay contract.
	if want := image.Rect(5, 3, 205, 103); got != want || hiddenCalled {
		t.Fatalf("overlay anchor/hidden = %v/%t, want %v/false", got, hiddenCalled, want)
	}
}
