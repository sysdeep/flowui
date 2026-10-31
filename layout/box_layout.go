package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func (b BoxWidget) layoutInteractiveResolved(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	prepareFieldAssociation(ctx, b.child)
	transformStyle := resolved.Trans
	visualStyle := resolved
	visualStyle.Trans = nil
	visual := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, visualStyle)
	}
	layoutHost := func(gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		if host == nil {
			dims = visual(gtx)
		} else {
			dims = host(gtx, visual)
		}
		return dims
	}
	if !boxTransformIsIdentity(transformStyle) {
		untransformedHost := layoutHost
		layoutHost = func(gtx layout.Context) layout.Dimensions {
			macro := op.Record(gtx.Ops)
			dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return untransformedHost(gtx)
			})
			call := macro.Stop()
			transform := boxTransform(gtx, transformStyle, dims.Size)
			placement.PlaceTransform(transform)
			stack := op.Affine(transform).Push(gtx.Ops)
			call.Add(gtx.Ops)
			stack.Pop()
			return dims
		}
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutHost)
	}
	return layoutHost(gtx)
}

func (b BoxWidget) layoutResolved(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	layoutBox := func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, resolved)
	}
	margin := boxInsets(resolved.Box, false)
	if hasInset(margin) {
		return layoutBoxInset(ctx, gtx, margin, layoutBox)
	}
	return layoutBox(gtx)
}

func (b BoxWidget) layoutVisual(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	b.applyConstraints(&gtx, resolved.Box)
	if core.CollectingVisualOverflow(ctx) {
		core.ReportVisualOverflow(ctx, boxVisualOutset(gtx, resolved.Paint))
	}
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return b.layoutContent(ctx, gtx, resolved)
	})
	content := macro.Stop()

	transform := boxTransform(gtx, resolved.Trans, dims.Size)
	placement.PlaceTransform(transform)
	opacityValue := boxOpacity(resolved.Paint)
	placement.SetOpacity(opacityValue)
	transformStack := op.Affine(transform).Push(gtx.Ops)
	defer transformStack.Pop()
	opacity := paint.PushOpacity(gtx.Ops, opacityValue)
	defer opacity.Pop()

	rect := image.Rectangle{Max: dims.Size}
	shape := boxRRect(gtx, resolved.Paint, rect)
	if resolved.Box != nil && resolved.Box.Cursor != nil {
		cursorClip := shape.Push(gtx.Ops)
		resolved.Box.Cursor.Add(gtx.Ops)
		cursorClip.Pop()
	}
	boxShadows(gtx, rect, resolved.Paint)
	if brush, ok := boxBrush(resolved.Paint); ok {
		draw.FillRRect(gtx, shape, brush)
	}
	content.Add(gtx.Ops)
	boxBorder(gtx, shape, resolved.Paint)
	boxOutline(gtx, shape, resolved.Paint)
	return dims
}

func (b BoxWidget) layoutContent(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {

	layoutChild := func(gtx layout.Context) layout.Dimensions {
		return b.layoutChild(ctx, gtx, resolved)
	}
	if b.hasAlign {
		layoutChild = func(gtx layout.Context) layout.Dimensions {
			return layoutTrackedDirection(ctx, gtx, b.align.Direction(), func(gtx layout.Context) layout.Dimensions {
				return b.layoutChild(ctx, gtx, resolved)
			})
		}
	}
	padding := boxInsets(resolved.Box, true)
	if !hasInset(padding) {
		return b.layoutOverflow(ctx, gtx, resolved.Paint, resolved.Box, layoutChild)
	}
	return b.layoutOverflow(ctx, gtx, resolved.Paint, resolved.Box, func(gtx layout.Context) layout.Dimensions {
		return layoutBoxInset(ctx, gtx, padding, layoutChild)
	})
}

func layoutBoxInset(ctx *Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	horizontal := gtx.Dp(inset.Left) + gtx.Dp(inset.Right)
	vertical := gtx.Dp(inset.Top) + gtx.Dp(inset.Bottom)
	gtx.Constraints.Min.X = max(gtx.Constraints.Min.X-horizontal, 0)
	gtx.Constraints.Min.Y = max(gtx.Constraints.Min.Y-vertical, 0)
	return layoutTrackedInset(ctx, gtx, inset, child)
}

func (b BoxWidget) layoutChild(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle) layout.Dimensions {
	if b.child == nil {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	foreground, background := ctx.ForegroundColor(), ctx.BackgroundColor()
	if resolved.Text != nil {
		if color, ok := computed.Color(resolved.Text.Color); ok {
			foreground = color
		}
	}
	if brush, ok := boxBrush(resolved.Paint); ok {
		if sampled := brush.ColorAt(.5); sampled.A != 0 {
			background = sampled
		}
	}
	restoreColors := core.PushColors(ctx, foreground, background)
	defer restoreColors()
	if resolved.Text != nil {
		restoreStyle := core.PushInheritedStyle(ctx, style.TextDeclaration(resolved.Text))
		defer restoreStyle()
	}
	return b.child.Layout(ctx, gtx)
}

func (b BoxWidget) layoutOverflow(ctx *Context, gtx layout.Context, paintStyle *style.PaintStyle, boxStyle *style.BoxStyle, child layout.Widget) layout.Dimensions {
	if boxStyle == nil || boxStyle.Overflow == nil || *boxStyle.Overflow == style.OverflowVisible {
		return child(gtx)
	}
	// This box owns the clip, so visual overflow from descendants must not be
	// reserved again by an ancestor viewport.
	restoreCollector := core.PushVisualOverflowCollector(ctx, new(core.VisualOverflowCollector))
	defer restoreCollector()
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child(gtx)
	})
	call := macro.Stop()
	dims.Size = gtx.Constraints.Constrain(dims.Size)
	placement.PlaceOffset(image.Point{})
	placement.ClipTo(image.Rectangle{Max: dims.Size})
	shape := boxRRect(gtx, paintStyle, image.Rectangle{Max: dims.Size})
	defer shape.Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	return dims
}

func (b BoxWidget) applyConstraints(gtx *layout.Context, box *style.BoxStyle) {
	if box == nil {
		box = &style.BoxStyle{}
	}
	if box.MaxWidth != nil {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, max(gtx.Dp(*box.MaxWidth), 0))
		gtx.Constraints.Min.X = min(gtx.Constraints.Min.X, gtx.Constraints.Max.X)
	}
	if box.MaxHeight != nil {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, max(gtx.Dp(*box.MaxHeight), 0))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
	if box.MinWidth != nil {
		gtx.Constraints.Min.X = min(max(gtx.Constraints.Min.X, gtx.Dp(*box.MinWidth)), gtx.Constraints.Max.X)
	}
	if box.MinHeight != nil {
		gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(*box.MinHeight)), gtx.Constraints.Max.Y)
	}
	aspectBounds := gtx.Constraints
	if box.FillWidth != nil && *box.FillWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if box.FillHeight != nil && *box.FillHeight {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	if box.Width != nil {
		width := min(max(max(gtx.Dp(*box.Width), 0), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	if box.Height != nil {
		height := min(max(max(gtx.Dp(*box.Height), 0), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height
	}
	applyAspectRatio(gtx, box, aspectBounds)
}

func applyAspectRatio(gtx *layout.Context, box *style.BoxStyle, bounds layout.Constraints) {
	if box.AspectRatio == nil || *box.AspectRatio <= 0 {
		return
	}
	widthFixed := box.Width != nil || box.FillWidth != nil && *box.FillWidth
	heightFixed := box.Height != nil || box.FillHeight != nil && *box.FillHeight
	if widthFixed == heightFixed {
		return
	}
	ratio := *box.AspectRatio
	if widthFixed {
		width := gtx.Constraints.Max.X
		height := int(float32(width)/ratio + .5)
		if height > bounds.Max.Y {
			height = bounds.Max.Y
			width = int(float32(height)*ratio + .5)
		}
		height = min(max(height, bounds.Min.Y), bounds.Max.Y)
		width = min(max(width, bounds.Min.X), bounds.Max.X)
		gtx.Constraints.Min = image.Pt(width, height)
		gtx.Constraints.Max = gtx.Constraints.Min
		return
	}
	height := gtx.Constraints.Max.Y
	width := int(float32(height)*ratio + .5)
	if width > bounds.Max.X {
		width = bounds.Max.X
		height = int(float32(width)/ratio + .5)
	}
	width = min(max(width, bounds.Min.X), bounds.Max.X)
	height = min(max(height, bounds.Min.Y), bounds.Max.Y)
	gtx.Constraints.Min = image.Pt(width, height)
	gtx.Constraints.Max = gtx.Constraints.Min
}

func boxInsets(box *style.BoxStyle, padding bool) layout.Inset {
	if box == nil {
		return layout.Inset{}
	}
	value := box.Margin
	if padding {
		value = box.Padding
	}
	if value == nil {
		return layout.Inset{}
	}
	return layout.Inset{
		Top:    value.Top,
		Right:  value.Right,
		Bottom: value.Bottom,
		Left:   value.Left,
	}
}

func hasInset(value layout.Inset) bool {
	return value.Top != 0 || value.Right != 0 || value.Bottom != 0 || value.Left != 0
}
