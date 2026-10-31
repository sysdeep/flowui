package base

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

// HoverInteraction observes a caller-owned target without focusing it or
// consuming presses. Retain the state between frames.
type HoverInteraction struct {
	Hovered bool
}

func (s *HoverInteraction) Update(gtx layout.Context, target event.Tag, enabled bool) {
	enabled = enabled && gtx.Enabled()
	if !enabled {
		s.Hovered = false
	}
	for {
		e, ok := core.NextPointerEvent(gtx, target, pointer.Enter|pointer.Leave|pointer.Cancel)
		if !ok {
			return
		}
		s.Hovered = enabled && e.Kind == pointer.Enter
	}
}

// AddInput mounts a pass-through region. A nil cursor leaves the child's cursor
// unchanged; no focus filter or Tab stop is added.
func (s *HoverInteraction) AddInput(gtx layout.Context, target event.Tag, bounds image.Rectangle, enabled bool, cursor *pointer.Cursor) {
	if !enabled || !gtx.Enabled() || bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	if cursor != nil {
		cursor.Add(gtx.Ops)
	}
	event.Op(gtx.Ops, target)
	pass.Pop()
	area.Pop()
}
