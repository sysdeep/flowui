package alertdialog

import (
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestAlertDialogDefaultsUseThemeTokens(t *testing.T) {
	active := theme.Light()

	ctx := core.New(nil, &active, locale.LanguageAuto)
	value := New("Title", "Body")
	defaults := value.defaultsWithCache(ctx, value.status)
	if len(defaults.icon.glyph) == 0 {
		t.Fatal("alert dialog defaults are missing the status glyph")
	}
	if defaults.icon.background.A == 0 || defaults.icon.foreground.A == 0 {
		t.Fatalf("status icon colors = %#v/%#v, want non-zero", defaults.icon.foreground, defaults.icon.background)
	}
	if defaults.metrics.IconSize <= 0 || defaults.metrics.IconGlyphSize <= 0 {
		t.Fatalf("icon metrics = %#v, want positive", defaults.metrics)
	}
	resolved := style.Cascade(style.State{}, defaults.header)
	if resolved.Text == nil || resolved.Text.FontSize == nil {
		t.Fatalf("header declaration is incomplete: %#v", resolved.Text)
	}
}

func TestAlertDialogStyleIsLastCascadeLayer(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageAuto)
	want := style.RGB(0x123456)
	value := New("Title", "Body").Style(style.New().TextColor(want))
	defaults := value.defaultsWithCache(ctx, value.status)

	resolved := style.Cascade(style.State{}, defaults.header, value.style)
	if resolved.Text == nil || resolved.Text.Color != want {
		t.Fatalf("header text color = %#v, want the instance declaration", resolved.Text)
	}

	plain := New("Title", "Body").defaultsWithCache(ctx, value.status)
	if defaults.header.Hash64() != plain.header.Hash64() {
		t.Fatal("instance style leaked into the cached theme defaults")
	}
}
