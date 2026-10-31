package radiogroup

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

func drawRadio(ctx *core.Context, gtx layout.Context, defaults radioDefaults, resolved radioStyle, scale float32) layout.Dimensions {
	controlSize := radioIndicatorSize(gtx, resolved)
	focusSpace := max(gtx.Dp(defaults.Metrics.FocusSpace), 1)
	controlSize.X = min(controlSize.X, max(gtx.Constraints.Max.X-focusSpace*2, 0))
	controlSize.Y = min(controlSize.Y, max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := controlSize.Add(image.Pt(focusSpace*2, focusSpace*2))
	dims := gtx.Constraints.Constrain(bounds)
	if controlSize.X <= 0 || controlSize.Y <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-controlSize.X)/2, (dims.Y-controlSize.Y)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(controlSize),
	}
	center := f32.Pt(
		float32(rect.Min.X)+float32(controlSize.X)/2,
		float32(rect.Min.Y)+float32(controlSize.Y)/2,
	)
	stack := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
	// The default selected dot is a glyph over the accent fill. Keep it opaque
	// white instead of inheriting the page foreground used by light themes.
	foreground := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	dot := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return drawRadioIndicator(
			gtx,
			gtx.Constraints.Min,
			foreground,
			resolved.selected,
			resolved.pressed,
			defaults.Metrics.DotScale,
			defaults.Metrics.DotPressedScale,
		)
	})
	layoutRadioLayer(ctx, gtx, rect, resolved.indicatorOff, 1-resolved.selected, resolved.focus, nil)
	layoutRadioLayer(ctx, gtx, rect, resolved.indicatorOn, resolved.selected, resolved.focus, dot)
	stack.Pop()

	return layout.Dimensions{Size: dims}
}

func radioIndicatorSize(gtx layout.Context, resolved radioStyle) image.Point {
	var size image.Point
	for _, endpoint := range [...]style.ResolvedStyle{resolved.indicatorOff, resolved.indicatorOn} {
		if endpoint.Box == nil {
			panic("flowui/radiogroup: missing indicator box")
		}
		width := endpoint.Box.Width
		if width == nil {
			width = endpoint.Box.MinWidth
		}
		height := endpoint.Box.Height
		if height == nil {
			height = endpoint.Box.MinHeight
		}
		if width == nil || height == nil || *width <= 0 || *height <= 0 {
			panic("flowui/radiogroup: missing or non-positive indicator dimensions")
		}
		size.X = max(size.X, gtx.Dp(*width))
		size.Y = max(size.Y, gtx.Dp(*height))
	}
	if size.X <= 0 || size.Y <= 0 {
		panic("flowui/radiogroup: indicator has no pixel area")
	}
	return size
}

func layoutRadioLayer(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity, focus float32, child core.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, layerGtx, computed.ApplyOutlineOpacity(resolved, focus), child)
	fade.Pop()
	offset.Pop()
}
