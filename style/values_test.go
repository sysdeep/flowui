package style

import (
	"image/color"
	"testing"
)

func TestPackedColors(t *testing.T) {
	if got := RGB(0x9333ea).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xff}) {
		t.Fatalf("RGB = %#v", got)
	}
	if got := RGBA(0x9333eacc).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xcc}) {
		t.Fatalf("RGBA = %#v", got)
	}
}
