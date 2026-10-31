package theme

import (
	"image/color"
	"testing"
)

func TestSeedCustomizesDefaultAppearance(t *testing.T) {
	primary := color.NRGBA{R: 180, G: 50, B: 80, A: 255}
	seed := DefaultSeed()
	seed.Primary = primary
	seed.Motion.Enabled = false
	for _, active := range []Theme{New(seed), NewDark(seed)} {
		if got := active.Semantic().Tone(TonePrimary).Fill; got != primary {
			t.Fatalf("primary = %v, want %v", got, primary)
		}
		if active.Motion.Enabled {
			t.Fatal("seed did not disable motion")
		}
	}
}

func TestDefaultModes(t *testing.T) {
	for _, test := range []struct {
		name  string
		new   func() Theme
		light bool
	}{
		{"light", Light, true},
		{"dark", Dark, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			active := test.new()
			semantic := active.Semantic()
			bg, fg := semantic.Background, semantic.Foreground
			contrast := int(bg.R) + int(bg.G) + int(bg.B) - int(fg.R) - int(fg.G) - int(fg.B)
			if contrast == 0 || (contrast > 0) != test.light {
				t.Fatalf("%s mode has background %v and foreground %v", test.name, bg, fg)
			}
		})
	}
}
