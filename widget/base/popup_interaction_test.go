package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestPopupInteractionEscapePolicy(t *testing.T) {
	for _, tt := range []struct {
		name     string
		policy   base.PopupPolicy
		disabled bool
		want     base.PopupDismiss
	}{
		{name: "open owner", policy: base.PopupPolicy{Open: true, Interactive: true, DismissEscape: true}, want: base.PopupDismissEscape},
		{name: "not owner", policy: base.PopupPolicy{Open: true, DismissEscape: true}},
		{name: "exit panel", policy: base.PopupPolicy{Interactive: true, DismissEscape: true}},
		{name: "keyboard disabled", policy: base.PopupPolicy{Open: true, Interactive: true}},
		{name: "disabled context", policy: base.PopupPolicy{Open: true, Interactive: true, DismissEscape: true}, disabled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 200), theme.Light())
			var interaction base.PopupInteraction
			var intents []base.PopupDismiss
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if tt.disabled {
					gtx = gtx.Disabled()
				}
				if intent := interaction.Update(ctx, gtx, tt.policy); intent != base.PopupNotDismissed {
					intents = append(intents, intent)
				}
				return layout.Dimensions{}
			})
			h.Frame(view)
			h.Key(key.NameEscape, 0)
			h.Frame(view)
			h.Frame(view)
			if tt.want == base.PopupNotDismissed {
				if len(intents) != 0 {
					t.Fatalf("unexpected close intents: %v", intents)
				}
			} else if len(intents) != 1 || intents[0] != tt.want {
				t.Fatalf("close intents = %v, want [%v]", intents, tt.want)
			}
		})
	}
}

func TestPopupInteractionOutsidePolicy(t *testing.T) {
	for _, tt := range []struct {
		name        string
		onPress     bool
		dismissable bool
		position    f32.Point
		wantPress   int
		wantRelease int
	}{
		{name: "press dismissal", onPress: true, dismissable: true, position: f32.Pt(200, 180), wantPress: 1, wantRelease: 1},
		{name: "click dismissal", dismissable: true, position: f32.Pt(200, 180), wantRelease: 1},
		{name: "outside disabled", onPress: true, position: f32.Pt(200, 180)},
		{name: "panel protected", onPress: true, dismissable: true, position: f32.Pt(40, 60)},
		{name: "arrow protected", onPress: true, dismissable: true, position: f32.Pt(40, 25)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 200), theme.Light())
			var interaction base.PopupInteraction
			requests, open := 0, true
			panel, arrow := image.Rect(20, 40, 120, 100), image.Rect(30, 20, 50, 40)
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				intent := interaction.Update(ctx, gtx, base.PopupPolicy{
					Open: open, Interactive: true, DismissOutside: tt.dismissable, DismissOnPress: tt.onPress,
				})
				if intent != base.PopupNotDismissed {
					if intent != base.PopupDismissOutside {
						t.Fatalf("close intent = %v, want outside", intent)
					}
					requests++
					open = false
				}
				interaction.LayoutOutside(gtx, gtx.Constraints.Max, false, panel, arrow)
				interaction.LayoutSurface(gtx, panel)
				interaction.LayoutArrow(gtx, arrow)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			h.Frame(view)
			h.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
				Buttons: pointer.ButtonPrimary, Position: tt.position})
			h.Frame(view)
			if requests != tt.wantPress {
				t.Fatalf("press requests = %d, want %d", requests, tt.wantPress)
			}
			h.Router().Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: tt.position})
			h.Frame(view)
			h.Frame(view)
			if requests != tt.wantRelease {
				t.Fatalf("release requests = %d, want %d", requests, tt.wantRelease)
			}
		})
	}
}
