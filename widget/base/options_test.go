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

func optionList(selected *string) func() ui.Widget {
	return func() ui.Widget {
		return base.Options("list", []base.Option{
			{Key: "a", Label: "Alpha"},
			{Key: "b", Label: "Beta"},
			{Key: "c", Label: "Gamma", Disabled: true},
			{Key: "d", Label: "Delta"},
		}).
			Label("Letters").
			Selected("a").
			Item(func(base.OptionState) ui.Widget { return layout.Spacer(120, 30) }).
			OnSelect(func(next string) { *selected = next })
	}
}

func TestOptionsClickSelectsRow(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := ""
	build := optionList(&selected)
	h.Frame(build())
	h.Click(f32.Pt(20, 45)) // second row
	h.Frame(build())
	if selected != "b" {
		t.Fatalf("selected = %q, want b", selected)
	}
}

func TestOptionsKeyboardNavigatesAndSkipsDisabled(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := ""
	build := optionList(&selected)
	h.Frame(build())
	h.Click(f32.Pt(20, 15)) // first row; the click also focuses the list
	h.Frame(build())

	h.Key(key.NameDownArrow, 0)
	h.Frame(build())
	h.Key(key.NameSpace, 0)
	h.Frame(build())
	if selected != "b" {
		t.Fatalf("selected after Down+Space = %q, want b", selected)
	}

	h.Key(key.NameDownArrow, 0) // skips disabled c
	h.Frame(build())
	h.Key(key.NameSpace, 0)
	h.Frame(build())
	if selected != "d" {
		t.Fatalf("selected after second Down+Space = %q, want d", selected)
	}

	h.Key(key.NameDownArrow, 0) // wraps to first enabled
	h.Frame(build())
	h.Key(key.NameReturn, 0)
	h.Frame(build())
	if selected != "a" {
		t.Fatalf("selected after wrap = %q, want a", selected)
	}
}

func TestOptionsDisabledListIgnoresActivation(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	selected := ""
	build := func() ui.Widget {
		return base.Options("list", []base.Option{{Key: "a", Label: "Alpha"}}).
			Disabled(true).
			Item(func(base.OptionState) ui.Widget { return layout.Spacer(120, 30) }).
			OnSelect(func(next string) { selected = next })
	}
	h.Frame(build())
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	if selected != "" {
		t.Fatalf("disabled list selected %q", selected)
	}
}

func TestOptionsKeyboardActivationCancelledByDisable(t *testing.T) {
	h := uitest.New(image.Pt(160, 80), theme.Light())
	disabled, actions := false, 0
	build := func() ui.Widget {
		return base.Options("list", []base.Option{{Key: "a", Label: "Alpha"}}).
			Disabled(disabled).
			Item(func(base.OptionState) ui.Widget { return layout.Spacer(120, 30) }).
			OnSelect(func(string) { actions++ })
	}
	h.Frame(build())
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	actions = 0
	h.Router().Queue(key.Event{Name: key.NameSpace, State: key.Press})
	h.Frame(build())
	if actions != 0 {
		t.Fatal("activated before key release")
	}
	disabled = true
	h.Frame(build())
	disabled = false
	h.Frame(build())
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	actions = 0
	h.Router().Queue(key.Event{Name: key.NameSpace, State: key.Release})
	h.Frame(build())
	if actions != 0 {
		t.Fatal("stale key release activated the re-enabled group")
	}
	h.Key(key.NameSpace, 0)
	h.Frame(build())
	if actions != 1 {
		t.Fatalf("fresh activation = %d, want 1", actions)
	}
}
