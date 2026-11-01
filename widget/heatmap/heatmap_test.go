package heatmap

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestHeatmapBuildsCalendarWeekGrid(t *testing.T) {
	start := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	widget := New(start, start.AddDate(0, 0, 6), []CalendarValue{{Date: start, Value: 4}}).Key("calendar")
	if len(widget.values) != 7 || len(widget.values[0]) != 2 {
		t.Fatalf("calendar grid = %d rows x %d columns", len(widget.values), len(widget.values[0]))
	}
	if widget.values[int(start.Weekday())][0] != 4 {
		t.Fatalf("calendar value = %v", widget.values[int(start.Weekday())][0])
	}
}

func TestHeatmapCalendarLabelsFollowWindowLanguage(t *testing.T) {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	widget := New(start, start.AddDate(0, 0, 6), nil).Key("calendar-labels")
	english, englishWeekdays := widget.calendarLabels(core.New(nil, new(theme.Light()), locale.LanguageEnglish))
	chinese, chineseWeekdays := widget.calendarLabels(core.New(nil, new(theme.Light()), locale.LanguageChinese))
	if english[0] != "Jan" || englishWeekdays[int(time.Monday)] != "Mon" {
		t.Fatalf("English labels = %q, %q", english[0], englishWeekdays[int(time.Monday)])
	}
	if chinese[0] != "1月" || chineseWeekdays[int(time.Monday)] != "一" {
		t.Fatalf("Chinese labels = %q, %q", chinese[0], chineseWeekdays[int(time.Monday)])
	}
}

func TestHeatmapMapsValuesAcrossTimeZones(t *testing.T) {
	// Range bounds in UTC; values carry a different location but the same civil
	// days. The cells must still be filled (regression: time.Time map keys
	// compare their location, which dropped every value across zones).
	tokyo := time.FixedZone("JST", 9*60*60)
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC) // Wednesday
	end := start.AddDate(0, 0, 6)
	values := []CalendarValue{
		{Date: time.Date(2025, time.January, 1, 0, 0, 0, 0, tokyo), Value: 7},
		{Date: time.Date(2025, time.January, 3, 0, 0, 0, 0, tokyo), Value: 9},
	}
	w := New(start, end, values).Key("cross-tz")

	jan1 := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	if got := w.values[int(jan1.Weekday())][0]; got != 7 {
		t.Fatalf("Jan 1 cell = %v, want 7", got)
	}
	jan3 := time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC)
	if got := w.values[int(jan3.Weekday())][0]; got != 9 {
		t.Fatalf("Jan 3 cell = %v, want 9", got)
	}
}

func TestHeatmapComparesRangeAsCivilDates(t *testing.T) {
	west := time.FixedZone("UTC-12", -12*60*60)
	east := time.FixedZone("UTC+14", 14*60*60)
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, west)
	end := time.Date(2025, time.January, 1, 0, 0, 0, 0, east)

	w := New(start, end, nil).Key("civil-range")
	if len(w.values) != 7 || len(w.values[0]) != 1 {
		t.Fatalf("calendar grid = %d rows x %d columns", len(w.values), len(w.values[0]))
	}
}

func TestHeatmapValueRangeOverridesData(t *testing.T) {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	w := New(start, start, []CalendarValue{{Date: start, Value: 10}}).Key("test").ValueRange(0, 100)
	minimum, maximum := w.valueExtent(1, 1)
	if minimum != 0 || maximum != 100 {
		t.Fatalf("range = %v, %v", minimum, maximum)
	}
}

func TestHeatmapVisualCacheTracksThemeLifecycle(t *testing.T) {
	active := new(theme.Light())
	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("cache")

	value.resolveVisual(active)
	value.resolveVisual(active)
	replacement := *active
	value.resolveVisual(&replacement)

	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 0 {
		t.Fatalf("heatmap visual cache stats = %#v", stats)
	}
}

func TestHeatmapDefaultsUseComponentDefaults(t *testing.T) {
	active := new(theme.Light())

	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("token-theme")
	visual := value.resolveVisual(active)
	if visual.metrics.Height != defaultHeatmapMetrics.Height {
		t.Fatalf("heatmap height = %v, want %v", visual.metrics.Height, defaultHeatmapMetrics.Height)
	}
	if visual.colors.Min != (color.NRGBA{R: 0xe8, G: 0xf1, B: 0xff, A: 0xff}) ||
		visual.colors.Max != (color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff}) {
		t.Fatalf("heatmap colors = %#v/%#v, want component palette defaults", visual.colors.Min, visual.colors.Max)
	}
	root := style.Cascade(style.State{}, visual.defaults)
	if root.Box == nil || root.Box.FillWidth == nil || !*root.Box.FillWidth {
		t.Fatalf("heatmap defaults do not fill their width: %#v", root.Box)
	}
}

func TestHeatmapStyleIsLastCascadeLayer(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("styled").
		Style(style.Style{}.Padding(9))
	visual := value.resolveVisual(active)

	resolved := computed.ResolveStatic(ctx, style.State{}, visual.defaults, style.Style{}, style.Style{}, value.style)
	if resolved.Box == nil || resolved.Box.Padding == nil || resolved.Box.Padding.Left != 9 {
		t.Fatalf("instance style did not reach the root cascade: %#v", resolved.Box)
	}

	plain := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("styled").resolveVisual(active)
	if visual.defaults.Hash64() != plain.defaults.Hash64() {
		t.Fatal("instance style leaked into the cached theme defaults")
	}
}

func TestHeatmapStyleDoesNotChangeDefaultsCacheKey(t *testing.T) {
	active := new(theme.Light())
	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("style-cache")
	value.resolveVisual(active)

	styled := value.Style(style.Style{}.Padding(9))
	styled.resolveVisual(active)

	stats := styled.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.KeyInvalidations != 0 {
		t.Fatalf("instance style changed the defaults cache: %#v", stats)
	}
}

func TestHeatmapColorOverrides(t *testing.T) {
	active := new(theme.Light())
	minColor := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	maxColor := color.NRGBA{R: 0xd0, G: 0xe0, B: 0xf0, A: 0xff}
	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("colors").
		MinColor(minColor).
		MaxColor(maxColor)

	visual := value.resolveVisual(active)
	if visual.colors.Min != minColor || visual.colors.Max != maxColor {
		t.Fatalf("heatmap colors = %#v, %#v; want %#v, %#v", visual.colors.Min, visual.colors.Max, minColor, maxColor)
	}
	if stats := value.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 1 {
		t.Fatalf("heatmap defaults cache = %#v, want one miss", stats)
	}

	// The overrides are applied to the resolved copy, so the cached theme
	// defaults must stay untouched.
	plain := value
	plain.hasMinColor, plain.hasMaxColor = false, false
	defaults := plain.resolveVisual(active)
	if defaults.colors.Min != (color.NRGBA{R: 0xe8, G: 0xf1, B: 0xff, A: 0xff}) ||
		defaults.colors.Max != (color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 0xff}) {
		t.Fatalf("instance colors leaked into the cached defaults: %#v", defaults.colors)
	}
	if stats := value.visualCache.Stats(); stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("resolving other instance colors changed the defaults cache key: %#v", stats)
	}
}

func TestHeatmapSelectionUsesColorScale(t *testing.T) {
	value := New(time.Unix(1, 0), time.Unix(2, 0), nil).Key("scale")
	value.values = [][]float64{{5}}
	value.cellLabels = [][]string{{"cell"}}
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	calls := 0
	scale := func(minValue, maxValue, value float64) color.NRGBA {
		calls++
		if minValue != 0 || maxValue != 10 || value != 5 {
			t.Fatalf("scale inputs = %v, %v, %v; want 0, 10, 5", minValue, maxValue, value)
		}
		return want
	}
	selection := value.selection(f32.Pt(5, 5), true, image.Rect(0, 0, 10, 10), 10, 10, 1, 1, 0, 10, scale)
	if !selection.valid || selection.color != want {
		t.Fatalf("selection = %#v, want visual color %#v", selection, want)
	}
	if calls != 1 {
		t.Fatalf("scale calls = %d, want 1", calls)
	}
}
