package style

import (
	"image/color"
	"testing"
)

func TestStyleDeepCopiesGradientAndNestedRule(t *testing.T) {
	gradient := Gradient{
		Stops: []GradientStop{{Offset: 0, Color: SolidColor{Color: color.NRGBA{A: 0xff}}}},
	}
	source := Style{}.
		Background(gradient).
		On(Pressed, Style{}.Scale(0.95))

	gradient.Stops[0].Offset = 1
	storedSource := source.Resolve(State{})
	if storedSource.Paint == nil {
		t.Fatal("source lost paint")
	}
	stored := storedSource.Paint.Background.(Gradient)
	if stored.Stops[0].Offset != 0 {
		t.Fatalf("gradient was aliased: %v", stored.Stops[0].Offset)
	}

	resolved := source.Resolve(State{Pressed: true})
	if resolved.Trans == nil || resolved.Trans.ScaleX == nil || *resolved.Trans.ScaleX != 0.95 {
		t.Fatalf("pressed transform = %#v, want 0.95", resolved.Trans)
	}
}

func TestResolveDeepCopiesPointerSources(t *testing.T) {
	background := &SolidColor{Color: color.NRGBA{R: 1, A: 0xff}}
	source := Style{
		paint: &PaintStyle{Background: background},
	}
	built := source.Resolve(State{})

	background.Color.R = 2
	if got := built.Paint.Background.(*SolidColor).Color.R; got != 1 {
		t.Fatalf("background red = %d, want 1", got)
	}
}

func TestMutatingResolvedStyleDoesNotChangeDeclaration(t *testing.T) {
	declaration := Style{}.
		Padding(8).
		Background(RGB(0x112233)).
		Opacity(.75).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000040))

	resolved := declaration.Resolve(State{})
	resolved.Box.Padding.Left = 99
	*resolved.Paint.Opacity = 0
	resolved.Paint.Background = RGB(0xffffff)
	resolved.Paint.Shadows[0].Blur = 99

	again := declaration.Resolve(State{})
	if again.Box.Padding.Left != 8 || *again.Paint.Opacity != .75 || again.Paint.Background != RGB(0x112233) || again.Paint.Shadows[0].Blur != 4 {
		t.Fatalf("resolved mutation leaked into declaration: %#v", again)
	}
}

func TestFluentBranchesDoNotMutateTheirSource(t *testing.T) {
	base := Style{}.
		BorderColor(RGB(0x112233)).
		BorderWidth(1).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000040)).
		On(Hovered, Style{}.Opacity(.8)).
		Part(PartLabel, Style{}.TextColor(RGB(0x445566)))
	changed := base.
		BorderWidth(3).
		BoxShadow(0, 4, 8, 0, RGBA(0x00000080)).
		On(Pressed, Style{}.Scale(.95)).
		Part(PartLabel, Style{}.FontWeight(600))

	baseResolved := base.Resolve(State{Hovered: true, Pressed: true})
	if baseResolved.Paint == nil || *baseResolved.Paint.Border.Width != 1 || len(baseResolved.Paint.Shadows) != 1 || baseResolved.Trans != nil {
		t.Fatalf("source branch changed: %#v", baseResolved)
	}
	baseLabel := CascadePart(State{}, PartLabel, base)
	if baseLabel.Text == nil || baseLabel.Text.FontWeight != nil || baseLabel.Text.Color != RGB(0x445566) {
		t.Fatalf("source part changed: %#v", baseLabel.Text)
	}

	changedResolved := changed.Resolve(State{Hovered: true, Pressed: true})
	if *changedResolved.Paint.Border.Width != 3 || len(changedResolved.Paint.Shadows) != 2 || changedResolved.Trans == nil || *changedResolved.Trans.ScaleX != .95 {
		t.Fatalf("changed branch = %#v", changedResolved)
	}
	changedLabel := CascadePart(State{}, PartLabel, changed)
	if changedLabel.Text == nil || changedLabel.Text.FontWeight == nil || *changedLabel.Text.FontWeight != 600 {
		t.Fatalf("changed part = %#v", changedLabel.Text)
	}
}
