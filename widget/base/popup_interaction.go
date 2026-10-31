package base

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
)

// PopupDismiss is a close intent, not a change to the component's open value.
type PopupDismiss uint8

const (
	PopupNotDismissed PopupDismiss = iota
	PopupDismissOutside
	PopupDismissEscape
)

// PopupPolicy separates input ownership from visibility. An exiting panel can
// still block pointers without accepting another close request.
type PopupPolicy struct {
	Open           bool
	Interactive    bool
	DismissOutside bool
	DismissOnPress bool
	DismissEscape  bool
	// SurfaceFocus optionally routes panel presses/clicks to a composite's
	// retained focus target. Nil preserves the existing focus.
	SurfaceFocus event.Tag
}

// PopupInteraction owns pointer-only popup surfaces and dismissal input. The
// component retains its value binding, placement, animation and focus-return
// policy. Its zero value is ready to use and must be retained between frames.
type PopupInteraction struct {
	outside [16]core.ClickArea
	surface core.ClickArea
	arrow   core.ClickArea
}

// Update processes popup policy input only for the overlay host's input owner.
// Call before laying out the surfaces so presses preserve the existing focus.
// At most one close intent is returned per update, including controlled popups.
func (p *PopupInteraction) Update(ctx *core.Context, gtx layout.Context, policy PopupPolicy) PopupDismiss {
	return p.update(ctx, gtx, policy).dismiss
}

type popupEvents struct {
	dismiss        PopupDismiss
	pressed        bool
	surfaceClicked bool
}

func (p *PopupInteraction) update(ctx *core.Context, gtx layout.Context, policy PopupPolicy) popupEvents {
	if !policy.Interactive {
		return popupEvents{}
	}
	surface := drainPopupArea(gtx, &p.surface)
	arrow := drainPopupArea(gtx, &p.arrow)
	pressed := surface.pressed || arrow.pressed
	outside := false
	for i := range p.outside {
		activity := drainPopupArea(gtx, &p.outside[i])
		pressed = pressed || activity.pressed
		dismissPress := policy.DismissOnPress && activity.pressed
		outside = outside || activity.clicked || dismissPress
	}
	if pressed && gtx.Enabled() {
		core.PreserveFocus(ctx)
	}
	if gtx.Enabled() && policy.SurfaceFocus != nil && (surface.pressed || surface.clicked) {
		core.RequestFocusVisible(ctx, policy.SurfaceFocus, false)
	}
	events := popupEvents{pressed: pressed, surfaceClicked: surface.clicked}
	if !gtx.Enabled() || !policy.Open {
		return events
	}
	if outside && policy.DismissOutside {
		events.dismiss = PopupDismissOutside
		return events
	}
	if policy.DismissEscape {
		for {
			e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
			if !ok {
				break
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				events.dismiss = PopupDismissEscape
				return events
			}
		}
	}
	return events
}

type popupAreaActivity struct{ clicked, pressed bool }

func drainPopupArea(gtx layout.Context, area *core.ClickArea) popupAreaActivity {
	var activity popupAreaActivity
	for area.Clicked(gtx) {
		activity.clicked = true
	}
	activity.pressed = area.TakePressed()
	return activity
}

// Hovered reports whether the pointer is over the panel or its arrow.
func (p *PopupInteraction) Hovered() bool { return p.surface.Hovered() || p.arrow.Hovered() }

// LayoutOutside mounts dismissal regions in viewport coordinates. Excluded
// regions can include the trigger, animated panel and arrow. Pass-through is
// useful for menus whose background press also reaches another pointer target.
func (p *PopupInteraction) LayoutOutside(
	gtx layout.Context,
	bounds image.Point,
	passThrough bool,
	excluded ...image.Rectangle,
) {
	areas := core.DismissRectsExcluding(image.Rectangle{Max: bounds}, excluded...)
	for i, area := range areas {
		if i >= len(p.outside) {
			break
		}
		if area.Empty() {
			continue
		}
		if passThrough {
			pass := pointer.PassOp{}.Push(gtx.Ops)
			layoutPopupArea(gtx, &p.outside[i], area)
			pass.Pop()
		} else {
			layoutPopupArea(gtx, &p.outside[i], area)
		}
	}
}

// LayoutSurface mounts the pointer blocker behind the panel's interactive
// children. Coordinates are relative to the current ops transform.
func (p *PopupInteraction) LayoutSurface(gtx layout.Context, rect image.Rectangle) {
	layoutPopupArea(gtx, &p.surface, rect)
}

// LayoutArrow mounts a pointer blocker for an arrow outside the panel bounds.
func (p *PopupInteraction) LayoutArrow(gtx layout.Context, rect image.Rectangle) {
	layoutPopupArea(gtx, &p.arrow, rect)
}

func layoutPopupArea(gtx layout.Context, area *core.ClickArea, rect image.Rectangle) {
	if rect.Empty() {
		return
	}
	gtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	area.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: rect.Size()}
	})
	offset.Pop()
}
