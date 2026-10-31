package layout

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
)

func TestLayoutItemsPropagatesOverlayPositions(t *testing.T) {
	tests := []struct {
		name        string
		horizontal  bool
		constraints layout.Constraints
		want        image.Rectangle
	}{
		{
			name:        "vertical second item",
			constraints: layout.Constraints{Max: image.Pt(100, 100)},
			want:        image.Rect(0, 15, 20, 25),
		},
		{
			name:        "horizontal wrapped item",
			horizontal:  true,
			constraints: layout.Constraints{Max: image.Pt(70, 100)},
			want:        image.Rect(0, 15, 20, 25),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			core.BeginFrameWithViewport(ctx, image.Pt(100, 100))
			gtx := layout.Context{Constraints: test.constraints, Ops: new(op.Ops)}
			probe := &overlayProbeWidget{key: "items", size: image.Pt(20, 10), anchor: image.Rect(0, 0, 20, 10)}
			var got image.Rectangle
			probe.got = &got
			children := []layout.Widget{
				func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(60, 10)} },
				func(gtx layout.Context) layout.Dimensions { return probe.Layout(ctx, gtx) },
			}
			Items(ctx, gtx, test.horizontal, 10, 5, children)
			core.LayoutOverlays(ctx, gtx)
			core.EndFrame(ctx)
			if got != test.want {
				t.Fatalf("item anchor = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLayoutItemsClampsNegativeGaps(t *testing.T) {
	dims := Items(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	}, true, -10, -10, []layout.Widget{
		func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(20, 10)} },
		func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(20, 10)} },
	})

	if dims.Size != image.Pt(40, 10) {
		t.Fatalf("items size = %v, want (40,10)", dims.Size)
	}
}
