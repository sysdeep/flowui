package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

type SwitchWidget struct {
	style       style.Style
	key         string
	checked     bool
	label       string
	description string
	onChange    func(bool)
	disabled    bool
	invalid     bool
	size        SwitchSize
	labelBefore bool
	thumb       func(bool) core.Widget
}

// NewSwitch creates a controlled switch. Checked supplies the value; OnChange
// reports the requested next value.
func NewSwitch(label string) SwitchWidget {
	return SwitchWidget{
		key:   core.AutoKey("switch"),
		label: label,
	}
}

// Key sets an explicit identity for the switch state.
func (s SwitchWidget) Key(key string) SwitchWidget {
	s.key = key
	return s
}

func (s SwitchWidget) OnChange(fn func(bool)) SwitchWidget {
	s.onChange = fn
	return s
}

func (s SwitchWidget) Checked(checked bool) SwitchWidget {
	s.checked = checked
	return s
}

func (s SwitchWidget) Description(description string) SwitchWidget {
	s.description = description
	return s
}

func (s SwitchWidget) Disabled(disabled bool) SwitchWidget {
	s.disabled = disabled
	return s
}

func (s SwitchWidget) Invalid(invalid bool) SwitchWidget {
	s.invalid = invalid
	return s
}

func (s SwitchWidget) Size(size SwitchSize) SwitchWidget {
	s.size = size
	return s
}

func (s SwitchWidget) LabelBefore() SwitchWidget {
	s.labelBefore = true
	return s
}

func (s SwitchWidget) Thumb(content func(checked bool) core.Widget) SwitchWidget {
	s.thumb = content
	return s
}

func (s SwitchWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.FullKey(ctx, s.key)
	description := s.description
	if description == "" {
		description = core.FieldDescription(ctx, key)
	}
	if description != "" {
		core.PrepareFieldDescription(ctx, key, description)
	}

	return base.Toggle(s.key, s.checked).
		Label(s.label).
		Description(description).
		Role(base.ToggleSwitch).
		Disabled(s.disabled).
		Invalid(s.invalid).
		OnChange(s.onChange).
		Content(func(state base.ToggleState) core.Widget {
			if description != "" {
				core.PrepareFieldDescription(ctx, state.Key, description)
			}
			componentState := switchStateFor(ctx, state.Key)
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				styleState := style.State{
					Hovered:      state.Hovered,
					Pressed:      state.Pressed,
					Focused:      state.Focused,
					FocusVisible: state.FocusVisible,
					Disabled:     state.Disabled,
					Selected:     state.Checked,
					Checked:      state.Checked,
					Invalid:      state.Invalid,
				}
				resolved := s.resolveStyle(ctx, gtx, state.Key, styleState, &componentState.styleCache)
				motion := resolved.defaults.Motion
				resolved.selected = componentState.selection(gtx, state.Checked, resolved.defaults.SelectionDuration, motion)
				resolved.focus = componentState.focusOpacity(gtx, state.FocusVisible && !state.Disabled, resolved.defaults.FocusDuration, motion)
				return layoutui.Styled(ctx, gtx, state.Key, styleState, s.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return s.layoutContent(ctx, gtx, resolved, state.Checked)
				}))
			})
		}).
		Layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w SwitchWidget) Style(values ...style.Style) SwitchWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
