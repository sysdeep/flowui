package base

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
)

// ButtonState is the interaction state a Button reports to its Content
// callback. It changes every frame and must not be retained.
type ButtonState struct {
	Interaction
	// Key is the resolved identity of the button's interaction state.
	Key string
	// Selected is the selected state for buttons that stay pressed.
	Selected bool
}

// ButtonWidget is a clickable control without a visual design. It owns
// pointer and keyboard activation, focus and button semantics, and reports
// interaction state to Content; the application owns the visual and the
// action.
//
//	base.Button("save").
//		Label("Save").
//		OnClick(func() { send(Save{}) }).
//		Content(func(state base.ButtonState) ui.Widget {
//			return card(state.Hovered, state.Pressed, state.FocusVisible)
//		})
type ButtonWidget struct {
	key         string
	label       string
	disabled    bool
	loading     bool
	selected    bool
	hasSelected bool
	onClick     func()
	content     func(ButtonState) core.Widget
	clickable   *widget.Clickable
	noEvents    bool
}

// Button creates a headless button with the given identity.
func Button(key string) ButtonWidget {
	return ButtonWidget{key: key}
}

// Label sets the accessible name. Content remains responsible for visible
// text.
func (b ButtonWidget) Label(label string) ButtonWidget {
	b.label = label
	return b
}

// Disabled stops activation and reports the disabled state.
func (b ButtonWidget) Disabled(disabled bool) ButtonWidget {
	b.disabled = disabled
	return b
}

// Loading stops activation and reports the loading state.
func (b ButtonWidget) Loading(loading bool) ButtonWidget {
	b.loading = loading
	return b
}

// Selected reports a selected/pressed state for toggle-style buttons. Ordinary
// buttons leave it unset, so they do not emit a misleading selected semantic.
func (b ButtonWidget) Selected(selected bool) ButtonWidget {
	b.selected = selected
	b.hasSelected = true
	return b
}

// Clickable supplies caller-owned interaction state. The base keeps its own
// state slot when it is nil.
func (b ButtonWidget) Clickable(value *widget.Clickable) ButtonWidget {
	b.clickable = value
	return b
}

// Events reports whether the base processes pointer and keyboard events. Disable
// it when a parent owns the gesture, such as a draggable title bar.
func (b ButtonWidget) Events(enabled bool) ButtonWidget {
	b.noEvents = !enabled
	return b
}

// OnClick runs when the pointer, Enter or Space activates the button.
func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.onClick = fn
	return b
}

// Content builds the visual for the current state. It is required.
func (b ButtonWidget) Content(fn func(ButtonState) core.Widget) ButtonWidget {
	b.content = fn
	return b
}

func (b ButtonWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if b.content == nil {
		panic("flowui: base.Button requires Content")
	}
	enabled := !b.disabled && !b.loading
	click := core.Begin(ctx, gtx, b.key, b.clickable, enabled, !b.noEvents, b.onClick)
	state := ButtonState{
		Interaction: Interaction{
			Hovered:      enabled && click.State.Hovered,
			Pressed:      enabled && click.State.Pressed,
			Focused:      click.State.Focused,
			FocusVisible: click.State.FocusVisible,
			Disabled:     !enabled || !gtx.Enabled(),
			Loading:      b.loading,
		},
		Key:      click.Key,
		Selected: b.selected,
	}
	content := b.content(state)
	if content == nil {
		panic("flowui: base.Button Content returned nil")
	}

	draw := func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := content.Layout(ctx, gtx)
		call := macro.Stop()
		area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if b.label != "" {
			semantic.LabelOp(b.label).Add(gtx.Ops)
		}
		if b.hasSelected {
			semantic.SelectedOp(b.selected).Add(gtx.Ops)
		}
		call.Add(gtx.Ops)
		area.Pop()
		return dims
	}
	if !enabled {
		// A disabled button must not own the pointer region, or controls below
		// it never receive the press. Content still draws with an enabled gtx
		// so theme transitions keep requesting frames.
		return draw(gtx)
	}
	return click.Clickable.Layout(gtx, draw)
}
