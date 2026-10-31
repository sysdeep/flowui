package nodegraph

import (
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// nodeGraphDefaults is the theme-derived appearance shared by every graph
// instance. Instance declarations are resolved separately and are never stored
// here, so the cache key only contains inputs that change the defaults.
type nodeGraphDefaults struct {
	root     style.Style
	metrics  nodeGraphMetrics
	colors   nodeGraphColors
	material *material.Theme
}

type nodeGraphSettingsCache = computed.VisualCache[struct{}, nodeGraphDefaults]

type nodeGraphDefaultsCache struct {
	cache          nodeGraphSettingsCache
	layoutHost     nodeGraphLayoutHost
	minimapScratch nodeGraphMinimapScratch
}

// nodeGraphDefaultsCacheStats exposes cache counters for diagnostics and
// benchmark assertions without exposing the cache implementation.
type nodeGraphDefaultsCacheStats = computed.VisualCacheStats

func (c *nodeGraphDefaultsCache) Stats() nodeGraphDefaultsCacheStats {
	if c == nil {
		return nodeGraphDefaultsCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveDefaults(active *theme.Theme) nodeGraphDefaults {
	if w.visualCache == nil {
		return defaultNodeGraphDefaults(active)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() nodeGraphDefaults {
		return defaultNodeGraphDefaults(active)
	})
}
