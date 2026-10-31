package splitpane

import (
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type splitPaneMetrics struct {
	DividerWidth unit.Dp
	HitSize      unit.Dp
	ActiveWidth  unit.Dp
	HandleLength unit.Dp
}

var defaultSplitPaneMetrics = splitPaneMetrics{DividerWidth: 1, HitSize: 12, ActiveWidth: 2, HandleLength: 32}

func defaultStyles(activeTheme *theme.Theme, orientation Orientation, state style.State) style.Style {
	sem := activeTheme.Semantic()
	tokens := defaultSplitPaneMetrics
	base := sem.Separator
	if state.Disabled {
		base = activeTheme.DisabledColor(base)
	}
	cursor := pointer.CursorColResize
	if orientation == Vertical {
		cursor = pointer.CursorRowResize
	}
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.
			Width(tokens.DividerWidth).
			Background(style.SolidColor{Color: base}).
			Cursor(cursor)).
		Part(style.PartIndicator, style.Style{}.
			Width(tokens.ActiveWidth).
			Height(tokens.HandleLength).
			Background(style.SolidColor{Color: sem.Tone(theme.TonePrimary).Fill}).
			Radius(tokens.ActiveWidth/2))
}
