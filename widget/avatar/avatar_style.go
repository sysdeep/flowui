package avatar

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type avatarResolvedStyle struct{ root, label, icon style.ResolvedStyle }

type avatarVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type avatarVisualCache = computed.VisualCache[avatarVisualKey, avatarVisual]

func (a Widget) visualWithCache(ctx *core.Context) avatarVisual {
	active := core.ActiveTheme(ctx)
	key := avatarVisualKey{color: a.color, variant: a.variant, size: a.size}
	return a.visualCache.Resolve(active, key, func() avatarVisual {
		return buildAvatarVisual(active, a.color, a.variant, a.size)
	})
}

func (a Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual avatarVisual) avatarResolvedStyle {
	state := style.State{}
	resolved := avatarResolvedStyle{
		root:  computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, a.style),
		label: computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, a.style),
		icon:  computed.ResolvePartStatic(ctx, style.PartIcon, state, visual.Defaults, visual.Variant, visual.Size, a.style),
	}
	if !computed.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := core.ClaimDerivedKey(ctx, core.KindStyle, a.key, "style")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.label)
	resolved.icon = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIcon, resolved.icon)
	return resolved
}
