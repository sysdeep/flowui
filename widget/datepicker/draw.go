package datepicker

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/widget/icon"
)

func dateDrawCalendarIcon(gtx layout.Context, size image.Point, foreground color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(lucide.Calendar, iconGtx, foreground)
}

func dateDrawYearIndicator(gtx layout.Context, size image.Point, open bool, foreground color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	data := lucide.ChevronRight
	if open {
		data = lucide.ChevronDown
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(data, iconGtx, foreground)
}

func dateDrawNavButton(
	gtx layout.Context,
	size image.Point,
	direction int,
	background, foreground color.NRGBA,
	radius int,
	chevronStroke unit.Dp,
) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if background.A != 0 {
		paint.FillShape(gtx.Ops, background,
			clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
	}
	data := lucide.ChevronRight
	if direction < 0 {
		data = lucide.ChevronLeft
	}
	diameter := min(icon.LucideSizeForStroke(gtx, chevronStroke), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(data, iconGtx, foreground)
	offset.Pop()
}

func dateDrawCell(gtx layout.Context, size image.Point, background color.NRGBA, radius int) {
	if size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return
	}
	radius = min(radius, min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, background,
		clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func dateDrawRange(gtx layout.Context, size image.Point, track color.NRGBA, startRadius, endRadius int) {
	if size.X <= 0 || size.Y <= 4 {
		return
	}
	rect := image.Rect(0, 2, size.X, size.Y-2)
	startRadius = min(max(startRadius, 0), rect.Dy()/2)
	endRadius = min(max(endRadius, 0), rect.Dy()/2)
	paint.FillShape(gtx.Ops, track, clip.RRect{
		Rect: rect,
		NW:   startRadius,
		SW:   startRadius,
		NE:   endRadius,
		SE:   endRadius,
	}.Op(gtx.Ops))
}

func dateDrawStrike(gtx layout.Context, size image.Point, foreground color.NRGBA, halfSize, width unit.Dp) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	strike := foreground
	strike.A = byte(uint16(strike.A) * 3 / 4)
	var path clip.Path
	path.Begin(gtx.Ops)
	y := float32(size.Y) / 2
	half := draw.DpFloat(gtx, halfSize)
	path.MoveTo(f32.Pt(float32(size.X)/2-half, y))
	path.LineTo(f32.Pt(float32(size.X)/2+half, y))
	stroke := clip.Stroke{Path: path.End(), Width: max(draw.DpFloat(gtx, width), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, strike)
	stroke.Pop()
}

func dateDrawFocus(
	gtx layout.Context,
	size image.Point,
	foreground color.NRGBA,
	radius int,
	visible, trigger bool,
	focusWidth, iconSize, segmentRadius unit.Dp,
) {
	if !visible || size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	if trigger {
		diameter := min(gtx.Dp(iconSize+8), min(size.X, size.Y))
		rect = image.Rect((size.X-diameter)/2, (size.Y-diameter)/2,
			dateSizeCenter(size.X, diameter), dateSizeCenter(size.Y, diameter))
		radius = gtx.Dp(segmentRadius)
	}
	width := max(gtx.Dp(focusWidth), 1)
	rect = rect.Inset((width + 1) / 2)
	if rect.Empty() {
		return
	}
	radius = max(radius-(width+1)/2, 0)
	draw.RoundedInsetStroke(gtx, rect, radius, width, 0, foreground)
}

func dateSizeCenter(size, span int) int { return (size + span) / 2 }
