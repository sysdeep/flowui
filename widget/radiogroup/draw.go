package radiogroup

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// drawRadioIndicator owns the selected-dot geometry. The caller supplies the
// resolved foreground, selection progress, and press state.
func drawRadioIndicator(
	gtx layout.Context,
	size image.Point,
	foreground color.NRGBA,
	selection float32,
	pressed bool,
	dotScale, dotPressedScale float32,
) layout.Dimensions {
	if selection <= 0 || size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	targetScale := dotScale
	if pressed {
		targetScale = dotPressedScale
	}
	diameter := min(size.X, size.Y)
	dotSize := max(int(float32(diameter)*targetScale*selection+.5), 1)
	center := image.Pt(size.X/2, size.Y/2)
	dot := image.Rectangle{Min: center.Sub(image.Pt(dotSize/2, dotSize/2))}
	dot.Max = dot.Min.Add(image.Pt(dotSize, dotSize))
	if foreground.A != 0 {
		paint.FillShape(gtx.Ops, foreground, clip.Ellipse(dot).Op(gtx.Ops))
	}
	return layout.Dimensions{Size: size}
}
