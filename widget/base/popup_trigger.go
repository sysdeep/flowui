package base

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
)

// PopupTriggerPolicy selects opening gestures. Focus targets retain ownership
// of their focus events; only the trigger's own tag is registered here.
type PopupTriggerPolicy struct {
	Focus     []event.Tag
	Secondary bool
	LongPress bool
	TouchOnly bool
	Delay     time.Duration
}

// PopupTriggerEvent reports an opening gesture in trigger-local coordinates.
// The component chooses its anchor rectangle, open binding and focus policy.
type PopupTriggerEvent struct {
	Position  f32.Point
	Keyboard  bool
	LongPress bool
}

// PopupTrigger observes secondary clicks, long presses and Shift+F10 without
// consuming the primary click or drawing the component's trigger.
type PopupTrigger struct {
	holding bool
	pointer pointer.ID
	start   f32.Point
	at      time.Time
	size    image.Point
}

// Reset cancels a pending long press.
func (t *PopupTrigger) Reset() { t.holding = false }

// Holding reports whether a long press is waiting for its deadline.
func (t *PopupTrigger) Holding() bool { return t.holding }

func (t *PopupTrigger) Update(gtx layout.Context, policy PopupTriggerPolicy, open func(PopupTriggerEvent)) {
	if !gtx.Enabled() {
		t.Reset()
		return
	}
	if !policy.LongPress {
		t.Reset()
	}
	filters := []event.Filter{
		pointer.Filter{Target: t, Kinds: pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Cancel},
	}
	for _, target := range policy.Focus {
		if target == nil {
			continue
		}
		filters = append(filters, key.Filter{Focus: target, Name: key.NameF10, Required: key.ModShift})
		if target == t {
			filters = append(filters, key.FocusFilter{Target: t})
		}
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		switch e := e.(type) {
		case key.Event:
			if e.State == key.Press {
				t.Reset()
				open(PopupTriggerEvent{Position: f32.Pt(float32(t.size.X)/2, float32(t.size.Y)/2), Keyboard: true})
			}
		case pointer.Event:
			// Gio cancellation resets the handler, without a source or pointer ID.
			if e.Kind == pointer.Cancel {
				t.Reset()
				continue
			}
			if policy.Secondary && e.Source == pointer.Mouse && e.Kind == pointer.Press && e.Buttons.Contain(pointer.ButtonSecondary) {
				t.Reset()
				open(PopupTriggerEvent{Position: e.Position})
				continue
			}
			if !policy.LongPress || (policy.TouchOnly && e.Source != pointer.Touch) {
				continue
			}
			switch e.Kind {
			case pointer.Press:
				if e.Buttons == 0 || core.IsPrimaryPointerPress(e) {
					t.holding, t.pointer, t.start, t.at = true, e.PointerID, e.Position, gtx.Now
				}
			case pointer.Move, pointer.Drag:
				if t.holding && e.PointerID == t.pointer {
					threshold := float32(gtx.Dp(10))
					dx, dy := e.Position.X-t.start.X, e.Position.Y-t.start.Y
					if dx > threshold || dx < -threshold || dy > threshold || dy < -threshold {
						t.Reset()
					}
				}
			case pointer.Release:
				if e.PointerID == t.pointer {
					t.Reset()
				}
			}
		}
	}
	if !t.holding {
		return
	}
	deadline := t.at.Add(policy.Delay)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return
	}
	gtx.Execute(pointer.GrabCmd{Tag: t, ID: t.pointer})
	t.Reset()
	open(PopupTriggerEvent{Position: t.start, LongPress: true})
}

// Layout mounts a pass-through observer over the already laid out trigger.
func (t *PopupTrigger) Layout(gtx layout.Context, size image.Point) {
	t.size = size
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, t)
	pass.Pop()
	area.Pop()
}
