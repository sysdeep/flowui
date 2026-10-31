package button

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type buttonPreparedContent struct {
	call         op.CallOp
	dims         layout.Dimensions
	width        int
	declarations buttonDefaults
}

// Measure resolves static geometry without updating interaction or transition state.
func (b Widget) Measure(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	gtx.Ops = new(op.Ops)
	resolved := b.staticStyle(ctx, style.State{Disabled: b.disabled || !gtx.Enabled(), Loading: b.loading})
	child := b.child
	b.child = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return core.MeasureWidget(ctx, gtx, child)
	})
	if b.width != nil {
		width := min(max(gtx.Dp(*b.width), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
	} else if b.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return b.layoutContent(ctx, gtx, resolved)
		})
	}))
}

func (b Widget) prepareContent(ctx *core.Context, gtx layout.Context) buttonPreparedContent {
	activeTheme := core.ActiveTheme(ctx)
	key := core.FullKey(ctx, b.key)
	declarations := b.declarations(activeTheme, core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle))
	style := b.staticContentStyleWithDeclarations(ctx, style.State{Disabled: b.disabled, Loading: b.loading}, declarations)
	measureGtx := gtx
	if b.disabled || b.loading {
		measureGtx = measureGtx.Disabled()
	}
	measureGtx.Constraints.Min.X = 0
	measureGtx.Constraints.Min.Y = 0
	macro := op.Record(gtx.Ops)
	dims := b.layoutContent(ctx, measureGtx, style)
	call := macro.Stop()
	visualMacro := op.Record(gtx.Ops)
	visualDims := layoutui.Resolved(ctx, measureGtx, style.root, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return dims
		})
	}))
	_ = visualMacro.Stop()
	return buttonPreparedContent{call: call, dims: dims, width: visualDims.Size.X, declarations: declarations}
}

func (b Widget) layoutContent(ctx *core.Context, gtx layout.Context, resolved buttonStyle) layout.Dimensions {
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved.content, b.child)
	})
	if !b.loading {
		return content.Layout(ctx, gtx)
	}
	spinner := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved.indicatorPart, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			// The indicator part has already gone through the component cascade.
			// Keep the spinner on that resolved color so instance and scoped
			// PartIndicator overrides apply to the loading visual as well.
			indicator := buttonForeground(resolved.indicatorPart)
			return layoutSpinner(gtx, resolved.defaults, indicator)
		}))
	})
	if b.iconOnly {
		return spinner.Layout(ctx, gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(8),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return spinner.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return content.Layout(ctx, gtx)
		}),
	)
}

func layoutSpinner(gtx layout.Context, defaults buttonDefaults, col color.NRGBA) layout.Dimensions {
	period := theme.ResolveMotionDuration(defaults.motion, 900*time.Millisecond)
	return drawSpinner(gtx, defaults.spinner, 2, col, period)
}
