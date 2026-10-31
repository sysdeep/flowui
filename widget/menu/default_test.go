package menu

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// TestMenuHighlightText checks that a highlighted item keeps its label readable,
// and that the destructive item uses white text on the red fill.
func TestMenuHighlightText(t *testing.T) {
	light := theme.Light()
	dark := theme.Dark()
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for _, tc := range []struct {
		name   string
		active *theme.Theme
	}{
		{"light", &light},
		{"dark", &dark},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defaults := buildMenu(tc.active, false)
			sem := tc.active.Semantic()
			for _, test := range []struct {
				name    string
				variant int
				fill    color.NRGBA
			}{
				{"default", 0, sem.Tone(theme.TonePrimary).Fill},
				{"danger", 1, sem.Tone(theme.ToneDanger).Fill},
			} {
				item := defaults.Items[test.variant][0]
				part := style.CascadePart(style.State{Hovered: true}, style.PartItem, item.Defaults, item.Variant)
				background, ok := part.Paint.Background.(style.SolidColor)
				if !ok {
					t.Fatalf("%s fill = %#v", test.name, part.Paint.Background)
				}
				if background.Color != test.fill {
					t.Fatalf("%s fill = %v, want %v", test.name, background.Color, test.fill)
				}
				text := item.resolvedColors(part).Foreground
				if text != white {
					t.Fatalf("%s highlight text = %v, want white", test.name, text)
				}
				if ratio := theme.ContrastRatio(test.fill, text); ratio < 3 {
					t.Fatalf("%s contrast = %.2f", test.name, ratio)
				}
			}
		})
	}
}

func TestMenuVisualCacheTracksThemeAndInstanceStyle(t *testing.T) {
	light := theme.Light()
	dark := theme.Dark()
	ctx := core.New(nil, &light, locale.LanguageEnglish)
	var cache menuDefaultsCache
	value := Widget{}
	first := value.visualWithCache(ctx, &cache)
	value.visualWithCache(ctx, &cache)
	if stats := cache.Stats(); stats.Misses != 1 || stats.Hits != 1 {
		t.Fatalf("stable cache stats = %+v", stats)
	}
	// The instance style is the last cascade layer and is not a cache input.
	custom := value.Style(style.Style{}.Width(234)).visualWithCache(ctx, &cache)
	if custom.panel.Box == nil || custom.panel.Box.Width == nil || *custom.panel.Box.Width != 234 {
		t.Fatalf("instance style = %#v", custom.panel.Box)
	}
	if stats := cache.Stats(); stats.Misses != 1 || stats.KeyInvalidations != 0 {
		t.Fatalf("instance style invalidated the defaults cache: %+v", stats)
	}
	value = value.Compact(true)
	value.visualWithCache(ctx, &cache)
	if stats := cache.Stats(); stats.Misses != 2 || stats.KeyInvalidations != 1 {
		t.Fatalf("density invalidation stats = %+v", stats)
	}
	core.ReplaceTheme(ctx, dark)
	last := value.visualWithCache(ctx, &cache)
	if stats := cache.Stats(); stats.ThemeInvalidations != 1 {
		t.Fatalf("theme invalidation stats = %+v", stats)
	}
	if first.defaults.Colors.PanelBackground == last.defaults.Colors.PanelBackground {
		t.Fatal("theme replacement retained light menu colors")
	}
}
