package tooltip

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	tooltipEnterDuration = 150 * time.Millisecond
	tooltipExitDuration  = 100 * time.Millisecond
	tooltipMoveDuration  = 400 * time.Millisecond
)

// Arrow sides, in the same order the placement code reports them.
const (
	sideTop uint8 = iota
	sideLeft
	sideRight
	sideBottom
)

// tooltipMetrics is the component-owned geometry and timing consumed by layout.
type tooltipMetrics struct {
	Offset            unit.Dp
	ArrowOffset       unit.Dp
	Padding           unit.Dp
	Radius            unit.Dp
	BorderWidth       unit.Dp
	MaxWidth          unit.Dp
	ArrowSize         unit.Dp
	TextSize          unit.Sp
	AnimationScale    float32
	ExitScale         float32
	AnimationDistance unit.Dp
	Delay             time.Duration
	CloseDelay        time.Duration
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	MoveDuration      time.Duration
}

// tooltipPanelColors is the resolved surface and border the arrow takes from
// the panel declaration.
type tooltipPanelColors struct {
	Surface color.NRGBA
	Border  color.NRGBA
}

// tooltipArrowRequest carries the measured facts needed to paint one arrow.
type tooltipArrowRequest struct {
	Side        uint8
	Panel       image.Point
	Anchor      float32
	ArrowSize   int
	BorderWidth int
	Colors      tooltipPanelColors
}

// tooltipVisual is the resolved appearance of one tooltip configuration.
type tooltipVisual struct {
	Defaults   style.Style
	Content    style.Style
	Metrics    tooltipMetrics
	Motion     theme.MotionTheme
	PanelColor func(style.ResolvedStyle) tooltipPanelColors
}

func buildTooltipVisual(active *theme.Theme) tooltipVisual {
	semantic := active.Semantic()
	return tooltipVisual{
		Defaults: style.Style{}.
			Background(style.Solid(semantic.Overlay)).
			TextColor(style.Solid(semantic.OverlayForeground)).
			FontSize(12).
			Padding(8).
			Radius(4).
			BorderColor(style.Solid(semantic.Border)).
			BorderWidth(1).
			Shadow(style.ShadowOverlay).
			Overflow(style.OverflowHidden),
		Content: style.Style{}.
			FontSize(12).
			TextColor(style.Solid(semantic.OverlayForeground)),
		Metrics: tooltipMetrics{
			Offset: 3, ArrowOffset: 7,
			Padding: 8, Radius: 4,
			BorderWidth: 1, MaxWidth: 320,
			ArrowSize: 12, TextSize: 12,
			AnimationScale:    1,
			ExitScale:         1,
			AnimationDistance: 4,
			Delay:             1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond,
			EnterDuration: tooltipEnterDuration,
			ExitDuration:  tooltipExitDuration,
			MoveDuration:  tooltipMoveDuration,
		},
		Motion:     active.Motion,
		PanelColor: tooltipPanelColorsOf(semantic),
	}
}

func tooltipAnimationScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .90
	}
	return value
}

func tooltipPanelColorsOf(semantic theme.Semantic) func(style.ResolvedStyle) tooltipPanelColors {
	return func(resolved style.ResolvedStyle) tooltipPanelColors {
		result := tooltipPanelColors{Surface: semantic.Overlay, Border: semantic.Border}
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

func tooltipArrow(gtx layout.Context, request tooltipArrowRequest) {
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
	case sideTop:
		y := float32(request.Panel.Y) - inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y+depth)
		control1 = f32.Pt(centerX+span*0.041667, y+depth)
		to = f32.Pt(centerX+half, y)
	case sideBottom:
		y := inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y-depth)
		control1 = f32.Pt(centerX+span*0.041667, y-depth)
		to = f32.Pt(centerX+half, y)
	case sideLeft:
		x := float32(request.Panel.X) - inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x+depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x+depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	case sideRight:
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
