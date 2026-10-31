package style

import (
	"math"
	"testing"

	"gioui.org/unit"
)

func TestTransformIgnoresNonFiniteValues(t *testing.T) {
	base := Style{}.Translate(2, 3).Scale(.5).Rotate(.7)
	for _, value := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		resolved := base.
			Translate(unit.Dp(value), 4).
			Translate(4, unit.Dp(value)).
			Scale(value).
			Rotate(value).
			Resolve(State{})
		if *resolved.Trans.TranslateX != 2 || *resolved.Trans.TranslateY != 3 ||
			*resolved.Trans.ScaleX != .5 || *resolved.Trans.ScaleY != .5 || *resolved.Trans.Rotate != .7 {
			t.Fatalf("non-finite %v changed transform: %#v", value, resolved.Trans)
		}
	}
}
