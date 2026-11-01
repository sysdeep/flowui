package spinner

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	key         string
	style       style.Style
	color       Color
	size        Size
	label       string
	visualCache *spinnerVisualCache
}

func New() Widget {
	return Widget{key: core.AutoKey("spinner"), visualCache: &spinnerVisualCache{}}
}

// Key sets the stable identity used by style transitions.
func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

func (s Widget) Color(color Color) Widget {
	s.color = color
	return s
}

func (s Widget) Size(size Size) Widget {
	s.size = size
	return s
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	resolved := s.resolveStyle(ctx, gtx)
	return s.layoutResolved(ctx, gtx, resolved)
}

func (s Widget) layoutResolved(ctx *core.Context, gtx layout.Context, resolved spinnerResolvedStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)
			stack := op.Offset(offset).Push(gtx.Ops)
			period := theme.ResolveMotionDuration(resolved.motion, resolved.metrics.SpinDuration)
			spinnerDraw(
				gtx,
				gtx.Metric.PxToDp(iconSize),
				resolved.metrics.StrokeRatio,
				resolved.metrics.InsetRatio,
				resolved.color,
				period,
			)
			if period > 0 {
				gtx.Execute(op.InvalidateCmd{})
			}
			stack.Pop()
		}
		clipStack := clip.Rect{Max: size}.Push(gtx.Ops)
		label := s.label
		if label == "" {
			label = ctx.Text(i18n.SpinnerLoading)
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		clipStack.Pop()
		return layout.Dimensions{Size: size}
	}))
}

// Style merges instance declarations into the widget. The declaration is
// applied as the last cascade layer when the spinner resolves its style, so
// Style can sit anywhere in the chain and does not invalidate cached defaults.
func (s Widget) Style(values ...style.Style) Widget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}
