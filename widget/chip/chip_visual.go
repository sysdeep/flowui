package chip

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// chipVisual is the resolved appearance of one chip configuration. The chip
// owns both its behavior and its default look; the active theme only supplies
// tokens, so this struct is package-private.
type chipVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	// LabelPaddingX reserves the label's inset when computing accessory space.
	LabelPaddingX unit.Dp
	// ContentGap separates the label from start/end content.
	ContentGap unit.Dp
}

func buildChipVisual(active *theme.Theme, chipColor Color, variant Variant, size Size) chipVisual {
	height, paddingX, paddingY, textSize := chipSize(size)
	background, foreground := chipColors(active, chipColor, variant)
	visible := style.Style{}.TextColor(style.Solid(foreground))
	if variant != VariantTertiary {
		visible = visible.Background(style.Solid(background))
	}
	return chipVisual{
		Defaults: style.Style{}.
			Radius(4).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.
				FontWeight(int(font.Medium)).
				MaxLines(1)),
		Variant: visible,
		Size: style.Style{}.
			Height(height).
			Padding(paddingY, paddingX).
			Part(style.PartLabel, style.Style{}.
				FontSize(textSize).
				LineHeight(20)),
		LabelPaddingX: 2,
		ContentGap:    2,
	}
}

func chipSize(size Size) (height, paddingX, paddingY unit.Dp, textSize unit.Sp) {
	height, paddingX, paddingY, textSize = 24, 8, 2, 12
	switch size {
	case SizeSmall:
		height, paddingX, paddingY, textSize = 20, 4, 0, 12
	case SizeLarge:
		height, paddingX, paddingY, textSize = 28, 12, 4, 14
	}
	return
}

func chipColors(active *theme.Theme, chipColor Color, variant Variant) (color.NRGBA, color.NRGBA) {
	semantic := active.Semantic()
	defaultBackground := semantic.SurfaceTertiary
	defaultForeground := semantic.SurfaceTertiaryForeground
	tone, tinted := chipTone(semantic, chipColor)

	solidBackground, solidForeground := defaultBackground, defaultForeground
	softBackground, softForeground := chipSoftColor(defaultBackground), defaultForeground
	if tinted {
		solidBackground, solidForeground = tone.Fill, tone.OnFill
		softBackground, softForeground = tone.Soft, tone.OnSoft
	}

	switch variant {
	case VariantPrimary:
		return solidBackground, solidForeground
	case VariantTertiary:
		return color.NRGBA{}, softForeground
	case VariantSoft:
		return softBackground, softForeground
	default:
		return defaultBackground, softForeground
	}
}

func chipTone(semantic theme.Semantic, chipColor Color) (theme.TonePair, bool) {
	switch chipColor {
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

func chipSoftColor(value color.NRGBA) color.NRGBA {
	value.A = byte((uint16(value.A) + 1) / 2)
	return value
}
