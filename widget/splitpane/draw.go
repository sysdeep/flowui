package splitpane

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

func drawSplitPaneDivider(ctx *core.Context, gtx layout.Context, axis layout.Axis, size image.Point, first, divider int, hovered, dragging bool, focus float32, disabled bool, style splitPaneStyle) {
	dividerRect := splitPaneDividerRect(axis, size, first, divider)
	layoutSplitPanePart(ctx, gtx, dividerRect, style.track, 1)

	active := focus
	if hovered || dragging {
		active = 1
	}
	if active <= 0 || disabled {
		return
	}
	activeWidth := max(gtx.Dp(style.activeWidth), divider)
	handleLength := max(gtx.Dp(style.handleLength), activeWidth)
	handle := splitPaneHandleRect(axis, size, first, divider, activeWidth, handleLength).Intersect(image.Rectangle{Max: size})
	if handle.Empty() {
		return
	}
	layoutSplitPanePart(ctx, gtx, handle, style.indicator, min(max(active, 0), 1))
}

func layoutSplitPanePart(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity float32) {
	if rect.Empty() {
		return
	}
	if resolved.Box != nil && resolved.Box.Cursor != nil {
		box := *resolved.Box
		box.Cursor = nil
		resolved.Box = &box
	}
	partGtx := gtx
	partGtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, partGtx, resolved, nil)
	fade.Pop()
	offset.Pop()
}

func splitPaneDividerRect(axis layout.Axis, size image.Point, first, divider int) image.Rectangle {
	if axis == layout.Vertical {
		return image.Rect(0, first, size.X, first+divider)
	}
	return image.Rect(first, 0, first+divider, size.Y)
}

func splitPaneHandleRect(axis layout.Axis, size image.Point, first, divider, width, length int) image.Rectangle {
	mainCenter := first + divider/2
	crossSize := axis.Convert(size).Y
	length = min(length, crossSize)
	crossStart := (crossSize - length) / 2
	mainStart := mainCenter - width/2
	return axisRect(axis, mainStart, crossStart, width, length)
}

func axisRect(axis layout.Axis, main, cross, mainSize, crossSize int) image.Rectangle {
	minPoint := axis.Convert(image.Pt(main, cross))
	return image.Rectangle{Min: minPoint, Max: minPoint.Add(axis.Convert(image.Pt(mainSize, crossSize)))}
}
