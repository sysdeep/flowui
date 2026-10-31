package tree

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/toggle"
)

func (t Widget) layout(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, visible []flatItem, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	maxHeight := metrics.MaxHeight
	if t.maxHeight > 0 {
		maxHeight = unit.Dp(t.maxHeight)
	}
	if maxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := core.PushColors(ctx, defaults.colors.RootForeground, defaults.colors.RootBackground)
		defer restore()
		contentDims = layout.UniformInset(metrics.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if len(visible) == 0 {
				return t.layoutEmpty(ctx, gtx, defaults)
			}
			treeStateValue.list.Axis = layout.Vertical
			treeStateValue.list.Gap = gtx.Dp(metrics.Gap)
			treeStateValue.list.Alignment = layout.Start
			treeStateValue.list.ScrollAnyAxis = false
			return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &treeStateValue.list, &treeStateValue.bar, &treeStateValue.visualOutset, len(visible), t.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
				return t.layoutItem(ctx, gtx, treeStateValue, visible[index], index, defaults)
			})
		})
	}()
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	radius := treeRootRadius(gtx, metrics, t.variant, size)
	treeDrawRoot(core.ActiveTheme(ctx), gtx, size, metrics.SurfaceRadius, defaults.colors.RootBackground, t.variant == VariantSurface)
	if t.variant == VariantSurface {
		root := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
		content.Add(gtx.Ops)
		root.Pop()
	} else {
		root := clip.Rect{Max: size}.Push(gtx.Ops)
		content.Add(gtx.Ops)
		root.Pop()
	}
	return layout.Dimensions{Size: size, Baseline: contentDims.Baseline}
}

func (t Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	height := min(gtx.Dp(metrics.RowHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value := t.emptyText
		if !t.hasEmptyText {
			value = ctx.Text(i18n.CollectionNoItems)
		}
		return treeDrawText(core.ActiveTheme(ctx), gtx, value, metrics.ItemTextSize, defaults.colors.Description, 0)
	})
}

func (t Widget) layoutItem(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, entry flatItem, index int, defaults treeDefaults) layout.Dimensions {
	itemState := treeStateValue.item(entry.item.Key)
	itemState.index = index
	disabled := t.itemDisabled(entry.item)
	selected := t.itemSelected(entry.item.Key)
	expanded := treeContainsKey(t.expandedKeys, entry.item.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	}

	metrics := defaults.metrics
	height := min(max(treeItemHeight(gtx, metrics, entry.item), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	width := gtx.Constraints.Max.X
	size := image.Pt(width, height)
	rowGtx := gtx
	rowGtx.Constraints = layout.Exact(size)

	focusVisible := core.FocusVisible(ctx, &itemState.clickable, rowGtx.Focused(&itemState.clickable))
	focus := itemState.focus.Opacity(animGtx, focusVisible && !disabled, defaults.motion)
	hovered := (itemState.clickable.Hovered() || itemState.toggle.Hovered() || itemState.check.Hovered()) && !disabled
	rowPaint := defaults.rowPaint(selected, hovered, disabled)
	if treeContainsKey(treeStateValue.dragSources, entry.item.Key) {
		rowPaint.Opacity *= 0.6
	}
	expansionTarget := float32(0)
	if expanded {
		expansionTarget = 1
	}
	expansion := itemState.expansion.Value(animGtx, expansionTarget, treeExpandDuration, animation.EaseSmoothstep, defaults.motion)

	macro := op.Record(gtx.Ops)
	func() {
		restore := core.PushColors(ctx, rowPaint.Foreground, rowPaint.Background)
		defer restore()
		if t.guides {
			drawTreeGuides(rowGtx, entry, metrics, expanded, t.guideConnectors, t.guideStyle, defaults)
		}
		t.layoutItemContent(ctx, rowGtx, treeStateValue, itemState, entry, rowPaint, selected, expanded, expansion, disabled, defaults)
	}()
	content := macro.Stop()

	row := func(gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, rowPaint.Opacity)
		treeDrawRow(gtx, size, rowPaint.Background, rowPaint.Focus, focus, metrics.RowRadius, metrics.FocusRingWidth)
		content.Add(gtx.Ops)
		opacity.Pop()
		if treeStateValue.dropTarget.drawKey == entry.item.Key {
			treeDrawDropIndicator(gtx, size, treeStateValue.dropTarget.position, treeStateValue.dropTarget.depth, metrics, defaults.colors.DropIndicator)
		}
		return layout.Dimensions{Size: size}
	}
	interactiveRow := func(gtx layout.Context) layout.Dimensions {
		if t.onDrop == nil || disabled {
			return row(gtx)
		}
		dims := row(gtx)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		var preview layout.Widget
		if treeStateValue.dragSource == entry.item.Key {
			label := entry.item.Label
			if count := len(treeStateValue.dragSources); count > 1 {
				label = fmt.Sprintf("%s +%d", label, count-1)
			}
			offsetPx := gtx.Dp(metrics.DragPreviewOffset)
			offset := itemState.drag.Press().Round().Add(image.Pt(offsetPx, offsetPx))
			preview = t.dragPreview(ctx, label, offset, defaults)
		}
		itemState.drag.Layout(gtx, func(layout.Context) layout.Dimensions { return dims }, preview)
		pass.Pop()
		registerTreeDragAreas(gtx, itemState, dims.Size, treeItemAcceptsChildren(entry.item), !entry.item.DragDisabled, !entry.item.DropDisabled)
		return dims
	}
	if !t.hasContextMenu || disabled || treeStateValue.renameKey == entry.item.Key {
		return interactiveRow(rowGtx)
	}
	trigger := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return interactiveRow(gtx)
	})
	owner := core.FullKey(ctx, t.key)
	key := core.DerivedKey(ctx, owner, "context-menu:"+entry.item.Key)
	return menu.ContextMenu(key, trigger, t.contextMenu).
		FocusTarget(&itemState.clickable).
		OnOpenChange(func(open bool) {
			if open && t.onContextMenu != nil {
				t.onContextMenu(entry.item.Key)
			}
		}).
		Layout(ctx, rowGtx)
}

func (t Widget) dragPreview(ctx *core.Context, label string, offset image.Point, defaults treeDefaults) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		stack := op.Offset(offset).Push(gtx.Ops)
		defer stack.Pop()
		return t.layoutDragPreview(ctx, gtx, label, defaults)
	}
}

func (t Widget) layoutDragPreview(ctx *core.Context, gtx layout.Context, label string, defaults treeDefaults) layout.Dimensions {
	return treeDragPreview(core.ActiveTheme(ctx), gtx, label, defaults)
}

func treeItemHeight(gtx layout.Context, metrics treeGeometry, item Item) int {
	height := metrics.RowHeight
	if treeItemDescription(item) != "" {
		height = metrics.DescriptionRowHeight
	}
	return gtx.Dp(height)
}

func registerTreeDragAreas(gtx layout.Context, state *treeItemState, size image.Point, acceptsChildren, draggable, droppable bool) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	edge := max(size.Y/4, 1)
	breaks := [4]int{0, edge, size.Y - edge, size.Y}
	if !acceptsChildren {
		breaks = [4]int{0, size.Y / 2, size.Y / 2, size.Y}
	}
	if droppable {
		for index := range state.dropTargets {
			if breaks[index] >= breaks[index+1] {
				continue
			}
			state.dropTargets[index].Register(gtx, image.Rect(0, breaks[index], size.X, breaks[index+1]))
		}
	}
	if draggable {
		state.drag.RegisterSource(gtx, image.Rectangle{Max: size})
	}
}

func (t Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, itemState *treeItemState, entry flatItem, rowPaint treeRowPaint, selected, expanded bool, expansion float32, disabled bool, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	left := metrics.RowPaddingX + unit.Dp(entry.depth)*metrics.Indent
	return layout.Inset{
		Top:    metrics.RowPaddingY,
		Right:  metrics.RowPaddingX,
		Bottom: metrics.RowPaddingY,
		Left:   left,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 3)
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutToggle(ctx, gtx, itemState, entry.item, rowPaint, expansion, disabled, defaults)
		}))
		if t.checkable {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return t.layoutCheckbox(ctx, gtx, treeStateValue, itemState, entry.item)
			}))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			content := func(gtx layout.Context) layout.Dimensions {
				semantic.LabelOp(entry.item.Label).Add(gtx.Ops)
				if description := treeSemanticDescription(ctx, entry.item, expanded); description != "" {
					semantic.DescriptionOp(description).Add(gtx.Ops)
				}
				semantic.SelectedOp(selected).Add(gtx.Ops)
				semantic.EnabledOp(!disabled).Add(gtx.Ops)
				return t.layoutMainContent(ctx, gtx, treeStateValue, entry.item, rowPaint, selected, expanded, disabled, defaults)
			}
			if treeStateValue.renameKey == entry.item.Key {
				return content(gtx)
			}
			return itemState.clickable.Layout(gtx, content)
		}))
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Gap:       gtx.Dp(metrics.ContentGap),
		}.Layout(gtx, children...)
	})
}

func (t Widget) layoutCheckbox(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, itemState *treeItemState, item Item) layout.Dimensions {
	if !t.checkable {
		return layout.Dimensions{}
	}
	disabled := t.itemCheckboxDisabled(item)
	checked := t.itemChecked(item.Key)
	halfChecked := t.itemHalfChecked(item.Key)
	content := func(gtx layout.Context) layout.Dimensions {
		semantic.CheckBox.Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		if halfChecked {
			description := ctx.Text(i18n.TreePartiallySelected)
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		selection := float32(0)
		if checked || halfChecked {
			selection = 1
		}
		return toggle.DrawCheckboxControl(ctx, gtx, toggle.CheckboxControlOptions{
			Variant:       toggle.CheckboxPrimary,
			Selection:     selection,
			Indeterminate: halfChecked,
			Hovered:       itemState.check.Hovered(),
			Pressed:       itemState.check.Pressed(),
			Disabled:      disabled,
			VisualCache:   &treeStateValue.checkboxVisual,
		})
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.check.Layout(gtx, content)
}

func (t Widget) layoutToggle(ctx *core.Context, gtx layout.Context, itemState *treeItemState, item Item, rowPaint treeRowPaint, expansion float32, disabled bool, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	slot := min(gtx.Dp(metrics.ChevronSlotSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(slot, 0), max(slot, 0))
	gtx.Constraints = layout.Exact(size)
	if !treeItemExpandable(item) {
		return layout.Dimensions{Size: size}
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.toggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if item.ChildrenState == ChildrenLoading {
			return spinner.New().Color(spinner.Current).Size(spinner.Small).Label(ctx.Text(i18n.TreeLoadingChildren)).Layout(ctx, gtx)
		}
		if item.ChildrenState == ChildrenError {
			treeDrawRetry(gtx, size, metrics, defaults.colors.Retry)
			return layout.Dimensions{Size: size}
		}
		switcher := item.Switcher
		if expansion >= .5 && item.ExpandedSwitcher != nil {
			switcher = item.ExpandedSwitcher
		}
		if switcher != nil {
			restore := core.PushColors(ctx, rowPaint.Chevron, defaults.colors.ToggleBackground)
			defer restore()
			return switcher.Layout(ctx, gtx)
		}
		treeDrawToggle(
			gtx,
			size,
			metrics,
			expansion,
			t.guides && t.guideConnectors,
			rowPaint.Chevron,
			defaults.colors.ToggleBackground,
		)
		return layout.Dimensions{Size: size}
	})
}

func (t Widget) layoutMainContent(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, item Item, rowPaint treeRowPaint, selected, expanded, disabled bool, defaults treeDefaults) layout.Dimensions {
	if item.Content != nil {
		return item.Content.Layout(ctx, gtx)
	}
	metrics := defaults.metrics
	children := make([]layout.FlexChild, 0, 3)
	leading := item.Leading
	if expanded && item.ExpandedLeading != nil {
		leading = item.ExpandedLeading
	}
	if leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return leading.Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return t.layoutItemText(ctx, gtx, treeStateValue, item, rowPaint, selected, disabled, defaults)
	}))
	if item.Trailing != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Trailing.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(metrics.ContentGap),
	}.Layout(gtx, children...)
}

func (t Widget) layoutItemText(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, item Item, rowPaint treeRowPaint, selected, disabled bool, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	if treeStateValue.renameKey == item.Key {
		editor := material.Editor(core.ActiveMaterial(ctx), &treeStateValue.renameEditor, "")
		editor.TextSize = metrics.ItemTextSize
		editor.Color = rowPaint.Foreground
		editor.SelectionColor = defaults.colors.EditorSelection
		editor.Font.Weight = metrics.ItemTextWeight
		return editor.Layout(gtx)
	}
	descriptionText := treeItemDescription(item)
	labelWidget := func(gtx layout.Context) layout.Dimensions {
		return treeDrawText(core.ActiveTheme(ctx), gtx, item.Label, metrics.ItemTextSize, rowPaint.Foreground, metrics.ItemTextWeight)
	}
	if descriptionText == "" {
		return labelWidget(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(labelWidget),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			color := rowPaint.Description
			if item.ChildrenState == ChildrenError {
				color = defaults.colors.Retry
			}
			return treeDrawText(core.ActiveTheme(ctx), gtx, descriptionText, metrics.ItemDescriptionSize, color, 0)
		}),
	)
}

func treeSemanticDescription(ctx *core.Context, item Item, expanded bool) string {
	parts := make([]string, 0, 2)
	if description := treeItemDescription(item); description != "" {
		parts = append(parts, description)
	}
	if treeItemExpandable(item) {
		state := ctx.Text(i18n.TreeCollapsed)
		if expanded {
			state = ctx.Text(i18n.TreeExpanded)
		}
		parts = append(parts, state)
	}
	loadingState := ""
	switch item.ChildrenState {
	case ChildrenUnloaded:
		loadingState = ctx.Text(i18n.TreeChildrenNotLoaded)
	case ChildrenLoading:
		loadingState = ctx.Text(i18n.TreeLoadingChildren)
	case ChildrenError:
		loadingState = ctx.Text(i18n.TreeLoadChildrenFailed)
	}
	if loadingState != "" {
		parts = append(parts, loadingState)
	}
	return strings.Join(parts, ". ")
}

func treeItemDescription(item Item) string {
	if item.ChildrenState == ChildrenError && item.LoadError != "" {
		return item.LoadError
	}
	return item.Description
}
