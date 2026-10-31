package divider

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

// Widget draws a horizontal or vertical separator line.
type Widget struct {
	thickness unit.Dp
	color     color.NRGBA
	hasColor  bool
	axis      layout.Axis
}

// New creates a horizontal separator line.
func New() Widget {
	return Widget{}
}

// Vertical changes the divider orientation to vertical.
func (d Widget) Vertical() Widget {
	d.axis = layout.Vertical
	return d
}

// Thickness sets the divider thickness in density-independent pixels.
func (d Widget) Thickness(dp int) Widget {
	d.thickness = unit.Dp(dp)
	return d
}

// Color sets the divider color.
func (d Widget) Color(c color.NRGBA) Widget {
	d.color = c
	d.hasColor = true
	return d
}

func (d Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	thickness := lineThickness(gtx, d.thickness)
	size := image.Pt(gtx.Constraints.Max.X, thickness)
	if d.axis == layout.Vertical {
		size = image.Pt(thickness, gtx.Constraints.Max.Y)
	}
	size = gtx.Constraints.Constrain(size)
	drawLine(ctx, gtx, size, d.color, d.hasColor)
	return layout.Dimensions{Size: size}
}
func lineThickness(gtx layout.Context, thickness unit.Dp) int {
	if thickness == 0 {
		thickness = 1
	}
	return max(gtx.Dp(thickness), 1)
}

func drawLine(ctx *core.Context, gtx layout.Context, size image.Point, c color.NRGBA, hasColor bool) {
	if !hasColor {
		c = core.ActiveTheme(ctx).Semantic().Foreground
		c.A = 0x33
	}
	paint.FillShape(gtx.Ops, c, clip.Rect{Max: size}.Op())
}
