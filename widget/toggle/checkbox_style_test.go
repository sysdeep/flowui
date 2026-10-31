package toggle

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestCheckboxDisabledCheckedUsesDisabledSemanticColors(t *testing.T) {
	active := new(theme.Light())
	defaults := buildCheckboxDefaults(active, CheckboxPrimary)
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for _, state := range []style.State{
		{Checked: true, Selected: true, Disabled: true},
		{Checked: true, Selected: true, Disabled: true, Invalid: true},
	} {
		resolved := style.CascadePart(state, style.PartIndicator, defaults.Defaults)
		if resolved.Paint == nil || resolved.Text == nil {
			t.Fatalf("state %+v: disabled checked indicator is incomplete", state)
		}
		background, ok := computed.Brush(resolved.Paint.Background)
		if !ok {
			t.Fatalf("state %+v: background = %#v", state, resolved.Paint.Background)
		}
		if !state.Invalid && background.ColorAt(.5) != active.Semantic().DisabledBackground {
			t.Fatalf("state %+v: background = %#v, want disabled background", state, resolved.Paint.Background)
		}
		foreground, ok := computed.Color(resolved.Text.Color)
		if !ok || foreground != white {
			t.Fatalf("state %+v: foreground = %#v, want white selected foreground", state, foreground)
		}
	}
}

func TestCheckboxSelectedForegroundIsWhiteInBothThemes(t *testing.T) {
	for name, active := range map[string]theme.Theme{"light": theme.Light(), "dark": theme.Dark()} {
		t.Run(name, func(t *testing.T) {
			defaults := buildCheckboxDefaults(&active, CheckboxPrimary)
			resolved := style.CascadePart(style.State{Checked: true, Selected: true}, style.PartIndicator, defaults.Defaults)
			foreground, ok := computed.Color(resolved.Text.Color)
			want := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			if !ok || foreground != want {
				t.Fatalf("foreground = %#v, want white", foreground)
			}
		})
	}
}

func TestCheckboxIndicatorForegroundUsesResolvedOnStyle(t *testing.T) {
	got := checkboxIndicatorForeground()
	want := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if got != want {
		t.Fatalf("indicator foreground = %#v, want resolved on-state foreground %#v", got, want)
	}
}

func TestVisualCacheReusesVariantsAndInvalidatesTheme(t *testing.T) {
	active := new(theme.Light())
	var cache CheckboxVisualCache

	cache.resolve(active, CheckboxPrimary)
	cache.resolve(active, CheckboxPrimary)
	cache.resolve(active, CheckboxSecondary)
	cache.resolve(active, CheckboxSecondary)
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 2 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %+v", stats)
	}

	replacement := *active
	cache.resolve(&replacement, CheckboxPrimary)
	if got := cache.Stats().ThemeInvalidations; got != 1 {
		t.Fatalf("theme invalidations = %d, want 1", got)
	}
}
