package titlebar

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	titleBarDefaultHeight         unit.Dp = 35
	titleBarDefaultPaddingX       unit.Dp = 8
	titleBarDefaultLeadingGap     unit.Dp = 8
	titleBarDefaultControlWidth   unit.Dp = 46
	titleBarDefaultIconSize       unit.Dp = 12
	titleBarDefaultIconStroke     unit.Dp = 1.25
	titleBarDefaultTextSize       unit.Sp = 12
	titleBarDefaultBorderWidth    unit.Dp = 1
	titleBarDefaultFocusRingWidth unit.Dp = 2
)

type titleBarControlGlyph uint8

const (
	titleBarGlyphNone titleBarControlGlyph = iota
	titleBarGlyphMinimize
	titleBarGlyphMaximize
	titleBarGlyphRestore
	titleBarGlyphClose
)

// titleBarControlRenderRequest supplies layout and interaction facts to the
// window-control drawing code.
type titleBarControlRenderRequest struct {
	Size         image.Point
	Glyph        titleBarControlGlyph
	Foreground   color.NRGBA
	Focus        color.NRGBA
	FocusVisible bool
}

// titleBarVisual is the resolved default appearance of a title bar.
type titleBarVisual struct {
	Defaults          style.Style
	Focus             color.NRGBA
	ControlStyle      style.Style
	CloseControlStyle style.Style
	Metrics           titleBarControlMetrics
}

type titleBarControlMetrics struct {
	controlWidth   unit.Dp
	iconSize       unit.Dp
	iconStroke     unit.Dp
	focusRingWidth unit.Dp
}

// buildTitleBar builds the component-owned title-bar appearance.
func buildTitleBar(active *theme.Theme) titleBarVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	metrics := titleBarControlMetrics{
		controlWidth:   titleBarDefaultControlWidth,
		iconSize:       titleBarDefaultIconSize,
		iconStroke:     titleBarDefaultIconStroke,
		focusRingWidth: titleBarDefaultFocusRingWidth,
	}
	controlStyle := titleBarControlDeclaration(active, false)
	closeControlStyle := titleBarControlDeclaration(active, true)
	defaults := style.Style{}.
		Height(titleBarDefaultHeight).
		Padding(0, 0, 0, titleBarDefaultPaddingX).
		Background(style.Solid(sem.SurfaceSecondary)).
		TextColor(style.Solid(sem.SurfaceSecondaryForeground)).
		Part(style.PartPrefix, style.Style{}.Padding(0, titleBarDefaultLeadingGap)).
		Part(style.PartLabel, style.Style{}.FontSize(titleBarDefaultTextSize)).
		Part(style.PartIndicator, style.Style{}.
			Height(titleBarDefaultBorderWidth).
			Background(style.Solid(sem.Separator)))
	return titleBarVisual{
		Defaults:          defaults,
		Focus:             sem.Ring,
		ControlStyle:      controlStyle,
		CloseControlStyle: closeControlStyle,
		Metrics:           metrics,
	}
}

// titleBarControlDeclaration returns one component-owned window-control declaration.
func titleBarControlDeclaration(active *theme.Theme, closeAction bool) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	control := style.Style{}.
		Width(titleBarDefaultControlWidth).
		FillHeight()
	sem := active.Semantic()
	if closeAction {
		return style.Style{}.Part(style.PartItem, control).
			Rules(
				style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(sem.Tone(theme.ToneDanger).Hover)).
					TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
				style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(sem.Tone(theme.ToneDanger).Fill)).
					TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
			)
	}
	return style.Style{}.Part(style.PartItem, control).
		Rules(
			style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem,
				style.Style{}.Background(style.Solid(sem.SurfaceTertiary)))},
			style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem,
				style.Style{}.Background(style.Solid(sem.SurfacePressed)))},
		)
}

func titleBarRenderControl(gtx layout.Context, request titleBarControlRenderRequest, metrics titleBarControlMetrics) {
	titleBarDrawControlGlyph(gtx, request.Size, request.Glyph, request.Foreground, metrics)
	if !request.FocusVisible || request.Focus.A == 0 {
		return
	}
	width := max(gtx.Dp(metrics.focusRingWidth), 1)
	ring := image.Rectangle{Max: request.Size}.Inset(width)
	if ring.Empty() {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(ring, 0).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Focus)
	stroke.Pop()
}

func titleBarDrawControlGlyph(gtx layout.Context, size image.Point, glyph titleBarControlGlyph, foreground color.NRGBA, metrics titleBarControlMetrics) {
	if glyph == titleBarGlyphNone || foreground.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	glyphSize := min(max(gtx.Dp(metrics.iconSize), 1), min(size.X, size.Y))
	stroke := titleBarControlGlyphStroke(gtx.Metric, metrics.iconStroke, glyphSize)
	left := (float32(size.X)-float32(glyphSize))/2 + stroke
	top := (float32(size.Y)-float32(glyphSize))/2 + stroke
	inner := titleBarControlGlyphBounds{left: left, top: top, right: left + float32(glyphSize) - stroke*2, bottom: top + float32(glyphSize) - stroke*2}
	if inner.empty() {
		return
	}
	switch glyph {
	case titleBarGlyphMinimize:
		y := (inner.top + inner.bottom) / 2
		titleBarDrawControlLine(gtx.Ops, f32.Pt(inner.left, y), f32.Pt(inner.right, y), stroke, foreground)
	case titleBarGlyphMaximize:
		titleBarDrawControlOutline(gtx.Ops, inner, stroke, foreground)
	case titleBarGlyphRestore:
		titleBarDrawRestoreGlyph(gtx.Ops, inner, stroke, foreground)
	case titleBarGlyphClose:
		titleBarDrawControlLine(gtx.Ops, f32.Pt(inner.left, inner.top), f32.Pt(inner.right, inner.bottom), stroke, foreground)
		titleBarDrawControlLine(gtx.Ops, f32.Pt(inner.right, inner.top), f32.Pt(inner.left, inner.bottom), stroke, foreground)
	}
}

type titleBarControlGlyphBounds struct{ left, top, right, bottom float32 }

func (b titleBarControlGlyphBounds) empty() bool { return b.right <= b.left || b.bottom <= b.top }

func titleBarControlGlyphStroke(metric unit.Metric, width unit.Dp, glyphSize int) float32 {
	if width <= 0 {
		width = 1
	}
	scale := metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	stroke := max(float32(width)*scale, 1)
	return min(stroke, max(float32(glyphSize)/3, 1))
}

func titleBarDrawRestoreGlyph(ops *op.Ops, bounds titleBarControlGlyphBounds, stroke float32, foreground color.NRGBA) {
	offset := stroke * 2
	side := min(bounds.right-bounds.left, bounds.bottom-bounds.top) - offset
	if side <= stroke*2 {
		titleBarDrawControlOutline(ops, bounds, stroke, foreground)
		return
	}
	back := titleBarControlGlyphBounds{left: bounds.left + offset, top: bounds.top, right: bounds.left + offset + side, bottom: bounds.top + side}
	front := titleBarControlGlyphBounds{left: bounds.left, top: bounds.top + offset, right: bounds.left + side, bottom: bounds.top + offset + side}
	titleBarDrawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.right, back.top), stroke, foreground)
	titleBarDrawControlLine(ops, f32.Pt(back.right, back.top), f32.Pt(back.right, back.bottom), stroke, foreground)
	titleBarDrawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.left, front.top), stroke, foreground)
	titleBarDrawControlLine(ops, f32.Pt(front.right, back.bottom), f32.Pt(back.right, back.bottom), stroke, foreground)
	titleBarDrawControlOutline(ops, front, stroke, foreground)
}

func titleBarDrawControlOutline(ops *op.Ops, bounds titleBarControlGlyphBounds, stroke float32, foreground color.NRGBA) {
	if bounds.empty() || stroke <= 0 {
		return
	}
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(f32.Pt(bounds.left, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.bottom))
	path.LineTo(f32.Pt(bounds.left, bounds.bottom))
	path.Close()
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}

func titleBarDrawControlLine(ops *op.Ops, from, to f32.Point, stroke float32, foreground color.NRGBA) {
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(from)
	path.LineTo(to)
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}
