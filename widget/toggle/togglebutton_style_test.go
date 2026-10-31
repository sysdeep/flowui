package toggle

import (
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestVisualCacheTracksRequestKeyAndThemeInvalidation(t *testing.T) {
	active := new(theme.Light())
	var cache toggleButtonVisualCache
	key := toggleButtonVisualKey{
		variant: ToggleButtonDefault,
		size:    ToggleButtonMedium,
		foreground: color.NRGBA{
			R: 0x10, G: 0x20, B: 0x30, A: 0xff,
		},
	}

	cache.resolve(active, key)
	cache.resolve(active, key)
	key.size = ToggleButtonLarge
	cache.resolve(active, key)
	key.iconOnly = true
	cache.resolve(active, key)
	key.foreground.R++
	cache.resolve(active, key)

	stats := cache.stats()
	if stats.Hits != 1 || stats.Misses != 4 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 3 {
		t.Fatalf("warm cache stats = %#v", stats)
	}

	replacement := *active
	cache.resolve(&replacement, key)
	stats = cache.stats()
	if stats.Misses != 5 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 3 {
		t.Fatalf("theme invalidation stats = %#v", stats)
	}
}

func TestToggleInstanceStyleResolvesAsLastLayer(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var cache toggleButtonVisualCache
	widget := ToggleButtonWidget{}.Style(style.Style{}.Height(91))

	resolved := widget.resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "toggle", style.State{}, &cache)
	if resolved.root.Box == nil || resolved.root.Box.Height == nil || *resolved.root.Box.Height != 91 {
		t.Fatalf("visual declaration did not receive instance style: %#v", resolved.root)
	}
	before := cache.stats()

	other := ToggleButtonWidget{}.Style(style.Style{}.Width(120))
	other.resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "other", style.State{}, &cache)
	after := cache.stats()
	// A different instance style must still reuse the same theme defaults.
	if after.Hits != before.Hits+1 || after.Misses != before.Misses ||
		after.KeyInvalidations != before.KeyInvalidations || after.ThemeInvalidations != before.ThemeInvalidations {
		t.Fatalf("instance style changed defaults cache key: before=%#v after=%#v", before, after)
	}
}
