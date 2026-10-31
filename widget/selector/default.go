package selector

import (
	"math"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	selectDefaultHeight            unit.Dp       = 32
	selectDefaultRadius            unit.Dp       = 4
	selectDefaultTextSize          unit.Sp       = 14
	selectDefaultContentGap        unit.Dp       = 4
	selectDefaultTriggerPaddingX   unit.Dp       = 8
	selectDefaultTriggerPaddingY   unit.Dp       = 5
	selectDefaultIndicatorWidth    unit.Dp       = 28
	selectDefaultIndicatorSize     unit.Dp       = 16
	selectDefaultIndicatorStroke   unit.Dp       = 1.7
	selectDefaultPanelGap          unit.Dp       = 6
	selectDefaultPanelRadius       unit.Dp       = 8
	selectDefaultPanelMaxHeight    unit.Dp       = 280
	selectDefaultPanelPadding      unit.Dp       = 4
	selectDefaultPanelBorderWidth  unit.Dp       = 1
	selectDefaultAnimationScale    float32       = 1
	selectDefaultAnimationDistance unit.Dp       = 4
	selectDefaultOpenDuration      time.Duration = 150 * time.Millisecond
	selectDefaultCloseDuration     time.Duration = 100 * time.Millisecond
	selectDefaultIndicatorDuration time.Duration = 150 * time.Millisecond
)

// selectDefaults is the theme-derived select appearance. It contains the
// declaration layers and geometry the layout needs. The instance style is
// resolved separately as the last cascade layer and is deliberately not part
// of this cache value.
type selectDefaults struct {
	trigger     style.Style
	placeholder style.Style
	indicator   style.Style
	icon        style.Style
	panel       style.Style
	item        style.Style

	textSize        unit.Sp
	indicatorWidth  unit.Dp
	indicatorSize   unit.Dp
	indicatorStroke unit.Dp
	contentGap      unit.Dp
	panelGap        int
	itemTextSize    unit.Sp
}

// Appearance is the shared visual contract for select-like controls. It is
// intentionally data-only so editable controls such as Combobox can reuse the
// Select trigger, panel and item cascade without depending on Select's state
// machine.
type Appearance struct {
	Trigger, Placeholder, Indicator, Icon, Panel, Item style.Style
	TextSize                                           unit.Sp
	IndicatorWidth                                     unit.Dp
	IndicatorSize                                      unit.Dp
	IndicatorStroke                                    unit.Dp
	ContentGap                                         unit.Dp
	PanelGap                                           int
	ItemTextSize                                       unit.Sp
}

// DefaultAppearance returns the active theme's Select appearance for a
// select-like control.
func DefaultAppearance(active *theme.Theme, variant SelectVariant) Appearance {
	defaults := buildSelectDefaults(active, variant)
	return Appearance{
		Trigger: defaults.trigger, Placeholder: defaults.placeholder,
		Indicator: defaults.indicator, Icon: defaults.icon,
		Panel: defaults.panel, Item: defaults.item,
		TextSize: defaults.textSize, IndicatorWidth: defaults.indicatorWidth,
		IndicatorSize: defaults.indicatorSize, IndicatorStroke: defaults.indicatorStroke,
		ContentGap: defaults.contentGap, PanelGap: defaults.panelGap,
		ItemTextSize: defaults.itemTextSize,
	}
}

// buildSelectDefaults derives the trigger and option appearance from the
// active theme. The select owns its own trigger parts while option rows reuse
// the shared list-box metrics and colours.
func buildSelectDefaults(active *theme.Theme, variant SelectVariant) selectDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	primary := sem.Tone(theme.TonePrimary)
	danger := sem.Tone(theme.ToneDanger).Fill

	duration := selectFieldDuration(active)
	indicatorDuration := selectDefaultIndicatorDuration
	borderWidth := unit.Dp(1)
	focusWidth := unit.Dp(1)

	background, hover := sem.FieldBackground, sem.FieldHover
	if variant == SelectSecondary {
		neutral := sem.Tone(theme.ToneDefault)
		background, hover = neutral.Fill, neutral.Hover
	}

	trigger := style.New().
		Height(selectDefaultHeight).
		Padding(selectDefaultTriggerPaddingY, selectDefaultTriggerPaddingX).
		Radius(selectDefaultRadius).
		Background(style.Solid(background)).
		TextColor(style.Solid(sem.FieldForeground)).
		FontSize(selectDefaultTextSize).
		BorderWidth(borderWidth).
		BorderColor(style.Solid(sem.Border)).
		Outline(1, 0, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Cursor(pointer.CursorPointer).
		Transition(style.PropBackgroundColor, duration).
		Transition(style.PropTextColor, duration).
		Transition(style.PropBorderColor, duration).
		Transition(style.PropOutlineColor, duration).
		Transition(style.PropOpacity, duration).
		On(style.Hovered, style.New().
			Background(style.Solid(hover)).
			BorderColor(style.Solid(theme.Mix(sem.Border, sem.Foreground, 18)))).
		On(style.FocusVisible, style.New().
			BorderWidth(focusWidth).
			BorderColor(style.Solid(sem.Ring))).
		On(style.Invalid, style.New().
			BorderColor(style.Solid(danger))).
		On(style.Invalid|style.FocusVisible, style.New().
			BorderWidth(focusWidth).
			BorderColor(style.Solid(danger))).
		On(style.Open, style.New().Background(style.Solid(hover))).
		Rules(style.Rule{Rank: 2, When: style.Disabled, Set: selectDisabledStyle(active)})

	placeholder := style.New().Part(style.PartPlaceholder, style.New().
		TextColor(style.Solid(sem.FieldPlaceholder)))

	indicator := style.New().Part(style.PartIndicator, style.New().
		Width(selectDefaultIndicatorWidth).
		TextColor(style.Solid(sem.FieldPlaceholder)))

	icon := style.New().Part(style.PartIcon, style.New().
		Width(selectDefaultIndicatorSize).
		Height(selectDefaultIndicatorSize).
		TextColor(style.Solid(sem.FieldPlaceholder)).
		Transition(style.PropTransform, indicatorDuration).
		On(style.Open, style.New().Rotate(float32(math.Pi))))

	panel := style.New().Part(style.PartPanel, style.New().
		MaxHeight(selectDefaultPanelMaxHeight).
		Background(style.Solid(sem.Overlay)).
		TextColor(style.Solid(sem.OverlayForeground)).
		BorderWidth(selectDefaultPanelBorderWidth).
		BorderColor(style.Solid(sem.Border)).
		Radius(selectDefaultPanelRadius).
		Padding(selectDefaultPanelPadding).
		Shadow(style.ShadowMenu).
		Overflow(style.OverflowHidden))

	item := style.New().
		MinHeight(32).
		FillWidth().
		Padding(6, 10).
		Radius(4).
		Background(style.RGBA(0x00000000)).
		TextColor(style.Solid(sem.Foreground)).
		BorderWidth(2).
		BorderColor(style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Transition(style.PropBackgroundColor, duration).
		Transition(style.PropBorderColor, duration).
		Transition(style.PropTextColor, duration).
		Rules(
			style.Rule{When: style.Hovered, Set: style.New().
				Background(style.Solid(sem.SurfaceRaised))},
			style.Rule{When: style.Pressed, Set: style.New().
				Background(style.Solid(sem.SurfacePressed))},
			style.Rule{When: style.FocusVisible, Set: style.New().
				BorderColor(style.Solid(sem.Ring))},
			style.Rule{When: style.Selected, Set: style.New().
				Background(style.Solid(primary.Soft))},
			style.Rule{When: style.Selected | style.Hovered, Set: style.New().
				Background(style.Solid(primary.SoftHover))},
			style.Rule{When: style.Selected | style.Pressed, Set: style.New().
				Background(style.Solid(primary.SoftHover))},
			style.Rule{Rank: 2, When: style.Disabled, Set: style.New().
				Opacity(active.DisabledOpacityValue()).
				TextColor(style.Solid(active.DisabledColor(sem.Foreground)))},
		)

	return selectDefaults{
		trigger:         trigger,
		placeholder:     placeholder,
		indicator:       indicator,
		icon:            icon,
		panel:           panel,
		item:            style.New().Part(style.PartItem, item),
		textSize:        selectDefaultTextSize,
		indicatorWidth:  selectDefaultIndicatorWidth,
		indicatorSize:   selectDefaultIndicatorSize,
		indicatorStroke: selectDefaultIndicatorStroke,
		contentGap:      selectDefaultContentGap,
		panelGap:        int(selectDefaultPanelGap),
		itemTextSize:    14,
	}
}

func selectFieldDuration(active *theme.Theme) time.Duration {
	if active.Motion.ControlDuration > 0 {
		return active.Motion.ControlDuration
	}
	return 150 * time.Millisecond
}

func selectDisabledStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	if sem.DisabledBackground.A == 0 {
		return style.New().Opacity(active.DisabledOpacityValue())
	}
	return style.New().
		Background(style.Solid(sem.DisabledBackground)).
		TextColor(style.Solid(sem.DisabledForeground)).
		BorderColor(style.Solid(sem.DisabledBorder)).
		Cursor(pointer.CursorNotAllowed).
		Opacity(1)
}
