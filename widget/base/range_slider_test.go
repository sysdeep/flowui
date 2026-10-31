package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/layout"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func rangeContent(base.RangeSliderState) ui.Widget {
	return layout.Spacer(200, 40)
}

func TestRangeSliderPointerSelectsNearestThumb(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	lower, upper := 20.0, 80.0
	changes := 0
	build := func() ui.Widget {
		return base.RangeSlider("price", lower, upper, 0, 100).
			Step(1).
			Label("Price").
			Track(rangeContent).
			OnRangeChange(func(nextLower, nextUpper float64) {
				lower, upper = nextLower, nextUpper
				changes++
			})
	}
	h.Frame(build())
	h.Click(f32.Pt(50, 20))
	h.Frame(build())
	if changes != 1 || lower == 20 || upper != 80 {
		t.Fatalf("left press: lower=%v upper=%v changes=%d", lower, upper, changes)
	}
	h.Click(f32.Pt(150, 20))
	h.Frame(build())
	if changes != 2 || upper == 80 || lower >= upper {
		t.Fatalf("right press: lower=%v upper=%v changes=%d", lower, upper, changes)
	}
}

func TestRangeSliderPointerDragMovesSelectedThumb(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	lower, upper := 20.0, 80.0
	build := func() ui.Widget {
		return base.RangeSlider("price", lower, upper, 0, 100).
			Step(1).
			Track(rangeContent).
			OnRangeChange(func(nextLower, nextUpper float64) { lower, upper = nextLower, nextUpper })
	}
	h.Frame(build())
	h.Touch(1, pointer.Press, f32.Pt(48, 20))
	h.Frame(build())
	h.Touch(1, pointer.Move, f32.Pt(96, 20))
	h.Frame(build())
	h.Touch(1, pointer.Release, f32.Pt(96, 20))
	h.Frame(build())
	if lower <= 20 || upper != 80 {
		t.Fatalf("dragged range: lower=%v upper=%v, want lower > 20 and upper 80", lower, upper)
	}
}

func TestRangeSliderKeyboardAndSemantics(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	lower, upper := 20.0, 80.0
	build := func() ui.Widget {
		return base.RangeSlider("price", lower, upper, 0, 100).
			Step(1).
			Label("Price").
			Track(rangeContent).
			OnRangeChange(func(nextLower, nextUpper float64) { lower, upper = nextLower, nextUpper })
	}
	h.Frame(build())
	tag := core.FieldFocusTag(h.Context(), "price")
	h.Router().Source().Execute(key.FocusCmd{Tag: tag})
	h.Frame(build())
	h.Key(key.NameRightArrow, 0)
	h.Frame(build())
	if lower != 21 || upper != 80 {
		t.Fatalf("keyboard: lower=%v upper=%v", lower, upper)
	}
	found := false
	uitest.WalkSemantics(h.Semantics(), func(node uitest.Semantic) bool {
		if node.Desc.Label == "Price" {
			found = true
			return false
		}
		return true
	})
	if !found {
		t.Fatalf("semantics = %v", uitest.SemanticLabels(h.Semantics()))
	}
}

func TestRangeSliderDisabledIgnoresInput(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	changes := 0
	view := base.RangeSlider("price", 20, 80, 0, 100).
		Track(rangeContent).
		Disabled(true).
		OnRangeChange(func(float64, float64) { changes++ })
	h.Frame(view)
	h.Click(f32.Pt(50, 20))
	h.Frame(view)
	if changes != 0 {
		t.Fatalf("disabled changes = %d", changes)
	}
}
