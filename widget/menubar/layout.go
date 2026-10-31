package menubar

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type menubarTriggerPlacement struct {
	dims      layout.Dimensions
	call      op.CallOp
	placement core.OverlayPlacement
}

func (m Widget) layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if !gtx.Enabled() {
		m.disabled = true
	}
	state := menubarStateFor(ctx, m.key)
	defaults := m.defaultsWithCache(ctx, &state.visualCache)
	state.bind(m)
	state.beginFrame(m.items)
	defer state.endFrame()

	openKey := state.current(m)
	if !m.hasOpenKey && state.openKey != "" && openKey == "" {
		openKey = state.requestOpen(ctx, m, "")
	}
	state.observeOpen(ctx, openKey)
	state.updateInteractions(ctx, gtx, m)
	openKey = state.current(m)
	state.observeOpen(ctx, openKey)
	progress := state.progress(gtx, openKey != "", defaults.motion)

	macro := op.Record(gtx.Ops)
	var panelAnchor image.Rectangle
	var hasPanelAnchor bool
	hovered, pressed, focused := false, false, false
	for _, item := range m.items {
		trigger := state.trigger(item.key)
		hovered = hovered || trigger.clickable.Hovered()
		pressed = pressed || trigger.clickable.Pressed()
		focused = focused || gtx.Focused(&trigger.clickable)
	}
	root := m.resolveRoot(ctx, gtx, state.key, style.State{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: m.disabled || !gtx.Enabled(),
		Open:     openKey != "",
	}, defaults)
	dims := layoutui.Resolved(ctx, gtx, root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		dims, panelAnchor, hasPanelAnchor = m.layoutTriggers(ctx, gtx, state, openKey, state.panelKey, defaults)
		return dims
	}))
	call := macro.Stop()
	rootClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	semantic.EnabledOp(!m.disabled).Add(gtx.Ops)
	label := m.alt
	if label == "" {
		label = ctx.Text(i18n.MenubarLabel)
	}
	semantic.DescriptionOp(label).Add(gtx.Ops)
	call.Add(gtx.Ops)
	rootClip.Pop()

	if progress > 0 && state.panelKey != "" {
		if item, ok := m.itemByKey(state.panelKey); ok {
			if hasPanelAnchor && !panelAnchor.Empty() {
				m.registerOverlay(ctx, state, item, image.Rectangle{Max: dims.Size}, panelAnchor, openKey == state.panelKey, progress, defaults)
			}
		}
	} else if progress <= 0 {
		state.panelKey = ""
	}
	return dims
}

func (m Widget) layoutTriggers(ctx *core.Context, gtx layout.Context, state *menubarState, openKey, panelKey string, defaults menubarDefaults) (layout.Dimensions, image.Rectangle, bool) {
	axis := layout.Horizontal
	if m.orientation == Vertical {
		axis = layout.Vertical
	}
	var inlinePlacements [16]menubarTriggerPlacement
	var placements []menubarTriggerPlacement
	if len(m.items) > len(inlinePlacements) {
		placements = make([]menubarTriggerPlacement, len(m.items))
	} else {
		placements = inlinePlacements[:len(m.items)]
	}
	gap := max(gtx.Dp(defaults.gap), 0)
	axisMin := axis.Convert(gtx.Constraints.Min)
	axisMax := axis.Convert(gtx.Constraints.Max)
	remaining := axisMax.X - max(len(m.items)-1, 0)*gap
	remaining = max(remaining, 0)
	mainSize := max(len(m.items)-1, 0) * gap
	maxCross := axisMin.Y
	maxBaseline := 0
	for index, item := range m.items {
		childGtx := gtx
		childGtx.Constraints = layout.Constraints{
			Min: axis.Convert(image.Pt(0, axisMin.Y)),
			Max: axis.Convert(image.Pt(remaining, axisMax.Y)),
		}
		macro := op.Record(gtx.Ops)
		dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return m.layoutTrigger(ctx, childGtx, state, item, openKey == item.key, defaults)
		})
		placements[index] = menubarTriggerPlacement{dims: dims, call: macro.Stop(), placement: placement}
		childSize := axis.Convert(dims.Size)
		mainSize += childSize.X
		remaining = max(remaining-childSize.X, 0)
		maxCross = max(maxCross, childSize.Y)
		maxBaseline = max(maxBaseline, dims.Size.Y-dims.Baseline)
	}
	if axisMin.X > mainSize {
		mainSize = axisMin.X
	}
	size := gtx.Constraints.Constrain(axis.Convert(image.Pt(mainSize, maxCross)))
	dims := layout.Dimensions{Size: size, Baseline: size.Y - maxBaseline}
	var panelAnchor image.Rectangle
	hasPanelAnchor := false
	main := 0
	for index, child := range placements {
		size := axis.Convert(child.dims.Size)
		cross := max((maxCross-size.Y)/2, 0)
		position := axis.Convert(image.Pt(main, cross))
		child.placement.PlaceOffset(position)
		offset := op.Offset(position).Push(gtx.Ops)
		child.call.Add(gtx.Ops)
		offset.Pop()
		if m.items[index].key == panelKey {
			panelAnchor = image.Rectangle{Min: position, Max: position.Add(child.dims.Size)}
			hasPanelAnchor = true
		}
		main += size.X
		if index < len(placements)-1 {
			main += gap
		}
	}
	return dims, panelAnchor, hasPanelAnchor
}

func (m Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, state *menubarState, item Item, open bool, defaults menubarDefaults) layout.Dimensions {
	triggerState := state.trigger(item.key)
	disabled := m.itemDisabled(item) || !gtx.Enabled()
	eventGtx := gtx
	if disabled {
		eventGtx = eventGtx.Disabled()
	}
	animGtx := gtx
	focused := gtx.Focused(&triggerState.clickable)
	focusVisible := core.FocusVisible(ctx, &triggerState.clickable, focused)
	resolved := m.resolveTrigger(ctx, gtx, state.key, item.key, style.State{
		Open:         open,
		Hovered:      triggerState.clickable.Hovered(),
		Pressed:      triggerState.clickable.Pressed(),
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
	}, defaults)

	height := 0
	if resolved.Box != nil && resolved.Box.Height != nil {
		height = min(gtx.Dp(*resolved.Box.Height), gtx.Constraints.Max.Y)
	}
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)

	return triggerState.clickable.Layout(eventGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(item.label).Add(gtx.Ops)
		semantic.SelectedOp(open).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)

		focus := triggerState.focus.Opacity(animGtx, focusVisible && !disabled, defaults.motion)
		padding := 0
		if resolved.Box != nil && resolved.Box.Padding != nil {
			padding = max(gtx.Dp(resolved.Box.Padding.Left), 0) + max(gtx.Dp(resolved.Box.Padding.Right), 0)
		}
		contentGtx := gtx
		contentGtx.Constraints.Min = image.Point{}
		contentGtx.Constraints.Max.X = max(contentGtx.Constraints.Max.X-padding, 0)
		macro := op.Record(gtx.Ops)
		foreground := menubarTextColor(resolved, defaults.foreground)
		background := menubarPaintColor(resolved.Paint)
		restore := core.PushColors(ctx, foreground, background)
		var contentDims layout.Dimensions
		if item.trigger != nil {
			contentDims = item.trigger.Layout(ctx, contentGtx)
		} else {
			contentDims = menubarLabelStyle(core.ActiveTheme(ctx), resolved, defaults, item.label).Layout(contentGtx)
		}
		restore()
		contentCall := macro.Stop()
		size := gtx.Constraints.Constrain(image.Pt(contentDims.Size.X+padding, max(height, contentDims.Size.Y)))
		opacity := float32(1)
		if resolved.Paint != nil && resolved.Paint.Opacity != nil {
			opacity = min(max(*resolved.Paint.Opacity, 0), 1)
		}
		opacityStack := paint.PushOpacity(gtx.Ops, opacity)
		menubarRenderTrigger(gtx, size, resolved, focus)
		position := image.Pt(max((size.X-contentDims.Size.X)/2, 0), max((size.Y-contentDims.Size.Y)/2, 0))
		offset := op.Offset(position).Push(gtx.Ops)
		contentCall.Add(gtx.Ops)
		offset.Pop()
		opacityStack.Pop()
		return layout.Dimensions{Size: size, Baseline: contentDims.Baseline + position.Y}
	})
}
