package gantt

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type ganttDefaultsCache = computed.VisualCache[struct{}, ganttDefaults]

func (w Widget) resolveVisual(active *theme.Theme) ganttDefaults {
	load := func() ganttDefaults {
		return buildGantt(active)
	}
	if w.visualCache == nil {
		return load()
	}
	return w.visualCache.Resolve(active, struct{}{}, load)
}
