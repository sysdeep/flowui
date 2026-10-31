package gantt

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	ganttTaskBackgroundOpacity = 0.24
	ganttBaselineOpacity       = 0xa0 / 255.0
	ganttSelectedOpacity       = 0.42
	ganttLegendHiddenOpacity   = 0.38
	ganttEmphasisWidth         = 2
	ganttDependencyArrowLength = 5
	ganttDependencyArrowHeight = 3
	ganttDependencyRouteMargin = 4
)

// ganttMetrics contains geometry and typography consumed by Gantt layout and
// drawing.
type ganttMetrics struct {
	AnimationDuration     time.Duration
	UpdateDuration        time.Duration
	Height                unit.Dp
	PlotPaddingTop        unit.Dp
	PlotPaddingRight      unit.Dp
	PlotPaddingBottom     unit.Dp
	PlotPaddingLeft       unit.Dp
	AxisNameGap           unit.Dp
	TickLabelGap          unit.Dp
	PlotLabelGap          unit.Dp
	AxisTextSize          unit.Sp
	TextWeight            font.Weight
	EditEdge              unit.Dp
	GridWidth             unit.Dp
	AxisWidth             unit.Dp
	RowHeight             unit.Dp
	BarHeight             unit.Dp
	BarRadius             unit.Dp
	TaskIndent            unit.Dp
	TaskToggleSize        unit.Dp
	TaskToggleGap         unit.Dp
	TaskLabelPaddingX     unit.Dp
	DependencyWidth       unit.Dp
	DependencyDash        unit.Dp
	MarkerWidth           unit.Dp
	MarkerLabelGap        unit.Dp
	LegendGap             unit.Dp
	LegendTextSize        unit.Sp
	LegendMarkerSize      unit.Dp
	LegendMarkerGap       unit.Dp
	LegendRadius          unit.Dp
	LegendItemGap         unit.Dp
	LegendLineGap         unit.Dp
	TooltipGap            unit.Dp
	TooltipRowGap         unit.Dp
	TooltipMarkerSize     unit.Dp
	TooltipTextSize       unit.Sp
	EmphasisWidth         unit.Dp
	DependencyArrowLength unit.Dp
	DependencyArrowHeight unit.Dp
	DependencyRouteMargin unit.Dp
}

var defaultGanttMetrics = ganttMetrics{
	AnimationDuration:     700 * time.Millisecond,
	UpdateDuration:        450 * time.Millisecond,
	Height:                360,
	PlotPaddingTop:        28,
	PlotPaddingRight:      20,
	PlotPaddingBottom:     34,
	PlotPaddingLeft:       140,
	AxisNameGap:           6,
	TickLabelGap:          8,
	PlotLabelGap:          6,
	AxisTextSize:          12,
	TextWeight:            font.Normal,
	EditEdge:              6,
	GridWidth:             1,
	AxisWidth:             1,
	RowHeight:             36,
	BarHeight:             18,
	BarRadius:             3,
	TaskIndent:            14,
	TaskToggleSize:        12,
	TaskToggleGap:         4,
	TaskLabelPaddingX:     6,
	DependencyWidth:       1,
	DependencyDash:        4,
	MarkerWidth:           1,
	MarkerLabelGap:        4,
	LegendGap:             10,
	LegendTextSize:        12,
	LegendMarkerSize:      10,
	LegendMarkerGap:       5,
	LegendRadius:          4,
	LegendItemGap:         14,
	LegendLineGap:         6,
	TooltipGap:            12,
	TooltipRowGap:         4,
	TooltipMarkerSize:     7,
	TooltipTextSize:       12,
	EmphasisWidth:         ganttEmphasisWidth,
	DependencyArrowLength: ganttDependencyArrowLength,
	DependencyArrowHeight: ganttDependencyArrowHeight,
	DependencyRouteMargin: ganttDependencyRouteMargin,
}

var defaultGanttSeries = [9]color.NRGBA{
	{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}, {R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
	{R: 0x3f, G: 0xbe, B: 0x95, A: 0xff}, {R: 0xff, G: 0x99, B: 0x4d, A: 0xff},
	{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff}, {R: 0xfb, G: 0x62, B: 0x8b, A: 0xff},
	{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff}, {R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
	{R: 0x50, G: 0x53, B: 0x72, A: 0xff},
}

// ganttColors contains the palette used by Gantt drawing.
type ganttColors struct {
	Axis              color.NRGBA
	Label             color.NRGBA
	Grid              color.NRGBA
	Dependency        color.NRGBA
	Marker            color.NRGBA
	Emphasis          color.NRGBA
	TaskLabel         color.NRGBA
	Empty             color.NRGBA
	LegendHover       color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

func (c ganttColors) seriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

func (c ganttColors) seriesSignature() uint64 {
	hash := uint64(1469598103934665603)
	for _, col := range c.Series {
		hash = (hash ^ uint64(col.R)) * 1099511628211
		hash = (hash ^ uint64(col.G)) * 1099511628211
		hash = (hash ^ uint64(col.B)) * 1099511628211
		hash = (hash ^ uint64(col.A)) * 1099511628211
	}
	return hash
}

// ganttDefaults is the theme-derived appearance shared by Gantt instances.
type ganttDefaults struct {
	Defaults            style.Style
	Metrics             ganttMetrics
	Colors              ganttColors
	LegendHiddenOpacity float32
}

func buildGantt(active *theme.Theme) ganttDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	return ganttDefaults{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: defaultGanttMetrics,
		Colors: ganttColors{
			Axis:              semantic.Border,
			Label:             semantic.MutedForeground,
			Grid:              ganttWithAlpha(semantic.Border, 0.75),
			Dependency:        ganttWithAlpha(semantic.MutedForeground, 0.7),
			Marker:            semantic.Tone(theme.TonePrimary).Fill,
			Emphasis:          semantic.Foreground,
			TaskLabel:         semantic.OverlayForeground,
			Empty:             semantic.MutedForeground,
			LegendHover:       semantic.SurfaceHover,
			OverlayForeground: semantic.OverlayForeground,
			Series:            defaultGanttSeries,
		},
		LegendHiddenOpacity: ganttLegendHiddenOpacity,
	}
}

func ganttDrawTask(
	gtx layout.Context,
	rect image.Rectangle,
	progressRect image.Rectangle,
	baseline image.Rectangle,
	value color.NRGBA,
	emphasis color.NRGBA,
	hasSelection bool,
	selected bool,
	milestone bool,
	radius int,
	emphasisWidth float32,
) {
	if !baseline.Empty() {
		background := ganttWithAlpha(value, ganttBaselineOpacity)
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(baseline, min(baseline.Dy()/2, 2)).Op(gtx.Ops))
	}
	if rect.Empty() {
		return
	}
	opacity := float32(1)
	if hasSelection && !selected {
		opacity = ganttSelectedOpacity
	}
	fade := paint.PushOpacity(gtx.Ops, opacity)
	if milestone {
		ganttDrawMilestone(gtx, rect, value, emphasis, selected, emphasisWidth)
		fade.Pop()
		return
	}
	background := ganttWithAlpha(value, ganttTaskBackgroundOpacity)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	if !progressRect.Empty() {
		progressRadius := min(radius, min(progressRect.Dx(), progressRect.Dy())/2)
		paint.FillShape(gtx.Ops, value, clip.UniformRRect(progressRect, progressRadius).Op(gtx.Ops))
	}
	fade.Pop()
}

func ganttDrawMilestone(
	gtx layout.Context,
	rect image.Rectangle,
	value color.NRGBA,
	emphasis color.NRGBA,
	selected bool,
	emphasisWidth float32,
) {
	center := f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Min.Y+rect.Dy()/2))
	radius := float32(min(rect.Dx(), rect.Dy())) / 2
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X, center.Y-radius))
	path.LineTo(f32.Pt(center.X+radius, center.Y))
	path.LineTo(f32.Pt(center.X, center.Y+radius))
	path.LineTo(f32.Pt(center.X-radius, center.Y))
	path.Close()
	spec := path.End()
	paint.FillShape(gtx.Ops, value, clip.Outline{Path: spec}.Op())
	if selected {
		paint.FillShape(gtx.Ops, emphasis, clip.Stroke{Path: spec, Width: emphasisWidth}.Op())
	}
}

func ganttDrawDependency(
	gtx layout.Context,
	from f32.Point,
	to f32.Point,
	middle f32.Point,
	forward bool,
	width float32,
	dash float32,
	arrowLength float32,
	arrowHeight float32,
	value color.NRGBA,
) {
	ganttStrokeDependency(gtx, from, middle, width, value, dash)
	ganttStrokeDependency(gtx, middle, f32.Pt(middle.X, to.Y), width, value, dash)
	ganttStrokeDependency(gtx, f32.Pt(middle.X, to.Y), to, width, value, dash)
	direction := float32(1)
	if !forward {
		direction = -1
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(to)
	path.LineTo(f32.Pt(to.X-arrowLength*direction, to.Y-arrowHeight))
	path.LineTo(f32.Pt(to.X-arrowLength*direction, to.Y+arrowHeight))
	path.Close()
	paint.FillShape(gtx.Ops, value, clip.Outline{Path: path.End()}.Op())
}

func ganttStrokeDependency(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA, dash float32) {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if distance <= 0 || dash <= 0 || width <= 0 || col.A == 0 {
		return
	}
	for offset := float32(0); offset < distance; offset += dash * 2 {
		end := min(offset+dash, distance)
		ganttStrokeLine(gtx,
			f32.Pt(from.X+dx*offset/distance, from.Y+dy*offset/distance),
			f32.Pt(from.X+dx*end/distance, from.Y+dy*end/distance), width, col)
	}
}

func ganttStrokeLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	if width <= 0 || col.A == 0 {
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func ganttDrawToggle(gtx layout.Context, area image.Rectangle, collapsed bool, value color.NRGBA) {
	center := f32.Pt(float32(area.Min.X+area.Dx()/2), float32(area.Min.Y+area.Dy()/2))
	size := float32(min(max(area.Dx()/2, 3), max(area.Dy()/3, 3)))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	if collapsed {
		path.MoveTo(f32.Pt(center.X-size/2, center.Y-size))
		path.LineTo(f32.Pt(center.X-size/2, center.Y+size))
		path.LineTo(f32.Pt(center.X+size, center.Y))
	} else {
		path.MoveTo(f32.Pt(center.X-size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X+size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X, center.Y+size))
	}
	path.Close()
	paint.FillShape(gtx.Ops, value, clip.Outline{Path: path.End()}.Op())
}

func ganttWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
