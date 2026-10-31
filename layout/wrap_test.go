package layout

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

func TestWrapAndItemsSharePixelSpacingAndSingleLayout(t *testing.T) {
	for _, tc := range []struct {
		name string
		wrap bool
	}{
		{name: "Widget Wrap", wrap: true},
		{name: "Gio Items"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newContext(nil)
			viewport := image.Pt(100, 100)
			core.BeginFrameWithViewport(ctx, viewport)
			defer core.EndFrame(ctx)
			gtx := layout.Context{
				Constraints: layout.Constraints{Max: image.Pt(70, 100)},
				Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2}, Ops: new(op.Ops),
			}
			var calls [3]int
			var anchors [3]image.Rectangle
			widgets := make([]Widget, 3)
			gioChildren := make([]layout.Widget, 3)
			for index := range widgets {
				widgets[index] = WidgetFunc(func(ctx *Context, gtx layout.Context) layout.Dimensions {
					calls[index]++
					return (&overlayProbeWidget{
						key: string(rune('a' + index)), size: image.Pt(30, 10),
						anchor: image.Rect(0, 0, 30, 10), got: &anchors[index],
					}).Layout(ctx, gtx)
				})
				gioChildren[index] = func(gtx layout.Context) layout.Dimensions {
					return widgets[index].Layout(ctx, gtx)
				}
			}
			var dims layout.Dimensions
			if tc.wrap {
				dims = Wrap(widgets...).GapDp(3.5).LineGapDp(2.5).Layout(ctx, gtx)
			} else {
				dims = Items(ctx, gtx, true, 7, 5, gioChildren)
			}
			core.LayoutOverlays(ctx, gtx)
			if dims.Size != image.Pt(67, 25) || calls != [3]int{1, 1, 1} {
				t.Fatalf("size/calls = %v/%v, want (67,25)/[1 1 1]", dims.Size, calls)
			}
			want := [3]image.Rectangle{image.Rect(0, 0, 30, 10), image.Rect(37, 0, 67, 10), image.Rect(0, 15, 30, 25)}
			if anchors != want {
				t.Fatalf("anchors = %v, want %v", anchors, want)
			}
		})
	}
}

func TestWrapMovesChildrenToNextLine(t *testing.T) {
	var ops op.Ops

	dims := Wrap(
		Spacer(60, 10),
		Spacer(60, 20),
		Spacer(60, 30),
	).Gap(10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(130, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(130, 60) {
		t.Fatalf("wrap size = %v, want (130,60)", dims.Size)
	}
}

func TestWrapClampsNegativeGaps(t *testing.T) {
	dims := Wrap(Spacer(60, 10), Spacer(60, 10)).Gap(-10).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         new(op.Ops),
	})

	if dims.Size != image.Pt(120, 10) {
		t.Fatalf("wrap size = %v, want (120,10)", dims.Size)
	}
}

func TestWrapPropagatesWrappedAndAlignedPosition(t *testing.T) {
	got := resolveOverlayAnchor(t, layout.Constraints{Max: image.Pt(130, 100)}, Wrap(
		Spacer(60, 10),
		Spacer(60, 20),
		&overlayProbeWidget{key: "wrap", size: image.Pt(30, 10), anchor: image.Rect(0, 0, 30, 10)},
	).Gap(10).AlignEnd())
	want := image.Rect(100, 30, 130, 40)
	if got != want {
		t.Fatalf("wrap anchor = %v, want %v", got, want)
	}
}

func TestWrapPreservesZeroHeightRowBoundaries(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 20)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	var first, second image.Rectangle
	firstProbe := &overlayProbeWidget{key: "first-zero-row", size: image.Pt(90, 0), anchor: image.Rect(0, 0, 1, 1), got: &first}
	secondProbe := &overlayProbeWidget{key: "second-zero-row", size: image.Pt(30, 0), anchor: image.Rect(0, 0, 1, 1), got: &second}

	core.BeginFrameWithViewport(ctx, viewport)
	Wrap(firstProbe, secondProbe).Gap(10).AlignEnd().Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	if first.Min.X != 0 || second.Min.X != 60 {
		t.Fatalf("zero-height row positions = %v/%v, want x=0/60", first, second)
	}
}
