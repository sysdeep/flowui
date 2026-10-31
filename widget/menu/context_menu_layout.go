package menu

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
)

func (c ContextMenuWidget) layoutTrigger(ctx *core.Context, gtx layout.Context, state *contextMenuState, open *bool) layout.Dimensions {
	if !c.disabled && gtx.Enabled() {
		c.updateTrigger(ctx, gtx, state, open)
	} else {
		state.trigger.Reset()
	}
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	if c.trigger != nil {
		dims = c.trigger.Layout(ctx, gtx)
	}
	call := macro.Stop()
	state.triggerSize = dims.Size

	call.Add(gtx.Ops)
	state.trigger.Layout(gtx, dims.Size)
	return dims
}

func (c ContextMenuWidget) updateTrigger(ctx *core.Context, gtx layout.Context, state *contextMenuState, open *bool) {
	targets := append([]event.Tag{c.defaultFocusTarget(state)}, c.focusTargets...)
	state.trigger.Update(gtx, base.PopupTriggerPolicy{
		Focus: targets, Secondary: true, LongPress: !c.longPressDisabled,
		TouchOnly: true, Delay: contextMenuLongPressDelay,
	}, func(gesture base.PopupTriggerEvent) {
		state.focusTarget = c.focusedTarget(gtx, state)
		size := 1
		if gesture.LongPress {
			size = max(gtx.Dp(10), 1)
		}
		x, y := int(gesture.Position.X+.5), int(gesture.Position.Y+.5)
		state.anchor = image.Rect(x, y, x+size, y+size)
		state.hasAnchor = true
		state.focusVisible = gesture.Keyboard
		if !gesture.Keyboard {
			core.RequestFocusVisible(ctx, c.triggerFocusTarget(state), false)
		}
		*open = state.requestOpen(ctx, c, true)
	})
}

func (c ContextMenuWidget) registerOverlay(
	ctx *core.Context,
	state *contextMenuState,
	open bool,
	progress float32,
	disabled bool,
	visual RuntimeVisual,
) {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       state.key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    state.anchor,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen, dismissed := c.handleOverlayEvents(ctx, gtx, state, open, interactive)
			if dismissed {
				core.DismissActiveOverlay(ctx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return c.layoutOverlay(
				ctx,
				gtx,
				state,
				anchor,
				overlayOpen,
				progress,
				interactive && gtx.Enabled(),
				visual,
			)
		},
	})
}

func (c ContextMenuWidget) handleOverlayEvents(ctx *core.Context, gtx layout.Context, state *contextMenuState, open, interactive bool) (bool, bool) {
	switch state.popup.Update(ctx, gtx, base.PopupPolicy{
		Open: open, Interactive: interactive, DismissOutside: true, DismissEscape: true,
	}) {
	case base.PopupDismissOutside:
		state.skipRestore = true
		open = state.requestOpen(ctx, c, false)
		state.transition.Set(0, 0, gtx.Now)
		return open, true
	case base.PopupDismissEscape:
		core.RequestFocusVisible(ctx, c.triggerFocusTarget(state), true)
		open = state.requestOpen(ctx, c, false)
	}
	return open, false
}

func (c ContextMenuWidget) layoutOverlay(
	ctx *core.Context,
	gtx layout.Context,
	state *contextMenuState,
	anchor image.Rectangle,
	open bool,
	progress float32,
	interactive bool,
	visual RuntimeVisual,
) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	rootMenu := c.menu.withDerivedIdentity(state.key, "menu")
	menuState := rootMenu.stateFor(ctx)
	rootMenu = rootMenu.withClose(func(focusVisible bool) {
		menuState.openSubmenu = ""
		core.RequestFocusVisible(ctx, c.triggerFocusTarget(state), focusVisible)
		state.skipRestore = true
		state.requestOpen(ctx, c, false)
	})
	if !open {
		menuState.openSubmenu = ""
	}
	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return rootMenu.layoutRootWithVisual(ctx, panelGtx, menuState, open && (interactive || menuState.openSubmenu != "" || menuState.submenuActive), visual)
	})
	panelCall := macro.Stop()

	metrics := visual.defaults.Metrics
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gtx.Dp(metrics.ContextMenuOffset),
		Placement:        core.Placement{Side: core.SideBottom, Align: core.AlignStart},
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := core.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	baseScale := metrics.EnterScale
	if !open {
		baseScale = metrics.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	slide := core.SlideOffset(gtx.Dp(metrics.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	state.popup.LayoutOutside(gtx, bounds, true, animatedPanel)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	state.popup.LayoutSurface(gtx, image.Rectangle{Max: panelDims.Size})
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive {
		core.AfterOverlays(ctx, func() {
			if core.OverlayBecameTopmost(ctx, core.OverlayLayerPopup, state.key) {
				menuState.focusFirstEntry(ctx, rootMenu, state.focusVisible)
			}
		})
	}
	return layout.Dimensions{Size: bounds}
}
