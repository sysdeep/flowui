package dropdown

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/button"
)

func (d Widget) layoutRoot(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	state := dropdownStateFor(ctx, d.key)
	defaults := d.visualWithCache(ctx, &state.visualCache)
	state.bind(d)
	open := state.isOpen(d)
	if open && state.openSource != OpenChangeContextMenu {
		state.hasContextAnchor = false
	}
	if d.disabled || !gtx.Enabled() {
		open = state.requestOpen(ctx, d, false)
	}
	if !open || core.OverlayInteractive(ctx, core.OverlayLayerPopup, state.key) {
		open = d.handleTrigger(ctx, gtx, state, open, defaults)
	}
	eventGtx := gtx
	if d.disabled {
		eventGtx = eventGtx.Disabled()
	}
	focused := eventGtx.Focused(&state.trigger)
	focusVisible := core.FocusVisible(ctx, &state.trigger, focused)
	dims := layoutui.Styled(ctx, eventGtx, state.key, style.State{
		Hovered:      state.trigger.Hovered(),
		Pressed:      state.trigger.Pressed(),
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     d.disabled || !gtx.Enabled(),
		Open:         open,
	}, d.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutTrigger(ctx, gtx, state, defaults)
	}))
	state.triggerRect = image.Rectangle{Max: dims.Size}

	if open && !state.wasOpen {
		core.ActivateExclusive(ctx, dropdownExclusive, state.key)
	} else if !open {
		core.ReleaseExclusive(ctx, dropdownExclusive, state.key)
	}
	restoreFocus := !open && state.wasOpen && !d.disabled && gtx.Enabled() && !state.skipRestore
	state.observeOpen(open)
	if restoreFocus {
		core.AfterOverlays(ctx, func() {
			if !core.HasTopOverlay(ctx) {
				core.RequestFocusVisible(ctx, &state.trigger, state.focusVisible)
			}
		})
	}

	progress := state.progressFor(gtx, open && !d.disabled, defaults.EnterDuration, defaults.ExitDuration, defaults.Motion)
	if progress > 0 && !state.triggerRect.Empty() {
		d.registerRootOverlay(ctx, state, open, progress, !gtx.Enabled(), defaults)
	}
	return dims
}

func (d Widget) handleTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, defaults dropdownDefaults) bool {
	if d.disabled || !gtx.Enabled() {
		state.gesture.Reset()
		return open
	}
	focusVisible := core.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
	// Opening keys are component policy and must run before Clickable consumes them.
	open = d.handleTriggerKeys(ctx, gtx, state, open)
	base.RowInteraction(ctx, gtx, &state.trigger, false, func() {
		if d.triggerMode != TriggerPress {
			return
		}
		state.focusFirst = !open
		state.focusLast = false
		state.focusVisible = focusVisible
		state.hasContextAnchor = false
		open = state.requestOpenFrom(ctx, d, !open, OpenChangeTrigger)
		core.RequestFocusVisible(ctx, &state.trigger, focusVisible)
	})
	if d.triggerMode == TriggerLongPress {
		open = d.handlePopupGesture(ctx, gtx, state, open, defaults.LongPressDuration)
	} else if d.triggerMode == TriggerHover {
		state.gesture.Reset()
		open = d.handleHover(ctx, gtx, state, open, defaults.HoverOpenDuration, defaults.HoverCloseDuration)
	} else if d.triggerMode == TriggerContextMenu {
		open = d.handlePopupGesture(ctx, gtx, state, open, defaults.LongPressDuration)
	} else {
		state.gesture.Reset()
	}
	return open
}

func (d Widget) handleHover(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, defaultOpen, defaultClose time.Duration) bool {
	if d.disabled || !gtx.Enabled() {
		state.hoverOpenAt = time.Time{}
		state.hoverCloseAt = time.Time{}
		return false
	}

	triggerHovered := state.trigger.Hovered()
	menuHovered := state.menuHovered
	if open && (triggerHovered || menuHovered) {
		state.hoverOpenAt = time.Time{}
		state.hoverCloseAt = time.Time{}
		return open
	}
	if triggerHovered || menuHovered {
		state.hoverCloseAt = time.Time{}
	} else if open {
		if state.hoverCloseAt.IsZero() {
			closeDelay := defaultClose
			if d.hasHoverCloseDelay {
				closeDelay = d.hoverCloseDelay
			}
			state.hoverCloseAt = gtx.Now.Add(closeDelay)
		}
		if gtx.Now.Before(state.hoverCloseAt) {
			gtx.Execute(op.InvalidateCmd{At: state.hoverCloseAt})
			return open
		}
		state.hoverCloseAt = time.Time{}
		state.hoverOpenAt = time.Time{}
		return state.requestOpenFrom(ctx, d, false, OpenChangeOutside)
	}

	if !triggerHovered {
		state.hoverOpenAt = time.Time{}
		return open
	}
	if state.hoverOpenAt.IsZero() {
		openDelay := defaultOpen
		if d.hasHoverOpenDelay {
			openDelay = d.hoverOpenDelay
		}
		state.hoverOpenAt = gtx.Now.Add(openDelay)
	}
	if gtx.Now.Before(state.hoverOpenAt) {
		gtx.Execute(op.InvalidateCmd{At: state.hoverOpenAt})
		return open
	}
	state.hoverOpenAt = time.Time{}
	state.focusFirst = true
	state.focusLast = false
	state.focusVisible = false
	return state.requestOpenFrom(ctx, d, true, OpenChangeTrigger)
}

func (d Widget) handleTriggerKeys(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	filters := []event.Filter{
		key.Filter{Focus: &state.trigger, Name: key.NameDownArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameUpArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameEnter},
		key.Filter{Focus: &state.trigger, Name: key.NameReturn},
		key.Filter{Focus: &state.trigger, Name: key.NameSpace},
		key.Filter{Focus: &state.trigger, Name: key.NameEscape},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return open
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		if event.Name == key.NameEscape {
			if open {
				state.focusFirst = false
				state.focusLast = false
				open = state.requestOpenFrom(ctx, d, false, OpenChangeKeyboard)
			}
			continue
		}

		if open {
			continue
		}
		state.focusVisible = true
		state.focusFirst = event.Name != key.NameUpArrow
		state.focusLast = event.Name == key.NameUpArrow
		state.hasContextAnchor = false
		open = state.requestOpenFrom(ctx, d, true, OpenChangeKeyboard)
	}
}

func (d Widget) handlePopupGesture(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, defaultDuration time.Duration) bool {
	delay := defaultDuration
	if d.hasLongPressDelay {
		delay = d.longPressDelay
	}
	policy := base.PopupTriggerPolicy{LongPress: d.triggerMode == TriggerLongPress, Delay: delay}
	if d.triggerMode == TriggerContextMenu {
		policy.Secondary = true
		policy.Focus = []event.Tag{&state.trigger}
	}
	state.gesture.Update(gtx, policy, func(gesture base.PopupTriggerEvent) {
		source := OpenChangeTrigger
		if d.triggerMode == TriggerContextMenu {
			source = OpenChangeContextMenu
			x, y := int(gesture.Position.X+.5), int(gesture.Position.Y+.5)
			state.contextAnchor = image.Rect(x, y, x+1, y+1)
			state.hasContextAnchor = true
		}
		state.focusFirst, state.focusLast, state.focusVisible = true, false, gesture.Keyboard
		core.RequestFocusVisible(ctx, &state.trigger, gesture.Keyboard)
		open = state.requestOpenFrom(ctx, d, true, source)
	})
	return open
}

func (d Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, defaults dropdownDefaults) layout.Dimensions {
	if trigger, ok := d.trigger.(button.Widget); ok {
		return d.layoutButtonTrigger(ctx, gtx, state, trigger)
	}
	if trigger, ok := d.trigger.(*button.Widget); ok && trigger != nil {
		return d.layoutButtonTrigger(ctx, gtx, state, *trigger)
	}
	animGtx := gtx
	macro := op.Record(gtx.Ops)
	dims := state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		var childDims layout.Dimensions
		if d.trigger != nil {
			pass := pointer.PassOp{}.Push(gtx.Ops)
			childDims = d.trigger.Layout(ctx, gtx)
			pass.Pop()
		}
		call := macro.Stop()
		size := gtx.Constraints.Constrain(childDims.Size)
		focusVisible := core.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
		focus := state.triggerFocus.Opacity(animGtx, focusVisible && !d.disabled, defaults.Motion)
		scale := dropdownTriggerScale(animGtx, state.trigger.History(), defaults, d.disabled)
		transform := draw.Scale(size, scale).Push(gtx.Ops)
		call.Add(gtx.Ops)
		defaults.layoutFocus(gtx, size, focus)
		transform.Pop()
		return layout.Dimensions{Size: size, Baseline: childDims.Baseline}
	})
	call := macro.Stop()
	d.registerTriggerPointerEvents(gtx, state, dims.Size)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	pass.Pop()
	return dims
}

func (d Widget) layoutButtonTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, trigger button.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := button.LayoutWithClickableNoEvents(trigger.Style(d.style), ctx, gtx, &state.trigger)
	call := macro.Stop()
	d.registerTriggerPointerEvents(gtx, state, dims.Size)
	call.Add(gtx.Ops)
	return dims
}

func (d Widget) registerTriggerPointerEvents(gtx layout.Context, state *dropdownState, size image.Point) {
	if (d.triggerMode != TriggerLongPress && d.triggerMode != TriggerContextMenu) || size.X <= 0 || size.Y <= 0 {
		return
	}
	state.gesture.Layout(gtx, size)
}
