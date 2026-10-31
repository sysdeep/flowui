package tree

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type treeVisualKey struct {
	Variant Variant
	Size    Size
}

type treeVisualCache struct {
	cache computed.VisualCacheMap[treeVisualKey, treeDefaults]
}

type treeVisualCacheStats = computed.VisualCacheStats

func (c *treeVisualCache) Stats() treeVisualCacheStats {
	if c == nil {
		return treeVisualCacheStats{}
	}
	return c.cache.Stats()
}

// resolve returns the theme-derived defaults for key. The instance style is
// resolved separately as the last cascade layer and is deliberately not part
// of the cache key.
func (c *treeVisualCache) resolve(active *theme.Theme, key treeVisualKey) treeDefaults {
	if c == nil {
		return buildTreeDefaults(active, key)
	}
	return c.cache.Resolve(active, key, func() treeDefaults {
		return buildTreeDefaults(active, key)
	})
}
