package base

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
)

// ToggleRole selects what assistive technology calls a Toggle. The role changes
// semantics only; the visual and the value are the caller's.
type ToggleRole uint8

const (
	// ToggleCheckbox reports a check box, the default.
	ToggleCheckbox ToggleRole = iota
	// ToggleSwitch reports a switch.
	ToggleSwitch
	// ToggleButton reports a button that stays pressed.
	ToggleButton
)

// ToggleState is the interaction state a Toggle reports to its Content
// callback. It changes every frame and must not be retained.
type ToggleState struct {
	Interaction
	// Key is the resolved identity of the toggle's interaction state.value.
	Key string
	// Checked is the value the control paints this frame.
	Checked bool
	// Indeterminate reports the mixed state; the value stays binary.
	Indeterminate bool
}

// ToggleWidget is a check box, switch or toggle button without a visual
// design. It owns pointer and keyboard activation, focus, and semantics, and
// reports the requested value through OnChange; the application keeps the
// value in its model.
//
// Give it a stable key and a Content callback. Content draws the whole control;
// the base only measures it and wraps it in interaction and semantics:
//
//	base.Toggle("sync", model.Sync).
//		Label("Sync files").
//		OnChange(func(next bool) { send(SetSync{next}) }).
//		Content(func(state base.ToggleState) ui.Widget {
//			return indicator(state.value.Checked, state.value.Hovered)
//		})
type ToggleWidget struct {
	key           string
	checked       bool
	label         string
	description   string
	role          ToggleRole
	disabled      bool
	invalid       bool
	readOnly      bool
	indeterminate bool
	onChange      func(bool)
	content       func(ToggleState) core.Widget
}

// Toggle creates a headless check box with the given value.
func Toggle(key string, checked bool) ToggleWidget {
	return ToggleWidget{key: key, checked: checked}
}

// OnChange reports the value the user requested. The callback runs during
// layout, so it should report intent instead of retaining state from the view.
func (t ToggleWidget) OnChange(fn func(bool)) ToggleWidget {
	t.onChange = fn
	return t
}

// Label sets the accessible name without changing what Content draws.
func (t ToggleWidget) Label(label string) ToggleWidget {
	t.label = label
	return t
}

// Description sets the accessible description.
func (t ToggleWidget) Description(description string) ToggleWidget {
	t.description = description
	return t
}

// Role selects the semantics reported to assistive technology.
func (t ToggleWidget) Role(role ToggleRole) ToggleWidget {
	t.role = role
	return t
}

// Disabled stops activation and reports the disabled state to Content.
func (t ToggleWidget) Disabled(disabled bool) ToggleWidget {
	t.disabled = disabled
	return t
}

// Invalid reports the invalid state to Content and assistive technology.
func (t ToggleWidget) Invalid(invalid bool) ToggleWidget {
	t.invalid = invalid
	return t
}

// ReadOnly keeps the control focusable and hoverable but ignores activation.
func (t ToggleWidget) ReadOnly(readOnly bool) ToggleWidget {
	t.readOnly = readOnly
	return t
}

// Indeterminate reports the mixed state to Content. The value stays binary.
func (t ToggleWidget) Indeterminate(indeterminate bool) ToggleWidget {
	t.indeterminate = indeterminate
	return t
}

// Content builds the control for the current state.value. It is required and draws
// the whole visual, including its interaction box.
func (t ToggleWidget) Content(fn func(ToggleState) core.Widget) ToggleWidget {
	t.content = fn
	return t
}

func (t ToggleWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.content == nil {
		panic("flowui: base.Toggle requires Content")
	}
	stateKey := core.ClaimKey(ctx, core.KindBaseToggle, t.key)
	state := core.UseState[toggleState](ctx, stateKey, "base-toggle")

	disabled := t.disabled || !gtx.Enabled()
	if disabled {
		gtx = gtx.Disabled()
	}
	presses := core.ActivePresses(state.value.History())
	focusedBefore := gtx.Focused(&state.value)
	state.value.Value = t.checked
	checked := t.checked
	if !disabled && state.value.Update(gtx) {
		if !focusedBefore {
			core.RequestFocusVisible(ctx, &state.value, false)
		}
		if !t.readOnly {
			checked = !t.checked
			if t.onChange != nil {
				t.onChange(checked)
			}
		}
	}
	state.value.Value = checked

	return state.value.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !disabled {
			core.FocusOnPress(ctx, &state.value, state.value.History(), presses)
			core.RegisterFocusGroupItem(ctx, &state.value, true)
		}
		focused := gtx.Focused(&state.value)
		focusVisible := core.FocusVisible(ctx, &state.value, focused)
		content := t.content(ToggleState{
			Interaction: Interaction{
				Hovered:      !disabled && state.value.Hovered(),
				Pressed:      !disabled && state.value.Pressed(),
				Focused:      focused,
				FocusVisible: focusVisible,
				Disabled:     disabled,
				Invalid:      t.invalid,
				ReadOnly:     t.readOnly,
			},
			Key:           stateKey,
			Checked:       checked,
			Indeterminate: t.indeterminate,
		})
		if content == nil {
			panic("flowui: base.Toggle Content returned nil")
		}

		dims := content.Layout(ctx, gtx)

		area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		t.role.class().Add(gtx.Ops)
		if t.label != "" {
			semantic.LabelOp(t.label).Add(gtx.Ops)
		}
		if t.description != "" {
			semantic.DescriptionOp(t.description).Add(gtx.Ops)
		}
		semantic.SelectedOp(checked || t.indeterminate).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		area.Pop()
		return dims
	})
}

func (r ToggleRole) class() semantic.ClassOp {
	switch r {
	case ToggleSwitch:
		return semantic.Switch
	case ToggleButton:
		return semantic.Button
	default:
		return semantic.CheckBox
	}
}

type toggleState struct {
	value widget.Bool
}
