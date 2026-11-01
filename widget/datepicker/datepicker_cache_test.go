package datepicker

import (
	"image/color"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestDatePickerMonthDaysCacheReusesAndInvalidates(t *testing.T) {
	var state datePickerState
	month := testDate(2026, 7, 1)

	first := state.monthDays(month, time.Sunday)
	if len(first) != 42 {
		t.Fatalf("cached month days = %d, want 42", len(first))
	}
	firstAddress := &first[0]
	firstDate := first[0].date
	for _, day := range first {
		if want := strconv.Itoa(day.date.Day()); day.label != want {
			t.Fatalf("day %s label = %q, want %q", day.date.Format("2006-01-02"), day.label, want)
		}
	}

	second := state.monthDays(month, time.Sunday)
	if &second[0] != firstAddress {
		t.Fatal("same month and week start did not reuse the fixed calendar buffer")
	}
	if !sameDate(second[0].date, firstDate) {
		t.Fatalf("cached first date = %v, want %v", second[0].date, firstDate)
	}

	changedMonth := state.monthDays(testDate(2026, 8, 1), time.Sunday)
	if sameDate(changedMonth[0].date, firstDate) {
		t.Fatal("month change reused stale calendar contents")
	}
	changedWeekStart := state.monthDays(month, time.Monday)
	if dateKey(changedWeekStart[0].date) != "2026-06-29" {
		t.Fatalf("week-start change first date = %s, want 2026-06-29", dateKey(changedWeekStart[0].date))
	}

	otherLocation := time.FixedZone("UTC+8", 8*60*60)
	changedLocation := state.monthDays(time.Date(2026, 7, 1, 0, 0, 0, 0, otherLocation), time.Monday)
	if changedLocation[0].date.Location() != otherLocation {
		t.Fatalf("location change retained %v, want %v", changedLocation[0].date.Location(), otherLocation)
	}
}

func TestDatePickerMonthDaysIntoWritesOnlyCalendarSlots(t *testing.T) {
	days := make([]datePickerDay, 43)
	days[42] = datePickerDay{label: "sentinel"}

	datePickerMonthDaysInto(days, testDate(2026, 7, 1), time.Sunday)
	if days[42].label != "sentinel" {
		t.Fatalf("month generator overwrote slot 42 with %q", days[42].label)
	}
}

func TestDatePickerResolvedPartCacheTracksThemeAndScope(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache

	cache.prepare(ctx, style.Style{})
	if !cache.enabled {
		t.Fatal("transition-free theme should enable the resolved part cache")
	}
	value := style.ResolvedStyle{}
	cache.storeCell(7, value)
	if _, ok := cache.cellValue(7); !ok {
		t.Fatal("stored cell style was not returned")
	}
	if cache.hits != 1 || cache.misses != 0 {
		t.Fatalf("cache counters = hits %d misses %d, want one hit", cache.hits, cache.misses)
	}

	restore := core.PushStyle(ctx, style.Style{}.TextColor(style.RGB(0x102030)))
	cache.prepare(ctx, style.Style{})
	restore()
	if cache.scopeInvalidation == 0 {
		t.Fatal("style scope change did not invalidate resolved parts")
	}
	stats := cache.stats()
	if stats.ScopeInvalidations != 1 {
		t.Fatalf("scope invalidations = %d, want 1", stats.ScopeInvalidations)
	}
	if _, ok := cache.cellValue(7); ok {
		t.Fatal("resolved cell style survived a style scope change")
	}

	other := theme.Light()
	cache.storeWeekday(0, value)
	core.ReplaceTheme(ctx, other)
	cache.prepare(ctx, style.Style{})
	if cache.themeInvalidation == 0 {
		t.Fatal("theme replacement did not invalidate resolved parts")
	}
	stats = cache.stats()
	if stats.ThemeInvalidations != 1 {
		t.Fatalf("theme invalidations = %d, want 1", stats.ThemeInvalidations)
	}
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("weekday style survived a theme replacement")
	}
}

func TestDatePickerResolvedPartCacheTracksGenerationInvalidation(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, Primary)

	core.ReplaceLanguage(ctx, locale.LanguageChinese)
	cache.prepare(ctx, style.Style{}, Primary)

	stats := cache.stats()
	if stats.GenerationInvalidations != 1 {
		t.Fatalf("generation invalidations = %d, want 1", stats.GenerationInvalidations)
	}
	if stats.ThemeInvalidations != 0 {
		t.Fatalf("theme invalidations = %d, want 0 for language-only change", stats.ThemeInvalidations)
	}
}

func TestDatePickerResolvedPartCacheTracksVariant(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, Primary)
	cache.storeHeader(3, style.ResolvedStyle{})
	if _, ok := cache.headerValue(3); !ok {
		t.Fatal("stored header style was not returned")
	}

	cache.prepare(ctx, style.Style{}, Secondary)
	if _, ok := cache.headerValue(3); ok {
		t.Fatal("header style survived a variant change")
	}
	stats := cache.stats()
	if stats.VariantInvalidations != 1 {
		t.Fatalf("variant invalidations = %d, want 1", stats.VariantInvalidations)
	}
}

func TestDatePickerResolvedPartCacheStoresTextPartsByStateAndColor(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, Primary)
	state := style.State{Focused: true}
	base := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	key := datePickerTextResolvedKey(state, base)
	cache.storeSegment(key, style.ResolvedStyle{})
	if _, ok := cache.segmentValue(key); !ok {
		t.Fatal("stored segment style was not returned")
	}
	if _, ok := cache.segmentValue(datePickerTextResolvedKey(state, color.NRGBA{R: 0x11, G: 0x20, B: 0x30, A: 0xff})); ok {
		t.Fatal("segment style ignored its base color")
	}
	cache.storeLiteral(key, style.ResolvedStyle{})
	if _, ok := cache.literalValue(key); !ok {
		t.Fatal("stored literal style was not returned")
	}
}

func TestDatePickerResolvedPartCacheBypassesTransitions(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond))
	if !cache.enabled {
		t.Fatal("duration-only transition should remain cache-safe until resolution")
	}
	cache.storeCell(1, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.cellValue(1); ok {
		t.Fatal("transition-bearing resolved style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	if cache.enabled {
		t.Fatal("custom easing declaration should disable the resolved part cache")
	}
	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	if stats := cache.stats(); stats.UnsafeBypasses != 2 {
		t.Fatalf("unsafe bypasses = %d, want 2 across repeated uncached prepares", stats.UnsafeBypasses)
	}
}

func TestDatePickerResolvedPartCacheTracksWeekdayStates(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})

	value := style.ResolvedStyle{}
	cache.storeWeekday(0, value)
	if _, ok := cache.weekdayValue(0); !ok {
		t.Fatal("enabled weekday style was not returned")
	}
	if _, ok := cache.weekdayValue(1); ok {
		t.Fatal("disabled weekday style unexpectedly hit before storing")
	}
	cache.storeWeekday(1, value)
	if _, ok := cache.weekdayValue(1); !ok {
		t.Fatal("disabled weekday style was not returned")
	}

	restore := core.PushStyle(ctx, style.Style{}.TextColor(style.RGB(0x102030)))
	cache.prepare(ctx, style.Style{})
	restore()
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("weekday style survived a style scope change")
	}
}

func TestDatePickerResolvedPartCacheDoesNotStoreTransitionWeekday(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})
	cache.storeWeekday(0, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("transition-bearing weekday style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	cache.storeWeekday(0, style.ResolvedStyle{})
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("custom easing declaration unexpectedly enabled weekday cache")
	}
}

func TestDatePickerResolvedPartCacheDoesNotStoreTransitionFocus(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})
	cache.storeFocus(3, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.focusValue(3); ok {
		t.Fatal("transition-bearing focus style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	cache.storeFocus(3, style.ResolvedStyle{})
	if _, ok := cache.focusValue(3); ok {
		t.Fatal("custom easing declaration unexpectedly enabled focus cache")
	}
}

func TestDatePickerResolvedPartCacheUsedByOpenCalendar(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	picker := New(testDate(2026, 7, 9)).Key("resolved-cache")
	state := datePickerStateFor(ctx, picker.key)
	var router input.Router
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()

	layoutDatePickerFrameAt(ctx, &router, picker, testDate(2026, 7, 1))
	misses := state.resolvedParts.misses
	if misses == 0 {
		t.Fatal("open calendar did not resolve any parts")
	}
	layoutDatePickerFrameAt(ctx, &router, picker, testDate(2026, 7, 1).AddDate(0, 0, 1))
	if state.resolvedParts.hits == 0 {
		t.Fatal("second open calendar frame did not hit resolved part cache")
	}
	if state.resolvedParts.misses < misses {
		t.Fatalf("cache misses moved backwards: %d -> %d", misses, state.resolvedParts.misses)
	}
}
