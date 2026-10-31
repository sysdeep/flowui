package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

func selectTrigger(base.SelectState) ui.Widget { return popupTrigger(base.PopupState{}) }

func selectItem(state base.SelectItemState) ui.Widget {
	return text.New(state.Label)
}

func TestSelectOpensAndSelects(t *testing.T) {
	h := uitest.New(image.Pt(240, 260), theme.Light())
	selected, open, changes := "", false, 0
	build := func() ui.Widget {
		return base.Select("size", selected, []base.SelectItem{
			{Key: "s", Label: "Small"},
			{Key: "m", Label: "Medium"},
			{Key: "l", Label: "Large", Disabled: true},
		}).Label("Size").Open(open).
			OnOpenChange(func(next bool) { open = next }).
			OnChange(func(key string) { selected = key; changes++ }).
			Trigger(selectTrigger).Item(selectItem)
	}
	h.Frame(build())
	h.Click(f32.Pt(40, 16))
	h.Frame(build())
	if !open {
		t.Fatal("trigger did not open the select")
	}
	h.Frame(build())
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Small")
	if !ok {
		t.Fatalf("missing item semantics: %v", uitest.SemanticLabels(h.Semantics()))
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	h.Frame(build())
	if selected != "s" || changes != 1 || open {
		t.Fatalf("selected=%q changes=%d open=%v", selected, changes, open)
	}
}

func TestSelectOutsideClickDismisses(t *testing.T) {
	h := uitest.New(image.Pt(240, 260), theme.Light())
	open := true
	build := func() ui.Widget {
		return base.Select("size", "", []base.SelectItem{{Key: "s", Label: "Small"}}).
			Label("Size").Open(open).
			OnOpenChange(func(next bool) { open = next }).
			OnChange(func(string) {}).
			Trigger(selectTrigger).Item(selectItem)
	}
	h.Frame(build())
	h.Frame(build())
	h.Click(f32.Pt(200, 200))
	h.Frame(build())
	if open {
		t.Fatal("outside click did not dismiss the select")
	}
}

func TestSelectEscapeCloses(t *testing.T) {
	h := uitest.New(image.Pt(240, 260), theme.Light())
	open := true
	build := func() ui.Widget {
		return base.Select("size", "", []base.SelectItem{{Key: "s", Label: "Small"}}).
			Label("Size").Open(open).
			OnOpenChange(func(next bool) { open = next }).
			OnChange(func(string) {}).
			Trigger(selectTrigger).Item(selectItem)
	}
	h.Frame(build())
	h.Frame(build())
	tag := core.FieldFocusTag(h.Context(), "size")
	if tag == nil {
		t.Fatal("missing select field focus tag")
	}
	h.Router().Source().Execute(key.FocusCmd{Tag: tag})
	h.Frame(build())
	h.Key(key.NameEscape, 0)
	h.Frame(build())
	if open {
		t.Fatal("Escape did not close the select")
	}
}

func TestSelectKeyboardNavigation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		keys     []key.Name
		disabled bool
		want     string
	}{
		{name: "skip disabled", keys: []key.Name{key.NameDownArrow}, want: "c"},
		{name: "wrap", keys: []key.Name{key.NameUpArrow}, want: "c"},
		{name: "edges", keys: []key.Name{key.NameEnd, key.NameHome}, want: "a"},
		{name: "all disabled", keys: []key.Name{key.NameDownArrow, key.NameHome}, disabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 260), theme.Light())
			selected, open := "a", true
			changes := 0
			build := func() ui.Widget {
				return base.Select("size", selected, []base.SelectItem{
					{Key: "a", Label: "Alpha", Disabled: tc.disabled},
					{Key: "b", Label: "Beta", Disabled: true},
					{Key: "c", Label: "Charlie", Disabled: tc.disabled},
				}).Open(open).
					OnOpenChange(func(next bool) { open = next }).
					OnChange(func(next string) { selected = next; changes++ }).
					Trigger(selectTrigger).Item(selectItem)
			}
			h.Frame(build())
			h.Frame(build())
			for _, name := range tc.keys {
				h.Key(name, 0)
				h.Frame(build())
			}
			h.Key(key.NameReturn, 0)
			h.Frame(build())
			if tc.disabled {
				if changes != 0 || !open {
					t.Fatalf("disabled options activated: changes=%d open=%v", changes, open)
				}
			} else if selected != tc.want || changes != 1 || open {
				t.Fatalf("selected=%q changes=%d open=%v, want %q/1/false", selected, changes, open, tc.want)
			}
		})
	}
}
