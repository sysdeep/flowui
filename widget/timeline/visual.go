package timeline

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type timelineVisualCache struct {
	cache      computed.VisualCache[struct{}, timelineVisual]
	visualHost timelineVisualHost
}

type timelineVisualHost struct {
	value  Widget
	visual timelineVisual
}

func (v *timelineVisualHost) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return v.value.layout(ctx, gtx, v.visual)
}

func (c *timelineVisualCache) Stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}

func (w Widget) resolveVisual(active *theme.Theme) timelineVisual {
	if w.visualCache == nil {
		return buildTimeline(active)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() timelineVisual {
		return buildTimeline(active)
	})
}
