package toggle

import (
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

func TestSwitchDefaultThumbAndTrackStates(t *testing.T) {
	for name, active := range map[string]theme.Theme{"light": theme.Light(), "dark": theme.Dark()} {
		t.Run(name, func(t *testing.T) {
			defaults := buildSwitchDefaults(&active, SwitchMedium)
			for _, state := range []style.State{
				{}, {Checked: true}, {Hovered: true}, {Checked: true, Pressed: true},
				{Checked: true, Invalid: true}, {Checked: true, Disabled: true},
			} {
				thumb := style.CascadePart(state, style.PartThumb, defaults.Defaults)
				foreground := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				if state.Disabled {
					foreground = active.DisabledColor(foreground)
				}
				white := style.Solid(foreground)
				if thumb.Paint == nil || thumb.Paint.Background != white || len(thumb.Paint.Shadows) != 1 {
					t.Fatalf("state %+v: unexpected thumb appearance: %#v", state, thumb.Paint)
				}
				shadow := thumb.Paint.Shadows[0]
				if shadow.OffsetY != 1 || shadow.Blur != 2 || shadow.Spread != 0 {
					t.Fatalf("state %+v: unexpected thumb shadow: %#v", state, shadow)
				}
				track := style.CascadePart(state, style.PartTrack, defaults.Defaults)
				if track.Paint == nil || track.Paint.Border == nil {
					t.Fatalf("state %+v: missing track border", state)
				}
				border, ok := resolve.Color(track.Paint.Border.Color)
				want := active.Semantic().Separator
				if state.Checked {
					want = color.NRGBA{}
				}
				if !ok || border != want {
					t.Fatalf("state %+v: track border = %v, want %v", state, border, want)
				}
			}
		})
	}
}

func TestSwitchVisualCacheTracksSizeAndThemeInvalidation(t *testing.T) {
	activeTheme := new(theme.Light())
	var cache switchVisualCache

	cache.resolve(activeTheme, SwitchMedium)
	cache.resolve(activeTheme, SwitchMedium)
	cache.resolve(activeTheme, SwitchLarge)

	stats := cache.stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 0 {
		t.Fatalf("warm cache stats = %#v", stats)
	}

	replaced := *activeTheme
	cache.resolve(&replaced, SwitchMedium)
	stats = cache.stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.ThemeInvalidations != 1 {
		t.Fatalf("theme invalidation stats = %#v", stats)
	}
}

func TestSwitchVisualCachePanicDoesNotDiscardPreviousValue(t *testing.T) {
	active := new(theme.Light())
	var cache switchVisualCache

	cache.resolve(active, SwitchMedium)
	func() {
		defer func() {
			if recovered := recover(); recovered != theme.ErrMissingTheme {
				t.Fatalf("nil theme panic = %v, want %v", recovered, theme.ErrMissingTheme)
			}
		}()
		cache.resolve(nil, SwitchMedium)
	}()

	cache.resolve(active, SwitchMedium)
	stats := cache.stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("panic cache stats = %#v, want one hit, one miss, one theme invalidation", stats)
	}
}

func TestSwitchInstanceStyleResolvesAsLastLayer(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	widget := NewSwitch("").Key("styled").Checked(false).Style(style.Style{}.
		Part(style.PartTrack, style.Style{}.Width(30)).
		Part(style.PartThumb, style.Style{}.Width(10)))

	resolved := widget.resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "styled", style.State{}, nil)
	if resolved.trackOff.Box == nil || resolved.trackOff.Box.Width == nil || *resolved.trackOff.Box.Width != 30 {
		t.Fatalf("track width = %#v, want 30", resolved.trackOff.Box)
	}
	if resolved.thumbOff.Box == nil || resolved.thumbOff.Box.Width == nil || *resolved.thumbOff.Box.Width != 10 {
		t.Fatalf("thumb width = %#v, want 10", resolved.thumbOff.Box)
	}
}

func TestSwitchInstanceStyleDoesNotEnterDefaultsCache(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var cache switchVisualCache

	first := NewSwitch("").Key("first").Style(style.Style{}.
		Part(style.PartTrack, style.Style{}.Width(30)))
	first.resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "first", style.State{}, &cache)
	before := cache.stats()

	second := NewSwitch("").Key("second").Style(style.Style{}.
		Part(style.PartTrack, style.Style{}.Width(50)))
	second.resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "second", style.State{}, &cache)

	// A different instance style must still reuse the same theme defaults.
	after := cache.stats()
	if after.Hits != before.Hits+1 || after.Misses != before.Misses ||
		after.KeyInvalidations != before.KeyInvalidations || after.ThemeInvalidations != before.ThemeInvalidations {
		t.Fatalf("instance style changed defaults cache key: before=%#v after=%#v", before, after)
	}
}
