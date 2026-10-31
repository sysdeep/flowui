package base

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"

	"github.com/qianniancn/flowui/internal/core"
)

// DragInteraction captures a primary pointer and reports positions in the
// caller's coordinate system. The caller owns the hit area and value mapping.
type DragInteraction struct {
	Dragging bool
	pointer  pointer.ID
}

// Update reports each position with started=true only for the primary press.
func (s *DragInteraction) Update(ctx *core.Context, gtx layout.Context, target event.Tag, enabled bool, position func(position f32.Point, started bool)) {
	enabled = enabled && gtx.Enabled()
	if !enabled {
		s.Dragging = false
	}
	for {
		e, ok := core.NextPointerEvent(gtx, target, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			return
		}
		if !enabled {
			continue
		}
		switch e.Kind {
		case pointer.Press:
			if s.Dragging || !core.IsPrimaryPointerPress(e) {
				continue
			}
			s.Dragging, s.pointer = true, e.PointerID
			core.PreserveFocus(ctx)
			core.RequestFocusVisible(ctx, target, false)
			core.GrabPointer(gtx, target, e)
			position(e.Position, true)
		case pointer.Drag:
			if s.Dragging && e.PointerID == s.pointer {
				position(e.Position, false)
			}
		case pointer.Release:
			if s.Dragging && e.PointerID == s.pointer {
				position(e.Position, false)
				s.Dragging = false
			}
		case pointer.Cancel:
			if e.PointerID == s.pointer {
				s.Dragging = false
			}
		}
	}
}

// AddInput mounts caller-supplied geometry, cursor and accessible text.
func (s *DragInteraction) AddInput(gtx layout.Context, target event.Tag, bounds image.Rectangle, enabled bool, cursor pointer.Cursor, label, description string) {
	area := clip.Rect(bounds).Push(gtx.Ops)
	semantic.LabelOp(label).Add(gtx.Ops)
	semantic.DescriptionOp(description).Add(gtx.Ops)
	enabled = enabled && gtx.Enabled()
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if enabled {
		cursor.Add(gtx.Ops)
		event.Op(gtx.Ops, target)
	}
	area.Pop()
}
