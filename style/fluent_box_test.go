package style

import (
	"math"
	"reflect"
	"testing"

	"gioui.org/unit"
)

func TestLayoutValuesIgnoreNonFiniteValues(t *testing.T) {
	base := Style{}.
		Width(100).Height(40).
		MinWidth(80).MaxWidth(120).MinHeight(20).MaxHeight(60).
		Padding(4).Margin(5).
		BorderWidth(1).Radius(6).
		BoxShadow(1, 2, 3, 4, RGB(0x112233)).
		Outline(2, 3, RGB(0x445566)).
		Opacity(.75).
		FontSize(14).LineHeight(20).LineHeightScale(1.25)

	for _, value := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		length := unit.Dp(value)
		textLength := unit.Sp(value)
		resolved := base.
			Width(length).Height(length).
			MinWidth(length).MaxWidth(length).MinHeight(length).MaxHeight(length).
			Padding(length).
			Margin(length).MarginX(length).MarginY(length).
			MarginTop(length).MarginRight(length).MarginBottom(length).MarginLeft(length).
			BorderWidth(length).
			Radius(length).
			RadiusTopLeft(length).RadiusTopRight(length).RadiusBottomRight(length).RadiusBottomLeft(length).
			BoxShadow(length, 2, 3, 4, RGB(0)).
			BoxShadow(1, length, 3, 4, RGB(0)).
			BoxShadow(1, 2, length, 4, RGB(0)).
			BoxShadow(1, 2, 3, length, RGB(0)).
			Outline(length, 2, RGB(0)).Outline(1, length, RGB(0)).
			Opacity(value).
			FontSize(textLength).LineHeight(textLength).LineHeightScale(value).
			Resolve(State{})

		want := base.Resolve(State{})
		if !reflect.DeepEqual(resolved, want) {
			t.Fatalf("non-finite %v changed layout style:\n got %#v\nwant %#v", value, resolved, want)
		}
	}
}

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
