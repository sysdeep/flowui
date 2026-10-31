package colorpicker

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/qianniancn/flowui/draw"
)

type colorWheelImageCache struct {
	size image.Point
	op   paint.ImageOp
}

func colorPickerTriggerLabel(gtx layout.Context, defaults colorPickerDefaults, value string) layout.Dimensions {
	label := material.Label(defaults.Material, defaults.TriggerTextSize, value)
	label.Color = defaults.Foreground
	return label.Layout(gtx)
}

func colorPickerTriggerFocus(gtx layout.Context, defaults colorPickerDefaults, size image.Point, opacity float32) {
	if opacity <= 0 {
		return
	}
	width := max(gtx.Dp(defaults.FocusRingWidth), 1)
	rect := image.Rectangle{Max: size}.Inset(width)
	if rect.Empty() {
		return
	}
	col := defaults.FocusColor
	col.A = byte(float32(col.A)*opacity + .5)
	radius := max(gtx.Dp(defaults.TriggerRadius)-width, 0)
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func colorDrawColorSwatch(gtx layout.Context, size image.Point, value color.NRGBA, radius int) {
	colorDrawColorSwatchFill(gtx, size, value, radius)
	colorDrawRoundedStroke(gtx, image.Rectangle{Max: size}, radius, 1, color.NRGBA{A: 26})
}

func colorDrawColorSwatchFill(gtx layout.Context, size image.Point, value color.NRGBA, radius int) {
	rect := image.Rectangle{Max: size}
	if value.A < 255 {
		colorDrawCheckerboard(gtx, rect, radius)
	}
	paint.FillShape(gtx.Ops, value, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func colorDrawColorSwatchPickerItem(
	gtx layout.Context,
	defaults colorSwatchPickerDefaults,
	size image.Point,
	value, background color.NRGBA,
	swatchSize uint8,
	shape ColorSwatchShape,
	selection, check, hover, focus float32,
) {
	rect := image.Rectangle{Max: size}
	minimum := min(size.X, size.Y)
	borderWidth := colorSwatchPickerBorderWidth(gtx, defaults, swatchSize)
	itemRadius := colorSwatchPickerItemRadius(gtx, defaults, minimum, swatchSize, shape)
	if focus > 0 {
		focusColor := defaults.FocusColor
		focusColor.A = byte(float32(focusColor.A)*focus + .5)
		focusWidth := max(gtx.Dp(defaults.FocusRingWidth), 1)
		focusGap := max(gtx.Dp(defaults.FocusRingGap), 0)
		expand := focusGap + focusWidth
		focusRect := rect.Inset(-expand)
		focusRadius := itemRadius + expand
		if shape == ColorSwatchCircle {
			focusRadius = min(focusRect.Dx(), focusRect.Dy()) / 2
		}
		colorDrawRoundedBorder(gtx, focusRect, focusRadius, focusWidth, focusColor, background)
	}
	if selection > 0 {
		shadowOpacity := defaults.ShadowOpacity * selection
		if shadowOpacity > 0 {
			shadowShape := draw.EllipseShadow()
			if shape == ColorSwatchSquare {
				radius := colorSwatchPickerItemRadiusDp(defaults, swatchSize)
				shadowShape = draw.RoundedShadowCorners(radius, radius, radius, radius)
			}
			draw.Shadow(gtx, rect, shadowShape, draw.ThemeShadow(defaults.Shadows, defaults.ShadowColor, shadowOpacity))
		}
		border := value
		border.A = byte(float32(255)*selection + .5)
		colorDrawRoundedBorder(gtx, rect, itemRadius, borderWidth, border, background)
	}
	scale := float32(1) + .1*hover
	if selection > 0 {
		scale = 1 + (defaults.SelectedScale-1)*selection
	}
	swatchPixels := colorSwatchPickerVisualSize(size, borderWidth, scale)
	swatchRect := image.Rectangle{Min: image.Pt((size.X-swatchPixels.X)/2, (size.Y-swatchPixels.Y)/2)}
	swatchRect.Max = swatchRect.Min.Add(swatchPixels)
	offset := op.Offset(swatchRect.Min).Push(gtx.Ops)
	colorDrawColorSwatchFill(gtx, swatchRect.Size(), value, colorSwatchPickerSwatchRadius(gtx, defaults, min(swatchRect.Dx(), swatchRect.Dy()), swatchSize, shape, selection))
	offset.Pop()
	if check > 0 {
		colorDrawColorSwatchCheck(gtx, size, value, check, float32(gtx.Dp(defaults.CheckStroke)))
	}
}

func colorSwatchPickerBorderWidth(gtx layout.Context, defaults colorSwatchPickerDefaults, size uint8) int {
	switch size {
	case 0:
		return max(gtx.Dp(defaults.ExtraSmallBorderWidth), 1)
	case 3, 4:
		return max(gtx.Dp(defaults.LargeBorderWidth), 1)
	default:
		return max(gtx.Dp(defaults.BorderWidth), 1)
	}
}

func colorSwatchPickerItemRadius(gtx layout.Context, defaults colorSwatchPickerDefaults, side int, size uint8, shape ColorSwatchShape) int {
	if shape == ColorSwatchCircle {
		return side / 2
	}
	radius := defaults.SquareItemRadius
	if size == 0 {
		radius = defaults.SquareItemRadiusExtraSmall
	} else if size == 1 {
		radius = defaults.SquareItemRadiusSmall
	}
	return min(max(gtx.Dp(radius), 1), side/2)
}

func colorSwatchPickerItemRadiusDp(defaults colorSwatchPickerDefaults, size uint8) unit.Dp {
	value := defaults.SquareItemRadius
	if size == 0 {
		value = defaults.SquareItemRadiusExtraSmall
	} else if size == 1 {
		value = defaults.SquareItemRadiusSmall
	}
	return value
}

func colorSwatchPickerSwatchRadius(gtx layout.Context, defaults colorSwatchPickerDefaults, side int, size uint8, shape ColorSwatchShape, selection float32) int {
	if shape == ColorSwatchCircle {
		return side / 2
	}
	radius := defaults.SquareSwatchRadius
	switch size {
	case 0:
		radius = defaults.SquareSwatchRadiusExtraSmall
	case 1:
		radius = unit.Dp(float32(defaults.SquareSwatchRadiusSmall) + (float32(defaults.SquareSelectedSmallRadius)-float32(defaults.SquareSwatchRadiusSmall))*selection)
	}
	return min(max(gtx.Dp(radius), 1), side/2)
}

func colorDrawColorSwatchCheck(gtx layout.Context, size image.Point, value color.NRGBA, progress, strokeWidth float32) {
	checkSize := float32(min(size.X, size.Y)) / 3
	x := (float32(size.X) - checkSize) / 2
	y := (float32(size.Y) - checkSize) / 2
	points := [3]f32.Point{
		f32.Pt(x+checkSize*.05, y+checkSize*.56),
		f32.Pt(x+checkSize*.40, y+checkSize*.86),
		f32.Pt(x+checkSize*.95, y+checkSize*.14),
	}
	path := draw.CheckPath(gtx.Ops, points, progress)
	foreground := color.NRGBA{R: 255, G: 255, B: 255, A: byte(255*progress + .5)}
	if colorColorLuminance(value) > .5 {
		foreground.R, foreground.G, foreground.B = 0, 0, 0
	}
	stroke := clip.Stroke{Path: path, Width: max(strokeWidth, 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
}

func colorColorLuminance(value color.NRGBA) float64 {
	return (.2126*float64(value.R) + .7152*float64(value.G) + .0722*float64(value.B)) / 255
}

func colorDrawColorArea(gtx layout.Context, size image.Point, value hsvColor, radius, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA, showDots bool, dotSize, dotGap int) {
	rect := image.Rectangle{Max: size}
	hue := colorHsvToNRGBA(hsvColor{h: value.h, s: 1, v: 1, a: 1})
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, hue)
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), color.NRGBA{R: 255, G: 255, B: 255, A: 255}, f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)), color.NRGBA{R: 255, G: 255, B: 255})
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), color.NRGBA{}, f32.Pt(float32(rect.Min.X), float32(rect.Max.Y)), color.NRGBA{A: 255})
	clipped.Pop()
	if showDots {
		colorDrawColorAreaDots(gtx, rect, radius, dotSize, dotGap)
	}
	colorDrawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})
	center := image.Pt(int(float64(max(size.X-1, 0))*value.s+0.5), int(float64(max(size.Y-1, 0))*(1-value.v)+0.5))
	thumbColor := colorHsvToNRGBA(hsvColor{h: value.h, s: value.s, v: value.v, a: 1})
	colorDrawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func colorDrawColorAreaDots(gtx layout.Context, rect image.Rectangle, radius, size, gap int) {
	if rect.Empty() || size <= 0 || gap <= 0 {
		return
	}
	clipped := clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Push(gtx.Ops)
	defer clipped.Pop()
	value := color.NRGBA{R: 255, G: 255, B: 255, A: 51}
	for y := rect.Min.Y + gap/2; y < rect.Max.Y; y += gap {
		for x := rect.Min.X + gap/2; x < rect.Max.X; x += gap {
			colorDrawCircle(gtx, image.Pt(x, y), size, value)
		}
	}
}

func colorDrawHueSlider(gtx layout.Context, size image.Point, value hsvColor, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	colorDrawHueSliderTrack(gtx, rect, min(size.X, size.Y)/2)
	colorDrawRoundedStroke(gtx, rect, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	center := colorColorSliderCenter(size, value.h)
	thumbColor := colorHsvToNRGBA(hsvColor{h: value.h, s: 1, v: 1, a: 1})
	colorDrawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func colorDrawAlphaSlider(gtx layout.Context, size image.Point, value color.NRGBA, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	colorDrawCheckerboard(gtx, rect, radius)
	transparent := value
	transparent.A = 0
	opaque := value
	opaque.A = 255
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	draw.PaintLinearGradient(gtx, f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)), transparent, f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)), opaque)
	clipped.Pop()
	colorDrawRoundedStroke(gtx, rect, radius, 1, color.NRGBA{A: 26})
	center := colorColorSliderCenter(size, float64(value.A)/255)
	colorDrawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, value, focus, focusColor)
}

func colorColorSliderCenter(size image.Point, value float64) image.Point {
	edge := float64(min(size.X, size.Y)) / 2
	length := max(float64(size.X)-edge*2, 0)
	return image.Pt(int(edge+colorClampUnit(value)*length+0.5), size.Y/2)
}

func colorDrawColorThumb(gtx layout.Context, center image.Point, size, border, focusWidth int, value color.NRGBA, focus float32, focusColor color.NRGBA) {
	if size <= 0 {
		return
	}
	if focus > 0 {
		col := focusColor
		col.A = byte(float32(col.A)*focus + .5)
		colorDrawCircle(gtx, center, size+focusWidth*4, col)
		colorDrawCircle(gtx, center, size+focusWidth*2, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	}
	colorDrawCircle(gtx, center, size+2, color.NRGBA{A: 38})
	colorDrawCircle(gtx, center, size, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	colorDrawCircle(gtx, center, max(size-border*2, 1), value)
}

func colorDrawCircle(gtx layout.Context, center image.Point, diameter int, value color.NRGBA) {
	if diameter <= 0 || value.A == 0 {
		return
	}
	half := diameter / 2
	rect := image.Rect(center.X-half, center.Y-half, center.X-half+diameter, center.Y-half+diameter)
	paint.FillShape(gtx.Ops, value, clip.Ellipse(rect).Op(gtx.Ops))
}

func colorDrawRoundedBorder(gtx layout.Context, rect image.Rectangle, radius, width int, border, background color.NRGBA) {
	if rect.Empty() || width <= 0 || border.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, border, clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(inner, min(max(radius-width, 0), min(inner.Dx(), inner.Dy())/2)).Op(gtx.Ops))
}

func colorDrawRoundedStroke(gtx layout.Context, rect image.Rectangle, radius, width int, value color.NRGBA) {
	if rect.Empty() || width <= 0 || value.A == 0 {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, value)
	stroke.Pop()
}

func colorDrawCheckerboard(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() {
		return
	}
	clipped := clip.UniformRRect(rect, min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)).Push(gtx.Ops)
	defer clipped.Pop()
	cell := max(min(rect.Dx(), rect.Dy())/4, 2)
	colors := [2]color.NRGBA{{R: 0xef, G: 0xef, B: 0xef, A: 255}, {R: 0xf7, G: 0xf7, B: 0xf7, A: 255}}
	for y, row := rect.Min.Y, 0; y < rect.Max.Y; y, row = y+cell, row+1 {
		for x, column := rect.Min.X, 0; x < rect.Max.X; x, column = x+cell, column+1 {
			tile := image.Rect(x, y, min(x+cell, rect.Max.X), min(y+cell, rect.Max.Y))
			paint.FillShape(gtx.Ops, colors[(row+column)%2], clip.Rect(tile).Op())
		}
	}
}

func colorDrawHueSliderTrack(gtx layout.Context, rect image.Rectangle, radius int) {
	if rect.Empty() {
		return
	}
	imageOp := colorHueSliderImage()
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	scale := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(float32(rect.Dx())/float32(imageOp.Size().X), float32(rect.Dy())/float32(imageOp.Size().Y)))).Push(gtx.Ops)
	imageOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	scale.Pop()
	clipped.Pop()
}

var colorHueSliderImageOp = colorNewHueSliderImage()

func colorHueSliderImage() paint.ImageOp {
	return colorHueSliderImageOp
}

func colorNewHueSliderImage() paint.ImageOp {
	const width = 360
	value := image.NewNRGBA(image.Rect(0, 0, width, 1))
	for x := range width {
		value.SetNRGBA(x, 0, colorHsvToNRGBA(hsvColor{h: float64(x) / (width - 1), s: 1, v: 1, a: 1}))
	}
	return paint.NewImageOp(value)
}

func colorDrawColorWheel(gtx layout.Context, size image.Point, value hsvColor, cache *colorWheelImageCache, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if cache.size != size || cache.op.Size() != size {
		cache.size = size
		cache.op = paint.NewImageOp(colorBuildColorWheelImage(size))
	}
	clipped := clip.Ellipse(image.Rectangle{Max: size}).Push(gtx.Ops)
	cache.op.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	clipped.Pop()
	colorDrawRoundedStroke(gtx, image.Rectangle{Max: size}, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	center := colorColorWheelSelectorCenter(size, value, (thumbSize+focusWidth*4+1)/2)
	thumbColor := colorHsvToNRGBA(hsvColor{h: value.h, s: value.s, v: value.v, a: 1})
	colorDrawColorThumb(gtx, center, thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func colorColorWheelSelectorCenter(size image.Point, value hsvColor, inset int) image.Point {
	radius := max(float64(min(size.X, size.Y))/2-float64(max(inset, 0)), 0)
	angle := colorWrapUnit(value.h) * 2 * math.Pi
	centerX := float64(size.X) / 2
	centerY := float64(size.Y) / 2
	distance := radius * colorClampUnit(value.s)
	return image.Pt(int(math.Round(centerX+math.Cos(angle)*distance)), int(math.Round(centerY+math.Sin(angle)*distance)))
}

func colorBuildColorWheelImage(size image.Point) *image.NRGBA {
	value := image.NewNRGBA(image.Rectangle{Max: size})
	centerX := float64(size.X-1) / 2
	centerY := float64(size.Y-1) / 2
	radius := float64(min(size.X, size.Y)) / 2
	if radius <= 0 {
		return value
	}
	for y := range size.Y {
		for x := range size.X {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Hypot(dx, dy)
			coverage := min(max(radius+.5-distance, 0), 1)
			if coverage == 0 {
				continue
			}
			pixel := colorHsvToNRGBA(hsvColor{h: colorWrapUnit(math.Atan2(dy, dx) / (2 * math.Pi)), s: colorClampUnit(distance / radius), v: 1, a: coverage})
			value.SetNRGBA(x, y, pixel)
		}
	}
	return value
}

func colorHsvToNRGBA(value hsvColor) color.NRGBA {
	hue := value.h - math.Floor(value.h)
	saturation := colorClampUnit(value.s)
	brightness := colorClampUnit(value.v)
	chroma := brightness * saturation
	section := hue * 6
	x := chroma * (1 - math.Abs(math.Mod(section, 2)-1))
	var r, g, b float64
	switch int(section) % 6 {
	case 0:
		r, g = chroma, x
	case 1:
		r, g = x, chroma
	case 2:
		g, b = chroma, x
	case 3:
		g, b = x, chroma
	case 4:
		r, b = x, chroma
	default:
		r, b = chroma, x
	}
	match := brightness - chroma
	channel := func(value float64) uint8 { return uint8(colorClampUnit(value+match)*255 + .5) }
	return color.NRGBA{R: channel(r), G: channel(g), B: channel(b), A: uint8(colorClampUnit(value.a)*255 + .5)}
}

func colorClampUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(max(value, 0), 1)
}

func colorWrapUnit(value float64) float64 {
	value -= math.Floor(value)
	if value < 0 {
		value++
	}
	return value
}

func colorSwatchPickerVisualSize(item image.Point, borderWidth int, scale float32) image.Point {
	content := image.Rectangle{Max: item}.Inset(max(borderWidth, 0)).Size()
	if content.X <= 0 || content.Y <= 0 {
		return image.Point{}
	}
	return image.Pt(max(int(float32(content.X)*scale+.5), 1), max(int(float32(content.Y)*scale+.5), 1))
}
