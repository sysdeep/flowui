package button

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

type buttonDefaults struct {
	defaults style.Style
	variant  style.Style
	size     style.Style
	spinner  unit.Dp
	motion   theme.MotionTheme
}

type buttonStyle struct {
	root          style.ResolvedStyle
	content       style.ResolvedStyle
	indicatorPart style.ResolvedStyle
	defaults      buttonDefaults
}

const stateSlotStyle = "style"

type buttonStyleKey struct {
	variant  Variant
	size     Size
	iconOnly bool
	group    buttonGroupItemStyle
}

type buttonDeclarationCache = computed.VisualCache[buttonStyleKey, buttonDefaults]

func (b Widget) declarations(active *theme.Theme, cache *buttonDeclarationCache) buttonDefaults {
	key := buttonStyleKey{variant: b.variant, size: b.size, iconOnly: b.iconOnly, group: b.group}
	return cache.Resolve(active, key, func() buttonDefaults { return b.defaultStyles(active) })
}

func (b Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State) buttonStyle {
	defaults := b.declarations(core.ActiveTheme(ctx), core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle))
	return b.resolveStyleWithDeclarations(ctx, gtx, key, state, defaults)
}

func (b Widget) resolveStyleWithDeclarations(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	defaults buttonDefaults,
) buttonStyle {
	return b.resolveStyleParts(ctx, gtx, key, state, defaults, b.loading)
}

func (b Widget) resolvePreparedStyleWithDeclarations(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	defaults buttonDefaults,
) buttonStyle {
	// Prepared content already contains the indicator's recorded drawing operations.
	return b.resolveStyleParts(ctx, gtx, key, state, defaults, false)
}

func (b Widget) resolveStyleParts(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	defaults buttonDefaults,
	includeIndicator bool,
) buttonStyle {
	parts, count := b.styleParts(includeIndicator)
	var resolved [3]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:count],
		resolved[:count],
		defaults.defaults,
		defaults.variant,
		defaults.size,
		b.style,
	)
	return buttonStyle{root: resolved[0], content: resolved[1], indicatorPart: resolved[2], defaults: defaults}
}

func (b Widget) staticStyle(ctx *core.Context, state style.State) buttonStyle {
	key := core.FullKey(ctx, b.key)
	defaults := b.declarations(core.ActiveTheme(ctx), core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle))
	return b.staticStyleParts(ctx, state, defaults, true)
}

func (b Widget) staticContentStyleWithDeclarations(
	ctx *core.Context,
	state style.State,
	defaults buttonDefaults,
) buttonStyle {
	return b.staticStyleParts(ctx, state, defaults, b.loading)
}

func (b Widget) staticStyleParts(
	ctx *core.Context,
	state style.State,
	defaults buttonDefaults,
	includeIndicator bool,
) buttonStyle {
	parts, count := b.styleParts(includeIndicator)
	var resolved [3]style.ResolvedStyle
	computed.ResolvePartsStaticInto(
		ctx,
		state,
		parts[:count],
		resolved[:count],
		defaults.defaults,
		defaults.variant,
		defaults.size,
		b.style,
	)
	return buttonStyle{root: resolved[0], content: resolved[1], indicatorPart: resolved[2], defaults: defaults}
}

func (b Widget) styleParts(includeIndicator bool) ([3]style.Part, int) {
	content := style.PartLabel
	if b.iconOnly {
		content = style.PartIcon
	}
	count := 2
	if includeIndicator {
		count++
	}
	return [3]style.Part{style.PartRoot, content, style.PartIndicator}, count
}

func buttonForeground(resolved style.ResolvedStyle) color.NRGBA {
	if resolved.Text == nil {
		panic("flowui/button: missing foreground text style")
	}
	foreground, ok := resolve.Color(resolved.Text.Color)
	if !ok {
		panic("flowui/button: unresolved foreground color")
	}
	return foreground
}

type closeStyles struct {
	root style.Style
	icon style.Style
}

type closeResolvedStyle struct {
	root style.ResolvedStyle
	icon style.ResolvedStyle
}

type closeStyleCache = computed.VisualCache[struct{}, closeStyles]

func (b CloseButtonWidget) styles(ctx *core.Context) closeStyles {
	active := core.ActiveTheme(ctx)
	return b.styleCache.Resolve(active, struct{}{}, func() closeStyles { return defaultCloseStyles(active) })
}

func (b CloseButtonWidget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	defaults closeStyles,
) closeResolvedStyle {
	return closeResolvedStyle{
		root: computed.Resolve(ctx, gtx, key, state, defaults.root, style.Style{}, style.Style{}, b.style),
		icon: computed.ResolvePart(
			ctx,
			gtx,
			key,
			style.PartIcon,
			state,
			defaults.icon,
			style.Style{},
			style.Style{},
			b.style,
		),
	}
}
