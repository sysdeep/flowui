// Package tooltipstyle contains theme-owned visual plans for tooltip
// behavior. It does not import widget/tooltip: the widget supplies placement
// facts and executes the returned declarations and renderer.
package overlaystyle

import (
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	tooltipTooltipEnterDuration = 150 * time.Millisecond
	tooltipTooltipExitDuration  = 100 * time.Millisecond
	tooltipTooltipMoveDuration  = 400 * time.Millisecond
)

// BuildTooltip returns the complete tooltip visual protocol for the active
// theme. All values are copied so the provider does not retain mutable theme
// state beyond the current visual plan.
func BuildTooltip(active *theme.Theme, request theme.TooltipRequest) theme.TooltipVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Tooltip
	sem := active.Semantic()
	return theme.TooltipVisual{
		Defaults: style.Style{}.
			Background(style.Solid(sem.Overlay)).
			TextColor(style.Solid(sem.OverlayForeground)).
			FontSize(tokens.TextSize).
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(tokens.BorderWidth).
			Shadow(style.ShadowOverlay).
			Overflow(style.OverflowHidden),
		Content: style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.OverlayForeground)),
		Override: theme.TooltipOverrideStyle(request),
		Metrics: theme.TooltipMetrics{
			Offset: tokens.Offset, ArrowOffset: tokens.ArrowOffset,
			Padding: tokens.Padding, Radius: tokens.Radius,
			BorderWidth: tokens.BorderWidth, MaxWidth: tokens.MaxWidth,
			ArrowSize: tokens.ArrowSize, TextSize: tokens.TextSize,
			AnimationScale:    tooltipAnimationScale(tokens.AnimationScale),
			ExitScale:         tooltipAnimationScale(tokens.ExitScale),
			AnimationDistance: tokens.AnimationDistance,
			Delay:             tokens.Delay, CloseDelay: tokens.CloseDelay,
			EnterDuration: tooltipTooltipEnterDuration,
			ExitDuration:  tooltipTooltipExitDuration,
			MoveDuration:  tooltipTooltipMoveDuration,
		},
		Motion:     active.Motion,
		Arrow:      tooltipTooltipArrowRenderer,
		PanelColor: tooltipTooltipPanelColors(sem),
	}
}

func tooltipAnimationScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .90
	}
	return value
}

func tooltipTooltipPanelColors(sem theme.Semantic) theme.TooltipPanelColorResolver {
	return func(resolved style.ResolvedStyle) theme.TooltipPanelColors {
		result := theme.TooltipPanelColors{Surface: sem.Overlay, Border: sem.Border}
		if resolved.Paint == nil {
			return result
		}
		if value, ok := tooltipSolidColor(resolved.Paint.Background); ok {
			result.Surface = value
		}
		if resolved.Paint.Border != nil {
			if value, ok := tooltipSolidColor(resolved.Paint.Border.Color); ok {
				result.Border = value
			}
		}
		return result
	}
}

func tooltipSolidColor(source any) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

func tooltipTooltipArrowRenderer(gtx layout.Context, request theme.TooltipArrowRequest) {
	if request.ArrowSize <= 0 || request.Panel.X <= 0 || request.Panel.Y <= 0 {
		return
	}
	span := float32(request.ArrowSize)
	depth := span * 2 / 3
	half := span / 2
	inset := tooltipMinFloat(float32(tooltipMaxInt(request.BorderWidth, 0)), depth)
	centerX := request.Anchor
	centerY := request.Anchor
	var from, control0, control1, to f32.Point
	switch request.Side {
	case theme.TooltipSideTop:
		y := float32(request.Panel.Y) - inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y+depth)
		control1 = f32.Pt(centerX+span*0.041667, y+depth)
		to = f32.Pt(centerX+half, y)
	case theme.TooltipSideBottom:
		y := inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y-depth)
		control1 = f32.Pt(centerX+span*0.041667, y-depth)
		to = f32.Pt(centerX+half, y)
	case theme.TooltipSideLeft:
		x := float32(request.Panel.X) - inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x+depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x+depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	case theme.TooltipSideRight:
		x := inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x-depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x-depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	default:
		return
	}

	var fillPath clip.Path
	fillPath.Begin(gtx.Ops)
	fillPath.MoveTo(from)
	fillPath.CubeTo(control0, control1, to)
	fillPath.Close()
	paint.FillShape(gtx.Ops, request.Colors.Surface, clip.Outline{Path: fillPath.End()}.Op())
	if request.BorderWidth <= 0 || request.Colors.Border.A == 0 {
		return
	}
	var curve clip.Path
	curve.Begin(gtx.Ops)
	curve.MoveTo(from)
	curve.CubeTo(control0, control1, to)
	stroke := clip.Stroke{Path: curve.End(), Width: float32(request.BorderWidth)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Colors.Border)
	stroke.Pop()
}

func tooltipMaxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}

func tooltipMinFloat(value, limit float32) float32 {
	if value < limit {
		return value
	}
	return limit
}
