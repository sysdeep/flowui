package navigation

import (
	"testing"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestPaginationVisualCacheTracksSizeAndTheme(t *testing.T) {
	first := theme.Light()

	second := first
	firstContext := core.New(nil, &first, locale.LanguageEnglish)
	secondContext := core.New(nil, &second, locale.LanguageEnglish)
	value := NewPagination("cache", 1, 5)
	value.visualWithCache(firstContext)
	value.visualWithCache(firstContext)
	value.Size(PaginationSizeSmall).visualWithCache(firstContext)
	value.visualWithCache(secondContext)
	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.ThemeInvalidations != 1 {
		t.Fatalf("pagination cache stats = %#v", stats)
	}
}

func TestPaginationInstanceStyleAppliesLast(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	value := NewPagination("style", 1, 5).Style(style.Style{}.Padding(3))
	visual := value.visualWithCache(ctx)
	resolved := style.Cascade(style.State{}, visual.Defaults, value.style)
	if resolved.Box == nil || resolved.Box.Padding == nil {
		t.Fatal("pagination instance style did not reach the root declaration")
	}
	if got := resolved.Box.Padding.Left; got != unit.Dp(3) {
		t.Fatalf("padding = %v, want 3", got)
	}
}
