package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type ColorSwatchWidget struct {
	style style.Style
	value color.NRGBA
	size  ColorSwatchSize
	shape ColorSwatchShape
	alt   string
}

func ColorSwatch(value color.NRGBA) ColorSwatchWidget {
	return ColorSwatchWidget{value: value, size: ColorSwatchMedium, shape: ColorSwatchCircle}
}

func (swatch ColorSwatchWidget) Size(size ColorSwatchSize) ColorSwatchWidget {
	swatch.size = size
	return swatch
}

func (swatch ColorSwatchWidget) Shape(shape ColorSwatchShape) ColorSwatchWidget {
	swatch.shape = shape
	return swatch
}

func (swatch ColorSwatchWidget) Alt(alt string) ColorSwatchWidget {
	swatch.alt = alt
	return swatch
}

func (swatch ColorSwatchWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	defaults := standaloneColorSwatchDefaults(ctx)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		defaults.Defaults,
		style.Style{},
		style.Style{},
		swatch.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		side := min(max(colorSwatchPixels(gtx, defaults, swatch.size), 0), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
		size := gtx.Constraints.Constrain(image.Pt(side, side))
		if swatch.alt != "" {
			semantic.LabelOp(swatch.alt).Add(gtx.Ops)
		}
		radius := 0
		if swatch.shape == ColorSwatchCircle {
			radius = min(size.X, size.Y) / 2
		} else {
			radius = min(max(gtx.Dp(defaults.SquareRadius), 1), min(size.X, size.Y)/2)
		}
		colorDrawColorSwatch(gtx, size, swatch.value, radius)
		return layout.Dimensions{Size: size}
	}))
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w ColorSwatchWidget) Style(values ...style.Style) ColorSwatchWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
