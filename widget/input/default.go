package input

import (
	"time"

	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/internal/fieldstyle"
)

// defaultInputRules declares the built-in size tiers and field feedback.
func defaultInputRules(active *theme.Theme, variant Variant, size Size, fullWidth bool) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	height, radius, padding := unit.Dp(32), unit.Dp(4), unit.Dp(8)
	switch size {
	case Small:
		height, padding = 24, 6
	case Large:
		height, radius, padding = 40, 6, 12
	}
	return fieldstyle.Defaults(active, variant, fieldstyle.Geometry{
		Height: height, Radius: radius, PaddingX: padding,
		TextSize: active.Typography.ControlSize, LineHeight: 20,
		FillWidth: fullWidth,
	})
}

// defaultTextAreaRules sizes the field from its rows and built-in minimum.
func defaultTextAreaRules(active *theme.Theme, variant Variant, fullWidth bool, height unit.Dp) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return fieldstyle.Defaults(active, variant, fieldstyle.Geometry{
		Height: max(64, height), Radius: 4, PaddingX: 8, PaddingY: 8,
		TextSize: active.Typography.ControlSize, LineHeight: 20,
		FillWidth: fullWidth,
	})
}

// defaultGroupRules adds editor, slot, and divider parts to the shared field shell.
func defaultGroupRules(active *theme.Theme, variant Variant, fullWidth bool, minHeight unit.Dp) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	root := fieldstyle.Defaults(active, variant, fieldstyle.Geometry{
		MinHeight: max(32, minHeight), Radius: 4,
		TextSize: active.Typography.ControlSize, LineHeight: 20,
		FillWidth: fullWidth,
	})
	sem := active.Semantic()
	content := style.New().Padding(8).MinHeight(64).Cursor(pointer.CursorText)
	return root.
		Rules(style.Rule{
			Rank: 2, When: style.Disabled,
			Set: style.New().Part(style.PartContent, style.New().Cursor(pointer.CursorNotAllowed)),
		}).
		Cursor(pointer.CursorDefault).
		Part(style.PartContent, content).
		Part(style.PartPrefix, style.New().TextColor(style.Solid(sem.MutedForeground))).
		Part(style.PartSuffix, style.New().TextColor(style.Solid(sem.MutedForeground))).
		Part(style.PartIndicator, style.New().Width(0).Background(style.Solid(sem.Border)))
}

const (
	groupActionSize     = unit.Dp(24)
	groupActionDuration = 100 * time.Millisecond
)

// groupActionDefaults is the theme-derived appearance of a compact input-group
// action. The action's instance style is resolved separately as the last
// cascade layer, so it never enters the cache key.
type groupActionDefaults struct {
	declaration style.Style
	slotPadding unit.Dp
}

func defaultGroupAction(active *theme.Theme) groupActionDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return groupActionDefaults{
		declaration: groupActionRules(active),
		slotPadding: 4,
	}
}

func groupActionRules(activeTheme *theme.Theme) style.Style {
	sem := activeTheme.Semantic()
	neutral := sem.Tone(theme.ToneDefault)
	return style.Style{}.
		MinWidth(groupActionSize).
		Height(groupActionSize).
		Padding(0, 4).
		Radius(4).
		TextColor(style.Solid(sem.MutedForeground)).
		Cursor(pointer.CursorPointer).
		Outline(unit.Dp(2), 0, style.WithAlpha(sem.Ring, 0)).
		Transition(style.PropBackgroundColor, groupActionDuration).
		Transition(style.PropOutlineColor, groupActionDuration).
		Transition(style.PropTextColor, groupActionDuration).
		Rules(
			style.Rule{
				When: style.Hovered,
				Set: style.Style{}.
					Background(style.Solid(neutral.Hover)).
					TextColor(style.Solid(sem.Foreground)),
			},
			style.Rule{
				When: style.Pressed,
				Set:  style.Style{}.Background(style.Solid(neutral.Fill)),
			},
			style.Rule{
				When: style.FocusVisible,
				Not:  style.Disabled,
				Set:  style.Style{}.Outline(2, 0, style.Solid(sem.Ring)),
			},
			style.Rule{
				When: style.Disabled,
				Set: style.Style{}.
					Cursor(pointer.CursorDefault).
					Opacity(0.45),
			},
		)
}
