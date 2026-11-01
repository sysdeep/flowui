package button

import (
	"gioui.org/layout"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/icon"
)

type CloseButtonWidget struct {
	style      style.Style
	key        string
	onClick    func()
	disabled   bool
	icon       core.Widget
	label      string
	styleCache *closeStyleCache
}

func NewCloseButton() CloseButtonWidget {
	return CloseButtonWidget{key: core.AutoKey("close-button"), styleCache: &closeStyleCache{}}
}

// Key sets a stable identity for the close button's interaction state.
func (b CloseButtonWidget) Key(key string) CloseButtonWidget {
	b.key = key
	return b
}

func (b CloseButtonWidget) OnClick(fn func()) CloseButtonWidget {
	b.onClick = fn
	return b
}

func (b CloseButtonWidget) Disabled(disabled bool) CloseButtonWidget {
	b.disabled = disabled
	return b
}

func (b CloseButtonWidget) Icon(icon core.Widget) CloseButtonWidget {
	b.icon = icon
	return b
}

func (b CloseButtonWidget) Label(label string) CloseButtonWidget {
	b.label = label
	return b
}

func (b CloseButtonWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return closeLayoutWithClickable(b, ctx, gtx, nil, true)
}

// LayoutCloseButtonWithClickableNoEvents renders a close button with caller-owned state and events.
func LayoutCloseButtonWithClickableNoEvents(b CloseButtonWidget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return closeLayoutWithClickable(b, ctx, gtx, clickable, false)
}

func closeLayoutWithClickable(
	b CloseButtonWidget,
	ctx *core.Context,
	gtx layout.Context,
	clickable *widget.Clickable,
	handleEvents bool,
) layout.Dimensions {
	defaults := b.styles(ctx)
	return base.Button(b.key).
		Label(b.semanticLabel(ctx)).
		Disabled(b.disabled).
		OnClick(b.onClick).
		Clickable(clickable).
		Events(handleEvents).
		Content(func(state base.ButtonState) core.Widget {
			resolved := b.resolveStyle(ctx, gtx, state.Key, styleState(state), defaults)
			content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					child := b.icon
					if child == nil {
						child = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
							return icon.Layout(lucide.X, gtx, ctx.ForegroundColor())
						})
					}
					return layoutui.Resolved(ctx, gtx, resolved.icon, child)
				})
			})
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, gtx, resolved.root, content)
			})
		}).
		Layout(ctx, gtx)
}

func styleState(state base.ButtonState) style.State {
	return style.State{
		Hovered:      state.Hovered,
		Pressed:      state.Pressed,
		Focused:      state.Focused,
		FocusVisible: state.FocusVisible,
		Disabled:     state.Disabled,
		Loading:      state.Loading,
	}
}

func (b CloseButtonWidget) semanticLabel(ctx *core.Context) string {
	if b.label != "" {
		return b.label
	}
	return ctx.Text(i18n.ActionClose)
}

// Style merges instance declarations into the built-in close-button appearance.
func (b CloseButtonWidget) Style(values ...style.Style) CloseButtonWidget {
	b.style = style.Join(b.style, style.Join(values...))
	return b
}
