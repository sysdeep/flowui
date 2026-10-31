package base

import (
	"image"
	"math"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/internal/core"
)

// RangeSliderState is the interaction state a RangeSlider reports to its
// Header and Track callbacks. It changes every frame.
type RangeSliderState struct {
	Interaction
	Key         string
	Lower       float64
	Upper       float64
	Min         float64
	Max         float64
	LowerRatio  float32
	UpperRatio  float32
	Orientation Orientation
	// Dragging is 0 for the lower thumb, 1 for the upper thumb, -1 for none.
	Dragging int
	// Active is the thumb keyboard input currently adjusts.
	Active int
}

// RangeSliderWidget is a two-thumb range control without a visual design. It
// owns pointer dragging, keyboard input, focus and semantics; Content draws the
// track from RangeSliderState ratios.
type RangeSliderWidget struct {
	key         string
	lower       float64
	upper       float64
	minValue    float64
	maxValue    float64
	step        float64
	orientation Orientation
	label       string
	description string
	disabled    bool
	hitCross    unit.Dp
	inset       unit.Dp
	cursor      func(RangeSliderState) *pointer.Cursor
	onChange    func(float64, float64)
	header      func(RangeSliderState) core.Widget
	track       func(RangeSliderState) core.Widget
}

// RangeSlider creates a headless range slider. Values outside the range are
// clamped; an invalid range is a programming error.
func RangeSlider(key string, lower, upper, minValue, maxValue float64) RangeSliderWidget {
	if math.IsNaN(minValue) || math.IsInf(minValue, 0) ||
		math.IsNaN(maxValue) || math.IsInf(maxValue, 0) ||
		maxValue <= minValue {
		panic("flowui: base.RangeSlider maximum must be greater than minimum")
	}
	for _, value := range []float64{lower, upper} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			panic("flowui: base.RangeSlider values must be finite")
		}
	}
	return RangeSliderWidget{
		key:      key,
		lower:    clampFloat(lower, minValue, maxValue),
		upper:    clampFloat(upper, minValue, maxValue),
		minValue: minValue,
		maxValue: maxValue,
	}
}

func (s RangeSliderWidget) Step(step float64) RangeSliderWidget {
	if math.IsNaN(step) || math.IsInf(step, 0) || step < 0 {
		panic("flowui: base.RangeSlider step must be finite and non-negative")
	}
	s.step = step
	return s
}

func (s RangeSliderWidget) Vertical() RangeSliderWidget {
	s.orientation = Vertical
	return s
}

func (s RangeSliderWidget) Label(label string) RangeSliderWidget {
	s.label = label
	return s
}

func (s RangeSliderWidget) Description(description string) RangeSliderWidget {
	s.description = description
	return s
}

func (s RangeSliderWidget) Disabled(disabled bool) RangeSliderWidget {
	s.disabled = disabled
	return s
}

func (s RangeSliderWidget) OnRangeChange(fn func(float64, float64)) RangeSliderWidget {
	s.onChange = fn
	return s
}

func (s RangeSliderWidget) HitCross(value unit.Dp) RangeSliderWidget {
	s.hitCross = value
	return s
}

func (s RangeSliderWidget) Inset(value unit.Dp) RangeSliderWidget {
	s.inset = value
	return s
}

func (s RangeSliderWidget) Cursor(fn func(RangeSliderState) *pointer.Cursor) RangeSliderWidget {
	s.cursor = fn
	return s
}

func (s RangeSliderWidget) Header(fn func(RangeSliderState) core.Widget) RangeSliderWidget {
	s.header = fn
	return s
}

func (s RangeSliderWidget) Track(fn func(RangeSliderState) core.Widget) RangeSliderWidget {
	s.track = fn
	return s
}

func (s RangeSliderWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if s.track == nil {
		panic("flowui: base.RangeSlider requires Track")
	}
	stateKey := core.ClaimKey(ctx, core.KindBaseRangeSlider, s.key)
	state := core.UseState[rangeSliderState](ctx, stateKey, "base-range-slider")
	// A zero-valued int would look like the lower thumb is active. Keep the
	// idle state explicit so an unpressed pointer move cannot change values.
	if !state.initialized {
		state.dragging = -1
		state.initialized = true
	}
	state.lower, state.upper = s.lower, s.upper
	tag := &state.tag
	core.RegisterFieldFocus(ctx, stateKey, tag, !s.disabled && gtx.Enabled())

	disabled := s.disabled || !gtx.Enabled()
	if state.focusPending && !disabled {
		core.RequestFocusVisible(ctx, tag, false)
		gtx.Execute(key.FocusCmd{Tag: tag})
		state.focusPending = false
	}
	if !disabled {
		core.RegisterFocusGroupItem(ctx, tag, true)
		if next, ok := s.keyboard(gtx, tag, state); ok {
			core.RequestFocusVisible(ctx, tag, true)
			s.report(next)
			s.lower, s.upper = next[0], next[1]
		}
	}
	if s.lower > s.upper {
		s.lower, s.upper = s.upper, s.lower
	}
	lowerRatio := s.ratio(s.lower)
	upperRatio := s.ratio(s.upper)
	focused := gtx.Focused(tag)
	focusVisible := core.FocusVisible(ctx, tag, focused)
	public := RangeSliderState{
		Interaction: Interaction{
			Hovered:      !disabled && state.hovered,
			Pressed:      !disabled && state.pressed,
			Focused:      focused,
			FocusVisible: focusVisible,
			Disabled:     disabled,
		},
		Key:         stateKey,
		Lower:       s.lower,
		Upper:       s.upper,
		Min:         s.minValue,
		Max:         s.maxValue,
		LowerRatio:  lowerRatio,
		UpperRatio:  upperRatio,
		Orientation: s.orientation,
		Dragging:    state.dragging,
		Active:      state.active,
	}
	track := s.track(public)
	if track == nil {
		panic("flowui: base.RangeSlider Track returned nil")
	}
	var header core.Widget
	if s.header != nil {
		header = s.header(public)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if header == nil {
				return layout.Dimensions{}
			}
			return header.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutTrack(ctx, gtx, state, public, track)
		}),
	)
}

func (s RangeSliderWidget) layoutTrack(ctx *core.Context, gtx layout.Context, state *rangeSliderState, public RangeSliderState, track core.Widget) layout.Dimensions {
	tag := &state.tag
	macro := op.Record(gtx.Ops)
	trackDims := track.Layout(ctx, gtx)
	call := macro.Stop()

	axis := s.orientation.axis()
	main := axis.Convert(trackDims.Size).X
	cross := max(axis.Convert(trackDims.Size).Y, gtx.Dp(s.hitCross))
	hitSize := axis.Convert(image.Pt(main, cross))
	dragGtx := gtx
	var cursor *pointer.Cursor
	area := clip.Rect{Max: hitSize}.Push(dragGtx.Ops)
	if !public.Disabled {
		event.Op(dragGtx.Ops, tag)
		inset := min(max(gtx.Dp(s.inset), 0), main/2)
		innerMain := max(main-2*inset, 1)
		for {
			e, ok := core.NextPointerEvent(dragGtx, tag, pointer.Press|pointer.Move|pointer.Drag|pointer.Release|pointer.Cancel)
			if !ok {
				break
			}
			ratio := s.ratioAt(image.Pt(int(e.Position.X), int(e.Position.Y)), axis, innerMain, inset)
			switch e.Kind {
			case pointer.Press:
				if state.dragging >= 0 || !core.IsPrimaryPointerPress(e) {
					continue
				}
				state.pressed = true
				state.dragging = s.nearestThumb(ratio, public.LowerRatio, public.UpperRatio)
				state.active = state.dragging
				state.pointer = e.PointerID
				state.focusPending = true
				core.RequestFocusVisible(ctx, tag, false)
				core.GrabPointer(dragGtx, tag, e)
				s.updateThumb(state, ratio, public)
			case pointer.Move, pointer.Drag:
				if state.dragging >= 0 && e.PointerID == state.pointer {
					s.updateThumb(state, ratio, public)
				}
			case pointer.Release:
				if state.dragging >= 0 && e.PointerID == state.pointer {
					state.pressed = false
					state.dragging = -1
				}
			case pointer.Cancel:
				state.pressed = false
				state.dragging = -1
			}
		}
		state.hovered = state.hover.Update(gtx.Source)
		if s.cursor != nil {
			cursorState := public
			cursorState.Pressed = state.pressed
			cursorState.Dragging = state.dragging
			cursor = s.cursor(cursorState)
		}
	}
	semantic.EnabledOp(!public.Disabled).Add(dragGtx.Ops)
	if s.label != "" {
		semantic.LabelOp(s.label).Add(dragGtx.Ops)
	}
	if s.description != "" {
		semantic.DescriptionOp(s.description).Add(dragGtx.Ops)
	}
	call.Add(dragGtx.Ops)
	if cursor != nil {
		cursor.Add(dragGtx.Ops)
	}
	area.Pop()
	return trackDims
}

func (s RangeSliderWidget) updateThumb(state *rangeSliderState, ratio float32, public RangeSliderState) {
	value := s.valueAt(ratio)
	if state.active == 0 {
		value = min(value, state.upper)
		state.lower = value
	} else {
		value = max(value, state.lower)
		state.upper = value
	}
	s.report([2]float64{state.lower, state.upper})
}

func (s RangeSliderWidget) report(values [2]float64) {
	if s.onChange != nil {
		s.onChange(values[0], values[1])
	}
}

func (s RangeSliderWidget) ratioAt(position image.Point, axis layout.Axis, innerMain, inset int) float32 {
	var main int
	if axis == layout.Vertical {
		main = position.Y
	} else {
		main = position.X
	}
	ratio := float32(main-inset) / float32(innerMain)
	if axis == layout.Vertical {
		ratio = 1 - ratio
	}
	return min(max(ratio, 0), 1)
}

func (s RangeSliderWidget) nearestThumb(ratio, lower, upper float32) int {
	if absFloat32(ratio-lower) <= absFloat32(ratio-upper) {
		return 0
	}
	return 1
}

func (s RangeSliderWidget) valueAt(ratio float32) float64 {
	ratio = min(max(ratio, 0), 1)
	value := s.minValue + float64(ratio)*(s.maxValue-s.minValue)
	return s.clamp(s.snap(value))
}

func (s RangeSliderWidget) clamp(value float64) float64 {
	return clampFloat(value, s.minValue, s.maxValue)
}

func (s RangeSliderWidget) snap(value float64) float64 {
	if s.step <= 0 {
		return value
	}
	return s.minValue + math.Round((value-s.minValue)/s.step)*s.step
}

func (s RangeSliderWidget) ratio(value float64) float32 {
	return float32(clampFloat((value-s.minValue)/(s.maxValue-s.minValue), 0, 1))
}

func (s RangeSliderWidget) keyboard(gtx layout.Context, tag event.Tag, state *rangeSliderState) ([2]float64, bool) {
	step := s.step
	if step <= 0 {
		step = (s.maxValue - s.minValue) / 100
	}
	values := [2]float64{s.lower, s.upper}
	for {
		e, ok := gtx.Event(
			key.FocusFilter{Target: tag},
			key.Filter{Focus: tag, Name: key.NameLeftArrow},
			key.Filter{Focus: tag, Name: key.NameRightArrow},
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
			key.Filter{Focus: tag, Name: key.NamePageDown},
			key.Filter{Focus: tag, Name: key.NamePageUp},
		)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		index := state.active
		current := values[index]
		next := current
		switch event.Name {
		case key.NameLeftArrow, key.NameDownArrow:
			next -= step
		case key.NameRightArrow, key.NameUpArrow:
			next += step
		case key.NameHome:
			if index == 0 {
				next = s.minValue
			} else {
				next = values[0]
			}
		case key.NameEnd:
			if index == 0 {
				next = values[1]
			} else {
				next = s.maxValue
			}
		case key.NamePageDown:
			next -= step * 10
		case key.NamePageUp:
			next += step * 10
		}
		next = s.clamp(s.snap(next))
		if index == 0 {
			next = min(next, values[1])
		} else {
			next = max(next, values[0])
		}
		if next != current {
			values[index] = next
			state.lower, state.upper = values[0], values[1]
			return values, true
		}
	}
	return values, false
}

type rangeSliderState struct {
	tag          struct{}
	lower        float64
	upper        float64
	hover        gesture.Hover
	dragging     int
	active       int
	pressed      bool
	hovered      bool
	focusPending bool
	initialized  bool
	pointer      pointer.ID
}

func absFloat32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}
