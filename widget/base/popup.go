package base

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/internal/core"
)

// PopupState is the state a Popup reports to its callbacks.
type PopupState struct {
	Key  string
	Open bool
}

// PopupWidget anchors a panel to a trigger without owning selection or item
// behavior. The panel is registered as a popup-layer overlay and placed with
// the shared ResolvePosition logic.
type PopupWidget struct {
	key       string
	open      bool
	disabled  bool
	placement core.PopoverPlacement
	gap       unit.Dp
	onDismiss func()
	trigger   func(PopupState) core.Widget
	panel     func(PopupState) core.Widget
}

// Popup creates a controlled popup around a trigger.
func Popup(key string, open bool) PopupWidget {
	return PopupWidget{key: key, open: open}
}

func (p PopupWidget) Disabled(disabled bool) PopupWidget { p.disabled = disabled; return p }
func (p PopupWidget) Placement(placement core.PopoverPlacement) PopupWidget {
	p.placement = placement
	return p
}
func (p PopupWidget) Gap(gap unit.Dp) PopupWidget { p.gap = gap; return p }
func (p PopupWidget) Trigger(fn func(PopupState) core.Widget) PopupWidget {
	p.trigger = fn
	return p
}
func (p PopupWidget) Panel(fn func(PopupState) core.Widget) PopupWidget {
	p.panel = fn
	return p
}

// OnDismiss runs when the user presses outside the panel or Escape.
func (p PopupWidget) OnDismiss(fn func()) PopupWidget {
	p.onDismiss = fn
	return p
}

func (p PopupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if p.trigger == nil {
		panic("flowui: base.Popup requires Trigger")
	}
	popupKey := core.ClaimKey(ctx, core.KindBasePopup, p.key)
	public := PopupState{Key: popupKey, Open: p.open}
	disabled := p.disabled || !gtx.Enabled()
	var dims layout.Dimensions
	_, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		dims = p.trigger(public).Layout(ctx, gtx)
		if p.open && !disabled && p.panel != nil {
			anchor := image.Rectangle{Max: dims.Size}
			p.registerPanel(ctx, popupKey, anchor, public)
		}
		return dims
	})
	placement.PlaceOffset(image.Point{})
	return dims
}

func (p PopupWidget) registerPanel(ctx *core.Context, key string, anchor image.Rectangle, public PopupState) {
	interaction := core.UseState[PopupInteraction](ctx, key, "base-popup-interaction")
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       key + "-panel",
		Layer:     core.OverlayLayerPopup,
		Anchor:    anchor,
		HasAnchor: true,
		Tail:      func(layout.Context) {},
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			if interaction.Update(ctx, gtx, PopupPolicy{
				Open: p.open, Interactive: interactive,
				DismissOutside: p.onDismiss != nil, DismissOnPress: true,
				DismissEscape: p.onDismiss != nil,
			}) != PopupNotDismissed {
				p.onDismiss()
			}
			panel := p.panel(public)
			if panel == nil {
				return layout.Dimensions{}
			}
			viewport := core.OverlayViewport(ctx, gtx.Constraints.Max)
			panelGtx := gtx
			panelGtx.Constraints = layout.Constraints{Max: viewport}
			macro := op.Record(gtx.Ops)
			dims := panel.Layout(ctx, panelGtx)
			call := macro.Stop()
			gap := gtx.Dp(p.gap)
			position := core.ResolvePosition(core.PositionConfig{
				Trigger:          anchor.Size(),
				TriggerOrigin:    anchor.Min,
				HasTriggerOrigin: true,
				Panel:            dims.Size,
				Bounds:           viewport,
				Offset:           gap,
				Placement:        p.placement.Placement(),
				Flip:             true,
				AvoidOverflow:    true,
			})
			panelRect := image.Rectangle{Min: position.Position, Max: position.Position.Add(dims.Size)}
			interaction.LayoutOutside(gtx, viewport, false, panelRect)
			offset := op.Offset(position.Position).Push(gtx.Ops)
			interaction.LayoutSurface(gtx, image.Rectangle{Max: dims.Size})
			call.Add(gtx.Ops)
			offset.Pop()
			return layout.Dimensions{Size: viewport}
		},
	})
}
