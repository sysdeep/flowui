package badge

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type badgeResolvedStyle struct {
	root  style.ResolvedStyle
	label style.ResolvedStyle
}

type badgeVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type badgeVisualCache = computed.VisualCache[badgeVisualKey, badgeVisual]

func (b Widget) visualWithCache(ctx *core.Context) badgeVisual {
	active := core.ActiveTheme(ctx)
	load := func() badgeVisual {
		return buildBadgeVisual(active, b.color, b.variant, b.size)
	}
	key := badgeVisualKey{
		color: b.color, variant: b.variant, size: b.size,
	}
	return b.visualCache.Resolve(active, key, load)
}

func (b Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual badgeVisual) badgeResolvedStyle {
	state := style.State{}
	root := computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, b.style)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, b.style)
	if !computed.HasTransitions(root, label) {
		return badgeResolvedStyle{root: root, label: label}
	}
	key := core.ClaimKey(ctx, core.KindStyle, "badge")
	return badgeResolvedStyle{
		root:  computed.ApplyTransitions(ctx, gtx, key, root),
		label: computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label),
	}
}
