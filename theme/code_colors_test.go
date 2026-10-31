package theme

import (
	"image/color"
	"testing"
)

func TestNewDerivesCodeColors(t *testing.T) {
	for _, active := range []Theme{New(DefaultSeed()), NewDark(DefaultSeed())} {
		code := active.CodeColors()
		if code.Background.A != 0xff {
			t.Fatalf("code background is not opaque: %d", code.Background.A)
		}
		if code.Background == active.Semantic().SurfaceTertiary {
			t.Fatal("code background is not tinted")
		}
		if code.Background == code.Plain {
			t.Fatal("code background matches the text color")
		}
		if code.Plain == code.Keyword {
			t.Fatal("plain text and keywords share a color")
		}
		if code.String == code.Comment {
			t.Fatal("strings and comments share a color")
		}
		if code.Plain.A != 0xff {
			t.Fatalf("plain text is not opaque: %d", code.Plain.A)
		}
	}
}

func TestCodeColorsCanBeOverridden(t *testing.T) {
	active := New(DefaultSeed())
	custom := CodeColors{Keyword: color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}}
	active.Code = custom
	if got := active.CodeColors().Keyword; got != custom.Keyword {
		t.Fatalf("keyword color = %v, want the override", got)
	}
	// Unset roles still come from the derived table.
	if got := active.CodeColors().String; got == (color.NRGBA{}) {
		t.Fatal("unset role was not derived")
	}
}
