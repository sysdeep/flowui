package theme_test

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/selector"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/toggle"
)

func TestDefaultControlAlignment(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		harness := uitest.New(image.Pt(300, 100), active)
		for _, control := range []ui.Widget{
			button.New("Save"), input.New("FlowUI"),
			selector.New(nil).Key("select"), combobox.New(nil).Key("combo"),
			datepicker.New("date", time.Time{}),
		} {
			harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				dims := control.Layout(ctx, gtx)
				if dims.Size.Y != 32 {
					t.Errorf("%T height = %d, want 32dp", control, dims.Size.Y)
				}
				return dims
			}))
		}
	}
}

func TestDefaultSliderPointerAndKeyboard(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		for _, vertical := range []bool{false, true} {
			harness := uitest.New(image.Pt(300, 300), active)
			value := 0.0
			root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				control := slider.New("slider", value).Step(5).
					OnChange(func(next float64) { value = next })
				if vertical {
					control = control.Vertical()
				}
				return control.Layout(ctx, gtx)
			})
			harness.Frame(root)
			position, arrow := f32.Pt(150, 8), key.NameRightArrow
			if vertical {
				position, arrow = f32.Pt(8, 150), key.NameUpArrow
			}
			harness.Click(position)
			harness.Frame(root)
			if value != 50 {
				t.Fatalf("vertical=%v pointer value = %v, want 50", vertical, value)
			}
			harness.Key(arrow, 0)
			harness.Frame(root)
			if value != 55 {
				t.Fatalf("vertical=%v keyboard value = %v, want 55", vertical, value)
			}
		}
	}
}

func TestDefaultSelectInteraction(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		harness := uitest.New(image.Pt(300, 300), active)
		selected, open, disabled := "desktop", false, false
		root := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return selector.New([]selector.SelectItem{
				{Key: "desktop", Label: "Desktop"}, {Key: "tools", Label: "Tools"},
			}).Key("platform").SelectedKey(selected).Open(open).Disabled(disabled).
				OnChange(func(value string) { selected = value }).
				OnOpenChange(func(value bool) { open = value }).Layout(ctx, gtx)
		})
		frame := func() {
			// Controlled open state arrives on the next frame, then its entrance
			// animation needs another frame before the options accept input.
			for range 3 {
				harness.Frame(root)
				harness.Advance(time.Second)
			}
		}
		frame()
		harness.Click(f32.Pt(30, 16))
		frame()
		if !open {
			t.Fatal("pointer did not open select")
		}
		harness.Key(key.NameDownArrow, 0)
		frame()
		harness.Key(key.NameReturn, 0)
		frame()
		if selected != "tools" || open {
			t.Fatalf("keyboard selection = %q, open=%v; want tools, false", selected, open)
		}
		harness.Click(f32.Pt(30, 16))
		frame()
		harness.Key(key.NameEscape, 0)
		frame()
		if open {
			t.Fatal("Escape did not close select")
		}
		disabled = true
		frame()
		harness.Click(f32.Pt(30, 16))
		frame()
		if open || selected != "tools" {
			t.Fatal("disabled select accepted input")
		}
	}
}

func TestControlSizeTiers(t *testing.T) {
	for _, test := range []struct {
		name       string
		buttonSize button.Size
		inputSize  input.Size
		toggleSize toggle.ToggleButtonSize
		height     int
	}{
		{name: "small", buttonSize: button.Small, inputSize: input.Small, toggleSize: toggle.ToggleButtonSmall, height: 24},
		{name: "medium", buttonSize: button.Medium, inputSize: input.Medium, toggleSize: toggle.ToggleButtonMedium, height: 32},
		{name: "large", buttonSize: button.Large, inputSize: input.Large, toggleSize: toggle.ToggleButtonLarge, height: 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
				harness := uitest.New(image.Pt(300, 100), active)
				for _, control := range []ui.Widget{
					button.New("保存 Save").Size(test.buttonSize),
					input.New("名称 Name").Size(test.inputSize),
					toggle.NewToggleButton("Toggle").Key("toggle").Size(test.toggleSize),
				} {
					harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
						gtx.Constraints.Min = image.Point{}
						dims := control.Layout(ctx, gtx)
						if dims.Size.Y != test.height {
							t.Errorf("%T height = %d, want %d", control, dims.Size.Y, test.height)
						}
						return dims
					}))
				}
			}
		})
	}
}
