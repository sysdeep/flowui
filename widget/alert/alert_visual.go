package alert

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// alertMetrics is the geometry consumed by alert layout.
type alertMetrics struct {
	Gap              unit.Dp
	IndicatorPadding unit.Dp
	IconSize         unit.Dp
}

// alertVisual is the resolved appearance of one alert status.
type alertVisual struct {
	Defaults    style.Style
	Title       style.Style
	Description style.Style
	Content     style.Style
	Indicator   style.Style
	Metrics     alertMetrics
	Icon        []byte
}

func buildAlertVisual(active *theme.Theme, status Status) alertVisual {
	semantic := active.Semantic()
	statusColor := alertStatusColor(active, status)
	return alertVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			TextColor(style.Solid(semantic.SurfaceForeground)).
			Padding(12).
			Radius(8).
			Overflow(style.OverflowHidden).
			Shadow(style.ShadowSurface).
			Opacity(1),
		Title: style.Style{}.
			TextColor(style.Solid(statusColor)).
			FontSize(14).
			FontWeight(int(font.Medium)).
			LineHeight(24),
		Description: style.Style{}.
			TextColor(style.Solid(semantic.MutedForeground)).
			FontSize(14).
			LineHeight(20),
		Content: style.Style{}.
			TextColor(style.Solid(semantic.MutedForeground)).
			FontSize(14).
			LineHeight(20),
		Indicator: style.Style{}.
			TextColor(style.Solid(statusColor)),
		Metrics: alertMetrics{
			Gap:              12,
			IndicatorPadding: 4,
			IconSize:         16,
		},
		Icon: alertStatusIcon(status),
	}
}

func alertStatusColor(active *theme.Theme, status Status) color.NRGBA {
	semantic := active.Semantic()
	value := semantic.SurfaceForeground
	switch status {
	case Accent:
		value = semantic.Tone(theme.TonePrimary).OnSoft
	case Success:
		value = semantic.Tone(theme.ToneSuccess).OnSoft
	case Warning:
		value = semantic.Tone(theme.ToneWarning).OnSoft
	case Danger:
		value = semantic.Tone(theme.ToneDanger).OnSoft
	}
	return value
}

func alertStatusIcon(status Status) []byte {
	switch status {
	case Success:
		return lucide.CircleCheck
	case Warning:
		return lucide.TriangleAlert
	case Danger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
