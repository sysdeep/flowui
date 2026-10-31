package menubar

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/menu"
)

func (m Widget) registerOverlay(ctx *core.Context, state *menubarState, item Item, bar, trigger image.Rectangle, open bool, progress float32, defaults menubarDefaults) {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       state.key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    bar,
		HasAnchor: true,
		Disabled:  m.disabled,
		Layout: func(gtx layout.Context, resolvedBar image.Rectangle, interactive bool) layout.Dimensions {
			resolvedTrigger := menubarResolvedChildRect(bar, resolvedBar, trigger)
			overlayOpen, dismissed := m.handleOverlayEvents(ctx, gtx, state, item.key, open, interactive)
			if dismissed {
				core.DismissActiveOverlay(ctx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return m.layoutOverlay(ctx, gtx, state, item, resolvedBar, resolvedTrigger, overlayOpen, progress, interactive && gtx.Enabled(), defaults)
		},
	})
}

func (m Widget) handleOverlayEvents(ctx *core.Context, gtx layout.Context, state *menubarState, itemKey string, open, interactive bool) (bool, bool) {
	switch state.popup.Update(ctx, gtx, base.PopupPolicy{
		Open: open, Interactive: interactive, DismissOutside: true, DismissOnPress: true, DismissEscape: true,
	}) {
	case base.PopupDismissOutside:
		state.requestOpen(ctx, m, "")
		state.focusPanelKey = ""
		state.transition.Set(0, 0, gtx.Now)
		return false, true
	case base.PopupDismissEscape:
		state.requestOpen(ctx, m, "")
		state.focusPanelKey = ""
		state.focusTrigger(ctx, itemKey, true)
		return false, false
	}
	return open, false
}

func (m Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, state *menubarState, item Item, bar, trigger image.Rectangle, open bool, progress float32, interactive bool, defaults menubarDefaults) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	content := item.menu
	if m.compact {
		content = content.Compact(true)
	}
	if m.disabled || item.disabled {
		content = content.Disabled(true)
	}

	var runtime menu.Runtime
	runtime = content.Runtime(ctx, state.key, "menu:"+item.key, func(focusVisible bool) {
		runtime.CloseSubmenus()
		state.requestOpen(ctx, m, "")
		state.focusPanelKey = ""
		state.focusTrigger(ctx, item.key, focusVisible)
	})
	if m.orientation == Horizontal {
		runtime = runtime.RootNavigation(
			func() {
				runtime.CloseSubmenus()
				m.switchFromMenu(ctx, gtx, state, item.key, -1)
			},
			func() {
				runtime.CloseSubmenus()
				m.switchFromMenu(ctx, gtx, state, item.key, 1)
			},
		)
	}
	if !open {
		runtime.CloseSubmenus()
	}
	menuVisual := runtime.Visual(ctx)

	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return runtime.LayoutWithVisual(ctx, panelGtx, open && (interactive || runtime.HasActiveSubmenu()), menuVisual)
	})
	panelCall := macro.Stop()

	placement := core.Placement{Side: core.SideBottom, Align: core.AlignStart}
	if m.orientation == Vertical {
		placement.Side = core.SideRight
	}
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gtx.Dp(defaults.panelGap),
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	baseScale := menuVisual.Metrics.EnterScale
	if !open {
		baseScale = menuVisual.Metrics.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	origin := core.PanelTransformOriginAt(trigger, result.Position, panelDims.Size, result.Placement)
	slide := core.SlideOffset(gtx.Dp(menuVisual.Metrics.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	state.popup.LayoutOutside(gtx, bounds, !m.modal, bar, animatedPanel)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	state.popup.LayoutSurface(gtx, image.Rectangle{Max: panelDims.Size})
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive && state.focusPanelKey == item.key {
		core.AfterOverlays(ctx, func() {
			if !core.OverlayTopmost(ctx, core.OverlayLayerPopup, state.key) {
				return
			}
			focused := false
			if state.focusLast {
				focused = runtime.FocusLast(ctx, state.focusVisible)
			} else {
				focused = runtime.FocusFirst(ctx, state.focusVisible)
			}
			if focused {
				state.focusPanelKey = ""
				state.focusLast = false
			}
		})
	}
	return layout.Dimensions{Size: bounds}
}

func (m Widget) switchFromMenu(ctx *core.Context, gtx layout.Context, state *menubarState, currentKey string, delta int) {
	current := m.indexOf(currentKey)
	next := m.moveIndex(current, delta)
	if next < 0 || next >= len(m.items) || next == current {
		return
	}
	key := m.items[next].key
	if state.requestOpen(ctx, m, key) != key {
		return
	}
	state.panelKey = key
	state.focusPanelKey = key
	state.focusLast = false
	state.focusVisible = true
	gtx.Execute(op.InvalidateCmd{})
}

func menubarResolvedChildRect(localParent, resolvedParent, localChild image.Rectangle) image.Rectangle {
	if localParent.Empty() || resolvedParent.Empty() || localChild.Empty() {
		return image.Rectangle{}
	}
	mapX := func(value int) int {
		offset := value - localParent.Min.X
		return resolvedParent.Min.X + int(float64(offset)*float64(resolvedParent.Dx())/float64(localParent.Dx())+0.5)
	}
	mapY := func(value int) int {
		offset := value - localParent.Min.Y
		return resolvedParent.Min.Y + int(float64(offset)*float64(resolvedParent.Dy())/float64(localParent.Dy())+0.5)
	}
	return image.Rect(mapX(localChild.Min.X), mapY(localChild.Min.Y), mapX(localChild.Max.X), mapY(localChild.Max.Y))
}
