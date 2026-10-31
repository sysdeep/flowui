package style

import (
	"testing"
)

func TestCustomPartNamesAndStyleSnapshot(t *testing.T) {
	const thumb Part = "thumb"
	part := Style{}.Radius(4)
	built := Style{}.Part(thumb, part)
	part = part.Radius(8)
	_ = part

	resolved := CascadePart(State{}, thumb, built)
	if resolved.Paint == nil || resolved.Paint.Radius == nil || *resolved.Paint.Radius != 4 {
		t.Fatalf("custom part = %#v", resolved.Paint)
	}
}

func TestPartRootAppliesToRootDeclaration(t *testing.T) {
	resolved := Style{}.
		Part(PartRoot, Style{}.Padding(8).Background(RGB(0x123456))).
		Resolve(State{})
	if resolved.Box == nil || resolved.Box.Padding == nil || resolved.Box.Padding.Left != 8 {
		t.Fatalf("root box = %#v", resolved.Box)
	}
	if resolved.Paint == nil || resolved.Paint.Background != RGB(0x123456) {
		t.Fatalf("root paint = %#v", resolved.Paint)
	}
}
