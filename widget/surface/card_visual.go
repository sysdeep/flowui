package surface

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type cardVisual struct {
	Defaults, Variant style.Style
	Gap               unit.Dp
}

type cardVisualKey struct {
	variant    CardVariant
	paddingSet bool
	padding    unit.Dp
	gapSet     bool
	gap        unit.Dp
}

type cardVisualCache = computed.VisualCache[cardVisualKey, cardVisual]

func (c CardWidget) visualWithCache(ctx *core.Context) cardVisual {
	active := core.ActiveTheme(ctx)
	key := cardVisualKey{variant: c.variant}
	if c.padding != nil {
		key.paddingSet, key.padding = true, *c.padding
	}
	if c.gap != nil {
		key.gapSet, key.gap = true, *c.gap
	}
	load := func() cardVisual { return buildCardVisual(active, c.variant, c.padding, c.gap) }
	if c.visualCache == nil {
		return load()
	}
	return c.visualCache.Resolve(active, key, load)
}

func buildCardVisual(active *theme.Theme, variant CardVariant, paddingOverride, gapOverride *unit.Dp) cardVisual {
	sem := active.Semantic()
	background, foreground := sem.Surface, sem.SurfaceForeground
	switch variant {
	case CardSecondary:
		background, foreground = sem.SurfaceSecondary, sem.SurfaceSecondaryForeground
	case CardTertiary:
		background, foreground = sem.SurfaceTertiary, sem.SurfaceTertiaryForeground
	case CardTransparent:
		foreground = sem.Foreground
	}
	variantStyle := style.Style{}.TextColor(style.Solid(foreground))
	if variant != CardTransparent {
		variantStyle = variantStyle.Background(style.Solid(background))
	}
	padding, gap := unit.Dp(16), unit.Dp(16)
	if paddingOverride != nil {
		padding = max(*paddingOverride, 0)
	}
	if gapOverride != nil {
		gap = max(*gapOverride, 0)
	}
	defaults := style.Style{}.
		BorderColor(style.Solid(sem.Separator)).
		BorderWidth(1).
		Opacity(1).
		Radius(8).
		Padding(padding)
	if variant != CardTransparent {
		defaults = defaults.Shadow(style.ShadowSurface)
	}
	return cardVisual{Defaults: defaults, Variant: variantStyle, Gap: gap}
}
