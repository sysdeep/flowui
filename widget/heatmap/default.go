package heatmap

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// heatmapMetrics is the geometry and typography consumed by heatmap layout.
type heatmapMetrics struct {
	Height               unit.Dp
	PlotPaddingTop       unit.Dp
	PlotPaddingRight     unit.Dp
	PlotPaddingBottom    unit.Dp
	PlotPaddingLeft      unit.Dp
	AxisTextSize         unit.Sp
	AxisFontWeight       font.Weight
	TickLabelGap         unit.Dp
	CellSize             unit.Dp
	CellGap              unit.Dp
	CellRadius           unit.Dp
	SelectionBorderWidth unit.Dp
	TooltipGap           unit.Dp
	TooltipTextSize      unit.Sp
	TooltipMarkerSize    unit.Dp
	TooltipRowGap        unit.Dp
	EmptyTextSize        unit.Sp
	EmptyFontWeight      font.Weight
}

var defaultHeatmapMetrics = heatmapMetrics{
	Height:               320,
	PlotPaddingTop:       28,
	PlotPaddingRight:     16,
	PlotPaddingBottom:    16,
	PlotPaddingLeft:      64,
	AxisTextSize:         12,
	AxisFontWeight:       font.Normal,
	TickLabelGap:         8,
	CellSize:             16,
	CellGap:              3,
	CellRadius:           2,
	SelectionBorderWidth: 2,
	TooltipGap:           12,
	TooltipTextSize:      12,
	TooltipMarkerSize:    7,
	TooltipRowGap:        4,
	EmptyTextSize:        12,
	EmptyFontWeight:      font.Normal,
}

// heatmapColors contains the colors consumed by heatmap drawing.
type heatmapColors struct {
	Min               color.NRGBA
	Max               color.NRGBA
	Empty             color.NRGBA
	AxisText          color.NRGBA
	Selection         color.NRGBA
	Muted             color.NRGBA
	OverlayForeground color.NRGBA
}

// heatmapVisual is the theme-derived appearance of one heatmap. Instance
// declarations are joined after it at resolve time, so they never enter the
// cached defaults.
type heatmapVisual struct {
	defaults style.Style
	metrics  heatmapMetrics
	colors   heatmapColors
}

// buildHeatmapVisual returns the complete visual plan for the active theme.
func buildHeatmapVisual(active *theme.Theme) heatmapVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	empty := semantic.SurfaceSecondary
	return heatmapVisual{
		defaults: style.Style{}.
			FillWidth().
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		metrics: defaultHeatmapMetrics,
		colors: heatmapColors{
			Min:               color.NRGBA{R: 0xe8, G: 0xf1, B: 0xff, A: 0xff},
			Max:               color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff},
			Empty:             empty,
			AxisText:          semantic.MutedForeground,
			Selection:         semantic.Foreground,
			Muted:             semantic.MutedForeground,
			OverlayForeground: semantic.OverlayForeground,
		},
	}
}

type heatmapColorScale func(minValue, maxValue, value float64) color.NRGBA

func heatLinearColorScale(minColor, maxColor color.NRGBA) heatmapColorScale {
	return func(minValue, maxValue, value float64) color.NRGBA {
		if maxValue <= minValue {
			return minColor
		}
		t := float32((value - minValue) / (maxValue - minValue))
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		return color.NRGBA{
			R: byte(float32(minColor.R) + (float32(maxColor.R)-float32(minColor.R))*t),
			G: byte(float32(minColor.G) + (float32(maxColor.G)-float32(minColor.G))*t),
			B: byte(float32(minColor.B) + (float32(maxColor.B)-float32(minColor.B))*t),
			A: byte(float32(minColor.A) + (float32(maxColor.A)-float32(minColor.A))*t),
		}
	}
}
