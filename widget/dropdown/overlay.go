package dropdown

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/overlayarrow"
	"github.com/qianniancn/flowui/widget/menu"
)

func (d Widget) registerRootOverlay(ctx *core.Context, state *dropdownState, open bool, progress float32, disabled bool, defaults dropdownDefaults) {
	anchor := state.triggerRect
	if state.hasContextAnchor {
		anchor = state.contextAnchor
	}
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       state.key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    anchor,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen, dismissed := d.handleOverlayEvents(ctx, gtx, state, open, interactive)
			if dismissed {
				core.DismissActiveOverlay(ctx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return d.layoutRootOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled(), defaults)
		},
	})
}

func (d Widget) handleOverlayEvents(ctx *core.Context, gtx layout.Context, state *dropdownState, open, interactive bool) (bool, bool) {
	intent := state.popup.Update(ctx, gtx, base.PopupPolicy{
		Open: open, Interactive: interactive,
		DismissOutside: true, DismissOnPress: true, DismissEscape: true,
	})
	switch intent {
	case base.PopupDismissOutside:
		state.skipRestore = true
		open = state.requestOpenFrom(ctx, d, false, OpenChangeOutside)
		// Menus disappear on the press, even while the pointer is held down.
		state.transition.Set(0, 0, gtx.Now)
		return open, true
	case base.PopupDismissEscape:
		core.RequestFocusVisible(ctx, &state.trigger, true)
		state.skipRestore = true
		open = state.requestOpenFrom(ctx, d, false, OpenChangeKeyboard)
	}
	return open, false
}

func (d Widget) layoutRootOverlay(ctx *core.Context, gtx layout.Context, state *dropdownState, anchor image.Rectangle, open bool, progress float32, interactive bool, defaults dropdownDefaults) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	menuWidget := d.menu
	if d.matchTriggerWidth {
		menuWidget = menu.WithMinimumWidthPx(menuWidget, anchor.Dx())
	}
	var runtime menu.Runtime
	runtime = menuWidget.RuntimeWithStyle(ctx, state.key, "menu", func(focusVisible bool) {
		runtime.CloseSubmenus()
		core.RequestFocusVisible(ctx, &state.trigger, focusVisible)
		state.skipRestore = true
		state.requestOpenFrom(ctx, d, false, OpenChangeMenu)
	}, d.menuStyle)
	visual := runtime.Visual(ctx)
	if !open {
		runtime.CloseSubmenus()
	}
	state.menuHovered = state.popup.Hovered() || runtime.HoveredWithSubmenus(ctx)

	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return runtime.LayoutWithVisual(ctx, panelGtx, open && (interactive || runtime.HasActiveSubmenu()), visual)
	})
	panelCall := macro.Stop()
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           d.panelGapPx(gtx, defaults),
		Placement:        d.placement.Placement(),
		Flip:             d.flipEnabled(),
		AvoidOverflow:    d.overflowAvoidanceEnabled(),
	})
	placement := result.Placement.PopoverPlacement()
	arrowSize := 0
	arrowAnchor := float32(0)
	arrowRect := image.Rectangle{}
	if d.arrow {
		arrowSize = gtx.Dp(defaults.ArrowSize)
		radius, _ := visual.PanelChrome()
		panelRadius := min(max(gtx.Dp(radius), 0), min(panelDims.Size.X, panelDims.Size.Y)/2)
		arrowAnchor = overlayarrow.ArrowAnchor(anchor, result.Position, panelDims.Size, placement, panelRadius, arrowSize)
		arrowRect = overlayarrow.ArrowRect(panelDims.Size, placement, arrowAnchor, arrowSize)
	}
	motion := visual.Metrics
	baseScale := motion.EnterScale
	if !open {
		baseScale = motion.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	origin := core.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	slide := core.SlideOffset(gtx.Dp(motion.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	animatedArrow := core.AffineRectBounds(arrowRect, panelTransform)
	state.popup.LayoutOutside(gtx, bounds, true, anchor, animatedArrow)
	state.popup.LayoutSurface(gtx, animatedPanel)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	panelCall.Add(gtx.Ops)
	if d.arrow {
		surface, border := visual.PanelColors()
		_, borderWidth := visual.PanelChrome()
		overlayarrow.DrawArrow(gtx, placement, panelDims.Size, arrowAnchor, arrowSize, gtx.Dp(borderWidth), surface, border)
		state.popup.LayoutArrow(gtx, arrowRect)
	}
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive {
		core.AfterOverlays(ctx, func() {
			if !core.OverlayBecameTopmost(ctx, core.OverlayLayerPopup, state.key) {
				return
			}
			if state.focusLast {
				runtime.FocusLast(ctx, state.focusVisible)
			} else if state.focusFirst {
				runtime.FocusFirst(ctx, state.focusVisible)
			}
			state.focusFirst = false
			state.focusLast = false
		})
	}
	return layout.Dimensions{Size: bounds}
}

func (d Widget) panelGapPx(gtx layout.Context, defaults dropdownDefaults) int {
	gap := 0
	if d.hasOffset {
		gap = gtx.Dp(d.offset)
	} else {
		gap = gtx.Dp(defaults.PanelGap)
	}
	if d.arrow {
		gap += gtx.Dp(defaults.ArrowSize * 2 / 3)
	}
	return gap
}
