package base

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestTooltipShowAndCloseDelays(t *testing.T) {
	start := time.Unix(1, 0)
	state := new(TooltipInteraction)
	state.UpdateActive(tooltipContextAt(start), true, TooltipPolicy{Disabled: false, Delay: 1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond})
	if state.Open || state.showAt != start.Add(1500*time.Millisecond) {
		t.Fatalf("initial show state = open %v at %v", state.Open, state.showAt)
	}

	state.UpdateActive(tooltipContextAt(start.Add(1499*time.Millisecond)), true, TooltipPolicy{Disabled: false, Delay: 1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond})
	if state.Open {
		t.Fatal("tooltip opened before the show delay")
	}
	state.UpdateActive(tooltipContextAt(start.Add(1500*time.Millisecond)), true, TooltipPolicy{Disabled: false, Delay: 1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond})
	if !state.Open {
		t.Fatal("tooltip did not open when the show delay elapsed")
	}

	leave := start.Add(2 * time.Second)
	state.UpdateActive(tooltipContextAt(leave), false, TooltipPolicy{Disabled: false, Delay: 1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond})
	if !state.Open || state.hideAt != leave.Add(500*time.Millisecond) {
		t.Fatalf("close state = open %v at %v", state.Open, state.hideAt)
	}
	state.UpdateActive(tooltipContextAt(leave.Add(500*time.Millisecond)), false, TooltipPolicy{Disabled: false, Delay: 1500 * time.Millisecond, CloseDelay: 500 * time.Millisecond})
	if state.Open {
		t.Fatal("tooltip stayed open after the close delay")
	}
}

func TestTooltipLeaveCancelsPendingShow(t *testing.T) {
	start := time.Unix(1, 0)
	state := new(TooltipInteraction)
	state.UpdateActive(tooltipContextAt(start), true, TooltipPolicy{Disabled: false, Delay: time.Second, CloseDelay: time.Second})
	state.UpdateActive(tooltipContextAt(start.Add(100*time.Millisecond)), false, TooltipPolicy{Disabled: false, Delay: time.Second, CloseDelay: time.Second})
	if !state.showAt.IsZero() || state.Open {
		t.Fatalf("pending show was not cancelled: open %v at %v", state.Open, state.showAt)
	}
}

func TestTooltipReenterCancelsPendingClose(t *testing.T) {
	start := time.Unix(1, 0)
	state := &TooltipInteraction{Active: true, Open: true}
	state.UpdateActive(tooltipContextAt(start), false, TooltipPolicy{Disabled: false, Delay: 0, CloseDelay: time.Second})
	state.UpdateActive(tooltipContextAt(start.Add(100*time.Millisecond)), true, TooltipPolicy{Disabled: false, Delay: 0, CloseDelay: time.Second})
	if !state.hideAt.IsZero() || !state.Open {
		t.Fatalf("re-enter state = open %v hideAt %v", state.Open, state.hideAt)
	}
}

func TestTooltipPeerCloseSkipsCloseDelay(t *testing.T) {
	state := &TooltipInteraction{
		Active: true,
		Open:   true,
		hideAt: time.Unix(2, 0),
	}
	state.Close()
	if state.Open || !state.showAt.IsZero() || !state.hideAt.IsZero() {
		t.Fatalf("peer close did not close immediately: %+v", state)
	}
}

func TestTooltipDisabledClosesImmediately(t *testing.T) {
	state := &TooltipInteraction{Hovered: true, Active: true, Open: true}
	state.UpdateActive(tooltipContextAt(time.Unix(1, 0)), true, TooltipPolicy{Disabled: true, Delay: time.Second, CloseDelay: time.Second})
	if state.Open || state.Active || state.Hovered || !state.showAt.IsZero() || !state.hideAt.IsZero() {
		t.Fatalf("disabled state was not cleared: %+v", state)
	}
}

func tooltipContextAt(now time.Time) layout.Context {
	return layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source: new(input.Router).Source(), Ops: new(op.Ops), Now: now}
}
