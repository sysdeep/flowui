package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

// Viewport clips child to the incoming maximum size and reserves its reported
// visual overflow inside that viewport. Keep visual between frames. Child layout
// runs once, and overlay anchors follow the inset and viewport clip.
func Viewport(ctx *Context, gtx layout.Context, visual *VisualOutsetState, child Widget) layout.Dimensions {
	size := gtx.Constraints.Constrain(gtx.Constraints.Max)
	if child == nil {
		return layout.Dimensions{Size: size}
	}
	prepareFieldAssociation(ctx, child)
	outset := core.VisualOutset{}
	if visual != nil {
		outset = visual.outset
	}
	left := max(outset.Left, 0)
	top := max(outset.Top, 0)
	right := max(outset.Right, 0)
	bottom := max(outset.Bottom, 0)
	childGtx := gtx
	childGtx.Constraints = layout.Exact(image.Pt(
		max(size.X-visualOutsetSum(left, right), 0),
		max(size.Y-visualOutsetSum(top, bottom), 0),
	))
	collector := new(core.VisualOverflowCollector)
	restoreCollector := core.PushVisualOverflowCollector(ctx, collector)
	collecting := core.CollectingVisualOverflow(ctx)
	macro := op.Record(gtx.Ops)
	_, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child.Layout(ctx, childGtx)
	})
	call := macro.Stop()
	restoreCollector()
	changed := collecting && visual != nil && visual.observe(collector.Outset(), true)
	innerOffset := image.Pt(left, top)
	placement.PlaceOffset(innerOffset)
	placement.ClipTo(image.Rectangle{Max: size})
	clipped := clip.Rect{Max: size}.Push(gtx.Ops)
	inset := op.Offset(innerOffset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	inset.Pop()
	clipped.Pop()
	if changed && ctx != nil {
		ctx.Invalidate()
	}
	return layout.Dimensions{Size: size}
}
