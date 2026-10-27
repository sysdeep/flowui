// Package progressstyle owns the visual recipes used by progress providers.
// Progress behavior remains in widget/progress: this package only maps theme
// semantics and component tokens to declarations and neutral layout metrics.
package feedbackstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildProgressBar builds the provider-owned linear progress visual plan.
func BuildProgressBar(active *theme.Theme, request theme.ProgressBarRequest) theme.ProgressBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ProgressBarVisual{
		Defaults: progressBarDefaults(active),
		Variant:  progressBarVariant(active, request.Color),
		Size:     progressBarSize(active, request.Size),
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    progressProgressColor,
		Metrics: theme.ProgressBarMetrics{
			HeaderGap:             active.Components.ProgressBar.HeaderGap,
			ValueDuration:         progressProgressValueDuration(active.Motion),
			IndeterminatePeriod:   1500 * time.Millisecond,
			IndeterminateFillRate: .4,
		},
		Motion: active.Motion,
	}
}

// BuildProgressCircle builds the provider-owned circular progress visual plan.
func BuildProgressCircle(active *theme.Theme, request theme.ProgressCircleRequest) theme.ProgressCircleVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ProgressCircleVisual{
		Defaults: progressCircleDefaults(active, request.Color, request.Size),
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    progressProgressColor,
		Renderer: progressProgressCircleRenderer,
		Metrics: theme.ProgressCircleMetrics{
			StrokeRatio:   active.Components.ProgressCircle.StrokeRatio,
			ValueDuration: progressProgressValueDuration(active.Motion),
			SpinDuration:  active.Components.ProgressCircle.SpinDuration,
		},
		Motion: active.Motion,
	}
}

// progressColor keeps brush and opacity interpretation on the provider side.
// Circle drawing needs a concrete color, while the widget remains agnostic to
// whether a theme uses a solid fill or a gradient declaration.
func progressProgressColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Paint == nil {
		return color.NRGBA{}, false
	}
	brush, ok := computed.Brush(value.Paint.Background)
	if !ok {
		return color.NRGBA{}, false
	}
	return computed.ApplyOpacity(brush.ColorAt(.5), value.Paint.Opacity), true
}

func progressProgressValueDuration(motion theme.MotionTheme) time.Duration {
	if motion.ControlDuration > 0 {
		return motion.ControlDuration
	}
	return motion.DefaultDuration
}

func progressBarDefaults(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.
			FillWidth().
			Overflow(style.OverflowHidden).
			Background(style.Solid(sem.SurfaceRaised))).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(sem.Foreground)).
			FontSize(active.Components.ProgressBar.TextSize).
			FontWeight(int(font.Medium))).
		Rules(style.Rule{
			When: style.Disabled,
			Set: style.Style{}.
				Part(style.PartLabel, style.Style{}.
					TextColor(style.Solid(active.DisabledColor(sem.Foreground)))),
		})
}

func progressBarVariant(active *theme.Theme, value theme.ProgressBarColor) style.Style {
	enabled := progressBarColors(active, value, false)
	disabled := progressBarColors(active, value, true)
	parts := func(colors progressBarPalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return parts(enabled).Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

func progressBarSize(active *theme.Theme, size theme.ProgressBarSize) style.Style {
	height, radius := progressBarDimensions(active, size)
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.Height(height).Radius(radius)).
		Part(style.PartFill, style.Style{}.Radius(radius))
}

func progressCircleDefaults(active *theme.Theme, value theme.ProgressBarColor, size theme.ProgressBarSize) style.Style {
	enabled := progressCircleColors(active, value, false)
	disabled := progressCircleColors(active, value, true)
	diameter := progressCircleDiameter(active, size)
	parts := func(colors progressCirclePalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return style.Join(style.Style{}.Width(diameter).Height(diameter).AspectRatio(1), parts(enabled)).
		Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

type progressBarPalette struct{ track, fill color.NRGBA }

func progressBarColors(active *theme.Theme, value theme.ProgressBarColor, disabled bool) progressBarPalette {
	sem := active.Semantic()
	result := progressBarPalette{track: sem.SurfaceRaised, fill: progressTone(active, value).Fill}
	if value == theme.ProgressBarDefault {
		result.fill = sem.Foreground
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

type progressCirclePalette struct{ track, fill color.NRGBA }

func progressCircleColors(active *theme.Theme, value theme.ProgressBarColor, disabled bool) progressCirclePalette {
	neutral := appearance.Tone(active, theme.ToneDefault)
	result := progressCirclePalette{track: neutral.Fill, fill: progressTone(active, value).Fill}
	if value == theme.ProgressBarDefault {
		result.fill = neutral.OnFill
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

func progressTone(active *theme.Theme, value theme.ProgressBarColor) appearance.ToneColors {
	switch value {
	case theme.ProgressBarSuccess:
		return appearance.Tone(active, theme.ToneSuccess)
	case theme.ProgressBarWarning:
		return appearance.Tone(active, theme.ToneWarning)
	case theme.ProgressBarDanger:
		return appearance.Tone(active, theme.ToneDanger)
	default:
		return appearance.Tone(active, theme.TonePrimary)
	}
}

func progressBarDimensions(active *theme.Theme, size theme.ProgressBarSize) (unit.Dp, unit.Dp) {
	tokens := active.Components.ProgressBar
	switch size {
	case theme.ProgressBarSmall:
		return tokens.SmallHeight, tokens.SmallRadius
	case theme.ProgressBarLarge:
		return tokens.LargeHeight, tokens.LargeRadius
	default:
		return tokens.MediumHeight, tokens.MediumRadius
	}
}

func progressCircleDiameter(active *theme.Theme, size theme.ProgressBarSize) unit.Dp {
	tokens := active.Components.ProgressCircle
	switch size {
	case theme.ProgressBarSmall:
		return tokens.SmallSize
	case theme.ProgressBarLarge:
		return tokens.LargeSize
	default:
		return tokens.MediumSize
	}
}
