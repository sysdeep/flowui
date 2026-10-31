package popover

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type popoverDefaultsCache = computed.VisualCache[struct{}, popoverVisual]

// visualWithCache resolves the theme-derived defaults. The instance style is
// not part of the cache key because it is applied as the last cascade layer
// when the panel is resolved.
func (p Widget) visualWithCache(ctx *core.Context, cache *popoverDefaultsCache) popoverVisual {
	active := core.ActiveTheme(ctx)
	load := func() popoverVisual { return buildPopoverVisual(active) }
	return cache.Resolve(active, struct{}{}, load)
}
