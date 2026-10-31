package colorpicker

import (
	"image"
	"image/color"
	"slices"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

const (
	stateSlotColorSwatchPicker = "color-swatch-picker"
)

type ColorSwatchPickerWidget struct {
	style          style.Style
	key            string
	value          color.NRGBA
	colors         []color.NRGBA
	disabledColors []color.NRGBA
	onChange       func(color.NRGBA)
	size           ColorSwatchSize
	shape          ColorSwatchShape
	arrangement    ColorSwatchPickerLayout
	disabled       bool
}

type colorSwatchPickerState struct {
	items         map[colorSwatchItemKey]*colorSwatchItemState
	frameItems    map[colorSwatchItemKey]struct{}
	defaultsCache colorSwatchPickerDefaultsCache
}

type colorSwatchItemKey struct {
	index int
	value color.NRGBA
}

type colorSwatchItemState struct {
	clickable widget.Clickable
	focus     animation.FocusAnimation
	selection animation.FloatTransition
	check     animation.FloatTransition
	hover     animation.FloatTransition
}

type recordedColorSwatch struct {
	call op.CallOp
	size image.Point
}

func ColorSwatchPicker(key string, value color.NRGBA, colors []color.NRGBA) ColorSwatchPickerWidget {
	return ColorSwatchPickerWidget{
		key:         key,
		value:       value,
		colors:      append([]color.NRGBA(nil), colors...),
		size:        ColorSwatchMedium,
		shape:       ColorSwatchCircle,
		arrangement: ColorSwatchPickerGrid,
	}
}

func (picker ColorSwatchPickerWidget) OnChange(fn func(color.NRGBA)) ColorSwatchPickerWidget {
	picker.onChange = fn
	return picker
}

func (picker ColorSwatchPickerWidget) Size(size ColorSwatchSize) ColorSwatchPickerWidget {
	picker.size = size
	return picker
}

func (picker ColorSwatchPickerWidget) Shape(shape ColorSwatchShape) ColorSwatchPickerWidget {
	picker.shape = shape
	return picker
}

func (picker ColorSwatchPickerWidget) Arrangement(arrangement ColorSwatchPickerLayout) ColorSwatchPickerWidget {
	picker.arrangement = arrangement
	return picker
}

func (picker ColorSwatchPickerWidget) Disabled(disabled bool) ColorSwatchPickerWidget {
	picker.disabled = disabled
	return picker
}

func (picker ColorSwatchPickerWidget) DisabledColors(values []color.NRGBA) ColorSwatchPickerWidget {
	picker.disabledColors = append([]color.NRGBA(nil), values...)
	return picker
}

func (picker ColorSwatchPickerWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return picker.layoutWithStyle(ctx, gtx)
}

func (picker ColorSwatchPickerWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorSwatchPicker, picker.key)
	pickerState := core.UseState[colorSwatchPickerState](ctx, key, stateSlotColorSwatchPicker)
	defaults := colorSwatchPickerDefaultsFor(ctx, &pickerState.defaultsCache)
	if pickerState.items == nil {
		pickerState.items = make(map[colorSwatchItemKey]*colorSwatchItemState)
	}
	if pickerState.frameItems == nil {
		pickerState.frameItems = make(map[colorSwatchItemKey]struct{})
	} else {
		clear(pickerState.frameItems)
	}
	defer func() {
		for itemKey := range pickerState.items {
			if _, visible := pickerState.frameItems[itemKey]; !visible {
				delete(pickerState.items, itemKey)
			}
		}
	}()

	gap := max(gtx.Dp(defaults.Gap), 0)
	recorded := make([]recordedColorSwatch, 0, len(picker.colors))
	hovered, pressed := false, false
	for index, value := range picker.colors {
		itemKey := colorSwatchItemKey{index: index, value: value}
		pickerState.frameItems[itemKey] = struct{}{}
		itemState := pickerState.items[itemKey]
		if itemState == nil {
			itemState = new(colorSwatchItemState)
			pickerState.items[itemKey] = itemState
		}
		recorded = append(recorded, picker.recordItem(ctx, gtx, itemState, defaults, value, picker.colorDisabled(value)))
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
	}
	state := style.State{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  false,
		Disabled: picker.disabled || !gtx.Enabled(),
		Selected: slices.Contains(picker.colors, picker.value),
	}
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults.Defaults,
		style.Style{},
		style.Style{},
		picker.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return picker.layoutItems(gtx, recorded, gap)
	}))
}

func (picker ColorSwatchPickerWidget) recordItem(ctx *core.Context, gtx layout.Context, itemState *colorSwatchItemState, defaults colorSwatchPickerDefaults, value color.NRGBA, disabled bool) recordedColorSwatch {
	enabled := gtx.Enabled() && !picker.disabled && !disabled
	base.RowInteraction(ctx, gtx, &itemState.clickable, !enabled, func() {
		if value != picker.value && picker.onChange != nil {
			picker.onChange(value)
		}
	})

	macro := op.Record(gtx.Ops)
	itemGtx := gtx
	side := min(max(colorSwatchPixels(itemGtx, defaults.Swatch, picker.size), 0), min(itemGtx.Constraints.Max.X, itemGtx.Constraints.Max.Y))
	itemSize := image.Pt(side, side)
	itemGtx.Constraints = layout.Exact(itemSize)
	if !enabled {
		itemGtx = itemGtx.Disabled()
	}
	dimensions := itemState.clickable.Layout(itemGtx, func(gtx layout.Context) layout.Dimensions {
		if enabled {
			clipped := clip.Rect{Max: itemSize}.Push(gtx.Ops)
			pointer.CursorPointer.Add(gtx.Ops)
			clipped.Pop()
		}
		semantic.RadioButton.Add(gtx.Ops)
		semantic.LabelOp(formatHexColor(value, value.A != 255)).Add(gtx.Ops)
		semantic.SelectedOp(value == picker.value).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		selection := itemState.selection.Value(gtx, boolFloat(value == picker.value), defaults.SelectionDuration, animation.EaseSmoothstep, defaults.Motion)
		check := itemState.check.Value(gtx, boolFloat(value == picker.value), defaults.CheckDuration, animation.EaseSmoothstep, defaults.Motion)
		hover := itemState.hover.Value(gtx, boolFloat(itemState.clickable.Hovered() && enabled), defaults.HoverDuration, animation.EaseSmoothstep, defaults.Motion)
		focusVisible := core.FocusVisible(ctx, &itemState.clickable, gtx.Focused(&itemState.clickable))
		focus := itemState.focus.Opacity(gtx, focusVisible && enabled, defaults.Motion)
		opacity := paint.PushOpacity(gtx.Ops, colorControlOpacity(enabled, defaults.DisabledOpacity))
		colorDrawColorSwatchPickerItem(
			gtx,
			defaults,
			itemSize,
			value,
			ctx.BackgroundColor(),
			uint8(picker.size),
			picker.shape,
			selection,
			check,
			hover,
			focus,
		)
		opacity.Pop()
		return layout.Dimensions{Size: itemSize}
	})
	return recordedColorSwatch{call: macro.Stop(), size: dimensions.Size}
}

func (picker ColorSwatchPickerWidget) layoutItems(gtx layout.Context, items []recordedColorSwatch, gap int) layout.Dimensions {
	x, y, rowHeight, width := 0, 0, 0, 0
	maxWidth := max(gtx.Constraints.Max.X, 0)
	for index, item := range items {
		if picker.arrangement == ColorSwatchPickerStack {
			if index > 0 {
				y += rowHeight + gap
			}
			x = 0
			rowHeight = 0
		} else if x > 0 && x+item.size.X > maxWidth {
			x = 0
			y += rowHeight + gap
			rowHeight = 0
		}
		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		item.call.Add(gtx.Ops)
		offset.Pop()
		width = max(width, x+item.size.X)
		rowHeight = max(rowHeight, item.size.Y)
		if picker.arrangement != ColorSwatchPickerStack {
			x += item.size.X + gap
		}
	}
	height := y + rowHeight
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
}

func (picker ColorSwatchPickerWidget) colorDisabled(value color.NRGBA) bool {
	return slices.Contains(picker.disabledColors, value)
}

func boolFloat(value bool) float32 {
	if value {
		return 1
	}
	return 0
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w ColorSwatchPickerWidget) Style(values ...style.Style) ColorSwatchPickerWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
