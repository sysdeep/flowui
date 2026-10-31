package colorpicker

import (
	"image/color"

	"gioui.org/layout"
)

type ColorChannel uint8

const (
	ColorChannelHue ColorChannel = iota
	ColorChannelAlpha
)

type ColorSwatchSize uint8

const (
	ColorSwatchExtraSmall ColorSwatchSize = iota
	ColorSwatchSmall
	ColorSwatchMedium
	ColorSwatchLarge
	ColorSwatchExtraLarge
)

type ColorSwatchShape uint8

const (
	ColorSwatchCircle ColorSwatchShape = iota
	ColorSwatchSquare
)

type ColorSwatchPickerLayout uint8

const (
	ColorSwatchPickerGrid ColorSwatchPickerLayout = iota
	ColorSwatchPickerStack
)

type colorValueState struct {
	syncedColor color.NRGBA
	ready       bool
	retainedHue float64
	hueReady    bool
}

func (state *colorValueState) sync(value color.NRGBA) {
	if state.ready && state.syncedColor == value {
		return
	}
	resolved := nrgbaToHSV(value)
	rgbChanged := !state.ready || state.syncedColor.R != value.R || state.syncedColor.G != value.G || state.syncedColor.B != value.B
	if !state.hueReady || rgbChanged && resolved.s > 0 && resolved.v > 0 {
		state.retainedHue = resolved.h
		state.hueReady = true
	}
	state.syncedColor = value
	state.ready = true
}

func (state *colorValueState) accept(value color.NRGBA, hue float64) {
	state.syncedColor = value
	state.ready = true
	state.retainedHue = clampUnit(hue)
	state.hueReady = true
}

func (state *colorValueState) hsv() hsvColor {
	resolved := nrgbaToHSV(state.syncedColor)
	if state.hueReady {
		resolved.h = state.retainedHue
	}
	return resolved
}

func colorSwatchPixels(gtx layout.Context, defaults colorSwatchDefaults, size ColorSwatchSize) int {
	switch size {
	case ColorSwatchExtraSmall:
		return gtx.Dp(defaults.ExtraSmallSize)
	case ColorSwatchSmall:
		return gtx.Dp(defaults.SmallSize)
	case ColorSwatchLarge:
		return gtx.Dp(defaults.LargeSize)
	case ColorSwatchExtraLarge:
		return gtx.Dp(defaults.ExtraLargeSize)
	default:
		return gtx.Dp(defaults.MediumSize)
	}
}
