package listbox

import "github.com/qianniancn/flowui/internal/computed"

func listBoxDefaultsStats(cache *computed.VisualCache[struct{}, listBoxDefaults]) computed.VisualCacheStats {
	if cache == nil {
		return computed.VisualCacheStats{}
	}
	return cache.Stats()
}
