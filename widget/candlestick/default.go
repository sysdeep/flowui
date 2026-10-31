package candlestick

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// candlestickDefaults is the cached, theme-derived appearance and geometry of a
// candlestick chart. The instance style is resolved separately as the last
// cascade layer, so it never becomes part of the cache key.
type candlestickDefaults struct {
	Defaults        style.Style
	Metrics         candlestickDimensions
	Colors          candlestickPalette
	DisabledOpacity float32
	Motion          theme.MotionTheme
}

// candlestickDimensions contains the geometry, typography, and animation timing
// consumed by candlestick layout and drawing.
type candlestickDimensions struct {
	AnimationDuration     time.Duration
	UpdateDuration        time.Duration
	Height                unit.Dp
	PlotPaddingTop        unit.Dp
	PlotPaddingRight      unit.Dp
	PlotPaddingBottom     unit.Dp
	PlotPaddingLeft       unit.Dp
	AxisNameGap           unit.Dp
	TickLabelGap          unit.Dp
	AxisTextSize          unit.Sp
	GridWidth             unit.Dp
	AxisWidth             unit.Dp
	AnnotationLineWidth   unit.Dp
	AnnotationPointSize   unit.Dp
	CrosshairWidth        unit.Dp
	CrosshairLabelPadding unit.Dp
	CrosshairLabelRadius  unit.Dp
	WickWidth             unit.Dp
	BorderWidth           unit.Dp
	EmphasisBorderWidth   unit.Dp
	TooltipGap            unit.Dp
	TooltipRowGap         unit.Dp
	TooltipTextSize       unit.Sp
	TextWeight            font.Weight
	EmphasisWeight        font.Weight
}

var defaultCandlestickDimensions = candlestickDimensions{
	AnimationDuration:     300 * time.Millisecond,
	UpdateDuration:        300 * time.Millisecond,
	Height:                360,
	PlotPaddingTop:        12,
	PlotPaddingRight:      16,
	PlotPaddingBottom:     36,
	PlotPaddingLeft:       56,
	AxisNameGap:           6,
	TickLabelGap:          8,
	AxisTextSize:          12,
	GridWidth:             1,
	AxisWidth:             1,
	AnnotationLineWidth:   1,
	AnnotationPointSize:   9,
	CrosshairWidth:        1,
	CrosshairLabelPadding: 4,
	CrosshairLabelRadius:  4,
	WickWidth:             1,
	BorderWidth:           1,
	EmphasisBorderWidth:   2,
	TooltipGap:            12,
	TooltipRowGap:         5,
	TooltipTextSize:       12,
	TextWeight:            font.Normal,
	EmphasisWeight:        font.Medium,
}

// candlestickPalette contains the palette decisions used by candlestick
// drawing. Explicit OHLC colors remain widget inputs and replace these defaults
// when set.
type candlestickPalette struct {
	Axis                     color.NRGBA
	AxisLabel                color.NRGBA
	Grid                     color.NRGBA
	Crosshair                color.NRGBA
	CrosshairLabelBackground color.NRGBA
	CrosshairLabelForeground color.NRGBA
	OverlayForeground        color.NRGBA
	AnnotationArea           color.NRGBA
	AnnotationLine           color.NRGBA
	AnnotationPoint          color.NRGBA
	Up                       color.NRGBA
	Down                     color.NRGBA
	Doji                     color.NRGBA
}

// buildCandlestickDefaults returns the theme-derived defaults for a
// candlestick chart.
func buildCandlestickDefaults(active *theme.Theme) candlestickDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	background := semantic.MutedForeground
	foreground := semantic.Background
	return candlestickDefaults{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: defaultCandlestickDimensions,
		Colors: candlestickPalette{
			Axis:                     semantic.Border,
			AxisLabel:                semantic.MutedForeground,
			Grid:                     candleWithAlpha(semantic.Border, .8),
			Crosshair:                candleWithAlpha(semantic.MutedForeground, .75),
			CrosshairLabelBackground: background,
			CrosshairLabelForeground: foreground,
			OverlayForeground:        semantic.OverlayForeground,
			AnnotationArea:           candleWithAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:           candleWithAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:          semantic.Tone(theme.ToneWarning).Fill,
			Up:                       color.NRGBA{R: 0xeb, G: 0x54, B: 0x54, A: 0xff},
			Down:                     color.NRGBA{R: 0x47, G: 0xb2, B: 0x62, A: 0xff},
			Doji:                     semantic.MutedForeground,
		},
		DisabledOpacity: active.DisabledOpacityValue(),
		Motion:          active.Motion,
	}
}

func candleWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
