package titlebar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type titleBarVisualKey struct {
	clientDecorations bool
}

type titleBarVisualCache = computed.VisualCache[titleBarVisualKey, titleBarVisual]

func (w Widget) visualWithCache(ctx *core.Context) titleBarVisual {
	active := core.ActiveTheme(ctx)
	key := titleBarVisualKey{clientDecorations: w.clientDecorations}
	load := func() titleBarVisual { return buildTitleBar(active) }
	if w.visualCache == nil {
		return load()
	}
	return w.visualCache.Resolve(active, key, load)
}
