package chip

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type chipResolvedStyle struct {
	root          style.ResolvedStyle
	label         style.ResolvedStyle
	icon          style.ResolvedStyle
	labelPaddingX unit.Dp
	contentGap    unit.Dp
}

type chipVisualKey struct {
	color   Color
	variant Variant
	size    Size
}

type chipVisualCache = computed.VisualCache[chipVisualKey, chipVisual]

func (c Widget) visualWithCache(ctx *core.Context) chipVisual {
	active := core.ActiveTheme(ctx)
	key := chipVisualKey{color: c.color, variant: c.variant, size: c.size}
	return c.visualCache.Resolve(active, key, func() chipVisual {
		return buildChipVisual(active, c.color, c.variant, c.size)
	})
}

func (c Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual chipVisual) chipResolvedStyle {
	state := style.State{}
	resolved := chipResolvedStyle{
		root:          computed.ResolveStatic(ctx, state, visual.Defaults, visual.Variant, visual.Size, c.style),
		label:         computed.ResolvePartStatic(ctx, style.PartLabel, state, visual.Defaults, visual.Variant, visual.Size, c.style),
		icon:          computed.ResolvePartStatic(ctx, style.PartIcon, state, visual.Defaults, visual.Variant, visual.Size, c.style),
		labelPaddingX: visual.LabelPaddingX,
		contentGap:    visual.ContentGap,
	}
	if !computed.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := core.ClaimKey(ctx, core.KindStyle, "chip")
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, resolved.label)
	resolved.icon = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIcon, resolved.icon)
	return resolved
}
