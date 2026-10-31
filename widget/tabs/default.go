package tabs

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	tabsColorDuration     = 150 * time.Millisecond
	tabsIndicatorDuration = 250 * time.Millisecond
)

// tabsDefaults is the theme-derived appearance of one tabs variant. It is
// cached per theme and per variant/size/color; the instance style is resolved
// separately as the last cascade layer.
type tabsDefaults struct {
	root     style.Style
	panel    style.Style
	item     style.Style
	geometry tabsGeometry
	palette  tabsPalette
	timings  tabsTimings
	motion   theme.MotionTheme
}

// tabsTimings holds the label, indicator, and panel transition durations.
type tabsTimings struct {
	color     time.Duration
	indicator time.Duration
	panel     time.Duration
}

// tabsPalette holds the semantic colors tabs use while drawing chrome and
// item decorations.
type tabsPalette struct {
	listBackground   color.NRGBA
	scrollBackground color.NRGBA
	listBorder       color.NRGBA

	foreground                 color.NRGBA
	selectedForeground         color.NRGBA
	disabledSelectedForeground color.NRGBA
	indicator                  color.NRGBA
	disabledIndicator          color.NRGBA
	separator                  color.NRGBA
	disabledSeparator          color.NRGBA
	focus                      color.NRGBA
	scrollForeground           color.NRGBA
	disabledForeground         color.NRGBA
}

// itemColors returns the per-item decoration colors for an enabled or
// disabled tab.
func (p tabsPalette) itemColors(disabled bool) (selectedForeground, indicator, separator, focus color.NRGBA) {
	if disabled {
		return p.disabledSelectedForeground, p.disabledIndicator, p.disabledSeparator, color.NRGBA{}
	}
	return p.selectedForeground, p.indicator, p.separator, p.focus
}

func buildTabsDefaults(active *theme.Theme, key tabsDefaultsKey) tabsDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	palette := tabsPalette{
		listBackground:     sem.SurfaceHover,
		scrollBackground:   sem.Background,
		listBorder:         sem.Separator,
		foreground:         sem.MutedForeground,
		selectedForeground: sem.SegmentForeground,
		indicator:          sem.Segment,
		separator:          sem.MutedForeground,
		focus:              sem.Ring,
		scrollForeground:   sem.Foreground,
		disabledForeground: active.DisabledColor(sem.Foreground),
	}
	// The secondary strip draws only its divider, not a filled list surface.
	if key.variant == Secondary {
		palette.listBackground = color.NRGBA{}
		palette.selectedForeground = sem.Foreground
		palette.indicator = sem.Tone(theme.TonePrimary).Fill
	}
	if key.color == ColorAccent {
		palette.indicator = sem.Tone(theme.TonePrimary).Fill
		if key.variant == Secondary {
			palette.selectedForeground = sem.Tone(theme.TonePrimary).Fill
		} else {
			palette.selectedForeground = sem.Tone(theme.TonePrimary).OnFill
		}
	}
	palette.disabledSelectedForeground = active.DisabledColor(palette.selectedForeground)
	palette.disabledIndicator = active.DisabledColor(palette.indicator)
	palette.disabledSeparator = active.DisabledColor(palette.separator)

	fontSize, weight := unit.Sp(14), font.Medium
	switch key.size {
	case Small:
		weight = font.Normal
	case Large:
		fontSize = 16
	}
	declaration := style.New().
		FontSize(fontSize).
		FontWeight(int(weight)).
		Opacity(1).
		Transition(style.PropOpacity, theme.ResolveMotionDuration(active.Motion, tabsColorDuration)).
		OnNot(style.Hovered, style.Selected|style.Disabled, style.New().Opacity(.7)).
		On(style.Disabled, style.New().Opacity(active.DisabledOpacityValue()))

	panel := style.New().Padding(8).Opacity(1)
	return tabsDefaults{
		root:     style.New().TextColor(style.Solid(palette.foreground)),
		panel:    style.New().Part(style.PartPanel, panel),
		item:     style.New().Part(style.PartItem, declaration),
		geometry: tabsGeometryFor(key.variant),
		palette:  palette,
		timings: tabsTimings{
			color:     tabsColorDuration,
			indicator: tabsIndicatorDuration,
			panel:     200 * time.Millisecond,
		},
		motion: active.Motion,
	}
}
