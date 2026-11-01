package colorpicker

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type Widget struct {
	style        style.Style
	key          string
	value        color.NRGBA
	hasValue     bool
	defaultValue color.NRGBA
	hasDefault   bool
	label        string
	onChange     func(color.NRGBA)
	disabled     bool
	alpha        bool
	showField    bool
	showRGB      bool
	showHistory  bool
	historySize  int
	presets      []color.NRGBA
}

const defaultColorHistorySize = 8

func New(value color.NRGBA) Widget {
	return Widget{
		key:         core.AutoKey("color-picker"),
		value:       value,
		hasValue:    true,
		showField:   true,
		showRGB:     true,
		showHistory: true,
		historySize: defaultColorHistorySize,
	}
}

func (picker Widget) Key(key string) Widget {
	picker.key = key
	return picker
}

func (picker Widget) Value(value color.NRGBA) Widget {
	picker.value = value
	picker.hasValue = true
	return picker
}

func (picker Widget) DefaultValue(value color.NRGBA) Widget {
	picker.defaultValue = value
	picker.hasDefault = true
	picker.hasValue = false
	return picker
}

func (picker Widget) Label(label string) Widget {
	picker.label = label
	return picker
}

func (picker Widget) OnChange(fn func(color.NRGBA)) Widget {
	picker.onChange = fn
	return picker
}

func (picker Widget) changeHandler(state *colorPickerState, recordHistory bool) func(color.NRGBA) {
	return func(value color.NRGBA) {
		picker.reportChange(state, value, recordHistory)
	}
}

func (picker Widget) Disabled(disabled bool) Widget {
	picker.disabled = disabled
	return picker
}

func (picker Widget) Alpha(enabled bool) Widget {
	picker.alpha = enabled
	return picker
}

// ShowField shows a hex text field. Defaults to true for desktop pickers.
func (picker Widget) ShowField() Widget {
	picker.showField = true
	return picker
}

// HideField hides the hex text field.
func (picker Widget) HideField() Widget {
	picker.showField = false
	return picker
}

// ShowRGB shows R/G/B (and A when Alpha is enabled) channel fields.
// Defaults to true for desktop pickers.
func (picker Widget) ShowRGB(show bool) Widget {
	picker.showRGB = show
	return picker
}

// ShowHistory shows recent colors chosen in this picker instance.
// Defaults to true for desktop pickers.
func (picker Widget) ShowHistory(show bool) Widget {
	picker.showHistory = show
	return picker
}

// HistorySize sets how many recent colors to keep (default 8, max 16).
func (picker Widget) HistorySize(size int) Widget {
	if size < 0 {
		size = 0
	}
	if size > 16 {
		size = 16
	}
	picker.historySize = size
	return picker
}

func (picker Widget) Presets(values []color.NRGBA) Widget {
	picker.presets = append([]color.NRGBA(nil), values...)
	return picker
}

func (picker Widget) reportChange(state *colorPickerState, value color.NRGBA, recordHistory bool) {
	// Keep prior hue only for achromatic colors so presets/hex/RGB stay aligned
	// with the hue slider and saturation area.
	resolved := nrgbaToHSV(value)
	hue := state.color.hsv().h
	if resolved.s > 0 && resolved.v > 0 {
		hue = resolved.h
	}
	state.color.accept(value, hue)
	if recordHistory && picker.showHistory {
		state.pushHistory(value, picker.historySize)
	}
	state.requestValue(picker, value)
}

func (picker Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return picker.layoutWithStyle(ctx, gtx)
}

func (picker Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key, pickerState := colorPickerStateFor(ctx, picker.key)
	defaults := colorPickerDefaultsFor(ctx, &pickerState.defaultsCache)

	// Bind disclosure and get current value
	pickerState.bind(picker)
	currentValue := pickerState.currentValue(picker)
	picker.value = currentValue

	pickerState.color.sync(currentValue)

	enabled := gtx.Enabled() && !picker.disabled
	wasOpen := pickerState.open
	if !enabled {
		pickerState.open = false
	}
	// Commit the last color to history when the popover closes (not during drag).
	if wasOpen && !pickerState.open && picker.showHistory {
		pickerState.pushHistory(pickerState.color.syncedColor, picker.historySize)
	}

	triggerGtx := gtx
	if !enabled {
		triggerGtx = triggerGtx.Disabled()
	}
	state := style.State{
		Hovered:  pickerState.trigger.Hovered(),
		Pressed:  pickerState.trigger.Pressed(),
		Focused:  triggerGtx.Focused(&pickerState.trigger),
		Disabled: !enabled,
		Open:     pickerState.open,
	}
	resolved := computed.Resolve(
		ctx,
		triggerGtx,
		key,
		state,
		defaults.Defaults,
		style.Style{},
		style.Style{},
		picker.style,
	)
	dimensions := layoutui.Resolved(ctx, triggerGtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return picker.layoutTrigger(ctx, gtx, pickerState, defaults, enabled)
	}))
	progress := pickerState.popoverProgress(gtx, pickerState.open && enabled, defaults.Motion, defaults.EnterDuration, defaults.ExitDuration)
	if progress == 0 && !pickerState.open {
		return dimensions
	}

	picker.layoutPopover(ctx, gtx, pickerState, defaults, key, dimensions, progress, core.OverlayNaturallyDisabled(gtx))
	return dimensions
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
