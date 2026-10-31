package linechart

import (
	"image/color"
)

type chartStyle struct {
	axis        color.NRGBA
	axisLabel   color.NRGBA
	grid        color.NRGBA
	crosshair   color.NRGBA
	background  color.NRGBA
	legendHover color.NRGBA
	opacity     float32
}

func lineChartStyleFor(defaults lineChartDefaults, disabled bool) chartStyle {
	opacity := float32(1)
	if disabled {
		opacity = defaults.DisabledOpacity
	}
	return chartStyle{
		axis:        defaults.Colors.Axis,
		axisLabel:   defaults.Colors.AxisLabel,
		grid:        defaults.Colors.Grid,
		crosshair:   defaults.Colors.Crosshair,
		background:  defaults.Colors.Background,
		legendHover: defaults.Colors.LegendHover,
		opacity:     opacity,
	}
}
