package surface

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type cardCardVisualCache struct {
	cache computed.ProviderCache[CardVariant, theme.CardVisual]
}

func (c CardWidget) visualWithCache(ctx *core.Context, override *theme.CardOverride) theme.CardVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	request := theme.CardRequest{Variant: c.variant, Override: override}
	if override != nil || c.visualCache == nil {
		return cardResolveCardVisual(active, request)
	}
	return c.visualCache.cache.Resolve(active, c.variant, func() theme.CardVisual {
		return cardResolveCardVisual(active, request)
	})
}

func cardResolveCardVisual(active *theme.Theme, request theme.CardRequest) theme.CardVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Card, theme.ValidateCardVisual)
}
