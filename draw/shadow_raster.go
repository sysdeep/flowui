// SPDX-License-Identifier: Unlicense OR MIT

package draw

import (
	"image"
	"image/color"
	"math"
	"sync"

	"gioui.org/op/paint"
)

const (
	maxSoftShadowBlur     = 96
	maxShadowRasterPixels = 256 * 1024
	maxShadowRasterScale  = 8
)

type cornerRadiiPx struct {
	NW int
	NE int
	SE int
	SW int
}

type shadowShape struct {
	Kind  ShadowShapeKind
	Radii cornerRadiiPx
}

var byteSlicePool = sync.Pool{
	New: func() any {
		buf := make([]uint8, 0, 128*1024)
		return &buf
	},
}

func buildSoftShadowEntry(size image.Point, shape shadowShape, blur, spread, offX, offY int, col color.NRGBA) softShadowCacheEntry {
	if blur < 0 {
		return softShadowCacheEntry{}
	}
	blur = min(blur, maxSoftShadowBlur)
	padX := ShadowRasterPadding(blur, spread, offX)
	padY := ShadowRasterPadding(blur, spread, offY)
	rasterSize, ok := shadowRasterSize(size, padX, padY)
	if !ok {
		return softShadowCacheEntry{}
	}
	w := rasterSize.X
	h := rasterSize.Y
	if w <= 0 || h <= 0 {
		return softShadowCacheEntry{}
	}

	scale := shadowRasterScale(blur, w, h)
	lw := ceilDiv(w, scale)
	lh := ceilDiv(h, scale)
	if lw <= 0 || lh <= 0 {
		return softShadowCacheEntry{}
	}
	if lw > maxShadowRasterPixels/lh {
		return softShadowCacheEntry{}
	}

	alpha := borrowByteSlice(lw * lh)
	defer releaseByteSlice(alpha)

	shapeRect := image.Rect(
		padX+offX-spread,
		padY+offY-spread,
		padX+offX+size.X+spread,
		padY+offY+size.Y+spread,
	)
	lowShapeRect := scaleRectDown(shapeRect, scale)
	lowCutRect := scaleRectDown(image.Rect(padX, padY, padX+size.X, padY+size.Y), scale)
	lowShape := scaleShapeDown(shapeWithSpread(shape, spread), scale)
	lowCutShape := scaleShapeDown(shape, scale)
	drawShapeMask(alpha, lw, lh, lowShapeRect, lowShape, col.A)
	boxBlurAlpha(alpha, lw, lh, ceilDiv(blur, scale))
	cutShapeMask(alpha, lw, lh, lowCutRect, lowCutShape)

	img := image.NewRGBA(image.Rect(0, 0, lw, lh))
	for y := range lh {
		row := y * img.Stride
		for x := range lw {
			a := rasterAlpha(alpha[y*lw+x], scale, x, y)
			if a == 0 {
				continue
			}
			i := row + x*4
			img.Pix[i+0] = premul(col.R, a)
			img.Pix[i+1] = premul(col.G, a)
			img.Pix[i+2] = premul(col.B, a)
			img.Pix[i+3] = a
		}
	}
	op := paint.NewImageOp(img)
	op.Filter = paint.FilterLinear
	return softShadowCacheEntry{op: op, padX: padX, padY: padY, size: image.Pt(w, h), scale: scale, bytes: rgbaBytes(lw, lh)}
}

func shadowRasterScale(blur, w, h int) int {
	scale := 1
	switch {
	case blur >= 48 && pixelsAtLeast(w, h, 256*256):
		scale = 4
	case blur >= 16 && pixelsAtLeast(w, h, 128*128):
		scale = 2
	}
	for pixelsExceed(ceilDiv(w, scale), ceilDiv(h, scale), maxShadowRasterPixels) && scale < maxShadowRasterScale {
		scale *= 2
	}
	return scale
}

func shadowRasterSize(size image.Point, padX, padY int) (image.Point, bool) {
	if size.X <= 0 || size.Y <= 0 || padX < 0 || padY < 0 {
		return image.Point{}, false
	}
	maxInt := maxIntValue()
	if padX > (maxInt-size.X)/2 || padY > (maxInt-size.Y)/2 {
		return image.Point{}, false
	}
	return image.Pt(size.X+padX*2, size.Y+padY*2), true
}

func pixelsAtLeast(w, h, threshold int) bool {
	if threshold <= 0 {
		return true
	}
	if w <= 0 || h <= 0 {
		return false
	}
	return w > (threshold-1)/h
}

func pixelsExceed(w, h, limit int) bool {
	if limit < 0 {
		return true
	}
	if w <= 0 || h <= 0 {
		return false
	}
	return w > limit/h
}

func rgbaBytes(w, h int) int {
	if w <= 0 || h <= 0 {
		return 0
	}
	maxInt := maxIntValue()
	if w > maxInt/h/4 {
		return maxInt
	}
	return w * h * 4
}

func scaleRectDown(rect image.Rectangle, scale int) image.Rectangle {
	if scale <= 1 {
		return rect
	}
	return image.Rect(
		floorDiv(rect.Min.X, scale),
		floorDiv(rect.Min.Y, scale),
		ceilDiv(rect.Max.X, scale),
		ceilDiv(rect.Max.Y, scale),
	)
}

func scaleShapeDown(shape shadowShape, scale int) shadowShape {
	if scale <= 1 || shape.Kind == ShadowEllipse {
		return shape
	}
	shape.Radii.NW = ceilDiv(shape.Radii.NW, scale)
	shape.Radii.NE = ceilDiv(shape.Radii.NE, scale)
	shape.Radii.SE = ceilDiv(shape.Radii.SE, scale)
	shape.Radii.SW = ceilDiv(shape.Radii.SW, scale)
	return shape
}

func drawShapeMask(alpha []uint8, w, h int, rect image.Rectangle, shape shadowShape, a uint8) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || a == 0 {
		return
	}
	shape = normalizeShadowShape(rect, shape)
	if shape.Kind != ShadowEllipse {
		drawRoundedShapeMask(alpha, w, h, rect, shape.Radii, a)
		return
	}
	for y := max(rect.Min.Y, 0); y < min(rect.Max.Y, h); y++ {
		for x := max(rect.Min.X, 0); x < min(rect.Max.X, w); x++ {
			coverage := shapeCoverageNormalized(x, y, rect, shape)
			alpha[y*w+x] = uint8(float32(a)*coverage + .5)
		}
	}
}

func cutShapeMask(alpha []uint8, w, h int, rect image.Rectangle, shape shadowShape) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return
	}
	shape = normalizeShadowShape(rect, shape)
	if shape.Kind != ShadowEllipse {
		cutRoundedShapeMask(alpha, w, h, rect, shape.Radii)
		return
	}
	for y := max(rect.Min.Y-1, 0); y < min(rect.Max.Y+1, h); y++ {
		for x := max(rect.Min.X-1, 0); x < min(rect.Max.X+1, w); x++ {
			coverage := shapeCoverageNormalized(x, y, rect, shape)
			if coverage <= 0 {
				continue
			}
			i := y*w + x
			alpha[i] = uint8(float32(alpha[i])*(1-coverage) + .5)
		}
	}
}

func drawRoundedShapeMask(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	fillRoundedShapeRects(alpha, w, h, rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+radii.NW, rect.Min.Y+radii.NW), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.NE, rect.Min.Y, rect.Max.X, rect.Min.Y+radii.NE), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.SE, rect.Max.Y-radii.SE, rect.Max.X, rect.Max.Y), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Max.Y-radii.SW, rect.Min.X+radii.SW, rect.Max.Y), rect, radii, a)
}

func cutRoundedShapeMask(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx) {
	fillRoundedShapeRects(alpha, w, h, rect, radii, 0)
	cutCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+radii.NW, rect.Min.Y+radii.NW), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.NE, rect.Min.Y, rect.Max.X, rect.Min.Y+radii.NE), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.SE, rect.Max.Y-radii.SE, rect.Max.X, rect.Max.Y), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Max.Y-radii.SW, rect.Min.X+radii.SW, rect.Max.Y), rect, radii)
}

func fillRoundedShapeRects(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	for y := max(rect.Min.Y, 0); y < min(rect.Max.Y, h); y++ {
		leftInset := 0
		if y < rect.Min.Y+radii.NW {
			leftInset = radii.NW
		} else if y >= rect.Max.Y-radii.SW {
			leftInset = radii.SW
		}
		rightInset := 0
		if y < rect.Min.Y+radii.NE {
			rightInset = radii.NE
		} else if y >= rect.Max.Y-radii.SE {
			rightInset = radii.SE
		}
		minX := max(rect.Min.X+leftInset, 0)
		maxX := min(rect.Max.X-rightInset, w)
		row := y * w
		for x := minX; x < maxX; x++ {
			alpha[row+x] = a
		}
	}
}

func drawCornerMask(alpha []uint8, w, h int, corner, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	for y := max(corner.Min.Y, 0); y < min(corner.Max.Y, h); y++ {
		for x := max(corner.Min.X, 0); x < min(corner.Max.X, w); x++ {
			coverage := roundedRectCoverage(x, y, rect, radii)
			alpha[y*w+x] = uint8(float32(a)*coverage + .5)
		}
	}
}

func cutCornerMask(alpha []uint8, w, h int, corner, rect image.Rectangle, radii cornerRadiiPx) {
	for y := max(corner.Min.Y, 0); y < min(corner.Max.Y, h); y++ {
		for x := max(corner.Min.X, 0); x < min(corner.Max.X, w); x++ {
			coverage := roundedRectCoverage(x, y, rect, radii)
			if coverage <= 0 {
				continue
			}
			i := y*w + x
			alpha[i] = uint8(float32(alpha[i])*(1-coverage) + .5)
		}
	}
}

func shapeCoverage(x, y int, rect image.Rectangle, shape shadowShape) float32 {
	shape = normalizeShadowShape(rect, shape)
	return shapeCoverageNormalized(x, y, rect, shape)
}

func shapeCoverageNormalized(x, y int, rect image.Rectangle, shape shadowShape) float32 {
	if shape.Kind == ShadowEllipse {
		return ellipseCoverage(x, y, rect)
	}
	return roundedRectCoverage(x, y, rect, shape.Radii)
}

func roundedRectCoverage(x, y int, rect image.Rectangle, radii cornerRadiiPx) float32 {
	if rect.Empty() || x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return 0
	}

	var cx, cy float64
	var radius int
	switch {
	case x < rect.Min.X+radii.NW && y < rect.Min.Y+radii.NW:
		radius = radii.NW
		cx = float64(rect.Min.X + radius)
		cy = float64(rect.Min.Y + radius)
	case x >= rect.Max.X-radii.NE && y < rect.Min.Y+radii.NE:
		radius = radii.NE
		cx = float64(rect.Max.X - radius)
		cy = float64(rect.Min.Y + radius)
	case x >= rect.Max.X-radii.SE && y >= rect.Max.Y-radii.SE:
		radius = radii.SE
		cx = float64(rect.Max.X - radius)
		cy = float64(rect.Max.Y - radius)
	case x < rect.Min.X+radii.SW && y >= rect.Max.Y-radii.SW:
		radius = radii.SW
		cx = float64(rect.Min.X + radius)
		cy = float64(rect.Max.Y - radius)
	default:
		return 1
	}
	if radius <= 0 {
		return 1
	}

	return circlePixelCoverage(x, y, cx, cy, float64(radius))
}

const aaGrid = 4

func circlePixelCoverage(x, y int, cx, cy, radius float64) float32 {
	px := float64(x) + 0.5
	py := float64(y) + 0.5
	dist := math.Hypot(px-cx, py-cy)
	edge := radius - dist
	switch {
	case edge >= 0.75:
		return 1
	case edge <= -0.75:
		return 0
	default:
		r2 := radius * radius
		inside := 0
		for sy := range aaGrid {
			fy := float64(y) + (float64(sy)+0.5)/aaGrid
			dy := fy - cy
			for sx := range aaGrid {
				fx := float64(x) + (float64(sx)+0.5)/aaGrid
				dx := fx - cx
				if dx*dx+dy*dy <= r2 {
					inside++
				}
			}
		}
		return float32(inside) / float32(aaGrid*aaGrid)
	}
}

func ellipseCoverage(x, y int, rect image.Rectangle) float32 {
	if rect.Empty() || x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return 0
	}
	rx := float64(rect.Dx()) / 2
	ry := float64(rect.Dy()) / 2
	if rx <= 0 || ry <= 0 {
		return 0
	}
	cx := float64(rect.Min.X+rect.Max.X) / 2
	cy := float64(rect.Min.Y+rect.Max.Y) / 2
	dx := (float64(x) + 0.5 - cx) / rx
	dy := (float64(y) + 0.5 - cy) / ry
	dist := math.Hypot(dx, dy)
	edge := (1 - dist) * math.Min(rx, ry)
	switch {
	case edge >= 0.75:
		return 1
	case edge <= -0.75:
		return 0
	default:
		inside := 0
		for sy := range aaGrid {
			fy := float64(y) + (float64(sy)+0.5)/aaGrid
			ny := (fy - cy) / ry
			for sx := range aaGrid {
				fx := float64(x) + (float64(sx)+0.5)/aaGrid
				nx := (fx - cx) / rx
				if nx*nx+ny*ny <= 1 {
					inside++
				}
			}
		}
		return float32(inside) / float32(aaGrid*aaGrid)
	}
}

func shapeWithSpread(shape shadowShape, spread int) shadowShape {
	if shape.Kind == ShadowEllipse {
		return shape
	}
	shape.Radii.NW = max(0, shape.Radii.NW+spread)
	shape.Radii.NE = max(0, shape.Radii.NE+spread)
	shape.Radii.SE = max(0, shape.Radii.SE+spread)
	shape.Radii.SW = max(0, shape.Radii.SW+spread)
	return shape
}

func normalizeShadowShape(rect image.Rectangle, shape shadowShape) shadowShape {
	if shape.Kind == ShadowEllipse {
		return shadowShape{Kind: ShadowEllipse}
	}
	shape.Kind = ShadowRoundedRect
	shape.Radii = normalizeCornerRadii(rect, shape.Radii)
	return shape
}

func normalizeCornerRadii(rect image.Rectangle, radii cornerRadiiPx) cornerRadiiPx {
	radii.NW = max(radii.NW, 0)
	radii.NE = max(radii.NE, 0)
	radii.SE = max(radii.SE, 0)
	radii.SW = max(radii.SW, 0)

	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return cornerRadiiPx{}
	}

	scale := 1.0
	scale = min(scale, cornerScale(w, radii.NW+radii.NE))
	scale = min(scale, cornerScale(w, radii.SW+radii.SE))
	scale = min(scale, cornerScale(h, radii.NW+radii.SW))
	scale = min(scale, cornerScale(h, radii.NE+radii.SE))
	if scale < 1 {
		radii.NW = int(math.Floor(float64(radii.NW)*scale + .5))
		radii.NE = int(math.Floor(float64(radii.NE)*scale + .5))
		radii.SE = int(math.Floor(float64(radii.SE)*scale + .5))
		radii.SW = int(math.Floor(float64(radii.SW)*scale + .5))
	}
	return radii
}

func cornerScale(size, sum int) float64 {
	if sum <= 0 || sum <= size {
		return 1
	}
	return float64(size) / float64(sum)
}

func boxBlurAlpha(alpha []uint8, w, h, radius int) {
	if radius <= 0 || w <= 0 || h <= 0 {
		return
	}
	passes := 3
	boxRadius := max(int(math.Ceil(float64(radius)/2.0)), 1)
	tmp := borrowByteSlice(len(alpha))
	defer releaseByteSlice(tmp)
	for range passes {
		blurHorizontal(alpha, tmp, w, h, boxRadius)
		blurVertical(tmp, alpha, w, h, boxRadius)
	}
}

func borrowByteSlice(size int) []uint8 {
	bufp := byteSlicePool.Get().(*[]uint8)
	buf := *bufp
	if cap(buf) < size {
		buf = make([]uint8, size)
	} else {
		buf = buf[:size]
		clear(buf)
	}
	*bufp = nil
	return buf
}

func releaseByteSlice(buf []uint8) {
	if cap(buf) > 4*1024*1024 {
		return
	}
	buf = buf[:0]
	byteSlicePool.Put(&buf)
}

var bayer4 = [16]int{
	0, 8, 2, 10,
	12, 4, 14, 6,
	3, 11, 1, 9,
	15, 7, 13, 5,
}

func ditherAlpha(a uint8, x, y int) uint8 {
	if a == 0 || a == 255 {
		return a
	}
	rank := bayer4[(y&3)*4+(x&3)]
	threshold := rank - 8
	if rank >= 8 {
		threshold = rank - 7
	}
	if threshold < 0 && a <= uint8(-threshold) {
		return 0
	}
	if threshold > 0 && int(a)+threshold >= 255 {
		return 255
	}
	return uint8(int(a) + threshold)
}

func rasterAlpha(a uint8, scale, x, y int) uint8 {
	if scale > 1 {
		return a
	}
	return ditherAlpha(a, x, y)
}

func blurHorizontal(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for y := range h {
		sum := 0
		row := y * w
		for x := -r; x <= r; x++ {
			sum += int(src[row+clampInt(x, 0, w-1)])
		}
		for x := range w {
			dst[row+x] = uint8(sum / window)
			remove := clampInt(x-r, 0, w-1)
			add := clampInt(x+r+1, 0, w-1)
			sum += int(src[row+add]) - int(src[row+remove])
		}
	}
}

func blurVertical(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for x := range w {
		sum := 0
		for y := -r; y <= r; y++ {
			sum += int(src[clampInt(y, 0, h-1)*w+x])
		}
		for y := range h {
			dst[y*w+x] = uint8(sum / window)
			remove := clampInt(y-r, 0, h-1)
			add := clampInt(y+r+1, 0, h-1)
			sum += int(src[add*w+x]) - int(src[remove*w+x])
		}
	}
}

func premul(c, a uint8) uint8 {
	return uint8((uint16(c)*uint16(a) + 127) / 255)
}

// ShadowRasterPadding returns the symmetric pixel padding used by a soft
// shadow raster on one axis. It is shared with clipping containers so their
// reserved visual space matches the pixels Shadow may emit.
func ShadowRasterPadding(blur, spread, offset int) int {
	blur = min(max(blur, 0), maxSoftShadowBlur)
	return saturatedIntSum(blur, blur, absoluteInt(spread), absoluteInt(offset), 2)
}

func absoluteInt(value int) int {
	if value >= 0 {
		return value
	}
	if value == -value {
		return maxIntValue()
	}
	return -value
}

func saturatedIntSum(values ...int) int {
	limit := maxIntValue()
	result := 0
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if result > limit-value {
			return limit
		}
		result += value
	}
	return result
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}

func ceilDiv(v, d int) int {
	if d <= 0 {
		return v
	}
	if v >= 0 {
		return (v + d - 1) / d
	}
	return v / d
}

func floorDiv(v, d int) int {
	if d <= 0 {
		return v
	}
	if v >= 0 {
		return v / d
	}
	return -((-v + d - 1) / d)
}

func clampByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(math.Round(float64(v)))
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
