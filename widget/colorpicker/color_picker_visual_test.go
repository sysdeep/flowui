package colorpicker

import (
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func TestColorPickerComponentTreeLaysOutAndShowsOverlay(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	router := new(input.Router)
	now := time.Unix(1, 0)
	value := color.NRGBA{R: 4, G: 133, B: 247, A: 180}

	layoutColorWidgetFrame(ctx, router, ColorWheel(value).Key("renderer-wheel"), now)
	layoutColorWidgetFrame(ctx, router, ColorSlider(value, ColorChannelHue).Key("renderer-slider"), now)
	layoutColorWidgetFrame(ctx, router, ColorSwatch(value), now)
	layoutColorWidgetFrame(ctx, router, ColorSwatchPicker(value, []color.NRGBA{value}).Key("renderer-swatches"), now)

	picker := New(value).Key("renderer-picker").
		Label("Accent color").
		Alpha(true).
		Presets([]color.NRGBA{{R: 200, G: 100, B: 50, A: 255}})
	layoutColorPickerFrame(ctx, router, picker, now)
	state := colorPickerTestState(ctx, "renderer-picker")
	state.open = true
	layoutColorPickerFrame(ctx, router, picker, now.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, now.Add(colorPickerEnterDuration+time.Millisecond))

	if !core.HasVisibleOverlay(ctx) || !core.OverlayTopmost(ctx, core.OverlayLayerPopup, "renderer-picker") {
		t.Fatal("color picker did not register a visible topmost popup overlay")
	}
	state.open = false
	layoutColorPickerFrame(ctx, router, picker, now.Add(2*colorPickerEnterDuration+time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, now.Add(3*colorPickerEnterDuration+time.Millisecond))
	if core.HasVisibleOverlay(ctx) {
		t.Fatal("color picker overlay remained visible after close")
	}
}
