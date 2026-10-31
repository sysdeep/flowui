package radiogroup

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// buildRadioDefaults returns the theme-derived declarations, metrics, and
// motion for radio behavior.
func buildRadioDefaults(active *theme.Theme, variant Variant) radioDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return radioDefaults{
		Defaults: radioDefaultStyle(active, variant),
		Metrics: radioMetrics{
			FocusSpace: 2, ContentGap: 8,
			VerticalGap: 8, HorizontalGap: 16,
			DescriptionGap: 2, DotScale: .38,
			DotPressedScale: .5, PressedScale: 1,
		},
		Motion:            active.Motion,
		SelectionDuration: 200 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
		PressInDuration:   80 * time.Millisecond,
		PressOutDuration:  140 * time.Millisecond,
	}
}

// radioIndicatorColor reads the selected-dot color from the resolved indicator
// declaration. A missing or unsupported declaration is incomplete appearance.
func radioIndicatorColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
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

// radioPalette contains the semantic colors used by the radio defaults.
type radioPalette struct {
	Background, Border, SelectedBackground color.NRGBA
	Dot, Foreground, Description, Focus    color.NRGBA
}

// radioColors resolves radio colors from the active theme and state.
func radioColors(active *theme.Theme, variant Variant, hovered, pressed, disabled, invalid bool) radioPalette {
	sem := active.Semantic()
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	value := radioPalette{
		Background: sem.Surface, Border: sem.Border,
		SelectedBackground: sem.Tone(theme.TonePrimary).Fill,
		Dot:                white,
		Foreground:         sem.Foreground, Description: sem.MutedForeground, Focus: sem.Ring,
	}
	if variant == RadioSecondary {
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
		value.Dot = white
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

// radioDefaultStyle returns the state-aware radio declaration.
func radioDefaultStyle(active *theme.Theme, variant Variant) style.Style {
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
	repaint := func(p radioPalette) style.Style {
		return style.Style{}.On(style.Checked, style.Style{}.
			Background(style.Solid(p.SelectedBackground)).BorderColor(style.Solid(p.SelectedBackground)))
	}
	row := func(p radioPalette) style.Style {
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
	return style.Style{}.
		Part(style.PartIndicator, style.Style{}.
			Width(16).Height(16).Background(style.Solid(base.Background)).
			BorderWidth(1).BorderColor(style.Solid(base.Border)).Radius(8).
			Outline(ring, 0, style.Solid(base.Focus)).TextColor(style.Solid(base.Dot)).
			Rules(
				style.Rule{Set: repaint(base)},
				style.Rule{When: style.Invalid, Set: row(invalid)},
				style.Rule{When: style.Hovered, Not: style.Invalid, Set: row(hover)},
				style.Rule{When: style.Pressed, Not: style.Invalid, Set: repaint(pressed)},
				style.Rule{When: style.Hovered | style.Invalid, Set: row(invalidActive)},
				style.Rule{When: style.Pressed | style.Invalid, Set: row(invalidActive)},
				style.Rule{Rank: 1, When: style.Disabled, Set: disabledRow},
			)).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(base.Foreground)).FontSize(14).FontWeight(int(font.Medium)).
			On(style.Disabled, style.Style{}.TextColor(style.Solid(disabled.Foreground)))).
		Part(style.PartDescription, style.Style{}.
			TextColor(style.Solid(base.Description)).FontSize(12).
			On(style.Disabled, style.Style{}.TextColor(style.Solid(disabled.Description))))
}
