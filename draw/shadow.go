// SPDX-License-Identifier: Unlicense OR MIT

package draw

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
	"github.com/qianniancn/flowui/theme"
)

// ShadowShapeKind identifies the geometry used for a shadow.
type ShadowShapeKind uint8

const (
	// ShadowRoundedRect draws a rounded rectangle shadow.
	ShadowRoundedRect ShadowShapeKind = iota
	// ShadowEllipse draws an ellipse shadow.
	ShadowEllipse
)

// ShadowCornerRadii stores independent rounded-rectangle corner radii.
type ShadowCornerRadii struct {
	NW unit.Dp
	NE unit.Dp
	SE unit.Dp
	SW unit.Dp
}

// ShadowShape describes the surface casting a shadow.
type ShadowShape struct {
	Kind  ShadowShapeKind
	Radii ShadowCornerRadii
}

// RoundedShadowCorners returns a non-uniform rounded rectangle shadow shape.
func RoundedShadowCorners(nw, ne, se, sw unit.Dp) ShadowShape {
	return ShadowShape{
		Kind: ShadowRoundedRect,
		Radii: ShadowCornerRadii{
			NW: nw,
			NE: ne,
			SE: se,
			SW: sw,
		},
	}
}

// EllipseShadow returns an ellipse shadow shape.
func EllipseShadow() ShadowShape {
	return ShadowShape{Kind: ShadowEllipse}
}

func (s ShadowShape) pixels(gtx layout.Context, size image.Point) shadowShape {
	shape := shadowShape{Kind: s.Kind}
	if shape.Kind != ShadowEllipse {
		shape.Kind = ShadowRoundedRect
		shape.Radii = normalizeCornerRadii(image.Rectangle{Max: size}, cornerRadiiPx{
			NW: gtx.Dp(s.Radii.NW),
			NE: gtx.Dp(s.Radii.NE),
			SE: gtx.Dp(s.Radii.SE),
			SW: gtx.Dp(s.Radii.SW),
		})
	}
	return shape
}

// ClipOp returns a clip operation matching the shadow shape.
func (s ShadowShape) ClipOp(gtx layout.Context, rect image.Rectangle) clip.Op {
	shape := s.pixels(gtx, rect.Size())
	if shape.Kind == ShadowEllipse {
		return clip.Ellipse(rect).Op(gtx.Ops)
	}
	r := shape.Radii
	return clip.RRect{
		Rect: rect,
		NW:   r.NW,
		NE:   r.NE,
		SE:   r.SE,
		SW:   r.SW,
	}.Op(gtx.Ops)
}

// ShadowLayer is one soft shadow pass.
type ShadowLayer struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
}

// ShadowTheme is a theme-defined stack of shadow layers.
type ShadowTheme = theme.ShadowTheme

// BoxShadow describes a CSS-like shadow. Layers, when present, take precedence over
// the single shadow fields.
type BoxShadow struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
	Layers  []ShadowLayer
}

// EffectiveLayers returns the drawable layers for a box shadow.
func (b BoxShadow) EffectiveLayers() []ShadowLayer {
	if len(b.Layers) > 0 {
		layers := make([]ShadowLayer, 0, len(b.Layers))
		for _, layer := range b.Layers {
			if layer.Blur >= 0 && layer.Color.A > 0 {
				layers = append(layers, layer)
			}
		}
		return layers
	}
	if b.Blur < 0 || b.Color.A == 0 {
		return nil
	}
	return []ShadowLayer{{
		OffsetX: b.OffsetX,
		OffsetY: b.OffsetY,
		Blur:    b.Blur,
		Spread:  b.Spread,
		Color:   b.Color,
	}}
}

func (b BoxShadow) forEachLayer(fn func(ShadowLayer)) {
	if len(b.Layers) > 0 {
		for _, layer := range b.Layers {
			if layer.Blur >= 0 && layer.Color.A > 0 {
				fn(layer)
			}
		}
		return
	}
	if b.Blur < 0 || b.Color.A == 0 {
		return
	}
	fn(ShadowLayer{
		OffsetX: b.OffsetX,
		OffsetY: b.OffsetY,
		Blur:    b.Blur,
		Spread:  b.Spread,
		Color:   b.Color,
	})
}

// ThemeShadow resolves theme-controlled layers against a base color.
func ThemeShadow(style ShadowTheme, col color.NRGBA, opacity float32) BoxShadow {
	opacity = shadowOpacity(opacity)
	layers := make([]ShadowLayer, 0, len(style.Layers))
	for _, token := range style.Layers {
		layerOpacity := shadowOpacity(token.Opacity) * opacity
		layerColor := col
		layerColor.A = scaleAlpha(col.A, layerOpacity)
		if token.Blur < 0 || layerColor.A == 0 {
			continue
		}
		layers = append(layers, ShadowLayer{
			OffsetX: float32(token.OffsetX),
			OffsetY: float32(token.OffsetY),
			Blur:    float32(token.Blur),
			Spread:  float32(token.Spread),
			Color:   layerColor,
		})
	}
	return BoxShadow{Blur: -1, Layers: layers}
}

func shadowOpacity(value float32) float32 {
	if value <= 0 || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return 0
	}
	return min(value, 1)
}

// Shadow paints a cached soft shadow for bounds. The bounds rectangle may be
// offset; the shadow is drawn around that rectangle in the current operation
// coordinate space.
func Shadow(gtx layout.Context, bounds image.Rectangle, shape ShadowShape, box BoxShadow) {
	size := bounds.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	drawShadow(gtx, bounds, shape.pixels(gtx, size), box)
}

func drawShadow(gtx layout.Context, bounds image.Rectangle, pxShape shadowShape, box BoxShadow) {
	size := bounds.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if bounds.Min != (image.Point{}) {
		stack := op.Offset(bounds.Min).Push(gtx.Ops)
		defer stack.Pop()
	}

	box.forEachLayer(func(layer ShadowLayer) {
		blur := gtx.Dp(unit.Dp(layer.Blur))
		if blur < 0 || layer.Color.A == 0 {
			return
		}
		opacity := float32(layer.Color.A) / 255
		cacheColor := layer.Color
		cacheColor.A = 255
		entry := softShadowEntry(
			size,
			pxShape,
			blur,
			gtx.Dp(unit.Dp(layer.Spread)),
			gtx.Dp(unit.Dp(layer.OffsetX)),
			gtx.Dp(unit.Dp(layer.OffsetY)),
			cacheColor,
		)
		if entry.op.Size() == (image.Point{}) {
			return
		}
		stack := op.Offset(image.Pt(-entry.padX, -entry.padY)).Push(gtx.Ops)
		clipStack := clip.Rect(image.Rectangle{Max: entry.size}).Push(gtx.Ops)
		var scaleStack op.TransformStack
		if entry.scale > 1 {
			scale := float32(entry.scale)
			scaleStack = op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale))).Push(gtx.Ops)
		}
		var opacityStack paint.OpacityStack
		if opacity < 1 {
			opacityStack = paint.PushOpacity(gtx.Ops, opacity)
		}
		entry.op.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		if opacity < 1 {
			opacityStack.Pop()
		}
		if entry.scale > 1 {
			scaleStack.Pop()
		}
		clipStack.Pop()
		stack.Pop()
	})
}

// Surface paints a rounded surface with the supplied shadow. It is the
// themeable form used by higher-level widgets.
func Surface(gtx layout.Context, rect image.Rectangle, radius int, surface color.NRGBA, shadow BoxShadow) {
	radius = max(radius, 0)
	shape := shadowShape{
		Kind: ShadowRoundedRect,
		Radii: cornerRadiiPx{
			NW: radius,
			NE: radius,
			SE: radius,
			SW: radius,
		},
	}
	drawShadow(gtx, rect, shape, shadow)
	paint.FillShape(gtx.Ops, surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func scaleAlpha(a uint8, scale float32) uint8 {
	return clampByte(float32(a) * scale)
}
