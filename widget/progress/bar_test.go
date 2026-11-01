package progress

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func newContext(_ any) *core.Context {
	return core.New(nil, new(theme.Light()), locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.Light()
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func TestBarRatio(t *testing.T) {
	if got := Bar(50).Key("upload").ratio(); got != 0.5 {
		t.Fatalf("ratio = %v, want 0.5", got)
	}
	if got := Bar(-10).Key("upload").ratio(); got != 0 {
		t.Fatalf("low ratio = %v, want 0", got)
	}
	if got := Bar(140).Key("upload").ratio(); got != 1 {
		t.Fatalf("high ratio = %v, want 1", got)
	}
	if got := Bar(75).Key("upload").Range(50, 100).ratio(); got != 0.5 {
		t.Fatalf("range ratio = %v, want 0.5", got)
	}
	if got := Bar(75).Key("upload").Range(10, 10).ratio(); got != 0 {
		t.Fatalf("invalid range ratio = %v, want 0", got)
	}
	if got := Bar(math.NaN()).Key("upload").ratio(); got != 0 {
		t.Fatalf("nan ratio = %v, want 0", got)
	}
	if got := Bar(math.Inf(1)).Key("upload").ratio(); got != 0 {
		t.Fatalf("inf ratio = %v, want 0", got)
	}
}

func TestBarOutputText(t *testing.T) {
	if got := Bar(42).Key("upload").outputText(); got != "" {
		t.Fatalf("default output = %q, want empty", got)
	}
	if got := Bar(42).Key("upload").ShowValue().outputText(); got != "42%" {
		t.Fatalf("percent output = %q, want 42%%", got)
	}
	if got := Bar(42).Key("upload").ValueText("42 files").outputText(); got != "42 files" {
		t.Fatalf("custom output = %q, want 42 files", got)
	}
	if got := Bar(42).Key("upload").ShowValue().Indeterminate().outputText(); got != "" {
		t.Fatalf("indeterminate output = %q, want empty", got)
	}
}

func TestBarSemanticDescription(t *testing.T) {
	if got := Bar(42).Label("Upload").Key("upload").semanticDescription(); got != "Upload 42%" {
		t.Fatalf("semantic description = %q, want Upload 42%%", got)
	}
	if got := Bar(42).ValueText("42 files").Key("upload").semanticDescription(); got != "Progress 42 files" {
		t.Fatalf("custom semantic description = %q, want Progress 42 files", got)
	}
	if got := Bar(0).Label("Sync").Key("upload").Indeterminate().semanticDescription(); got != "Sync in progress" {
		t.Fatalf("indeterminate semantic description = %q, want Sync in progress", got)
	}
}

// TestBarSizeStyle reads the size layer through the widget's own cascade seam
// so the test fails if the size layer stops being installed.
func TestBarSizeStyle(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	small := Bar(50).Size(BarSmall).Key("upload").resolveStyle(ctx, testLayoutContext(), "upload", &progressBarStyleCache{}).track
	large := Bar(50).Size(BarLarge).Key("upload").resolveStyle(ctx, testLayoutContext(), "upload", &progressBarStyleCache{}).track
	if got := small.Box.Height; got == nil || *got != 4 {
		t.Fatalf("small height = %v, want 4", got)
	}
	if got := large.Box.Height; got == nil || *got != 12 {
		t.Fatalf("large height = %v, want 12", got)
	}
}

func TestBarSeparatesRootTrackAndIndicatorStyles(t *testing.T) {
	rootColor := color.NRGBA{R: 1, A: 0xff}
	trackColor := color.NRGBA{G: 2, A: 0xff}
	fillColor := color.NRGBA{B: 3, A: 0xff}
	custom := style.Style{}.
		Height(60).
		Background(style.SolidColor{Color: rootColor}).
		Part(style.PartTrack, style.Style{}.Height(4).Background(style.SolidColor{Color: trackColor})).
		Part(style.PartFill, style.Style{}.Background(style.SolidColor{Color: fillColor}))

	resolved := Bar(50).Style(custom).Key("upload").resolveStyle(newContext(nil), testLayoutContext(), "upload", &progressBarStyleCache{})
	if resolved.root.Box == nil || resolved.root.Box.Height == nil || *resolved.root.Box.Height != 60 {
		t.Fatalf("root height = %#v", resolved.root.Box)
	}
	if got := resolved.root.Paint.Background.(style.SolidColor).Color; got != rootColor {
		t.Fatalf("root background = %#v", got)
	}
	if resolved.track.Box == nil || resolved.track.Box.Height == nil || *resolved.track.Box.Height != 4 {
		t.Fatalf("track box = %#v", resolved.track.Box)
	}
	if got := resolved.track.Paint.Background.(style.SolidColor).Color; got != trackColor {
		t.Fatalf("track background = %#v", got)
	}
	if got := resolved.fill.Paint.Background.(style.SolidColor).Color; got != fillColor {
		t.Fatalf("fill background = %#v", got)
	}
}

func TestBarInstanceStyleStaysOutsideDefaultsCache(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	cache := &progressBarStyleCache{}

	first := Bar(50).Key("upload").
		Style(style.Style{}.Part(style.PartTrack, style.Style{}.Height(3))).
		resolveStyle(ctx, gtx, "upload", cache)
	second := Bar(50).Key("upload").
		Style(style.Style{}.Part(style.PartTrack, style.Style{}.Height(9))).
		resolveStyle(ctx, gtx, "upload", cache)

	if got := first.track.Box.Height; got == nil || *got != 3 {
		t.Fatalf("first track height = %v, want 3", got)
	}
	if got := second.track.Box.Height; got == nil || *got != 9 {
		t.Fatalf("second track height = %v, want 9", got)
	}
}

func TestBarAnimation(t *testing.T) {
	state := new(progressBarState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start
	defaults := testProgressBarDefaults(t)
	duration := defaults.valueDuration

	if got := state.progress(gtx, 0, false, duration, defaults.motion); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	if got := state.progress(gtx, 1, false, duration, defaults.motion); got != 0 {
		t.Fatalf("animation start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.progress(gtx, 1, false, duration, defaults.motion)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("animation midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.progress(gtx, 1, false, duration, defaults.motion); got != 1 {
		t.Fatalf("animation end = %v, want 1", got)
	}
}

func TestBarIndeterminateOffset(t *testing.T) {
	width := 40
	start := time.Unix(1, 0)
	period := testProgressBarDefaults(t).indeterminatePeriod
	if got := progressBarIndeterminateOffset(time.Time{}, width, period); got != -width {
		t.Fatalf("zero time offset = %d, want %d", got, -width)
	}
	first := progressBarIndeterminateOffset(start, width, period)
	mid := progressBarIndeterminateOffset(start.Add(period/2), width, period)
	if first == mid {
		t.Fatal("indeterminate offset did not advance")
	}
	if got := progressBarIndeterminatePosition(start, width, 0); got != 0 {
		t.Fatalf("static indeterminate offset = %d, want 0", got)
	}
}

func testProgressBarDefaults(t *testing.T) progressBarDefaults {
	t.Helper()
	return Bar(0).Key("upload").defaultStyles(new(theme.Light()))
}

func TestIndeterminateBarRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := theme.Light()
		themeValue.Motion.Enabled = enabled
		ctx := core.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		Bar(0).Indeterminate().Key("sync").Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(200, 100)},
			Source:      router.Source(),
			Ops:         &ops,
			Now:         time.Unix(1, 0),
		})
		router.Frame(&ops)
		_, wake := router.WakeupTime()
		return wake
	}
	if !wakes(true) {
		t.Fatal("enabled indeterminate progress bar did not request redraw")
	}
	if wakes(false) {
		t.Fatal("indeterminate progress bar requested redraw with motion disabled")
	}
}

func TestBarLayoutTrackOnly(t *testing.T) {
	dims := Bar(50).Key("upload").Layout(newContext(nil), progressBarTestContext())
	if dims.Size != image.Pt(300, 8) {
		t.Fatalf("track-only size = %v, want (300,8)", dims.Size)
	}
}

func TestBarLayoutWithHeader(t *testing.T) {
	dims := Bar(50).Key("upload").
		Label("Upload").
		ShowValue().
		Layout(newContext(nil), progressBarTestContext())
	if dims.Size.X != 300 {
		t.Fatalf("header width = %d, want 300", dims.Size.X)
	}
	if dims.Size.Y <= 8 {
		t.Fatalf("header height = %d, want greater than track height", dims.Size.Y)
	}
}

func TestBarLabelPartUsesCommonBoxRenderer(t *testing.T) {
	base := Bar(50).Key("upload").
		Label("Upload").
		Layout(newContext(nil), progressBarTestContext())
	styled := (Bar(50).Label("Upload").Key("upload")).Style(style.Style{}.Part(style.PartLabel, style.Style{}.Padding(7))).Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled label height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func TestBarTrackPartUsesCommonBoxRenderer(t *testing.T) {
	base := Bar(50).Key("upload").Layout(newContext(nil), progressBarTestContext())
	styled := (Bar(50)).Key("upload").Style(style.Style{}.Part(style.PartTrack, style.Style{}.MarginY(7))).Layout(newContext(nil), progressBarTestContext())
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled track height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

// paletteTheme installs a theme's v1 palette as its colour source, so edits to
func progressBarTestContext() layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 120)},
		Ops:         &ops,
	}
}
