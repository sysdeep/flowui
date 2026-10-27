package toggle

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type CheckboxVariant = theme.CheckboxVariant

const (
	CheckboxPrimary   = theme.CheckboxPrimary
	CheckboxSecondary = theme.CheckboxSecondary
)

// CheckboxIndicatorState describes the controlled state passed to a custom indicator.
type CheckboxIndicatorState struct {
	Checked       bool
	Indeterminate bool
	Disabled      bool
	Invalid       bool
}

type CheckboxWidget struct {
	override       *theme.CheckboxOverride
	key            string
	checked        bool
	hasChecked     bool
	defaultChecked bool
	hasDefault     bool
	label          string
	description    string
	errorMessage   string
	onChange       func(bool)
	indicator      func(CheckboxIndicatorState) core.Widget
	variant        CheckboxVariant
	disabled       bool
	invalid        bool
	indeterminate  bool
	readOnly       bool
	required       bool
}

// New creates an uncontrolled checkbox that starts unchecked. Use Checked for
// controlled state and DefaultChecked for an uncontrolled starting value.
func NewCheckbox(label string) CheckboxWidget {
	return CheckboxWidget{
		key:   core.AutoKey(),
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
	c.hasChecked = true
	return c
}

func (c CheckboxWidget) DefaultChecked(checked bool) CheckboxWidget {
	c.defaultChecked = checked
	c.hasDefault = true
	c.hasChecked = false // Switch to uncontrolled mode
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
	return c.layoutWithOverride(ctx, gtx, c.override)
}

func (c CheckboxWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CheckboxOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	key, valueState := core.BoolStateWithKey(ctx, c.key)
	anim := checkboxCheckboxStateFor(ctx, key)

	// Bind disclosure state
	anim.bind(c)
	checked := anim.currentChecked(c)

	valueState.Value = checked
	animGtx := gtx
	presses := core.ActivePresses(valueState.History())
	disabled := c.disabled || !gtx.Enabled()
	message := c.supportingText()
	if message != "" {
		core.PrepareFieldDescription(ctx, key, message)
	}
	if disabled {
		gtx = gtx.Disabled()
	} else if valueState.Update(gtx) {
		next := valueState.Value
		valueState.Value = checked
		if !c.readOnly {
			checked = anim.requestChecked(c, next)
		}
	}
	dims := valueState.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !disabled {
			core.FocusOnPress(ctx, valueState, valueState.History(), presses)
		}
		focusVisible := core.FocusVisible(ctx, valueState, gtx.Focused(valueState))
		styleState := style.State{
			Hovered:       valueState.Hovered(),
			Pressed:       valueState.Pressed(),
			Focused:       gtx.Focused(valueState),
			FocusVisible:  focusVisible,
			Disabled:      disabled,
			Selected:      checked || c.indeterminate,
			Checked:       checked,
			Indeterminate: c.indeterminate,
			ReadOnly:      c.readOnly,
			Invalid:       c.invalid,
		}
		resolved := c.resolveStyle(ctx, animGtx, key, styleState, override, &anim.styleCache)
		selection := anim.selection(animGtx, checked || c.indeterminate, resolved.selectionDuration, resolved.motion)
		focus := anim.focusOpacity(animGtx, focusVisible && !disabled, resolved.focusDuration, resolved.motion)
		indicatorState := CheckboxIndicatorState{
			Checked: checked, Indeterminate: c.indeterminate,
			Disabled: disabled, Invalid: c.invalid,
		}
		options := CheckboxControlOptions{
			Variant: c.variant, Selection: selection, Indeterminate: c.indeterminate,
			Hovered: valueState.Hovered(), Pressed: valueState.Pressed(),
			Focused: focus, Disabled: disabled, Invalid: c.invalid,
			CustomIndicator: c.indicator != nil,
			Indicator:       c.indicatorWidget(indicatorState),
			resolvedStyle:   &resolved.indicator,
			resolvedVisual:  &resolved.visual,
		}
		dims := layoutui.Styled(ctx, gtx, key, styleState, resolved.override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return c.layoutContent(ctx, gtx, options, indicatorState, resolved)
		}))
		semanticClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.CheckBox.Add(gtx.Ops)
		if c.label != "" {
			semantic.LabelOp(c.label).Add(gtx.Ops)
		}
		if description := c.semanticDescription(ctx, key); description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		semanticClip.Pop()
		return dims
	})
	return dims
}

func (c CheckboxWidget) supportingText() string {
	if c.invalid && c.errorMessage != "" {
		return c.errorMessage
	}
	return c.description
}

func (c CheckboxWidget) semanticDescription(ctx *core.Context, key string) string {
	if message := c.supportingText(); message != "" {
		return message
	}
	return core.FieldDescription(ctx, key)
}

func (c CheckboxWidget) indicatorWidget(state CheckboxIndicatorState) core.Widget {
	if c.indicator == nil {
		return nil
	}
	return c.indicator(state)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w CheckboxWidget) Style(values ...style.Style) CheckboxWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.CheckboxOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
