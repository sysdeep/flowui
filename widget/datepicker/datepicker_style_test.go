package datepicker

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

func TestDatePickerDefaultsCacheTracksVariantAndTheme(t *testing.T) {
	active := theme.Light()
	var cache datePickerPartDefaultsCache
	first := cache.defaultsFor(&active, Primary)
	if first.metrics.Height != defaultDatePickerGeometry.Height {
		t.Fatalf("cached height = %v, want %v", first.metrics.Height, defaultDatePickerGeometry.Height)
	}
	cache.defaultsFor(&active, Primary)
	cache.defaultsFor(&active, Secondary)
	replacement := theme.Dark()
	cache.defaultsFor(&replacement, Secondary)
	if stats := cache.defaultsCache.Stats(); stats.Hits != 1 || stats.Misses != 3 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("cache stats = %+v", stats)
	}
}

func TestDatePickerInstanceStyleIsNotPartOfDefaults(t *testing.T) {
	declaration := style.Style{}.TextColor(style.RGB(0x102030))
	for name, got := range map[string]style.Style{
		"picker": New("date", testDate(2026, 7, 1)).Style(declaration).style,
		"field":  DateField("date-field", testDate(2026, 7, 1)).Style(declaration).style,
		"range":  DateRangePicker("date-range", DateRange{}).Style(declaration).style,
	} {
		if got.Hash64() != declaration.Hash64() {
			t.Fatalf("%s instance style = %#x, want %#x", name, got.Hash64(), declaration.Hash64())
		}
	}

	active := theme.Light()
	var cache datePickerPartDefaultsCache
	cache.defaultsFor(&active, Primary)
	before := cache.defaultsCache.Stats()
	cache.defaultsFor(&active, Primary)
	after := cache.defaultsCache.Stats()
	if after.Hits != before.Hits+1 || after.Misses != before.Misses {
		t.Fatalf("defaults cache stats after instance-independent hit = %+v, want one more hit and no miss", after)
	}
}

func TestDatePickerInstanceStyleIsLastCascadeLayer(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	declaration := style.Style{}.Part(PartCalendarCell, style.Style{}.TextColor(style.RGB(0x123456)))
	defaults := buildDatePickerDefaults(core.ActiveTheme(ctx), Primary)
	picker := New("styled", testDate(2026, 7, 1)).Style(declaration)

	resolved := picker.resolvePart(ctx, gtx, PartCalendarCell, style.State{}, &defaults.calendar)
	got, ok := resolve.Color(resolved.Text.Color)
	if !ok {
		t.Fatal("styled cell has no text colour")
	}
	if got != want {
		t.Fatalf("styled cell text = %v, want instance colour %v", got, want)
	}
}

func TestDatePickerOutlineFallbackForUnborderedField(t *testing.T) {
	active := theme.Light()
	for _, variant := range []Variant{Primary, Filled, Borderless, Underlined} {
		defaults := buildDatePickerDefaults(&active, variant)
		resolved := style.CascadePart(style.State{}, style.PartContent, defaults.field)
		fallback := active.Semantic().Ring
		if got := dateResolveOutlineColor(resolved, fallback); variant == Borderless && got != fallback {
			t.Fatalf("borderless outline fallback = %v, want %v", got, fallback)
		}
	}
}

func TestDefaultDatePickerVariants(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		for _, variant := range []Variant{Primary, Secondary, Filled, Borderless, Underlined} {
			defaults := buildDatePickerDefaults(&active, variant)
			if defaults.metrics.Height <= 0 || defaults.metrics.CellSize <= 0 || defaults.metrics.CalendarWidth <= 0 {
				t.Fatalf("variant %v has incomplete geometry: %+v", variant, defaults.metrics)
			}
			panel := style.CascadePart(style.State{}, PartCalendarPanel, defaults.calendar)
			if panel.Paint == nil || panel.Box == nil || panel.Box.Padding == nil {
				t.Fatalf("variant %v calendar panel is incomplete", variant)
			}
			cell := style.CascadePart(style.State{}, PartCalendarCell, defaults.calendar)
			if cell.Text == nil || cell.Paint == nil || cell.Paint.Radius == nil {
				t.Fatalf("variant %v calendar cell is incomplete", variant)
			}
		}
	}
}

func TestDefaultDatePickerGeometryIsCompact(t *testing.T) {
	metrics := defaultDatePickerGeometry
	if metrics.CalendarWidth != metrics.CellSize*7 {
		t.Fatalf("calendar width = %v, want seven cells (%v)", metrics.CalendarWidth, metrics.CellSize*7)
	}
	if metrics.PopoverPadding > 10 || metrics.CellSize > 32 {
		t.Fatalf("datepicker geometry remains oversized: padding=%v cell=%v", metrics.PopoverPadding, metrics.CellSize)
	}
	if metrics.HeaderBottom <= 0 || metrics.CalendarTopGap <= 0 || metrics.WeekdayGap <= 0 {
		t.Fatal("datepicker calendar spacing must remain positive")
	}
}
