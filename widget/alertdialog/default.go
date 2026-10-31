package alertdialog

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// alertDialogMetrics is the header geometry consumed by alert dialog layout.
type alertDialogMetrics struct {
	IconSize      unit.Dp
	IconGlyphSize unit.Dp
	HeaderGap     unit.Dp
}

// alertDialogIcon is the status icon appearance the dialog paints itself.
type alertDialogIcon struct {
	background color.NRGBA
	foreground color.NRGBA
	glyph      []byte
}

// alertDialogDefaults is the theme-derived header appearance of one alert
// dialog status. The panel and body belong to the composed modal, so they are
// not part of it, and the instance style is resolved as the last cascade layer
// instead of being cached.
type alertDialogDefaults struct {
	header  style.Style
	icon    alertDialogIcon
	metrics alertDialogMetrics
}

func buildAlertDialogDefaults(active *theme.Theme, status Status) alertDialogDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	background := sem.SurfaceSecondary
	foreground := sem.OverlayForeground
	if tone, tinted := alertDialogStatusTone(active, status); tinted {
		background = tone.Soft
		foreground = tone.OnSoft
	}
	return alertDialogDefaults{
		header: style.Style{}.
			FontSize(20).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.OverlayForeground)),
		icon: alertDialogIcon{
			background: background,
			foreground: foreground,
			glyph:      alertDialogStatusIcon(status),
		},
		metrics: alertDialogMetrics{
			IconSize:      40,
			IconGlyphSize: 20,
			HeaderGap:     12,
		},
	}
}

func alertDialogStatusTone(active *theme.Theme, status Status) (theme.TonePair, bool) {
	switch status {
	case StatusAccent:
		return active.Semantic().Tone(theme.TonePrimary), true
	case StatusSuccess:
		return active.Semantic().Tone(theme.ToneSuccess), true
	case StatusWarning:
		return active.Semantic().Tone(theme.ToneWarning), true
	case StatusDanger:
		return active.Semantic().Tone(theme.ToneDanger), true
	default:
		return active.Semantic().Tone(theme.ToneDefault), false
	}
}

func alertDialogStatusIcon(status Status) []byte {
	switch status {
	case StatusSuccess:
		return lucide.CircleCheck
	case StatusWarning:
		return lucide.TriangleAlert
	case StatusDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
