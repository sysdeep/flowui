package table

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

func TestTableGridMeetsOuterBorder(t *testing.T) {
	for _, width := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("width-%d", width), func(t *testing.T) {
			size := image.Pt(80, 60)
			window, err := headless.NewWindow(size.X, size.Y)
			if err != nil {
				t.Skipf("headless renderer unavailable: %v", err)
			}
			defer window.Release()

			var ops op.Ops
			gtx := layout.Context{Ops: &ops}
			line := color.NRGBA{A: 255}
			paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			tableDrawHeaderSeparator(gtx, 40, 20, width, 20, line, true)
			offset := op.Offset(image.Pt(0, 20)).Push(&ops)
			tableDrawRowSeparators(gtx, 0, []int{40, 40}, 40, width, line)
			offset.Pop()
			tableDrawBorder(gtx, size, 6, width, line)
			if err := window.Frame(&ops); err != nil {
				t.Fatal(err)
			}
			pixels := image.NewRGBA(image.Rectangle{Max: size})
			if err := window.Screenshot(pixels); err != nil {
				t.Fatal(err)
			}
			// The grid junction and adjacent frame must share the outermost pixel.
			for _, y := range []int{0, size.Y - 1} {
				for x := 12; x < size.X-12; x++ {
					if got := pixels.RGBAAt(x, y); got != (color.RGBA{A: 255}) {
						t.Fatalf("frame pixel (%d, %d) = %v, want solid border", x, y, got)
					}
				}
			}
		})
	}
}
