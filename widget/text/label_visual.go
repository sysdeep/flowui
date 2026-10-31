package text

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type labelVisual struct {
	Defaults, Required style.Style
	RequiredMarkGap    unit.Dp
}

type labelVisualKey struct {
	Foreground color.NRGBA
}

type labelVisualCache = computed.VisualCache[labelVisualKey, labelVisual]

func (l LabelWidget) layoutWithStyle(
	ctx *core.Context,
	gtx layout.Context,
) layout.Dimensions {
	disabled := l.disabled || !gtx.Enabled()
	state := style.State{Disabled: disabled, Invalid: l.invalid}
	visual := l.visualWithCache(ctx)
	if l.forKey == "" {
		resolved := l.resolveLayoutStyle(ctx, gtx, state, visual)
		return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return l.layoutContent(ctx, gtx, resolved, visual.RequiredMarkGap)
		}))
	}

	fieldKey := l.registerFieldAssociation(ctx)
	key, clickable := core.DerivedClickableWithKey(ctx, l.forKey, "label")
	presses := core.SnapshotPresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			core.RequestFieldFocusVisible(ctx, fieldKey, presses.ClickFocusVisible(clickable.History()))
		}
	}
	state.Hovered = clickable.Hovered()
	state.Pressed = clickable.Pressed()
	state.Focused = gtx.Focused(clickable)
	state.FocusVisible = core.FocusVisible(ctx, clickable, state.Focused)
	resolved := l.resolveStyle(ctx, gtx, key, state, visual)
	child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return l.layoutContent(ctx, gtx, resolved, visual.RequiredMarkGap)
	})
	return layoutui.InteractiveResolved(ctx, gtx, resolved.root, child, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return clickable.Layout(gtx, visual)
	})
}

func labelResolveLabelVisual(active *theme.Theme, foreground color.NRGBA) labelVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}

	sem := active.Semantic()
	danger := sem.Tone(theme.ToneDanger).Fill
	defaults := style.Style{}.
		FontSize(active.Typography.ControlSize).
		FontWeight(500).
		TextColor(style.Solid(foreground)).
		Rules(
			style.Rule{
				When: style.Invalid,
				Set:  style.Style{}.TextColor(style.Solid(danger)),
			},
			style.Rule{
				Rank: 1,
				When: style.Disabled,
				Set: style.Style{}.
					TextColor(style.Solid(active.DisabledColor(foreground))).
					On(style.Invalid, style.Style{}.
						TextColor(style.Solid(active.DisabledColor(danger)))),
			},
		)
	required := style.Style{}.
		TextColor(style.Solid(danger)).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.TextColor(style.Solid(active.DisabledColor(danger))),
		})
	return labelVisual{
		Defaults:        defaults,
		Required:        required,
		RequiredMarkGap: 2,
	}
}
