package base_test

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestPopupTriggerLongPressLifecycle(t *testing.T) {
	for _, tt := range []struct {
		name    string
		send    bool
		kind    pointer.Kind
		move    f32.Point
		disable bool
		want    int
	}{
		{name: "hold fires once", want: 1},
		{name: "release cancels", send: true, kind: pointer.Release},
		{name: "cancel cancels", send: true, kind: pointer.Cancel},
		{name: "movement cancels", send: true, kind: pointer.Move, move: f32.Pt(40, 20)},
		{name: "small movement retains hold", send: true, kind: pointer.Move, move: f32.Pt(25, 20), want: 1},
		{name: "disabled cancels", disable: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := uitest.New(image.Pt(160, 100), theme.Light())
			var trigger base.PopupTrigger
			var gestures []base.PopupTriggerEvent
			disabled := false
			view := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
				if disabled {
					gtx = gtx.Disabled()
				}
				trigger.Update(gtx, base.PopupTriggerPolicy{LongPress: true, TouchOnly: true, Delay: time.Second}, func(e base.PopupTriggerEvent) {
					gestures = append(gestures, e)
				})
				trigger.Layout(gtx, image.Pt(160, 100))
				return layout.Dimensions{Size: image.Pt(160, 100)}
			})
			h.Frame(view)
			h.TouchDown(f32.Pt(20, 20))
			h.Frame(view)
			if tt.send {
				position := tt.move
				if position == (f32.Point{}) {
					position = f32.Pt(20, 20)
				}
				h.Touch(2, tt.kind, position)
			}
			disabled = tt.disable
			h.Frame(view)
			h.Advance(time.Second)
			h.Frame(view)
			h.Advance(time.Second)
			h.Frame(view)
			if len(gestures) != tt.want || trigger.Holding() {
				t.Fatalf("gestures = %v, holding = %v; want %d gestures", gestures, trigger.Holding(), tt.want)
			}
			if len(gestures) > 0 && (!gestures[0].LongPress || gestures[0].Position != f32.Pt(20, 20)) {
				t.Fatalf("long press lost its initial position: %v", gestures)
			}
		})
	}
}

func TestPopupTriggerPreservesPrimaryClick(t *testing.T) {
	h := uitest.New(image.Pt(160, 100), theme.Light())
	var trigger base.PopupTrigger
	var click widget.Clickable
	clicks, gestures := 0, 0
	view := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		trigger.Update(gtx, base.PopupTriggerPolicy{Secondary: true}, func(base.PopupTriggerEvent) { gestures++ })
		for click.Clicked(gtx) {
			clicks++
		}
		dims := click.Layout(gtx, func(layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(160, 100)}
		})
		trigger.Layout(gtx, dims.Size)
		return dims
	})
	h.Frame(view)
	h.Click(f32.Pt(20, 20))
	h.Frame(view)
	if clicks != 1 || gestures != 0 {
		t.Fatalf("primary click: clicks = %d, gestures = %d", clicks, gestures)
	}
	h.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 3,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(30, 40)})
	h.Frame(view)
	if clicks != 1 || gestures != 1 {
		t.Fatalf("secondary click: clicks = %d, gestures = %d", clicks, gestures)
	}
}
