package navigation

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type paginationPaginationVisualKey struct {
	size theme.PaginationSize
}

type paginationPaginationVisualCache struct {
	cache computed.ProviderCache[paginationPaginationVisualKey, theme.PaginationVisual]
}

type PaginationVisualCacheStats = computed.ProviderCacheStats

func (c *paginationPaginationVisualCache) stats() PaginationVisualCacheStats {
	if c == nil {
		return PaginationVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (p PaginationWidget) visualWithCache(ctx *core.Context, override *theme.PaginationOverride) theme.PaginationVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	request := theme.PaginationRequest{Size: theme.PaginationSize(p.size), Override: override}
	if override != nil || p.visualCache == nil {
		return paginationResolvePaginationVisual(active, request)
	}
	key := paginationPaginationVisualKey{size: request.Size}
	return p.visualCache.cache.Resolve(active, key, func() theme.PaginationVisual {
		return paginationResolvePaginationVisual(active, request)
	})
}

func paginationResolvePaginationVisual(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Pagination, theme.ValidatePaginationVisual)
}
