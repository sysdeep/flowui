package modal

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type modalDefaultsKey struct {
	size     Size
	backdrop BackdropVariant
}

type modalDefaultsCache = computed.VisualCache[modalDefaultsKey, modalDefaults]

// defaultsWithCache resolves the theme-derived defaults for this size and
// backdrop. The instance style is not part of the key because it is applied as
// the last cascade layer when a part is resolved.
func (m Widget) defaultsWithCache(ctx *core.Context, cache *modalDefaultsCache) modalDefaults {
	active := core.ActiveTheme(ctx)
	key := modalDefaultsKey{size: m.size, backdrop: m.backdrop}
	load := func() modalDefaults {
		return buildModalDefaults(active, m.size, m.backdrop)
	}
	return cache.Resolve(active, key, load)
}
