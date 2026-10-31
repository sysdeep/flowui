package slider

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/text"
)

func (s Widget) layoutOutput(ctx *core.Context, gtx layout.Context, style sliderStyle, output string) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, style.label, text.New(output))
}

func sliderThumbSize(gtx layout.Context, resolved style.ResolvedStyle) image.Point {
	if resolved.Box == nil {
		panic("flowui: slider visual is missing a thumb box")
	}
	width := resolved.Box.Width
	if width == nil {
		width = resolved.Box.MinWidth
	}
	height := resolved.Box.Height
	if height == nil {
		height = resolved.Box.MinHeight
	}
	if width == nil || height == nil {
		panic("flowui: slider visual is missing a thumb size")
	}
	return image.Pt(max(gtx.Dp(*width), 1), max(gtx.Dp(*height), 1))
}

func (s Widget) layoutFill(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, geometry sliderGeometry, rangeMode bool, trackRadius unit.Dp) {
	rect := sliderFillRect(geometry, rangeMode)
	if rect.Empty() {
		return
	}
	fillGtx := gtx
	fillGtx.Constraints = layout.Exact(rect.Size())
	radius := min(max(gtx.Dp(trackRadius), 0), min(geometry.size.X, geometry.size.Y)/2)
	clipped := clip.UniformRRect(image.Rectangle{Max: geometry.size}, radius).Push(gtx.Ops)
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	layoutui.Resolved(ctx, fillGtx, resolved, nil)
	stack.Pop()
	clipped.Pop()
}

func (s Widget) layoutThumbVisual(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, scale float32) {
	outer, inner, inset, layered := sliderThumbLayers(resolved)
	if !layered {
		center := f32.Pt(float32(gtx.Constraints.Min.X)/2, float32(gtx.Constraints.Min.Y)/2)
		transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
		layoutui.Resolved(ctx, gtx, resolved, nil)
		transform.Pop()
		return
	}
	layoutui.Resolved(ctx, gtx, outer, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		pixels := min(max(gtx.Dp(inset), 0), min(size.X, size.Y)/2)
		rect := image.Rectangle{Min: image.Pt(pixels, pixels), Max: size.Sub(image.Pt(pixels, pixels))}
		if !rect.Empty() {
			center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
			transform := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
			stack := op.Offset(rect.Min).Push(gtx.Ops)
			innerGtx := gtx
			innerGtx.Constraints = layout.Exact(rect.Size())
			layoutui.Resolved(ctx, innerGtx, inner, nil)
			stack.Pop()
			transform.Pop()
		}
		return layout.Dimensions{Size: size}
	}))
}

func sliderThumbLayers(resolved style.ResolvedStyle) (outer, inner style.ResolvedStyle, inset unit.Dp, ok bool) {
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width <= 0 || resolved.Paint.Border.Color == nil {
		return resolved, style.ResolvedStyle{}, 0, false
	}
	inset = *resolved.Paint.Border.Width
	outer = resolved
	outerPaint := *resolved.Paint
	outerPaint.Background = outerPaint.Border.Color
	outerPaint.Border = nil
	outerPaint.Shadows = nil
	outer.Paint = &outerPaint

	innerPaint := style.PaintStyle{Background: resolved.Paint.Background, Shadows: resolved.Paint.Shadows}
	if resolved.Paint.Radius != nil {
		radius := max(*resolved.Paint.Radius-inset, 0)
		innerPaint.Radius = &radius
	}
	if resolved.Paint.Radii != nil {
		radii := *resolved.Paint.Radii
		radii.TopLeft = max(radii.TopLeft-inset, 0)
		radii.TopRight = max(radii.TopRight-inset, 0)
		radii.BottomRight = max(radii.BottomRight-inset, 0)
		radii.BottomLeft = max(radii.BottomLeft-inset, 0)
		innerPaint.Radii = &radii
	}
	inner.Paint = &innerPaint
	return outer, inner, inset, true
}
