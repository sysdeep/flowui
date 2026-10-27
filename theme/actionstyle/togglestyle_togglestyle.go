// Package togglestyle contains the provider-owned visual recipe for toggle
// button behavior. It deliberately has no dependency on widget/togglebutton.
package actionstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	toggleColorDuration = 100 * time.Millisecond
	toggleScaleDuration = 250 * time.Millisecond
)

// BuildToggleButton returns the declaration layers consumed by toggle
// behavior. Foreground is supplied by the render context, not inferred by the
// provider, so inherited foreground remains authoritative.
func BuildToggleButton(active *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ToggleVisual{
		Defaults: toggleDefaultRules(active),
		Variant:  toggleVariantRules(active, request.Foreground, request.Variant),
		Size:     toggleSizeRules(active, request.Size, request.IconOnly),
		Override: theme.ToggleOverrideStyle(request),
	}
}

// visual contains resolved colors and opacity for the shared recipe.
type toggleVisual struct {
	Background, Foreground color.NRGBA
	Radius                 unit.Dp
	Opacity                float32
}

func toggleColors(active *theme.Theme, currentForeground color.NRGBA, variant theme.ToggleButtonVariant, selected, hovered, pressed, disabled bool) toggleVisual {
	sem := active.Semantic()
	background := sem.SurfaceRaised
	foreground := currentForeground
	if variant == theme.ToggleButtonGhost {
		background = color.NRGBA{}
		foreground = sem.Foreground
	}
	if hovered || pressed {
		background = sem.SurfacePressed
	}
	if selected {
		background = sem.Tone(theme.TonePrimary).Soft
		foreground = sem.Tone(theme.TonePrimary).OnSoft
		if hovered || pressed {
			background = sem.Tone(theme.TonePrimary).SoftHover
		}
	}
	opacity := float32(1)
	if disabled {
		opacity = active.DisabledOpacityValue()
	}
	return toggleVisual{Background: background, Foreground: foreground, Radius: active.Components.ToggleButton.Radius, Opacity: opacity}
}

func toggleDefaultRules(active *theme.Theme) style.Style {
	tokens := active.Components.ToggleButton
	sem := active.Semantic()
	return style.Style{}.
		Radius(tokens.Radius).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Cursor(pointer.CursorPointer).
		Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium))).
		Transition(style.PropBackgroundColor, toggleColorDuration).
		Transition(style.PropOutlineColor, toggleColorDuration).
		Transition(style.PropTransform, toggleScaleDuration).
		OnNot(style.FocusVisible, style.Disabled,
			style.Style{}.Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(sem.Ring)),
		).
		On(style.Disabled,
			style.Style{}.Opacity(active.DisabledOpacityValue()).Cursor(pointer.CursorDefault),
		)
}

func toggleVariantRules(active *theme.Theme, foreground color.NRGBA, variant theme.ToggleButtonVariant) style.Style {
	idle := toggleColors(active, foreground, variant, false, false, false, false)
	hovered := toggleColors(active, foreground, variant, false, true, false, false)
	selected := toggleColors(active, foreground, variant, true, false, false, false)
	selectedHovered := toggleColors(active, foreground, variant, true, true, false, false)
	background := func(value color.NRGBA) style.Style { return style.Style{}.Background(style.Solid(value)) }
	return style.Style{}.
		Background(style.Solid(idle.Background)).
		TextColor(style.Solid(idle.Foreground)).
		Rules(
			style.Rule{When: style.Hovered, Set: background(hovered.Background)},
			style.Rule{When: style.Pressed, Set: background(hovered.Background)},
			style.Rule{When: style.Selected, Set: style.Style{}.
				Background(style.Solid(selected.Background)).TextColor(style.Solid(selected.Foreground))},
			style.Rule{When: style.Selected | style.Hovered, Set: background(selectedHovered.Background)},
			style.Rule{When: style.Selected | style.Pressed, Set: background(selectedHovered.Background)},
		)
}

func toggleSizeRules(active *theme.Theme, size theme.ToggleButtonSize, iconOnly bool) style.Style {
	value := toggleSizeStyleFor(active, size)
	result := style.Style{}.
		Height(value.height).
		PaddingX(value.paddingX).
		Part(style.PartLabel, style.Style{}.FontSize(unit.Sp(value.textSize))).
		Part(style.PartIcon, style.Style{}.
			FontSize(unit.Sp(value.textSize)).
			MinWidth(unit.Dp(value.textSize)).
			MinHeight(unit.Dp(value.textSize))).
		Rules(style.Rule{When: style.Pressed, Set: style.Style{}.Scale(value.pressedScale)})
	if iconOnly {
		result = result.AspectRatio(1).PaddingX(0)
	}
	return result
}

type toggleSizeStyle struct {
	height       unit.Dp
	paddingX     unit.Dp
	textSize     float32
	pressedScale float32
}

func toggleSizeStyleFor(active *theme.Theme, size theme.ToggleButtonSize) toggleSizeStyle {
	tokens := active.Components.ToggleButton
	value := toggleSizeStyle{tokens.MediumHeight, tokens.MediumPaddingX, float32(tokens.MediumTextSize), tokens.PressedScaleMedium}
	switch size {
	case theme.ToggleButtonSmall:
		value = toggleSizeStyle{tokens.SmallHeight, tokens.SmallPaddingX, float32(tokens.SmallTextSize), tokens.PressedScaleSmall}
	case theme.ToggleButtonLarge:
		value = toggleSizeStyle{tokens.LargeHeight, tokens.LargePaddingX, float32(tokens.LargeTextSize), tokens.PressedScaleLarge}
	}
	if value.pressedScale <= 0 || value.pressedScale > 1 {
		switch size {
		case theme.ToggleButtonSmall:
			value.pressedScale = .98
		case theme.ToggleButtonLarge:
			value.pressedScale = .96
		default:
			value.pressedScale = .97
		}
	}
	return value
}
