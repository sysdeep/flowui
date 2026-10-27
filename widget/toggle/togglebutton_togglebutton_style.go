package toggle

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type toggleToggleButtonResolvedStyle struct {
	root    style.ResolvedStyle
	content style.ResolvedStyle
}

// ToggleButtonVisualCache retains provider-resolved toggle visuals for repeated layouts.
// The foreground is part of the key because inherited scope can change it
// without replacing the active theme. Instance overrides deliberately bypass
// this cache because they are scoped to one widget instance.
type ToggleButtonVisualCache struct {
	cache computed.ProviderCacheMap[toggleToggleVisualCacheKey, theme.ToggleVisual]
}

type toggleToggleVisualCacheKey struct {
	variant    theme.ToggleButtonVariant
	size       theme.ToggleButtonSize
	iconOnly   bool
	foreground color.NRGBA
}

// ToggleButtonVisualCacheStats exposes provider-cache behavior for benchmarks and
// regression tests without exposing cache internals.
type ToggleButtonVisualCacheStats = computed.ProviderCacheStats

// Stats returns a snapshot of cache activity.
func (c *ToggleButtonVisualCache) Stats() ToggleButtonVisualCacheStats {
	if c == nil {
		return ToggleButtonVisualCacheStats{}
	}
	return c.cache.Stats()
}

// Resolve returns the cached visual for a provider request.
func (c *ToggleButtonVisualCache) Resolve(activeTheme *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	if c == nil || request.Override != nil {
		return toggleResolveToggleProvider(activeTheme, request)
	}
	key := toggleToggleVisualCacheKey{
		variant:    request.Variant,
		size:       request.Size,
		iconOnly:   request.IconOnly,
		foreground: request.Foreground,
	}
	return c.cache.Resolve(activeTheme, key, func() theme.ToggleVisual {
		return toggleResolveToggleProvider(activeTheme, request)
	})
}

func (b ToggleButtonWidget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, override *theme.ToggleOverride, caches ...*ToggleButtonVisualCache) toggleToggleButtonResolvedStyle {
	visual := toggleToggleVisualCached(core.ActiveTheme(ctx), b, ctx.ForegroundColor(), override, caches)
	part := style.PartLabel
	if b.iconOnly {
		part = style.PartIcon
	}
	return toggleToggleButtonResolvedStyle{
		root:    computed.Resolve(ctx, gtx, key, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
		content: computed.ResolvePart(ctx, gtx, key, part, state, visual.Defaults, visual.Variant, visual.Size, visual.Override),
	}
}

func toggleResolveToggleProvider(activeTheme *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	return theme.ResolveVisual(activeTheme, request,
		theme.ComponentVisualProvider.ToggleButton, theme.ValidateToggleVisual)
}

func toggleToggleVisualCached(activeTheme *theme.Theme, b ToggleButtonWidget, foreground color.NRGBA, override *theme.ToggleOverride, caches []*ToggleButtonVisualCache) theme.ToggleVisual {
	request := theme.ToggleRequest{
		Variant: theme.ToggleButtonVariant(b.variant), Size: theme.ToggleButtonSize(b.size),
		IconOnly: b.iconOnly, Foreground: foreground, Override: override,
	}
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, request)
	}
	return toggleResolveToggleProvider(activeTheme, request)
}
