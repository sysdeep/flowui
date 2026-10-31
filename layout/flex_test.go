package layout

import (
	"fmt"
	"image"
	"math"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

func TestFlexFractionalDpGaps(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value Widget
		want  image.Point
	}{
		{name: "row", value: Row(Spacer(10, 10), Spacer(10, 10)).GapDp(2.5), want: image.Pt(45, 20)},
		{name: "column", value: Column(Spacer(10, 10), Spacer(10, 10)).GapDp(1.5), want: image.Pt(20, 43)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gtx := testLayoutContext()
			gtx.Metric = unit.Metric{PxPerDp: 2, PxPerSp: 2}
			if dims := tc.value.Layout(newContext(nil), gtx); dims.Size != tc.want {
				t.Fatalf("size = %v, want %v", dims.Size, tc.want)
			}
		})
	}
}

func TestFlexGeometryModifiers(t *testing.T) {
	var childConstraints layout.Constraints
	child := WidgetFunc(func(_ *Context, gtx layout.Context) layout.Dimensions {
		childConstraints = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(10, 10))}
	})
	gtx := testLayoutContext()
	dims := Row(child).Width(100).Padding(10, 20).Layout(newContext(nil), gtx)
	if dims.Size.X != 100 {
		t.Fatalf("row width = %d, want 100", dims.Size.X)
	}
	if childConstraints.Max.X != 60 {
		t.Fatalf("child max width = %d, want 60 after horizontal padding", childConstraints.Max.X)
	}
	if dims.Size.Y != 30 {
		t.Fatalf("row height = %d, want 30 from vertical padding", dims.Size.Y)
	}

	dims = Column(Spacer(10, 10)).FillWidth().Layout(newContext(nil), gtx)
	if dims.Size.X != gtx.Constraints.Max.X {
		t.Fatalf("filled column width = %d, want %d", dims.Size.X, gtx.Constraints.Max.X)
	}
}

func TestRowSpacingBetweenUsesRemainingMainAxisSpace(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 20)
	var first, second image.Rectangle
	firstProbe := &overlayProbeWidget{key: "first", size: image.Pt(10, 10), anchor: image.Rect(0, 0, 10, 10), got: &first}
	secondProbe := &overlayProbeWidget{key: "second", size: image.Pt(10, 10), anchor: image.Rect(0, 0, 10, 10), got: &second}
	gtx := layout.Context{Constraints: layout.Exact(viewport), Ops: new(op.Ops)}
	core.BeginFrameWithViewport(ctx, viewport)
	Row(firstProbe, secondProbe).Spacing(SpaceBetween).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	if first != image.Rect(0, 0, 10, 10) || second != image.Rect(90, 0, 100, 10) {
		t.Fatalf("spacing positions = %v/%v, want (0,0)-(10,10) and (90,0)-(100,10)", first, second)
	}
}

func TestFlexSpacingModes(t *testing.T) {
	for _, test := range []struct {
		name        string
		spacing     FlexSpacing
		wantStart   int
		wantBetween int
	}{
		{name: "end", spacing: SpaceEnd},
		{name: "start", spacing: SpaceStart, wantStart: 10},
		{name: "sides", spacing: SpaceSides, wantStart: 5},
		{name: "around", spacing: SpaceAround, wantStart: 2, wantBetween: 5},
		{name: "between", spacing: SpaceBetween, wantBetween: 10},
		{name: "evenly", spacing: SpaceEvenly, wantStart: 3, wantBetween: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			start, between := flexSpacing(test.spacing, 10, 2)
			if start != test.wantStart || between != test.wantBetween {
				t.Fatalf("spacing = %d/%d, want %d/%d", start, between, test.wantStart, test.wantBetween)
			}
		})
	}
}

func TestCenterPropagatesOverlayPosition(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 80)), Center(&overlayProbeWidget{
		key:    "center",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}))
	want := image.Rect(40, 35, 60, 45)
	if got != want {
		t.Fatalf("center anchor = %v, want %v", got, want)
	}
}

func TestFlexPropagatesOverlayPositions(t *testing.T) {
	t.Run("row alignment", func(t *testing.T) {
		got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 40)), Row(
			Spacer(10, 10),
			&overlayProbeWidget{key: "row", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)},
		).Gap(5).AlignEnd())
		want := image.Rect(15, 30, 35, 40)
		if got != want {
			t.Fatalf("row anchor = %v, want %v", got, want)
		}
	})

	t.Run("column flexible child", func(t *testing.T) {
		got := resolveOverlayAnchor(t, layout.Exact(image.Pt(60, 80)), Column(
			Spacer(10, 10),
			Flexible(1, &overlayProbeWidget{key: "column", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 10, 10)}),
		).Gap(5).AlignMiddle())
		want := image.Rect(20, 15, 30, 25)
		if got != want {
			t.Fatalf("column anchor = %v, want %v", got, want)
		}
	})
}

func TestFlexOverlayPositionMatchesIgnoredNegativeGap(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Exact(image.Pt(100, 30)), Row(
		Spacer(10, 10),
		&overlayProbeWidget{key: "negative-gap", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)},
	).Gap(-5))
	want := image.Rect(10, 0, 30, 10)
	if got != want {
		t.Fatalf("negative-gap anchor = %v, want %v", got, want)
	}
}

func TestRowAlignBaselinePositionsChildrenOnOneBaseline(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 30)
	firstBounds := image.Rectangle{}
	secondBounds := image.Rectangle{}
	first := &overlayProbeWidget{
		key:      "baseline-first",
		size:     image.Pt(10, 14),
		baseline: 2,
		anchor:   image.Rect(0, 0, 1, 1),
		got:      &firstBounds,
	}
	second := &overlayProbeWidget{
		key:      "baseline-second",
		size:     image.Pt(10, 10),
		baseline: 0,
		anchor:   image.Rect(0, 0, 1, 1),
		got:      &secondBounds,
	}
	gtx := layout.Context{Constraints: layout.Exact(viewport), Ops: new(op.Ops)}

	core.BeginFrameWithViewport(ctx, viewport)
	Row(first, second).AlignBaseline().Gap(4).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	if firstBounds.Min.Y != 0 || secondBounds.Min.Y != 2 {
		t.Fatalf("baseline offsets = %d/%d, want 0/2", firstBounds.Min.Y, secondBounds.Min.Y)
	}
}

func TestExpandedUsesRemainingRowSpace(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Row(
		Spacer(50, 10),
		Expanded(child),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 40)},
		Ops:         &ops,
	})

	if child.constraints.Min.X != 140 || child.constraints.Max.X != 140 {
		t.Fatalf("expanded width constraints = %v, want exact 140", child.constraints)
	}
}

func TestFlexibleOffersRemainingSpaceByWeight(t *testing.T) {
	left := &constraintWidget{}
	right := &constraintWidget{}
	var ops op.Ops

	Row(
		Flexible(1, left),
		Flexible(2, right),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 40)},
		Ops:         &ops,
	})

	if left.constraints.Min.X != 0 || left.constraints.Max.X != 100 {
		t.Fatalf("left constraints = %v, want loose maximum 100", left.constraints)
	}
	if right.constraints.Min.X != 0 || right.constraints.Max.X != 200 {
		t.Fatalf("right constraints = %v, want loose maximum 200", right.constraints)
	}
}

func TestFlexibleCanUseLessThanItsShare(t *testing.T) {
	dims := Row(Flexible(1, Spacer(20, 10))).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 40)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(20, 10) {
		t.Fatalf("loose Flexible size = %v, want (20,10)", dims.Size)
	}
}

func TestFlexibleRejectsInvalidWeight(t *testing.T) {
	for _, weight := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1))} {
		t.Run(fmt.Sprintf("weight-%v", weight), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid flex weight panic")
				}
			}()
			Flexible(weight, Spacer(1, 1))
		})
	}
}
