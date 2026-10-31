package piechart

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// pieChartMetrics contains the geometry and typography consumed by pie-chart
// layout and drawing.
type pieChartMetrics struct {
	Height             unit.Dp
	PlotPaddingTop     unit.Dp
	PlotPaddingRight   unit.Dp
	PlotPaddingBottom  unit.Dp
	PlotPaddingLeft    unit.Dp
	LegendTextSize     unit.Sp
	LegendMarkerSize   unit.Dp
	LegendMarkerGap    unit.Dp
	LegendMarkerRadius unit.Dp
	LegendItemGap      unit.Dp
	LegendLineGap      unit.Dp
	LegendGap          unit.Dp
	LabelTextSize      unit.Sp
	LabelLineLength    unit.Dp
	LabelLineLength2   unit.Dp
	LabelLineWidth     unit.Dp
	LabelGap           unit.Dp
	EmphasisSize       unit.Dp
	TooltipGap         unit.Dp
	TooltipRowGap      unit.Dp
	TooltipMarkerSize  unit.Dp
	TooltipTextSize    unit.Sp
	TextWeight         font.Weight
	EmphasisWeight     font.Weight
}

var defaultPieChartMetrics = pieChartMetrics{
	Height:             360,
	PlotPaddingTop:     12,
	PlotPaddingRight:   16,
	PlotPaddingBottom:  12,
	PlotPaddingLeft:    16,
	LegendTextSize:     12,
	LegendMarkerSize:   10,
	LegendMarkerGap:    7,
	LegendMarkerRadius: 2,
	LegendItemGap:      20,
	LegendLineGap:      8,
	LegendGap:          12,
	LabelTextSize:      12,
	LabelLineLength:    15,
	LabelLineLength2:   30,
	LabelLineWidth:     1,
	LabelGap:           5,
	EmphasisSize:       5,
	TooltipGap:         12,
	TooltipRowGap:      6,
	TooltipMarkerSize:  8,
	TooltipTextSize:    12,
	TextWeight:         font.Normal,
	EmphasisWeight:     font.Medium,
}

var defaultPieChartSeries = [9]color.NRGBA{
	{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}, {R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
	{R: 0x50, G: 0x53, B: 0x72, A: 0xff}, {R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
	{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff}, {R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
	{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff}, {R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
	{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
}

// pieChartColors contains the palette decisions used by pie-chart drawing.
// Explicit slice colors remain widget inputs; the defaults own fallback colors.
type pieChartColors struct {
	Text              color.NRGBA
	MutedText         color.NRGBA
	Empty             color.NRGBA
	LegendHover       color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// SeriesColor resolves a default slice color for behavior-owned data.
func (c pieChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// pieChartDefaults is the theme-derived appearance of a pie chart. Instance
// declarations are resolved separately as the last cascade layer, so they are
// not part of this value or its cache key.
type pieChartDefaults struct {
	Defaults            style.Style
	Metrics             pieChartMetrics
	Colors              pieChartColors
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	Motion              theme.MotionTheme
}

// buildPieChart returns the theme-derived defaults for the active theme.
func buildPieChart(active *theme.Theme) pieChartDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	return pieChartDefaults{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: defaultPieChartMetrics,
		Colors: pieChartColors{
			Text:              semantic.Foreground,
			MutedText:         semantic.MutedForeground,
			Empty:             pieWithAlpha(semantic.SurfaceSecondary, .72),
			LegendHover:       semantic.SurfaceHover,
			OverlayForeground: semantic.OverlayForeground,
			Series:            defaultPieChartSeries,
		},
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		Motion:              active.Motion,
	}
}

func pieWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
