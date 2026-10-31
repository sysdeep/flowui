package toggle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// buildCheckboxDefaults returns the theme-derived declarations, metrics, and
// motion for one checkbox variant.
func buildCheckboxDefaults(active *theme.Theme, variant CheckboxVariant) checkboxDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return checkboxDefaults{
		Defaults: checkboxDefaultStyle(active, variant),
		Metrics: checkboxMetrics{
			FocusSpace:          2,
			IndicatorSize:       12,
			IndeterminateStroke: 1.5,
			CheckStroke:         1.5,
			DescriptionGap:      4,
			DescriptionIndent:   28,
			LabelGap:            8,
			RequiredMarkOffset:  2,
		},
		Colors:            checkboxRequiredColors(active),
		Motion:            active.Motion,
		SelectionDuration: 200 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
	}
}

// checkboxRequiredColors returns the semantic colors used by the required mark.
func checkboxRequiredColors(active *theme.Theme) checkboxColors {
	sem := active.Semantic()
	value := sem.Tone(theme.ToneDanger).Fill
	return checkboxColors{
		RequiredMark:         value,
		DisabledRequiredMark: active.DisabledColor(value),
	}
}

// checkboxPalette contains the resolved semantic colors used by the checkbox
// declaration.
type checkboxPalette struct {
	Bg, Border, Accent, AccentHover, AccentFg, Fg, FocusColor color.NRGBA
}

// checkboxPaletteFor resolves checkbox colors from the active theme and state.
func checkboxPaletteFor(active *theme.Theme, variant CheckboxVariant, hovered, disabled, invalid bool) checkboxPalette {
	sem := active.Semantic()
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	danger := sem.Tone(theme.ToneDanger).Fill
	dangerHover := sem.Tone(theme.ToneDanger).Hover
	dangerFg := white
	value := checkboxPalette{
		Bg: sem.Surface, Border: sem.Border,
		Accent:      sem.Tone(theme.TonePrimary).Fill,
		AccentHover: sem.Tone(theme.TonePrimary).Hover,
		AccentFg:    white,
		Fg:          sem.Foreground, FocusColor: sem.Ring,
	}
	if variant == CheckboxSecondary {
		value.Bg = sem.SurfaceTertiary
	}
	if invalid {
		value.Border, value.Accent = danger, danger
		value.AccentHover, value.AccentFg = dangerHover, dangerFg
		value.FocusColor = danger
		value.FocusColor.A = sem.Ring.A
	}
	if disabled {
		value.Bg = active.DisabledColor(value.Bg)
		value.Border = active.DisabledColor(value.Border)
		value.Accent = active.DisabledColor(value.Accent)
		value.AccentHover = active.DisabledColor(value.AccentHover)
		value.AccentFg = active.DisabledColor(value.AccentFg)
		value.Fg = active.DisabledColor(value.Fg)
		value.FocusColor = color.NRGBA{}
		return value
	}
	if hovered {
		if invalid {
			value.Border = dangerHover
		} else {
			value.Border = sem.SurfacePressed
		}
		value.Accent = value.AccentHover
	}
	return value
}

// checkboxDefaultStyle returns the state-aware checkbox declaration for every
// part the control draws.
func checkboxDefaultStyle(active *theme.Theme, variant CheckboxVariant) style.Style {
	const ringWidth = 2
	base := checkboxPaletteFor(active, variant, false, false, false)
	hover := checkboxPaletteFor(active, variant, true, false, false)
	invalid := checkboxPaletteFor(active, variant, false, false, true)
	invalidHover := checkboxPaletteFor(active, variant, true, false, true)
	disabled := checkboxPaletteFor(active, variant, false, true, false)
	disabledInvalid := checkboxPaletteFor(active, variant, false, true, true)
	sem := active.Semantic()
	selected := func(p checkboxPalette) style.Style {
		return style.Style{}.OnAny(style.Checked|style.Indeterminate, style.Style{}.
			Background(style.Solid(p.Accent)).
			BorderColor(style.Solid(p.Accent)).
			TextColor(style.Solid(p.AccentFg)))
	}
	indicator := style.Style{}.
		Width(16).Height(16).
		BorderWidth(1).
		Background(style.Solid(base.Bg)).
		BorderColor(style.Solid(base.Border)).
		Radius(4).
		Outline(ringWidth, 0, style.Solid(base.FocusColor)).
		TextColor(style.Solid(base.Fg))
	row := func(p checkboxPalette) style.Style {
		return style.Style{}.
			BorderColor(style.Solid(p.Border)).
			Outline(ringWidth, 0, style.Solid(p.FocusColor)).
			TextColor(style.Solid(p.Fg)).
			OnAny(style.Checked|style.Indeterminate, style.Style{}.
				Background(style.Solid(p.Accent)).
				BorderColor(style.Solid(p.Accent)).
				TextColor(style.Solid(p.AccentFg)))
	}
	disabledSelected := style.Style{}.OnAny(style.Checked|style.Indeterminate, style.Style{}.
		Background(style.Solid(sem.DisabledBackground)).
		BorderColor(style.Solid(sem.DisabledBorder)).
		TextColor(style.Solid(color.NRGBA{R: 255, G: 255, B: 255, A: 255})))
	disabledInvalidSelected := style.Style{}.OnAny(style.Checked|style.Indeterminate, style.Style{}.
		Background(style.Solid(disabledInvalid.Accent)).
		BorderColor(style.Solid(disabledInvalid.Accent)).
		TextColor(style.Solid(color.NRGBA{R: 255, G: 255, B: 255, A: 255})))
	disabledPaint := style.Join(
		style.Style{}.
			Background(style.Solid(disabled.Bg)).
			BorderColor(style.Solid(disabled.Border)).
			Outline(ringWidth, 0, style.Solid(disabled.FocusColor)).
			TextColor(style.Solid(disabled.Fg)),
		disabledSelected,
	)
	disabledRow := disabledPaint.On(style.Invalid, style.Join(
		style.Style{}.
			Background(style.Solid(disabledInvalid.Bg)).
			BorderColor(style.Solid(disabledInvalid.Border)).
			Outline(ringWidth, 0, style.Solid(disabledInvalid.FocusColor)).
			TextColor(style.Solid(disabledInvalid.Fg)),
		disabledInvalidSelected,
	))
	return style.Style{}.
		Part(style.PartIndicator, indicator.Rules(
			style.Rule{Set: selected(base)},
			style.Rule{When: style.Invalid, Set: row(invalid)},
			style.Rule{When: style.Hovered, Not: style.Invalid, Set: row(hover)},
			style.Rule{When: style.Pressed, Not: style.Invalid, Set: row(hover)},
			style.Rule{When: style.Invalid | style.Hovered, Set: row(invalidHover)},
			style.Rule{When: style.Invalid | style.Pressed, Set: row(invalidHover)},
			style.Rule{Rank: 1, When: style.Disabled, Set: disabledRow},
		)).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(base.Fg)).
			FontSize(active.Typography.ControlSize).
			FontWeight(int(font.Medium)).
			Rules(style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.
				TextColor(style.Solid(disabled.Fg))})).
		Part(style.PartDescription, style.Style{}.
			TextColor(style.Solid(sem.MutedForeground)).
			Rules(
				style.Rule{When: style.Invalid, Set: style.Style{}.
					TextColor(style.Solid(sem.Tone(theme.ToneDanger).Fill))},
				style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.
					TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))).
					On(style.Invalid, style.Style{}.
						TextColor(style.Solid(active.DisabledColor(sem.Tone(theme.ToneDanger).Fill))))},
			))
}

// buildSwitchDefaults returns the theme-derived declaration, metrics, and
// motion for one switch size.
func buildSwitchDefaults(active *theme.Theme, size SwitchSize) switchDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return switchDefaults{
		Defaults: switchDefaultStyle(active, size),
		Metrics: switchMetrics{
			FocusSpace:     2,
			ContentGap:     8,
			DescriptionGap: 4,
		},
		Motion:            active.Motion,
		SelectionDuration: active.Motion.ControlDuration,
		FocusDuration:     100 * time.Millisecond,
	}
}

// switchPalette contains semantic switch colors for a given interaction row.
type switchPalette struct {
	TrackOff, TrackOn  color.NRGBA
	ThumbForegroundOff color.NRGBA
	ThumbForeground    color.NRGBA
	Focus              color.NRGBA
}

// switchPaletteFor resolves switch colors from the active theme and state.
func switchPaletteFor(active *theme.Theme, hovered, invalid bool) switchPalette {
	sem := active.Semantic()
	value := switchPalette{
		TrackOff:           sem.SurfaceRaised,
		TrackOn:            sem.Tone(theme.TonePrimary).Fill,
		ThumbForegroundOff: sem.MutedForeground,
		ThumbForeground:    sem.Tone(theme.TonePrimary).Fill,
		Focus:              sem.Ring,
	}
	if hovered {
		value.TrackOff = sem.SurfacePressed
		value.TrackOn = sem.Tone(theme.TonePrimary).Hover
	}
	if invalid {
		value.TrackOn = sem.Tone(theme.ToneDanger).Fill
		value.ThumbForeground = sem.Tone(theme.ToneDanger).Fill
		value.Focus = sem.Tone(theme.ToneDanger).Fill
		value.Focus.A = sem.Ring.A
		if hovered {
			value.TrackOn = sem.Tone(theme.ToneDanger).Hover
		}
	}
	return value
}

type switchDimensions struct {
	trackWidth  unit.Dp
	trackHeight unit.Dp
	thumbWidth  unit.Dp
	thumbHeight unit.Dp
}

func switchDimensionsFor(size SwitchSize) switchDimensions {
	switch size {
	case SwitchSmall:
		return switchDimensions{
			trackWidth: 28, trackHeight: 16,
			thumbWidth: 14, thumbHeight: 14,
		}
	case SwitchLarge:
		return switchDimensions{
			trackWidth: 42, trackHeight: 24,
			thumbWidth: 22, thumbHeight: 22,
		}
	default:
		return switchDimensions{
			trackWidth: 35, trackHeight: 20,
			thumbWidth: 18, thumbHeight: 18,
		}
	}
}

// switchDefaultStyle returns the state-aware switch declaration for every part
// the control draws.
func switchDefaultStyle(active *theme.Theme, size SwitchSize) style.Style {
	sizes := switchDimensionsFor(size)
	rest := switchPaletteFor(active, false, false)
	hovered := switchPaletteFor(active, true, false)
	invalid := switchPaletteFor(active, false, true)
	invalidHovered := switchPaletteFor(active, true, true)
	dim := active.DisabledOpacityValue()
	background := func(value style.ColorSource) style.Style { return style.Style{}.Background(value) }
	track := style.Style{}.
		Width(sizes.trackWidth).
		Height(sizes.trackHeight).
		Background(style.Solid(rest.TrackOff)).
		BorderWidth(1).
		BorderColor(style.Solid(active.Semantic().Separator)).
		Radius(sizes.trackHeight/2).
		Outline(2, 0, style.Solid(rest.Focus)).
		Rules(
			style.Rule{When: style.Checked, Set: style.New().BorderColor(style.RGBA(0))},
			style.Rule{When: style.Hovered, Not: style.Checked, Set: background(style.Solid(hovered.TrackOff))},
			style.Rule{When: style.Pressed, Not: style.Checked, Set: background(style.Solid(hovered.TrackOff))},
			style.Rule{When: style.Checked, Set: background(style.Solid(rest.TrackOn))},
			style.Rule{When: style.Checked | style.Hovered, Set: background(style.Solid(hovered.TrackOn))},
			style.Rule{When: style.Checked | style.Pressed, Set: background(style.Solid(hovered.TrackOn))},
			style.Rule{When: style.Checked | style.Invalid, Set: background(style.Solid(invalid.TrackOn))},
			style.Rule{When: style.Checked | style.Invalid | style.Hovered, Set: background(style.Solid(invalidHovered.TrackOn))},
			style.Rule{When: style.Checked | style.Invalid | style.Pressed, Set: background(style.Solid(invalidHovered.TrackOn))},
			style.Rule{When: style.Invalid, Not: style.Disabled, Set: style.Style{}.Outline(2, 0, style.Solid(invalid.Focus))},
			style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim).Outline(2, 0, style.SolidColor{})},
		)
	white := style.Solid(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	thumb := style.Style{}.
		Width(sizes.thumbWidth).
		Height(sizes.thumbHeight).
		Background(white).
		Radius(sizes.thumbHeight/2).
		TextColor(style.Solid(rest.ThumbForegroundOff)).
		BoxShadow(0, 1, 2, 0, style.WithAlpha(color.NRGBA{A: 255}, .16)).
		Rules(
			style.Rule{When: style.Checked, Set: style.Style{}.TextColor(style.Solid(rest.ThumbForeground))},
			style.Rule{When: style.Checked | style.Invalid, Set: style.Style{}.TextColor(style.Solid(invalid.ThumbForeground))},
			style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)},
		)
	label := style.Style{}.
		TextColor(style.Solid(active.Semantic().Foreground)).
		FontSize(active.Typography.ControlSize).
		FontWeight(int(font.Medium)).
		Rules(style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)})
	description := style.Style{}.
		TextColor(style.Solid(active.Semantic().MutedForeground)).
		FontSize(active.Typography.SmallSize).
		Rules(style.Rule{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)})

	return style.New().
		Part(style.PartTrack, track).
		Part(style.PartThumb, thumb).
		Part(style.PartLabel, label).
		Part(style.PartDescription, description)
}

const (
	toggleColorDuration = 100 * time.Millisecond
	toggleScaleDuration = 250 * time.Millisecond
)

// toggleButtonDefaults is the cached appearance of one toggle button request.
// It holds only theme- and context-derived layers; instance declarations are
// resolved as the last cascade layer and are never part of the cache key.
type toggleButtonDefaults struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
}

// buildToggleButtonDefaults returns the declaration layers consumed by toggle
// behavior. Foreground is supplied by the render context, not inferred by the
// declaration, so inherited foreground remains authoritative.
func buildToggleButtonDefaults(
	active *theme.Theme,
	variant ToggleButtonVariant,
	size ToggleButtonSize,
	iconOnly bool,
	foreground color.NRGBA,
) toggleButtonDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return toggleButtonDefaults{
		Defaults: toggleButtonDefaultStyle(active),
		Variant:  toggleButtonVariantStyle(active, foreground, variant),
		Size:     toggleButtonSizeStyle(size, iconOnly),
	}
}

// toggleButtonPalette contains resolved colors for the default declaration.
type toggleButtonPalette struct {
	Background, Foreground color.NRGBA
}

func toggleButtonColors(
	active *theme.Theme,
	currentForeground color.NRGBA,
	variant ToggleButtonVariant,
	selected, hovered bool,
) toggleButtonPalette {
	sem := active.Semantic()
	background := sem.SurfaceRaised
	foreground := currentForeground
	if variant == ToggleButtonGhost {
		background = color.NRGBA{}
		foreground = sem.Foreground
	}
	if hovered {
		background = sem.SurfacePressed
	}
	if selected {
		background = sem.Tone(theme.TonePrimary).Soft
		foreground = sem.Tone(theme.TonePrimary).OnSoft
		if hovered {
			background = sem.Tone(theme.TonePrimary).SoftHover
		}
	}
	return toggleButtonPalette{Background: background, Foreground: foreground}
}

func toggleButtonDefaultStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Radius(4).
		Outline(2, 2, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Cursor(pointer.CursorPointer).
		Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium))).
		Transition(style.PropBackgroundColor, toggleColorDuration).
		Transition(style.PropOutlineColor, toggleColorDuration).
		Transition(style.PropTransform, toggleScaleDuration).
		OnNot(style.FocusVisible, style.Disabled,
			style.Style{}.Outline(2, 2, style.Solid(sem.Ring)),
		).
		On(style.Disabled,
			style.Style{}.Opacity(active.DisabledOpacityValue()).Cursor(pointer.CursorDefault),
		)
}

func toggleButtonVariantStyle(active *theme.Theme, foreground color.NRGBA, variant ToggleButtonVariant) style.Style {
	idle := toggleButtonColors(active, foreground, variant, false, false)
	hovered := toggleButtonColors(active, foreground, variant, false, true)
	selected := toggleButtonColors(active, foreground, variant, true, false)
	selectedHovered := toggleButtonColors(active, foreground, variant, true, true)
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

func toggleButtonSizeStyle(size ToggleButtonSize, iconOnly bool) style.Style {
	value := toggleSizeMetricsFor(size)
	result := style.Style{}.
		Height(value.height).
		Padding(0, value.paddingX).
		Part(style.PartLabel, style.Style{}.FontSize(unit.Sp(value.textSize))).
		Part(style.PartIcon, style.Style{}.
			FontSize(unit.Sp(value.textSize)).
			MinWidth(unit.Dp(value.textSize)).
			MinHeight(unit.Dp(value.textSize)))
	if iconOnly {
		result = result.AspectRatio(1).Padding(0)
	}
	return result
}

type toggleSizeMetrics struct {
	height   unit.Dp
	paddingX unit.Dp
	textSize float32
}

func toggleSizeMetricsFor(size ToggleButtonSize) toggleSizeMetrics {
	switch size {
	case ToggleButtonSmall:
		return toggleSizeMetrics{height: 24, paddingX: 8, textSize: 14}
	case ToggleButtonLarge:
		return toggleSizeMetrics{height: 40, paddingX: 16, textSize: 16}
	default:
		return toggleSizeMetrics{height: 32, paddingX: 12, textSize: 14}
	}
}

// segmentedDefaults is the theme-derived appearance of a segmented track.
type segmentedDefaults struct {
	root style.Style
	gap  unit.Dp
}

func buildSegmentedDefaults(active *theme.Theme) segmentedDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return segmentedDefaults{
		root: style.New().
			Padding(3).
			Radius(8).
			Background(style.Color(active.Semantic().Muted)),
		gap: 2,
	}
}
