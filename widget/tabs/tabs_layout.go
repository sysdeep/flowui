package tabs

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	iconui "github.com/qianniancn/flowui/widget/icon"
	textui "github.com/qianniancn/flowui/widget/text"
)

func (t Widget) layout(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, defaults tabsDefaults) layout.Dimensions {
	if t.panelsForceRender() {
		t.layoutForcedPanels(ctx, gtx, state, selectedKey, defaults)
	}
	var listDims layout.Dimensions
	list := func(gtx layout.Context) layout.Dimensions {
		listDims = t.layoutList(ctx, gtx, state, selectedKey, disabled, defaults)
		return listDims
	}
	item, hasPanel := t.selectedItem(selectedKey)
	hasPanel = hasPanel && item.Panel != nil
	if !hasPanel {
		return list(gtx)
	}
	var panelPlacement core.OverlayPlacement
	panel := func(gtx layout.Context) layout.Dimensions {
		opacity := state.panelOpacityFor(gtx, selectedKey, t.panelTransition, t.panelAnimationDuration(ctx, defaults), defaults.motion)
		fade := paint.PushOpacity(gtx.Ops, opacity)
		defer fade.Pop()
		dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return t.layoutPanel(ctx, gtx, item, defaults)
		})
		panelPlacement = placement
		return dims
	}
	gap := max(gtx.Dp(defaults.geometry.rootGap+defaults.geometry.panelGap), 0)
	alignment := layout.Start
	if t.centered && t.orientation == Horizontal {
		alignment = layout.Middle
	}
	if t.orientation == Vertical {
		if t.placement == End {
			dims := layout.Flex{Axis: layout.Horizontal, Alignment: alignment, Gap: gap}.Layout(gtx,
				layout.Flexed(1, panel),
				layout.Rigid(list),
			)
			panelPlacement.PlaceOffset(image.Pt(0, 0))
			return dims
		}
		dims := layout.Flex{Axis: layout.Horizontal, Alignment: alignment, Gap: gap}.Layout(gtx,
			layout.Rigid(list),
			layout.Flexed(1, panel),
		)
		panelPlacement.PlaceOffset(image.Pt(listDims.Size.X+gap, 0))
		return dims
	}
	if t.placement == Bottom {
		dims := layout.Flex{Axis: layout.Vertical, Alignment: alignment, Gap: gap}.Layout(gtx,
			layout.Rigid(panel),
			layout.Rigid(list),
		)
		panelPlacement.PlaceOffset(image.Pt(0, 0))
		return dims
	}
	dims := layout.Flex{Axis: layout.Vertical, Alignment: alignment, Gap: gap}.Layout(gtx,
		layout.Rigid(list),
		layout.Rigid(panel),
	)
	panelPlacement.PlaceOffset(image.Pt(0, listDims.Size.Y+gap))
	return dims
}

// layoutForcedPanels performs a one-time hidden layout for inactive panels.
// The frame helper gives each panel a disabled input source and a private ops
// stream, so this pass can initialize component state without adding visible
// paint, semantics, focus targets, or overlays to the current core.
func (t Widget) layoutForcedPanels(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, defaults tabsDefaults) {
	for _, item := range t.items {
		if item.Key == selectedKey || item.Panel == nil {
			continue
		}
		if _, rendered := state.renderedPanels[item.Key]; rendered {
			continue
		}
		scope := t.panelStateScope(ctx, item.Key)
		core.LayoutHidden(ctx, gtx, scope, core.WidgetFunc(func(ctx *core.Context, hiddenGtx layout.Context) layout.Dimensions {
			return t.layoutPanel(ctx, hiddenGtx, item, defaults)
		}))
		state.renderedPanels[item.Key] = struct{}{}
	}
}

func (t Widget) layoutList(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, defaults tabsDefaults) layout.Dimensions {
	tabStrip := func(gtx layout.Context) layout.Dimensions {
		return t.layoutTabStrip(ctx, gtx, state, selectedKey, disabled, defaults)
	}
	if t.centered && t.orientation == Horizontal {
		// Keep the strip's intrinsic width while centering it in the available
		// bar area. Leading/trailing slots are handled by the surrounding flex,
		// so this centers only within their remaining space.
		centeredStrip := tabStrip
		tabStrip = func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, centeredStrip)
		}
	}
	if t.leading == nil && t.trailing == nil && t.onAdd == nil {
		return tabStrip(gtx)
	}

	children := make([]layout.FlexChild, 0, 4)
	if t.leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSlot(ctx, gtx, "leading", t.leading)
		}))
	}
	stripChild := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return tabStrip(gtx)
	})
	if t.usesIntrinsicTabWidths() {
		stripChild = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return tabStrip(gtx)
		})
	}
	children = append(children, stripChild)
	if t.trailing != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSlot(ctx, gtx, "trailing", t.trailing)
		}))
	}
	if t.onAdd != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutAddButton(ctx, gtx, state, disabled, defaults)
		}))
	}
	axis := layout.Horizontal
	if t.orientation == Vertical {
		axis = layout.Vertical
	}
	return layout.Flex{Axis: axis, Alignment: layout.Middle, Gap: gtx.Dp(defaults.geometry.extraContentGap)}.Layout(gtx, children...)
}

func (t Widget) layoutSlot(ctx *core.Context, gtx layout.Context, role string, content core.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	return tabsSlotWidget{hostKey: t.key, role: role, content: content}.Layout(ctx, gtx)
}

func (t Widget) slotWidget(role string, content core.Widget) core.Widget {
	if content == nil {
		return nil
	}
	return tabsSlotWidget{hostKey: t.key, role: role, content: content}
}

type tabsSlotWidget struct {
	hostKey string
	role    string
	content core.Widget
}

func (w tabsSlotWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	restoreHost := core.PushKey(ctx, w.hostKey)
	defer restoreHost()
	restoreRole := core.PushKey(ctx, w.role)
	defer restoreRole()
	return w.content.Layout(ctx, gtx)
}

func (w tabsSlotWidget) Measure(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	restoreHost := core.PushKey(ctx, w.hostKey)
	defer restoreHost()
	restoreRole := core.PushKey(ctx, w.role)
	defer restoreRole()
	return core.MeasureWidget(ctx, gtx, w.content)
}

func (t Widget) layoutTabStripDefault(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, defaults tabsDefaults) layout.Dimensions {
	geometry := defaults.geometry
	variant := t.variant
	padding := gtx.Dp(geometry.listPadding)
	if variant == Secondary {
		padding = 0
	}
	sizeStyle := tabsSizeStyleFor(geometry, t.size)
	tabHeight := max(gtx.Dp(sizeStyle.height), 1)
	widths := t.tabWidths(ctx, gtx, padding, sizeStyle, defaults)
	tabGap := gtx.Dp(geometry.tabGap)
	size := t.listSize(gtx, widths, tabHeight, padding, tabGap)
	tabsDrawList(gtx, geometry, size, t.orientation, variant, defaults.palette.listBackground, defaults.palette.listBorder)

	innerSize := image.Pt(max(size.X-padding*2, 0), max(size.Y-padding*2, 0))
	listGtx := gtx
	listGtx.Constraints = layout.Exact(innerSize)
	if disabled {
		listGtx = listGtx.Disabled()
	}
	if state.noteListLayout(innerSize, t.orientation, widths, len(t.items)) {
		state.selectionPending = true
	}
	state.list.Axis = layout.Horizontal
	state.list.Gap = tabGap
	if t.orientation == Vertical {
		state.list.Axis = layout.Vertical
	}
	stack := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	clipStack := clip.Rect(image.Rectangle{Max: innerSize}).Push(gtx.Ops)
	contentMacro := op.Record(gtx.Ops)
	listLabel := ctx.Text(i18n.TabsList)
	semantic.DescriptionOp(listLabel).Add(gtx.Ops)
	listSemanticKey := core.DerivedKey(ctx, core.FullKey(ctx, t.key), "tab-list")
	core.RegisterSemantic(ctx, core.SemanticNode{
		Key:   listSemanticKey,
		Role:  core.SemanticTabList,
		Label: listLabel,
	})
	state.list.Layout(listGtx, len(t.items), func(gtx layout.Context, index int) layout.Dimensions {
		itemSize := image.Pt(widths[index], tabHeight)
		gtx.Constraints = layout.Exact(itemSize)
		return t.layoutTab(ctx, gtx, state, sizeStyle, index, selectedKey, disabled, defaults)
	})
	content := contentMacro.Stop()
	position := state.list.Position
	if state.ensureSelectionVisible(t.items, selectedKey) {
		ctx.Invalidate()
	}
	t.drawSelectionIndicator(ctx, gtx, state, position, widths, tabHeight, tabGap, selectedKey, disabled, defaults)
	content.Add(gtx.Ops)
	clipStack.Pop()
	stack.Pop()

	previous := state.canScrollPrevious()
	next := state.canScrollNext(len(t.items))
	shadowBackground := defaults.palette.listBackground
	if shadowBackground.A == 0 {
		shadowBackground = defaults.palette.scrollBackground
	}
	tabsDrawScroll(
		gtx, geometry, size, t.orientation, variant, previous, next,
		shadowBackground, defaults.palette.listBorder,
	)
	t.layoutScrollButtons(ctx, gtx, state, size, disabled, defaults)
	return layout.Dimensions{Size: size}
}

func (t Widget) drawSelectionIndicator(ctx *core.Context, gtx layout.Context, state *tabsState, position layout.Position, widths []int, tabHeight, tabGap int, selectedKey string, tabsDisabled bool, defaults tabsDefaults) {
	if t.indicatorVisibleSet && !t.indicatorVisible {
		state.indicator.set = false
		return
	}
	index := tabsIndexByKey(t.items, selectedKey)
	if index < 0 {
		state.indicator.set = false
		return
	}
	target := t.tabRect(position, widths, index, tabHeight, tabGap)
	target = t.indicatorTarget(ctx, gtx, target, defaults)
	rect := state.indicator.transitionWithDuration(gtx, selectedKey, t.orientation, target, t.indicatorAnimationDuration(ctx, defaults), defaults.motion)
	disabled := tabsDisabled || t.items[index].Disabled
	_, indicator, _, _ := defaults.palette.itemColors(disabled)
	tabsDrawIndicator(gtx, defaults.geometry, rect, t.orientation, t.variant, indicator)
}

func (t Widget) indicatorTarget(ctx *core.Context, gtx layout.Context, target image.Rectangle, defaults tabsDefaults) image.Rectangle {
	tokens := defaults.geometry
	width := t.indicatorWidth
	if width <= 0 {
		width = tokens.indicatorWidth
	}
	// Align against the measured tab node. Its padding is part of the tab's
	// interactive and visual slot, so start/center/end use the same bounds as
	// the tab item rather than a separately reconstructed content rectangle.
	mainLength := target.Dx()
	if t.orientation == Vertical {
		mainLength = target.Dy()
	}
	inset := 0
	minLength := 0
	variant := t.variant
	if variant == Secondary {
		inset = gtx.Dp(tokens.indicatorInset)
		minLength = gtx.Dp(tokens.indicatorMinWidth)
	}
	mainLength = min(max(mainLength-inset*2, minLength), mainLength)
	if width > 0 {
		mainLength = min(mainLength, gtx.Dp(width))
	}
	mainLength = max(min(mainLength, max(target.Dx(), target.Dy())), 0)
	align := t.indicatorAlign
	if !t.indicatorAlignSet && variant == Secondary {
		align = IndicatorCenter
	}
	if t.orientation == Vertical {
		offset := 0
		switch align {
		case IndicatorCenter:
			offset = (target.Dy() - mainLength) / 2
		case IndicatorEnd:
			offset = target.Dy() - mainLength
		}
		target.Min.Y += offset
		target.Max.Y = target.Min.Y + mainLength
		if variant == Secondary {
			lineWidth := min(max(gtx.Dp(tokens.indicatorLineWidth), 1), max(target.Dx(), 1))
			if t.placement == End {
				target.Min.X = target.Max.X - lineWidth
				target.Max.X = target.Min.X + lineWidth
			}
		}
		return target
	}
	offset := 0
	switch align {
	case IndicatorCenter:
		offset = (target.Dx() - mainLength) / 2
	case IndicatorEnd:
		offset = target.Dx() - mainLength
	}
	target.Min.X += offset
	target.Max.X = target.Min.X + mainLength
	if variant == Secondary && t.placement == Bottom {
		lineWidth := min(max(gtx.Dp(tokens.indicatorLineWidth), 1), max(target.Dy(), 1))
		target.Max.Y = target.Min.Y + lineWidth
	}
	return target
}

func (t Widget) tabRect(position layout.Position, widths []int, index, tabHeight, tabGap int) image.Rectangle {
	gap := tabGap
	itemMainSize := func(item int) int {
		return widths[item]
	}
	if t.orientation == Vertical {
		itemMainSize = func(int) int {
			return tabHeight
		}
	}
	start := -position.Offset
	if index >= position.First {
		for item := position.First; item < index; item++ {
			start += itemMainSize(item) + gap
		}
	} else {
		for item := position.First - 1; item >= index; item-- {
			start -= itemMainSize(item) + gap
		}
	}
	if t.orientation == Vertical {
		return image.Rect(0, start, widths[index], start+tabHeight)
	}
	return image.Rect(start, 0, start+widths[index], tabHeight)
}

func (t Widget) tabWidths(ctx *core.Context, gtx layout.Context, padding int, sizeStyle tabsSizeStyle, values ...tabsDefaults) []int {
	effective := tabsDefaults{}
	if len(values) > 0 {
		effective = values[0]
	} else {
		effective = t.defaultsWithCache(ctx)
	}
	widths := make([]int, len(t.items))
	if len(widths) == 0 {
		return widths
	}
	geometry := effective.geometry
	available := max(gtx.Constraints.Max.X-padding*2, 0)
	tabGap := gtx.Dp(geometry.tabGap)
	total := 0
	for index, item := range t.items {
		width := t.measureTabWidth(ctx, gtx, item, sizeStyle, effective)
		if t.orientation == Vertical {
			width = max(width, gtx.Dp(geometry.tabMinWidth))
		}
		widths[index] = width
		total += width
	}
	if t.orientation != Vertical && !t.usesIntrinsicTabWidths() {
		gapTotal := max(len(widths)-1, 0) * tabGap
		extra := available - total - gapTotal
		if extra <= 0 {
			return widths
		}
		for index := range widths {
			share := extra / (len(widths) - index)
			widths[index] += share
			extra -= share
		}
	}
	if t.orientation == Vertical {
		maxWidth := 0
		for _, width := range widths {
			maxWidth = max(maxWidth, width)
		}
		maxWidth = min(maxWidth, available)
		for index := range widths {
			widths[index] = maxWidth
		}
	}
	return widths
}

func (t Widget) measureTabWidth(ctx *core.Context, gtx layout.Context, item TabItem, sizeStyle tabsSizeStyle, values ...tabsDefaults) int {
	effective := tabsDefaults{}
	if len(values) > 0 {
		effective = values[0]
	} else {
		effective = t.defaultsWithCache(ctx)
	}
	measure := gtx
	measure.Constraints = layout.Constraints{Max: image.Pt(1e6, max(gtx.Constraints.Max.Y, 1))}
	leftPadding := gtx.Dp(sizeStyle.paddingX)
	rightPadding := gtx.Dp(t.tabRightPadding(ctx, sizeStyle, item, effective))
	width := 0
	if item.Content != nil {
		width = t.measureTabContent(ctx, measure, item.Key, item.Content).Size.X
	} else {
		label := textui.WithResolvedText(textui.New(item.Label), t.itemTextStyle(ctx, effective, effective.palette.foreground))
		dims := textui.Measure(ctx, measure, label)
		width = dims.Size.X
	}
	if len(item.Icon) > 0 {
		width += gtx.Dp(effective.geometry.iconSize) + gtx.Dp(effective.geometry.iconGap)
	}
	for _, leading := range []core.Widget{item.Leading, item.Trailing} {
		if leading == nil {
			continue
		}
		child := t.measureTabContent(ctx, measure, item.Key, leading)
		width += child.Size.X + gtx.Dp(effective.geometry.iconGap)
	}
	if item.Closable && t.onClose != nil {
		closeSize, closeGap := t.closeButtonMetrics(effective)
		width += gtx.Dp(closeSize) + gtx.Dp(closeGap)
	}
	return width + leftPadding + rightPadding
}

// tabRightPadding keeps a closable tab's close slot inside the normal right
// padding whenever there is enough room. This avoids adding a full padding
// allowance and then adding the hidden close slot on top of it.
func (t Widget) tabRightPadding(ctx *core.Context, sizeStyle tabsSizeStyle, item TabItem, defaults tabsDefaults) unit.Dp {
	right := sizeStyle.paddingX
	if item.Closable && t.onClose != nil {
		closeSize, closeGap := t.closeButtonMetrics(defaults)
		slot := closeSize + closeGap
		right = max(right-slot, 0)
	}
	return right
}

func (t Widget) measureTabContent(ctx *core.Context, gtx layout.Context, itemKey string, content core.Widget) layout.Dimensions {
	restoreHost := core.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, itemKey)
	defer restoreItem()
	if measurable, ok := content.(core.Measurable); ok {
		return measurable.Measure(ctx, gtx)
	}
	return core.MeasureWidget(ctx, gtx, content)
}

func (t Widget) layoutTabContent(ctx *core.Context, gtx layout.Context, itemKey string, content core.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	restoreHost := core.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, itemKey)
	defer restoreItem()
	return content.Layout(ctx, gtx)
}

func (t Widget) listSize(gtx layout.Context, widths []int, tabHeight, padding, tabGap int) image.Point {
	if t.orientation == Vertical {
		width := 0
		for _, itemWidth := range widths {
			width = max(width, itemWidth)
		}
		height := tabHeight*len(widths) + max(len(widths)-1, 0)*tabGap + padding*2
		return gtx.Constraints.Constrain(image.Pt(width+padding*2, height))
	}
	height := tabHeight + padding*2
	width := gtx.Constraints.Max.X
	if t.usesIntrinsicTabWidths() {
		width = padding * 2
		for _, itemWidth := range widths {
			width += itemWidth
		}
		width += max(len(widths)-1, 0) * tabGap
	}
	return gtx.Constraints.Constrain(image.Pt(width, height))
}

func (t Widget) layoutTab(ctx *core.Context, gtx layout.Context, componentState *tabsState, sizeStyle tabsSizeStyle, index int, selectedKey string, tabsDisabled bool, defaults tabsDefaults) layout.Dimensions {
	item := t.items[index]
	itemState := componentState.item(item.Key)
	disabled := tabsDisabled || item.Disabled
	selected := item.Key == selectedKey
	if componentState.editingKey != item.Key {
		itemState.editReady = false
		itemState.editFocused = false
	}
	editing := t.itemIsEditing(componentState, item)
	animGtx := gtx
	closeClicked := false
	if !disabled && !editing && item.Editable {
		for {
			e, ok := gtx.Event(key.Filter{Focus: &itemState.clickable, Name: key.NameF2})
			if !ok {
				break
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				componentState.editingKey = item.Key
				itemState.editReady = false
				itemState.editFocused = false
				editing = true
				core.RequestFocus(ctx, &itemState.editor)
			}
		}
	}
	if disabled {
		gtx = gtx.Disabled()
	} else if !editing {
		if item.Closable && t.onClose != nil && t.closeVisible(itemState) {
			base.UpdateRow(ctx, gtx, base.RowInput{
				Clickable: &itemState.close, PreserveFocus: true,
				OnClick: func(widget.Click) {
					closeClicked = true
					t.onClose(item.Key)
					if selected {
						if fallback, ok := tabsCloseFallback(t.items, index); ok {
							selectedKey = componentState.requestSelectedKey(t, t.items[fallback].Key)
							if t.hasSelectedKey {
								// Controlled bindings return the old value until the
								// caller feeds the request back. Keep the current frame
								// usable while preserving that contract.
								selectedKey = t.items[fallback].Key
							}
							componentState.selectionPending = true
						} else {
							selectedKey = componentState.requestSelectedKey(t, "")
							if t.hasSelectedKey {
								selectedKey = ""
							}
							componentState.selectionPending = false
						}
					}
				},
			})
		}
		base.RowInteraction(ctx, gtx, &itemState.clickable, false, func() {
			if !closeClicked && !selected {
				componentState.requestSelectedKey(t, item.Key)
			}
		})
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(t.itemAccessibleLabel(item)).Add(gtx.Ops)
		description := item.Description
		if description == "" {
			description = ctx.Text(i18n.TabsTab)
			if selected {
				description = ctx.Text(i18n.TabsSelectedTab)
			}
		}
		semantic.DescriptionOp(description).Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		core.RegisterSemantic(ctx, core.SemanticNode{
			Key:         core.DerivedKey(ctx, core.FullKey(ctx, t.key), "tab:"+item.Key),
			Role:        core.SemanticTab,
			Label:       t.itemAccessibleLabel(item),
			Description: description,
			Controls:    core.DerivedKey(ctx, core.FullKey(ctx, t.key), "panel:"+item.Key),
			Owner:       core.DerivedKey(ctx, core.FullKey(ctx, t.key), "tab-list"),
			Selected:    selected,
			Disabled:    disabled,
			PosInSet:    index + 1,
			SetSize:     len(t.items),
		})

		focusVisible := core.FocusVisible(ctx, &itemState.clickable, gtx.Focused(&itemState.clickable))
		motion := defaults.motion
		focus := itemState.interaction.Opacity(animGtx, focusVisible && !disabled, motion)
		progress := itemState.selectionProgressWithDuration(animGtx, selected, t.selectionDuration(ctx, defaults), motion)
		variant := t.variant
		selectedForeground, _, separatorColor, focusColor := defaults.palette.itemColors(disabled)
		if t.separators && variant != Secondary && index > 0 && !selected && t.items[index-1].Key != selectedKey {
			tabsDrawSeparator(gtx, defaults.geometry, size, t.orientation, separatorColor)
		}
		foreground := animation.LerpColor(defaults.palette.foreground, selectedForeground, progress)
		styleState := style.State{
			Hovered: itemState.clickable.Hovered(), Pressed: itemState.clickable.Pressed(),
			Focused: gtx.Focused(&itemState.clickable), FocusVisible: focusVisible,
			Disabled: disabled, Selected: selected,
		}
		resolved := t.resolveItemStyle(ctx, animGtx, item.Key, styleState, foreground, defaults)
		if editing {
			t.prepareTabEditor(ctx, gtx, componentState, item, itemState)
		}
		return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
			if resolved.Text == nil || resolved.Text.FontSize == nil || resolved.Text.FontWeight == nil {
				panic("flowui/tabs: missing editable tab typography")
			}
			textSize := *resolved.Text.FontSize
			weight := font.Weight(*resolved.Text.FontWeight)
			color := foreground
			if resolved.Text != nil {
				if value, ok := computed.Color(resolved.Text.Color); ok {
					color = value
				}
			}
			children := make([]layout.FlexChild, 0, 4)
			if len(item.Icon) > 0 {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconui.New(item.Icon).Size(float32(defaults.geometry.iconSize)).Layout(ctx, gtx)
				}))
			}
			if item.Leading != nil {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Leading)
				}))
			}
			labelContent := func(gtx layout.Context) layout.Dimensions {
				if editing {
					semantic.Editor.Add(gtx.Ops)
					semantic.LabelOp(t.itemAccessibleLabel(item)).Add(gtx.Ops)
					return material.Editor(core.ActiveMaterial(ctx), &itemState.editor, "").Layout(gtx)
				}
				if item.Content != nil {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Content)
				}
				label := material.Label(core.ActiveMaterial(ctx), textSize, item.Label)
				label.Font.Weight = weight
				label.Color = color
				return layout.Center.Layout(gtx, label.Layout)
			}
			if t.usesIntrinsicTabWidths() {
				children = append(children, layout.Rigid(labelContent))
			} else {
				children = append(children, layout.Flexed(1, labelContent))
			}
			if item.Trailing != nil {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Trailing)
				}))
			}
			contentLayout := func(contentGtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: contentGtx.Dp(defaults.geometry.iconGap)}.Layout(contentGtx, children...)
			}
			layoutContent := contentLayout
			if item.Closable && t.onClose != nil {
				closeSize, closeGap := t.closeButtonMetrics(defaults)
				closeSizePx := gtx.Dp(closeSize)
				closeGapPx := gtx.Dp(closeGap)
				closeChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !t.closeVisible(itemState) {
						size := max(gtx.Dp(closeSize), 1)
						return layout.Dimensions{Size: image.Pt(size, size)}
					}
					return t.layoutCloseButton(ctx, gtx, item, itemState, disabled, defaults)
				})
				layoutContent = func(contentGtx layout.Context) layout.Dimensions {
					content := contentGtx
					// The outer Flex already reserves closeGapPx before this
					// rigid child, so only the close button's own width remains
					// to be removed from the content constraint.
					content.Constraints.Max.X = max(content.Constraints.Max.X-closeSizePx, 0)
					content.Constraints.Min.X = min(content.Constraints.Min.X, content.Constraints.Max.X)
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: content.Dp(defaults.geometry.iconGap)}.Layout(content, children...)
				}
				layout.Inset{
					Left:  sizeStyle.paddingX,
					Right: t.tabRightPadding(ctx, sizeStyle, item, defaults),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: closeGapPx}.Layout(gtx,
						layout.Rigid(layoutContent),
						closeChild,
					)
				})
			} else {
				layout.Inset{
					Left:  sizeStyle.paddingX,
					Right: t.tabRightPadding(ctx, sizeStyle, item, defaults),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutContent(gtx)
				})
			}
			tabsDrawFocus(gtx, defaults.geometry, size, t.variant, focus, focusColor)
			return layout.Dimensions{Size: size}
		}))
	})
}

func (t Widget) closeVisible(itemState *tabsItemState) bool {
	if !t.closeOnHover {
		return true
	}
	return itemState.clickable.Hovered() || itemState.close.Hovered() || itemState.close.Pressed()
}

func (t Widget) itemIsEditing(state *tabsState, item TabItem) bool {
	return t.onEdit != nil && item.Editable && state.editingKey == item.Key
}

func (t Widget) prepareTabEditor(ctx *core.Context, gtx layout.Context, componentState *tabsState, item TabItem, itemState *tabsItemState) {
	editKey := core.DerivedKey(ctx, core.FullKey(ctx, t.key), "edit:"+item.Key)
	editKey = core.ClaimKey(ctx, core.KindInput, editKey)
	if !itemState.editReady {
		itemState.editor.SetText(item.Label)
		itemState.editReady = true
		core.RequestFocus(ctx, &itemState.editor)
	}
	events := (base.TextEditor{SingleLine: true}).Update(ctx, gtx, base.TextEditorRequest{
		FieldKey: editKey,
		Editor:   &itemState.editor,
		Value:    itemState.editor.Text(),
	})
	if !gtx.Enabled() {
		return
	}
	if events.Submitted {
		t.finishTabEdit(componentState, item, itemState, true)
		return
	}
	for {
		event, ok := gtx.Event(key.Filter{Focus: &itemState.editor, Name: key.NameEscape})
		if !ok {
			break
		}
		if pressed, ok := event.(key.Event); ok && pressed.State == key.Press {
			itemState.editor.SetText(item.Label)
			t.finishTabEdit(componentState, item, itemState, false)
			return
		}
	}
	focused := gtx.Focused(&itemState.editor)
	if focused {
		itemState.editFocused = true
	} else if itemState.editFocused {
		t.finishTabEdit(componentState, item, itemState, true)
	}
}

func (t Widget) finishTabEdit(componentState *tabsState, item TabItem, itemState *tabsItemState, commit bool) {
	value := strings.TrimSpace(itemState.editor.Text())
	componentState.editingKey = ""
	itemState.editReady = false
	itemState.editFocused = false
	if commit && value != "" && t.onEdit != nil {
		t.onEdit(item.Key, value)
	}
}

func (t Widget) layoutCloseButton(ctx *core.Context, gtx layout.Context, item TabItem, itemState *tabsItemState, disabled bool, defaults tabsDefaults) layout.Dimensions {
	closeSize, _ := t.closeButtonMetrics(defaults)
	size := max(gtx.Dp(closeSize), 1)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	buttonGtx := gtx
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	}
	return itemState.close.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(ctx.Textf(i18n.TabsClose, t.itemAccessibleLabel(item))).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		col := defaults.palette.foreground
		if itemState.close.Hovered() {
			col = defaults.palette.scrollForeground
		}
		if disabled {
			col = defaults.palette.disabledForeground
		}
		tabsDrawClose(gtx, image.Pt(size, size), col)
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}

func (t Widget) itemAccessibleLabel(item TabItem) string {
	if item.AccessibleLabel != "" {
		return item.AccessibleLabel
	}
	if item.Label != "" {
		return item.Label
	}
	return item.Key
}

func (t Widget) resolveItemStyle(ctx *core.Context, gtx layout.Context, itemKey string, state style.State, foreground color.NRGBA, values ...tabsDefaults) style.ResolvedStyle {
	effective := tabsDefaults{}
	if len(values) > 0 {
		effective = values[0]
	} else {
		effective = t.defaultsWithCache(ctx)
	}
	item := style.Join(effective.item, style.Style{}.Part(style.PartItem, style.New().TextColor(style.Solid(foreground))))
	return computed.ResolvePart(
		ctx, gtx, core.DerivedKey(ctx, core.FullKey(ctx, t.key), "item:"+itemKey),
		style.PartItem, state, item, style.Style{}, style.Style{}, t.style,
	)
}

// itemTextStyle resolves the item declaration for text measurement.
// Typography is a theme decision and must be identical to the declaration
// consumed by the painted item.
func (t Widget) itemTextStyle(ctx *core.Context, defaults tabsDefaults, foreground color.NRGBA) *style.TextStyle {
	item := style.Join(defaults.item, style.Style{}.Part(style.PartItem, style.New().TextColor(style.Solid(foreground))))
	resolved := computed.ResolvePartStatic(ctx, style.PartItem, style.State{}, item, style.Style{}, style.Style{}, t.style)
	if resolved.Text == nil || resolved.Text.FontSize == nil || resolved.Text.FontWeight == nil {
		panic("flowui/tabs: missing item typography")
	}
	return resolved.Text
}

func (t Widget) layoutPanel(ctx *core.Context, gtx layout.Context, item TabItem, defaults tabsDefaults) layout.Dimensions {
	semanticRoot := core.FullKey(ctx, t.key)
	scope := t.panelStateScope(ctx, item.Key)
	restoreHost := core.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, item.Key)
	defer restoreItem()
	if t.panelsKeepAlive() {
		restoreRetention := core.PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	itemLabel := t.itemAccessibleLabel(item)
	panelDescription := ctx.Textf(i18n.TabsPanel, itemLabel)
	semantic.LabelOp(itemLabel).Add(gtx.Ops)
	semantic.DescriptionOp(panelDescription).Add(gtx.Ops)
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	core.RegisterSemantic(ctx, core.SemanticNode{
		Key:         core.DerivedKey(ctx, semanticRoot, "panel:"+item.Key),
		Role:        core.SemanticTabPanel,
		Label:       itemLabel,
		Description: panelDescription,
		Owner:       core.DerivedKey(ctx, semanticRoot, "tab-list"),
		Hidden:      false,
		Selected:    true,
	})
	if t.orientation != Vertical {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	resolved := computed.ResolvePart(
		ctx, gtx, semanticRoot, style.PartPanel,
		style.State{Selected: true, Disabled: !gtx.Enabled()},
		defaults.panel, style.Style{}, style.Style{}, t.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, item.Panel)
}

func (t Widget) layoutScrollButtons(ctx *core.Context, gtx layout.Context, state *tabsState, size image.Point, disabled bool, defaults tabsDefaults) {
	previous := state.canScrollPrevious()
	next := state.canScrollNext(len(t.items))
	if previous {
		t.layoutScrollButton(ctx, gtx, state, &state.previous, size, -1, disabled, defaults)
	}
	if next {
		t.layoutScrollButton(ctx, gtx, state, &state.next, size, 1, disabled, defaults)
	}
}

func (t Widget) layoutScrollButton(ctx *core.Context, gtx layout.Context, state *tabsState, clickable *widget.Clickable, listSize image.Point, direction int, disabled bool, defaults tabsDefaults) {
	geometry := defaults.geometry
	buttonSize := min(gtx.Dp(geometry.scrollButtonSize), min(listSize.X, listSize.Y))
	inset := gtx.Dp(geometry.scrollButtonInset)
	position := image.Pt(inset, max((listSize.Y-buttonSize)/2, 0))
	if direction > 0 {
		position.X = max(listSize.X-buttonSize-inset, 0)
	}
	if t.orientation == Vertical {
		position = image.Pt(max((listSize.X-buttonSize)/2, 0), inset)
		if direction > 0 {
			position.Y = max(listSize.Y-buttonSize-inset, 0)
		}
	}
	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(image.Pt(buttonSize, buttonSize))
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	} else {
		for clickable.Clicked(buttonGtx) {
			state.scrollPage(direction, len(t.items))
			ctx.Invalidate()
		}
	}
	stack := op.Offset(position).Push(gtx.Ops)
	clickable.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(t.scrollButtonLabel(ctx, direction)).Add(gtx.Ops)
		col := defaults.palette.scrollForeground
		if clickable.Hovered() {
			col.A = byte(float32(col.A)*0.7 + 0.5)
		}
		if disabled {
			col = defaults.palette.disabledForeground
		}
		tabsDrawChevron(gtx, geometry, image.Pt(buttonSize, buttonSize), t.orientation, direction, col)
		return layout.Dimensions{Size: image.Pt(buttonSize, buttonSize)}
	})
	stack.Pop()
}

func (t Widget) scrollButtonLabel(ctx *core.Context, direction int) string {
	if t.orientation == Vertical {
		if direction < 0 {
			return ctx.Text(i18n.TabsScrollUp)
		}
		return ctx.Text(i18n.TabsScrollDown)
	}
	if direction < 0 {
		return ctx.Text(i18n.TabsScrollLeft)
	}
	return ctx.Text(i18n.TabsScrollRight)
}

func (t Widget) layoutAddButton(ctx *core.Context, gtx layout.Context, state *tabsState, disabled bool, defaults tabsDefaults) layout.Dimensions {
	geometry := defaults.geometry
	sizeStyle := tabsSizeStyleFor(geometry, t.size)
	size := max(gtx.Dp(sizeStyle.height)+gtx.Dp(geometry.listPadding)*2, 1)
	if t.orientation == Vertical {
		size = max(gtx.Dp(geometry.tabHeight), 1)
	}
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	buttonGtx := gtx
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	}
	if !disabled {
		for state.add.Clicked(buttonGtx) {
			if t.onAdd != nil {
				t.onAdd()
			}
			core.PreserveFocus(ctx)
		}
	}
	return state.add.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(ctx.Text(i18n.TabsAddTab)).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		col := defaults.palette.scrollForeground
		if state.add.Hovered() {
			col.A = byte(float32(col.A)*0.7 + 0.5)
		}
		if disabled {
			col = defaults.palette.disabledForeground
		}
		iconSize := min(gtx.Dp(geometry.iconSize), size)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		iconOffset := op.Offset(image.Pt((size-iconSize)/2, (size-iconSize)/2)).Push(gtx.Ops)
		tabsDrawAdd(iconGtx, image.Pt(iconSize, iconSize), col)
		iconOffset.Pop()
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}
