package selector

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotSelectDefaults = "select-defaults"

type selectDefaultsKey struct {
	variant SelectVariant
}

type selectDeclarationCache = computed.VisualCache[selectDefaultsKey, selectDefaults]

// declarations resolves the theme-derived appearance for the current variant.
// The instance style is deliberately not part of the cache key.
func (s SelectWidget) declarations(active *theme.Theme, cache *selectDeclarationCache) selectDefaults {
	key := selectDefaultsKey{variant: s.variant}
	return cache.Resolve(active, key, func() selectDefaults {
		return buildSelectDefaults(active, s.variant)
	})
}

func (s SelectWidget) sizeStyle() style.Style {
	if !s.fullWidth {
		return style.Style{}
	}
	return style.New().FillWidth()
}

func (s SelectWidget) trigger(state base.SelectState, defaults selectDefaults, triggerWidth *int) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		styleState := s.triggerState(state)
		resolved := computed.Resolve(
			ctx,
			gtx,
			state.Key,
			styleState,
			defaults.trigger,
			style.Style{},
			s.sizeStyle(),
			s.style,
		)
		placeholder := computed.ResolvePart(
			ctx,
			gtx,
			state.Key,
			style.PartPlaceholder,
			styleState,
			defaults.placeholder,
			style.Style{},
			style.Style{},
			s.style,
		)
		icon := computed.ResolvePart(
			ctx,
			gtx,
			state.Key,
			style.PartIcon,
			styleState,
			defaults.icon,
			style.Style{},
			style.Style{},
			s.style,
		)
		indicator := computed.ResolvePart(
			ctx,
			gtx,
			state.Key,
			style.PartIndicator,
			styleState,
			defaults.indicator,
			style.Style{},
			style.Style{},
			s.style,
		)
		dims := layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return s.layoutTriggerContent(ctx, gtx, state, defaults, resolved, placeholder, icon, indicator)
		}))
		*triggerWidth = dims.Size.X
		return dims
	})
}

func (s SelectWidget) triggerState(state base.SelectState) style.State {
	return style.State{
		Hovered:      state.Hovered,
		Pressed:      state.Pressed,
		Focused:      state.Focused,
		FocusVisible: state.FocusVisible,
		Disabled:     state.Disabled,
		Invalid:      s.invalid,
		Open:         state.Open,
		Selected:     s.selected != "",
	}
}

func selectResolvedColor(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Text != nil {
		if value, ok := computed.Color(resolved.Text.Color); ok {
			return value
		}
	}
	return fallback
}
