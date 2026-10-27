package text

import "github.com/qianniancn/flowui/internal/computed"

func (c *labelLabelVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
