package candlestick

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

// candlestickDefaultsCache memoizes the theme-derived defaults. The instance
// style is resolved separately as the last cascade layer, so it never enters
// the cache key.
type candlestickDefaultsCache = computed.VisualCache[struct{}, candlestickDefaults]

func (w Widget) defaultsFor(ctx *core.Context) candlestickDefaults {
	active := core.ActiveTheme(ctx)
	load := func() candlestickDefaults {
		return buildCandlestickDefaults(active)
	}
	if w.defaultsCache == nil {
		return load()
	}
	return w.defaultsCache.Resolve(active, struct{}{}, load)
}
