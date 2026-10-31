package splitpane

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type splitPaneStyle struct {
	track        style.ResolvedStyle
	indicator    style.ResolvedStyle
	dividerWidth unit.Dp
	activeWidth  unit.Dp
	handleLength unit.Dp
	cursor       pointer.Cursor
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State) (style.ResolvedStyle, splitPaneStyle) {
	tokens := defaultSplitPaneMetrics
	cursor := pointer.CursorColResize
	if s.orientation == Vertical {
		cursor = pointer.CursorRowResize
	}
	defaults := defaultStyles(core.ActiveTheme(ctx), s.orientation, state)

	root := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults,
		style.Style{},
		style.Style{},
		s.customStyle,
	)
	track := computed.ResolvePart(ctx, gtx, key, style.PartTrack, state, defaults, style.Style{}, style.Style{}, s.customStyle)
	indicator := computed.ResolvePart(ctx, gtx, key, style.PartIndicator, state, defaults, style.Style{}, style.Style{}, s.customStyle)
	result := splitPaneStyle{
		track: track, indicator: indicator,
		dividerWidth: tokens.DividerWidth, activeWidth: tokens.ActiveWidth,
		handleLength: tokens.HandleLength,
		cursor:       cursor,
	}
	if track.Box != nil {
		if track.Box.Width != nil {
			result.dividerWidth = *track.Box.Width
		}
		if track.Box.Cursor != nil {
			result.cursor = *track.Box.Cursor
		}
	}
	if indicator.Box != nil {
		if indicator.Box.Width != nil {
			result.activeWidth = *indicator.Box.Width
		}
		if indicator.Box.Height != nil {
			result.handleLength = *indicator.Box.Height
		}
	}
	return root, result
}
