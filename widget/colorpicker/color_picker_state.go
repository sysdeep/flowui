package colorpicker

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/binding"
)

const stateSlotColorPicker = "color-picker"

func colorPickerStateFor(ctx *core.Context, key string) (string, *colorPickerState) {
	key = core.ClaimKey(ctx, core.KindColorPicker, key)
	return key, core.UseState[colorPickerState](ctx, key, stateSlotColorPicker)
}

type colorPickerState struct {
	valueBinding      binding.Binding[color.NRGBA]
	trigger           widget.Clickable
	triggerFocus      animation.FocusAnimation
	popup             base.PopupInteraction
	open              bool
	popoverTransition animation.FloatTransition
	panelList         layout.List
	color             colorValueState
	defaultsCache     colorPickerDefaultsCache
	history           []color.NRGBA
}

func (pickerState *colorPickerState) pushHistory(value color.NRGBA, limit int) {
	if limit <= 0 {
		return
	}
	// Skip duplicate of most recent.
	if len(pickerState.history) > 0 && pickerState.history[0] == value {
		return
	}
	next := make([]color.NRGBA, 0, limit+1)
	next = append(next, value)
	for _, item := range pickerState.history {
		if item == value {
			continue
		}
		next = append(next, item)
		if len(next) >= limit {
			break
		}
	}
	pickerState.history = next
}

type colorControlState struct {
	base.DragInteraction
	focus animation.FocusAnimation
}

func (pickerState *colorPickerState) popoverProgress(gtx layout.Context, open bool, motion theme.MotionTheme, enter, exit time.Duration) float32 {
	target := float32(0)
	duration := exit
	if open {
		target = 1
		duration = enter
	}
	return pickerState.popoverTransition.Value(gtx, target, duration, animation.EaseSmoothstep, motion)
}

func (control *colorControlState) updateArea(ctx *core.Context, gtx layout.Context, size image.Point, current hsvColor, enabled bool) (hsvColor, bool) {
	next := current
	changed := false
	control.DragInteraction.Update(ctx, gtx, control, enabled, func(position f32.Point, _ bool) {
		next.s, next.v = colorAreaPosition(position, size)
		changed = true
	})

	for {
		value, ok := gtx.Event(
			key.Filter{Focus: control, Name: key.NameLeftArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameRightArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameUpArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameDownArrow, Optional: key.ModShift},
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		step := .01
		if eventValue.Modifiers&key.ModShift != 0 {
			step = .1
		}
		switch eventValue.Name {
		case key.NameLeftArrow:
			next.s -= step
		case key.NameRightArrow:
			next.s += step
		case key.NameUpArrow:
			next.v += step
		case key.NameDownArrow:
			next.v -= step
		}
		next.s = clampUnit(next.s)
		next.v = clampUnit(next.v)
		changed = true
	}
	return next, changed
}

func (control *colorControlState) updateAxis(ctx *core.Context, gtx layout.Context, size image.Point, current, keyboardStep float64, enabled bool) (float64, bool) {
	next := current
	changed := false
	control.DragInteraction.Update(ctx, gtx, control, enabled, func(position f32.Point, _ bool) {
		next = colorAxisPosition(position, size)
		changed = true
	})

	keysGtx := gtx
	if !enabled {
		keysGtx = gtx.Disabled()
	}
	keys := base.SliderKeys{Min: 0, Max: 1, Step: keyboardStep, ShiftScale: 10}
	for {
		value, ok := keys.Update(keysGtx, control, next)
		if !ok {
			break
		}
		next, changed = value, true
	}

	return next, changed
}

func (control *colorControlState) focusOpacity(ctx *core.Context, gtx layout.Context, motion theme.MotionTheme) float32 {
	visible := core.FocusVisible(ctx, control, gtx.Focused(control))
	return control.focus.Opacity(gtx, visible, motion)
}

func colorAreaPosition(position f32.Point, size image.Point) (float64, float64) {
	width := max(size.X-1, 1)
	height := max(size.Y-1, 1)
	return clampUnit(float64(position.X) / float64(width)), clampUnit(1 - float64(position.Y)/float64(height))
}

func colorAxisPosition(position f32.Point, size image.Point) float64 {
	edge := float64(min(size.X, size.Y)) / 2
	length := max(float64(size.X)-edge*2, 1)
	return clampUnit((float64(position.X) - edge) / length)
}

type hsvColor struct {
	h float64
	s float64
	v float64
	a float64
}

func nrgbaToHSV(value color.NRGBA) hsvColor {
	r := float64(value.R) / 255
	g := float64(value.G) / 255
	b := float64(value.B) / 255
	maximum := max(r, g, b)
	minimum := min(r, g, b)
	delta := maximum - minimum
	hue := float64(0)
	if delta != 0 {
		switch maximum {
		case r:
			hue = math.Mod((g-b)/delta, 6)
		case g:
			hue = (b-r)/delta + 2
		default:
			hue = (r-g)/delta + 4
		}
		hue /= 6
		if hue < 0 {
			hue++
		}
	}
	saturation := float64(0)
	if maximum != 0 {
		saturation = delta / maximum
	}
	return hsvColor{h: hue, s: saturation, v: maximum, a: float64(value.A) / 255}
}

func hsvToNRGBA(value hsvColor) color.NRGBA {
	hue := value.h - math.Floor(value.h)
	saturation := clampUnit(value.s)
	brightness := clampUnit(value.v)
	chroma := brightness * saturation
	section := hue * 6
	x := chroma * (1 - math.Abs(math.Mod(section, 2)-1))
	var r, g, b float64
	switch int(section) % 6 {
	case 0:
		r, g = chroma, x
	case 1:
		r, g = x, chroma
	case 2:
		g, b = chroma, x
	case 3:
		g, b = x, chroma
	case 4:
		r, b = x, chroma
	default:
		r, b = chroma, x
	}
	match := brightness - chroma
	channel := func(value float64) uint8 {
		return uint8(clampUnit(value+match)*255 + .5)
	}
	return color.NRGBA{R: channel(r), G: channel(g), B: channel(b), A: uint8(clampUnit(value.a)*255 + .5)}
}

func clampUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(max(value, 0), 1)
}

// Disclosure helpers
func colorPickerBindingConfig(widget Widget) binding.Config[color.NRGBA] {
	return binding.Config[color.NRGBA]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *colorPickerState) currentValue(widget Widget) color.NRGBA {
	return s.valueBinding.Current(colorPickerBindingConfig(widget))
}

func (s *colorPickerState) bind(widget Widget) {
	s.valueBinding.Bind(colorPickerBindingConfig(widget))
}

func (s *colorPickerState) requestValue(widget Widget, value color.NRGBA) color.NRGBA {
	newValue, _ := s.valueBinding.Request(colorPickerBindingConfig(widget), value)
	return newValue
}
