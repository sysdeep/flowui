package base_test

import (
	"image"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestDialogInteractionClosePolicy(t *testing.T) {
	for _, tt := range []struct {
		name     string
		policy   base.DialogPolicy
		disabled bool
		want     int
	}{
		{name: "open owner", policy: base.DialogPolicy{Open: true, Interactive: true, DismissEscape: true}, want: 1},
		{name: "not owner", policy: base.DialogPolicy{Open: true, DismissEscape: true}},
		{name: "exiting", policy: base.DialogPolicy{Interactive: true, DismissEscape: true}},
		{name: "disabled", policy: base.DialogPolicy{Open: true, Interactive: true, DismissEscape: true}, disabled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 200), theme.Light())
			var dialog base.DialogInteraction
			var close widget.Clickable
			policy := tt.policy
			policy.Close = &close
			requests := 0
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if tt.disabled {
					gtx = gtx.Disabled()
				}
				if dialog.Update(ctx, gtx, policy) {
					requests++
				}
				return layout.Dimensions{}
			})
			h.Frame(view)
			// Simultaneous close sources must produce one request, not one per source.
			close.Click()
			h.Key(key.NameEscape, 0)
			h.Frame(view)
			if requests != tt.want {
				t.Fatalf("close requests = %d, want %d", requests, tt.want)
			}
		})
	}
}

func TestDialogInteractionCloseButtonKeyboard(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	var dialog base.DialogInteraction
	var close widget.Clickable
	requests := 0
	view := dialogTestView(&dialog, &close, func() { requests++ })
	h.Frame(view)
	h.Frame(view)
	h.Router().MoveFocus(key.FocusBackward)
	h.Frame(view)
	if !h.Router().Source().Focused(&close) {
		t.Fatal("Shift+Tab did not wrap to the close button")
	}
	h.Key(key.NameSpace, 0)
	h.Frame(view)
	if requests != 1 {
		t.Fatalf("keyboard close requests = %d, want 1", requests)
	}
	if !core.FocusVisible((*core.Context)(h.Context()), &close, true) {
		t.Fatal("keyboard close lost its focus-visible modality")
	}
}

func TestDialogInteractionWithoutControlsKeepsFocusInside(t *testing.T) {
	for _, tt := range []struct {
		name      string
		direction key.FocusDirection
	}{
		{name: "Tab", direction: key.FocusForward},
		{name: "Shift+Tab", direction: key.FocusBackward},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 200), theme.Light())
			var dialog base.DialogInteraction
			view := dialogTestView(&dialog, nil, nil)
			h.Frame(view)
			h.Frame(view)
			for range 3 {
				h.Router().MoveFocus(tt.direction)
				h.Frame(view)
				if !h.Router().Source().Focused(dialog.FocusTarget()) {
					t.Fatal("dialog without controls lost its neutral focus target")
				}
			}
		})
	}
}

func dialogTestView(dialog *base.DialogInteraction, close *widget.Clickable, onClose func()) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		core.RegisterOverlay(ctx, core.OverlayRequest{
			Key: "dialog", Layer: core.OverlayLayerModal,
			Layout: func(gtx layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
				if dialog.Update(ctx, gtx, base.DialogPolicy{Open: true, Interactive: interactive, Close: close}) {
					if onClose != nil {
						onClose()
					}
				}
				var last event.Tag
				if close != nil {
					last = close
				}
				dialog.LayoutStart(ctx, gtx, last)
				if close != nil {
					close.Layout(gtx, func(layout.Context) layout.Dimensions {
						return layout.Dimensions{Size: image.Pt(40, 24)}
					})
				}
				return layout.Dimensions{Size: gtx.Constraints.Max}
			},
			Tail: dialog.LayoutEnd,
		})
		dialog.ManageFocus(ctx, "dialog", true, true)
		return layout.Dimensions{}
	})
}
