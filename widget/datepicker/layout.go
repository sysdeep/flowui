package datepicker

import (
	"fmt"
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/text"
)

func (d Widget) layoutField(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle resolvedDateField, metrics datePickerGeometry, enabled, invalid bool) (layout.Dimensions, image.Rectangle) {
	var children [3]layout.FlexChild
	count := 0
	labelHeight := 0
	inputAnchor := image.Rectangle{}
	gap := gtx.Dp(metrics.FieldGap)
	descriptionStyle := d.resolvePart(ctx, gtx, style.PartDescription, style.State{
		Disabled: !enabled,
		Invalid:  invalid,
	}, &pickerState.calendarStyle)
	if d.label != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dimensions := text.NewLabel(d.label).
				For(d.key).
				Required(d.required).
				Disabled(!enabled).
				Invalid(invalid).
				Layout(ctx, gtx)
			labelHeight = dimensions.Size.Y
			return dimensions
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dimensions := d.layoutInput(ctx, gtx, pickerState, fieldStyle, metrics, enabled, invalid)
		y := labelHeight
		if d.label != "" {
			y += gap
		}
		inputAnchor = image.Rectangle{Min: image.Pt(0, y), Max: image.Pt(dimensions.Size.X, y+dimensions.Size.Y)}
		pickerState.hover.AddInput(gtx, &pickerState.hover, image.Rectangle{Max: dimensions.Size}, enabled, nil)
		return dimensions
	})
	count++
	if invalid && d.errorMessage != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(d.errorMessage).
				Style(style.TextDeclaration(descriptionStyle.Text)).
				Layout(ctx, gtx)
		})
		count++
	} else if d.description != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return (text.NewDescription(d.description).
				For(d.key).
				Disabled(!enabled)).Style(style.TextDeclaration(descriptionStyle.Text)).Layout(ctx, gtx)
		})
		count++
	}
	dimensions := layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children[:count]...)
	return dimensions, inputAnchor
}

func (d Widget) layoutInput(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle resolvedDateField, metrics datePickerGeometry, enabled, invalid bool) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, fieldStyle.Content, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutInputContent(ctx, gtx, pickerState, fieldStyle, metrics, enabled, invalid)
	}))
}

func (d Widget) layoutInputContent(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle resolvedDateField, metrics datePickerGeometry, enabled, invalid bool) layout.Dimensions {
	colors := fieldStyle.Colors
	base.RowInteraction(ctx, gtx, &pickerState.trigger, !enabled, func() {
		if pickerState.open {
			pickerState.open = false
		} else {
			pickerState.openCalendar()
		}
	})

	frameConstraints := gtx.Constraints
	if d.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	height := min(gtx.Dp(metrics.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	left := gtx.Dp(metrics.FieldPaddingX)
	right := gtx.Dp(metrics.TriggerWidth)
	horizontalPadding := left + right
	maxX := max(frameConstraints.Max.X-horizontalPadding, 0)
	minX := 0
	if !d.editable {
		minX = min(max(frameConstraints.Min.X-horizontalPadding, 0), maxX)
	}

	macro := op.Record(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, height),
	}
	var contentDims layout.Dimensions
	showHint := false
	if d.editable {
		contentDims = pickerState.segments.layout(ctx, contentGtx, d.locale, dateSegmentsLayoutOptions{
			key:           core.FullKey(ctx, d.key),
			instanceStyle: d.style,
			calendarStyle: &pickerState.calendarStyle,
			resolvedParts: &pickerState.resolvedParts,
			colors:        colors,
			metrics:       metrics,
			enabled:       enabled,
			invalid:       invalid,
			minDate:       d.minDate,
			maxDate:       d.maxDate,
		}, func(value time.Time) {
			pickerState.requestValue(d, value)
		})
		showHint = d.hintSet && pickerState.segments.empty() && !pickerState.segments.focused(gtx)
	} else {
		value := d.hint
		color := colors.Placeholder
		if !d.value.IsZero() {
			value = d.locale.DateLabel(d.value)
			color = colors.Foreground
		}
		contentGtx.Constraints.Min = image.Point{}
		contentDims = text.New(value).
			Size(float32(metrics.TextSize)).
			Color(color).
			MaxLines(1).
			Truncator("...").
			Layout(ctx, contentGtx)
	}
	call := macro.Stop()
	var hintCall op.CallOp
	var hintDims layout.Dimensions
	if showHint {
		hintMacro := op.Record(gtx.Ops)
		hintColor := colors.Placeholder
		if invalid {
			hintColor = d.datePickerDescriptionColor(ctx, gtx, true, enabled, &pickerState.calendarStyle)
		}
		hintGtx := contentGtx
		hintGtx.Constraints.Min = image.Point{}
		hintDims = text.New(d.hint).
			Size(float32(metrics.TextSize)).
			Color(hintColor).
			MaxLines(1).
			Truncator("...").
			Layout(ctx, hintGtx)
		hintCall = hintMacro.Stop()
		contentDims.Size.X = max(contentDims.Size.X, hintDims.Size.X)
		contentDims.Size.Y = max(contentDims.Size.Y, hintDims.Size.Y)
	}

	size := frameConstraints.Constrain(image.Pt(contentDims.Size.X+horizontalPadding, height))

	// The disclosure owns the whole field hit area. Date segments are emitted
	// after it so their own click targets remain above the field trigger and
	// continue to focus/edit the selected segment.
	triggerX := 0
	triggerSize := size
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	stack := op.Offset(image.Pt(triggerX, 0)).Push(gtx.Ops)
	if !enabled {
		triggerGtx = triggerGtx.Disabled()
	}
	pickerState.trigger.Layout(triggerGtx, func(gtx layout.Context) layout.Dimensions {
		if enabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(d.calendarLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		iconSize := image.Pt(gtx.Dp(metrics.IconSize), gtx.Dp(metrics.IconSize))
		iconOffset := op.Offset(image.Pt(size.X-right+(right-iconSize.X)/2, (size.Y-iconSize.Y)/2)).Push(gtx.Ops)
		dateDrawCalendarIcon(gtx, iconSize, colors.Placeholder)
		iconOffset.Pop()
		if d.editable {
			focusColor := dateResolveOutlineColor(fieldStyle.Content, core.ActiveTheme(ctx).Semantic().Ring)
			focusOffset := op.Offset(image.Pt(size.X-right, 0)).Push(gtx.Ops)
			dateDrawFocus(
				gtx,
				image.Pt(right, size.Y),
				focusColor,
				0,
				enabled && core.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger)),
				true,
				metrics.FocusRingWidth,
				metrics.IconSize,
				metrics.SegmentRadius,
			)
			focusOffset.Pop()
		}
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()

	// Draw the editable content after the trigger so segment click targets stay
	// on top of the field-wide disclosure target.
	stack = op.Offset(image.Pt(left, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
	contentClip := clip.Rect{Max: image.Pt(max(size.X-horizontalPadding, 0), contentDims.Size.Y)}.Push(gtx.Ops)
	if !showHint {
		call.Add(gtx.Ops)
	}
	if showHint {
		hintOffset := op.Offset(image.Pt(0, max((contentDims.Size.Y-hintDims.Size.Y)/2, 0))).Push(gtx.Ops)
		hintCall.Add(gtx.Ops)
		hintOffset.Pop()
	}
	contentClip.Pop()
	stack.Pop()
	return layout.Dimensions{Size: size}
}

func (d Widget) calendarLabel(ctx *core.Context) string {
	return ctx.Text(i18n.DatePickerChooseDate)
}

func (d Widget) layoutPopover(ctx *core.Context, gtx layout.Context, state *datePickerState, inputAnchor image.Rectangle, progress float32, now time.Time, naturallyDisabled bool) {
	metrics := state.partDefaults.currentDefaults(core.ActiveTheme(ctx)).metrics
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       core.FullKey(ctx, d.key),
		Layer:     core.OverlayLayerPopup,
		Anchor:    inputAnchor,
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			contentInteractive := interactive && state.open && gtx.Enabled()
			if contentInteractive {
				state.updateKeys(gtx, &state.trigger)
			}
			panelGtx := gtx
			// Keep an open panel enabled even when a nested overlay owns
			// policy input; disabling it would unregister that overlay.
			if !state.open || !gtx.Enabled() {
				panelGtx = panelGtx.Disabled()
			}
			return d.layoutCalendarOverlay(ctx, gtx, panelGtx, state, metrics, anchor, progress, now, contentInteractive)
		},
	})
}

func (d Widget) layoutCalendarOverlay(ctx *core.Context, gtx, panelGtx layout.Context, pickerState *datePickerState, metrics datePickerGeometry, anchor image.Rectangle, progress float32, now time.Time, interactive bool) layout.Dimensions {
	pickerState.popup.Update(ctx, gtx, base.PopupPolicy{
		Open: pickerState.open, Interactive: interactive, SurfaceFocus: &pickerState.trigger,
	})
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(metrics.PopoverGap)
	panelResolved := d.resolvePart(ctx, gtx, PartCalendarPanel, style.State{
		Disabled: !gtx.Enabled(),
		Open:     pickerState.open,
	}, &pickerState.calendarStyle)
	panelWidth := min(gtx.Dp(metrics.CalendarWidth)+datePickerHorizontalPadding(gtx, panelResolved), max(viewport.X, 0))
	panelMaxY := min(gtx.Dp(metrics.PopoverMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelMaxY <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelMaxY),
	}

	macro := op.Record(gtx.Ops)
	dims, tracked := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		resolved := panelResolved
		// A disabled overlay has a different style state. Preserve the old
		// disabled resolution only for that uncommon path; the interactive path
		// reuses the panel declaration already needed for width measurement.
		if panelGtx.Enabled() != gtx.Enabled() {
			resolved = d.resolvePart(ctx, panelGtx, PartCalendarPanel, style.State{
				Disabled: !panelGtx.Enabled(),
				Open:     pickerState.open,
			}, &pickerState.calendarStyle)
		}
		return d.layoutCalendarPanel(ctx, panelGtx, pickerState, metrics, now, resolved)
	})
	call := macro.Stop()
	placement := core.Placement{Side: core.SideBottom, Align: core.AlignStart}
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            dims.Size,
		Bounds:           viewport,
		Offset:           gap,
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := core.PanelTransformOriginAt(anchor, result.Position, dims.Size, result.Placement)
	motion := pickerState.partDefaults.motion(core.ActiveTheme(ctx))
	popoverScale := motion.PopoverScale
	if popoverScale <= 0 || popoverScale > 1 {
		popoverScale = 1
	}
	scale := popoverScale + (1-popoverScale)*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	tracked.PlaceTransform(f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform))
	tracked.SetOpacity(progress)

	stack := op.Offset(result.Position).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	transform := op.Affine(scaleTransform).Push(gtx.Ops)
	pickerState.popup.LayoutSurface(gtx, image.Rectangle{Max: dims.Size})
	call.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	stack.Pop()
	return dims
}

func (d Widget) layoutCalendarPanel(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time, resolved style.ResolvedStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutCalendar(ctx, gtx, state, metrics, now)
	}))
}

func datePickerHorizontalPadding(gtx layout.Context, resolved style.ResolvedStyle) int {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		panic("flowui/datepicker: missing horizontal padding")
	}
	return gtx.Dp(resolved.Box.Padding.Left) + gtx.Dp(resolved.Box.Padding.Right)
}

func datePickerPaintRadius(gtx layout.Context, resolved style.ResolvedStyle) int {
	if resolved.Paint == nil || resolved.Paint.Radius == nil {
		panic("flowui/datepicker: missing paint radius")
	}
	return max(gtx.Dp(*resolved.Paint.Radius), 0)
}

func (d Widget) layoutCalendar(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	gtx.Constraints.Min.X = min(gtx.Dp(metrics.CalendarWidth), gtx.Constraints.Max.X)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutCalendarHeader(ctx, gtx, state, metrics)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return divider.New().Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Spacer(0, gtx.Dp(metrics.CalendarTopGap)).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			switch state.viewMode {
			case datePickerViewMonths:
				return d.layoutMonths(ctx, gtx, state, metrics, now)
			case datePickerViewYears:
				return d.layoutYears(ctx, gtx, state, metrics, now)
			default:
				return d.layoutDayView(ctx, gtx, state, metrics, now)
			}
		}),
	)
}

func (d Widget) layoutCalendarHeader(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, metrics datePickerGeometry) layout.Dimensions {
	return layout.Inset{
		Left:   2,
		Right:  2,
		Bottom: metrics.HeaderBottom,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				base.RowInteractionWithFocus(ctx, gtx, &pickerState.header, &pickerState.trigger, false, func() {
					pickerState.toggleYearPicker(pickerState.viewMonth.Year())
				})
				return pickerState.header.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return d.layoutHeaderTrigger(ctx, gtx, pickerState, metrics)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, metrics, -1)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, metrics, 1)
			}),
		)
	})
}

func (d Widget) layoutHeaderTrigger(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry) layout.Dimensions {
	label := d.headerLabel(state)
	focused := gtx.Focused(&state.header)
	headerState := style.State{
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &state.header, focused),
		Selected:     state.viewMode == datePickerViewYears,
	}
	headerKey := datePickerStyleStateKey(headerState)
	resolved, cached := state.resolvedParts.headerValue(headerKey)
	if !cached {
		resolved = d.resolvePart(ctx, gtx, PartCalendarHeader, headerState, &state.calendarStyle)
		state.resolvedParts.storeHeader(headerKey, resolved)
	}
	gap := gtx.Dp(metrics.HeaderGap)
	iconSize := gtx.Dp(metrics.HeaderIconSize)
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(iconSize, iconSize)
			gtx.Constraints = layout.Exact(size)
			dateDrawYearIndicator(gtx, size, state.viewMode == datePickerViewYears, dateResolveTextColor(resolved))
			return layout.Dimensions{Size: size}
		}),
	)
}

func (d Widget) headerLabel(state *datePickerState) string {
	year, _, _ := state.viewMonth.Date()
	switch state.viewMode {
	case datePickerViewMonths:
		return d.locale.YearLabel(year)
	case datePickerViewYears:
		return d.locale.MonthLabel(state.viewMonth)
	default:
		return d.locale.MonthLabel(state.viewMonth)
	}
}

func (d Widget) layoutNavButton(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, metrics datePickerGeometry, delta int) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	defaults := pickerState.partDefaults.currentDefaults(active)
	size := image.Pt(gtx.Dp(metrics.NavButtonSize), gtx.Dp(metrics.NavButtonSize))
	if pickerState.viewMode == datePickerViewYears {
		return layout.Dimensions{Size: size}
	}

	clickable := &pickerState.prev
	if delta > 0 {
		clickable = &pickerState.next
	}
	disabled := !d.canMove(pickerState, delta)
	if disabled {
		gtx = gtx.Disabled()
	}
	interaction := base.RowInteractionWithFocus(ctx, gtx, clickable, &pickerState.trigger, disabled, func() { pickerState.move(delta) })
	focusVisible := interaction.FocusVisible
	gtx.Constraints = layout.Exact(size)
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := gtx.Focused(clickable)
		state := style.State{
			Hovered:      clickable.Hovered() && !disabled,
			Pressed:      clickable.Pressed() && !disabled,
			Focused:      focused,
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
		}
		navKey := datePickerStyleStateKey(state)
		resolved, cached := pickerState.resolvedParts.navValue(navKey)
		if !cached {
			resolved = d.resolvePart(ctx, gtx, PartCalendarNav, state, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeNav(navKey, resolved)
		}
		foreground := dateResolveTextColor(resolved)
		background := dateResolvePaintColor(resolved)
		radius := datePickerPaintRadius(gtx, resolved)
		focusKey := datePickerStyleStateKey(state)
		controlFocus, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			controlFocus = d.resolvePart(ctx, gtx, PartCalendarFocus, state, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, controlFocus)
		}
		focusColor := dateResolveOutlineColor(controlFocus, active.Semantic().Ring)
		motion := defaults.motion
		background = pickerState.navBackground(gtx, delta, background, motion)
		scale := datePickerPressScale(gtx, clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		dateDrawNavButton(gtx, size, delta, background, foreground, radius, metrics.NavChevronStroke)
		stack.Pop()
		dateDrawFocus(
			gtx,
			size,
			focusColor,
			radius,
			focusVisible && !disabled,
			false,
			metrics.FocusRingWidth,
			metrics.IconSize,
			metrics.SegmentRadius,
		)
		return layout.Dimensions{Size: size}
	})
}

func (d Widget) layoutDayView(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutWeekdays(ctx, gtx, state, metrics)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Spacer(0, gtx.Dp(metrics.WeekdayGap)).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutDays(ctx, gtx, state, metrics, now)
		}),
	)
}

func (d Widget) layoutWeekdays(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry) layout.Dimensions {
	cell := gtx.Dp(metrics.CellSize)
	weekdayState := style.State{Disabled: !gtx.Enabled()}
	weekdayKey := uint8(0)
	if weekdayState.Disabled {
		weekdayKey = 1
	}
	resolved, ok := state.resolvedParts.weekdayValue(weekdayKey)
	if !ok {
		resolved = d.resolvePart(ctx, gtx, PartCalendarWeekday, weekdayState, &state.calendarStyle)
		state.resolvedParts.storeWeekday(weekdayKey, resolved)
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, d.datePickerWeekdayChildren(ctx, cell, orderedDatePickerWeekdays(d.locale), resolved)...)
}

func (d Widget) datePickerWeekdayChildren(ctx *core.Context, cell int, weekdays [7]string, resolved style.ResolvedStyle) []layout.FlexChild {
	var children [len(weekdays)]layout.FlexChild
	for index, weekday := range weekdays {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				weekdayWidget := text.WithResolvedText(text.New(weekday), resolved.Text)
				return weekdayWidget.Layout(ctx, gtx)
			})
		})
	}
	return children[:]
}

func (d Widget) layoutDays(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	days := state.monthDays(state.viewMonth, d.locale.WeekStart)
	cell := gtx.Dp(metrics.CellSize)
	state.calendarLayout.prepare(ctx, d, state, days, cell, metrics, now)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, state.calendarLayout.rows[:]...)
}

func (d Widget) layoutMonths(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	year, _, _ := state.viewMonth.Date()
	cellWidth := gtx.Dp(metrics.CalendarWidth) / 3
	cellHeight := gtx.Dp(metrics.MonthCellHeight)
	children := make([]layout.FlexChild, 0, 4)
	for row := range 4 {
		monthStart := row*3 + 1
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart), metrics, now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+1), metrics, now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+2), metrics, now)
				}),
			)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (d Widget) layoutMonth(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, year int, month time.Month, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	date := time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
	key := fmt.Sprintf("%04d-%02d", year, int(month))
	monthState := pickerState.month(key)
	disabled := d.isMonthDisabled(date)
	if disabled {
		gtx = gtx.Disabled()
	}
	base.RowInteractionWithFocus(ctx, gtx, &monthState.clickable, &pickerState.trigger, disabled, func() {
		pickerState.viewMonth = date
		pickerState.viewMode = datePickerViewDays
	})

	valueYear, valueMonth, _ := d.value.Date()
	selected := !d.value.IsZero() && valueYear == year && valueMonth == month
	today := now.Year() == year && now.Month() == month
	active := pickerState.viewMonth.Year() == year && pickerState.viewMonth.Month() == month
	return d.layoutPickerCell(ctx, gtx, pickerState, monthState, d.locale.Months[int(month)-1], metrics, selected, active, today, false, disabled)
}

func (d Widget) layoutYears(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	minYear, maxYear := d.yearPickerRange(state, now)
	count := maxYear - minYear + 1
	if count <= 0 {
		return layout.Dimensions{}
	}

	rows := (count + 2) / 3
	cellHeight := gtx.Dp(metrics.YearCellHeight)
	if cellHeight == 0 {
		cellHeight = gtx.Dp(metrics.MonthCellHeight)
	}
	state.yearList.Axis = layout.Vertical
	state.yearList.Gap = gtx.Dp(metrics.YearGridGap)
	if state.yearScrollReady {
		target := min(max(state.yearScrollYear, minYear), maxYear)
		state.yearList.ScrollTo((target - minYear) / 3)
		state.yearScrollReady = false
	}
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.yearList, &state.yearBar, &state.yearVisualOutset, rows, !gtx.Enabled(), false, func(gtx layout.Context, row int) layout.Dimensions {
		cellWidth := gtx.Constraints.Max.X / 3
		yearStart := minYear + row*3
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart, maxYear, metrics, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+1, maxYear, metrics, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+2, maxYear, metrics, now)
			}),
		)
	})
}

func (d Widget) layoutYearCell(ctx *core.Context, gtx layout.Context, state *datePickerState, year, maxYear int, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	if year > maxYear {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return d.layoutYear(ctx, gtx, state, year, metrics, now)
}

func (d Widget) yearPickerRange(state *datePickerState, now time.Time) (int, int) {
	hasMin := !d.minDate.IsZero()
	hasMax := !d.maxDate.IsZero()
	minYear := 1900
	maxYear := 2099
	if hasMin {
		minYear = d.minDate.Year()
	}
	if hasMax {
		maxYear = d.maxDate.Year()
	}
	for _, year := range []int{state.viewMonth.Year(), now.Year()} {
		if !hasMin && year < minYear {
			minYear = year
		}
		if !hasMax && year > maxYear {
			maxYear = year
		}
	}
	if !d.value.IsZero() {
		year := d.value.Year()
		if !hasMin && year < minYear {
			minYear = year
		}
		if !hasMax && year > maxYear {
			maxYear = year
		}
	}
	if minYear > maxYear {
		minYear, maxYear = maxYear, minYear
	}
	return minYear, maxYear
}

func (d Widget) layoutYear(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, year int, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	yearState := pickerState.year(fmt.Sprintf("%04d", year))
	disabled := d.isYearDisabled(year)
	if disabled {
		gtx = gtx.Disabled()
	}
	base.RowInteractionWithFocus(ctx, gtx, &yearState.clickable, &pickerState.trigger, disabled, func() {
		_, month, _ := pickerState.viewMonth.Date()
		pickerState.viewMonth = time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
		pickerState.viewMode = datePickerViewDays
	})

	selected := !d.value.IsZero() && d.value.Year() == year
	today := now.Year() == year
	active := pickerState.viewMonth.Year() == year
	return d.layoutPickerCell(ctx, gtx, pickerState, yearState, d.locale.YearLabel(year), metrics, selected, active, today, false, disabled)
}

func (d Widget) layoutPickerCell(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, state *datePickerCellState, label string, metrics datePickerGeometry, selected, active, today, outside, disabled bool) layout.Dimensions {
	themeActive := core.ActiveTheme(ctx)
	defaults := pickerState.partDefaults.currentDefaults(themeActive)
	focusVisible := core.FocusVisible(ctx, &state.clickable, gtx.Focused(&state.clickable))
	return state.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		cellState := style.State{
			Hovered:      state.clickable.Hovered() && !disabled,
			Pressed:      state.clickable.Pressed() && !disabled,
			Focused:      gtx.Focused(&state.clickable),
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
			Selected:     selected,
		}
		cellContext := dateCellState{
			Hovered:  cellState.Hovered,
			Pressed:  cellState.Pressed,
			Selected: selected,
			Today:    today,
			Outside:  outside,
			Disabled: disabled,
			Active:   active,
		}
		cellKey := datePickerResolvedCellKey(cellContext, cellState)
		resolved, cached := pickerState.resolvedParts.cellValue(cellKey)
		if !cached {
			cellDefaults := pickerState.partDefaults.cellDefaults(themeActive, pickerState.calendarStyle, cellContext)
			resolved = d.resolvePartWithDefaults(ctx, gtx, PartCalendarCell, cellState, cellDefaults)
			pickerState.resolvedParts.storeCell(cellKey, resolved)
		}
		focusKey := datePickerStyleStateKey(cellState)
		focusResolved, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			focusResolved = d.resolvePart(ctx, gtx, PartCalendarFocus, cellState, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, focusResolved)
		}
		focusColor := dateResolveOutlineColor(focusResolved, themeActive.Semantic().Ring)
		motion := defaults.motion
		background := state.background(gtx, dateResolvePaintColor(resolved), motion)
		radius := datePickerPaintRadius(gtx, resolved)
		scale := datePickerPressScale(gtx, state.clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		dateDrawCell(gtx, size, background, radius)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		})
		stack.Pop()
		dateDrawFocus(
			gtx,
			size,
			focusColor,
			radius,
			focusVisible && !disabled,
			false,
			metrics.FocusRingWidth,
			metrics.IconSize,
			metrics.SegmentRadius,
		)
		return layout.Dimensions{Size: size}
	})
}

func (d Widget) layoutDay(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, day datePickerDay, column int, previousOutside, nextOutside bool, metrics datePickerGeometry, now time.Time) layout.Dimensions {
	themeActive := core.ActiveTheme(ctx)
	defaults := pickerState.partDefaults.currentDefaults(themeActive)
	key := dateKey(day.date)
	dayState := pickerState.day(key)
	dayState.date = day.date
	disabled := d.isDateDisabled(day.date)
	if disabled {
		gtx = gtx.Disabled()
	}
	interaction := base.RowInteractionWithFocus(ctx, gtx, &dayState.clickable, &pickerState.trigger, disabled, func() {
		date := dateOnly(day.date)
		if d.onDateSelect != nil {
			d.onDateSelect(date)
		} else {
			pickerState.open = false
			pickerState.sync(date, date)
			pickerState.requestValue(d, date)
		}
	})

	focusVisible := interaction.FocusVisible

	selectionStart := d.rangeMode && !day.outside && sameDate(d.rangeStart, day.date)
	selectionEnd := d.rangeMode && !day.outside && sameDate(d.rangeEnd, day.date)
	rangeSelected := d.rangeMode && dateBetween(day.date, d.rangeStart, d.rangeEnd)
	selected := sameDate(d.value, day.date)
	if d.rangeMode {
		selected = selectionStart || selectionEnd
	}
	today := sameDate(now, day.date)
	return dayState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		cellState := style.State{
			Hovered:      dayState.clickable.Hovered() && !disabled,
			Pressed:      dayState.clickable.Pressed() && !disabled,
			Focused:      gtx.Focused(&dayState.clickable),
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
			Selected:     selected,
		}
		cellContext := dateCellState{
			Hovered:  cellState.Hovered,
			Pressed:  cellState.Pressed,
			Selected: selected,
			Today:    today,
			Outside:  day.outside,
			Disabled: disabled,
			Range:    rangeSelected && !selected,
		}
		cellKey := datePickerResolvedCellKey(cellContext, cellState)
		resolved, cached := pickerState.resolvedParts.cellValue(cellKey)
		if !cached {
			cellDefaults := pickerState.partDefaults.cellDefaults(themeActive, pickerState.calendarStyle, cellContext)
			resolved = d.resolvePartWithDefaults(ctx, gtx, PartCalendarCell, cellState, cellDefaults)
			pickerState.resolvedParts.storeCell(cellKey, resolved)
		}
		foreground := dateResolveTextColor(resolved)
		focusKey := datePickerStyleStateKey(cellState)
		focusResolved, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			focusResolved = d.resolvePart(ctx, gtx, PartCalendarFocus, cellState, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, focusResolved)
		}
		focusColor := dateResolveOutlineColor(focusResolved, themeActive.Semantic().Ring)
		motion := defaults.motion
		background := dayState.background(gtx, dateResolvePaintColor(resolved), motion)
		radius := datePickerPaintRadius(gtx, resolved)
		if rangeSelected {
			startRadius, endRadius := 0, 0
			if column == 0 {
				startRadius = gtx.Dp(metrics.RangeRadius)
			}
			if column == 6 {
				endRadius = gtx.Dp(metrics.RangeRadius)
			}
			if !day.outside && previousOutside {
				startRadius = gtx.Dp(metrics.RangeRadius)
			}
			if !day.outside && nextOutside {
				endRadius = gtx.Dp(metrics.RangeRadius)
			}
			if selectionStart {
				startRadius = size.Y / 2
			}
			if selectionEnd {
				endRadius = size.Y / 2
			}
			rangeKey := focusKey << 1
			if day.outside {
				rangeKey |= 1
			}
			rangeResolved, cached := pickerState.resolvedParts.rangeValue(rangeKey)
			if !cached {
				rangeDefaults := pickerState.partDefaults.rangeDefaults(themeActive, pickerState.calendarStyle, day.outside)
				rangeResolved = d.resolvePartWithDefaults(ctx, gtx, PartCalendarRange, cellState, rangeDefaults)
				pickerState.resolvedParts.storeRange(rangeKey, rangeResolved)
			}
			dateDrawRange(gtx, size, dateResolvePaintColor(rangeResolved), startRadius, endRadius)
		}
		scale := datePickerPressScale(gtx, dayState.clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		dateDrawCell(gtx, size, background, radius)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(day.label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		})
		if disabled && !day.outside {
			dateDrawStrike(gtx, size, foreground, metrics.CellStrikeHalfSize, metrics.CellStrikeWidth)
		}
		stack.Pop()
		dateDrawFocus(
			gtx,
			size,
			focusColor,
			radius,
			focusVisible && !disabled,
			false,
			metrics.FocusRingWidth,
			metrics.IconSize,
			metrics.SegmentRadius,
		)
		return layout.Dimensions{Size: size}
	})
}
