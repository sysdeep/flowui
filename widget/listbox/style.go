package listbox

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const listBoxResolvedPartCacheCapacity = 64

type listBoxResolvedPartKey struct {
	variant ItemVariant
	state   style.StateMask
}

// listBoxResolvedPartCache stores immutable static Item Part snapshots. The
// item transition key stays outside this cache, so two rows can share a
// settled declaration without sharing animation state.
type listBoxResolvedPartCache struct {
	cache           computed.VisualCacheMap[listBoxResolvedPartKey, style.ResolvedStyle]
	theme           *theme.Theme
	themeGeneration uint64
	scopeHash       uint64
	overrideHash    uint64
	enabled         bool
	initialized     bool
	hits            uint64
	misses          uint64
	themeInvalid    uint64
	scopeInvalid    uint64
	overrideInvalid uint64
}

type listBoxResolvedPartCacheStats struct {
	Hits                  uint64
	Misses                uint64
	ThemeInvalidations    uint64
	ScopeInvalidations    uint64
	OverrideInvalidations uint64
	Size                  int
	Evictions             uint64
}

func (c *listBoxResolvedPartCache) prepare(
	ctx *core.Context,
	active *theme.Theme,
	instanceStyle style.Style,
) {
	if c == nil {
		return
	}
	enabled := computed.StyleLayersCacheSafe(ctx, instanceStyle)
	scopeHash := uint64(0)
	overrideHash := uint64(0)
	if enabled {
		scopeHash = computed.StyleScopeHash(ctx)
		overrideHash = instanceStyle.Hash64()
	}
	generation := core.ThemeGeneration(ctx)
	changed := !c.initialized || c.theme != active || c.themeGeneration != generation ||
		c.scopeHash != scopeHash || c.overrideHash != overrideHash || c.enabled != enabled
	if !changed {
		return
	}
	if c.initialized {
		switch {
		case c.theme != active || c.themeGeneration != generation:
			c.themeInvalid++
		case c.scopeHash != scopeHash || c.enabled != enabled:
			c.scopeInvalid++
		case c.overrideHash != overrideHash:
			c.overrideInvalid++
		}
	}
	c.theme = active
	c.themeGeneration = generation
	c.scopeHash = scopeHash
	c.overrideHash = overrideHash
	c.enabled = enabled
	c.initialized = true
	c.cache = computed.VisualCacheMap[listBoxResolvedPartKey, style.ResolvedStyle]{}
	c.cache.SetCapacity(listBoxResolvedPartCacheCapacity)
}

func (c *listBoxResolvedPartCache) resolve(
	active *theme.Theme,
	key listBoxResolvedPartKey,
	load func() style.ResolvedStyle,
) style.ResolvedStyle {
	value, hit := c.cache.ResolveWithStatus(active, key, load)
	if hit {
		c.hits++
	} else {
		c.misses++
	}
	return value
}

func (c *listBoxResolvedPartCache) stats() listBoxResolvedPartCacheStats {
	if c == nil {
		return listBoxResolvedPartCacheStats{}
	}
	cache := c.cache.Stats()
	return listBoxResolvedPartCacheStats{
		Hits:                  c.hits,
		Misses:                c.misses,
		ThemeInvalidations:    c.themeInvalid,
		ScopeInvalidations:    c.scopeInvalid,
		OverrideInvalidations: c.overrideInvalid,
		Size:                  cache.Size,
		Evictions:             cache.Evictions,
	}
}

// defaultsFor resolves the theme-derived list appearance. The instance style
// is intentionally resolved separately by the layout path.
func (s *listBoxState) defaultsFor(ctx *core.Context) listBoxDefaults {
	active := core.ActiveTheme(ctx)
	return s.defaultsCache.Resolve(active, struct{}{}, func() listBoxDefaults {
		return buildListBoxDefaults(active)
	})
}

func (s *listBoxState) resolvedItemPart(
	ctx *core.Context,
	gtx layout.Context,
	active *theme.Theme,
	transitionKey string,
	variant ItemVariant,
	state style.State,
	defaults style.Style,
	instanceStyle style.Style,
) style.ResolvedStyle {
	if !s.resolvedPartCache.enabled {
		return computed.ResolvePartWithKey(
			ctx,
			gtx,
			transitionKey,
			style.PartItem,
			state,
			defaults,
			style.Style{},
			style.Style{},
			instanceStyle,
		)
	}
	key := listBoxResolvedPartKey{
		variant: variant,
		state:   state.Bits(),
	}
	static := s.resolvedPartCache.resolve(active, key, func() style.ResolvedStyle {
		return computed.ResolvePartStaticShared(
			ctx,
			style.PartItem,
			state,
			defaults,
			style.Style{},
			style.Style{},
			instanceStyle,
		)
	})
	return computed.ApplyTransitionsWithSharedKey(ctx, gtx, transitionKey, static)
}
