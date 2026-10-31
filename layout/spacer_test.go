package layout

import (
	"image"
	"testing"

	"gioui.org/layout"
)

func TestSpacer(t *testing.T) {
	dims := Spacer(12, 8).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 300)},
	})
	if dims.Size != image.Pt(12, 8) {
		t.Fatalf("spacer size = %v, want (12,8)", dims.Size)
	}
}
