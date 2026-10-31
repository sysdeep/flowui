package base_test

import (
	"image"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/layout"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestSliderKeysAccelerateAndProcessQueuedValues(t *testing.T) {
	h := uitest.New(image.Pt(100, 40), theme.Light())
	var tag struct{ _ byte }
	value, disabled := .5, false
	view := core.WidgetFunc(func(_ *core.Context, gtx giolayout.Context) giolayout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		keys := base.SliderKeys{Min: 0, Max: 1, Step: .01, ShiftScale: 10}
		for {
			next, ok := keys.Update(gtx, &tag, value)
			if !ok {
				break
			}
			value = next
		}
		core.AddPointerArea(gtx, &tag, image.Rect(0, 0, 100, 40), 0)
		return giolayout.Dimensions{Size: image.Pt(100, 40)}
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &tag})
	h.Key(key.NameRightArrow, key.ModShift)
	h.Key(key.NameRightArrow, 0)
	h.Frame(view)
	if math.Abs(value-.61) > 1e-9 {
		t.Fatalf("queued axis value = %v, want .61", value)
	}
	disabled = true
	h.Key(key.NameEnd, 0)
	h.Frame(view)
	if math.Abs(value-.61) > 1e-9 {
		t.Fatalf("disabled axis changed to %v", value)
	}
}

func sliderContent(base.SliderState) ui.Widget {
	return layout.Spacer(200, 40)
}

func TestSliderPointerAndKeyboardReportValues(t *testing.T) {
	h := uitest.New(image.Pt(200, 40), theme.Light())
	value := 0.0
	var got []float64
	build := func() ui.Widget {
		return base.Slider("volume", value, 0, 100).
			Step(1).
			Label("Volume").
			Track(sliderContent).
			OnChange(func(next float64) {
				value = next
				got = append(got, next)
			})
	}

	h.Frame(build())
	h.Click(f32.Pt(150, 20))
	h.Frame(build())
	h.Frame(build())
	if len(got) != 1 || math.Abs(got[0]-75) > 0.001 {
		t.Fatalf("click changes = %v, want 75", got)
	}

	got = nil
	h.Key(key.NameRightArrow, 0)
	h.Frame(build())
	if len(got) != 1 || math.Abs(got[0]-76) > 0.001 {
		t.Fatalf("Right changes = %v, want 76", got)
	}
}

func TestSliderHomeEndAndPageKeys(t *testing.T) {
	h := uitest.New(image.Pt(200, 40), theme.Light())
	value := 50.0
	var got []float64
	build := func() ui.Widget {
		return base.Slider("volume", value, 0, 100).
			Step(5).
			Track(sliderContent).
			OnChange(func(next float64) {
				value = next
				got = append(got, next)
			})
	}

	h.Frame(build())
	h.Click(f32.Pt(100, 20))
	h.Frame(build())
	h.Frame(build())
	got = nil
	h.Key(key.NameHome, 0)
	h.Frame(build())
	h.Key(key.NameEnd, 0)
	h.Frame(build())
	h.Key(key.NamePageDown, 0)
	h.Frame(build())
	if len(got) != 3 || got[0] != 0 || got[1] != 100 || got[2] != 50 {
		t.Fatalf("Home/End/PageDown = %v, want [0 100 50]", got)
	}
}

func TestSliderDisabledIgnoresInput(t *testing.T) {
	h := uitest.New(image.Pt(200, 40), theme.Light())
	var got []float64
	view := base.Slider("volume", 25, 0, 100).
		Step(1).
		Disabled(true).
		Track(sliderContent).
		OnChange(func(next float64) { got = append(got, next) })
	h.Frame(view)
	h.Click(f32.Pt(150, 20))
	h.Frame(view)
	if len(got) != 0 {
		t.Fatalf("changes = %v, want none", got)
	}
}

func TestSliderExposesLabel(t *testing.T) {
	h := uitest.New(image.Pt(200, 40), theme.Light())
	h.Frame(base.Slider("volume", 25, 0, 100).
		Label("Volume").
		Track(sliderContent))
	if _, ok := uitest.SemanticByLabel(h.Semantics(), "Volume"); !ok {
		t.Fatalf("semantics = %v, want Volume", uitest.SemanticLabels(h.Semantics()))
	}
}
