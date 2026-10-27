package toggle

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/text"
)

func switchNewContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func switchTestComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func switchTestLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

type enabledProbeWidget struct {
	enabled bool
}

func (w *enabledProbeWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func TestSwitchSyncsValue(t *testing.T) {
	ctx := switchNewContext(nil)

	NewSwitch("Notifications").Key("notifications").Checked(true).Layout(ctx, switchTestLayoutContext())

	state := switchTestComponentState[switchSwitchState](ctx, "notifications", switchStateSlotSwitch)
	if state == nil {
		t.Fatal("missing switch state")
	}
	if !state.value.Value {
		t.Fatal("switch value = false, want true")
	}
}

func TestSwitchPointerToggleNotifies(t *testing.T) {
	ctx := switchNewContext(nil)
	router := new(input.Router)
	checked := false
	var notified bool
	layout := func() {
		layoutSwitchFrameWith(ctx, router, NewSwitch("Notifications").Key("notifications").Checked(checked).
			OnChange(func(next bool) {
				checked = next
				notified = next
			}))
	}

	layout()
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(10, 10),
	})
	layout()
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(10, 10),
	})
	layout()

	if !checked || !notified {
		t.Fatalf("switch did not notify pointer toggle, checked=%v notified=%v", checked, notified)
	}
}

func TestSwitchControlOnlyLayout(t *testing.T) {
	dims := NewSwitch("").Key("notifications").Checked(false).Layout(switchNewContext(nil), switchTestLayoutContext())

	if dims.Size != image.Pt(44, 24) {
		t.Fatalf("switch size = %v, want (44,24)", dims.Size)
	}
}

func TestSwitchTrackPartControlsGeometry(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	activeTheme := *core.ActiveTheme(ctx)
	gtx := switchTestLayoutContext()
	dims := (NewSwitch("").Key("compact").Checked(false)).Style(style.Style{}.
		Part(style.PartTrack, style.Style{}.Width(30).Height(14))).
		Layout(ctx, gtx)
	focusSpace := max(gtx.Dp(activeTheme.Components.Switch.FocusSpace), 1)
	want := image.Pt(gtx.Dp(30)+focusSpace*2, gtx.Dp(14)+focusSpace*2)
	if dims.Size != want {
		t.Fatalf("styled switch size = %v, want %v", dims.Size, want)
	}
}

func TestSwitchRespectsConstraints(t *testing.T) {
	gtx := switchTestLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(12, 12)}

	dims := NewSwitch("").Key("notifications").Checked(false).Layout(switchNewContext(nil), gtx)

	if dims.Size != image.Pt(12, 12) {
		t.Fatalf("switch size = %v, want (12,12)", dims.Size)
	}
}

func TestSwitchThumbReceivesCheckedState(t *testing.T) {
	got := false
	NewSwitch("").Key("power").Checked(true).
		Thumb(func(checked bool) core.Widget {
			got = checked
			return text.New("1")
		}).
		Layout(switchNewContext(nil), switchTestLayoutContext())

	if !got {
		t.Fatal("thumb content did not receive checked=true")
	}
}

func TestSwitchPassesDisabledContextToThumb(t *testing.T) {
	probe := &enabledProbeWidget{}

	NewSwitch("").Key("power").Checked(false).
		Disabled(true).
		Thumb(func(bool) core.Widget {
			return probe
		}).
		Layout(switchNewContext(nil), switchTestLayoutContext())

	if probe.enabled {
		t.Fatal("switch thumb content was laid out with enabled context")
	}
}

func TestSwitchLabelPartUsesCommonBoxRenderer(t *testing.T) {
	base := NewSwitch("Notifications").Key("notifications").Checked(false).
		Layout(switchNewContext(nil), switchTestLayoutContext())
	styled := (NewSwitch("Notifications").Key("notifications").Checked(false)).Style(style.Style{}.Part(style.PartLabel, style.Style{}.PaddingY(7))).
		Layout(switchNewContext(nil), switchTestLayoutContext())
	if styled.Size.Y <= base.Size.Y {
		t.Fatalf("styled label height = %d, want greater than %d", styled.Size.Y, base.Size.Y)
	}
}

func TestSwitchSelectionAnimation(t *testing.T) {
	state := new(switchSwitchState)
	start := time.Unix(1, 0)
	gtx := switchTestLayoutContext()
	gtx.Now = start

	selectionDuration := 300 * time.Millisecond
	if got := state.selection(gtx, false, selectionDuration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.selection(gtx, true, selectionDuration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(selectionDuration / 2)
	mid := state.selection(gtx, true, selectionDuration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(selectionDuration)
	if got := state.selection(gtx, true, selectionDuration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestSwitchFocusOpacityAnimation(t *testing.T) {
	state := new(switchSwitchState)
	start := time.Unix(1, 0)
	gtx := switchTestLayoutContext()
	gtx.Now = start

	focusDuration := 100 * time.Millisecond
	if got := state.focusOpacity(gtx, false, focusDuration); got != 0 {
		t.Fatalf("initial focus = %v, want 0", got)
	}
	if got := state.focusOpacity(gtx, true, focusDuration); got != 0 {
		t.Fatalf("focus start = %v, want 0", got)
	}

	gtx.Now = start.Add(focusDuration / 2)
	mid := state.focusOpacity(gtx, true, focusDuration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("focus midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(focusDuration)
	if got := state.focusOpacity(gtx, true, focusDuration); got != 1 {
		t.Fatalf("focus end = %v, want 1", got)
	}
}

func TestSwitchFocusesOnPress(t *testing.T) {
	ctx := switchNewContext(nil)
	router := new(input.Router)
	layoutSwitchFrame(ctx, router)
	state := switchTestComponentState[switchSwitchState](ctx, "notifications", switchStateSlotSwitch)
	if state == nil {
		t.Fatal("missing switch state")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(10, 10),
	})
	layoutSwitchFrame(ctx, router)

	if !router.Source().Focused(&state.value) {
		t.Fatal("switch did not gain focus after press")
	}
}

func TestSwitchGroupHorizontalWraps(t *testing.T) {
	gtx := switchTestLayoutContext()
	gtx.Constraints.Max = image.Pt(100, 200)

	dims := NewSwitchGroup(
		fixedSwitchChild{size: image.Pt(70, 10)},
		fixedSwitchChild{size: image.Pt(70, 10)},
	).Horizontal().Layout(switchNewContext(nil), gtx)

	if dims.Size != image.Pt(70, 36) {
		t.Fatalf("switch group size = %v, want (70,36)", dims.Size)
	}
}

type fixedSwitchChild struct {
	size image.Point
}

func (f fixedSwitchChild) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(f.size)}
}

func layoutSwitchFrame(ctx *core.Context, router *input.Router) {
	layoutSwitchFrameWith(ctx, router, NewSwitch("Notifications").Key("notifications").Checked(false))
}

func layoutSwitchFrameWith(ctx *core.Context, router *input.Router, sw SwitchWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	sw.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
