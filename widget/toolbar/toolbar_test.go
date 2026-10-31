package toolbar

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
)

func TestToolbarLayoutSpacing(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	horizontal := New(probe{size: image.Pt(20, 10)}, probe{size: image.Pt(30, 12)})
	if dims := layoutToolbarFrame(ctx, new(input.Router), horizontal, time.Unix(1, 0)); dims.Size != image.Pt(58, 12) {
		t.Fatalf("horizontal Toolbar size = %v, want (58,12)", dims.Size)
	}
	vertical := horizontal.Orientation(Vertical)
	if dims := layoutToolbarFrame(ctx, new(input.Router), vertical, time.Unix(2, 0)); dims.Size != image.Pt(30, 30) {
		t.Fatalf("vertical Toolbar size = %v, want (30,30)", dims.Size)
	}
	attached := horizontal.Attached(true)
	if dims := layoutToolbarFrame(ctx, new(input.Router), attached, time.Unix(3, 0)); dims.Size != image.Pt(66, 20) {
		t.Fatalf("attached Toolbar size = %v, want (66,20)", dims.Size)
	}
	if dims := layoutToolbarFrameWithConstraints(ctx, new(input.Router), attached, time.Unix(3, 0), layout.Exact(image.Pt(300, 100))); dims.Size != image.Pt(66, 20) {
		t.Fatalf("constrained attached Toolbar size = %v, want (66,20)", dims.Size)
	}
}

func TestCommandVisibleCountReservesOverflowTrigger(t *testing.T) {
	tests := []struct {
		name      string
		widths    []int
		available int
		want      int
	}{
		{name: "all fit", widths: []int{40, 40}, available: 100, want: 2},
		{name: "reserve more", widths: []int{40, 40}, available: 80, want: 1},
		{name: "no command fits", widths: []int{40}, available: 20, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := commandVisibleCount(test.widths, test.available, 8, 32); got != test.want {
				t.Fatalf("visible commands = %d, want %d", got, test.want)
			}
		})
	}
}

func TestCommandToolbarLaysOutOverflowTriggerAtNarrowWidth(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	commands := Commands("file-tools",
		command.New("open", "Open").OnExecute(func() {}),
		command.New("save", "Save changes").OnExecute(func() {}),
		command.New("export", "Export document").OnExecute(func() {}),
	)
	layoutToolbarFrameWithConstraints(ctx, router, commands, time.Unix(1, 0), layout.Exact(image.Pt(100, 48)))
	nodes := router.AppendSemantics(nil)
	if !semanticTreeHasLabel(nodes, "More") {
		t.Fatal("narrow command toolbar did not expose its overflow trigger")
	}
}

func semanticTreeHasLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || semanticTreeHasLabel(node.Children, label) {
			return true
		}
	}
	return false
}

func TestToolbarKeyboardNavigation(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	router := new(input.Router)
	toolbar := New(
		button.New("").Key("first").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly(),
		button.New("").Key("second").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly().Disabled(true),
		button.New("").Key("last").Content(probe{size: image.Pt(8, 8)}).Size(button.Small).IconOnly(),
	)
	now := time.Unix(4, 0)
	layoutToolbarFrame(ctx, router, toolbar, now)
	first, _ := core.PeekState[widget.Clickable](ctx, "first", "clickable")
	last, _ := core.PeekState[widget.Clickable](ctx, "last", "clickable")
	if first == nil || last == nil {
		t.Fatal("New button focus targets were not retained")
	}

	router.Source().Execute(key.FocusCmd{Tag: first})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(2*time.Millisecond))
	if !router.Source().Focused(last) {
		t.Fatal("Right Arrow did not skip the disabled New button")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutToolbarFrame(ctx, router, toolbar, now.Add(3*time.Millisecond))
	if !router.Source().Focused(first) {
		t.Fatal("New focus did not loop")
	}
}

func TestToolbarVerticalNavigationWithChildFocusTargets(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	router := new(input.Router)
	first := &focusProbe{key: "first"}
	last := &focusProbe{key: "last"}
	view := New(first, last).Orientation(Vertical).LoopFocus(false)
	now := time.Unix(1, 0)
	layoutToolbarFrame(ctx, router, view, now)
	router.Source().Execute(key.FocusCmd{Tag: first})
	layoutToolbarFrame(ctx, router, view, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameEnd, State: key.Press})
	layoutToolbarFrame(ctx, router, view, now.Add(2*time.Millisecond))
	if !router.Source().Focused(last) {
		t.Fatal("End did not focus the child's non-clickable target")
	}
	router.Queue(
		key.Event{Name: key.NameDownArrow, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Press},
	)
	layoutToolbarFrame(ctx, router, view, now.Add(3*time.Millisecond))
	if !router.Source().Focused(last) || last.actions != 1 || first.actions != 0 {
		t.Fatalf("edge/child activation = focused %v actions %d/%d", router.Source().Focused(last), first.actions, last.actions)
	}
	router.Queue(key.Event{Name: key.NameHome, State: key.Press})
	layoutToolbarFrame(ctx, router, view, now.Add(4*time.Millisecond))
	if !router.Source().Focused(first) {
		t.Fatal("Home did not focus the first child")
	}
}

type focusProbe struct {
	key     string
	actions int
}

func (p *focusProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	core.RegisterFieldFocus(ctx, p.key, p, gtx.Enabled())
	core.RegisterFocusGroupItem(ctx, p, gtx.Enabled())
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: p}, key.Filter{Focus: p, Name: key.NameReturn})
		if !ok {
			break
		}
		if e, ok := e.(key.Event); ok && e.State == key.Press {
			p.actions++
		}
	}
	area := clip.Rect{Max: image.Pt(40, 20)}.Push(gtx.Ops)
	event.Op(gtx.Ops, p)
	area.Pop()
	return layout.Dimensions{Size: image.Pt(40, 20)}
}

func TestToolbarSeparatorOrientation(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()), time.Unix(5, 0)); dims.Size != image.Pt(1, 20) {
		t.Fatalf("horizontal Toolbar separator = %v, want (1,20)", dims.Size)
	}
	if dims := layoutToolbarFrame(ctx, new(input.Router), New(Separator()).Orientation(Vertical), time.Unix(6, 0)); dims.Size != image.Pt(40, 1) {
		t.Fatalf("vertical Toolbar separator = %v, want (40,1)", dims.Size)
	}
}

func TestToolbarOwnsTransitioningSeparatorIdentity(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	transitioning := func() SeparatorWidget {
		return Separator()
	}
	first := (transitioning()).Style(style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, time.Second))
	second := (transitioning()).Style(style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, time.Second))
	layoutToolbarFrame(ctx, new(input.Router), New(first, second), time.Unix(7, 0))
	if got := core.StateLen(ctx); got != 2 {
		t.Fatalf("Separator transition states = %d, want 2", got)
	}
}

type probe struct {
	size image.Point
}

func (p probe) Layout(*core.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: p.size}
}

func layoutToolbarFrame(ctx *core.Context, router *input.Router, toolbar Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(400, 240)
	return layoutToolbarFrameWithConstraints(ctx, router, toolbar, now, layout.Constraints{Max: viewport})
}

func layoutToolbarFrameWithConstraints(ctx *core.Context, router *input.Router, toolbar Widget, now time.Time, constraints layout.Constraints) layout.Dimensions {
	viewport := image.Pt(400, 240)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: constraints,
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	dims := toolbar.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
