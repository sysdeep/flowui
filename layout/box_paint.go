package layout

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func boxTransform(gtx layout.Context, transform *style.TransformStyle, size image.Point) f32.Affine2D {
	matrix := f32.AffineId()
	if transform == nil {
		return matrix
	}
	translateX, translateY := float32(0), float32(0)
	if transform.TranslateX != nil {
		translateX = float32(gtx.Dp(*transform.TranslateX))
	}
	if transform.TranslateY != nil {
		translateY = float32(gtx.Dp(*transform.TranslateY))
	}
	matrix = matrix.Offset(f32.Pt(translateX, translateY))
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	scaleX, scaleY := float32(1), float32(1)
	if transform.ScaleX != nil {
		scaleX = *transform.ScaleX
	}
	if transform.ScaleY != nil {
		scaleY = *transform.ScaleY
	}
	matrix = matrix.Scale(center, f32.Pt(scaleX, scaleY))
	if transform.Rotate != nil {
		matrix = matrix.Rotate(center, *transform.Rotate)
	}
	return matrix
}

func boxTransformIsIdentity(transform *style.TransformStyle) bool {
	if transform == nil {
		return true
	}
	return (transform.TranslateX == nil || *transform.TranslateX == 0) &&
		(transform.TranslateY == nil || *transform.TranslateY == 0) &&
		(transform.ScaleX == nil || *transform.ScaleX == 1) &&
		(transform.ScaleY == nil || *transform.ScaleY == 1) &&
		(transform.Rotate == nil || *transform.Rotate == 0)
}

func boxOpacity(paint *style.PaintStyle) float32 {
	if paint == nil || paint.Opacity == nil {
		return 1
	}
	return min(max(*paint.Opacity, 0), 1)
}

func boxBrush(paint *style.PaintStyle) (draw.Brush, bool) {
	if paint == nil {
		return draw.Brush{}, false
	}
	return computed.Brush(paint.Background)
}

func boxRRect(gtx layout.Context, paint *style.PaintStyle, rect image.Rectangle) clip.RRect {
	shape := clip.RRect{Rect: rect}
	if paint == nil {
		return shape
	}
	limit := min(rect.Dx(), rect.Dy()) / 2
	toPixels := func(value unit.Dp) int { return min(max(gtx.Dp(value), 0), limit) }
	radii := boxCornerRadii(paint)
	shape.NW = toPixels(radii.TopLeft)
	shape.NE = toPixels(radii.TopRight)
	shape.SE = toPixels(radii.BottomRight)
	shape.SW = toPixels(radii.BottomLeft)
	return shape
}

func boxCornerRadii(paint *style.PaintStyle) style.CornerRadii {
	var radii style.CornerRadii
	if paint == nil {
		return radii
	}
	if paint.Radius != nil {
		radii = style.CornerRadii{
			TopLeft: *paint.Radius, TopRight: *paint.Radius,
			BottomRight: *paint.Radius, BottomLeft: *paint.Radius,
		}
	}
	if paint.Radii != nil {
		radii = *paint.Radii
	}
	return radii
}

func boxShadows(gtx layout.Context, rect image.Rectangle, paint *style.PaintStyle) {
	if paint == nil || rect.Empty() {
		return
	}
	radii := boxCornerRadii(paint)
	shape := draw.RoundedShadowCorners(radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft)
	for _, shadow := range paint.Shadows {
		col, ok := computed.Color(shadow.Color)
		if !ok {
			continue
		}
		draw.Shadow(gtx, rect, shape, draw.BoxShadow{
			OffsetX: float32(shadow.OffsetX),
			OffsetY: float32(shadow.OffsetY),
			Blur:    float32(shadow.Blur),
			Spread:  float32(shadow.Spread),
			Color:   col,
		})
	}
}

func boxVisualOutset(gtx layout.Context, paint *style.PaintStyle) core.VisualOutset {
	if paint == nil {
		return core.VisualOutset{}
	}
	result := core.VisualOutset{}
	if paint.VisualOutset != nil {
		result = result.Max(core.VisualOutset{
			Top:    max(gtx.Dp(paint.VisualOutset.Top), 0),
			Right:  max(gtx.Dp(paint.VisualOutset.Right), 0),
			Bottom: max(gtx.Dp(paint.VisualOutset.Bottom), 0),
			Left:   max(gtx.Dp(paint.VisualOutset.Left), 0),
		})
	}
	for _, shadow := range paint.Shadows {
		if shadow.Transient {
			continue
		}
		col, ok := computed.Color(shadow.Color)
		if !ok || col.A == 0 {
			continue
		}
		blur := gtx.Dp(shadow.Blur)
		if blur < 0 {
			continue
		}
		padX := draw.ShadowRasterPadding(blur, gtx.Dp(shadow.Spread), gtx.Dp(shadow.OffsetX))
		padY := draw.ShadowRasterPadding(blur, gtx.Dp(shadow.Spread), gtx.Dp(shadow.OffsetY))
		result = result.Max(core.VisualOutset{
			Top:    padY,
			Right:  padX,
			Bottom: padY,
			Left:   padX,
		})
	}
	if paint.Outline != nil {
		width := gtx.Dp(paint.Outline.Width)
		if width > 0 {
			expand := saturatedAdd(width, max(gtx.Dp(paint.Outline.Offset), 0))
			result = result.Max(core.VisualOutset{Top: expand, Right: expand, Bottom: expand, Left: expand})
		}
	}
	return result
}

func saturatedAdd(values ...int) int {
	limit := int(^uint(0) >> 1)
	result := 0
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if result > limit-value {
			return limit
		}
		result += value
	}
	return result
}

func boxBorder(gtx layout.Context, shape clip.RRect, paint *style.PaintStyle) {
	if paint == nil {
		return
	}
	if paint.Border != nil && paint.Border.Width != nil {
		color, ok := computed.Color(paint.Border.Color)
		if ok {
			width := max(gtx.Dp(*paint.Border.Width), 0)
			on, off := style.LineSolid.DashPattern()
			if paint.Border.LineStyle != nil {
				on, off = paint.Border.LineStyle.DashPattern()
			}
			if on > 0 {
				draw.DashedRRectBorder(gtx, shape, width, max(gtx.Dp(on), 1), gtx.Dp(off), color)
			} else {
				draw.RRectBorder(gtx, shape, width, color)
			}
		}
	}
	if paint.BorderBottom != nil && paint.BorderBottom.Width != nil {
		color, ok := computed.Color(paint.BorderBottom.Color)
		if ok {
			clipStack := shape.Push(gtx.Ops)
			draw.BottomBorder(gtx, shape.Rect, max(gtx.Dp(*paint.BorderBottom.Width), 0), color)
			clipStack.Pop()
		}
	}
}

func boxOutline(gtx layout.Context, shape clip.RRect, paint *style.PaintStyle) {
	if paint == nil || paint.Outline == nil || paint.Outline.Width <= 0 {
		return
	}
	col, ok := computed.Color(paint.Outline.Color)
	if !ok {
		return
	}
	width := max(gtx.Dp(paint.Outline.Width), 1)
	gap := max(gtx.Dp(paint.Outline.Offset), 0)
	expand := width + gap
	shape.Rect = shape.Rect.Inset(-expand)
	shape.NW += expand
	shape.NE += expand
	shape.SE += expand
	shape.SW += expand
	draw.RRectBorder(gtx, shape, width, col)
}
