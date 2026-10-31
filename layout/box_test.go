package layout

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func TestBoxAppliesWidthAndPadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Style(style.Style{}.Width(100).Padding(10)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size.X != 100 {
		t.Fatalf("width = %d, want 100", dims.Size.X)
	}
	if child.constraints.Min.X != 80 || child.constraints.Max.X != 80 {
		t.Fatalf("child width constraints = %v, want exact 80", child.constraints)
	}
}

func TestBoxOpacityCompositesOverlappingContent(t *testing.T) {
	window, err := headless.NewWindow(16, 16)
	if err != nil {
		t.Fatal(err)
	}
	defer window.Release()
	child := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{R: 255, A: 255}, clip.Rect(image.Rect(0, 0, 16, 16)).Op())
		paint.FillShape(gtx.Ops, color.NRGBA{B: 255, A: 255}, clip.Rect(image.Rect(0, 0, 8, 16)).Op())
		return layout.Dimensions{Size: image.Pt(16, 16)}
	})
	for _, tc := range []struct {
		name    string
		opacity float32
		alpha   uint8
	}{
		{name: "opaque", opacity: 1, alpha: 255},
		{name: "translucent", opacity: .5, alpha: 128},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var operations op.Ops
			Box(child).Style(style.New().Width(16).Height(16).Opacity(tc.opacity)).Layout(newContext(nil), layout.Context{
				Constraints: layout.Exact(image.Pt(16, 16)),
				Ops:         &operations,
			})
			if err := window.Frame(&operations); err != nil {
				t.Fatal(err)
			}
			pixels := image.NewRGBA(image.Rect(0, 0, 16, 16))
			if err := window.Screenshot(pixels); err != nil {
				t.Fatal(err)
			}
			covered, uncovered := pixels.RGBAAt(4, 8), pixels.RGBAAt(12, 8)
			if covered.A != tc.alpha || uncovered.A != tc.alpha {
				t.Fatalf("group alpha = %d/%d, want %d", covered.A, uncovered.A, tc.alpha)
			}
			if covered.R != 0 || covered.B == 0 || uncovered.B != 0 || uncovered.R == 0 {
				t.Fatalf("overlapping content colors = %v/%v", covered, uncovered)
			}
		})
	}
}

func TestBoxAllowsExplicitZeroSize(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(style.Style{}.Width(0).Height(0)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != (image.Point{}) || child.constraints != layout.Exact(image.Point{}) {
		t.Fatalf("zero-sized Box = %v/%v, want zero size and constraints", dims.Size, child.constraints)
	}
}

func TestInteractiveBoxHandlesClickAndDisabledState(t *testing.T) {
	for _, test := range []struct {
		key      string
		disabled bool
		want     int
	}{
		{key: "enabled", want: 1},
		{key: "disabled", disabled: true},
	} {
		ctx := newContext(nil)
		clickable := new(widget.Clickable)
		clickable.Click()
		core.UseStateWith(ctx, test.key, "clickable", func() *widget.Clickable { return clickable })
		clicks := 0
		dims := Box(nil).
			Key(test.key).
			Disabled(test.disabled).
			Style(style.Style{}.
				Width(80).
				Height(30).
				On(style.Disabled, style.Style{}.Width(90)),
			).
			OnClick(func() { clicks++ }).
			Layout(ctx, testLayoutContext())

		wantWidth := 80
		if test.disabled {
			wantWidth = 90
		}
		if clicks != test.want || dims.Size != image.Pt(wantWidth, 30) {
			t.Fatalf("%s clicks/size = %d/%v, want %d/(%d,30)", test.key, clicks, dims.Size, test.want, wantWidth)
		}
	}
}

func TestInteractiveBoxSemantics(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 40)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	Box(nil).
		Key("save").
		Label("Save changes").
		Disabled(true).
		Style(style.Style{}.Width(80).Height(30)).
		OnClick(nil).
		Layout(newContext(nil), gtx)
	router.Frame(&ops)
	node, ok := semanticButton(router.AppendSemantics(nil))
	if !ok || node.Desc.Label != "Save changes" || !node.Desc.Disabled {
		t.Fatalf("interactive Box semantics = %#v", node.Desc)
	}
}

func semanticButton(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			return node, true
		}
		if child, ok := semanticButton(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func TestBoxStyleCascadesScopeAndInstance(t *testing.T) {
	ctx := newContext(nil)
	restore := core.PushStyle(ctx, style.Style{}.
		Padding(0, 12).
		TextColor(style.RGB(0x112233)),
	)
	defer restore()
	probe := new(styledBoxProbe)
	instanceForeground := style.RGB(0x445566)
	instanceBackground := style.RGB(0x778899)

	// Scope supplies horizontal padding; instance owns size, vertical padding, and colors.
	dims := Box(probe).
		Style(style.Style{}.
			Width(120).
			Padding(4, 12, 4, 20). // overrides scope padding on the left only
			TextColor(instanceForeground).
			Background(instanceBackground),
		).
		Layout(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 100)}, Ops: new(op.Ops)})

	// Padding: L=20 (instance), R=12 (scope), T/B=4 (instance) → child width 120-32=88.
	wantConstraints := layout.Constraints{Min: image.Pt(88, 0), Max: image.Pt(88, 92)}
	if dims.Size.X != 120 || probe.constraints != wantConstraints {
		t.Fatalf("box size/child constraints = %v/%v", dims.Size, probe.constraints)
	}
	if probe.foreground != instanceForeground.Color || probe.background != instanceBackground.Color {
		t.Fatalf("box colors = %#v/%#v", probe.foreground, probe.background)
	}
}

func TestBoxStyleJoinsMultipleStyleCalls(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).
		Style(style.Style{}.Width(100).Padding(4)).
		Style(style.Style{}.Width(120).Padding(4, 4, 4, 20)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 300)},
			Ops:         new(op.Ops),
		})
	if dims.Size.X != 120 {
		t.Fatalf("joined width = %d, want 120", dims.Size.X)
	}
	// Padding all 4 then left 20 → child max width 120-24=96.
	if child.constraints.Max.X != 96 {
		t.Fatalf("joined padding child max X = %d, want 96", child.constraints.Max.X)
	}
}

func TestBoxReplaceStyleDropsEarlierLayer(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).
		Style(style.Style{}.Width(100).Padding(4)).
		ReplaceStyle(style.Style{}.Width(120)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 300)},
			Ops:         new(op.Ops),
		})
	if dims.Size.X != 120 {
		t.Fatalf("replaced width = %d, want 120", dims.Size.X)
	}
	if child.constraints.Max.X != 120 {
		t.Fatalf("replaced padding child max X = %d, want 120", child.constraints.Max.X)
	}
}

func TestBoxConditionalStyleTransition(t *testing.T) {
	ctx := newContext(nil)
	start := time.Unix(1, 0)
	from := style.RGB(0x102030).Color
	to := style.RGB(0x8090a0).Color
	layoutAt := func(now time.Time, active bool) color.NRGBA {
		probe := new(styledBoxProbe)
		statusStyle := style.Style{}.Background(style.SolidColor{Color: from}).
			Transition(style.PropBackgroundColor, 100*time.Millisecond)
		if active {
			statusStyle = statusStyle.Background(style.SolidColor{Color: to})
		}
		core.BeginFrame(ctx)
		Box(probe).
			Key("status").
			Style(statusStyle).
			Layout(ctx, layout.Context{
				Constraints: layout.Exact(image.Pt(20, 20)),
				Ops:         new(op.Ops),
				Now:         now,
			})
		core.EndFrame(ctx)
		return probe.background
	}

	if got := layoutAt(start, false); got != from {
		t.Fatalf("initial background = %#v, want %#v", got, from)
	}
	if got := layoutAt(start, true); got != from {
		t.Fatalf("transition start = %#v, want %#v", got, from)
	}
	middle := layoutAt(start.Add(50*time.Millisecond), true)
	if middle == from || middle == to {
		t.Fatalf("transition midpoint = %#v", middle)
	}
	if got := layoutAt(start.Add(100*time.Millisecond), true); got != to {
		t.Fatalf("transition end = %#v, want %#v", got, to)
	}
}

func TestBoxStyleTransformMovesOverlayPlacement(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(100, 80)}, Box(&overlayProbeWidget{
		key:    "styled-box",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}).Style(style.Style{}.Translate(10, 5)))
	if want := image.Rect(10, 5, 30, 15); got != want {
		t.Fatalf("transformed anchor = %v, want %v", got, want)
	}
}

func TestBoxStyleTranslateUsesDp(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 80)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2},
		Ops:         new(op.Ops),
	}
	var got image.Rectangle
	probe := &overlayProbeWidget{
		key:    "scaled-translate",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
		got:    &got,
	}

	core.BeginFrameWithViewport(ctx, viewport)
	Box(probe).Style(style.Style{}.Translate(10, 5)).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	if want := image.Rect(20, 10, 40, 20); got != want {
		t.Fatalf("translated anchor = %v, want %v", got, want)
	}
}

func TestBoxClampsExplicitSizeToParentConstraints(t *testing.T) {
	for _, test := range []struct {
		name        string
		requested   image.Point
		constraints layout.Constraints
		want        image.Point
	}{
		{name: "maximum", requested: image.Pt(120, 120), constraints: layout.Constraints{Min: image.Pt(40, 50), Max: image.Pt(80, 90)}, want: image.Pt(80, 90)},
		{name: "minimum", requested: image.Pt(20, 20), constraints: layout.Constraints{Min: image.Pt(40, 50), Max: image.Pt(80, 90)}, want: image.Pt(40, 50)},
	} {
		t.Run(test.name, func(t *testing.T) {
			child := &constraintWidget{}
			Box(child).Style(style.Style{}.Width(unit.Dp(test.requested.X)).Height(unit.Dp(test.requested.Y))).Layout(newContext(nil), layout.Context{
				Constraints: test.constraints,
				Ops:         new(op.Ops),
			})
			if child.constraints != layout.Exact(test.want) {
				t.Fatalf("child constraints = %v, want exact %v", child.constraints, test.want)
			}
		})
	}
}

func TestBoxAppliesMinMaxConstraints(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(style.Style{}.MinWidth(80).MaxWidth(120).MinHeight(40).MaxHeight(90)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 80 || child.constraints.Max.X != 120 {
		t.Fatalf("width constraints = %v, want min 80 max 120", child.constraints)
	}
	if child.constraints.Min.Y != 40 || child.constraints.Max.Y != 90 {
		t.Fatalf("height constraints = %v, want min 40 max 90", child.constraints)
	}
}

func TestBoxStyleFill(t *testing.T) {
	child := &constraintWidget{}
	Box(child).Style(style.Style{}.FillWidth().FillHeight()).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
	})

	if child.constraints != layout.Exact(image.Pt(300, 200)) {
		t.Fatalf("constraints = %v, want exact (300,200)", child.constraints)
	}
}

func TestBoxStyleAspectRatioResolvesFromWidth(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(style.Style{}.Width(80).AspectRatio(2)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(80, 40) || child.constraints != layout.Exact(image.Pt(80, 40)) {
		t.Fatalf("aspect box/constraints = %v/%v, want (80,40)", dims.Size, child.constraints)
	}
}

func TestBoxStyleAspectRatioShrinksFixedAxisToFit(t *testing.T) {
	dims := Box(nil).Style(style.Style{}.Width(80).AspectRatio(1)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 50)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(50, 50) {
		t.Fatalf("aspect box = %v, want (50,50)", dims.Size)
	}
}

func TestBoxStyleCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	box := Box(Spacer(40, 20)).Style(style.Style{}.Cursor(pointer.CursorPointer))
	layoutFrame := func() {
		var ops op.Ops
		box.Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Source:      router.Source(),
			Ops:         &ops,
		})
		router.Frame(&ops)
	}

	layoutFrame()
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(10, 10)})
	layoutFrame()
	if got := router.Cursor(); got != pointer.CursorPointer {
		t.Fatalf("cursor = %v, want %v", got, pointer.CursorPointer)
	}
}

func TestBoxPaddingUsesBorderBoxMinimum(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).Style(style.Style{}.MinHeight(36).Padding(6)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
	})

	if dims.Size.Y != 36 || child.constraints.Min.Y != 24 {
		t.Fatalf("box/child height = %d/%v, want total 36 and child minimum 24", dims.Size.Y, child.constraints)
	}
}

func TestBoxAppliesSidePadding(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(style.Style{}.Padding(1, 2, 3, 4)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAppliesMarginOutsideSize(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := Box(child).Style(style.Style{}.Width(100).Height(50).Margin(10)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(120, 70) {
		t.Fatalf("box size = %v, want (120,70)", dims.Size)
	}
	if child.constraints.Min != image.Pt(100, 50) || child.constraints.Max != image.Pt(100, 50) {
		t.Fatalf("child constraints = %v, want exact (100,50)", child.constraints)
	}
}

func TestBoxAppliesSideMargin(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(style.Style{}.MarginLeft(4).MarginRight(2).MarginTop(1).MarginBottom(3)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	if child.constraints.Max != image.Pt(94, 96) {
		t.Fatalf("child max constraints = %v, want (94,96)", child.constraints.Max)
	}
}

func TestBoxAlignClearsChildMinimum(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Box(child).Style(style.Style{}.FillWidth().FillHeight()).Align(AlignCenter).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != (image.Point{}) {
		t.Fatalf("child min constraints = %v, want zero", child.constraints.Min)
	}
	if child.constraints.Max != image.Pt(300, 200) {
		t.Fatalf("child max constraints = %v, want (300,200)", child.constraints.Max)
	}
}

func TestBoxClipConstrainsReportedSize(t *testing.T) {
	var ops op.Ops

	dims := Box(overflowWidget{}).Style(style.Style{}.Width(100).Height(50).Overflow(style.OverflowHidden)).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(100, 50) {
		t.Fatalf("clipped size = %v, want (100,50)", dims.Size)
	}
}

func TestBoxOverflowStyleDoesNotClip(t *testing.T) {
	// OverflowVisible is the default Style overflow; ensure an explicit value still layouts.
	dims := Box(Spacer(10, 10)).
		Style(style.Style{}.Width(20).Height(20).Overflow(style.OverflowVisible)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         new(op.Ops),
		})
	if dims.Size != image.Pt(20, 20) {
		t.Fatalf("overflow-visible box = %v, want (20,20)", dims.Size)
	}
}

func TestBoxOverflowHiddenStyleClips(t *testing.T) {
	dims := Box(overflowWidget{}).
		Style(style.Style{}.Width(100).Height(50).Overflow(style.OverflowHidden)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(200, 200)},
			Ops:         new(op.Ops),
		})
	if dims.Size != image.Pt(100, 50) {
		t.Fatalf("overflow-hidden box = %v, want (100,50)", dims.Size)
	}
}

func TestBoxPropagatesMarginPaddingAndAlignment(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(120, 100)}, Box(&overlayProbeWidget{
		key:    "box",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}).Style(style.Style{}.Width(80).Height(60).MarginLeft(7).MarginTop(9).Padding(5)).
		Align(AlignBottomEnd))
	want := image.Rect(62, 54, 82, 64)
	if got != want {
		t.Fatalf("box anchor = %v, want %v", got, want)
	}
}

func TestBoxClipHidesOverlayAnchorOutsideBox(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(200, 120)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	called := false
	probe := &overlayProbeWidget{
		key:    "clipped-box",
		size:   image.Pt(100, 100),
		anchor: image.Rect(80, 80, 90, 90),
		capture: func(image.Rectangle) {
			called = true
		},
	}
	core.BeginFrameWithViewport(ctx, viewport)
	Box(probe).Style(style.Style{}.Width(50).Height(50).Overflow(style.OverflowHidden)).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
	if called {
		t.Fatal("anchor outside a clipped Box produced an overlay")
	}
}

func TestBoxStyleAcceptsSeveralDeclarationsAtOnce(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).
		Style(style.Style{}.Width(100), style.Style{}.Padding(4), style.Style{}.Width(120).Padding(4, 4, 4, 20)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 300)},
			Ops:         new(op.Ops),
		})
	if dims.Size.X != 120 {
		t.Fatalf("joined width = %d, want 120", dims.Size.X)
	}
	if child.constraints.Max.X != 96 {
		t.Fatalf("joined padding child max X = %d, want 96", child.constraints.Max.X)
	}
}

func TestBoxReplaceStyleJoinsDeclarations(t *testing.T) {
	child := &constraintWidget{}
	dims := Box(child).
		Style(style.Style{}.Width(60)).
		ReplaceStyle(style.Style{}.Width(100).Padding(4), style.Style{}.Width(120).Padding(4, 4, 4, 20)).
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 300)},
			Ops:         new(op.Ops),
		})
	if dims.Size.X != 120 {
		t.Fatalf("replaced width = %d, want 120", dims.Size.X)
	}
	if child.constraints.Max.X != 96 {
		t.Fatalf("replaced padding child max X = %d, want 96", child.constraints.Max.X)
	}
}
