package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"

	"github.com/qianniancn/flowui/layout"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func toggleContent(base.ToggleState) ui.Widget {
	return layout.Spacer(200, 40)
}

func TestToggleClickAndKeyboardReportChanges(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	checked := false
	var got []bool
	build := func() ui.Widget {
		return base.Toggle("sync", checked).
			Label("Sync").
			Content(toggleContent).
			OnChange(func(next bool) {
				checked = next
				got = append(got, next)
			})
	}

	h.Frame(build())
	position := f32.Pt(100, 20)
	h.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	h.Frame(build())
	h.Router().Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	h.Frame(build())
	if len(got) != 1 || !got[0] || !checked {
		t.Fatalf("click changes = %v, checked = %v; want one true change", got, checked)
	}

	h.Key(key.NameTab, 0)
	h.Frame(build())
	h.Key(key.NameSpace, 0)
	h.Frame(build())
	if len(got) != 2 || got[1] || checked {
		t.Fatalf("Space changes = %v, checked = %v; want false after true", got, checked)
	}
}

func TestToggleExposesRoleAndLabel(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	h.Frame(base.Toggle("sync", true).
		Label("Sync").
		Description("Keep files in step").
		Content(toggleContent))
	if !hasSemanticClass(h.Semantics(), semantic.CheckBox, "Sync") {
		t.Fatalf("semantics = %v, want a check box named Sync", uitest.SemanticLabels(h.Semantics()))
	}

	h.Frame(base.Toggle("mode", true).
		Role(base.ToggleSwitch).
		Label("Mode").
		Content(toggleContent))
	if !hasSemanticClass(h.Semantics(), semantic.Switch, "Mode") {
		t.Fatalf("semantics = %v, want a switch named Mode", uitest.SemanticLabels(h.Semantics()))
	}
}

func TestToggleDisabledAndReadOnlyIgnoreActivation(t *testing.T) {
	for _, test := range []struct {
		name string
		view func(func(bool)) ui.Widget
	}{
		{
			name: "disabled",
			view: func(onChange func(bool)) ui.Widget {
				return base.Toggle("sync", false).Disabled(true).Content(toggleContent).OnChange(onChange)
			},
		},
		{
			name: "read-only",
			view: func(onChange func(bool)) ui.Widget {
				return base.Toggle("sync", false).ReadOnly(true).Content(toggleContent).OnChange(onChange)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 80), theme.Light())
			var got []bool
			view := test.view(func(next bool) { got = append(got, next) })
			h.Frame(view)
			h.Click(f32.Pt(100, 20))
			h.Frame(view)
			if len(got) != 0 {
				t.Fatalf("changes = %v, want none", got)
			}
		})
	}
}

func hasSemanticClass(nodes []uitest.Semantic, class semantic.ClassOp, label string) bool {
	found := false
	uitest.WalkSemantics(nodes, func(node uitest.Semantic) bool {
		if node.Desc.Class == class && node.Desc.Label == label {
			found = true
			return false
		}
		return true
	})
	return found
}
