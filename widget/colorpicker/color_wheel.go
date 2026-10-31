package colorpicker

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

const stateSlotColorWheel = "color-wheel"

type ColorWheelWidget struct {
	style    style.Style
	key      string
	value    color.NRGBA
	label    string
	size     int
	hasSize  bool
	onChange func(color.NRGBA)
	disabled bool
}

type colorWheelState struct {
	control       colorControlState
	color         colorValueState
	defaultsCache colorWheelDefaultsCache
	imageCache    colorWheelImageCache
}

func ColorWheel(key string, value color.NRGBA) ColorWheelWidget {
	return ColorWheelWidget{key: key, value: value}
}

func (wheel ColorWheelWidget) Size(size int) ColorWheelWidget {
	wheel.size = max(size, 0)
	wheel.hasSize = true
	return wheel
}

func (wheel ColorWheelWidget) Label(label string) ColorWheelWidget {
	wheel.label = label
	return wheel
}

func (wheel ColorWheelWidget) OnChange(fn func(color.NRGBA)) ColorWheelWidget {
	wheel.onChange = fn
	return wheel
}

func (wheel ColorWheelWidget) Disabled(disabled bool) ColorWheelWidget {
	wheel.disabled = disabled
	return wheel
}

func (wheel ColorWheelWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return wheel.layoutWithStyle(ctx, gtx)
}

func (wheel ColorWheelWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorWheel, wheel.key)
	wheelState := core.UseState[colorWheelState](ctx, key, stateSlotColorWheel)
	defaults := colorWheelDefaultsFor(ctx, &wheelState.defaultsCache)
	wheelState.color.sync(wheel.value)

	enabled := gtx.Enabled() && !wheel.disabled
	diameter := defaults.Size
	if wheel.hasSize {
		diameter = unit.Dp(wheel.size)
	}
	side := min(gtx.Dp(diameter), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(side, 0), max(side, 0))
	wheelGtx := gtx
	wheelGtx.Constraints = layout.Exact(size)
	current := wheelState.color.hsv()
	if next, changed := wheelState.control.updateWheel(ctx, wheelGtx, size, current, enabled); changed {
		nextColor := hsvToNRGBA(next)
		wheelState.color.accept(nextColor, next.h)
		if wheel.onChange != nil && nextColor != wheel.value {
			wheel.onChange(nextColor)
		}
		current = next
	}

	focused := gtx.Focused(&wheelState.control)
	state := style.State{
		Pressed:      wheelState.control.Dragging,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &wheelState.control, focused),
		Disabled:     !enabled,
		Dragging:     wheelState.control.Dragging,
	}
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults.Defaults,
		style.Style{},
		style.Style{},
		wheel.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, colorControlOpacity(enabled, defaults.DisabledOpacity))
		colorDrawColorWheel(
			gtx,
			size,
			current,
			&wheelState.imageCache,
			max(gtx.Dp(defaults.ThumbSize), 1),
			max(gtx.Dp(defaults.ThumbBorderWidth), 1),
			max(gtx.Dp(defaults.FocusRingWidth), 1),
			wheelState.control.focusOpacity(ctx, gtx, defaults.Motion),
			defaults.FocusColor,
		)
		opacity.Pop()
		wheelState.control.AddInput(gtx, &wheelState.control, image.Rectangle{Max: size}, enabled, pointer.CursorCrosshair, wheel.semanticLabel(ctx), formatHexColor(wheel.value, wheel.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (wheel ColorWheelWidget) semanticLabel(ctx *core.Context) string {
	if wheel.label != "" {
		return wheel.label
	}
	return ctx.Text(i18n.ColorPickerWheel)
}

func (control *colorControlState) updateWheel(ctx *core.Context, gtx layout.Context, size image.Point, current hsvColor, enabled bool) (hsvColor, bool) {
	next := current
	changed := false
	control.DragInteraction.Update(ctx, gtx, control, enabled, func(position f32.Point, _ bool) {
		next = colorWheelPosition(position, size, next)
		changed = true
	})

	for {
		value, ok := gtx.Event(
			key.Filter{Focus: control, Name: key.NameLeftArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameRightArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameUpArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameDownArrow, Optional: key.ModShift},
			key.Filter{Focus: control, Name: key.NameHome},
			key.Filter{Focus: control, Name: key.NameEnd},
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		hueStep := 1.0 / 360
		saturationStep := .01
		if eventValue.Modifiers&key.ModShift != 0 {
			hueStep *= 10
			saturationStep *= 10
		}
		switch eventValue.Name {
		case key.NameLeftArrow:
			next.h = wrapUnit(next.h - hueStep)
		case key.NameRightArrow:
			next.h = wrapUnit(next.h + hueStep)
		case key.NameUpArrow:
			next.s = clampUnit(next.s + saturationStep)
		case key.NameDownArrow:
			next.s = clampUnit(next.s - saturationStep)
		case key.NameHome:
			next.s = 0
		case key.NameEnd:
			next.s = 1
		}
		changed = true
	}
	return next, changed
}

func colorWheelPosition(position f32.Point, size image.Point, current hsvColor) hsvColor {
	radius := float64(min(size.X, size.Y)) / 2
	if radius <= 0 {
		return current
	}
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	dx := float64(position.X - center.X)
	dy := float64(position.Y - center.Y)
	distance := math.Hypot(dx, dy)
	current.s = clampUnit(distance / radius)
	if distance > 0 {
		current.h = wrapUnit(math.Atan2(dy, dx) / (2 * math.Pi))
	}
	return current
}

func wrapUnit(value float64) float64 {
	value -= math.Floor(value)
	if value < 0 {
		value++
	}
	return value
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w ColorWheelWidget) Style(values ...style.Style) ColorWheelWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
