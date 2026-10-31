package linechart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type lineChartDefaultsCache = computed.VisualCache[struct{}, lineChartDefaults]

func (w Widget) resolveDefaults(active *theme.Theme) lineChartDefaults {
	if w.visualCache == nil {
		return buildLineChart(active)
	}
	return w.visualCache.Resolve(active, struct{}{}, func() lineChartDefaults {
		return buildLineChart(active)
	})
}
