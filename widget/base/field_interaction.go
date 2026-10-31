package base

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

// FieldInteraction tracks pointer hover and routes field-shell presses to a caller-owned focus target.
// Retain it between frames; it does not edit text or draw the field.
type FieldInteraction struct {
	Hovered bool
}

func (s *FieldInteraction) Update(ctx *core.Context, gtx layout.Context, disabled bool, tag event.Tag) {
	s.UpdateWithFocus(ctx, gtx, disabled, tag, nil)
}

// UpdateWithFocus updates hover state and lets the caller decide whether a
// pointer press should focus the field. Compound controls use this to keep
// action slots interactive without stealing focus from the editor.
func (s *FieldInteraction) UpdateWithFocus(ctx *core.Context, gtx layout.Context, disabled bool, tag event.Tag, shouldFocus func(pointer.Event) bool) {
	if disabled {
		s.Hovered = false
		return
	}
	s.updatePointer(ctx, gtx, tag, shouldFocus)
}

func (s *FieldInteraction) updatePointer(ctx *core.Context, gtx layout.Context, tag event.Tag, shouldFocus func(pointer.Event) bool) {
	for {
		e, ok := gtx.Event(pointer.Filter{
			Target: s,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Cancel | pointer.Press,
		})
		if !ok {
			return
		}
		event, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Enter:
			s.Hovered = true
		case pointer.Leave, pointer.Cancel:
			s.Hovered = false
		case pointer.Press:
			if shouldFocus == nil || shouldFocus(event) {
				core.RequestFocus(ctx, tag)
			}
		}
	}
}

// AddPointer registers a pass-through hit area without painting or claiming child input.
func (s *FieldInteraction) AddPointer(gtx layout.Context, size image.Point, disabled bool) {
	if disabled {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	stack := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, s)
	stack.Pop()
	area.Pop()
}
