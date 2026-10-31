package menubar

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestMenubarVisualCacheKeys(t *testing.T) {
	active := new(theme.Light())

	ctx := core.New(nil, active, locale.LanguageEnglish)
	cache := &menubarDefaultsCache{}
	value := Widget{}
	value.defaultsWithCache(ctx, cache)
	value.defaultsWithCache(ctx, cache)
	// The instance style is the last cascade layer, not a cache input.
	value.Style(style.Style{}.Padding(2)).defaultsWithCache(ctx, cache)
	value.compact = true
	value.defaultsWithCache(ctx, cache)
	replacement := *active
	core.ReplaceTheme(ctx, replacement)
	value.defaultsWithCache(ctx, cache)
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 3 || stats.ThemeInvalidations != 1 {
		t.Fatalf("cache stats = %#v", stats)
	}
}

func TestMenubarTriggerDefaultsFollowTheme(t *testing.T) {
	active := new(theme.Light())
	defaults := defaultMenubar(active, false)
	sem := active.Semantic()

	hovered := style.Cascade(style.State{Hovered: true}, defaults.trigger)
	background, ok := hovered.Paint.Background.(style.SolidColor)
	if !ok || background.Color != sem.Tone(theme.ToneDefault).Fill {
		t.Fatalf("hovered trigger background = %#v", hovered.Paint.Background)
	}
	if hovered.Text == nil {
		t.Fatal("trigger has no text declaration")
	}
	foreground, ok := hovered.Text.Color.(style.SolidColor)
	if !ok || foreground.Color != sem.Foreground {
		t.Fatalf("trigger foreground = %#v", hovered.Text.Color)
	}

	// The instance style is resolved after the cached defaults, so it wins
	// over the theme-derived state rules.
	customFill := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	custom := style.Cascade(
		style.State{Hovered: true},
		defaults.trigger,
		style.Style{}.Background(style.Solid(customFill)),
	)
	customBackground, ok := custom.Paint.Background.(style.SolidColor)
	if !ok || customBackground.Color != customFill {
		t.Fatalf("instance trigger background = %#v", custom.Paint.Background)
	}

	disabled := style.Cascade(style.State{Disabled: true}, defaults.trigger)
	if disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != active.DisabledOpacityValue() {
		t.Fatalf("disabled trigger opacity = %#v", disabled.Paint.Opacity)
	}
}
