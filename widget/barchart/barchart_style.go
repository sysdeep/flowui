package barchart

import (
	"image/color"
)

type chartStyle struct {
	axis          color.NRGBA
	axisLabel     color.NRGBA
	grid          color.NRGBA
	categoryHover color.NRGBA
	barBackground color.NRGBA
	legendHover   color.NRGBA
	opacity       float32
}

func barChartStyleFor(defaults barChartDefaults, disabled bool) chartStyle {
	opacity := float32(1)
	if disabled {
		opacity = defaults.DisabledOpacity
	}
	return chartStyle{
		axis:          defaults.Colors.Axis,
		axisLabel:     defaults.Colors.AxisLabel,
		grid:          defaults.Colors.Grid,
		categoryHover: defaults.Colors.CategoryHover,
		barBackground: defaults.Colors.BarBackground,
		legendHover:   defaults.Colors.LegendHover,
		opacity:       opacity,
	}
}
