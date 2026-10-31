package toggle

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type toggleButtonResolvedStyle struct {
	root    style.ResolvedStyle
	content style.ResolvedStyle
}

// toggleButtonVisualKey identifies the theme- and context-derived inputs that
// change a toggle button's defaults. The foreground is part of the key because
// inherited scope can change it without replacing the active theme. Instance
// declarations are deliberately absent: they are resolved as the last cascade
// layer for one widget instance.
type toggleButtonVisualKey struct {
	variant    ToggleButtonVariant
	size       ToggleButtonSize
	iconOnly   bool
	foreground color.NRGBA
}

type toggleButtonVisualCache struct {
	cache computed.VisualCacheMap[toggleButtonVisualKey, toggleButtonDefaults]
}

func (c *toggleButtonVisualCache) resolve(active *theme.Theme, key toggleButtonVisualKey) toggleButtonDefaults {
	load := func() toggleButtonDefaults {
		return buildToggleButtonDefaults(active, key.variant, key.size, key.iconOnly, key.foreground)
	}
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}

func (c *toggleButtonVisualCache) stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}

func (b ToggleButtonWidget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	cache *toggleButtonVisualCache,
) toggleButtonResolvedStyle {
	defaults := cache.resolve(core.ActiveTheme(ctx), toggleButtonVisualKey{
		variant: b.variant, size: b.size, iconOnly: b.iconOnly,
		foreground: ctx.ForegroundColor(),
	})
	part := style.PartLabel
	if b.iconOnly {
		part = style.PartIcon
	}
	parts := [2]style.Part{style.PartRoot, part}
	var resolved [2]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:],
		resolved[:],
		defaults.Defaults,
		defaults.Variant,
		defaults.Size,
		b.style,
	)
	return toggleButtonResolvedStyle{root: resolved[0], content: resolved[1]}
}
