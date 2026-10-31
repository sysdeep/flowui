package animation

import (
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func TestTweenSnapsOnFirstFrameByDefault(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("position", 1), time.Unix(1, 0))
	if value != 1 || wake {
		t.Fatalf("default first frame = %v, wake %v", value, wake)
	}
}

func TestTweenAnimatesExplicitInitialValue(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(2, 0)
	tween := Tween("position", 1).Initial(0).Duration(time.Second).Easing(EaseCubicInOut)
	if value, wake := tweenFrame(ctx, router, tween, start); value != 0 || !wake {
		t.Fatalf("entry start = %v, wake %v", value, wake)
	}
	if value, wake := tweenFrame(ctx, router, tween, start.Add(500*time.Millisecond)); value != 0.5 || !wake {
		t.Fatalf("entry midpoint = %v, wake %v", value, wake)
	}
	if value, wake := tweenFrame(ctx, router, tween, start.Add(time.Second)); value != 1 || wake {
		t.Fatalf("entry end = %v, wake %v", value, wake)
	}
}

func TestTweenTargetChangesRemainContinuous(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(3, 0)
	forward := Tween("position", 1).Initial(0).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, forward, start)
	if value, _ := tweenFrame(ctx, router, forward, start.Add(500*time.Millisecond)); value != 0.5 {
		t.Fatalf("forward midpoint = %v", value)
	}
	reverse := Tween("position", 0).Duration(time.Second).Easing(EaseLinear)
	if value, _ := tweenFrame(ctx, router, reverse, start.Add(500*time.Millisecond)); value != 0.5 {
		t.Fatalf("reversal jumped to %v", value)
	}
	if value, _ := tweenFrame(ctx, router, reverse, start.Add(750*time.Millisecond)); value != 0.375 {
		t.Fatalf("reverse midpoint = %v, want 0.375", value)
	}
}

func TestTweenHonorsGlobalMotionControls(t *testing.T) {
	activeTheme := theme.Light()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("disabled", 1).Initial(0), time.Unix(4, 0))
	if value != 1 || wake {
		t.Fatalf("disabled motion = %v, wake %v", value, wake)
	}

	activeTheme = theme.Light()
	activeTheme.Motion.DurationScale = 0.5
	ctx = tweenTestContext(&activeTheme)
	router = new(input.Router)
	start := time.Unix(5, 0)
	tween := Tween("scaled", 1).Initial(0).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, tween, start)
	if value, wake := tweenFrame(ctx, router, tween, start.Add(500*time.Millisecond)); value != 1 || wake {
		t.Fatalf("scaled duration end = %v, wake %v", value, wake)
	}
}

func TestTweenInvalidatesUntilTimeCompletes(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(6, 0)
	tween := Tween("constant-output", 1).
		Initial(0).
		Duration(time.Second).
		Easing(func(float32) float32 { return 1 })
	tweenFrame(ctx, router, tween, start)
	if value, wake := tweenFrame(ctx, router, tween, start.Add(250*time.Millisecond)); value != 1 || !wake {
		t.Fatalf("early target value = %v, wake %v", value, wake)
	}
}

func TestTweenRevisionRestartsFromInitialValue(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(9, 0)
	first := Tween("revision", 1).Initial(0).Revision(1).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, first, start)
	if value, _ := tweenFrame(ctx, router, first, start.Add(time.Second)); value != 1 {
		t.Fatalf("first revision end = %v", value)
	}
	second := Tween("revision", 1).Initial(0).Revision(2).Duration(time.Second).Easing(EaseLinear)
	if value, wake := tweenFrame(ctx, router, second, start.Add(2*time.Second)); value != 0 || !wake {
		t.Fatalf("second revision start = %v, wake %v", value, wake)
	}
}

func TestTweenRejectsInvalidConfigurationAndEasing(t *testing.T) {
	tests := []func(){
		func() { Tween("bad", float32(math.NaN())) },
		func() { Tween("bad", 1).Initial(float32(math.Inf(1))) },
		func() { Tween("bad", 1).Duration(-time.Second) },
		func() { Tween("bad", 1).Easing(nil) },
	}
	for _, run := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid tween configuration did not panic")
				}
			}()
			run()
		}()
	}

	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(7, 0)
	tween := Tween("bad-easing", 1).Initial(0).Duration(time.Second).Easing(func(float32) float32 {
		return float32(math.NaN())
	})
	tweenFrame(ctx, router, tween, start)
	defer func() {
		if recover() == nil {
			t.Fatal("non-finite easing output did not panic")
		}
	}()
	tweenFrame(ctx, router, tween, start.Add(time.Millisecond))
}

func TestTweenStateIsReleasedWhenUnused(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	tweenFrame(ctx, router, Tween("temporary", 1), time.Unix(8, 0))
	if core.StateLen(ctx) != 1 {
		t.Fatalf("Tween retained states = %d, want 1", core.StateLen(ctx))
	}
	core.BeginFrame(ctx)
	core.EndFrame(ctx)
	if core.StateLen(ctx) != 0 {
		t.Fatalf("unused Tween states = %d, want 0", core.StateLen(ctx))
	}
}

func tweenTestContext(activeTheme *theme.Theme) *core.Context {
	return core.New(nil, activeTheme, locale.LanguageEnglish)
}

func neutralTestTheme() *theme.Theme {
	value := theme.Light()
	return &value
}

func tweenFrame(ctx *core.Context, router *input.Router, tween TweenValue, now time.Time) (float32, bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	value := tween.Value(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	_, wake := router.WakeupTime()
	return value, wake
}

func TestTweenSpringApproachesTarget(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(20, 0)
	tween := Tween("spring", 100).Initial(0).Spring(SpringSnappy())

	value, wake := tweenFrame(ctx, router, tween, start)
	if value != 0 || !wake {
		t.Fatalf("spring start = %v, wake %v", value, wake)
	}

	// Step several frames; value should move toward 100 without jumping instantly.
	var last float32
	for i := 1; i <= 30; i++ {
		value, _ = tweenFrame(ctx, router, tween, start.Add(time.Duration(i)*16*time.Millisecond))
		if value < last {
			// allow tiny numerical noise
			if last-value > 1 {
				t.Fatalf("spring went backwards: %v -> %v", last, value)
			}
		}
		last = value
	}
	if last <= 10 {
		t.Fatalf("spring did not progress enough: %v", last)
	}

	// Eventually settles at target.
	for i := 31; i <= 200; i++ {
		value, wake = tweenFrame(ctx, router, tween, start.Add(time.Duration(i)*16*time.Millisecond))
		if !wake {
			if value != 100 {
				t.Fatalf("settled value = %v, want 100", value)
			}
			return
		}
	}
	t.Fatalf("spring did not settle; last=%v", value)
}

func TestTweenSpringHonorsMotionDisabled(t *testing.T) {
	activeTheme := theme.Light()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("spring-off", 1).Initial(0).Spring(DefaultSpring()), time.Unix(21, 0))
	if value != 1 || wake {
		t.Fatalf("spring with motion off = %v, wake %v", value, wake)
	}
}

func TestTweenModeSwitchFromSpringToDuration(t *testing.T) {
	ctx := tweenTestContext(neutralTestTheme())
	router := new(input.Router)
	start := time.Unix(22, 0)
	spring := Tween("mode", 100).Initial(0).Spring(SpringSnappy())
	// Drive spring until settled at target.
	var value float32
	var wake bool
	for i := range 200 {
		value, wake = tweenFrame(ctx, router, spring, start.Add(time.Duration(i)*16*time.Millisecond))
		if !wake {
			break
		}
	}
	if value != 100 || wake {
		t.Fatalf("spring settle = %v, wake %v", value, wake)
	}

	// Same target, switch to duration mode — must remain at target, not re-run spring.
	duration := Tween("mode", 100).Duration(time.Second).Easing(EaseLinear)
	value, wake = tweenFrame(ctx, router, duration, start.Add(5*time.Second))
	if value != 100 || wake {
		t.Fatalf("duration after mode switch = %v, wake %v", value, wake)
	}
}

func TestTweenSpringRetargetsFromCurrentTime(t *testing.T) {
	start := time.Unix(23, 0)
	config := SpringSnappy()

	referenceCtx := tweenTestContext(neutralTestTheme())
	referenceRouter := new(input.Router)
	forward := Tween("forward", 1).Initial(0).Spring(config)
	tweenFrame(referenceCtx, referenceRouter, forward, start)
	for step := 1; step <= 3; step++ {
		tweenFrame(referenceCtx, referenceRouter, forward, start.Add(time.Duration(step)*100*time.Millisecond))
	}
	want, _ := tweenFrame(referenceCtx, referenceRouter, forward, start.Add(400*time.Millisecond))

	retargetCtx := tweenTestContext(neutralTestTheme())
	retargetRouter := new(input.Router)
	retarget := Tween("retarget", 1).Initial(0).Spring(config)
	tweenFrame(retargetCtx, retargetRouter, retarget, start)
	for step := 1; step <= 3; step++ {
		tweenFrame(retargetCtx, retargetRouter, retarget, start.Add(time.Duration(step)*100*time.Millisecond))
	}
	got, _ := tweenFrame(retargetCtx, retargetRouter, Tween("retarget", 0).Spring(config), start.Add(400*time.Millisecond))
	if got != want {
		t.Fatalf("retarget value = %v, want current-time value %v", got, want)
	}
}
