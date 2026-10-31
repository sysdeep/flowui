package base

import (
	"image"
	"math"

	"gioui.org/f32"
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

// Orientation is the direction a Slider travels.
type Orientation uint8

const (
	// Horizontal places the low end at the start of the content box.
	Horizontal Orientation = iota
	// Vertical places the low end at the bottom of the content box.
	Vertical
)

// SliderState is the interaction state a Slider reports to its Header and Track
// callback. It changes every frame and must not be retained.
type SliderState struct {
	Interaction
	// Key is the resolved identity of the slider's interaction state.
	Key string
	// Value is the controlled value the control paints this frame.
	Value float64
	Min   float64
	Max   float64
	// Ratio is Value mapped to [0, 1] along the slider axis.
	Ratio       float32
	Orientation Orientation
	// Dragging reports an active pointer drag.
	Dragging bool
}

// SliderWidget is a slider without a visual design. It owns pointer dragging,
// the arrow, Home, End, Page Up and Page Down keys, focus, and semantics, and
// reports the requested value through OnChange; the application keeps the
// value in its model.
//
// Give it a stable key, an optional Header and a Track callback. Track draws
// the interactive area; Header is optional chrome above it:
//
//	base.Slider("volume", model.Volume, 0, 100).
//		Step(1).
//		Label("Volume").
//		OnChange(func(next float64) { send(SetVolume{next}) }).
//		Track(func(state base.SliderState) ui.Widget {
//			return gauge(state.Ratio, state.Dragging)
//		})
type SliderWidget struct {
	key         string
	value       float64
	minValue    float64
	maxValue    float64
	step        float64
	orientation Orientation
	label       string
	description string
	disabled    bool
	onChange    func(float64)
	header      func(SliderState) core.Widget
	track       func(SliderState) core.Widget
	hitCross    unit.Dp
	inset       unit.Dp
	cursor      func(SliderState) *pointer.Cursor
}

// Slider creates a headless slider. Values outside the range are clamped; an
// invalid range is a programming error.
func Slider(key string, value, minValue, maxValue float64) SliderWidget {
	if math.IsNaN(minValue) || math.IsInf(minValue, 0) ||
		math.IsNaN(maxValue) || math.IsInf(maxValue, 0) ||
		maxValue <= minValue {
		panic("flowui: base.Slider maximum must be greater than minimum")
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic("flowui: base.Slider value must be finite")
	}
	return SliderWidget{
		key:      key,
		value:    clampFloat(value, minValue, maxValue),
		minValue: minValue,
		maxValue: maxValue,
	}
}

// Step snaps dragging and keyboard changes to multiples of step from the
// minimum. Zero keeps dragging continuous and uses a hundredth of the range
// for the keyboard.
func (s SliderWidget) Step(step float64) SliderWidget {
	if math.IsNaN(step) || math.IsInf(step, 0) || step < 0 {
		panic("flowui: base.Slider step must be finite and non-negative")
	}
	s.step = step
	return s
}

// Vertical places the low end at the bottom of the content box.
func (s SliderWidget) Vertical() SliderWidget {
	s.orientation = Vertical
	return s
}

// Label sets the accessible name.
func (s SliderWidget) Label(label string) SliderWidget {
	s.label = label
	return s
}

// Description sets the accessible description.
func (s SliderWidget) Description(description string) SliderWidget {
	s.description = description
	return s
}

// Disabled stops interaction and reports the disabled state.
func (s SliderWidget) Disabled(disabled bool) SliderWidget {
	s.disabled = disabled
	return s
}

// OnChange reports the value the user requested.
func (s SliderWidget) OnChange(fn func(float64)) SliderWidget {
	s.onChange = fn
	return s
}

// Header draws optional chrome above the track. It is not interactive.
func (s SliderWidget) Header(fn func(SliderState) core.Widget) SliderWidget {
	s.header = fn
	return s
}

// Track draws the slider track for the current state. It is required, and its
// measured size is the drag area.
func (s SliderWidget) Track(fn func(SliderState) core.Widget) SliderWidget {
	s.track = fn
	return s
}

// HitCross reserves a minimum cross-axis hit area around the track. The
// rendered track keeps its measured size.
func (s SliderWidget) HitCross(value unit.Dp) SliderWidget {
	s.hitCross = value
	return s
}

// Inset keeps the value mapping away from the track ends, matching a thumb
// that must stay inside the track.
func (s SliderWidget) Inset(value unit.Dp) SliderWidget {
	s.inset = value
	return s
}

// Cursor resolves the pointer cursor for the current state over the whole hit
// area. Returning nil leaves the cursor unchanged.
func (s SliderWidget) Cursor(fn func(SliderState) *pointer.Cursor) SliderWidget {
	s.cursor = fn
	return s
}

func (s SliderWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if s.track == nil {
		panic("flowui: base.Slider requires Track")
	}
	stateKey := core.ClaimKey(ctx, core.KindBaseSlider, s.key)
	state := core.UseState[sliderState](ctx, stateKey, "base-slider")
	tag := state
	core.RegisterFieldFocus(ctx, stateKey, tag, !s.disabled && gtx.Enabled())

	disabled := s.disabled || !gtx.Enabled()
	if state.focusPending && !disabled {
		// RequestFocusVisible also records the target, so the frame's
		// pointer-press clear does not undo the focus command.
		core.RequestFocusVisible(ctx, tag, false)
		gtx.Execute(key.FocusCmd{Tag: tag})
		state.focusPending = false
	}

	ratio := s.ratio(s.value)
	current := s.value
	if !disabled {
		if next, ok := s.keyboard(gtx, tag, current); ok {
			core.RequestFocusVisible(ctx, tag, true)
			current = next
			ratio = s.ratio(next)
			if s.onChange != nil {
				s.onChange(next)
			}
		}
	}

	dragging := !disabled && state.dragging
	focused := gtx.Focused(tag)
	focusVisible := core.FocusVisible(ctx, tag, focused)
	public := SliderState{
		Interaction: Interaction{
			Hovered:      !disabled && state.hovered,
			Pressed:      state.pressed || dragging,
			Focused:      focused,
			FocusVisible: focusVisible,
			Disabled:     disabled,
		},
		Key:         stateKey,
		Value:       current,
		Min:         s.minValue,
		Max:         s.maxValue,
		Ratio:       ratio,
		Orientation: s.orientation,
		Dragging:    dragging,
	}
	track := s.track(public)
	if track == nil {
		panic("flowui: base.Slider Track returned nil")
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

func (s SliderWidget) layoutTrack(ctx *core.Context, gtx layout.Context, state *sliderState, public SliderState, track core.Widget) layout.Dimensions {
	tag := state
	macro := op.Record(gtx.Ops)
	trackDims := track.Layout(ctx, gtx)
	call := macro.Stop()

	axis := s.orientation.axis()
	main := axis.Convert(trackDims.Size).X
	cross := max(axis.Convert(trackDims.Size).Y, gtx.Dp(s.hitCross))
	hitSize := axis.Convert(image.Pt(main, cross))
	inset := min(max(gtx.Dp(s.inset), 0), main/2)
	innerMain := max(main-2*inset, 1)
	dragGtx := gtx
	var cursor *pointer.Cursor
	area := clip.Rect{Max: hitSize}.Push(dragGtx.Ops)
	pointer.CursorDefault.Add(dragGtx.Ops)
	if !public.Disabled {
		event.Op(dragGtx.Ops, tag)
		for {
			e, ok := core.NextPointerEvent(dragGtx, tag, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
			if !ok {
				break
			}
			switch e.Kind {
			case pointer.Press:
				if state.dragging || !core.IsPrimaryPointerPress(e) {
					continue
				}
				state.pressed = true
				state.dragging = true
				state.pointer = e.PointerID
				state.focusPending = true
				core.RequestFocusVisible(ctx, tag, false)
				core.GrabPointer(dragGtx, tag, e)
				s.reportPointerChange(s.valueAt(float64(s.ratioAt(e.Position, axis, innerMain, inset))))
			case pointer.Drag:
				if state.dragging && e.PointerID == state.pointer {
					s.reportPointerChange(s.valueAt(float64(s.ratioAt(e.Position, axis, innerMain, inset))))
				}
			case pointer.Release:
				if state.dragging && e.PointerID == state.pointer {
					state.pressed = false
					state.dragging = false
				}
			case pointer.Cancel:
				state.pressed = false
				state.dragging = false
			}
		}
		state.hover.Add(dragGtx.Ops)

		state.hovered = state.hover.Update(gtx.Source)
		if s.cursor != nil {
			cursorState := public
			cursorState.Pressed = state.pressed
			cursorState.Dragging = state.pressed
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
	if cursor != nil && *cursor != pointer.CursorDefault {
		cursor.Add(dragGtx.Ops)
	}
	area.Pop()
	return trackDims
}

func (s SliderWidget) reportPointerChange(value float64) {
	if s.onChange != nil {
		s.onChange(value)
	}
}

func (s SliderWidget) ratioAt(position f32.Point, axis layout.Axis, innerMain, inset int) float32 {
	main := position.X
	if axis == layout.Vertical {
		main = position.Y
	}
	ratio := (main - float32(inset)) / float32(innerMain)
	if axis == layout.Vertical {
		ratio = 1 - ratio
	}
	return min(max(ratio, 0), 1)
}
func (s SliderWidget) keyboard(gtx layout.Context, tag event.Tag, current float64) (float64, bool) {
	return (SliderKeys{
		Min: s.minValue, Max: s.maxValue, Step: s.effectiveStep(), Snap: s.step > 0,
	}).Update(gtx, tag, current)
}

func (s SliderWidget) effectiveStep() float64 {
	if s.step > 0 {
		return s.step
	}
	return (s.maxValue - s.minValue) / 100
}

func (s SliderWidget) snap(value float64) float64 {
	if s.step <= 0 {
		return value
	}
	return s.minValue + math.Round((value-s.minValue)/s.step)*s.step
}

func (s SliderWidget) clamp(value float64) float64 {
	return clampFloat(value, s.minValue, s.maxValue)
}

func (s SliderWidget) valueAt(ratio float64) float64 {
	return s.clamp(s.snap(s.minValue + ratio*(s.maxValue-s.minValue)))
}

func (s SliderWidget) ratio(value float64) float32 {
	return float32(clampFloat((value-s.minValue)/(s.maxValue-s.minValue), 0, 1))
}

func clampFloat(value, minimum, maximum float64) float64 {
	return min(max(value, minimum), maximum)
}

func (o Orientation) axis() layout.Axis {
	if o == Vertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

type sliderState struct {
	pressed      bool
	dragging     bool
	pointer      pointer.ID
	hover        gesture.Hover
	hovered      bool
	focusPending bool
}
