package toast

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/button"
)

// toastVisualKey contains only the inputs that change the built-in defaults.
// The instance style is resolved as the last cascade layer instead of being
// part of the key.
type toastVisualKey struct {
	variant       Variant
	actionVariant button.Variant
}

type toastVisualCache struct {
	cache computed.VisualCacheMap[toastVisualKey, toastVisual]
}

func (p Widget) baseVisual(ctx *core.Context) toastVisual {
	return p.visualWithCache(ctx, Default, button.Ghost)
}

func (p Widget) visualWithCache(ctx *core.Context, variant Variant, actionVariant button.Variant) toastVisual {
	active := core.ActiveTheme(ctx)
	key := toastVisualKey{variant: variant, actionVariant: actionVariant}
	load := func() toastVisual {
		return buildToastVisual(active, variant, actionVariant)
	}
	if p.visualCache == nil {
		return load()
	}
	return p.visualCache.cache.Resolve(active, key, load)
}

// toastResolvedStyle is the cascade result for one toast: the panel, title,
// description and indicator layers with the instance style applied last.
type toastResolvedStyle struct {
	panel       style.ResolvedStyle
	title       style.ResolvedStyle
	description style.ResolvedStyle
	indicator   style.ResolvedStyle
}

func (p Widget) resolveStyle(ctx *core.Context, gtx layout.Context, state style.State, visual toastVisual) toastResolvedStyle {
	key := toastStyleKey(ctx, "toast")
	panelParts := [1]style.Part{style.PartRoot}
	textParts := [2]style.Part{style.PartLabel, style.PartDescription}
	var panel [1]style.ResolvedStyle
	var text [2]style.ResolvedStyle
	computed.ResolvePartsInto(ctx, gtx, key, state, panelParts[:], panel[:],
		visual.panel, style.Style{}, style.Style{}, p.style)
	computed.ResolvePartsInto(ctx, gtx, key, state, textParts[:], text[:],
		visual.text, style.Style{}, style.Style{}, p.style)
	return toastResolvedStyle{
		panel:       panel[0],
		title:       text[0],
		description: text[1],
		indicator: computed.ResolvePart(ctx, gtx, key, style.PartIndicator, state,
			visual.indicator, style.Style{}, style.Style{}, p.style),
	}
}

func (p Widget) resolveActionStyle(ctx *core.Context, gtx layout.Context, state style.State, visual toastVisual) style.ResolvedStyle {
	return computed.Resolve(ctx, gtx, toastStyleKey(ctx, "toast-action"), state,
		visual.action, style.Style{}, style.Style{}, p.style)
}

func (p Widget) resolveCloseStyle(ctx *core.Context, gtx layout.Context, state style.State, visual toastVisual) style.ResolvedStyle {
	return computed.Resolve(ctx, gtx, toastStyleKey(ctx, "toast-close"), state,
		visual.close, style.Style{}, style.Style{}, p.style)
}

// toastStyleKey derives an animation identity below the current toast item.
// An empty owner means there is no item scope, so transitions soft-fail.
func toastStyleKey(ctx *core.Context, role string) string {
	owner := core.FullKey(ctx, "")
	if owner == "" {
		return ""
	}
	return core.DerivedKey(ctx, owner, role)
}
