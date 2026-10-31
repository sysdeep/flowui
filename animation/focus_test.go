package animation

import (
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/theme"
)

func TestFocusAnimationOpacityTracksVisible(t *testing.T) {
	var animation FocusAnimation
	gtx := layout.Context{Ops: new(op.Ops), Now: time.Unix(1, 0)}
	if got := animation.Opacity(gtx, true); got != 1 || animation.TargetOpacity() != 1 {
		t.Fatalf("visible opacity = %v, target = %v, want 1", got, animation.TargetOpacity())
	}

	gtx.Now = gtx.Now.Add(focusAnimationDuration + time.Millisecond)
	if got := animation.Opacity(gtx, false); got != 0 || animation.TargetOpacity() != 0 {
		t.Fatalf("hidden opacity = %v, target = %v, want 0", got, animation.TargetOpacity())
	}
}

func TestFocusAnimationRespectsMotionTheme(t *testing.T) {
	var animation FocusAnimation
	gtx := layout.Context{Ops: new(op.Ops), Now: time.Unix(1, 0)}
	motion := theme.MotionTheme{Enabled: false, DurationScale: 1}
	animation.Opacity(gtx, false, motion)
	gtx.Now = gtx.Now.Add(time.Millisecond)
	if got := animation.Opacity(gtx, true, motion); got != 1 {
		t.Fatalf("disabled motion focus opacity = %v, want 1", got)
	}
}
