package animation

import (
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/theme"
)

func TestLocalTransitionsInvalidateUntilDurationEnds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		easing Easing
	}{
		{name: "early endpoint", easing: func(float32) float32 { return 1 }},
		{name: "overshoot", easing: EaseBackOut},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Unix(1, 0)
			var scalar FloatTransition
			var colors ColorTransition
			scalar.Set(0, 1, start)
			colors.Set(color.NRGBA{}, color.NRGBA{R: 255, A: 255}, start)
			for _, frame := range []struct {
				elapsed time.Duration
				wake    bool
			}{
				{elapsed: 700 * time.Millisecond, wake: true},
				{elapsed: time.Second},
			} {
				for _, kind := range []string{"float", "color"} {
					var ops op.Ops
					var router input.Router
					gtx := layout.Context{Ops: &ops, Source: router.Source(), Now: start.Add(frame.elapsed)}
					if kind == "float" {
						scalar.Value(gtx, 1, time.Second, tc.easing)
					} else {
						colors.Value(gtx, color.NRGBA{R: 255, A: 255}, time.Second, tc.easing)
					}
					router.Frame(&ops)
					if _, wake := router.WakeupTime(); wake != frame.wake {
						t.Errorf("%s at %v: wake = %v, want %v", kind, frame.elapsed, wake, frame.wake)
					}
				}
			}
		})
	}
}

func TestLocalTransitionsTrackTargets(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := layout.Context{Ops: new(op.Ops), Now: start}
	var scalar FloatTransition
	if got := scalar.Value(gtx, 0, time.Second, EaseLinear); got != 0 {
		t.Fatalf("initial float = %v", got)
	}
	gtx.Now = start.Add(time.Millisecond)
	scalar.Value(gtx, 1, time.Second, EaseLinear)
	gtx.Now = start.Add(501 * time.Millisecond)
	if got := scalar.Value(gtx, 1, time.Second, EaseLinear); got != .5 {
		t.Fatalf("mid float = %v, want .5", got)
	}
	if !scalar.Ready() || scalar.Target() != 1 || !scalar.Increasing() {
		t.Fatalf("float state = ready %v, target %v, increasing %v", scalar.Ready(), scalar.Target(), scalar.Increasing())
	}

	var colors ColorTransition
	from := color.NRGBA{R: 10, A: 255}
	to := color.NRGBA{R: 110, A: 255}
	gtx.Now = start
	colors.Value(gtx, from, time.Second, EaseLinear)
	gtx.Now = start.Add(time.Millisecond)
	colors.Value(gtx, to, time.Second, EaseLinear)
	gtx.Now = start.Add(501 * time.Millisecond)
	if got := colors.Value(gtx, to, time.Second, EaseLinear); got.R != 60 {
		t.Fatalf("mid color = %#v, want red 60", got)
	}

	scalar.Reset()
	colors.Reset()
	if got := scalar.Value(gtx, 3, time.Second, EaseLinear); got != 3 {
		t.Fatalf("float after reset = %v, want new initial value 3", got)
	}
	if got := colors.Value(gtx, from, time.Second, EaseLinear); got != from {
		t.Fatalf("color after reset = %#v, want new initial value %#v", got, from)
	}
}

func TestLocalTransitionsRespectMotionTheme(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := layout.Context{Ops: new(op.Ops), Now: start}
	motion := theme.MotionTheme{Enabled: false, DurationScale: 1}
	var scalar FloatTransition
	scalar.Value(gtx, 0, time.Second, EaseLinear, motion)
	gtx.Now = start.Add(time.Millisecond)
	if got := scalar.Value(gtx, 1, time.Second, EaseLinear, motion); got != 1 {
		t.Fatalf("disabled motion float = %v, want 1", got)
	}

	var colors ColorTransition
	from := color.NRGBA{R: 10, A: 255}
	to := color.NRGBA{R: 110, A: 255}
	colors.Value(gtx, from, time.Second, EaseLinear, motion)
	if got := colors.Value(gtx, to, time.Second, EaseLinear, motion); got != to {
		t.Fatalf("disabled motion color = %#v, want %#v", got, to)
	}
}
