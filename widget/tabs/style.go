package tabs

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

// tabsDefaultsKey holds only the inputs that change the theme-derived
// defaults. The instance style resolves separately as the last cascade layer.
type tabsDefaultsKey struct {
	variant Variant
	size    Size
	color   Color
}

type tabsDefaultsCache = computed.VisualCache[tabsDefaultsKey, tabsDefaults]

func (t Widget) defaultsWithCache(ctx *core.Context) tabsDefaults {
	active := core.ActiveTheme(ctx)
	key := tabsDefaultsKey{variant: t.variant, size: t.size, color: t.color}
	load := func() tabsDefaults { return buildTabsDefaults(active, key) }
	if t.visualCache == nil {
		return load()
	}
	return t.visualCache.Resolve(active, key, load)
}
