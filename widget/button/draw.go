package button

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func drawSpinner(
	gtx layout.Context,
	size, stroke unit.Dp,
	col color.NRGBA,
	period time.Duration,
) layout.Dimensions {
	d := max(gtx.Dp(size), 1)
	strokeWidth := float32(max(gtx.Dp(stroke), 1))
	radius := float32(d)/2 - strokeWidth/2
	if radius <= 0 {
		return layout.Dimensions{Size: image.Pt(d, d)}
	}

	phase := float32(0)
	if period > 0 && !gtx.Now.IsZero() {
		elapsed := gtx.Now.UnixNano() % int64(period)
		if elapsed < 0 {
			elapsed += int64(period)
		}
		phase = float32(elapsed) / float32(period) * float32(math.Pi*2)
	}
	const segments = 12
	const sweep = float32(math.Pi * 1.6)
	center := f32.Pt(float32(d)/2, float32(d)/2)
	var path clip.Path
	path.Begin(gtx.Ops)
	for index := range segments + 1 {
		angle := phase + sweep*float32(index)/segments
		point := f32.Pt(
			center.X+float32(math.Cos(float64(angle)))*radius,
			center.Y+float32(math.Sin(float64(angle)))*radius,
		)
		if index == 0 {
			path.MoveTo(point)
			continue
		}
		path.LineTo(point)
	}
	strokeClip := clip.Stroke{Path: path.End(), Width: strokeWidth}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	strokeClip.Pop()
	return layout.Dimensions{Size: image.Pt(d, d)}
}

func drawGroupSeparator(
	gtx layout.Context,
	size image.Point,
	mainOffset int,
	orientation GroupOrientation,
	foreground color.NRGBA,
) {
	thickness := max(gtx.Dp(1), 1)
	foreground.A = uint8(float32(foreground.A)*.15 + .5)
	var rect image.Rectangle
	if orientation == GroupVertical {
		length := max(size.X/2, 1)
		rect = image.Rect((size.X-length)/2, mainOffset-thickness/2, (size.X+length)/2, mainOffset-thickness/2+thickness)
	} else {
		length := max(size.Y/2, 1)
		rect = image.Rect(mainOffset-thickness/2, (size.Y-length)/2, mainOffset-thickness/2+thickness, (size.Y+length)/2)
	}
	paint.FillShape(gtx.Ops, foreground, clip.Rect(rect).Op())
}
