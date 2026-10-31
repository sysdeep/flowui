package base

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

// TooltipPolicy controls passive visibility; Visible allows immediate re-entry
// while the caller's exit animation is still on screen.
type TooltipPolicy struct {
	Focus             bool
	Disabled          bool
	Visible           bool
	Delay, CloseDelay time.Duration
}

// TooltipInteraction tracks hover/focus and visibility deadlines. It owns no
// overlay, pointer blocker, focus command, theme or animation.
type TooltipInteraction struct {
	Hovered, Focused, Active, Open bool
	showAt, hideAt                 time.Time
}

// Close dismisses immediately without resetting the observed trigger activity.
func (s *TooltipInteraction) Close() {
	s.showAt, s.hideAt = time.Time{}, time.Time{}
	s.Open = false
}

func (s *TooltipInteraction) Update(gtx layout.Context, target event.Tag, policy TooltipPolicy) {
	s.updateEvents(gtx, target, policy.Focus)
	active := s.Hovered
	if policy.Focus {
		active = s.Focused
	}
	s.UpdateActive(gtx, active, policy)
}

// UpdateActive accepts caller-observed activity instead of reading trigger input.
func (s *TooltipInteraction) UpdateActive(gtx layout.Context, active bool, policy TooltipPolicy) {
	if policy.Disabled || !gtx.Enabled() {
		s.Hovered = false
		s.Focused = false
		active = false
		s.showAt = time.Time{}
		s.hideAt = time.Time{}
		s.Open = false
	}
	if active != s.Active {
		s.Active = active
		if active {
			s.hideAt = time.Time{}
			if s.Open || policy.Visible || policy.Delay <= 0 {
				s.Open = true
				s.showAt = time.Time{}
			} else {
				s.showAt = gtx.Now.Add(policy.Delay)
			}
		} else {
			s.showAt = time.Time{}
			if !s.Open {
				s.hideAt = time.Time{}
			} else if policy.CloseDelay <= 0 {
				s.Open = false
				s.hideAt = time.Time{}
			} else {
				s.hideAt = gtx.Now.Add(policy.CloseDelay)
			}
		}
	}
	s.resolveDeadline(gtx)
}

func (s *TooltipInteraction) updateEvents(gtx layout.Context, target event.Tag, focus bool) {
	for {
		eventValue, ok := core.NextPointerEvent(gtx, target, pointer.Enter|pointer.Leave|pointer.Cancel)
		if !ok {
			break
		}
		switch eventValue.Kind {
		case pointer.Enter:
			s.Hovered = true
		case pointer.Leave, pointer.Cancel:
			s.Hovered = false
		}
	}
	if !focus {
		return
	}
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: target})
		if !ok {
			break
		}
		if event, ok := e.(key.FocusEvent); ok {
			s.Focused = event.Focus
		}
	}
}

func (s *TooltipInteraction) resolveDeadline(gtx layout.Context) {
	if !s.showAt.IsZero() {
		if !gtx.Now.Before(s.showAt) {
			s.Open = s.Active
			s.showAt = time.Time{}
		} else {
			gtx.Execute(op.InvalidateCmd{At: s.showAt})
		}
	}
	if !s.hideAt.IsZero() {
		if !gtx.Now.Before(s.hideAt) {
			s.Open = s.Active
			s.hideAt = time.Time{}
		} else {
			gtx.Execute(op.InvalidateCmd{At: s.hideAt})
		}
	}
}

func (s *TooltipInteraction) AddInput(gtx layout.Context, target event.Tag, size image.Point, disabled bool) {
	if disabled || !gtx.Enabled() || size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, target)
	pass.Pop()
	area.Pop()
}
