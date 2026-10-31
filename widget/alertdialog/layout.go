package alertdialog

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
)

type dialogHeader struct {
	title       string
	description string
	status      Status
	icon        core.Widget
	cache       *alertDialogDefaultsCache
	style       style.Style
}

func (h dialogHeader) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	defaults := h.defaults(ctx)
	if h.title != "" {
		semantic.LabelOp(h.title).Add(gtx.Ops)
	}
	if h.description != "" {
		semantic.DescriptionOp(h.description).Add(gtx.Ops)
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return h.layoutIcon(ctx, gtx, defaults)
		}),
	}
	if h.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			resolved := computed.ResolveStatic(ctx, style.State{}, defaults.header, style.Style{}, style.Style{}, h.style)
			return text.WithResolvedText(text.New(h.title), resolved.Text).Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
		Gap:       gtx.Dp(defaults.metrics.HeaderGap),
	}.Layout(gtx, children...)
}

func (h dialogHeader) layoutIcon(ctx *core.Context, gtx layout.Context, defaults alertDialogDefaults) layout.Dimensions {
	diameter := min(gtx.Dp(defaults.metrics.IconSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	if diameter <= 0 {
		return layout.Dimensions{}
	}
	size := image.Pt(diameter, diameter)
	gtx.Constraints = layout.Exact(size)
	frame := clip.UniformRRect(image.Rectangle{Max: size}, diameter/2)
	paint.FillShape(gtx.Ops, defaults.icon.background, frame.Op(gtx.Ops))
	stack := frame.Push(gtx.Ops)
	defer stack.Pop()
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		glyphSize := min(gtx.Dp(defaults.metrics.IconGlyphSize), diameter)
		gtx.Constraints = layout.Exact(image.Pt(glyphSize, glyphSize))
		restore := core.PushColors(ctx, defaults.icon.foreground, defaults.icon.background)
		defer restore()
		if h.icon != nil {
			return h.icon.Layout(ctx, gtx)
		}
		if len(defaults.icon.glyph) == 0 {
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}
		return icon.Layout(defaults.icon.glyph, gtx, defaults.icon.foreground)
	})
	return layout.Dimensions{Size: size, Baseline: dims.Baseline}
}
