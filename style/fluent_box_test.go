package style

import (
	"testing"

	"gioui.org/unit"
)

func TestPaddingShorthand(t *testing.T) {
	for _, test := range []struct {
		name   string
		values []unit.Dp
		want   Insets
	}{
		{name: "all", values: []unit.Dp{8}, want: Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}},
		{name: "vertical-horizontal", values: []unit.Dp{4, 10}, want: Insets{Top: 4, Right: 10, Bottom: 4, Left: 10}},
		{name: "top-horizontal-bottom", values: []unit.Dp{2, 6, 8}, want: Insets{Top: 2, Right: 6, Bottom: 8, Left: 6}},
		{name: "four-sides", values: []unit.Dp{1, 2, 3, 4}, want: Insets{Top: 1, Right: 2, Bottom: 3, Left: 4}},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolved := Style{}.Padding(test.values...).Resolve(State{})
			if resolved.Box == nil || resolved.Box.Padding == nil || *resolved.Box.Padding != test.want {
				t.Fatalf("padding = %#v, want %#v", resolved.Box, test.want)
			}
		})
	}
}
