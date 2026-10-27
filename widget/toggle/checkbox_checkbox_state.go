package toggle

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const checkboxStateSlotCheckbox = "checkbox"

func checkboxCheckboxStateFor(ctx *core.Context, key string) *checkboxCheckboxState {
	return core.UseState[checkboxCheckboxState](ctx, key, checkboxStateSlotCheckbox)
}

type checkboxCheckboxState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	CheckboxSelectionAnimation
	focus      animation.FocusAnimation
	styleCache CheckboxVisualCache
}

// checkboxDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func checkboxCheckboxDisclosureCfg(widget CheckboxWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *checkboxCheckboxState) currentChecked(widget CheckboxWidget) bool {
	s.checked = s.disclosure.Current(checkboxCheckboxDisclosureCfg(widget))
	return s.checked
}

func (s *checkboxCheckboxState) bind(widget CheckboxWidget) {
	s.disclosure.Bind(checkboxCheckboxDisclosureCfg(widget))
}

func (s *checkboxCheckboxState) requestChecked(widget CheckboxWidget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(checkboxCheckboxDisclosureCfg(widget), checked)
	return s.checked
}

// CheckboxSelectionAnimation drives the shared checkbox fill and check-path progress.
type CheckboxSelectionAnimation struct {
	transition animation.FloatTransition
}

func (s *checkboxCheckboxState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.CheckboxSelectionAnimation.Progress(gtx, checked, duration, motions...)
}

func (s *CheckboxSelectionAnimation) Progress(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *checkboxCheckboxState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}
