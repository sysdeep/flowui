package colorpicker

import (
	"image/color"
	"testing"
)

func TestColorFieldFillWidth(t *testing.T) {
	base := ColorField(color.NRGBA{}).Key("color")
	configured := base.FillWidth()
	if base.fullWidth {
		t.Fatal("FillWidth mutated the original widget")
	}
	if !configured.fullWidth {
		t.Fatal("FillWidth did not enable full width")
	}
}
