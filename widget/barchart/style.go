package barchart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type barChartDefaultsCache = computed.VisualCache[struct{}, barChartDefaults]

func (w Widget) resolveDefaults(active *theme.Theme) barChartDefaults {
	if w.visualCache == nil {
		return buildBarChart(active)
	}
	return w.visualCache.Resolve(active, struct{}{}, func() barChartDefaults {
		return buildBarChart(active)
	})
}
