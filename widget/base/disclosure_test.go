package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/layout"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/button"
)

func disclosureView(expanded *bool, contentLayouts *int) func() ui.Widget {
	return func() ui.Widget {
		return base.Disclosure("more", *expanded).
			Label("More").
			OnToggle(func(next bool) { *expanded = next }).
			Trigger(func(base.DisclosureState) ui.Widget { return layout.Spacer(120, 30) }).
			Content(func() ui.Widget {
				*contentLayouts++
				return layout.Spacer(120, 30)
			})
	}
}

func TestDisclosureClickExpandsAndLaysOutContent(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	expanded := false
	contentLayouts := 0
	build := disclosureView(&expanded, &contentLayouts)
	h.Frame(build())
	if contentLayouts != 0 {
		t.Fatalf("collapsed disclosure laid out content %d times", contentLayouts)
	}
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	if !expanded {
		t.Fatal("click did not expand")
	}
	h.Frame(build())
	if contentLayouts != 1 {
		t.Fatalf("expanded disclosure content layouts = %d, want 1", contentLayouts)
	}
}

func TestDisclosureDisabledIgnoresActivation(t *testing.T) {
	h := uitest.New(image.Pt(160, 160), theme.Light())
	expanded := false
	build := func() ui.Widget {
		return base.Disclosure("more", false).
			Disabled(true).
			OnToggle(func(next bool) { expanded = next }).
			Trigger(func(base.DisclosureState) ui.Widget { return layout.Spacer(120, 30) })
	}
	h.Frame(build())
	h.Click(f32.Pt(20, 15))
	h.Frame(build())
	if expanded {
		t.Fatal("disabled disclosure toggled")
	}
}

func TestDisclosureUpdateOwnsTriggerBehaviorAndSemantics(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	var click widget.Clickable
	expanded, disabled, toggles := false, false, 0
	view := core.WidgetFunc(func(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
		state := base.Disclosure("details", expanded).Label("Details").
			Clickable(&click).Disabled(disabled).
			OnToggle(func(next bool) { expanded = next; toggles++ }).Update(ctx, gtx)
		visual := button.New("Details").Visual(base.ButtonState{
			Interaction: state.Interaction, Key: state.Key,
		})
		return state.LayoutTrigger(ctx, gtx, visual)
	})
	h.Frame(view)
	h.Click(f32.Pt(20, 15))
	h.Frame(view)
	h.Frame(view)
	if toggles != 1 || !expanded {
		t.Fatalf("pointer activation = %d/%v, want 1/true", toggles, expanded)
	}
	buttons := 0
	uitest.WalkSemantics(h.Semantics(), func(node uitest.Semantic) bool {
		if node.Desc.Class == semantic.Button {
			buttons++
			if node.Desc.Label != "Details" || !node.Desc.Selected || node.Desc.Bounds.Empty() {
				t.Errorf("trigger semantics = %+v", node.Desc)
			}
		}
		return true
	})
	if buttons != 1 {
		t.Fatalf("button semantic nodes = %d, want 1", buttons)
	}
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Key(key.NameSpace, 0)
	h.Frame(view)
	if toggles != 2 || expanded {
		t.Fatalf("keyboard activation = %d/%v, want 2/false", toggles, expanded)
	}
	disabled = true
	h.Frame(view)
	h.Click(f32.Pt(20, 15))
	h.Frame(view)
	if toggles != 2 {
		t.Fatalf("disabled trigger toggles = %d, want 2", toggles)
	}
}
