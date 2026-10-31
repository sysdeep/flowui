package menubar

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type menubarDefaultsCache = computed.VisualCache[bool, menubarDefaults]

func (m Widget) defaultsWithCache(ctx *core.Context, cache *menubarDefaultsCache) menubarDefaults {
	active := core.ActiveTheme(ctx)
	key := m.compact
	load := func() menubarDefaults { return defaultMenubar(active, m.compact) }
	return cache.Resolve(active, key, load)
}

// resolveRoot applies the instance style as the final cascade layer of the bar.
func (m Widget) resolveRoot(ctx *core.Context, gtx layout.Context, key string, state style.State, defaults menubarDefaults) style.ResolvedStyle {
	return computed.Resolve(ctx, gtx, key, state, defaults.root, style.Style{}, style.Style{}, m.style)
}

// resolveTrigger applies the instance style as the final cascade layer of one
// trigger. Each trigger owns its transition identity.
func (m Widget) resolveTrigger(ctx *core.Context, gtx layout.Context, owner, itemKey string, state style.State, defaults menubarDefaults) style.ResolvedStyle {
	key := core.DerivedKey(ctx, owner, "trigger:"+itemKey)
	return computed.Resolve(ctx, gtx, key, state, defaults.trigger, style.Style{}, style.Style{}, m.style)
}
