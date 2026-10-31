package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

// CheckboxIndicatorState describes the controlled state passed to a custom indicator.
type CheckboxIndicatorState struct {
	Checked       bool
	Indeterminate bool
	Disabled      bool
	Invalid       bool
}

type CheckboxWidget struct {
	style         style.Style
	key           string
	checked       bool
	label         string
	description   string
	errorMessage  string
	onChange      func(bool)
	indicator     func(CheckboxIndicatorState) core.Widget
	variant       CheckboxVariant
	disabled      bool
	invalid       bool
	indeterminate bool
	readOnly      bool
	required      bool
}

// NewCheckbox creates a controlled checkbox. Checked supplies the value; OnChange
// reports the requested next value.
func NewCheckbox(label string) CheckboxWidget {
	return CheckboxWidget{
		key:   core.AutoKey("checkbox"),
		label: label,
	}
}

// Key sets an explicit identity for the checkbox state.
func (c CheckboxWidget) Key(key string) CheckboxWidget {
	c.key = key
	return c
}

func (c CheckboxWidget) OnChange(fn func(bool)) CheckboxWidget {
	c.onChange = fn
	return c
}

func (c CheckboxWidget) Checked(checked bool) CheckboxWidget {
	c.checked = checked
	return c
}

func (c CheckboxWidget) Variant(variant CheckboxVariant) CheckboxWidget {
	c.variant = variant
	return c
}

func (c CheckboxWidget) Indeterminate(indeterminate bool) CheckboxWidget {
	c.indeterminate = indeterminate
	return c
}

func (c CheckboxWidget) ReadOnly(readOnly bool) CheckboxWidget {
	c.readOnly = readOnly
	return c
}

func (c CheckboxWidget) Required(required bool) CheckboxWidget {
	c.required = required
	return c
}

func (c CheckboxWidget) Description(description string) CheckboxWidget {
	c.description = description
	return c
}

func (c CheckboxWidget) ErrorMessage(message string) CheckboxWidget {
	c.errorMessage = message
	return c
}

func (c CheckboxWidget) Indicator(indicator func(CheckboxIndicatorState) core.Widget) CheckboxWidget {
	c.indicator = indicator
	return c
}

func (c CheckboxWidget) Disabled(disabled bool) CheckboxWidget {
	c.disabled = disabled
	return c
}

func (c CheckboxWidget) Invalid(invalid bool) CheckboxWidget {
	c.invalid = invalid
	return c
}

func (c CheckboxWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	message := c.supportingText()
	return base.Toggle(c.key, c.checked).
		Label(c.label).
		Description(message).
		Disabled(c.disabled).
		Invalid(c.invalid).
		ReadOnly(c.readOnly).
		Indeterminate(c.indeterminate).
		OnChange(c.onChange).
		Content(func(state base.ToggleState) core.Widget {
			if message != "" {
				core.PrepareFieldDescription(ctx, state.Key, message)
			}
			anim := checkboxStateFor(ctx, state.Key)
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				styleState := style.State{
					Hovered:       state.Hovered,
					Pressed:       state.Pressed,
					Focused:       state.Focused,
					FocusVisible:  state.FocusVisible,
					Disabled:      state.Disabled,
					Selected:      state.Checked || c.indeterminate,
					Checked:       state.Checked,
					Indeterminate: c.indeterminate,
					ReadOnly:      state.ReadOnly,
					Invalid:       state.Invalid,
				}
				resolved := c.resolveStyle(ctx, gtx, state.Key, styleState, &anim.styleCache)
				selection := anim.selection(gtx, state.Checked || c.indeterminate, resolved.defaults.SelectionDuration, resolved.defaults.Motion)
				focus := anim.focusOpacity(gtx, state.FocusVisible && !state.Disabled, resolved.defaults.FocusDuration, resolved.defaults.Motion)
				indicatorState := CheckboxIndicatorState{
					Checked: state.Checked, Indeterminate: c.indeterminate,
					Disabled: state.Disabled, Invalid: c.invalid,
				}
				options := CheckboxControlOptions{
					Variant: c.variant, Selection: selection, Indeterminate: c.indeterminate,
					Hovered: state.Hovered, Pressed: state.Pressed,
					Focused: focus, Disabled: state.Disabled, Invalid: c.invalid,
					CustomIndicator:  c.indicator != nil,
					Indicator:        c.indicatorWidget(indicatorState),
					resolvedStyle:    &resolved.indicator,
					resolvedDefaults: &resolved.defaults,
				}
				return layoutui.Styled(ctx, gtx, state.Key, styleState, c.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return c.layoutContent(ctx, gtx, options, indicatorState, resolved)
				}))
			})
		}).
		Layout(ctx, gtx)
}

func (c CheckboxWidget) supportingText() string {
	if c.invalid && c.errorMessage != "" {
		return c.errorMessage
	}
	return c.description
}

func (c CheckboxWidget) indicatorWidget(state CheckboxIndicatorState) core.Widget {
	if c.indicator == nil {
		return nil
	}
	return c.indicator(state)
}

// Style merges instance declarations into the checkbox. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (c CheckboxWidget) Style(values ...style.Style) CheckboxWidget {
	c.style = style.Join(c.style, style.Join(values...))
	return c
}
