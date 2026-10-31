package toolbar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
)

type toolbarVisualKey struct {
	orientation toolbarAxis
	attached    bool
}

type toolbarVisualCache = computed.VisualCache[toolbarVisualKey, toolbarVisual]

func (w Widget) visualWithCache(ctx *core.Context) toolbarVisual {
	active := core.ActiveTheme(ctx)
	key := toolbarVisualKey{
		orientation: toolbarOrientation(w.orientation),
		attached:    w.attached,
	}
	load := func() toolbarVisual { return buildToolbar(active, w.attached) }
	if w.visualCache == nil {
		return load()
	}
	return w.visualCache.Resolve(active, key, load)
}

func toolbarOrientation(value Orientation) toolbarAxis {
	if value == Vertical {
		return toolbarVertical
	}
	return toolbarHorizontal
}
