package alert

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type alertVisualKey struct {
	status Status
}

type alertVisualCache = computed.VisualCache[alertVisualKey, alertVisual]

func (a Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := a.visualWithCache(ctx)
	return a.layoutWithVisual(ctx, gtx, visual)
}

func (a Widget) visualWithCache(ctx *core.Context) alertVisual {
	active := core.ActiveTheme(ctx)
	key := alertVisualKey{status: a.status}
	load := func() alertVisual {
		return buildAlertVisual(active, a.status)
	}
	if a.visualCache == nil {
		return load()
	}
	return a.visualCache.Resolve(active, key, load)
}

type alertResolvedStyle struct {
	root        style.ResolvedStyle
	title       style.ResolvedStyle
	description style.ResolvedStyle
	content     style.ResolvedStyle
	indicator   style.ResolvedStyle
}

func (a Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual alertVisual) alertResolvedStyle {
	state := style.State{}
	resolved := alertResolvedStyle{
		root:        computed.ResolveStatic(ctx, state, visual.Defaults, style.Style{}, style.Style{}, a.style),
		title:       computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Title, style.Style{}, style.Style{}, a.style),
		description: computed.ResolvePartStatic(ctx, style.PartDescription, state, visual.Description, style.Style{}, style.Style{}, a.style),
		content:     computed.ResolvePartStatic(ctx, style.PartContent, state, visual.Content, style.Style{}, style.Style{}, a.style),
		indicator:   computed.ResolvePartStatic(ctx, style.PartIndicator, state, visual.Indicator, style.Style{}, style.Style{}, a.style),
	}
	if !computed.HasTransitions(resolved.root, resolved.title, resolved.description, resolved.content, resolved.indicator) {
		return resolved
	}
	key := core.ClaimDerivedKey(ctx, core.KindStyle, a.key, "style")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.title = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.title)
	resolved.description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, resolved.description)
	resolved.content = computed.ApplyPartTransitions(ctx, gtx, key, style.PartContent, resolved.content)
	resolved.indicator = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIndicator, resolved.indicator)
	return resolved
}
