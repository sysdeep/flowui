package modal

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/internal/binding"
)

const stateSlotModal = "modal"

func modalStateFor(ctx *core.Context, key string) *modalState {
	key = core.ClaimKey(ctx, core.KindModal, key)
	return core.UseState[modalState](ctx, key, stateSlotModal)
}

func deleteModalState(ctx *core.Context, key string) {
	core.DeleteState(ctx, key, stateSlotModal)
}

type modalState struct {
	dialog              base.DialogInteraction
	close               widget.Clickable
	bodyList            layout.List
	bodyBar             widget.Scrollbar
	bodyVisualOutset    layoutui.VisualOutsetState
	outsideList         layout.List
	outsideVisualOutset layoutui.VisualOutsetState
	transition          animation.FloatTransition
	visualCache         modalDefaultsCache
	closeButton         *button.CloseButtonWidget
	valueBinding        binding.Binding[bool]
	open                bool // cached effective open, updated by isOpen/requestOpen
}

// modalBindingConfig builds a binding.Config from the widget's open-state fields.
func modalBindingConfig(widget Widget) binding.Config[bool] {
	return binding.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *modalState) isOpen(widget Widget) bool {
	s.open = s.valueBinding.Current(modalBindingConfig(widget))
	return s.open
}

func (s *modalState) bind(widget Widget) {
	s.valueBinding.Bind(modalBindingConfig(widget))
}

func (s *modalState) requestOpen(widget Widget, open bool) bool {
	s.open, _ = s.valueBinding.Request(modalBindingConfig(widget), open)
	return s.open
}

// progressFor is the production path: the modal defaults own the motion
// timing.
func (s *modalState) progressFor(gtx layout.Context, open bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	s.transition.Initialize(0, gtx.Now)
	easing := animation.EaseSmoothstep
	if len(motions) > 0 {
		motion := motions[0]
		if resolved := animation.EasingFor(motion.Ease); resolved != nil {
			easing = resolved
		}
	}
	return s.transition.Value(gtx, target, duration, easing, motions...)
}

func (s *modalState) visible() bool {
	return s.transition.Ready() && s.transition.Current() > 0
}

func (s *modalState) opening() bool {
	return s.transition.Increasing()
}

// closeButtonWidget retains the visual cache across modal frames while
// keeping CloseButton's behavior and clickable state owned by modalState.
func (s *modalState) closeButtonWidget() *button.CloseButtonWidget {
	if s.closeButton == nil {
		value := button.NewCloseButton().Key("close")
		s.closeButton = &value
	}
	return s.closeButton
}
