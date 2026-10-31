package toast

import "github.com/qianniancn/flowui/internal/computed"

func (c *toastVisualCache) stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}
