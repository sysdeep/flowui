package linechart

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// lineChartMetrics contains geometry and typography consumed by line-chart
// layout and drawing. It contains no theme or semantic objects.
type lineChartMetrics struct {
	AnimationDuration   time.Duration
	UpdateDuration      time.Duration
	Height              unit.Dp
	PlotPaddingTop      unit.Dp
	PlotPaddingRight    unit.Dp
	PlotPaddingBottom   unit.Dp
	PlotPaddingLeft     unit.Dp
	AxisNameGap         unit.Dp
	TickLabelGap        unit.Dp
	AxisTextSize        unit.Sp
	TextWeight          font.Weight
	EmphasisWeight      font.Weight
	LegendTextSize      unit.Sp
	LegendMarkerWidth   unit.Dp
	LegendMarkerSize    unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	LineWidth           unit.Dp
	PointSize           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	HoverPointSize      unit.Dp
	CrosshairWidth      unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	TooltipTextSize     unit.Sp
}

var defaultLineChartMetrics = lineChartMetrics{
	AnimationDuration:   1 * time.Second,
	UpdateDuration:      500 * time.Millisecond,
	Height:              320,
	PlotPaddingTop:      12,
	PlotPaddingRight:    16,
	PlotPaddingBottom:   36,
	PlotPaddingLeft:     56,
	AxisNameGap:         6,
	TickLabelGap:        8,
	AxisTextSize:        12,
	TextWeight:          font.Normal,
	EmphasisWeight:      font.Medium,
	LegendTextSize:      12,
	LegendMarkerWidth:   20,
	LegendMarkerSize:    6,
	LegendItemGap:       20,
	LegendLineGap:       8,
	LegendGap:           12,
	GridWidth:           1,
	AxisWidth:           1,
	LineWidth:           2,
	PointSize:           6,
	AnnotationLineWidth: 1,
	AnnotationPointSize: 9,
	HoverPointSize:      10,
	CrosshairWidth:      1,
	TooltipGap:          12,
	TooltipRowGap:       6,
	TooltipMarkerSize:   7,
	TooltipTextSize:     12,
}

var defaultLineChartSeries = [9]color.NRGBA{
	{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}, {R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
	{R: 0x50, G: 0x53, B: 0x72, A: 0xff}, {R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
	{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff}, {R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
	{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff}, {R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
	{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
}

// lineChartColors contains all palette decisions used by line-chart behavior.
// Explicit series colors remain widget inputs; the defaults own the fallback
// palette and annotation/interaction colors.
type lineChartColors struct {
	Axis              color.NRGBA
	AxisLabel         color.NRGBA
	Grid              color.NRGBA
	Crosshair         color.NRGBA
	Background        color.NRGBA
	LegendHover       color.NRGBA
	AnnotationArea    color.NRGBA
	AnnotationLine    color.NRGBA
	AnnotationPoint   color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// SeriesColor resolves a default series color for behavior-owned data.
func (c lineChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// lineChartDefaults is the theme-derived appearance of a line chart. Instance
// declarations are resolved separately as the last cascade layer, so they are
// not part of this value or its cache key.
type lineChartDefaults struct {
	Defaults            style.Style
	Metrics             lineChartMetrics
	Colors              lineChartColors
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	AreaOpacity         float32
	Motion              theme.MotionTheme
}

// buildLineChart returns the theme-derived defaults for the active theme.
func buildLineChart(active *theme.Theme) lineChartDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	grid := lineWithAlpha(semantic.Border, 0.8)
	crosshair := lineWithAlpha(semantic.MutedForeground, 0.75)
	return lineChartDefaults{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: defaultLineChartMetrics,
		Colors: lineChartColors{
			Axis:              semantic.Border,
			AxisLabel:         semantic.MutedForeground,
			Grid:              grid,
			Crosshair:         crosshair,
			Background:        semantic.Surface,
			LegendHover:       semantic.SurfaceHover,
			AnnotationArea:    lineWithAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:    lineWithAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:   semantic.Tone(theme.ToneWarning).Fill,
			OverlayForeground: semantic.OverlayForeground,
			Series:            defaultLineChartSeries,
		},
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		AreaOpacity:         0x30 / 255.0,
		Motion:              active.Motion,
	}
}

func lineWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
