package base

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

const (
	splitPaneKeyboardStep     = float32(.02)
	splitPaneKeyboardPageStep = float32(.1)
)

// SplitPaneInteraction retains resize-handle input state without drawing a visual.
// Retain it between frames and use its address as the focus identity.
type SplitPaneInteraction struct {
	Hovered     bool
	Dragging    bool
	pointerID   pointer.ID
	dragOffset  float32
	orientation layout.Axis
	axisReady   bool
}

// SplitPaneGeometry describes the resize limits in physical pixels. Axis is the
// direction in which the first pane grows, not the direction of the divider line.
type SplitPaneGeometry struct {
	Axis      layout.Axis
	Available int
	First     int
	Divider   int
	Minimum   int
	Maximum   int
}

// Update handles pointer capture and keyboard resizing. Disabled or immovable
// handles cancel hover and drag state. Ratios are clamped to the supplied limits.
func (s *SplitPaneInteraction) Update(ctx *core.Context, gtx layout.Context, geometry SplitPaneGeometry, disabled bool) (float32, bool) {
	s.setOrientation(geometry.Axis)
	current := splitPaneSizeRatio(geometry.First, geometry.Available)
	if disabled || !gtx.Enabled() || geometry.Available <= 0 || geometry.Minimum >= geometry.Maximum {
		s.Hovered = false
		s.Dragging = false
		return current, false
	}
	next, changed := s.updatePointer(ctx, gtx, geometry.Axis, geometry.Available, geometry.First, geometry.Divider)
	if changed {
		first := int(math.Round(float64(splitPaneRatio(next) * float32(geometry.Available))))
		return splitPaneSizeRatio(min(max(first, geometry.Minimum), geometry.Maximum), geometry.Available), true
	}
	return s.updateKeyboard(gtx, geometry.Axis, current,
		splitPaneSizeRatio(geometry.Minimum, geometry.Available),
		splitPaneSizeRatio(geometry.Maximum, geometry.Available))
}

// AddHandle registers semantics and a caller-sized hit target. It paints nothing.
func (s *SplitPaneInteraction) AddHandle(gtx layout.Context, hit image.Rectangle, label string, disabled bool, cursor pointer.Cursor) {
	if hit.Empty() {
		return
	}
	area := clip.Rect(hit).Push(gtx.Ops)
	defer area.Pop()
	if label != "" {
		semantic.LabelOp(label).Add(gtx.Ops)
	}
	enabled := !disabled && gtx.Enabled()
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if enabled {
		core.AddPointerArea(gtx, s, hit, cursor)
	}
}

func (s *SplitPaneInteraction) setOrientation(orientation layout.Axis) {
	if s.axisReady && s.orientation != orientation {
		s.Hovered = false
		s.Dragging = false
	}
	s.orientation = orientation
	s.axisReady = true
}

func (s *SplitPaneInteraction) updatePointer(ctx *core.Context, gtx layout.Context, axis layout.Axis, available, first, divider int) (float32, bool) {
	next := splitPaneSizeRatio(first, available)
	changed := false
	for {
		eventValue, ok := core.NextPointerEvent(gtx, s, pointer.Enter|pointer.Leave|pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			return next, changed
		}
		switch eventValue.Kind {
		case pointer.Enter:
			s.Hovered = true
		case pointer.Leave:
			if !s.Dragging {
				s.Hovered = false
			}
		case pointer.Press:
			if !core.IsPrimaryPointerPress(eventValue) {
				continue
			}
			s.Dragging = true
			s.Hovered = true
			s.pointerID = eventValue.PointerID
			s.dragOffset = splitPanePointerPosition(axis, eventValue.Position) - (float32(first) + float32(divider)/2)
			core.RequestFocusVisible(ctx, s, false)
			core.GrabPointer(gtx, s, eventValue)
		case pointer.Drag:
			if !s.Dragging || eventValue.PointerID != s.pointerID {
				continue
			}
			position := splitPanePointerPosition(axis, eventValue.Position) - s.dragOffset - float32(divider)/2
			next = position / float32(available)
			changed = true
		case pointer.Release:
			if eventValue.PointerID == s.pointerID {
				s.Dragging = false
			}
		case pointer.Cancel:
			if eventValue.PointerID == s.pointerID {
				s.Dragging = false
				s.Hovered = false
			}
		}
	}
}

func (s *SplitPaneInteraction) updateKeyboard(gtx layout.Context, orientation layout.Axis, current, minimum, maximum float32) (float32, bool) {
	filters := []event.Filter{
		key.Filter{Focus: s, Name: key.NameHome},
		key.Filter{Focus: s, Name: key.NameEnd},
		key.Filter{Focus: s, Name: key.NamePageDown},
		key.Filter{Focus: s, Name: key.NamePageUp},
	}
	if orientation == layout.Vertical {
		filters = append(filters,
			key.Filter{Focus: s, Name: key.NameUpArrow},
			key.Filter{Focus: s, Name: key.NameDownArrow},
		)
	} else {
		filters = append(filters,
			key.Filter{Focus: s, Name: key.NameLeftArrow},
			key.Filter{Focus: s, Name: key.NameRightArrow},
		)
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return current, false
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		next := current
		switch event.Name {
		case key.NameLeftArrow, key.NameUpArrow:
			next -= splitPaneKeyboardStep
		case key.NameRightArrow, key.NameDownArrow:
			next += splitPaneKeyboardStep
		case key.NamePageDown:
			next -= splitPaneKeyboardPageStep
		case key.NamePageUp:
			next += splitPaneKeyboardPageStep
		case key.NameHome:
			next = minimum
		case key.NameEnd:
			next = maximum
		}
		next = min(max(next, minimum), maximum)
		return next, next != current
	}
}

func splitPanePointerPosition(axis layout.Axis, point f32.Point) float32 {
	if axis == layout.Vertical {
		return float32(point.Y)
	}
	return float32(point.X)
}

func splitPaneSizeRatio(size, available int) float32 {
	if available <= 0 {
		return 0.5
	}
	return float32(size) / float32(available)
}

func splitPaneRatio(ratio float32) float32 {
	if math.IsNaN(float64(ratio)) || math.IsInf(float64(ratio), 0) {
		return 0.5
	}
	return min(max(ratio, 0), 1)
}
