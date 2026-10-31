package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

func applyFixedWidth(gtx *layout.Context, width *unit.Dp) {
	if width == nil {
		return
	}
	fixed := min(max(gtx.Dp(*width), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
	gtx.Constraints.Min.X = fixed
	gtx.Constraints.Max.X = fixed
}

func (i Widget) layoutFrame(ctx *core.Context, gtx layout.Context, state *inputState, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.FieldInteraction, resolved, enabled, func(gtx layout.Context) layout.Dimensions {
		minWidth := gtx.Constraints.Min.X
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = min(minWidth, gtx.Constraints.Max.X)
			return child(gtx)
		})
	})
}

// layoutFieldFrame draws the field box and registers its pointer area.
func layoutFieldFrame(ctx *core.Context, gtx layout.Context, state *base.FieldInteraction, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutui.InteractiveResolved(
		ctx,
		gtx,
		resolved,
		core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions { return child(gtx) }),
		func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			dims := visual(gtx)
			state.AddPointer(gtx, dims.Size, !enabled)
			return dims
		},
	)
}

func (t TextAreaWidget) layoutFrame(ctx *core.Context, gtx layout.Context, state *textAreaState, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.FieldInteraction, resolved, enabled, child)
}
