package fieldstyle

import (
	"time"

	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Geometry supplies component-owned measurements for the built-in field shell.
// Border and feedback styling are fixed by this package, not caller-configurable.
type Geometry struct {
	TargetPart                                    style.Part
	Height, MinHeight, Radius, PaddingX, PaddingY unit.Dp
	TextSize, LineHeight                          unit.Sp
	FillWidth                                     bool
}

// fieldDuration uses the shared control duration, falling back to 150ms.
func fieldDuration(activeTheme *theme.Theme) time.Duration {
	if activeTheme.Motion.ControlDuration > 0 {
		return activeTheme.Motion.ControlDuration
	}
	return 150 * time.Millisecond
}

// Defaults builds the one shared field appearance. Callers own geometry, layout and state.
func Defaults(activeTheme *theme.Theme, variant Variant, options Geometry) style.Style {
	background, hover := surfaceColors(activeTheme, variant)
	const borderWidth unit.Dp = 1
	sem := activeTheme.Semantic()
	accentTone := sem.Tone(theme.TonePrimary)
	dangerTone := sem.Tone(theme.ToneDanger)
	warningTone := sem.Tone(theme.ToneWarning)
	accent, accentHover := accentTone.Fill, accentTone.Hover
	danger, dangerHover := dangerTone.Fill, dangerTone.Hover
	warning, warningHover := warningTone.Fill, warningTone.Hover

	// Focus and validation share the same inset stroke; feedback never changes geometry.
	stroke := func(col style.ColorSource) style.Style {
		return style.Style{}.BorderWidth(borderWidth).BorderColor(col)
	}

	focusRow := style.Rule{When: style.Focused}
	switch variant {
	case Primary, Secondary:
		focusRow.Set = stroke(style.Solid(sem.Ring))
	case Filled:
		focusRow.Set = stroke(style.Solid(accent))
	case Borderless:
		focusRow.Set = style.Style{}
	case Underlined:
		focusRow.Set = style.Style{}.BorderBottom(borderWidth, style.Solid(sem.Ring))
	}
	if variant != Borderless && variant != Underlined {
		focusRow.Set = focusRow.Set.BoxShadowTransient(0, 0, 5, 0, style.WithAlpha(accent, .60))
	}
	// Bordered fields darken their stroke on hover; the tinted variants keep
	// their own hover face, and the underlined rule is a hover affordance of its own.
	hoverSet := style.Style{}.Background(hover)
	switch variant {
	case Primary, Secondary:
		hoverSet = hoverSet.BorderColor(style.Solid(theme.Mix(sem.Border, sem.Foreground, 18)))
	case Underlined:
		hoverSet = style.Style{}.BorderBottom(borderWidth, style.Solid(accentHover))
	}
	invalidSet := stroke(style.Solid(danger))
	invalidHoverSet := invalidSet
	invalidFocusedSet := stroke(style.Solid(danger))

	switch variant {
	case Filled:
		invalidSet, invalidHoverSet = validationTint(sem, theme.ToneDanger)
		invalidFocusedSet = stroke(style.Solid(danger))
	case Borderless:
		invalidSet = style.Style{}.TextColor(style.Solid(danger))
		invalidHoverSet = invalidSet
		invalidFocusedSet = invalidSet
	case Underlined:
		invalidSet = style.Style{}.BorderBottom(borderWidth, style.Solid(danger))
		invalidHoverSet = style.Style{}.BorderBottom(borderWidth, style.Solid(dangerHover))
		invalidFocusedSet = style.Style{}.BorderBottom(borderWidth, style.Solid(danger))
	}
	warningSet := stroke(style.Solid(warning))
	warningHoverSet := warningSet
	warningFocusedSet := stroke(style.Solid(warning))
	switch variant {
	case Filled:
		warningSet, warningHoverSet = validationTint(sem, theme.ToneWarning)
		warningFocusedSet = stroke(style.Solid(warning))
	case Borderless:
		warningSet = style.Style{}.TextColor(style.Solid(warning))
		warningHoverSet = warningSet
		warningFocusedSet = warningSet
	case Underlined:
		warningSet = style.Style{}.BorderBottom(borderWidth, style.Solid(warning))
		warningHoverSet = style.Style{}.BorderBottom(borderWidth, style.Solid(warningHover))
		warningFocusedSet = style.Style{}.BorderBottom(borderWidth, style.Solid(warning))
	}
	duration := fieldDuration(activeTheme)
	content := style.Style{}.
		Background(background).
		TextColor(style.Solid(sem.FieldForeground)).
		Radius(options.Radius).
		Overflow(style.OverflowHidden).
		Cursor(pointer.CursorText).
		Opacity(1).
		Transition(style.PropBackgroundColor, duration).
		Transition(style.PropTextColor, duration).
		Transition(style.PropBorderColor, duration).
		Transition(style.PropOutlineColor, duration).
		Transition(style.PropOpacity, duration)
	switch variant {
	case Underlined:
		content = content.Radius(0).BorderBottom(borderWidth, style.Solid(sem.Border))
	case Filled, Borderless:
		// The filled and borderless variants rest without a box border; focus draws one.
	default:
		content = content.BorderColor(style.Solid(sem.Border)).BorderWidth(borderWidth)
	}
	if options.Height > 0 {
		content = content.Height(options.Height)
	}
	if options.MinHeight > 0 {
		content = content.MinHeight(options.MinHeight)
	}
	if options.PaddingX != 0 || options.PaddingY != 0 {
		content = content.Padding(options.PaddingY, options.PaddingX)
	}
	if options.TextSize > 0 {
		content = content.FontSize(options.TextSize)
	}
	if options.LineHeight > 0 {
		content = content.LineHeight(options.LineHeight)
	}
	if options.FillWidth {
		content = content.FillWidth()
	}

	// Disabled must override the more specific validation and focus rules.
	content = content.Rules(
		style.Rule{
			When: style.Hovered,
			Set:  hoverSet,
		},
		focusRow,
		style.Rule{
			When: style.Invalid,
			Set:  invalidSet,
		},
		style.Rule{
			When: style.Invalid | style.Hovered,
			Set:  invalidHoverSet,
		},
		style.Rule{
			When: style.Invalid | style.Focused,
			Set:  invalidFocusedSet,
		},
		style.Rule{
			When: style.Warning,
			Set:  warningSet,
		},
		style.Rule{
			When: style.Warning | style.Hovered,
			Set:  warningHoverSet,
		},
		style.Rule{
			When: style.Warning | style.Focused,
			Set:  warningFocusedSet,
		},
		// A read-only field shows the disabled face but keeps the text cursor
		// and its focus ring: it is still selectable and focusable.
		style.Rule{
			When: style.ReadOnly,
			Not:  style.Focused,
			Set:  readOnlySet(activeTheme, variant),
		},
		style.Rule{
			Rank: 2,
			When: style.Disabled,
			Set:  disabledSet(activeTheme, variant),
		},
	)
	return style.Style{}.
		Part(options.TargetPart, content).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(style.PartSelection, style.Style{}.Background(style.Solid(sem.Selection)))
}

// validationTint returns the filled variant's validation colors at rest and on hover.
func validationTint(sem theme.Semantic, tone theme.Tone) (resting, hovered style.Style) {
	pair := sem.Tone(tone)
	return style.Style{}.
			Background(style.Solid(pair.Soft)).
			TextColor(style.Solid(pair.OnSoft)),
		style.Style{}.
			Background(style.Solid(pair.SoftHover)).
			TextColor(style.Solid(pair.OnSoft))
}
