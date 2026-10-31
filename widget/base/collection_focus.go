package base

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

// The group's single keyboard target must be mounted but never intercept row
// pointers. Navigation consumes its keys before Clickable lays out the target.
func layoutCollectionFocus(ctx *core.Context, gtx layout.Context, focus *widget.Clickable) {
	core.RegisterFocusGroupItem(ctx, focus, gtx.Enabled())
	if !gtx.Enabled() {
		return
	}
	offset := op.Offset(image.Pt(-1, -1)).Push(gtx.Ops)
	defer offset.Pop()
	gtx.Constraints = layout.Exact(image.Pt(1, 1))
	focus.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{} })
}
