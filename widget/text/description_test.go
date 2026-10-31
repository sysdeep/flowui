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

func TestDescriptionVisualUsesMutedForeground(t *testing.T) {
	active := theme.Light()
	visual := buildDescriptionVisual(&active)

	text := func(state style.State) color.NRGBA {
		t.Helper()
		resolved := style.Cascade(state, visual.Defaults)
		if resolved.Text == nil {
			t.Fatalf("description declared no text properties for state %v", state)
		}
		value, ok := computed.Color(resolved.Text.Color)
		if !ok {
			t.Fatalf("description text color for state %v is not a literal color", state)
		}
		return value
	}

	muted := active.Semantic().MutedForeground
	if got := text(style.State{}); got != muted {
		t.Fatalf("description did not use muted foreground: got %#v", got)
	}
	if got := text(style.State{Disabled: true}); got != active.DisabledColor(muted) {
		t.Fatalf("disabled description color was not applied: got %#v", got)
	}
}

func TestDescriptionUsesThemeTokens(t *testing.T) {
	active := theme.Light()

	ctx := core.New(nil, &active, locale.LanguageEnglish)
	visual := NewDescription("Help").visualWithCache(ctx)
	resolved := style.Cascade(style.State{}, visual.Defaults)
	if resolved.Text == nil || resolved.Text.FontSize == nil {
		t.Fatal("missing description text style")
	}
}
