package piechart

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type pieChartVisualCache = computed.VisualCache[struct{}, pieChartDefaults]

func (w Widget) resolveDefaults(active *theme.Theme) pieChartDefaults {
	if w.visualCache == nil {
		return buildPieChart(active)
	}
	return w.visualCache.Resolve(active, struct{}{}, func() pieChartDefaults {
		return buildPieChart(active)
	})
}
