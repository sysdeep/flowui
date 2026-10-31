package toggle

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

const checkboxStateSlot = "checkbox"

func checkboxStateFor(ctx *core.Context, key string) *checkboxState {
	return core.UseState[checkboxState](ctx, key, checkboxStateSlot)
}

type checkboxState struct {
	CheckboxSelectionAnimation
	focus      animation.FocusAnimation
	styleCache CheckboxVisualCache
}

// CheckboxSelectionAnimation drives the shared checkbox fill and check-path progress.
type CheckboxSelectionAnimation struct {
	transition animation.FloatTransition
}

func (s *checkboxState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.CheckboxSelectionAnimation.Progress(gtx, checked, duration, motions...)
}

func (s *CheckboxSelectionAnimation) Progress(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *checkboxState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}

const switchStateSlot = "switch"

func switchStateFor(ctx *core.Context, key string) *switchState {
	return core.UseState[switchState](ctx, key, switchStateSlot)
}

type switchState struct {
	selected   animation.FloatTransition
	focus      animation.FocusAnimation
	styleCache switchVisualCache
}

func (s *switchState) selection(gtx layout.Context, checked bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.selected.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *switchState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}

const toggleButtonStateSlot = "togglebutton-visual"

type toggleButtonState struct {
	visualCache toggleButtonVisualCache
}

func toggleButtonStateFor(ctx *core.Context, key string) *toggleButtonState {
	return core.UseState[toggleButtonState](ctx, key, toggleButtonStateSlot)
}
