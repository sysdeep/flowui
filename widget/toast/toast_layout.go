package toast

import (
	"image"
	"sort"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/spinner"
	textui "github.com/qianniancn/flowui/widget/text"
)

type toastRecord struct {
	entry    *toastEntryState
	call     op.CallOp
	dims     layout.Dimensions
	progress float32
	index    int
	stack    float32
	exiting  bool
}

func (p Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, hostState *toastHostState, baseVisual toastVisual) layout.Dimensions {
	viewport := gtx.Constraints.Max
	if viewport.X <= 0 || viewport.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(viewport)

	p.handleEvents(ctx, gtx, hostState)
	hostState.updateRegionEvents(gtx)
	expanded := hostState.regionHover.Hovered || hostState.paused(gtx)
	motion := baseVisual.motion
	expansion := hostState.expansionProgress(gtx, expanded, baseVisual.metrics.AnimationDuration, motion)
	hostState.updateTimers(gtx, p.paused || expanded || expansion > 0, func(entry *toastEntryState) {
		entry.requestClose(p.onClose)
	})

	maxVisible := p.resolvedMaxVisible(baseVisual)
	inset := gtx.Dp(baseVisual.metrics.Inset)
	offset := gtx.Dp(p.resolvedOffset(baseVisual))
	maxWidth := min(gtx.Dp(p.resolvedWidth(baseVisual)), max(viewport.X-2*inset, 0))
	if maxWidth <= 0 {
		return layout.Dimensions{Size: viewport}
	}
	mobile := viewport.X <= gtx.Dp(768)
	records := make([]toastRecord, 0, maxVisible+1)
	presentIndex := 0
	frontHeight := 0
	restoreKey := core.PushKey(ctx, p.key)
	defer restoreKey()
	for _, key := range hostState.order {
		entry := hostState.entry(key)
		if entry == nil {
			continue
		}
		exiting := !entry.present || entry.closeRequested
		targetIndex := entry.stack.Target()
		if !exiting {
			targetIndex = float32(presentIndex)
			presentIndex++
		} else if !entry.stack.Ready() {
			targetIndex = float32(presentIndex)
		}
		stackPosition := entry.stackPosition(gtx, targetIndex, baseVisual.metrics.AnimationDuration, motion)
		progress := entry.progress(gtx, baseVisual.metrics.AnimationDuration, motion)
		if progress <= 0 && (!entry.present || entry.closeRequested) {
			continue
		}
		if targetIndex >= float32(maxVisible) && stackPosition >= float32(maxVisible) {
			continue
		}

		toastGtx := gtx
		toastGtx.Constraints.Min = image.Pt(0, 0)
		toastGtx.Constraints.Max = image.Pt(maxWidth, max(viewport.Y-inset-offset, 0))
		if !entry.item.compact {
			toastGtx.Constraints.Min.X = maxWidth
		}
		expandedLayout := expanded || expansion > 0
		interactive := !exiting && gtx.Enabled() && (targetIndex == 0 || expanded)
		if !expandedLayout && targetIndex != 0 && frontHeight > 0 {
			height := min(frontHeight, toastGtx.Constraints.Max.Y)
			toastGtx.Constraints.Min.Y = height
			toastGtx.Constraints.Max.Y = height
		}
		restoreItem := core.PushKey(ctx, entry.item.key)
		visual := p.visualWithCache(ctx, entry.item.variant, entry.item.actionVariant)
		state := style.State{
			Hovered:      entry.hover.Hovered,
			Focused:      toastGtx.Focused(&entry.action) || toastGtx.Focused(&entry.close),
			FocusVisible: core.FocusVisible(ctx, &entry.root, toastGtx.Focused(&entry.root)),
			Disabled:     !interactive,
			Expanded:     expanded,
			Loading:      entry.item.loading,
		}
		resolved := p.resolveStyle(ctx, toastGtx, state, visual)
		macro := op.Record(gtx.Ops)
		dims := p.layoutToast(ctx, toastGtx, entry, interactive, mobile, expandedLayout || hostState.touchMode, visual, resolved)
		call := macro.Stop()
		restoreItem()
		if targetIndex == 0 && !exiting {
			frontHeight = dims.Size.Y
		}
		records = append(records, toastRecord{
			entry:    entry,
			call:     call,
			dims:     dims,
			progress: progress,
			index:    int(targetIndex),
			stack:    stackPosition,
			exiting:  exiting,
		})
	}

	bounds := p.paintRecords(ctx, gtx, viewport, records, expansion, baseVisual)
	hostState.addRegionInput(gtx, bounds)
	return layout.Dimensions{Size: viewport}
}

func (p Widget) handleEvents(ctx *core.Context, gtx layout.Context, hostState *toastHostState) {
	for _, keyValue := range hostState.order {
		entry := hostState.entry(keyValue)
		if entry == nil {
			continue
		}
		entry.updateRootEvents(gtx)
		for {
			_, ok := gtx.Event(key.FocusFilter{Target: &entry.root})
			if !ok {
				break
			}
		}
		disabled := !entry.present || entry.closeRequested
		base.UpdateRow(ctx, gtx, base.RowInput{
			Clickable: &entry.close, Disabled: disabled, PreserveFocus: true,
			OnClick: func(widget.Click) { entry.requestClose(p.onClose) },
		})
		base.UpdateRow(ctx, gtx, base.RowInput{
			Clickable: &entry.action, Disabled: disabled, PreserveFocus: true,
			OnClick: func(widget.Click) {
				if p.onAction != nil {
					p.onAction(entry.item.key)
				}
			},
		})
		if base.EscapePressed(gtx, &entry.root) {
			entry.requestClose(p.onClose)
		}
	}
}

func (p Widget) paintRecords(ctx *core.Context, gtx layout.Context, viewport image.Point, records []toastRecord, expansion float32, baseVisual toastVisual) image.Rectangle {
	if len(records) == 0 {
		return image.Rectangle{}
	}
	motion := baseVisual.motion
	inset := gtx.Dp(baseVisual.metrics.Inset)
	offset := gtx.Dp(p.resolvedOffset(baseVisual))
	gap := gtx.Dp(p.resolvedGap(baseVisual))
	scaleFactor := p.resolvedScaleFactor(baseVisual)
	frontHeight := toastFrontHeight(records)
	expandedOffsets := toastExpandedOffsets(records, gap)
	paintOrder := append([]toastRecord(nil), records...)
	sort.SliceStable(paintOrder, func(i, j int) bool {
		if paintOrder[i].stack != paintOrder[j].stack {
			return paintOrder[i].stack > paintOrder[j].stack
		}
		return !paintOrder[i].exiting && paintOrder[j].exiting
	})

	var bounds image.Rectangle
	for _, record := range paintOrder {
		height := record.dims.Size.Y
		if frontHeight > 0 {
			height = int(animation.LerpFloat(float32(frontHeight), float32(height), expansion) + 0.5)
		}
		visibleSize := image.Pt(record.dims.Size.X, height)
		collapsedScale := max(1-record.stack*scaleFactor, 0.5)
		scale := animation.LerpFloat(collapsedScale, 1, expansion)
		x := toastRegionX(viewport.X, record.dims.Size.X, inset, p.placement)
		collapsedOffset := record.stack * float32(gap)
		expandedTarget := toastExpandedOffset(expandedOffsets, float32(record.index))
		expandedOffset := record.entry.expandedPosition(gtx, expandedTarget, baseVisual.metrics.AnimationDuration, motion)
		stackOffset := int(animation.LerpFloat(collapsedOffset, expandedOffset, expansion) + 0.5)
		y := offset + stackOffset
		enterOffset := -int(float32(height) * (1 - record.progress))
		if p.bottomPlacement() {
			y = viewport.Y - offset - height - stackOffset
			enterOffset = -enterOffset
		}
		y += enterOffset
		rect := image.Rectangle{Min: image.Pt(x, y), Max: image.Pt(x, y).Add(visibleSize)}
		if bounds.Empty() {
			bounds = rect
		} else {
			bounds = bounds.Union(rect)
		}

		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		transform := draw.Scale(visibleSize, scale).Push(gtx.Ops)
		opacity := paint.PushOpacity(gtx.Ops, record.progress)
		record.call.Add(gtx.Ops)
		opacity.Pop()
		transform.Pop()
		offset.Pop()
	}
	return bounds
}

func toastFrontHeight(records []toastRecord) int {
	for _, record := range records {
		if record.index == 0 && !record.exiting {
			return record.dims.Size.Y
		}
	}
	return 0
}

func toastExpandedOffsets(records []toastRecord, gap int) []float32 {
	maxIndex := 0
	for _, record := range records {
		maxIndex = max(maxIndex, record.index)
	}
	heights := make([]int, maxIndex+1)
	for _, record := range records {
		if record.index < 0 || record.index >= len(heights) {
			continue
		}
		if heights[record.index] == 0 || !record.exiting {
			heights[record.index] = record.dims.Size.Y
		}
	}
	offsets := make([]float32, len(heights))
	for index := 1; index < len(offsets); index++ {
		offsets[index] = offsets[index-1] + float32(heights[index-1]+gap)
	}
	return offsets
}

func toastExpandedOffset(offsets []float32, stack float32) float32 {
	if len(offsets) == 0 || stack <= 0 {
		return 0
	}
	last := len(offsets) - 1
	if stack >= float32(last) {
		return offsets[last]
	}
	lower := int(stack)
	progress := stack - float32(lower)
	return animation.LerpFloat(offsets[lower], offsets[lower+1], progress)
}

func (p Widget) layoutToast(ctx *core.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile, expanded bool, visual toastVisual, resolved toastResolvedStyle) layout.Dimensions {
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		if entry.item.title != "" {
			semantic.LabelOp(entry.item.title).Add(gtx.Ops)
		}
		if entry.item.description != "" {
			semantic.DescriptionOp(entry.item.description).Add(gtx.Ops)
		}
		return p.layoutToastContent(ctx, gtx, entry, interactive, mobile, visual, resolved)
	})
	macro := op.Record(gtx.Ops)
	dims := layoutui.Resolved(ctx, gtx, resolved.panel, content)
	call := macro.Stop()
	call.Add(gtx.Ops)
	if !interactive {
		entry.hover.Hovered = false
		return dims
	}
	entry.addRootInput(gtx, dims.Size)
	if !entry.item.compact {
		p.layoutToastClose(ctx, gtx, entry, dims.Size, mobile || expanded, visual)
	}
	return dims
}

func (p Widget) layoutToastContent(ctx *core.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile bool, visual toastVisual, resolved toastResolvedStyle) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if entry.item.showIndicator() {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastIndicator(ctx, gtx, entry.item, visual, resolved.indicator)
		}))
	}
	if entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastText(ctx, gtx, entry, false, visual, resolved)
		}))
	} else {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastText(ctx, gtx, entry, interactive && mobile, visual, resolved)
		}))
		if interactive && !mobile && entry.item.actionLabel != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry, visual)
			}))
		}
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(visual.metrics.ContentGap),
	}.Layout(gtx, children...)
}

func (p Widget) layoutToastText(ctx *core.Context, gtx layout.Context, entry *toastEntryState, mobileAction bool, visual toastVisual, resolved toastResolvedStyle) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if entry.item.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved.title, textui.New(entry.item.title))
		}))
	}
	if entry.item.description != "" && !entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved.description, textui.New(entry.item.description))
		}))
	}
	if mobileAction && entry.item.actionLabel != "" && !entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: visual.metrics.ActionInsetTop}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry, visual)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (p Widget) layoutToastIndicator(ctx *core.Context, gtx layout.Context, item Item, visual toastVisual, resolved style.ResolvedStyle) layout.Dimensions {
	outer := max(gtx.Dp(visual.metrics.IndicatorSize+2*visual.metrics.IndicatorPadding), 1)
	gtx.Constraints = layout.Exact(image.Pt(outer, outer))
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := max(gtx.Dp(visual.metrics.IndicatorSize), 1)
		gtx.Constraints = layout.Exact(image.Pt(size, size))
		if item.loading {
			return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return spinner.Draw(gtx, core.ActiveTheme(ctx), visual.metrics.IndicatorSize, ctx.ForegroundColor())
			}))
		}
		if item.hasIndicator && item.indicator != nil {
			if !gtx.Enabled() {
				gtx = gtx.Disabled()
			}
			return item.indicator.Layout(ctx, gtx)
		}
		return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return icon.Layout(visual.icon, gtx, ctx.ForegroundColor())
		}))
	})
}

func (p Widget) layoutToastAction(ctx *core.Context, gtx layout.Context, entry *toastEntryState, visual toastVisual) layout.Dimensions {
	gtx.Constraints.Min.X = 0
	state := style.State{
		Hovered: entry.action.Hovered(),
		Pressed: entry.action.Pressed(),
		Focused: gtx.Focused(&entry.action),
	}
	resolved := p.resolveActionStyle(ctx, gtx, state, visual)
	return entry.action.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(true).Add(gtx.Ops)
		semantic.LabelOp(entry.item.actionLabel).Add(gtx.Ops)
		return layoutui.Resolved(ctx, gtx, resolved, textui.New(entry.item.actionLabel))
	})
}

func (p Widget) layoutToastClose(ctx *core.Context, gtx layout.Context, entry *toastEntryState, toastSize image.Point, alwaysVisible bool, visual toastVisual) {
	show := alwaysVisible || entry.hover.Hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close)
	if !show {
		return
	}
	size := gtx.Dp(visual.metrics.CloseButtonSize)
	inset := gtx.Dp(visual.metrics.CloseInset)
	position := image.Pt(toastSize.X-size-inset, inset)
	closeGtx := gtx
	closeGtx.Constraints = layout.Exact(image.Pt(size, size))
	state := style.State{
		Hovered: entry.close.Hovered(),
		Pressed: entry.close.Pressed(),
		Focused: gtx.Focused(&entry.close),
	}
	resolved := p.resolveCloseStyle(ctx, gtx, state, visual)
	offset := op.Offset(position).Push(gtx.Ops)
	entry.close.Layout(closeGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(true).Add(gtx.Ops)
		semantic.LabelOp(ctx.Text("Close")).Add(gtx.Ops)
		return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			iconSize := max(size/2, 1)
			iconGtx := gtx
			iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
			return layout.Center.Layout(iconGtx, func(gtx layout.Context) layout.Dimensions {
				return icon.Layout(lucide.X, gtx, ctx.ForegroundColor())
			})
		}))
	})
	offset.Pop()
}

func (p Widget) resolvedOffset(visual toastVisual) unit.Dp {
	if p.hasOffset {
		return p.offset
	}
	return visual.metrics.Inset
}

func (p Widget) resolvedGap(visual toastVisual) unit.Dp {
	if p.hasGap {
		return p.gap
	}
	return visual.metrics.Gap
}

func (p Widget) resolvedMaxVisible(visual toastVisual) int {
	if p.hasMaxVisible {
		return p.maxVisible
	}
	return max(visual.metrics.MaxVisible, 1)
}

func (p Widget) resolvedScaleFactor(visual toastVisual) float32 {
	if p.hasScale {
		return p.scaleFactor
	}
	return min(max(visual.metrics.ScaleFactor, 0), 1)
}

func (p Widget) resolvedWidth(visual toastVisual) unit.Dp {
	if p.hasWidth {
		return p.width
	}
	return visual.metrics.Width
}

func (p Widget) bottomPlacement() bool {
	return p.placement == Bottom || p.placement == BottomStart || p.placement == BottomEnd
}

func toastRegionX(viewportWidth, toastWidth, inset int, placement Placement) int {
	switch placement {
	case BottomStart, TopStart:
		return inset
	case BottomEnd, TopEnd:
		return max(viewportWidth-inset-toastWidth, 0)
	default:
		return max((viewportWidth-toastWidth)/2, 0)
	}
}
