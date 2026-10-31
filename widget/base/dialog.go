package base

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

// DialogPolicy controls close requests without owning the dialog's open value.
// Close is the optional caller-owned close button; its visual is laid out by
// the component without processing events a second time.
type DialogPolicy struct {
	Open           bool
	Interactive    bool
	DismissOutside bool
	DismissEscape  bool
	Close          *widget.Clickable
}

// DialogInteraction composes pointer-only popup surfaces with a modal focus
// boundary. Retain it between frames. Layout, animation, open binding and the
// lifetime of an exiting overlay remain the component's responsibility.
type DialogInteraction struct {
	PopupInteraction
	start        dialogFocusTag
	target       dialogFocusTag
	end          dialogFocusTag
	open         bool
	focusPending bool
}

// Non-zero size gives each boundary a distinct address when used as an event tag.
type dialogFocusTag struct{ _ byte }

// Update returns at most one close request from Escape, a completed outside
// click or the close button. Visible exit surfaces keep focus inside the dialog
// but cannot request another close. Interactive comes from the overlay host.
func (d *DialogInteraction) Update(ctx *core.Context, gtx layout.Context, policy DialogPolicy) bool {
	events := d.PopupInteraction.update(ctx, gtx, PopupPolicy{
		Open: policy.Open, Interactive: policy.Interactive,
		DismissOutside: policy.DismissOutside, DismissEscape: policy.DismissEscape,
	})
	if !policy.Interactive || !gtx.Enabled() {
		return false
	}
	if events.pressed || events.surfaceClicked {
		core.RequestFocus(ctx, d.FocusTarget())
	}
	dismiss := events.dismiss != PopupNotDismissed
	if policy.Close != nil && policy.Open {
		RowInteraction(ctx, gtx, policy.Close, false, func() { dismiss = true })
	}
	return dismiss
}

// FocusTarget is the neutral initial focus target, before the dialog's controls
// in Tab order. It also holds focus while those controls are disabled on exit.
func (d *DialogInteraction) FocusTarget() event.Tag { return &d.target }

// ManageFocus schedules initial focus and focus-scope recovery after overlays
// have resolved. Call once for each visible frame, including exit animation.
// key is the modal-layer overlay identity, not a separately claimed base key.
func (d *DialogInteraction) ManageFocus(ctx *core.Context, key string, open, enabled bool) {
	if open != d.open {
		d.focusPending = open
	}
	d.open = open
	if !enabled {
		if open {
			d.focusPending = true
		}
		return
	}
	core.AfterOverlays(ctx, func() {
		if !core.OverlayFocusScopeTopmost(ctx, core.OverlayLayerModal, key) {
			return
		}
		becameTopmost := core.OverlayFocusScopeBecameTopmost(ctx, core.OverlayLayerModal, key)
		if !open || d.focusPending || becameTopmost {
			core.RequestFocus(ctx, d.FocusTarget())
			d.focusPending = false
		}
	})
}

// LayoutStart mounts the leading boundary and neutral focus target before the
// dialog's controls. last is the backward-wrap destination, or nil for the
// neutral target when no controls are enabled.
func (d *DialogInteraction) LayoutStart(ctx *core.Context, gtx layout.Context, last event.Tag) {
	if !gtx.Enabled() {
		return
	}
	if last == nil {
		last = d.FocusTarget()
	}
	redirectDialogFocus(ctx, gtx, &d.start, last)
	redirectDialogFocus(ctx, gtx, &d.end, d.FocusTarget())
	for {
		if _, ok := gtx.Event(key.FocusFilter{Target: d.FocusTarget()}); !ok {
			break
		}
	}
	layoutDialogFocusTag(gtx, &d.start)
	layoutDialogFocusTag(gtx, d.FocusTarget())
}

// LayoutEnd is the overlay Tail: mount it after descendant popups, not merely
// after the dialog's own content, so nested controls stay inside the boundary.
func (d *DialogInteraction) LayoutEnd(gtx layout.Context) {
	if gtx.Enabled() {
		layoutDialogFocusTag(gtx, &d.end)
	}
}

func redirectDialogFocus(ctx *core.Context, gtx layout.Context, target, redirect event.Tag) {
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: target})
		if !ok {
			return
		}
		if event, ok := e.(key.FocusEvent); ok && event.Focus {
			core.RequestFocus(ctx, redirect)
		}
	}
}

func layoutDialogFocusTag(gtx layout.Context, target event.Tag) {
	stack := clip.Rect{Max: image.Pt(1, 1)}.Push(gtx.Ops)
	event.Op(gtx.Ops, target)
	stack.Pop()
}
