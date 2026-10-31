package popover

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotPopover = "popover"

func popoverStateFor(ctx *core.Context, key string) *popoverState {
	key = core.ClaimKey(ctx, core.KindPopover, key)
	return core.UseState[popoverState](ctx, key, stateSlotPopover)
}

func deletePopoverState(ctx *core.Context, key string) {
	core.DeleteState(ctx, key, stateSlotPopover)
}

type popoverState struct {
	popup        base.PopupInteraction
	transition   animation.FloatTransition
	valueBinding binding.Binding[bool]
	visualCache  popoverDefaultsCache
	open         bool // cached effective open, updated by isOpen/requestOpen
}

// popoverBindingConfig builds a binding.Config from the widget's open-state fields.
func popoverBindingConfig(widget Widget) binding.Config[bool] {
	return binding.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *popoverState) isOpen(widget Widget) bool {
	s.open = s.valueBinding.Current(popoverBindingConfig(widget))
	return s.open
}

func (s *popoverState) bind(widget Widget) {
	s.valueBinding.Bind(popoverBindingConfig(widget))
}

func (s *popoverState) requestOpen(widget Widget, open bool) bool {
	s.open, _ = s.valueBinding.Request(popoverBindingConfig(widget), open)
	return s.open
}

// progressFor is the production path: overlay timing comes from the active
// recipe visual.
func (s *popoverState) progressFor(gtx layout.Context, open bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	duration := enter
	if !open {
		duration = exit
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *popoverState) visible() bool {
	return s.transition.Ready() && s.transition.Current() > 0
}
