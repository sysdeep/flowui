package slider

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// sliderVisual is the component-owned appearance of one slider orientation.
type sliderVisual struct {
	Defaults         style.Style
	Metrics          sliderMetrics
	Motion           theme.MotionTheme
	FocusDuration    time.Duration
	DraggingDuration time.Duration
}

// sliderMetrics contains component geometry consumed by slider layout.
type sliderMetrics struct {
	TrackThickness  unit.Dp
	TrackRadius     unit.Dp
	EdgeInset       unit.Dp
	ThumbLength     unit.Dp
	ThumbExtra      unit.Dp
	HeaderGap       unit.Dp
	VerticalGap     unit.Dp
	DraggingScale   float32
	FocusRingWidth  unit.Dp
	FocusRingOffset unit.Dp
}

// buildSlider builds the component-owned slider appearance and metrics.
func buildSlider(active *theme.Theme, orientation Orientation) sliderVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := sliderMetrics{
		TrackThickness: 8, TrackRadius: 4, EdgeInset: 8,
		ThumbLength: 16, ThumbExtra: 0, HeaderGap: 4,
		VerticalGap: 8, DraggingScale: 1,
		FocusRingWidth: 2, FocusRingOffset: 2,
	}
	focusDuration := 100 * time.Millisecond
	draggingDuration := 150 * time.Millisecond
	if active.Motion.ControlDuration > 0 {
		focusDuration = active.Motion.ControlDuration
		draggingDuration = active.Motion.ControlDuration
	}
	return sliderVisual{
		Defaults:         sliderDefaultStyle(active, orientation),
		Metrics:          tokens,
		Motion:           active.Motion,
		FocusDuration:    focusDuration,
		DraggingDuration: draggingDuration,
	}
}

// sliderDefaultStyle lowers the slider parts to one style declaration. Every
// part keeps its own declaration so instance Style layers can target it.
func sliderDefaultStyle(active *theme.Theme, orientation Orientation) style.Style {
	tokens := sliderMetrics{
		TrackThickness: 8, TrackRadius: 4, ThumbLength: 16,
		ThumbExtra: 0, FocusRingWidth: 2, FocusRingOffset: 2,
	}
	enabled := sliderColors(active, false)
	disabled := sliderColors(active, true)

	thumbWidth := tokens.ThumbLength + tokens.ThumbExtra
	thumbHeight := tokens.TrackThickness
	thumbRadius := tokens.TrackRadius
	thumbWidth, thumbHeight = 16, 16
	thumbRadius = 8
	borderWidth := max(tokens.ThumbExtra/2, unit.Dp(1))

	track := style.Style{}.
		Background(style.Solid(enabled.Track)).
		Radius(tokens.TrackRadius).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.Background(style.Solid(disabled.Track)),
		})
	fill := style.Style{}.
		Background(style.Solid(enabled.Fill)).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.Background(style.Solid(disabled.Fill)),
		})
	if orientation == Vertical {
		fill = fill.RadiusBottomLeft(tokens.TrackRadius).RadiusBottomRight(tokens.TrackRadius)
		track = track.Width(tokens.TrackThickness).FillHeight()
		thumbWidth, thumbHeight = thumbHeight, thumbWidth
	} else {
		fill = fill.RadiusTopLeft(tokens.TrackRadius).RadiusBottomLeft(tokens.TrackRadius)
		track = track.FillWidth().Height(tokens.TrackThickness)
	}

	thumb := style.Style{}.
		Width(thumbWidth).
		Height(thumbHeight).
		Background(style.Solid(enabled.ThumbBackground)).
		BorderWidth(borderWidth).
		BorderColor(style.Solid(enabled.ThumbBorder)).
		Radius(thumbRadius).
		BoxShadow(0, 2, 0, 2, style.Solid(color.NRGBA{A: sliderShadowAlpha(0x0a, enabled.ThumbBackground)})).
		BoxShadow(0, 1, 0, 1, style.Solid(color.NRGBA{A: sliderShadowAlpha(0x0f, enabled.ThumbBackground)})).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(enabled.Focus)).
		Rules(style.Rule{
			When: style.Disabled,
			Set: style.Style{}.
				Background(style.Solid(disabled.ThumbBackground)).
				BorderColor(style.Solid(disabled.ThumbBorder)).
				BoxShadowNone().
				BoxShadow(0, 2, 0, 2, style.Solid(color.NRGBA{A: sliderShadowAlpha(0x0a, disabled.ThumbBackground)})).
				BoxShadow(0, 1, 0, 1, style.Solid(color.NRGBA{A: sliderShadowAlpha(0x0f, disabled.ThumbBackground)})).
				Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(disabled.Focus)),
		})

	sem := active.Semantic()
	label := style.Style{}.
		TextColor(style.Solid(sem.Foreground)).
		FontSize(14).
		FontWeight(int(font.Medium)).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.TextColor(style.Solid(active.DisabledColor(sem.Foreground))),
		})

	margin := max((unit.Dp(16)-tokens.TrackThickness)/2, 0)
	trackMargin := style.Style{}.MarginY(margin)
	if orientation == Vertical {
		trackMargin = style.Style{}.MarginX(margin)
	}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	thumbOverlay := style.Style{}.
		Background(style.Solid(white)).
		BorderColor(style.Solid(sem.Border)).
		BoxShadowNone().
		BoxShadow(0, 1, 3, 0, style.WithAlpha(black, .16)).
		On(style.Disabled, style.Style{}.
			Background(style.Solid(active.DisabledColor(white))).
			BorderColor(style.Solid(active.DisabledColor(sem.Border))).
			BoxShadowNone())

	return style.Style{}.
		Part(style.PartTrack, style.Join(track, trackMargin)).
		Part(style.PartFill, fill).
		Part(style.PartThumb, style.Join(thumb, thumbOverlay)).
		Part(style.PartLabel, label)
}

type sliderPalette struct {
	Track, Fill, ThumbBorder, ThumbBackground, Focus color.NRGBA
}

func sliderColors(active *theme.Theme, disabled bool) sliderPalette {
	sem := active.Semantic()
	value := sliderPalette{
		Track:           sem.SurfaceRaised,
		Fill:            sem.Tone(theme.TonePrimary).Fill,
		ThumbBorder:     sem.Tone(theme.TonePrimary).Fill,
		ThumbBackground: sem.Tone(theme.TonePrimary).OnFill,
		Focus:           sem.Ring,
	}
	if disabled {
		value.Track = active.DisabledColor(value.Track)
		value.Fill = active.DisabledColor(value.Fill)
		value.ThumbBorder = active.DisabledColor(value.ThumbBorder)
		value.ThumbBackground = active.DisabledColor(value.ThumbBackground)
		value.Focus = color.NRGBA{}
	}
	return value
}

func sliderShadowAlpha(alpha uint8, background color.NRGBA) uint8 {
	return uint8(uint16(alpha) * uint16(background.A) / 0xff)
}
