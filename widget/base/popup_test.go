package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

func popupTrigger(base.PopupState) ui.Widget {
	return core.WidgetFunc(func(*core.Context, giolayout.Context) giolayout.Dimensions {
		return giolayout.Dimensions{Size: image.Pt(120, 32)}
	})
}

func TestPopupAnchorsPanelBelowTrigger(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	view := base.Popup("menu", true).
		Placement(core.PopoverBottomStart).
		Gap(8).
		Trigger(popupTrigger).
		Panel(func(base.PopupState) ui.Widget {
			return text.New("Panel")
		})
	h.Frame(view)
	h.Frame(view)
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Panel")
	if !ok {
		t.Fatalf("panel semantics missing: %v", uitest.SemanticLabels(h.Semantics()))
	}
	if node.Desc.Bounds.Min.Y != 40 {
		t.Fatalf("panel y = %d, want 40", node.Desc.Bounds.Min.Y)
	}
}

func TestPopupClosedDoesNotLayoutPanel(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	view := base.Popup("menu", false).
		Trigger(popupTrigger).
		Panel(func(base.PopupState) ui.Widget {
			return text.New("Panel")
		})
	h.Frame(view)
	if _, ok := uitest.SemanticByLabel(h.Semantics(), "Panel"); ok {
		t.Fatal("closed popup laid out a panel")
	}
}

func TestPopupDismissesOnOutsideClick(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	open, dismissed := true, false
	build := func() ui.Widget {
		return base.Popup("menu", open).
			OnDismiss(func() { dismissed = true; open = false }).
			Trigger(popupTrigger).
			Panel(func(base.PopupState) ui.Widget { return text.New("Panel") })
	}
	h.Frame(build())
	h.Frame(build())
	h.Click(f32.Pt(200, 100))
	h.Frame(build())
	if !dismissed || open {
		t.Fatalf("outside click: dismissed=%v open=%v", dismissed, open)
	}
}

func TestPopupDismissesOnOutsidePress(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	requests := 0
	view := base.Popup("menu", true).
		OnDismiss(func() { requests++ }).Trigger(popupTrigger).
		Panel(func(base.PopupState) ui.Widget { return text.New("Panel") })
	h.Frame(view)
	h.Frame(view)
	h.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(200, 180)})
	h.Frame(view)
	if requests != 1 {
		t.Fatalf("outside press requests = %d, want 1", requests)
	}
}

func TestPopupEscapeOnlyDismissesNestedOwner(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	parentRequests, childRequests := 0, 0
	view := base.Popup("parent", true).Trigger(popupTrigger).
		OnDismiss(func() { parentRequests++ }).
		Panel(func(base.PopupState) ui.Widget {
			return base.Popup("child", true).Trigger(popupTrigger).
				OnDismiss(func() { childRequests++ }).
				Panel(func(base.PopupState) ui.Widget { return text.New("Child") })
		})
	h.Frame(view)
	h.Frame(view)
	h.Key(key.NameEscape, 0)
	h.Frame(view)
	if parentRequests != 0 || childRequests != 1 {
		t.Fatalf("dismiss requests: parent=%d child=%d", parentRequests, childRequests)
	}
}

func TestPopupSurfaceBlocksBackgroundWithoutDismissing(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	requests, backgroundClicks := 0, 0
	view := core.WidgetFunc(func(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
		base.Button("background").OnClick(func() { backgroundClicks++ }).Content(buttonContent).Layout(ctx, gtx)
		return base.Popup("menu", true).Trigger(popupTrigger).
			OnDismiss(func() { requests++ }).
			Panel(func(base.PopupState) ui.Widget {
				return core.WidgetFunc(func(*core.Context, giolayout.Context) giolayout.Dimensions {
					return giolayout.Dimensions{Size: image.Pt(120, 40)}
				})
			}).Layout(ctx, gtx)
	})
	h.Frame(view)
	h.Frame(view)
	h.Click(f32.Pt(60, 35))
	h.Frame(view)
	h.Frame(view)
	if requests != 0 || backgroundClicks != 0 {
		t.Fatalf("panel click: dismiss=%d background=%d", requests, backgroundClicks)
	}
}

func TestPopupBlockersDoNotEnterFocusOrder(t *testing.T) {
	h := uitest.New(image.Pt(240, 200), theme.Light())
	activations := 0
	view := base.Popup("menu", true).Trigger(popupTrigger).OnDismiss(func() {}).
		Panel(func(base.PopupState) ui.Widget {
			return base.Button("action").Content(buttonContent).OnClick(func() { activations++ })
		})
	h.Frame(view)
	h.Frame(view)
	h.Router().MoveFocus(key.FocusForward)
	h.Key(key.NameSpace, 0)
	h.Frame(view)
	if activations != 1 {
		t.Fatalf("first Tab target was not the panel button: activations=%d", activations)
	}
}
