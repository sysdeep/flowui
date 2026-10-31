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

func buttonContent(base.ButtonState) ui.Widget {
	return layout.Spacer(120, 40)
}

func TestButtonClickAndKeyboardReportAction(t *testing.T) {
	h := uitest.New(image.Pt(160, 60), theme.Light())
	clicks := 0
	build := func() ui.Widget {
		return base.Button("save").
			Label("Save").
			Content(buttonContent).
			OnClick(func() { clicks++ })
	}

	h.Frame(build())
	position := f32.Pt(60, 20)
	h.Router().Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: position,
	})
	h.Frame(build())
	h.Router().Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: position,
	})
	h.Frame(build())
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}

	h.Key(key.NameSpace, 0)
	h.Frame(build())
	if clicks != 2 {
		t.Fatalf("clicks after Space = %d, want 2", clicks)
	}
}

func TestButtonDisabledIgnoresActivation(t *testing.T) {
	h := uitest.New(image.Pt(160, 60), theme.Light())
	clicks := 0
	view := base.Button("save").
		Label("Save").
		Disabled(true).
		Content(buttonContent).
		OnClick(func() { clicks++ })
	h.Frame(view)
	h.Click(f32.Pt(60, 20))
	h.Frame(view)
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func TestButtonExposesLabelAndRole(t *testing.T) {
	h := uitest.New(image.Pt(160, 60), theme.Light())
	h.Frame(base.Button("save").Label("Save").Content(buttonContent))
	nodes := h.Semantics()
	found := false
	uitest.WalkSemantics(nodes, func(node uitest.Semantic) bool {
		if node.Desc.Class == semantic.Button && node.Desc.Label == "Save" {
			found = true
			return false
		}
		return true
	})
	if !found {
		t.Fatalf("semantics = %v, want a button named Save", uitest.SemanticLabels(nodes))
	}
}
