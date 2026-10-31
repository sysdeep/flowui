package input

import (
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestGroupActionDefaultsCacheTracksTheme(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	defaults := groupActionDefaultsFor(ctx, "clear")
	groupActionDefaultsFor(ctx, "clear")
	if defaults.slotPadding != 4 {
		t.Fatalf("slot padding = %v, want %v", defaults.slotPadding, 4)
	}
	cache, ok := core.PeekState[groupActionDefaultsCache](ctx, core.FullKey(ctx, "clear"), stateSlotInputActionDefaults)
	if !ok {
		t.Fatal("missing action defaults cache")
	}
	if stats := cache.Stats(); stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("stable cache stats = %+v", stats)
	}

	replacement := *active
	replacement.Colors.MutedForeground.R ^= 0xff
	core.ReplaceTheme(ctx, replacement)
	defaults = groupActionDefaultsFor(ctx, "clear")
	if defaults.slotPadding != 4 {
		t.Fatalf("replacement slot padding = %v, want 4", defaults.slotPadding)
	}
	if stats := cache.Stats(); stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 0 {
		t.Fatalf("invalidation stats = %+v", stats)
	}
}

// TestGroupActionInstanceStyleIsLastLayer pins that an action's own Style
// resolves over the theme defaults instead of entering the cached defaults.
func TestGroupActionInstanceStyleIsLastLayer(t *testing.T) {
	active := new(theme.Light())
	defaults := defaultGroupAction(active)
	resolved := style.Cascade(style.State{}, defaults.declaration, style.Style{}.Width(41))
	if width := resolved.Box.Width; width == nil || *width != 41 {
		t.Fatalf("instance width = %v, want 41", width)
	}
}
