package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"

	"github.com/qianniancn/flowui/layout"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func radioGroup(selected *string) func() ui.Widget {
	return func() ui.Widget {
		return base.RadioGroup("group", *selected, []base.RadioItem{
			{Key: "a", Label: "Alpha"},
			{Key: "b", Label: "Beta"},
			{Key: "c", Label: "Gamma", Disabled: true},
			{Key: "d", Label: "Delta"},
		}).
			Label("Letters").
			Item(func(base.RadioState) ui.Widget { return layout.Spacer(120, 30) }).
			OnChange(func(next string) { *selected = next })
	}
}

func TestRadioGroupClickSelectsItem(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := "a"
	build := radioGroup(&selected)
	h.Frame(build())
	h.Click(f32.Pt(20, 45)) // second row
	h.Frame(build())
	if selected != "b" {
		t.Fatalf("selected = %q, want b", selected)
	}
}

func TestRadioGroupArrowsSelectAndSkipDisabled(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := "a"
	build := radioGroup(&selected)
	h.Frame(build())
	h.Click(f32.Pt(20, 15)) // first row; the click focuses the group
	h.Frame(build())

	h.Key(key.NameRightArrow, 0)
	h.Frame(build())
	if selected != "b" {
		t.Fatalf("selected after Right = %q, want b", selected)
	}

	h.Key(key.NameDownArrow, 0) // skips disabled c
	h.Frame(build())
	if selected != "d" {
		t.Fatalf("selected after Down = %q, want d", selected)
	}

	h.Key(key.NameHome, 0)
	h.Frame(build())
	if selected != "a" {
		t.Fatalf("selected after Home = %q, want a", selected)
	}
}

func TestRadioGroupDisabledIgnoresActivation(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := ""
	build := func() ui.Widget {
		return base.RadioGroup("group", "", []base.RadioItem{{Key: "a", Label: "Alpha"}}).
			Disabled(true).
			Item(func(base.RadioState) ui.Widget { return layout.Spacer(120, 30) }).
			OnChange(func(next string) { selected = next })
	}
	h.Frame(build())
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	if selected != "" {
		t.Fatalf("disabled group selected %q", selected)
	}
}
