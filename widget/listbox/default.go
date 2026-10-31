package listbox

import (
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	listBoxDefaultColorDuration     = 100 * time.Millisecond
	listBoxDefaultFocusDuration     = 100 * time.Millisecond
	listBoxDefaultPressInDuration   = 80 * time.Millisecond
	listBoxDefaultPressOutDuration  = 140 * time.Millisecond
	listBoxDefaultSelectionDuration = 200 * time.Millisecond
)

// listBoxDefaults is the theme-derived list appearance and geometry. The
// instance style is resolved separately and is deliberately not part of this
// cache value.
type listBoxDefaults struct {
	itemDefault style.Style
	itemDanger  style.Style

	padding               unit.Dp
	gap                   unit.Dp
	maxHeight             unit.Dp
	sectionHeaderPaddingX unit.Dp
	sectionHeaderPaddingY unit.Dp
	sectionHeaderTextSize unit.Sp
	itemMinHeight         unit.Dp
	itemContentGap        unit.Dp
	itemTextSize          unit.Sp
	itemDescriptionSize   unit.Sp
	itemIndicatorSize     unit.Dp
	itemIndicatorInset    unit.Dp
	itemIndicatorStroke   unit.Dp
	selectionDuration     time.Duration
	motion                theme.MotionTheme
}

func (d listBoxDefaults) item(variant ItemVariant) style.Style {
	if variant == ItemDanger {
		return d.itemDanger
	}
	return d.itemDefault
}

// buildListBoxDefaults derives the list appearance from the active theme. All
// fields are read on this call so copied themes and runtime replacement cannot
// retain stale values.
func buildListBoxDefaults(active *theme.Theme) listBoxDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return listBoxDefaults{
		itemDefault: buildListBoxItemDefaults(active, ItemDefault),
		itemDanger:  buildListBoxItemDefaults(active, ItemDanger),
		padding:     4, gap: 0, maxHeight: 280,
		sectionHeaderPaddingX: 10, sectionHeaderPaddingY: 4,
		sectionHeaderTextSize: 12, itemMinHeight: 32, itemContentGap: 12,
		itemTextSize: 14, itemDescriptionSize: 12,
		itemIndicatorSize: 18, itemIndicatorInset: 4,
		itemIndicatorStroke: 1.7,
		selectionDuration:   listBoxDefaultSelectionDuration,
		motion:              active.Motion,
	}
}

func buildListBoxItemDefaults(active *theme.Theme, variant ItemVariant) style.Style {
	sem := active.Semantic()
	pressed := listBoxPressedScale(1)
	foreground := sem.Foreground
	if variant == ItemDanger {
		foreground = sem.Tone(theme.ToneDanger).Fill
	}
	item := style.New().
		MinHeight(32).
		Padding(6, 10).
		Radius(4).
		Background(style.RGBA(0x00000000)).
		TextColor(style.Solid(foreground)).
		BorderWidth(2).
		BorderColor(style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Transition(style.PropBackgroundColor, listBoxDefaultColorDuration).
		Transition(style.PropBorderColor, listBoxDefaultFocusDuration).
		Transition(style.PropTransform, listBoxDefaultPressOutDuration)

	itemPart := func(declaration style.Style) style.Style {
		return style.New().Part(style.PartItem, declaration)
	}
	defaults := itemPart(item).Rules(
		style.Rule{
			When: style.Hovered,
			Set:  itemPart(style.New().Background(style.Solid(sem.SurfaceRaised))),
		},
		style.Rule{
			When: style.Pressed,
			Set: itemPart(style.New().
				Background(style.Solid(sem.SurfacePressed)).
				Scale(pressed).
				Transition(style.PropTransform, listBoxDefaultPressInDuration)),
		},
		style.Rule{
			When: style.FocusVisible,
			Set:  itemPart(style.New().BorderColor(style.Solid(sem.Ring))),
		},
		style.Rule{
			When: style.Selected,
			Set:  itemPart(style.New().Background(style.Solid(sem.Tone(theme.TonePrimary).Soft))),
		},
		style.Rule{
			When: style.Selected | style.Hovered,
			Set:  itemPart(style.New().Background(style.Solid(sem.Tone(theme.TonePrimary).SoftHover))),
		},
		style.Rule{
			When: style.Disabled,
			Set: itemPart(style.New().
				Opacity(active.DisabledOpacityValue()).
				TextColor(style.Solid(active.DisabledColor(foreground)))),
		},
	)

	if variant == ItemDanger {
		danger := sem.Tone(theme.ToneDanger).Fill
		defaults = defaults.Rules(style.Rule{
			When: style.FocusVisible,
			Set:  itemPart(style.New().BorderColor(style.WithAlpha(danger, float32(sem.Ring.A)/255))),
		})
	}
	return defaults
}

func listBoxPressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return 0.98
	}
	return value
}
