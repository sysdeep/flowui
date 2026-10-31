package button

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (b Widget) defaultStyles(active *theme.Theme) buttonDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	height, padding, radius, spinner := unit.Dp(32), unit.Dp(12), unit.Dp(4), unit.Dp(16)
	fontSize := active.Typography.ControlSize
	switch b.size {
	case Small:
		height, padding, radius, spinner, fontSize = 24, 8, 3, 14, 12
	case Large:
		height, padding, radius, spinner = 40, 16, 6, 18
		fontSize += 2
	}
	if b.iconOnly {
		padding = 0
	}
	size := style.New().Height(height).Padding(0, padding).FontSize(fontSize)
	if b.iconOnly {
		size = size.Width(height)
	}
	if b.group.grouped {
		size = style.Join(size, b.group.cornerStyle(radius))
	} else {
		pressDuration := active.Motion.ControlDuration
		if pressDuration <= 0 {
			pressDuration = 90 * time.Millisecond
		}
		size = size.Rules(style.Rule{
			When: style.Pressed,
			Set:  style.New().Scale(1),
			Tween: &style.Tween{
				Property: style.PropTransform,
				Duration: pressDuration,
			},
		})
	}
	sem := active.Semantic()
	duration := active.Motion.ControlDuration
	if duration <= 0 {
		duration = 100 * time.Millisecond
	}
	scaleDuration := active.Motion.ControlDuration
	if scaleDuration <= 0 {
		scaleDuration = 160 * time.Millisecond
	}
	defaults := style.New().
		Radius(radius).
		FontWeight(int(font.Medium)).
		Outline(2, 1, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Transition(style.PropBackgroundColor, duration).
		Transition(style.PropTextColor, duration).
		Transition(style.PropBorderColor, duration).
		Transition(style.PropOpacity, duration).
		Transition(style.PropOutlineColor, duration).
		Transition(style.PropTransform, scaleDuration).
		On(style.Disabled, buttonDisabledStyle(active)).
		OnNot(style.FocusVisible, style.Disabled, style.New().Outline(2, 2, style.Solid(sem.Ring)))
	return buttonDefaults{
		defaults: defaults,
		variant:  buttonVariantStyle(active, b.variant),
		size:     size,
		spinner:  spinner,
		motion:   active.Motion,
	}
}

func (g buttonGroupItemStyle) cornerStyle(radius unit.Dp) style.Style {
	if g.position == buttonGroupSingle {
		return style.New().Radius(0).
			RadiusTopLeft(radius).RadiusTopRight(radius).
			RadiusBottomLeft(radius).RadiusBottomRight(radius)
	}
	result := style.New().Radius(0)
	if g.orientation == GroupVertical {
		switch g.position {
		case buttonGroupStart:
			return result.RadiusTopLeft(radius).RadiusTopRight(radius)
		case buttonGroupEnd:
			return result.RadiusBottomLeft(radius).RadiusBottomRight(radius)
		}
		return result
	}
	switch g.position {
	case buttonGroupStart:
		return result.RadiusTopLeft(radius).RadiusBottomLeft(radius)
	case buttonGroupEnd:
		return result.RadiusTopRight(radius).RadiusBottomRight(radius)
	}
	return result
}

func buttonVariantStyle(active *theme.Theme, variant Variant) style.Style {
	sem := active.Semantic()
	neutral := sem.Tone(theme.ToneDefault)
	primary := sem.Tone(theme.TonePrimary)
	danger := sem.Tone(theme.ToneDanger)
	var background, hover, pressed, foreground color.NRGBA
	switch variant {
	case Primary:
		background, hover, pressed, foreground = primary.Fill, primary.Hover, primary.Pressed, primary.OnFill
	case Secondary:
		background, hover, foreground = primary.Soft, primary.SoftHover, primary.OnSoft
		pressed = theme.Mix(hover, foreground, 8)
	case Tertiary:
		background, hover, foreground = sem.ControlSurface, sem.ControlSurfaceHover, neutral.OnSoft
		pressed = theme.Mix(hover, foreground, 8)
	case Ghost:
		hover, pressed, foreground = neutral.Fill, neutral.Hover, neutral.OnFill
	case Danger:
		background, hover, pressed, foreground = danger.Fill, danger.Hover, danger.Pressed, danger.OnFill
	case DangerSoft:
		background, hover, foreground = danger.Soft, danger.SoftHover, danger.OnSoft
		pressed = theme.Mix(hover, foreground, 8)
	case Outline, DangerOutline:
		pair := primary
		if variant == DangerOutline {
			pair = danger
		}
		edge := pair.Fill
		dark := theme.ContrastRatio(sem.Background, color.NRGBA{A: 255}) <
			theme.ContrastRatio(sem.Foreground, color.NRGBA{A: 255})
		if value, ok := buttonAccentEdge(pair.Fill, dark); ok {
			edge = value
		}
		return style.New().Background(style.RGBA(0)).
			TextColor(style.Solid(pair.OnSoft)).BorderWidth(1).BorderColor(style.Solid(edge)).
			On(style.Hovered, style.New().Background(style.Solid(pair.Soft))).
			On(style.Pressed, style.New().Background(style.Solid(pair.SoftHover))).
			On(style.Disabled, disabledButton(sem))
	default:
		// Keep malformed or forward-incompatible values usable instead of
		// producing a zero-color, zero-state declaration.
		variant = Primary
		background, hover, pressed, foreground = primary.Fill, primary.Hover, primary.Pressed, primary.OnFill
	}
	result := style.New().Background(style.Solid(background)).TextColor(style.Solid(foreground))
	if gradient := sem.PrimaryGradient; variant == Primary && gradient != nil && len(gradient.Stops) > 0 {
		stops := make([]style.GradientStop, len(gradient.Stops))
		for index, stop := range gradient.Stops {
			stops[index] = style.GradientStop{Offset: stop.Offset, Color: style.Solid(stop.Color)}
		}
		result = result.Background(style.Gradient{AngleDegrees: gradient.AngleDegrees, Stops: stops})
	}
	disabled := buttonDisabledStyle(active)
	if variant == Secondary {
		disabled = disabledButton(sem)
	}
	return result.
		On(style.Hovered, style.New().Background(style.Solid(hover))).
		On(style.Pressed, style.New().Background(style.Solid(pressed))).
		On(style.Disabled, disabled)
}

func buttonDisabledStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	if sem.DisabledBackground.A == 0 {
		return style.New().Opacity(active.DisabledOpacityValue())
	}
	return disabledButton(sem)
}

func disabledButton(sem theme.Semantic) style.Style {
	return style.New().Background(style.Solid(sem.DisabledBackground)).
		TextColor(style.Solid(sem.DisabledForeground)).BorderColor(style.Solid(sem.DisabledBorder)).
		BoxShadowNone().Opacity(1)
}

func buttonAccentEdge(fill color.NRGBA, dark bool) (color.NRGBA, bool) {
	edges := map[uint32][2]uint32{
		0x3e63dd: {0x8da4ef, 0x435db1},
		0xe5484d: {0xeb8e90, 0xb54548},
		0x30a46c: {0x5bb98b, 0x2f7c57},
		0xffc53d: {0xe2a336, 0x8f6424},
	}
	pair, ok := edges[uint32(fill.R)<<16|uint32(fill.G)<<8|uint32(fill.B)]
	index := 0
	if dark {
		index = 1
	}
	value := pair[index]
	return color.NRGBA{R: byte(value >> 16), G: byte(value >> 8), B: byte(value), A: 255}, ok
}

func defaultCloseStyles(active *theme.Theme) closeStyles {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	root := style.New().
		Width(24).
		Height(24).
		Padding(4).
		Background(style.Solid(sem.SurfaceRaised)).
		TextColor(style.Solid(sem.MutedForeground)).
		Radius(4).
		Outline(2, 0, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Cursor(pointer.CursorPointer).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		Transition(style.PropOutlineColor, 100*time.Millisecond).
		Transition(style.PropTransform, 250*time.Millisecond).
		OnNot(style.Hovered, style.Disabled, style.New().Background(style.Solid(sem.SurfacePressed))).
		On(style.Pressed, style.New().Scale(1)).
		OnNot(style.FocusVisible, style.Disabled, style.New().Outline(2, 0, style.Solid(sem.Ring))).
		On(style.Disabled, style.New().
			Background(style.Solid(active.DisabledColor(sem.SurfaceRaised))).
			TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))).
			Cursor(pointer.CursorDefault))
	return closeStyles{
		root: root,
		icon: style.New().Part(style.PartIcon, style.New().Width(16).Height(16)),
	}
}
