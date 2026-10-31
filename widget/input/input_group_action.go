package input

import (
	"gioui.org/io/event"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

// GroupActionWidget is a compact clickable control for an input prefix
// or suffix. Its size and interaction style are stable across field heights.
type GroupActionWidget struct {
	style       style.Style
	key         string
	child       core.Widget
	label       string
	onClick     func()
	disabled    bool
	focusTarget event.Tag
}

// GroupAction creates a compact text action for an input prefix or suffix.
// The text is also its accessible label. Use Content to display an icon or
// composite widget; the default appearance supplies the action's chrome.
func GroupAction(key, label string) GroupActionWidget {
	return GroupActionWidget{
		key:   key,
		label: label,
		child: text.New(label),
	}
}

// Content replaces the visible content, preserving the accessible label.
func (a GroupActionWidget) Content(value core.Widget) GroupActionWidget {
	a.child = value
	return a
}

// OnClick sets the action callback.
func (a GroupActionWidget) OnClick(fn func()) GroupActionWidget {
	a.onClick = fn
	return a
}

// Disabled controls whether the action can be activated.
func (a GroupActionWidget) Disabled(disabled bool) GroupActionWidget {
	a.disabled = disabled
	return a
}

// Label changes the accessible action label.
func (a GroupActionWidget) Label(label string) GroupActionWidget {
	a.label = label
	return a
}

func (a GroupActionWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return a.layoutWithDefaults(ctx, gtx, groupActionDefaultsFor(ctx, a.key))
}

func (a GroupActionWidget) layoutWithDefaults(ctx *core.Context, gtx layout.Context, defaults groupActionDefaults) layout.Dimensions {
	return base.Button(a.key).Label(a.label).Disabled(a.disabled).OnClick(func() {
		if a.focusTarget != nil {
			core.RequestFocus(ctx, a.focusTarget)
		}
		if a.onClick != nil {
			a.onClick()
		}
	}).Content(func(state base.ButtonState) core.Widget {
		return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			resolved := computed.Resolve(ctx, gtx, state.Key, style.State{
				Hovered: state.Hovered, Pressed: state.Pressed, Focused: state.Focused,
				FocusVisible: state.FocusVisible, Disabled: state.Disabled,
			}, defaults.declaration, style.Style{}, style.Style{}, a.style)
			child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if state.Disabled {
					gtx = gtx.Disabled()
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if a.child == nil {
						return layout.Dimensions{}
					}
					return a.child.Layout(ctx, gtx)
				})
			})
			return layoutui.Resolved(ctx, gtx, resolved, child)
		})
	}).Layout(ctx, gtx)
}

func (a GroupActionWidget) withFocusTarget(target event.Tag) GroupActionWidget {
	a.focusTarget = target
	return a
}

func inputGroupActionWithFocus(widget core.Widget, target event.Tag) core.Widget {
	switch action := widget.(type) {
	case GroupActionWidget:
		return action.withFocusTarget(target)
	default:
		return widget
	}
}

// Style merges instance declarations into the group action. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (a GroupActionWidget) Style(values ...style.Style) GroupActionWidget {
	a.style = style.Join(a.style, style.Join(values...))
	return a
}
