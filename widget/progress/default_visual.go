package progress

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// progressBarDefaults is the cached, theme-derived appearance and geometry of
// one linear progress variant. The instance style is not part of it: it is
// resolved as the last cascade layer on every layout.
type progressBarDefaults struct {
	defaults              style.Style
	variant               style.Style
	size                  style.Style
	headerGap             unit.Dp
	valueDuration         time.Duration
	indeterminatePeriod   time.Duration
	indeterminateFillRate float32
	motion                theme.MotionTheme
}

// progressCircleDefaults is the cached, theme-derived appearance and geometry
// of one circular progress variant.
type progressCircleDefaults struct {
	defaults      style.Style
	variant       style.Style
	size          style.Style
	strokeRatio   float32
	valueDuration time.Duration
	spinDuration  time.Duration
	motion        theme.MotionTheme
}

func (p BarWidget) defaultStyles(active *theme.Theme) progressBarDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return progressBarDefaults{
		defaults:              progressBarBaseStyle(active),
		variant:               progressBarVariantStyle(active, p.color),
		size:                  progressBarSizeStyle(active, p.size),
		headerGap:             4,
		valueDuration:         progressValueDuration(active.Motion),
		indeterminatePeriod:   1500 * time.Millisecond,
		indeterminateFillRate: .4,
		motion:                active.Motion,
	}
}

func (p CircleWidget) defaultStyles(active *theme.Theme) progressCircleDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return progressCircleDefaults{
		defaults:      style.Style{},
		variant:       progressCircleVariantStyle(active, p.color),
		size:          progressCircleSizeStyle(active, p.size),
		strokeRatio:   4.0 / 36.0,
		valueDuration: progressValueDuration(active.Motion),
		spinDuration:  time.Second,
		motion:        active.Motion,
	}
}

// progressColor keeps brush and opacity interpretation on the component side.
// Circle drawing needs a concrete color, while declarations remain free to use
// a solid fill or a gradient.
func progressColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Paint == nil {
		return color.NRGBA{}, false
	}
	brush, ok := computed.Brush(value.Paint.Background)
	if !ok {
		return color.NRGBA{}, false
	}
	return computed.ApplyOpacity(brush.ColorAt(.5), value.Paint.Opacity), true
}

func progressValueDuration(motion theme.MotionTheme) time.Duration {
	if motion.ControlDuration > 0 {
		return motion.ControlDuration
	}
	return motion.DefaultDuration
}

func progressBarBaseStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.
			FillWidth().
			Overflow(style.OverflowHidden).
			Background(style.Solid(sem.SurfaceRaised))).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(sem.Foreground)).
			FontSize(14).
			FontWeight(int(font.Medium))).
		Rules(style.Rule{
			When: style.Disabled,
			Set: style.Style{}.
				Part(style.PartLabel, style.Style{}.
					TextColor(style.Solid(active.DisabledColor(sem.Foreground)))),
		})
}

func progressBarVariantStyle(active *theme.Theme, value BarColor) style.Style {
	enabled := progressBarColors(active, value, false)
	disabled := progressBarColors(active, value, true)
	parts := func(colors progressBarPalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return parts(enabled).Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

func progressBarSizeStyle(active *theme.Theme, size BarSize) style.Style {
	height, radius := progressBarDimensions(size)
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.Height(height).Radius(radius)).
		Part(style.PartFill, style.Style{}.Radius(radius))
}

func progressCircleVariantStyle(active *theme.Theme, value BarColor) style.Style {
	enabled := progressCircleColors(active, value, false)
	disabled := progressCircleColors(active, value, true)
	parts := func(colors progressCirclePalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return parts(enabled).Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

func progressCircleSizeStyle(active *theme.Theme, size BarSize) style.Style {
	diameter := progressCircleDiameter(size)
	return style.Style{}.Width(diameter).Height(diameter).AspectRatio(1)
}

type progressBarPalette struct{ track, fill color.NRGBA }

func progressBarColors(active *theme.Theme, value BarColor, disabled bool) progressBarPalette {
	sem := active.Semantic()
	result := progressBarPalette{track: sem.SurfaceRaised, fill: progressTone(active, value).Fill}
	if value == BarDefault {
		result.fill = sem.Foreground
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

type progressCirclePalette struct{ track, fill color.NRGBA }

func progressCircleColors(active *theme.Theme, value BarColor, disabled bool) progressCirclePalette {
	neutral := active.Semantic().Tone(theme.ToneDefault)
	result := progressCirclePalette{track: neutral.Fill, fill: progressTone(active, value).Fill}
	if value == BarDefault {
		result.fill = neutral.OnFill
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

func progressTone(active *theme.Theme, value BarColor) theme.TonePair {
	switch value {
	case BarSuccess:
		return active.Semantic().Tone(theme.ToneSuccess)
	case BarWarning:
		return active.Semantic().Tone(theme.ToneWarning)
	case BarDanger:
		return active.Semantic().Tone(theme.ToneDanger)
	default:
		return active.Semantic().Tone(theme.TonePrimary)
	}
}

func progressBarDimensions(size BarSize) (unit.Dp, unit.Dp) {
	switch size {
	case BarSmall:
		return 4, 2
	case BarLarge:
		return 12, 6
	default:
		return 8, 4
	}
}

func progressCircleDiameter(size BarSize) unit.Dp {
	switch size {
	case BarSmall:
		return 20
	case BarLarge:
		return 36
	default:
		return 28
	}
}
