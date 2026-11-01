package colorpicker

import (
	"image"
	"image/color"
	"slices"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/input"
)

func (picker Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults, enabled bool) layout.Dimensions {
	base.RowInteraction(ctx, gtx, &pickerState.trigger, !enabled, func() {
		pickerState.open = !pickerState.open
	})

	return pickerState.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(picker.semanticLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		dimensions := picker.layoutTriggerContent(ctx, gtx, defaults)
		call := macro.Stop()
		if !enabled {
			opacity := paint.PushOpacity(gtx.Ops, defaults.DisabledOpacity)
			call.Add(gtx.Ops)
			opacity.Pop()
		} else {
			call.Add(gtx.Ops)
		}
		focusVisible := core.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger))
		focus := pickerState.triggerFocus.Opacity(gtx, focusVisible && enabled, defaults.Motion)
		colorPickerTriggerFocus(gtx, defaults, dimensions.Size, focus)
		return dimensions
	})
}

func (picker Widget) layoutTriggerContent(ctx *core.Context, gtx layout.Context, defaults colorPickerDefaults) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorSwatch(picker.value).Size(ColorSwatchLarge).Layout(ctx, gtx)
		}),
	}
	if picker.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return colorPickerTriggerLabel(gtx, defaults, picker.label)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(defaults.TriggerGap),
	}.Layout(gtx, children...)
}

func (picker Widget) layoutPopover(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults, key string, trigger layout.Dimensions, progress float32, naturallyDisabled bool) {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: trigger.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			intent := pickerState.popup.Update(ctx, gtx, base.PopupPolicy{
				Open: pickerState.open, Interactive: interactive, DismissOutside: true,
				DismissOnPress: true, DismissEscape: true,
			})
			if intent != base.PopupNotDismissed {
				if picker.showHistory {
					pickerState.pushHistory(pickerState.color.syncedColor, picker.historySize)
				}
				pickerState.open = false
				if intent == base.PopupDismissEscape {
					core.RequestFocusVisible(ctx, &pickerState.trigger, true)
				}
			}

			panelGtx := gtx
			// Keep the open panel enabled even when a nested overlay (such as
			// an input's context menu) owns policy input. Disabling it would
			// stop that overlay from being registered here, so it would
			// disappear on every frame it is not the event owner.
			if !pickerState.open || !gtx.Enabled() {
				panelGtx = panelGtx.Disabled()
			}
			return picker.layoutOverlay(ctx, gtx, panelGtx, pickerState, defaults, anchor, progress)
		},
	})
}

func (picker Widget) layoutOverlay(ctx *core.Context, gtx, panelGtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults, anchor image.Rectangle, progress float32) layout.Dimensions {
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(defaults.PanelGap)
	panelWidth := min(gtx.Dp(defaults.PanelWidth), max(viewport.X, 0))
	panelHeight := min(gtx.Dp(defaults.PanelMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelHeight <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelHeight),
	}

	macro := op.Record(gtx.Ops)
	dimensions, tracked := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return picker.layoutPanel(ctx, panelGtx, pickerState, defaults)
	})
	call := macro.Stop()
	placement := core.Placement{Side: core.SideBottom, Align: core.AlignStart}
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            dimensions.Size,
		Bounds:           viewport,
		Offset:           gap,
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := core.PanelTransformOriginAt(anchor, result.Position, dimensions.Size, result.Placement)
	scale := defaults.EnterScale + (1-defaults.EnterScale)*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform)
	tracked.PlaceTransform(panelTransform)
	tracked.SetOpacity(progress)
	pickerState.popup.LayoutOutside(gtx, viewport, true, anchor, core.AffineRectBounds(image.Rectangle{Max: dimensions.Size}, panelTransform))

	offset := op.Offset(result.Position).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	transform := op.Affine(scaleTransform).Push(gtx.Ops)
	pickerState.popup.LayoutSurface(gtx, image.Rectangle{Max: dimensions.Size})
	call.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	offset.Pop()
	return dimensions
}

func (picker Widget) layoutPanel(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults) layout.Dimensions {
	restoreColors := core.PushColors(ctx, defaults.PanelForeground, defaults.PanelBackground)
	defer restoreColors()
	pickerState.panelList.Axis = layout.Vertical
	macro := op.Record(gtx.Ops)
	dimensions := pickerState.panelList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return flowlayout.TrackedInset(ctx, gtx, layout.UniformInset(defaults.PanelPadding), func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return picker.layoutPanelContent(ctx, gtx, pickerState, defaults)
		})
	})
	call := macro.Stop()
	radius := min(max(gtx.Dp(defaults.PanelRadius), 1), min(dimensions.Size.X, dimensions.Size.Y)/2)
	draw.Surface(gtx, image.Rectangle{Max: dimensions.Size}, radius, defaults.PanelBackground, defaults.PanelShadow)
	clipped := clip.UniformRRect(image.Rectangle{Max: dimensions.Size}, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipped.Pop()
	return dimensions
}

func (picker Widget) layoutPanelContent(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults) layout.Dimensions {
	restore := core.PushKey(ctx, picker.key)
	defer restore()
	// Continuous drag (area/slider) must not flood recent-history; discrete
	// commits (hex, RGB, presets, history swatches) record history.
	dragChange := picker.changeHandler(pickerState, false)
	commitChange := picker.changeHandler(pickerState, true)
	var children [8]layout.Widget
	count := 0
	children[count] = func(gtx layout.Context) layout.Dimensions {
		return ColorArea(picker.value).Key("area").
			withColorState(&pickerState.color).
			OnChange(dragChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	}
	count++
	children[count] = func(gtx layout.Context) layout.Dimensions {
		return ColorSlider(picker.value, ColorChannelHue).Key("hue").
			withColorState(&pickerState.color).
			OnChange(dragChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	}
	count++
	if picker.alpha {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSlider(picker.value, ColorChannelAlpha).Key("alpha").
				withColorState(&pickerState.color).
				OnChange(dragChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	if picker.showRGB {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return picker.layoutRGBChannels(ctx, gtx, pickerState, defaults, commitChange)
		}
		count++
	}
	if picker.showField {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorField(picker.value).Key("field").
				Alpha(picker.alpha).
				Swatch(true).
				Variant(input.Secondary).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	if len(picker.presets) > 0 {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSwatchPicker(picker.value, picker.presets).Key("presets").
				Size(ColorSwatchExtraSmall).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	gap := defaults.ContentGap
	if picker.alpha || picker.showField || picker.showRGB || len(picker.presets) > 0 || picker.showHistory {
		gap = defaults.CompactContentGap
	}
	history := historyWithoutPresets(pickerState.history, picker.presets)
	if picker.showHistory && len(history) > 0 {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSwatchPicker(picker.value, history).Key("history").
				Size(ColorSwatchExtraSmall).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	return flowlayout.Items(ctx, gtx, false, 0, gtx.Dp(gap), children[:count])
}

func historyWithoutPresets(history, presets []color.NRGBA) []color.NRGBA {
	if len(history) == 0 || len(presets) == 0 {
		return history
	}
	filtered := make([]color.NRGBA, 0, len(history))
	for _, recent := range history {
		preset := slices.Contains(presets, recent)
		if !preset {
			filtered = append(filtered, recent)
		}
	}
	return filtered
}

func (picker Widget) layoutRGBChannels(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, defaults colorPickerDefaults, onChange func(color.NRGBA)) layout.Dimensions {
	value := pickerState.color.syncedColor
	if !pickerState.color.ready {
		value = picker.value
	}
	type channelSpec struct {
		key   string
		label string
		get   func(color.NRGBA) byte
		set   func(color.NRGBA, byte) color.NRGBA
	}
	channels := []channelSpec{
		{"r", "R", func(c color.NRGBA) byte { return c.R }, func(c color.NRGBA, v byte) color.NRGBA { c.R = v; return c }},
		{"g", "G", func(c color.NRGBA) byte { return c.G }, func(c color.NRGBA, v byte) color.NRGBA { c.G = v; return c }},
		{"b", "B", func(c color.NRGBA) byte { return c.B }, func(c color.NRGBA, v byte) color.NRGBA { c.B = v; return c }},
	}
	if picker.alpha {
		channels = append(channels, channelSpec{
			"a", "A",
			func(c color.NRGBA) byte { return c.A },
			func(c color.NRGBA, v byte) color.NRGBA { c.A = v; return c },
		})
	}
	children := make([]layout.Widget, 0, len(channels))
	for _, channel := range channels {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return picker.layoutChannelField(ctx, gtx, channel.key, channel.label, channel.get(value), func(next byte) {
				onChange(channel.set(value, next))
			})
		})
	}
	return layoutTrackedRow(ctx, gtx, gtx.Dp(defaults.ChannelGap), children)
}

// layoutTrackedRow lays out equal-width children in a row while recording each
// child's horizontal offset, so overlays anchored inside them (such as an
// input's context menu) resolve to the right viewport position.
func layoutTrackedRow(ctx *core.Context, gtx layout.Context, gap int, children []layout.Widget) layout.Dimensions {
	if len(children) == 0 {
		return layout.Dimensions{}
	}
	width := (gtx.Constraints.Max.X - gap*(len(children)-1)) / len(children)
	width = max(width, 0)
	type placed struct {
		call      op.CallOp
		dims      layout.Dimensions
		placement core.OverlayPlacement
		x         int
	}
	calls := make([]placed, 0, len(children))
	height := 0
	for index, child := range children {
		childGtx := gtx
		childGtx.Constraints.Min.X = width
		childGtx.Constraints.Max.X = width
		macro := op.Record(gtx.Ops)
		dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(childGtx)
		})
		call := macro.Stop()
		calls = append(calls, placed{call: call, dims: dims, placement: placement, x: index * (width + gap)})
		height = max(height, dims.Size.Y)
	}
	for _, child := range calls {
		offset := image.Pt(child.x, height-child.dims.Size.Y)
		child.placement.PlaceOffset(offset)
		push := op.Offset(offset).Push(gtx.Ops)
		child.call.Add(gtx.Ops)
		push.Pop()
	}
	return layout.Dimensions{Size: image.Pt(width*len(children)+gap*(len(children)-1), height)}
}

func (picker Widget) layoutChannelField(ctx *core.Context, gtx layout.Context, key, label string, value byte, onChange func(byte)) layout.Dimensions {
	restore := core.PushKey(ctx, key)
	defer restore()
	control := input.New(itoaChannel(value)).Key("value").
		Placeholder("0").
		Label(label).
		Variant(input.Secondary).
		Disabled(picker.disabled).
		OnChange(func(text string) {
			if parsed, ok := parseChannel(text); ok {
				onChange(parsed)
			}
		})
	return control.Layout(ctx, gtx)
}

func itoaChannel(value byte) string {
	if value == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func parseChannel(text string) (byte, bool) {
	text = trimSpaceASCII(text)
	// Incomplete input (empty or still typing) must not commit a value, or
	// clearing the field while typing "128" would snap to 0 mid-edit.
	if text == "" {
		return 0, false
	}
	var n int
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
		if n > 255 {
			return 0, false
		}
	}
	return byte(n), true
}

func trimSpaceASCII(value string) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
}

func (picker Widget) semanticLabel(ctx *core.Context) string {
	if picker.label != "" {
		return picker.label
	}
	return ctx.Text(i18n.ColorPickerPickColor)
}
