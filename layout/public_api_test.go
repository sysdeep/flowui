package layout_test

import (
	"image"
	"testing"

	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestPublicLayoutComposition(t *testing.T) {
	var calls int
	var constraints giolayout.Constraints
	var decoration image.Rectangle
	child := ui.WidgetFunc(func(_ *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		calls++
		constraints = gtx.Constraints
		return giolayout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(10, 10))}
	})
	decorated := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		return layout.VisualOverflow(ctx, gtx, ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			return layout.VisualOutset(ctx, gtx, child, 2, 2, 2, 2)
		}), func(_ *ui.Context, _ giolayout.Context, bounds image.Rectangle) {
			decoration = bounds
		})
	})
	root := layout.Box(layout.Row(
		layout.Spacer(20, 10), layout.Expanded(decorated),
	).Gap(8)).Style(style.Style{}.FillWidth().FillHeight().Padding(10))

	h := uitest.New(image.Pt(160, 100), theme.Light())
	dims := h.Frame(root)
	if dims.Size != image.Pt(160, 100) || calls != 1 {
		t.Fatalf("size/calls = %v/%d, want (160,100)/1", dims.Size, calls)
	}
	if constraints.Min.X != 112 || constraints.Max.X != 112 {
		t.Fatalf("expanded child constraints = %v, want exact width 112", constraints)
	}
	if decoration != image.Rect(0, 0, 112, 80) {
		t.Fatalf("decoration = %v, want (0,0)-(112,80)", decoration)
	}
}
