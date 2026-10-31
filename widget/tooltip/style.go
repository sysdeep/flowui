package tooltip

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type tooltipDefaultsCache = computed.VisualCache[struct{}, tooltipVisual]

func tooltipVisualWithCache(ctx *core.Context, cache *tooltipDefaultsCache) tooltipVisual {
	active := core.ActiveTheme(ctx)
	load := func() tooltipVisual { return buildTooltipVisual(active) }
	return cache.Resolve(active, struct{}{}, load)
}

func (p Popup) visualWithCache(ctx *core.Context) tooltipVisual {
	return tooltipVisualWithCache(ctx, p.visualCache)
}
