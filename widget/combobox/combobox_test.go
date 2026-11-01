package combobox

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	fieldinput "github.com/qianniancn/flowui/widget/input"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
)

func TestComboboxSizes(t *testing.T) {
	for _, test := range []struct {
		name   string
		size   fieldinput.Size
		height int
	}{
		{name: "small", size: fieldinput.Small, height: 24},
		{name: "medium", size: fieldinput.Medium, height: 32},
		{name: "large", size: fieldinput.Large, height: 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(300, 240), theme.Light())
			root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return New(nil).Size(test.size).Layout(ctx, gtx)
			})
			if got := h.Frame(root).Size.Y; got != test.height {
				t.Fatalf("height=%d, want %d", got, test.height)
			}
		})
	}
}

func TestComboboxClosedArrowOpensAndScrollsToLastOption(t *testing.T) {
	h := uitest.New(image.Pt(300, 240), theme.Light())
	items := make([]ComboboxItem, 30)
	for index := range items {
		items[index] = ComboboxItem{Key: fmt.Sprint(index), Label: fmt.Sprintf("Option %02d", index)}
	}
	open, selected := false, ""
	root := func() core.Widget {
		return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return New(items).Key("long-combo").Open(open).
				OnOpenChange(func(value bool) { open = value }).
				OnChange(func(value string) { selected = value }).Layout(ctx, gtx)
		})
	}
	h.Frame(root())
	h.Click(f32.Pt(20, 15))
	h.Frame(root())
	h.Frame(root())
	h.Key(key.NameDownArrow, 0)
	h.Frame(root())
	h.Frame(root())
	if !open {
		t.Fatal("ArrowDown did not open focused combobox")
	}
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

func TestComboboxWidgetOpensAndSelects(t *testing.T) {
	h := uitest.New(image.Pt(260, 240), theme.Light())
	selected, input, open := "go", "Go", false
	build := func() ComboboxWidget {
		return New([]ComboboxItem{
			{Key: "go", Label: "Go"},
			{Key: "rust", Label: "Rust"},
		}).Key("lang-v2").SelectedKey(selected).InputValue(input).Open(open).
			OnChange(func(value string) { selected = value }).
			OnInputChange(func(value string) { input = value }).
			OnOpenChange(func(value bool) { open = value })
	}
	h.Frame(build())
	if node, ok := uitest.SemanticByLabel(h.Semantics(), "Open options"); ok {
		b := node.Desc.Bounds
		h.Click(f32.Pt(float32(b.Min.X+b.Max.X)/2, float32(b.Min.Y+b.Max.Y)/2))
	} else {
		t.Fatalf("missing chevron: %v", uitest.SemanticLabels(h.Semantics()))
	}
	h.Frame(build())
	if !open {
		t.Fatal("chevron did not open combobox")
	}
	h.Frame(build())
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Rust")
	if !ok {
		t.Fatalf("missing Rust item: %v", uitest.SemanticLabels(h.Semantics()))
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	h.Frame(build())
	if selected != "rust" || input != "Rust" || open {
		t.Fatalf("selected=%q input=%q open=%v", selected, input, open)
	}
}

func TestComboboxKeyboardSelects(t *testing.T) {
	h := uitest.New(image.Pt(260, 240), theme.Light())
	selected, input, open := "go", "Go", true
	build := func() ComboboxWidget {
		return New([]ComboboxItem{{Key: "go", Label: "Go"}, {Key: "rust", Label: "Rust"}}).
			Key("lang-keys").SelectedKey(selected).InputValue(input).Open(open).
			OnChange(func(value string) { selected = value }).
			OnInputChange(func(value string) { input = value }).
			OnOpenChange(func(value bool) { open = value })
	}
	h.Frame(build())
	h.Frame(build())
	inputKey := core.DerivedKey(h.Context(), "lang-keys", "input")
	tag := core.FieldFocusTag(h.Context(), core.FullKey(h.Context(), inputKey))
	if tag == nil {
		t.Fatal("missing combobox input focus tag")
	}
	h.Router().Source().Execute(key.FocusCmd{Tag: tag})
	h.Frame(build())
	h.Key(key.NameDownArrow, 0)
	h.Frame(build())
	h.Key(key.NameReturn, 0)
	h.Frame(build())
	if selected != "rust" || input != "Rust" || open {
		t.Fatalf("selected=%q input=%q open=%v", selected, input, open)
	}
}

func TestComboboxSelectsFilteredResultWithKeyboard(t *testing.T) {
	h := uitest.New(image.Pt(260, 240), theme.Light())
	selected, query := "", ""
	build := func() ComboboxWidget {
		return New([]ComboboxItem{{Key: "go", Label: "Go"}, {Key: "rust", Label: "Rust"}, {Key: "java", Label: "Java"}}).
			Key("filter").Open(true).InputValue(query).
			OnChange(func(value string) { selected = value }).
			OnInputChange(func(value string) { query = value })
	}
	h.Frame(build())
	h.Frame(build())
	inputKey := core.DerivedKey(h.Context(), "filter", "input")
	h.Router().Source().Execute(key.FocusCmd{Tag: core.FieldFocusTag(h.Context(), core.FullKey(h.Context(), inputKey))})
	h.Frame(build())
	h.Key(key.NameEnd, 0)
	h.Frame(build())
	h.Router().Queue(key.EditEvent{Range: key.Range{}, Text: "Go"})
	h.Frame(build())
	h.Frame(build())
	h.Key(key.NameReturn, 0)
	h.Frame(build())
	if query != "Go" || selected != "go" {
		t.Fatalf("query=%q selected=%q; want Go/go", query, selected)
	}
}

func TestComboboxKeyboardSkipsDisabledOptions(t *testing.T) {
	for _, test := range []struct {
		name string
		keys []key.Name
		want string
	}{
		{name: "down", keys: []key.Name{key.NameDownArrow}, want: "c"},
		{name: "up wraps", keys: []key.Name{key.NameUpArrow}, want: "c"},
		{name: "home", keys: []key.Name{key.NameEnd, key.NameHome}, want: "a"},
		{name: "end", keys: []key.Name{key.NameEnd}, want: "c"},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(260, 240), theme.Light())
			selected := ""
			build := func() ComboboxWidget {
				return New([]ComboboxItem{
					{Key: "first", Label: "First", Disabled: true},
					{Key: "a", Label: "A"}, {Key: "b", Label: "B", Disabled: true},
					{Key: "c", Label: "C"}, {Key: "last", Label: "Last", Disabled: true},
				}).Key("disabled-options").Open(true).OnChange(func(value string) { selected = value })
			}
			h.Frame(build())
			h.Frame(build())
			inputKey := core.DerivedKey(h.Context(), "disabled-options", "input")
			h.Router().Source().Execute(key.FocusCmd{Tag: core.FieldFocusTag(h.Context(), core.FullKey(h.Context(), inputKey))})
			h.Frame(build())
			for _, name := range test.keys {
				h.Key(name, 0)
				h.Frame(build())
			}
			h.Key(key.NameReturn, 0)
			h.Frame(build())
			if selected != test.want {
				t.Fatalf("selected=%q, want %q", selected, test.want)
			}
		})
	}
}

func TestComboboxFillWidth(t *testing.T) {
	h := uitest.New(image.Pt(360, 80), theme.Light())
	widget := New([]ComboboxItem{{Key: "go", Label: "Go"}}).Key("fill-width").FillWidth()
	if got := h.Frame(widget).Size.X; got != 360 {
		t.Fatalf("filled width = %d, want 360", got)
	}
}

func TestComboboxInputSemanticBounds(t *testing.T) {
	h := uitest.New(image.Pt(260, 80), theme.Light())
	h.Frame(New([]ComboboxItem{{Key: "go", Label: "Go"}}).Key("input-bounds").InputValue("Go").FillWidth())
	node, ok := uitest.FindSemantic(h.Semantics(), func(node uitest.Semantic) bool {
		return node.Desc.Class == semantic.Editor
	})
	if !ok {
		t.Fatalf("missing combobox editor semantic: %v", uitest.SemanticLabels(h.Semantics()))
	}
	if node.Desc.Bounds.Dx() <= 0 || node.Desc.Bounds.Dy() <= 0 {
		t.Fatalf("combobox editor bounds = %v", node.Desc.Bounds)
	}
}

func TestComboboxInputTextIsPainted(t *testing.T) {
	window, err := headless.NewWindow(260, 80)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageAuto)
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(image.Pt(260, 80)), Source: router.Source(), Ops: &ops}
	New([]ComboboxItem{{Key: "go", Label: "Go"}}).Key("paint").InputValue("Go").FillWidth().Layout(ctx, gtx)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 260, 80))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	fieldPixel := color.NRGBAModel.Convert(pixels.At(20, 5)).(color.NRGBA)
	if fieldPixel != active.Semantic().FieldBackground {
		t.Fatalf("combobox field background = %#v, want %#v", fieldPixel, active.Semantic().FieldBackground)
	}
	background := color.NRGBAModel.Convert(pixels.At(0, 0)).(color.NRGBA)
	ink := 0
	for y := 8; y < 28; y++ {
		for x := 8; x < 100; x++ {
			if color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA) != background {
				ink++
			}
		}
	}
	if ink == 0 {
		t.Fatalf("combobox input produced no visible pixels; background=%#v", background)
	}
}

func TestComboboxClickFocusesAndEditsInput(t *testing.T) {
	h := uitest.New(image.Pt(260, 80), theme.Light())
	value := "Go"
	build := func() ComboboxWidget {
		return New([]ComboboxItem{{Key: "go", Label: "Go"}}).
			Key("click-edit").InputValue(value).OnInputChange(func(next string) { value = next })
	}
	h.Frame(build())
	h.Click(f32.Pt(40, 16))
	h.Frame(build())
	inputKey := core.DerivedKey(h.Context(), "click-edit", "input")
	tag := core.FieldFocusTag(h.Context(), core.FullKey(h.Context(), inputKey))
	if tag == nil || !h.Router().Source().Focused(tag) {
		t.Fatalf("combobox input did not receive focus: tag=%v", tag)
	}
	h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: len(value)}, Text: "Rust"})
	h.Frame(build())
	if value != "Rust" {
		t.Fatalf("combobox input value = %q, want Rust", value)
	}
}
