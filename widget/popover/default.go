package popover

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// popoverMetrics contains the geometry and timing used by popover layout. The
// popover owns these values; the theme only supplies the tokens they come from.
type popoverMetrics struct {
	Offset            unit.Dp
	MaxWidth          unit.Dp
	ArrowWidth        unit.Dp
	ArrowHeight       unit.Dp
	BodyTextSize      unit.Sp
	SectionGap        unit.Dp
	AnimationScale    float32
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	AnimationDistance unit.Dp
}

// popoverVisual is the resolved appearance of one popover configuration.
type popoverVisual struct {
	Defaults style.Style
	Heading  style.Style
	Content  style.Style
	Metrics  popoverMetrics
	Motion   theme.MotionTheme
	Surface  color.NRGBA
}

const (
	popoverDefaultOffset            unit.Dp       = 8
	popoverDefaultPadding           unit.Dp       = 16
	popoverDefaultRadius            unit.Dp       = 8
	popoverDefaultMaxWidth          unit.Dp       = 320
	popoverDefaultArrowWidth        unit.Dp       = 12
	popoverDefaultArrowHeight       unit.Dp       = 7
	popoverDefaultHeadingSize       unit.Sp       = 14
	popoverDefaultBodyTextSize      unit.Sp       = 14
	popoverDefaultSectionGap        unit.Dp       = 8
	popoverDefaultAnimationScale    float32       = 1
	popoverDefaultEnterDuration     time.Duration = 150 * time.Millisecond
	popoverDefaultExitDuration      time.Duration = 100 * time.Millisecond
	popoverDefaultAnimationDistance unit.Dp       = 4
)

func buildPopoverVisual(active *theme.Theme) popoverVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	return popoverVisual{
		Defaults: style.Style{}.
			Background(style.Solid(sem.Overlay)).
			TextColor(style.Solid(sem.OverlayForeground)).
			Padding(popoverDefaultPadding).
			Radius(popoverDefaultRadius).
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(1).
			Shadow(style.ShadowOverlay).
			Overflow(style.OverflowHidden),
		Heading: style.Style{}.
			FontSize(popoverDefaultHeadingSize).
			FontWeight(500).
			TextColor(style.Solid(sem.OverlayForeground)),
		Content: style.Style{}.
			FontSize(popoverDefaultBodyTextSize).
			TextColor(style.Solid(sem.MutedForeground)),
		Metrics: popoverMetrics{
			Offset: popoverDefaultOffset, MaxWidth: popoverDefaultMaxWidth,
			ArrowWidth: popoverDefaultArrowWidth, ArrowHeight: popoverDefaultArrowHeight,
			BodyTextSize: popoverDefaultBodyTextSize, SectionGap: popoverDefaultSectionGap,
			AnimationScale:    popoverDefaultAnimationScale,
			EnterDuration:     popoverDefaultEnterDuration,
			ExitDuration:      popoverDefaultExitDuration,
			AnimationDistance: popoverDefaultAnimationDistance,
		},
		Motion:  active.Motion,
		Surface: sem.Overlay,
	}
}

// popoverArrowColor resolves the solid panel color the arrow should borrow. A
// non-solid or missing background falls back to the semantic overlay color.
func popoverArrowColor(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Paint != nil {
		switch value := resolved.Paint.Background.(type) {
		case style.SolidColor:
			return value.Color
		case *style.SolidColor:
			if value != nil {
				return value.Color
			}
		}
	}
	return fallback
}

func popoverArrow(
	gtx layout.Context,
	side core.Side,
	panel, arrow image.Point,
	resolved style.ResolvedStyle,
	fallback color.NRGBA,
) {
	fill := popoverArrowColor(resolved, fallback)
	if arrow.X <= 0 || arrow.Y <= 0 || fill.A == 0 {
		return
	}
	var border color.NRGBA
	var width int
	if resolved.Paint != nil && resolved.Paint.Border != nil {
		declaration := resolved.Paint.Border
		if declaration.Width != nil {
			width = max(gtx.Dp(*declaration.Width), 0)
		}
		switch value := declaration.Color.(type) {
		case style.SolidColor:
			border = value.Color
		case *style.SolidColor:
			if value != nil {
				border = value.Color
			}
		}
	}
	inset := float32(min(width, arrow.Y))
	var from, tip, to f32.Point
	switch side {
	case core.SideTop:
		x := float32(panel.X) / 2
		y := float32(panel.Y)
		from = f32.Pt(x-float32(arrow.X)/2, y-inset)
		tip = f32.Pt(x, y+float32(arrow.Y))
		to = f32.Pt(x+float32(arrow.X)/2, y-inset)
	case core.SideLeft:
		x := float32(panel.X)
		y := float32(panel.Y) / 2
		from = f32.Pt(x-inset, y-float32(arrow.X)/2)
		tip = f32.Pt(x+float32(arrow.Y), y)
		to = f32.Pt(x-inset, y+float32(arrow.X)/2)
	case core.SideRight:
		y := float32(panel.Y) / 2
		from = f32.Pt(inset, y-float32(arrow.X)/2)
		tip = f32.Pt(-float32(arrow.Y), y)
		to = f32.Pt(inset, y+float32(arrow.X)/2)
	default:
		x := float32(panel.X) / 2
		from = f32.Pt(x-float32(arrow.X)/2, inset)
		tip = f32.Pt(x, -float32(arrow.Y))
		to = f32.Pt(x+float32(arrow.X)/2, inset)
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(tip)
	path.LineTo(to)
	outline := path.End()
	paint.FillShape(gtx.Ops, fill, clip.Outline{Path: outline}.Op())
	if width > 0 && border.A > 0 {
		paint.FillShape(gtx.Ops, border, clip.Stroke{Path: outline, Width: float32(width)}.Op())
	}
}
