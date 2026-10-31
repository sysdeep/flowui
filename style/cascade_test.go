package style

import (
	"testing"

	"gioui.org/unit"
)

func TestCascadeReplacesShadowsAndOutline(t *testing.T) {
	shadowColor := RGBA(0x33445566)
	focusColor := RGB(0x00ccff)
	base := Style{}.
		BoxShadow(0, 2, 4, 0, shadowColor).
		Outline(1, 2, focusColor).
		VisualOutset(1, 2, 3, 4)

	override := Style{}.
		BoxShadow(0, 6, 12, 1, RGBA(0x11223344)).
		VisualOutset(5, 6, 7, 8)

	resolved := Cascade(State{}, base, override)
	if resolved.Paint == nil || len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("shadows = %#v, want one overriding layer", resolved.Paint)
	}
	if got := resolved.Paint.Shadows[0]; got.OffsetY != 6 || got.Blur != 12 || got.Color != RGBA(0x11223344) {
		t.Fatalf("shadow = %#v", got)
	}
	if resolved.Paint.Outline == nil || resolved.Paint.Outline.Width != 1 || resolved.Paint.Outline.Color != focusColor {
		t.Fatalf("outline = %#v", resolved.Paint.Outline)
	}
	if resolved.Paint.VisualOutset == nil || *resolved.Paint.VisualOutset != (Insets{Top: 5, Right: 6, Bottom: 7, Left: 8}) {
		t.Fatalf("visual outset = %#v", resolved.Paint.VisualOutset)
	}
}

func TestResolveAppliesOnWithoutMutatingSource(t *testing.T) {
	normal := RGB(0x0066ff)
	hover := RGB(0x3388ff)
	source := Style{}.
		Background(normal).
		On(Hovered, Style{}.Background(hover))

	resolved := source.Resolve(State{Hovered: true})
	if resolved.Paint == nil || resolved.Paint.Background != hover {
		t.Fatalf("resolved background = %#v, want %#v", resolved.Paint.Background, hover)
	}
	normalResolved := source.Resolve(State{})
	if normalResolved.Paint == nil || normalResolved.Paint.Background != normal {
		t.Fatalf("source background mutated: %#v", normalResolved.Paint)
	}
}

func TestCascadeCanClearBackgroundAndShadows(t *testing.T) {
	base := Style{}.
		Background(RGB(0x112233)).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000044))

	resolved := Cascade(State{}, base, Style{}.BackgroundNone().BoxShadowNone().VisualOutsetNone())

	if resolved.Paint == nil || resolved.Paint.Background != nil || len(resolved.Paint.Shadows) != 0 || resolved.Paint.VisualOutset != nil {
		t.Fatalf("cleared paint = %#v", resolved.Paint)
	}
}

func TestResolveAppliesDistinctComponentStates(t *testing.T) {
	source := Style{}.
		On(FocusVisible, Style{}.Opacity(.9)).
		On(Checked, Style{}.Scale(.8)).
		On(Indeterminate, Style{}.Radius(3)).
		On(ReadOnly, Style{}.BorderWidth(2)).
		On(Open, Style{}.Padding(4)).
		On(Expanded, Style{}.Margin(5)).
		On(Dragging, Style{}.Translate(6, 7)).
		On(DropTarget, Style{}.Background(RGB(0x010203)))

	resolved := source.Resolve(State{
		FocusVisible: true, Checked: true, Indeterminate: true, ReadOnly: true,
		Open: true, Expanded: true, Dragging: true, DropTarget: true,
	})
	if resolved.Paint == nil || *resolved.Paint.Opacity != .9 || *resolved.Paint.Radius != 3 || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("paint state overrides = %#v", resolved.Paint)
	}
	if resolved.Trans == nil || *resolved.Trans.ScaleX != .8 || *resolved.Trans.TranslateX != 6 {
		t.Fatalf("transform state overrides = %#v", resolved.Trans)
	}
	if resolved.Box == nil || resolved.Box.Padding.Top != 4 || resolved.Box.Margin.Top != 5 {
		t.Fatalf("box state overrides = %#v", resolved.Box)
	}
}

func TestCascadePartInheritsTextWithoutRootPaint(t *testing.T) {
	base := Style{}.
		TextColor(RGB(0x111111)).
		Background(RGB(0xeeeeee)).
		Part(PartLabel, Style{}.FontWeight(600))

	custom := Style{}.
		TextColor(RGB(0x222222)).
		Part(PartLabel, Style{}.
			On(Hovered, Style{}.TextColor(RGB(0x333333))))

	resolved := CascadePart(State{Hovered: true}, PartLabel, base, custom)
	if resolved.Paint != nil {
		t.Fatalf("root paint leaked into label: %#v", resolved.Paint)
	}
	if resolved.Text == nil || resolved.Text.Color != RGB(0x333333) || *resolved.Text.FontWeight != 600 {
		t.Fatalf("label style = %#v", resolved.Text)
	}
}

func TestCascadePartStateTextOverridesRootText(t *testing.T) {
	root := Style{}.TextColor(RGB(0x111111))
	part := Style{}.Part(PartLabel, Style{}.
		On(Selected, Style{}.TextColor(RGB(0xffffff))))

	resolved := CascadePart(State{Selected: true}, PartLabel, root, part)
	if resolved.Text == nil || resolved.Text.Color != RGB(0xffffff) {
		t.Fatalf("selected part text = %#v, want white", resolved.Text)
	}
}

func TestCascadeUsesLaterSetPropertiesOnly(t *testing.T) {
	base := Style{}.
		Padding(unit.Dp(8)).
		Background(RGB(0x111111)).
		Opacity(1)

	override := Style{}.
		Padding(8, 12).
		Opacity(0)

	resolved := Cascade(State{}, base, override)
	if resolved.Box == nil || resolved.Box.Padding == nil {
		t.Fatal("cascade lost padding")
	}
	if resolved.Box.Padding.Top != unit.Dp(8) || resolved.Box.Padding.Left != unit.Dp(12) {
		t.Fatalf("padding = %#v, want later shorthand top 8 and left 12", *resolved.Box.Padding)
	}
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0 {
		t.Fatalf("opacity = %#v, want explicit zero", resolved.Paint.Opacity)
	}
}

func TestCascadeMergesIndividualCornerRadii(t *testing.T) {
	base := Style{}.Radius(8)
	override := Style{}.RadiusTopLeft(1).RadiusBottomRight(3)
	resolved := Cascade(State{}, base, override)

	want := CornerRadii{TopLeft: 1, TopRight: 8, BottomRight: 3, BottomLeft: 8}
	if resolved.Paint == nil || resolved.Paint.Radii == nil || *resolved.Paint.Radii != want {
		t.Fatalf("radii = %#v, want %#v", resolved.Paint, want)
	}

	reset := Cascade(State{}, base, override, Style{}.Radius(4))
	if reset.Paint.Radius == nil || *reset.Paint.Radius != 4 || reset.Paint.Radii != nil {
		t.Fatalf("uniform radius did not replace corners: %#v", reset.Paint)
	}
}

func TestColorSourceWorksAcrossPaintAndText(t *testing.T) {
	accent := RGB(0x0066ff)
	foreground := RGB(0x101418)
	base := Style{}.BorderColor(accent).TextColor(foreground)
	resolved := Cascade(State{}, base, Style{}.BorderWidth(2))

	if resolved.Text == nil || resolved.Text.Color != foreground {
		t.Fatalf("text color = %#v, want %#v", resolved.Text, foreground)
	}
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Color != accent {
		t.Fatalf("border color = %#v, want %#v", resolved.Paint, accent)
	}
	if resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("border width = %#v, want 2", resolved.Paint.Border.Width)
	}
}
