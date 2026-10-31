package selector

import (
	"fmt"
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestSelectOpensWithKeyboard(t *testing.T) {
	for _, test := range []struct {
		name string
		key  key.Name
	}{
		{name: "return", key: key.NameReturn}, {name: "enter", key: key.NameEnter},
		{name: "down", key: key.NameDownArrow}, {name: "up", key: key.NameUpArrow},
		{name: "space", key: key.NameSpace},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(260, 240), theme.Light())
			open := false
			build := func() ui.Widget {
				return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
					gtx.Constraints.Min = image.Point{}
					return New([]SelectItem{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}}).
						Key("keyboard-open").Open(open).OnOpenChange(func(value bool) { open = value }).Layout(ctx, gtx)
				})
			}
			h.Frame(build())
			h.Router().Source().Execute(key.FocusCmd{Tag: core.FieldFocusTag(h.Context(), "keyboard-open")})
			h.Frame(build())
			h.Key(test.key, 0)
			h.Frame(build())
			h.Frame(build())
			if !open {
				t.Fatal("focused select did not open")
			}
		})
	}
}

func TestSelectWidgetHeightAndSelection(t *testing.T) {
	h := uitest.New(image.Pt(300, 300), theme.Light())
	selected, open := "desktop", false
	build := func() SelectWidget {
		return New([]SelectItem{
			{Key: "desktop", Label: "Desktop"},
			{Key: "tools", Label: "Tools"},
		}).Key("platform-v2").SelectedKey(selected).Open(open).
			OnChange(func(value string) { selected = value }).
			OnOpenChange(func(value bool) { open = value })
	}
	root := func() ui.Widget {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return build().Layout(ctx, gtx)
		})
	}
	dims := h.Frame(root())
	if dims.Size.Y != 32 {
		t.Fatalf("select height = %d, want 32", dims.Size.Y)
	}
	h.Click(f32.Pt(40, 16))
	for range 3 {
		h.Frame(root())
	}
	if !open {
		t.Fatal("trigger did not open")
	}
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Tools")
	if !ok {
		t.Fatalf("missing item semantics: %v", uitest.SemanticLabels(h.Semantics()))
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	for range 3 {
		h.Frame(root())
	}
	if selected != "tools" || open {
		t.Fatalf("selected=%q open=%v", selected, open)
	}
}

func TestSelectWidgetStyleAndSections(t *testing.T) {
	h := uitest.New(image.Pt(300, 300), theme.Light())
	selected, open := "go", false
	build := func() SelectWidget {
		return NewSelectSections([]SelectSection{
			{Title: "Compiled", Items: []SelectItem{
				{Key: "go", Label: "Go"},
				{Key: "rust", Label: "Rust"},
			}},
			{Title: "Scripts", Items: []SelectItem{
				{Key: "python", Label: "Python"},
				{Key: "legacy", Label: "Legacy", Disabled: true},
			}},
		}).Key("language-sections").SelectedKey(selected).Open(open).
			FillWidth().
			Style(style.New().Height(40)).
			OnChange(func(value string) { selected = value }).
			OnOpenChange(func(value bool) { open = value })
	}
	root := func() ui.Widget {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return build().Layout(ctx, gtx)
		})
	}
	dims := h.Frame(root())
	if dims.Size.Y != 40 {
		t.Fatalf("styled select height = %d, want 40", dims.Size.Y)
	}
	h.Click(f32.Pt(40, 20))
	for range 3 {
		h.Frame(root())
	}
	if !open {
		t.Fatal("sectioned trigger did not open")
	}
	for _, label := range []string{"Compiled", "Scripts", "Go", "Rust", "Python", "Legacy"} {
		if _, ok := uitest.SemanticByLabel(h.Semantics(), label); !ok {
			t.Fatalf("missing %q semantics: %v", label, uitest.SemanticLabels(h.Semantics()))
		}
	}
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Python")
	if !ok {
		t.Fatalf("missing item semantics: %v", uitest.SemanticLabels(h.Semantics()))
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	for range 3 {
		h.Frame(root())
	}
	if selected != "python" || open {
		t.Fatalf("selected=%q open=%v", selected, open)
	}
}

func TestSelectLongListScrollsToLastOption(t *testing.T) {
	h := uitest.New(image.Pt(300, 240), theme.Light())
	items := make([]SelectItem, 30)
	for index := range items {
		items[index] = SelectItem{Key: fmt.Sprint(index), Label: fmt.Sprintf("Option %02d", index)}
	}
	selected, open := "", true
	root := func() ui.Widget {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return New(items).Key("long-select").Open(open).
				OnChange(func(value string) { selected = value }).
				OnOpenChange(func(value bool) { open = value }).Layout(ctx, gtx)
		})
	}
	h.Frame(root())
	h.Frame(root())
	h.Key(key.NameEnd, 0)
	h.Frame(root())
	h.Frame(root())
	if node, ok := uitest.SemanticByLabel(h.Semantics(), "Option 29"); !ok || node.Desc.Bounds.Dy() <= 0 {
		t.Fatal("keyboard highlight did not reveal the last option")
	}
	h.Key(key.NameHome, 0)
	h.Frame(root())
	h.Frame(root())
	h.Router().Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(50, 130), Scroll: f32.Pt(0, 2000)})
	h.Frame(root())
	h.Frame(root())
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Option 29")
	if !ok || node.Desc.Bounds.Dy() <= 0 || node.Desc.Bounds.Max.Y > 240 {
		t.Fatalf("last option not visible after scrolling: %v", node.Desc.Bounds)
	}
	b := node.Desc.Bounds
	h.Click(f32.Pt(float32(b.Min.X+b.Max.X)/2, float32(b.Min.Y+b.Max.Y)/2))
	h.Frame(root())
	h.Frame(root())
	if selected != "29" || open {
		t.Fatalf("selected=%q open=%v", selected, open)
	}
}
