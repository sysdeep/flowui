package base

import (
	"math"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
)

// SliderKeys handles numeric axis keys without owning a focus target or visual.
// Step must be positive; Snap also rounds values to that increment.
// ShiftScale is optional and leaves Shift unclaimed when zero.
type SliderKeys struct {
	Min, Max   float64
	Step       float64
	Snap       bool
	ShiftScale float64
}

func (s SliderKeys) Update(gtx layout.Context, target event.Tag, current float64) (float64, bool) {
	modifiers := key.Modifiers(0)
	if s.ShiftScale > 0 {
		modifiers = key.ModShift
	}
	for {
		e, ok := gtx.Event(
			key.FocusFilter{Target: target},
			key.Filter{Focus: target, Name: key.NameLeftArrow, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NameRightArrow, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NameDownArrow, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NameUpArrow, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NameHome, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NameEnd, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NamePageDown, Optional: modifiers},
			key.Filter{Focus: target, Name: key.NamePageUp, Optional: modifiers},
		)
		if !ok {
			return current, false
		}
		press, ok := e.(key.Event)
		if !ok || press.State != key.Press || !gtx.Enabled() {
			continue
		}
		step := s.Step
		if press.Modifiers&key.ModShift != 0 {
			step *= s.ShiftScale
		}
		next := current
		switch press.Name {
		case key.NameLeftArrow, key.NameDownArrow:
			next -= step
		case key.NameRightArrow, key.NameUpArrow:
			next += step
		case key.NameHome:
			next = s.Min
		case key.NameEnd:
			next = s.Max
		case key.NamePageDown:
			next -= s.Step * 10
		case key.NamePageUp:
			next += s.Step * 10
		}
		if s.Snap {
			next = s.Min + math.Round((next-s.Min)/s.Step)*s.Step
		}
		next = clampFloat(next, s.Min, s.Max)
		if next != current {
			return next, true
		}
	}
}
