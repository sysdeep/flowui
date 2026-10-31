package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

func (b ToggleButtonWidget) layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	selected := b.selected
	return base.Button(b.key).
		Label(b.label).
		Disabled(b.disabled).
		Selected(selected).
		OnClick(func() {
			if b.onChange != nil {
				selected = !b.selected
				b.onChange(selected)
			}
		}).
		Content(func(state base.ButtonState) core.Widget {
			componentState := toggleButtonStateFor(ctx, state.Key)
			styleState := style.State{
				Hovered:      state.Hovered,
				Pressed:      state.Pressed,
				Focused:      state.Focused,
				FocusVisible: state.FocusVisible,
				Disabled:     state.Disabled,
				Selected:     selected,
				Checked:      selected,
			}
			resolved := b.resolveStyle(ctx, gtx, state.Key, styleState, &componentState.visualCache)
			layoutGtx := gtx
			if state.Disabled {
				layoutGtx = layoutGtx.Disabled()
			}

			content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutui.Resolved(ctx, layoutGtx, resolved.content, b.child)
				})
			})
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, layoutGtx, resolved.root, content)
			})
		}).
		Layout(ctx, gtx)
}
