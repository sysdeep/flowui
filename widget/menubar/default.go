package menubar

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	menubarDefaultTriggerHeight          unit.Dp = 32
	menubarDefaultTriggerPaddingX        unit.Dp = 12
	menubarDefaultTriggerRadius          unit.Dp = 4
	menubarDefaultTriggerTextSize        unit.Sp = 14
	menubarDefaultTriggerFocusRingWidth  unit.Dp = 2
	menubarDefaultTriggerFocusRingOffset unit.Dp = 1
	menubarDefaultGap                    unit.Dp = 0
	menubarDefaultPanelGap               unit.Dp = 4
)

// menubarDefaults is the theme-derived appearance and geometry of one menubar
// density. It is cached by menubarDefaultsCache; the instance style is resolved
// separately because it is not one of the cache inputs.
type menubarDefaults struct {
	root       style.Style
	trigger    style.Style
	gap        unit.Dp
	panelGap   unit.Dp
	textSize   unit.Sp
	foreground color.NRGBA
	selection  color.NRGBA
	motion     theme.MotionTheme
}

// defaultMenubar derives one menubar density from the active theme.
func defaultMenubar(active *theme.Theme, compact bool) menubarDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	height := menubarDefaultTriggerHeight
	paddingX := menubarDefaultTriggerPaddingX
	radius := menubarDefaultTriggerRadius
	textSize := menubarDefaultTriggerTextSize
	if compact {
		height = menubarMinDp(height, 28)
		paddingX = menubarMinDp(paddingX, 8)
		radius = menubarMinDp(radius, 4)
		textSize = menubarMinSp(textSize, 13)
	}
	sem := active.Semantic()
	neutral := sem.Tone(theme.ToneDefault)
	trigger := style.Style{}.
		Height(height).
		Padding(0, paddingX).
		Radius(radius).
		Background(style.RGBA(0)).
		TextColor(style.Solid(sem.Foreground)).
		FontSize(textSize).
		FontWeight(int(font.Normal)).
		Outline(menubarDefaultTriggerFocusRingWidth, menubarDefaultTriggerFocusRingOffset, style.Solid(sem.Ring)).
		Opacity(1).
		On(style.Open, style.Style{}.Background(style.Solid(neutral.Fill))).
		On(style.Hovered, style.Style{}.Background(style.Solid(neutral.Fill))).
		On(style.Pressed, style.Style{}.Background(style.Solid(neutral.Hover))).
		On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue()))
	return menubarDefaults{
		root:       style.Style{},
		trigger:    trigger,
		gap:        menubarDefaultGap,
		panelGap:   menubarDefaultPanelGap,
		textSize:   textSize,
		foreground: sem.Foreground,
		selection:  sem.Selection,
		motion:     active.Motion,
	}
}

func menubarMinDp(value, limit unit.Dp) unit.Dp {
	if value < limit {
		return value
	}
	return limit
}

func menubarMinSp(value, limit unit.Sp) unit.Sp {
	if value < limit {
		return value
	}
	return limit
}

// menubarLabelStyle builds the component-owned text primitive for a trigger.
// The resolved trigger part supplies typography and foreground, so an instance
// style that changes either reaches the label.
func menubarLabelStyle(active *theme.Theme, resolved style.ResolvedStyle, defaults menubarDefaults, text string) material.LabelStyle {
	size, weight := defaults.textSize, font.Normal
	foreground := menubarTextColor(resolved, defaults.foreground)
	if resolved.Text != nil {
		textStyle := resolved.Text
		if textStyle.FontSize != nil && *textStyle.FontSize > 0 {
			size = *textStyle.FontSize
		}
		if textStyle.FontWeight != nil {
			weight = font.Weight(*textStyle.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, text)
	label.Color = foreground
	label.SelectionColor = defaults.selection
	label.Font.Weight = weight
	if resolved.Text != nil {
		textStyle := resolved.Text
		if textStyle.Typeface != nil {
			label.Font.Typeface = *textStyle.Typeface
		}
		if textStyle.FontStyle != nil {
			label.Font.Style = *textStyle.FontStyle
		}
		if textStyle.LineHeight != nil {
			label.LineHeight = *textStyle.LineHeight
		}
		if textStyle.LineHeightScale != nil {
			label.LineHeightScale = *textStyle.LineHeightScale
		}
	}
	return label
}

// menubarTextColor resolves the trigger foreground, falling back to the
// theme-derived color when a declaration intentionally inherits it.
func menubarTextColor(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Text != nil {
		if value, ok := computed.Color(resolved.Text.Color); ok {
			return value
		}
	}
	return fallback
}

// menubarRenderTrigger paints the trigger surface and its animated focus ring.
func menubarRenderTrigger(gtx layout.Context, size image.Point, resolved style.ResolvedStyle, focusOpacity float32) {
	if size.X <= 0 || size.Y <= 0 || resolved.Paint == nil {
		return
	}
	rect := image.Rectangle{Max: size}
	if brush, ok := computed.Brush(resolved.Paint.Background); ok {
		draw.FillRRect(gtx, menubarRRect(gtx, rect, resolved.Paint), brush)
	}
	outline := resolved.Paint.Outline
	if focusOpacity <= 0 || outline == nil {
		return
	}
	focus, ok := computed.Color(outline.Color)
	if !ok || focus.A == 0 {
		return
	}
	radius := min(menubarRadius(gtx, resolved.Paint), min(size.X, size.Y)/2)
	width := max(gtx.Dp(outline.Width), 1)
	inset := max(gtx.Dp(outline.Offset), 0) + (width+1)/2
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	focus.A = byte(float32(focus.A)*min(max(focusOpacity, 0), 1) + 0.5)
	draw.RoundedInsetStroke(gtx, rect, max(radius-inset, 0), width, inset, focus)
}

func menubarRRect(gtx layout.Context, rect image.Rectangle, paintStyle *style.PaintStyle) clip.RRect {
	shape := clip.RRect{Rect: rect}
	if paintStyle == nil {
		return shape
	}
	if paintStyle.Radii != nil {
		radii := *paintStyle.Radii
		limit := min(rect.Dx(), rect.Dy()) / 2
		shape.NW = min(max(gtx.Dp(radii.TopLeft), 0), limit)
		shape.NE = min(max(gtx.Dp(radii.TopRight), 0), limit)
		shape.SE = min(max(gtx.Dp(radii.BottomRight), 0), limit)
		shape.SW = min(max(gtx.Dp(radii.BottomLeft), 0), limit)
		return shape
	}
	radius := min(menubarRadius(gtx, paintStyle), min(rect.Dx(), rect.Dy())/2)
	shape.NW, shape.NE, shape.SE, shape.SW = radius, radius, radius, radius
	return shape
}

func menubarRadius(gtx layout.Context, paintStyle *style.PaintStyle) int {
	if paintStyle == nil || paintStyle.Radius == nil {
		return 0
	}
	return max(gtx.Dp(*paintStyle.Radius), 0)
}

// menubarPaintColor returns the concrete fill an inherited scope should expose
// to custom trigger content. Gradients have no single color, so they leave the
// caller's background untouched.
func menubarPaintColor(paintStyle *style.PaintStyle) color.NRGBA {
	if paintStyle == nil {
		return color.NRGBA{}
	}
	switch value := paintStyle.Background.(type) {
	case style.SolidColor:
		return value.Color
	case *style.SolidColor:
		if value != nil {
			return value.Color
		}
	}
	return color.NRGBA{}
}
