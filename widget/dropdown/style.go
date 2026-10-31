package dropdown

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type dropdownDefaultsCache = computed.VisualCache[struct{}, dropdownDefaults]

// visual is the only production bridge from dropdown behavior to theme
// defaults. A layout pass resolves it once and passes the immutable result to
// the trigger and overlay paths.
func (d Widget) visual(ctx *core.Context) dropdownDefaults {
	return d.visualWithCache(ctx, nil)
}

func (d Widget) visualWithCache(ctx *core.Context, cache *dropdownDefaultsCache) dropdownDefaults {
	active := core.ActiveTheme(ctx)
	load := func() dropdownDefaults { return buildDropdown(active) }
	return cache.Resolve(active, struct{}{}, load)
}
