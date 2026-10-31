package radiogroup

import (
	"testing"

	"github.com/qianniancn/flowui/theme"
)

func TestVisualCacheReusesVariantsAndInvalidatesTheme(t *testing.T) {
	active := new(theme.Light())
	var cache radioDefaultsCache

	resolveRadioDefaults(&cache, active, RadioPrimary)
	resolveRadioDefaults(&cache, active, RadioPrimary)
	resolveRadioDefaults(&cache, active, RadioSecondary)
	resolveRadioDefaults(&cache, active, RadioSecondary)
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 2 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %+v", stats)
	}

	replacement := *active
	resolveRadioDefaults(&cache, &replacement, RadioPrimary)
	if got := cache.Stats().ThemeInvalidations; got != 1 {
		t.Fatalf("theme invalidations = %d, want 1", got)
	}
}
