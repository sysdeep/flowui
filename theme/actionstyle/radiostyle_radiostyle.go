// Package radiostyle owns the provider-side visual recipe for radio groups.
package actionstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildRadioGroup returns declarations, metrics, and motion for radio behavior.
func BuildRadioGroup(active *theme.Theme, request theme.RadioGroupRequest) theme.RadioGroupVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.RadioGroup
	return theme.RadioGroupVisual{
		Defaults: radioRuleSet(active, request.Variant).Compile(),
		Override: theme.RadioGroupOverrideStyle(request),
		Metrics: theme.RadioGroupMetrics{
			FocusSpace: tokens.FocusSpace, ContentGap: tokens.ContentGap,
			VerticalGap: tokens.VerticalGap, HorizontalGap: tokens.HorizontalGap,
			DescriptionGap: tokens.DescriptionGap, DotScale: tokens.DotScale,
			DotPressedScale: tokens.DotPressedScale, PressedScale: tokens.PressedScale,
		},
		IndicatorColor:    radioRadioIndicatorColor,
		IndicatorRenderer: radioRadioIndicatorRenderer,
		Motion:            active.Motion,
		SelectionDuration: 200 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
		PressInDuration:   80 * time.Millisecond,
		PressOutDuration:  140 * time.Millisecond,
	}
}

// radioIndicatorColor keeps the selected-dot color interpretation on the
// provider side. The widget only receives the already-resolved color and
// treats a missing declaration as an incomplete provider contract.
func radioRadioIndicatorColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	switch value := resolved.Text.Color.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

// radioPalette contains the semantic colors used by the radio recipe.
type radioRadioPalette struct {
	Background, Border, SelectedBackground color.NRGBA
	Dot, Foreground, Description, Focus    color.NRGBA
}

// colors resolves radio colors from the active theme and state.
func radioColors(active *theme.Theme, variant theme.RadioGroupVariant, hovered, pressed, disabled, invalid bool) radioRadioPalette {
	sem := active.Semantic()
	value := radioRadioPalette{
		Background: sem.Surface, Border: sem.Border,
		SelectedBackground: sem.Tone(theme.TonePrimary).Fill,
		Dot:                sem.Tone(theme.TonePrimary).OnFill,
		Foreground:         sem.Foreground, Description: sem.MutedForeground, Focus: sem.Ring,
	}
	if variant == theme.RadioSecondary {
		value.Background = sem.SurfaceRaised
	}
	if hovered {
		value.Border = sem.SurfacePressed
	}
	if pressed {
		value.SelectedBackground = sem.Tone(theme.TonePrimary).Hover
	}
	if invalid {
		value.Border = sem.Tone(theme.ToneDanger).Fill
		value.SelectedBackground = sem.Tone(theme.ToneDanger).Fill
		value.Dot = sem.Tone(theme.ToneDanger).OnFill
		value.Focus = sem.Tone(theme.ToneDanger).Fill
		value.Focus.A = sem.Ring.A
		if hovered || pressed {
			value.Border = sem.Tone(theme.ToneDanger).Hover
			value.SelectedBackground = sem.Tone(theme.ToneDanger).Hover
		}
	}
	if disabled {
		value.Background = active.DisabledColor(value.Background)
		value.Border = active.DisabledColor(value.Border)
		value.SelectedBackground = active.DisabledColor(value.SelectedBackground)
		value.Dot = active.DisabledColor(value.Dot)
		value.Foreground = active.DisabledColor(value.Foreground)
		value.Description = active.DisabledColor(value.Description)
		value.Focus = color.NRGBA{}
	}
	return value
}

// ruleSet returns the state-aware radio recipe.
func radioRuleSet(active *theme.Theme, variant theme.RadioGroupVariant) style.Recipe {
	tokens := active.Components.RadioGroup
	ring := unit.Dp(2)
	base := radioColors(active, variant, false, false, false, false)
	hover := radioColors(active, variant, true, false, false, false)
	pressed := radioColors(active, variant, false, true, false, false)
	invalid := radioColors(active, variant, false, false, false, true)
	invalidActive := radioColors(active, variant, true, false, false, true)
	disabled := radioColors(active, variant, false, false, true, false)
	disabledHover := radioColors(active, variant, true, false, true, false)
	disabledPressed := radioColors(active, variant, false, true, true, false)
	disabledInvalid := radioColors(active, variant, false, false, true, true)
	disabledInvalidActive := radioColors(active, variant, true, false, true, true)
	repaint := func(p radioRadioPalette) style.Style {
		return style.Style{}.On(style.Checked, style.Style{}.
			Background(style.Solid(p.SelectedBackground)).BorderColor(style.Solid(p.SelectedBackground)))
	}
	row := func(p radioRadioPalette) style.Style {
		return style.Join(style.Style{}.
			BorderColor(style.Solid(p.Border)).Outline(ring, 0, style.Solid(p.Focus)).TextColor(style.Solid(p.Dot)), repaint(p))
	}
	disabledBorder := style.Style{}.
		Background(style.Solid(disabled.Background)).BorderColor(style.Solid(disabled.Border)).
		Outline(ring, 0, style.Solid(disabled.Focus)).TextColor(style.Solid(disabled.Dot)).
		On(style.Hovered, style.Style{}.BorderColor(style.Solid(disabledHover.Border))).
		On(style.Invalid, style.Join(style.Style{}.
			Background(style.Solid(disabledInvalid.Background)).BorderColor(style.Solid(disabledInvalid.Border)).
			Outline(ring, 0, style.Solid(disabledInvalid.Focus)).TextColor(style.Solid(disabledInvalid.Dot)),
			style.Style{}.On(style.Hovered, style.Style{}.BorderColor(style.Solid(disabledInvalidActive.Border))).
				On(style.Pressed, style.Style{}.BorderColor(style.Solid(disabledInvalidActive.Border)))))
	disabledSelected := style.Style{}.
		Background(style.Solid(disabled.SelectedBackground)).BorderColor(style.Solid(disabled.SelectedBackground)).
		On(style.Pressed, style.Style{}.
			Background(style.Solid(disabledPressed.SelectedBackground)).BorderColor(style.Solid(disabledPressed.SelectedBackground)))
	disabledInvalidSelected := style.Style{}.
		Background(style.Solid(disabledInvalid.SelectedBackground)).BorderColor(style.Solid(disabledInvalid.SelectedBackground)).
		On(style.Hovered, style.Style{}.
			Background(style.Solid(disabledInvalidActive.SelectedBackground)).BorderColor(style.Solid(disabledInvalidActive.SelectedBackground))).
		On(style.Pressed, style.Style{}.
			Background(style.Solid(disabledInvalidActive.SelectedBackground)).BorderColor(style.Solid(disabledInvalidActive.SelectedBackground)))
	disabledRow := style.Join(disabledBorder, style.Style{}.
		On(style.Checked, disabledSelected.On(style.Invalid, disabledInvalidSelected)))
	return style.Recipe{
		style.PartIndicator: {Base: style.Style{}.
			Width(tokens.Size).Height(tokens.Size).Background(style.Solid(base.Background)).
			BorderWidth(tokens.BorderWidth).BorderColor(style.Solid(base.Border)).Radius(tokens.Size/2).
			Outline(ring, 0, style.Solid(base.Focus)).TextColor(style.Solid(base.Dot)), Rules: []style.Rule{
			{Set: repaint(base)}, {When: style.Invalid, Set: row(invalid)},
			{When: style.Hovered, Not: style.Invalid, Set: row(hover)},
			{When: style.Pressed, Not: style.Invalid, Set: repaint(pressed)},
			{When: style.Hovered | style.Invalid, Set: row(invalidActive)},
			{When: style.Pressed | style.Invalid, Set: row(invalidActive)},
			{Rank: 1, When: style.Disabled, Set: disabledRow},
		}},
		style.PartLabel: {Base: style.Style{}.
			TextColor(style.Solid(base.Foreground)).FontSize(tokens.TextSize).FontWeight(int(font.Medium)),
			Rules: []style.Rule{{When: style.Disabled, Set: style.Style{}.TextColor(style.Solid(disabled.Foreground))}}},
		style.PartDescription: {Base: style.Style{}.
			TextColor(style.Solid(base.Description)).FontSize(tokens.DescriptionSize),
			Rules: []style.Rule{{When: style.Disabled, Set: style.Style{}.TextColor(style.Solid(disabled.Description))}}},
	}
}
