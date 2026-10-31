package spinner

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// spinnerMetrics is the geometry and base period of the indicator.
type spinnerMetrics struct {
	StrokeRatio  float32
	InsetRatio   float32
	SpinDuration time.Duration
}

func buildSpinnerDefaults(active *theme.Theme, spinnerColor Color, size Size) spinnerDefaults {
	diameter := spinnerDiameter(size)
	return spinnerDefaults{
		root: style.New().
			Width(diameter).
			Height(diameter).
			AspectRatio(1).
			Part(style.PartIndicator, style.New().
				TextColor(style.Solid(spinnerResolveColor(active, spinnerColor)))),
		metrics: spinnerMetrics{
			StrokeRatio:  .125,
			InsetRatio:   .0625,
			SpinDuration: 750 * time.Millisecond,
		},
		motion: active.Motion,
	}
}

func spinnerDraw(
	gtx layout.Context,
	size unit.Dp,
	strokeRatio, insetRatio float32,
	col color.NRGBA,
	period time.Duration,
) layout.Dimensions {
	diameter := max(gtx.Dp(size), 1)
	if diameter <= 0 || col.A == 0 {
		return layout.Dimensions{Size: image.Pt(diameter, diameter)}
	}
	geometry, ok := spinnerGeometry(diameter, strokeRatio, insetRatio)
	if !ok {
		return layout.Dimensions{Size: image.Pt(diameter, diameter)}
	}
	transform := op.Affine(f32.AffineId().Rotate(geometry.center, spinnerPhase(gtx.Now, period))).Push(gtx.Ops)
	for _, arc := range spinnerArcs {
		spinnerDrawArc(gtx, diameter, geometry, arc, col)
	}
	transform.Pop()
	return layout.Dimensions{Size: image.Pt(diameter, diameter)}
}

type spinnerArcGeometry struct {
	center      f32.Point
	radius      float32
	strokeWidth float32
}

type spinnerArcSpec struct {
	startAngle    float32
	sweep         float32
	gradientStart float32
	gradientEnd   float32
	startAlpha    float32
	endAlpha      float32
}

var spinnerArcs = [...]spinnerArcSpec{
	{startAngle: -99.61 * math.Pi / 180, sweep: -170.39 * math.Pi / 180, gradientStart: 2.745 / 24, gradientEnd: 20.79 / 24, startAlpha: 1, endAlpha: .55},
	{startAngle: -48.11 * math.Pi / 180, sweep: 138.11 * math.Pi / 180, gradientStart: 6.974 / 24, gradientEnd: 20.145 / 24, startAlpha: 0, endAlpha: .55},
}

func spinnerGeometry(diameter int, strokeRatio, insetRatio float32) (spinnerArcGeometry, bool) {
	if diameter <= 0 || math.IsNaN(float64(strokeRatio)) || math.IsInf(float64(strokeRatio), 0) ||
		math.IsNaN(float64(insetRatio)) || math.IsInf(float64(insetRatio), 0) ||
		strokeRatio <= 0 || strokeRatio >= .5 || insetRatio < 0 || insetRatio >= .5 {
		return spinnerArcGeometry{}, false
	}
	strokeWidth := max(float32(diameter)*strokeRatio, 1)
	outerRadius := float32(diameter) * (.5 - insetRatio)
	if outerRadius <= strokeWidth {
		return spinnerArcGeometry{}, false
	}
	return spinnerArcGeometry{center: f32.Pt(float32(diameter)/2, float32(diameter)/2), radius: outerRadius - strokeWidth/2, strokeWidth: strokeWidth}, true
}

func spinnerDrawArc(gtx layout.Context, diameter int, geometry spinnerArcGeometry, arc spinnerArcSpec, col color.NRGBA) {
	path := spinnerArcPath(gtx, geometry, arc)
	shape := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	draw.PaintLinearGradient(
		gtx,
		f32.Pt(geometry.center.X, float32(diameter)*arc.gradientStart),
		spinnerAlpha(col, arc.startAlpha),
		f32.Pt(geometry.center.X, float32(diameter)*arc.gradientEnd),
		spinnerAlpha(col, arc.endAlpha),
	)
	shape.Pop()
}

func spinnerArcPath(gtx layout.Context, geometry spinnerArcGeometry, arc spinnerArcSpec) clip.PathSpec {
	outerRadius := geometry.radius + geometry.strokeWidth/2
	innerRadius := geometry.radius - geometry.strokeWidth/2
	outerStart := spinnerPoint(geometry.center, outerRadius, arc.startAngle)
	outerEnd := spinnerPoint(geometry.center, outerRadius, arc.startAngle+arc.sweep)
	innerEnd := spinnerPoint(geometry.center, innerRadius, arc.startAngle+arc.sweep)
	capCenter := spinnerPoint(geometry.center, geometry.radius, arc.startAngle)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(outerStart)
	if arc.sweep < 0 {
		path.ArcTo(capCenter, capCenter, float32(math.Pi))
		path.ArcTo(geometry.center, geometry.center, arc.sweep)
		path.LineTo(outerEnd)
		path.ArcTo(geometry.center, geometry.center, -arc.sweep)
	} else {
		path.ArcTo(geometry.center, geometry.center, arc.sweep)
		path.LineTo(innerEnd)
		path.ArcTo(geometry.center, geometry.center, -arc.sweep)
		path.ArcTo(capCenter, capCenter, float32(math.Pi))
	}
	path.Close()
	return path.End()
}

func spinnerPoint(center f32.Point, radius, angle float32) f32.Point {
	return f32.Pt(center.X+float32(math.Cos(float64(angle)))*radius, center.Y+float32(math.Sin(float64(angle)))*radius)
}

func spinnerAlpha(col color.NRGBA, opacity float32) color.NRGBA {
	opacity = min(max(opacity, 0), 1)
	col.A = byte(float32(col.A)*opacity + .5)
	return col
}

func spinnerPhase(now time.Time, period time.Duration) float32 {
	if now.IsZero() || period <= 0 {
		return 0
	}
	elapsed := now.UnixNano() % int64(period)
	if elapsed < 0 {
		elapsed += int64(period)
	}
	return float32(elapsed) / float32(period) * float32(math.Pi*2)
}

func spinnerDiameter(size Size) unit.Dp {
	switch size {
	case Small:
		return 16
	case Large:
		return 32
	case ExtraLarge:
		return 40
	default:
		return 24
	}
}

func spinnerResolveColor(active *theme.Theme, value Color) color.NRGBA {
	semantic := active.Semantic()
	if value == Current {
		return semantic.Foreground
	}
	switch value {
	case Success:
		return semantic.Tone(theme.ToneSuccess).Fill
	case Warning:
		return semantic.Tone(theme.ToneWarning).Fill
	case Danger:
		return semantic.Tone(theme.ToneDanger).Fill
	default:
		return semantic.Tone(theme.TonePrimary).Fill
	}
}

func spinnerIndicatorColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Text == nil {
		return color.NRGBA{}, false
	}
	resolved, ok := computed.Color(value.Text.Color)
	if !ok {
		return color.NRGBA{}, false
	}
	if value.Paint != nil {
		resolved = computed.ApplyOpacity(resolved, value.Paint.Opacity)
	}
	return resolved, true
}

// Draw paints the default spinner indicator at size with the supplied color.
// Peer components that embed a spinner (for example a toast's loading
// indicator) use it instead of depending on spinner widget layout.
func Draw(gtx layout.Context, active *theme.Theme, size unit.Dp, col color.NRGBA) layout.Dimensions {
	period := theme.ResolveMotionDuration(active.Motion, 750*time.Millisecond)
	dims := spinnerDraw(gtx, size, .125, .0625, col, period)
	if period > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return dims
}
