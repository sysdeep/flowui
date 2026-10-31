package alertdialog

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type alertDialogDefaultsCache = computed.VisualCache[Status, alertDialogDefaults]

// defaultsWithCache resolves the theme-derived header defaults for one status.
// The instance style is not part of the key because it is applied as the last
// cascade layer when the header is laid out.
func (a Widget) defaultsWithCache(ctx *core.Context, status Status) alertDialogDefaults {
	active := core.ActiveTheme(ctx)
	load := func() alertDialogDefaults {
		return buildAlertDialogDefaults(active, status)
	}
	return a.visualCache.Resolve(active, status, load)
}

func (h dialogHeader) defaults(ctx *core.Context) alertDialogDefaults {
	return Widget{status: h.status, visualCache: h.cache}.defaultsWithCache(ctx, h.status)
}
