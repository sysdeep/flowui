package text

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestLabelVisualStates(t *testing.T) {
	active := theme.Light()
	foreground := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	visual := labelResolveLabelVisual(&active, foreground)

	textColor := func(state style.State) color.NRGBA {
		t.Helper()
		resolved := style.Cascade(state, visual.Defaults)
		if resolved.Text == nil {
			t.Fatalf("label declared no text properties for state %v", state)
		}
		value, ok := computed.Color(resolved.Text.Color)
		if !ok {
			t.Fatalf("label text color for state %v is not literal", state)
		}
		return value
	}

	if got := textColor(style.State{}); got != foreground {
		t.Fatalf("base label color = %#v, want %#v", got, foreground)
	}
	if got := textColor(style.State{Invalid: true}); got != active.Semantic().Tone(theme.ToneDanger).Fill {
		t.Fatalf("invalid label color = %#v", got)
	}
	if got := textColor(style.State{Disabled: true}); got != active.DisabledColor(foreground) {
		t.Fatalf("disabled label color = %#v", got)
	}
	if got := textColor(style.State{Disabled: true, Invalid: true}); got != active.DisabledColor(active.Semantic().Tone(theme.ToneDanger).Fill) {
		t.Fatalf("disabled invalid label color = %#v", got)
	}
}

func TestLabelInstanceStyleOnlyChangesRoot(t *testing.T) {
	ctx := textTestContext()
	base := NewLabel("Field name")
	styled := base.Style(style.Style{}.TextColor(style.RGB(0x123456)))
	baseVisual := base.visualWithCache(ctx)
	styledVisual := styled.visualWithCache(ctx)
	baseResolved := base.resolveStyleStatic(ctx, style.State{}, baseVisual)
	styledResolved := styled.resolveStyleStatic(ctx, style.State{}, styledVisual)

	got, ok := computed.Color(styledResolved.root.Text.Color)
	if !ok || got != style.RGB(0x123456).Color {
		t.Fatalf("instance label color = %#v", styledResolved.root.Text.Color)
	}
	if baseResolved.root.Text.FontSize == nil || styledResolved.root.Text.FontSize == nil ||
		*baseResolved.root.Text.FontSize != *styledResolved.root.Text.FontSize {
		t.Fatal("instance style changed theme typography")
	}
	if styledVisual.Defaults.Hash64() != baseVisual.Defaults.Hash64() ||
		styledVisual.Required.Hash64() != baseVisual.Required.Hash64() {
		t.Fatal("instance style changed cached label defaults")
	}
}

func TestLabelUsesThemeTokens(t *testing.T) {
	active := theme.Light()

	ctx := core.New(nil, &active, locale.LanguageEnglish)
	visual := NewLabel("Name").visualWithCache(ctx)
	if style.Cascade(style.State{}, visual.Defaults).Text == nil {
		t.Fatal("missing label text style")
	}
}
