package barchart

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// barChartMetrics contains geometry and typography consumed by bar-chart
// layout and drawing. It contains no theme or semantic objects.
type barChartMetrics struct {
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
	LegendTextSize      unit.Sp
	TextWeight          font.Weight
	EmphasisWeight      font.Weight
	LegendMarkerSize    unit.Dp
	LegendMarkerGap     unit.Dp
	LegendMarkerRadius  unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	BarRadius           unit.Dp
	MinBarHeight        unit.Dp
	BackgroundRadius    unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	TooltipTextSize     unit.Sp
}

var defaultBarChartMetrics = barChartMetrics{
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
	LegendTextSize:      12,
	TextWeight:          font.Normal,
	EmphasisWeight:      font.Medium,
	LegendMarkerSize:    10,
	LegendMarkerGap:     7,
	LegendMarkerRadius:  2,
	LegendItemGap:       20,
	LegendLineGap:       8,
	LegendGap:           12,
	GridWidth:           1,
	AxisWidth:           1,
	AnnotationLineWidth: 1,
	AnnotationPointSize: 9,
	BarRadius:           0,
	MinBarHeight:        0,
	BackgroundRadius:    0,
	TooltipGap:          12,
	TooltipRowGap:       6,
	TooltipMarkerSize:   8,
	TooltipTextSize:     12,
}

var defaultBarChartSeries = [9]color.NRGBA{
	{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}, {R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
	{R: 0x50, G: 0x53, B: 0x72, A: 0xff}, {R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
	{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff}, {R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
	{R: 0xfb, G: 0x62, B: 0x8b, A: 0xff}, {R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
	{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff},
}

// barChartColors contains all palette decisions used by bar-chart behavior.
// Explicit series colors remain widget inputs; the defaults own the fallback
// palette and annotation/interaction colors.
type barChartColors struct {
	Axis              color.NRGBA
	AxisLabel         color.NRGBA
	Grid              color.NRGBA
	CategoryHover     color.NRGBA
	BarBackground     color.NRGBA
	LegendHover       color.NRGBA
	AnnotationArea    color.NRGBA
	AnnotationLine    color.NRGBA
	AnnotationPoint   color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// SeriesColor resolves a default series color for behavior-owned data.
func (c barChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// barChartDefaults is the theme-derived appearance of a bar chart. Instance
// declarations are resolved separately as the last cascade layer, so they are
// not part of this value or its cache key.
type barChartDefaults struct {
	Defaults            style.Style
	Metrics             barChartMetrics
	Colors              barChartColors
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	Motion              theme.MotionTheme
}

// buildBarChart returns the theme-derived defaults for the active theme.
func buildBarChart(active *theme.Theme) barChartDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	return barChartDefaults{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: defaultBarChartMetrics,
		Colors: barChartColors{
			Axis:              semantic.Border,
			AxisLabel:         semantic.MutedForeground,
			Grid:              barWithAlpha(semantic.Border, 0.8),
			CategoryHover:     barWithAlpha(semantic.Tone(theme.TonePrimary).Fill, 18.0/255.0),
			BarBackground:     barWithAlpha(semantic.MutedForeground, 28.0/255.0),
			LegendHover:       semantic.SurfaceHover,
			AnnotationArea:    barWithAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:    barWithAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:   semantic.Tone(theme.ToneWarning).Fill,
			OverlayForeground: semantic.OverlayForeground,
			Series:            defaultBarChartSeries,
		},
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		Motion:              active.Motion,
	}
}

func barChartLabelForeground(background color.NRGBA) color.NRGBA {
	luminance := (299*int(background.R) + 587*int(background.G) + 114*int(background.B)) / 1000
	if luminance > 150 {
		return color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}

func barWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
