package badge

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// badgeVisual is the resolved appearance of one badge configuration. The
// badge owns both its behavior and its default look; the active theme only
// supplies tokens, so this struct is package-private.
type badgeVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	// PlacementOffsetRatio shifts the badge outward from its anchor corner.
	PlacementOffsetRatio float32
}

// buildBadgeVisual maps the badge's semantic axes onto declarations. Colors
// come from the active semantic layer and geometry from the component tokens.
func buildBadgeVisual(active *theme.Theme, badgeColor Color, variant Variant, size Size) badgeVisual {
	background, foreground := badgeColors(active, badgeColor, variant)
	minSize, radius, textSize, lineHeight := badgeSize(size)
	return badgeVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(active.Semantic().Surface)).
			BorderWidth(1).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.
				FontWeight(int(font.Medium)).
				MaxLines(1)),
		Variant: style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)),
		Size: style.Style{}.
			MinWidth(minSize).
			MinHeight(minSize).
			Padding(0, 2).
			Radius(radius).
			Part(style.PartLabel, style.Style{}.
				FontSize(textSize).
				LineHeight(lineHeight)),
		PlacementOffsetRatio: .25,
	}
}

func badgeSize(size Size) (minSize, radius unit.Dp, textSize, lineHeight unit.Sp) {
	minSize, radius, textSize, lineHeight = 28, 4, 12, 16
	switch size {
	case SizeSmall:
		minSize, radius, textSize, lineHeight = 16, 3, 10, 14
	case SizeLarge:
		minSize, radius, textSize, lineHeight = 32, 6, 14, 20
	}
	return
}

func badgeColors(active *theme.Theme, badgeColor Color, variant Variant) (color.NRGBA, color.NRGBA) {
	semantic := active.Semantic()
	neutral := semantic.Tone(theme.ToneDefault)
	tone, tinted := badgeTone(semantic, badgeColor)
	switch variant {
	case VariantSecondary:
		return neutral.Fill, tone.OnSoft
	case VariantSoft:
		if tinted {
			return tone.Soft, tone.OnSoft
		}
		background := neutral.Fill
		background.A = byte((uint16(background.A) + 1) / 2)
		return background, neutral.OnFill
	default:
		return tone.Fill, tone.OnFill
	}
}

func badgeTone(semantic theme.Semantic, badgeColor Color) (theme.TonePair, bool) {
	switch badgeColor {
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
