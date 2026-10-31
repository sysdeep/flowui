package style

import (
	"testing"
)

func TestBottomBorderFluentStyle(t *testing.T) {
	accent := RGB(0x1677ff)
	declaration := Style{}.BorderBottom(1, accent)
	resolved := declaration.Resolve(State{})
	if resolved.Paint == nil || resolved.Paint.BorderBottom == nil {
		t.Fatalf("bottom border = %#v", resolved.Paint)
	}
	if resolved.Paint.BorderBottom.Color != accent || resolved.Paint.BorderBottom.Width == nil || *resolved.Paint.BorderBottom.Width != 1 {
		t.Fatalf("bottom border = %#v", resolved.Paint.BorderBottom)
	}

	changed := declaration.BorderBottom(2, RGB(0xff4d4f))
	base := declaration.Resolve(State{})
	if base.Paint.BorderBottom.Color != accent || *base.Paint.BorderBottom.Width != 1 {
		t.Fatalf("source bottom border changed: %#v", base.Paint.BorderBottom)
	}
	if changed.Hash64() == declaration.Hash64() {
		t.Fatal("bottom border changes were not included in style hash")
	}
}

func TestLineStyleIsHashedAndResolved(t *testing.T) {
	base := New().BorderWidth(1).BorderColor(RGB(0x000000))
	dashed := base.BorderLine(LineDashed)

	if base.Hash64() == dashed.Hash64() {
		t.Fatal("LineDashed hashed identically to an unset line style")
	}
	if dashed.Hash64() == base.BorderLine(LineSolid).Hash64() {
		t.Fatal("LineDashed hashed identically to LineSolid")
	}

	resolved := dashed.Resolve(State{})
	if resolved.Paint == nil || resolved.Paint.Border == nil {
		t.Fatal("border did not survive resolution")
	}
	if resolved.Paint.Border.LineStyle == nil || *resolved.Paint.Border.LineStyle != LineDashed {
		t.Fatal("LineDashed did not survive resolution")
	}
}

func TestLineStyleComposesWithStateRules(t *testing.T) {
	declaration := New().
		BorderWidth(1).
		BorderColor(RGB(0x111111)).
		On(Disabled, New().BorderLine(LineDashed))

	resolved := declaration.Resolve(State{Disabled: true})
	if resolved.Paint == nil || resolved.Paint.Border == nil {
		t.Fatal("border missing after state resolution")
	}
	if resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 1 {
		t.Fatal("state row dropped the base border width")
	}
	if resolved.Paint.Border.LineStyle == nil || *resolved.Paint.Border.LineStyle != LineDashed {
		t.Fatal("state row did not apply LineDashed")
	}
}
