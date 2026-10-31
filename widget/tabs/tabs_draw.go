package tabs

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/widget/icon"
)

// tabsDrawIndicator paints the selected-tab indicator inside the supplied
// local operation stream.
func tabsDrawIndicator(
	gtx layout.Context,
	geometry tabsGeometry,
	rect image.Rectangle,
	orientation Orientation,
	variant Variant,
	col color.NRGBA,
) {
	if rect.Empty() || col.A == 0 {
		return
	}
	if variant == Secondary {
		width := max(gtx.Dp(geometry.indicatorLineWidth), 1)
		if orientation == Vertical {
			bounds := image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)
			paint.FillShape(gtx.Ops, col, clip.Rect(bounds).Op())
			return
		}
		bounds := image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y)
		paint.FillShape(gtx.Ops, col, clip.Rect(bounds).Op())
		return
	}
	radius := min(max(gtx.Dp(geometry.indicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

// tabsDrawFocus paints the focus ring without owning focus state.
func tabsDrawFocus(
	gtx layout.Context,
	geometry tabsGeometry,
	size image.Point,
	variant Variant,
	opacity float32,
	col color.NRGBA,
) {
	if opacity <= 0 || col.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	width := max(gtx.Dp(geometry.focusRingWidth), 1)
	rect := image.Rectangle{Max: size}.Inset(max(width/2, 1))
	radius := 0
	if variant != Secondary {
		radius = min(max(gtx.Dp(geometry.indicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
	}
	col.A = byte(float32(col.A)*opacity + 0.5)
	draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, radius, width, max(width/2, 1), col)
}

// tabsDrawSeparator paints the optional divider between neighboring tabs.
func tabsDrawSeparator(
	gtx layout.Context,
	geometry tabsGeometry,
	size image.Point,
	orientation Orientation,
	col color.NRGBA,
) {
	if col.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	width := max(gtx.Dp(geometry.separatorWidth), 1)
	if orientation == Vertical {
		inset := max(size.X/20, 1)
		bounds := image.Rect(inset, 0, max(size.X-inset, inset), width)
		paint.FillShape(gtx.Ops, col, clip.Rect(bounds).Op())
		return
	}
	inset := size.Y / 4
	bounds := image.Rect(0, inset, width, max(size.Y-inset, inset))
	paint.FillShape(gtx.Ops, col, clip.Rect(bounds).Op())
}

// tabsDrawList paints the tab list surface and, for the secondary variant, its
// divider.
func tabsDrawList(
	gtx layout.Context,
	geometry tabsGeometry,
	size image.Point,
	orientation Orientation,
	variant Variant,
	background, border color.NRGBA,
) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if variant == Secondary {
		width := max(gtx.Dp(geometry.separatorWidth), 1)
		if orientation == Vertical {
			paint.FillShape(gtx.Ops, border, clip.Rect(image.Rect(0, 0, width, size.Y)).Op())
			return
		}
		paint.FillShape(gtx.Ops, border, clip.Rect(image.Rect(0, size.Y-width, size.X, size.Y)).Op())
		return
	}
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(geometry.listRadius), 1), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

type tabsScrollShadow struct {
	bounds image.Rectangle
	stop1  f32.Point
	color1 color.NRGBA
	stop2  f32.Point
	color2 color.NRGBA
}

func tabsScrollShadowFor(
	size image.Point,
	orientation Orientation,
	length, direction int,
	background color.NRGBA,
) (tabsScrollShadow, bool) {
	mainSize := size.X
	if orientation == Vertical {
		mainSize = size.Y
	}
	length = min(max(length, 0), mainSize/2)
	if length == 0 || direction == 0 || size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return tabsScrollShadow{}, false
	}
	transparent := background
	transparent.A = 0
	if orientation == Vertical {
		if direction < 0 {
			return tabsScrollShadow{
				bounds: image.Rect(0, 0, size.X, length),
				stop1:  f32.Pt(0, 0), color1: background,
				stop2: f32.Pt(0, float32(length)), color2: transparent,
			}, true
		}
		return tabsScrollShadow{
			bounds: image.Rect(0, size.Y-length, size.X, size.Y),
			stop1:  f32.Pt(0, float32(size.Y-length)), color1: transparent,
			stop2: f32.Pt(0, float32(size.Y)), color2: background,
		}, true
	}
	if direction < 0 {
		return tabsScrollShadow{
			bounds: image.Rect(0, 0, length, size.Y),
			stop1:  f32.Pt(0, 0), color1: background,
			stop2: f32.Pt(float32(length), 0), color2: transparent,
		}, true
	}
	return tabsScrollShadow{
		bounds: image.Rect(size.X-length, 0, size.X, size.Y),
		stop1:  f32.Pt(float32(size.X-length), 0), color1: transparent,
		stop2: f32.Pt(float32(size.X), 0), color2: background,
	}, true
}

// tabsDrawScroll paints the overflow shadows and their optional edge border.
func tabsDrawScroll(
	gtx layout.Context,
	geometry tabsGeometry,
	size image.Point,
	orientation Orientation,
	variant Variant,
	previous, next bool,
	background, border color.NRGBA,
) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	length := gtx.Dp(geometry.scrollShadowSize)
	var shadowClip clip.Stack
	hasShadowClip := false
	if variant != Secondary {
		radius := min(max(gtx.Dp(geometry.listRadius), 1), min(size.X, size.Y)/2)
		shadowClip = clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
		hasShadowClip = true
	}
	edges := []struct {
		visible   bool
		direction int
	}{
		{visible: previous, direction: -1},
		{visible: next, direction: 1},
	}
	for _, edge := range edges {
		if !edge.visible {
			continue
		}
		spec, ok := tabsScrollShadowFor(size, orientation, length, edge.direction, background)
		if !ok {
			continue
		}
		stack := clip.Rect(spec.bounds).Push(gtx.Ops)
		draw.PaintLinearGradient(gtx, spec.stop1, spec.color1, spec.stop2, spec.color2)
		stack.Pop()
	}
	if hasShadowClip {
		shadowClip.Pop()
	}
	if variant != Secondary || border.A == 0 {
		return
	}
	width := max(gtx.Dp(geometry.separatorWidth), 1)
	for _, edge := range edges {
		if !edge.visible {
			continue
		}
		spec, ok := tabsScrollShadowFor(size, orientation, length, edge.direction, border)
		if !ok {
			continue
		}
		bounds := spec.bounds
		if orientation == Vertical {
			bounds.Max.X = min(bounds.Min.X+width, bounds.Max.X)
		} else {
			bounds.Min.Y = max(bounds.Max.Y-width, bounds.Min.Y)
		}
		paint.FillShape(gtx.Ops, border, clip.Rect(bounds).Op())
	}
}

// tabsDrawChevron paints an overflow button glyph.
func tabsDrawChevron(
	gtx layout.Context,
	geometry tabsGeometry,
	size image.Point,
	orientation Orientation,
	direction int,
	col color.NRGBA,
) {
	data := lucide.ChevronRight
	if orientation == Vertical {
		data = lucide.ChevronDown
		if direction < 0 {
			data = lucide.ChevronUp
		}
	} else if direction < 0 {
		data = lucide.ChevronLeft
	}
	diameter := max(
		gtx.Dp(geometry.scrollChevronSize*2),
		icon.LucideSizeForStroke(gtx, geometry.scrollChevronStroke),
	)
	diameter = min(diameter, min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(data, iconGtx, col)
	offset.Pop()
}

// tabsDrawClose paints the close glyph for a closable tab.
func tabsDrawClose(gtx layout.Context, size image.Point, col color.NRGBA) {
	diameter := min(size.X, size.Y)
	if diameter <= 0 {
		return
	}
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.X, iconGtx, col)
	offset.Pop()
}

// tabsDrawAdd paints the add-tab glyph.
func tabsDrawAdd(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	diameter := min(size.X, size.Y)
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.Plus, iconGtx, col)
	offset.Pop()
}

// tabsDrawOverflow paints the overflow-menu trigger glyph.
func tabsDrawOverflow(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	diameter := min(size.X, size.Y)
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.Ellipsis, iconGtx, col)
	offset.Pop()
}
