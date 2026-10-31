package fieldstyle

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestSurfaceColorsByFieldFace(t *testing.T) {
	for _, testTheme := range []struct {
		name   string
		active theme.Theme
	}{
		{name: "light", active: theme.Light()},
		{name: "dark", active: theme.Dark()},
	} {
		t.Run(testTheme.name, func(t *testing.T) {
			active := testTheme.active
			semantic := active.Semantic()
			neutral := semantic.Tone(theme.ToneDefault)
			for _, test := range []struct {
				name      string
				face      Variant
				bg, hover color.NRGBA
			}{
				{name: "primary", face: Primary, bg: semantic.FieldBackground, hover: semantic.FieldHover},
				{name: "secondary", face: Secondary, bg: neutral.Fill, hover: neutral.Hover},
				{name: "filled", face: Filled, bg: neutral.Fill, hover: neutral.Hover},
				{name: "borderless", face: Borderless},
				{name: "underlined", face: Underlined, bg: semantic.FieldBackground, hover: semantic.FieldBackground},
			} {
				t.Run(test.name, func(t *testing.T) {
					bg, hover := surfaceColors(&active, test.face)
					if got := solidColor(bg); got != test.bg {
						t.Errorf("background = %v, want %v", got, test.bg)
					}
					if got := solidColor(hover); got != test.hover {
						t.Errorf("hover = %v, want %v", got, test.hover)
					}
				})
			}
		})
	}
}

func solidColor(source style.ColorSource) color.NRGBA {
	if value, ok := source.(style.SolidColor); ok {
		return value.Color
	}
	return color.NRGBA{}
}
