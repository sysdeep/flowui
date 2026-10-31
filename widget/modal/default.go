package modal

import (
	"image/color"
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// modalMetrics is the geometry and timing consumed by modal layout.
type modalMetrics struct {
	XSmallWidth       unit.Dp
	SmallWidth        unit.Dp
	MediumWidth       unit.Dp
	LargeWidth        unit.Dp
	Margin            unit.Dp
	DesktopMargin     unit.Dp
	DesktopBreakpoint unit.Dp
	HeaderGap         unit.Dp
	BodyGap           unit.Dp
	FooterGap         unit.Dp
	SectionGap        unit.Dp
	IconSize          unit.Dp
	Padding           unit.Dp
	CloseInset        unit.Dp
	CloseButtonSize   unit.Dp
	AnimationScale    float32
	AnimationDuration time.Duration
	AnimationDistance unit.Dp

	AnimationBounceScale       float32
	AnimationBounceSettleScale float32
	AnimationZoomScale         float32
	AnimationPopStartScale     float32
	AnimationPopOvershootScale float32
	AnimationPopSettleScale    float32
}

// modalDefaults is the theme-derived appearance and geometry of one modal
// configuration. The instance style is deliberately absent: it is resolved as
// the last cascade layer when each part is laid out.
type modalDefaults struct {
	defaults       style.Style
	header         style.Style
	body           style.Style
	metrics        modalMetrics
	motion         theme.MotionTheme
	iconBackground color.NRGBA
}

var defaultModalMetrics = modalMetrics{
	XSmallWidth: 320, SmallWidth: 384, MediumWidth: 448, LargeWidth: 512,
	Margin: 16, DesktopMargin: 40, DesktopBreakpoint: 640,
	HeaderGap: 12, BodyGap: 8, FooterGap: 8, SectionGap: 20,
	IconSize: 40, Padding: 24, CloseInset: 16, CloseButtonSize: 24, AnimationScale: 1,
	AnimationDuration: 250 * time.Millisecond, AnimationDistance: 24,
	AnimationBounceScale: 1, AnimationBounceSettleScale: 1,
	AnimationZoomScale: 1, AnimationPopStartScale: 1,
	AnimationPopOvershootScale: 1, AnimationPopSettleScale: 1,
}

func buildModalDefaults(active *theme.Theme, size Size, backdrop BackdropVariant) modalDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	backdropColor := color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x66}
	switch backdrop {
	case BackdropBlur:
		backdropColor = color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x7a}
	case BackdropTransparent:
		backdropColor = color.NRGBA{}
	}
	defaults := style.Style{}.
		Background(style.Solid(sem.Overlay)).
		TextColor(style.Solid(sem.OverlayForeground)).
		Padding(defaultModalMetrics.Padding).
		Overflow(style.OverflowHidden).
		Part(style.PartBackdrop, style.Style{}.Background(style.Solid(backdropColor)))
	if size != Full {
		defaults = defaults.Radius(12).Shadow(style.ShadowOverlay)
	}
	return modalDefaults{
		defaults: defaults,
		header: style.Style{}.
			FontSize(20).
			FontWeight(500).
			TextColor(style.Solid(sem.OverlayForeground)),
		body: style.Style{}.
			FontSize(14).
			TextColor(style.Solid(sem.MutedForeground)),
		metrics:        defaultModalMetrics,
		motion:         active.Motion,
		iconBackground: sem.Tone(theme.TonePrimary).Soft,
	}
}

// dialogPadding reads the panel insets from the resolved root declaration so
// close-button placement follows the instance cascade.
func dialogPadding(resolved style.ResolvedStyle) style.Insets {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		return style.Insets{}
	}
	return *resolved.Box.Padding
}
