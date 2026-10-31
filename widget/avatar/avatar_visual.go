package avatar

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

// avatarVisual is the resolved appearance of one avatar configuration.
type avatarVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
}

func buildAvatarVisual(active *theme.Theme, avatarColor Color, variant Variant, size Size) avatarVisual {
	diameter, radius, textSize, iconSize := avatarSize(size)
	background, foreground := avatarColors(active, avatarColor, variant)
	return avatarVisual{
		Defaults: style.Style{}.
			Overflow(style.OverflowHidden).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium)).MaxLines(1)),
		Variant: style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)),
		Size: style.Style{}.
			Width(diameter).
			AspectRatio(1).
			Radius(radius).
			Part(style.PartLabel, style.Style{}.FontSize(textSize)).
			Part(style.PartIcon, style.Style{}.Width(iconSize).Height(iconSize)),
	}
}

func avatarSize(size Size) (diameter, radius unit.Dp, textSize unit.Sp, iconSize unit.Dp) {
	diameter, radius, textSize, iconSize = 40, 24, 14, 20
	switch size {
	case SizeSmall:
		diameter, radius, textSize, iconSize = 32, 16, 14, 16
	case SizeLarge:
		diameter, radius, textSize, iconSize = 48, 24, 16, 24
	}
	return
}

func avatarColors(active *theme.Theme, avatarColor Color, variant Variant) (color.NRGBA, color.NRGBA) {
	semantic := active.Semantic()
	neutral := semantic.Tone(theme.ToneDefault)
	tone, tinted := avatarTone(semantic, avatarColor)
	foreground := neutral.OnFill
	if tinted {
		foreground = tone.OnSoft
	}
	background := neutral.Fill
	if variant != VariantSoft {
		return background, foreground
	}
	if tinted {
		return tone.Soft, foreground
	}
	background.A = byte((uint16(background.A) + 1) / 2)
	return background, foreground
}

func avatarTone(semantic theme.Semantic, avatarColor Color) (theme.TonePair, bool) {
	switch avatarColor {
	case ColorAccent:
		return semantic.Tone(theme.TonePrimary), true
	case ColorSuccess:
		return semantic.Tone(theme.ToneSuccess), true
	case ColorWarning:
		return semantic.Tone(theme.ToneWarning), true
	case ColorDanger:
		return semantic.Tone(theme.ToneDanger), true
	default:
		return semantic.Tone(theme.ToneDefault), false
	}
}

func avatarIcon(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	return icon.Layout(lucide.UserRound, gtx, col)
}
