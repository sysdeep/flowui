package toggle

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const switchStateSlotSwitch = "switch"

func switchSwitchStateFor(ctx *core.Context, key string) *switchSwitchState {
	key = core.ClaimKey(ctx, core.KindSwitch, key)
	return core.UseState[switchSwitchState](ctx, key, switchStateSlotSwitch)
}

type switchSwitchState struct {
	disclosure disclosure.Binding[bool]
	checked    bool
	value      widget.Bool
	selected   animation.FloatTransition
	focus      animation.FocusAnimation
	styleCache SwitchVisualCache
}

// switchDisclosureCfg builds a disclosure.Config from the widget's checked fields.
func switchSwitchDisclosureCfg(widget SwitchWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasChecked,
		Value:      widget.checked,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultChecked,
		OnChange:   widget.onChange,
	}
}

func (s *switchSwitchState) currentChecked(widget SwitchWidget) bool {
	s.checked = s.disclosure.Current(switchSwitchDisclosureCfg(widget))
	return s.checked
}

func (s *switchSwitchState) bind(widget SwitchWidget) {
	s.disclosure.Bind(switchSwitchDisclosureCfg(widget))
}

func (s *switchSwitchState) requestChecked(widget SwitchWidget, checked bool) bool {
	s.checked, _ = s.disclosure.Request(switchSwitchDisclosureCfg(widget), checked)
	return s.checked
}

func (s *switchSwitchState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.selected.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *switchSwitchState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}
