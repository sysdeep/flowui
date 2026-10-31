package spinner

import (
	"image"
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

func TestSpinnerLayoutUsesSelectedSize(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	var ops op.Ops
	dims := New().Size(Large).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	})
	if dims.Size != image.Pt(32, 32) {
		t.Fatalf("spinner dimensions = %v, want (32,32)", dims.Size)
	}
}

func TestSpinnerInstanceStyleAppliesLast(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	var ops op.Ops
	dims := New().Size(Large).Style(style.New().Width(64).Height(64)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	})
	if dims.Size != image.Pt(64, 64) {
		t.Fatalf("spinner dimensions = %v, want (64,64)", dims.Size)
	}
}

func TestSpinnerStyleDoesNotInvalidateDefaults(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageAuto)
	value := New().Size(Large)
	value.visualDefaults(ctx)
	baseline := value.visualCache.Stats()
	value.Style(style.New().Width(64)).visualDefaults(ctx)
	after := value.visualCache.Stats()
	if after.Misses != baseline.Misses || after.KeyInvalidations != baseline.KeyInvalidations {
		t.Fatalf(
			"instance style changed the defaults cache key: misses %d -> %d, key invalidations %d -> %d",
			baseline.Misses,
			after.Misses,
			baseline.KeyInvalidations,
			after.KeyInvalidations,
		)
	}
	if after.Hits != baseline.Hits+1 {
		t.Fatalf("defaults cache hits = %d, want %d", after.Hits, baseline.Hits+1)
	}
}

func TestSpinnerRespectsMotionTheme(t *testing.T) {
	wakes := func(enabled bool) bool {
		themeValue := theme.Light()
		themeValue.Motion.Enabled = enabled
		ctx := core.New(nil, &themeValue, locale.LanguageAuto)
		var router input.Router
		var ops op.Ops
		New().Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Source:      router.Source(),
			Ops:         &ops,
			Now:         time.Unix(1, 0),
		})
		router.Frame(&ops)
		_, wake := router.WakeupTime()
		return wake
	}
	if !wakes(true) {
		t.Fatal("enabled spinner did not request redraw")
	}
	if wakes(false) {
		t.Fatal("spinner requested redraw with motion disabled")
	}
}
