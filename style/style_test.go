package style

import (
	"testing"
)

func TestStyleIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		style Style
		want  bool
	}{
		{name: "zero value", style: Style{}, want: true},
		{name: "box declaration", style: Style{}.MinHeight(1), want: false},
		{name: "empty part declaration", style: Style{}.Part(PartItem, Style{}), want: false},
		{name: "conditional rule", style: Style{}.On(Hovered, Style{}), want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.style.IsEmpty(); got != test.want {
				t.Fatalf("Style.IsEmpty() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLineStyleDashPattern(t *testing.T) {
	if on, off := LineSolid.DashPattern(); on != 0 || off != 0 {
		t.Fatalf("LineSolid pattern = (%v, %v), want (0, 0)", on, off)
	}
	if on, off := LineDashed.DashPattern(); on <= 0 || off <= 0 {
		t.Fatalf("LineDashed pattern = (%v, %v), want positive lengths", on, off)
	}
	if on, off := LineDotted.DashPattern(); on <= 0 || off <= 0 {
		t.Fatalf("LineDotted pattern = (%v, %v), want positive lengths", on, off)
	}
}
