package style_test

import (
	"image"
	"testing"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestPublicScopedStyleComposition(t *testing.T) {
	fontSize := unit.Sp(13)
	base := style.Join(
		style.Style{}.Width(120).Height(32).Padding(4),
		style.TextDeclaration(&style.TextStyle{FontSize: &fontSize}),
	).On(style.Hovered, style.Style{}.Opacity(.5)).
		Part(style.PartLabel, style.Style{}.TextColor(style.RGB(0xff0000)))
	var root, label style.ResolvedStyle
	child := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		state := style.State{Hovered: true}
		instance := style.Style{}.Padding(20)
		root = resolve.Resolve(ctx, gtx, "", state, base, instance)
		label = resolve.ResolvePart(ctx, gtx, "", style.PartLabel, state, base, instance)
		return giolayout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(120, 32))}
	})
	h := uitest.New(image.Pt(120, 32), theme.Light())
	dims := h.Frame(resolve.Scope(style.Style{}.Padding(12), child))
	if dims.Size != image.Pt(120, 32) {
		t.Fatalf("size = %v, want (120,32)", dims.Size)
	}
	if root.Box == nil || root.Box.Width == nil || root.Box.Height == nil || root.Box.Padding == nil {
		t.Fatalf("missing box declarations: %#v", root.Box)
	}
	if *root.Box.Width != 120 || *root.Box.Height != 32 || root.Box.Padding.Left != 20 {
		t.Fatalf("scope/instance cascade changed box declarations: %#v", root.Box)
	}
	if root.Paint == nil || root.Paint.Opacity == nil || *root.Paint.Opacity != .5 {
		t.Fatalf("hover opacity = %#v, want .5", root.Paint)
	}
	if label.Text == nil || label.Text.FontSize == nil || *label.Text.FontSize != fontSize {
		t.Fatalf("label did not inherit text size: %#v", label.Text)
	}
	if got, ok := resolve.Color(label.Text.Color); !ok || got != style.RGB(0xff0000).Color {
		t.Fatalf("label color = %v/%v, want red", got, ok)
	}
}
