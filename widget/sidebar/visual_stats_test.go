package sidebar

import "github.com/qianniancn/flowui/internal/computed"

func (c *sidebarDefaultsCache) stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}
