package main

import (
	"image"
	"slices"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
)

func TestOverviewInteractions(t *testing.T) {
	active := theme.Light()
	harness := uitest.NewWithConfig(uitest.Config{
		Size: image.Pt(1440, 1100), Theme: &active, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
	})
	model := initialModel()
	frame := func() {
		for range 3 {
			harness.Advance(time.Second)
			harness.Frame(View(harness.Context(), model, func(msg Msg) { update(&model, msg) }))
		}
	}
	click := func(label string) {
		t.Helper()
		node, ok := uitest.SemanticByLabel(harness.Semantics(), label)
		if control, found := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
			return node.Desc.Label == label && node.Desc.Class != semantic.Unknown
		}); found {
			node, ok = control, true
		}
		if !ok {
			t.Fatalf("missing %q: %v", label, uitest.SemanticLabels(harness.Semantics()))
		}
		bounds := node.Desc.Bounds
		position := f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2)
		if node.Desc.Class == semantic.Unknown {
			// Plain table text inherits the scroll area's unconstrained height.
			position = f32.Pt(float32(bounds.Min.X+10), float32(bounds.Min.Y+10))
		}
		harness.Click(position)
		frame()
	}
	edit := func(field, old, value string) {
		t.Helper()
		tag := core.FieldFocusTag(harness.Context(), field)
		if tag == nil {
			t.Fatalf("missing editor %s", field)
		}
		harness.Router().Source().Execute(key.FocusCmd{Tag: tag})
		frame()
		harness.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: len([]rune(old))}, Text: value})
		frame()
	}
	frame()
	click("Month")
	if model.demo.overview.period != "month" {
		t.Fatal("month control did not update the chart period")
	}
	click("Auto-save")
	edit("overview-name", "FlowUI Desktop", "Unsaved project")
	if model.demo.overview.settings.name != "Unsaved project" || model.demo.overview.saved.name != "FlowUI Desktop" {
		t.Fatalf("manual edit: settings=%+v saved=%+v", model.demo.overview.settings, model.demo.overview.saved)
	}
	click("Cancel")
	if model.demo.overview.settings.name != "FlowUI Desktop" {
		t.Fatal("cancel did not restore saved settings")
	}
	click("Auto-save")
	edit("overview-name", "FlowUI Desktop", "Saved project")
	click("Save changes")
	if model.demo.overview.saved.name != "Saved project" || model.status != "All changes saved" {
		t.Fatal("save did not commit settings")
	}
	click("Next")
	if model.demo.overview.page != 2 || !uitest.HasSemanticLabel(harness.Semantics(), "Desktop workspace") {
		t.Fatal("pagination did not show the second page")
	}
	click("Desktop workspace")
	if !slices.Contains(model.demo.overview.selected, "project-5") ||
		!slices.Contains(model.demo.overview.selected, "project-1") {
		t.Fatalf("selection did not persist across pages: %v", model.demo.overview.selected)
	}
	click("New project")
	edit("overview-new-name", "", "My new app")
	click("Create project")
	if len(model.demo.overview.projects) != 13 || model.demo.overview.creating ||
		model.demo.overview.page != 4 || !uitest.HasSemanticLabel(harness.Semantics(), "My new app") {
		t.Fatal("create did not add the project and navigate to its page")
	}
	click("Reset demos")
	if len(model.demo.overview.projects) != 12 || model.demo.overview.page != 1 ||
		model.demo.overview.settings.name != "FlowUI Desktop" {
		t.Fatal("reset did not restore the overview")
	}
}
