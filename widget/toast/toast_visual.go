package toast

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

// toastMetrics is the geometry and timing consumed by toast layout.
type toastMetrics struct {
	Width             unit.Dp
	Inset             unit.Dp
	Gap               unit.Dp
	ContentGap        unit.Dp
	IndicatorPadding  unit.Dp
	IndicatorSize     unit.Dp
	CloseInset        unit.Dp
	CloseButtonSize   unit.Dp
	ActionInsetTop    unit.Dp
	MaxVisible        int
	ScaleFactor       float32
	AnimationDuration time.Duration
	DefaultTimeout    time.Duration
}

var defaultToastMetrics = toastMetrics{
	Width: 460, Inset: 16, Gap: 12, ContentGap: 6,
	IndicatorPadding: 4, IndicatorSize: 16, CloseInset: -4,
	CloseButtonSize: 28, ActionInsetTop: 6, MaxVisible: 3,
	ScaleFactor: 0, AnimationDuration: 350 * time.Millisecond,
	DefaultTimeout: 4 * time.Second,
}

// toastVisual is the appearance of one toast variant. Every visible element is
// a style layer that the layout resolves with the instance style as the last
// cascade layer; the metrics and motion are plain values the overlay math
// reads directly.
type toastVisual struct {
	panel     style.Style
	text      style.Style
	indicator style.Style
	action    style.Style
	close     style.Style
	metrics   toastMetrics
	motion    theme.MotionTheme
	icon      []byte
}

func buildToastVisual(active *theme.Theme, variant Variant, actionVariant button.Variant) toastVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := defaultToastMetrics
	semantic := active.Semantic()
	title := semantic.OverlayForeground
	indicator := title
	switch variant {
	case Accent:
		title = semantic.Tone(theme.TonePrimary).OnSoft
	case Success:
		title = semantic.Tone(theme.ToneSuccess).OnSoft
		indicator = title
	case Warning:
		title = semantic.Tone(theme.ToneWarning).OnSoft
		indicator = title
	case Danger:
		title = semantic.Tone(theme.ToneDanger).OnSoft
		indicator = title
	}
	closeSize := tokens.CloseButtonSize
	return toastVisual{
		panel: style.Style{}.
			Background(style.Solid(semantic.Surface)).
			TextColor(style.Solid(semantic.OverlayForeground)).
			Padding(12, 16).
			Radius(8).
			Shadow(style.ShadowOverlay).
			Overflow(style.OverflowHidden).
			Opacity(1).
			OnNot(style.FocusVisible, style.Disabled,
				style.Style{}.Outline(2, 0, style.Solid(semantic.Ring))),
		text: style.Style{}.
			TextColor(style.Solid(title)).
			FontSize(14).
			FontWeight(int(font.Medium)).
			Part(style.PartDescription, style.Style{}.
				TextColor(style.Solid(semantic.MutedForeground)).
				FontSize(14).
				FontWeight(int(font.Normal))),
		indicator: style.Style{}.TextColor(style.Solid(indicator)),
		action:    toastActionStyle(active, actionVariant),
		close:     toastCloseStyle(active, closeSize),
		metrics:   tokens,
		motion:    active.Motion,
		icon:      toastIcon(variant),
	}
}

func toastActionStyle(active *theme.Theme, variant button.Variant) style.Style {
	semantic := active.Semantic()
	background, foreground := toastButtonColors(active, variant)
	result := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		FontSize(14).
		FontWeight(int(font.Medium)).
		Padding(6, 12).
		Radius(4).
		On(style.Hovered, style.Style{}.Background(style.Solid(semantic.SurfacePressed))).
		On(style.Pressed, style.Style{}.Background(style.Solid(semantic.SurfacePressed)))
	if variant == button.Secondary {
		result = result.BorderWidth(1).BorderColor(style.Solid(semantic.Tone(theme.TonePrimary).Fill))
	}
	return result
}

func toastCloseStyle(active *theme.Theme, size unit.Dp) style.Style {
	semantic := active.Semantic()
	return style.Style{}.
		Width(size).
		Height(size).
		Radius(max(size/2, 1)).
		Background(style.Solid(semantic.SurfacePressed)).
		TextColor(style.Solid(semantic.MutedForeground)).
		Cursor(pointer.CursorPointer)
}

func toastButtonColors(active *theme.Theme, variant button.Variant) (color.NRGBA, color.NRGBA) {
	semantic := active.Semantic()
	switch variant {
	case button.Danger, button.DangerOutline, button.DangerSoft:
		tone := semantic.Tone(theme.ToneDanger)
		if variant == button.Danger {
			return tone.Fill, tone.OnFill
		}
		return tone.Soft, tone.OnSoft
	case button.Secondary:
		tone := semantic.Tone(theme.TonePrimary)
		return tone.Soft, tone.OnSoft
	case button.Tertiary, button.Ghost, button.Outline:
		return color.NRGBA{}, semantic.Tone(theme.TonePrimary).OnSoft
	default:
		tone := semantic.Tone(theme.TonePrimary)
		return tone.Fill, tone.OnFill
	}
}

func toastIcon(variant Variant) []byte {
	switch variant {
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
