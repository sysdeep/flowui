package theme_test

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/qianniancn/flowui/theme"
)

func TestResolveMotionDuration(t *testing.T) {
	if got := theme.ResolveMotionDuration(theme.MotionTheme{Enabled: true, DurationScale: 0.5}, time.Second); got != 500*time.Millisecond {
		t.Fatalf("scaled duration = %v, want 500ms", got)
	}
	if got := theme.ResolveMotionDuration(theme.MotionTheme{Enabled: false, DurationScale: 1}, time.Second); got != 0 {
		t.Fatalf("disabled duration = %v, want 0", got)
	}
	for _, scale := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		motion := theme.MotionTheme{Enabled: true, DurationScale: scale}
		if theme.MotionEnabled(motion) || theme.ResolveMotionDuration(motion, time.Second) != 0 {
			t.Fatalf("non-finite scale %v enabled motion", scale)
		}
	}
}

func TestDisabledColor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		opacity float32
		alpha   uint8
	}{
		{name: "scaled", opacity: .25, alpha: 50},
		{name: "transparent", opacity: 0, alpha: 0},
		{name: "above range", opacity: 2, alpha: 200},
		{name: "below range", opacity: -1, alpha: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			active := theme.Light()
			active.DisabledOpacity = tc.opacity
			got := active.DisabledColor(color.NRGBA{R: 1, G: 2, B: 3, A: 200})
			if want := (color.NRGBA{R: 1, G: 2, B: 3, A: tc.alpha}); got != want {
				t.Fatalf("disabled color = %#v, want %#v", got, want)
			}
		})
	}
}

func TestDefaultThemeSyncsMaterialBridge(t *testing.T) {
	active := theme.Light()
	bridge := theme.MaterialOf(&active)
	if bridge == nil {
		t.Fatal("missing material bridge")
	}
	if bridge.Palette.ContrastBg != active.Colors.Tones[theme.TonePrimary].Fill {
		t.Fatalf("material accent = %#v, want %#v", bridge.Palette.ContrastBg, active.Colors.Tones[theme.TonePrimary].Fill)
	}
	if bridge.Palette.Fg != active.Colors.Foreground {
		t.Fatalf("material foreground = %#v, want %#v", bridge.Palette.Fg, active.Colors.Foreground)
	}
	if bridge.Face != active.Typography.Typeface {
		t.Fatalf("material face = %q, want %q", bridge.Face, active.Typography.Typeface)
	}
	if bridge.Shaper == nil {
		t.Fatal("missing material text shaper")
	}
}

func TestThemesUseIndependentTextShapers(t *testing.T) {
	first := theme.Light()
	second := theme.Light()
	if theme.MaterialOf(&first).Shaper == theme.MaterialOf(&second).Shaper {
		t.Fatal("default themes share a text shaper")
	}

	copy := first
	theme.DetachMaterial(&copy)
	bridge := theme.MaterialOf(&copy)
	if bridge.Shaper == nil || bridge.Shaper == theme.MaterialOf(&first).Shaper {
		t.Fatal("detached theme did not acquire an independent text shaper")
	}
	if theme.MaterialOf(&copy).Shaper != bridge.Shaper {
		t.Fatal("reading the material bridge replaced its text shaper")
	}
	copy.Colors.Foreground = color.NRGBA{R: 12, A: 255}
	theme.SyncMaterialTheme(&copy)
	if theme.MaterialOf(&first).Shaper == theme.MaterialOf(&copy).Shaper {
		t.Fatal("detached themes share a text shaper")
	}
	theme.SyncMaterialTheme(&copy)
	if theme.MaterialOf(&first) == theme.MaterialOf(&copy) {
		t.Fatal("detached themes share a material bridge")
	}
	if theme.MaterialOf(&first).Palette.Fg == theme.MaterialOf(&copy).Palette.Fg {
		t.Fatal("detached palette mutation reached the original theme")
	}
}

func TestSyncMaterialThemeUsesFontConfig(t *testing.T) {
	active := theme.Light()
	active.Typography.Typeface = "FlowUI Sans"
	active.Fonts.SystemFonts = false
	theme.SyncMaterialTheme(&active)

	bridge := theme.MaterialOf(&active)
	if bridge.Face != "FlowUI Sans" {
		t.Fatalf("material face = %q, want FlowUI Sans", bridge.Face)
	}
	if bridge.Shaper == nil {
		t.Fatal("font configuration removed the text shaper")
	}
}

func TestDarkThemeDefinesThemedSurfaceAndShadow(t *testing.T) {
	dark := theme.Dark()
	if dark.Colors.Surface == theme.Light().Colors.Surface {
		t.Fatal("dark theme did not override surface")
	}
	if dark.Colors.OverlayShadow == theme.Light().Colors.OverlayShadow {
		t.Fatal("dark theme did not override overlay shadow")
	}
}

// Text selection blends the primary color over each mode's field background.
func TestSelectionColorsMatchTheme(t *testing.T) {
	for name, want := range map[string]struct {
		th    theme.Theme
		value color.NRGBA
	}{
		"light": {theme.Light(), color.NRGBA{R: 0xc5, G: 0xd0, B: 0xf5, A: 0xff}},
		"dark":  {theme.Dark(), color.NRGBA{R: 0x1f, G: 0x2b, B: 0x50, A: 0xff}},
	} {
		if got := want.th.Colors.Selection; got != want.value {
			t.Fatalf("%s selection = %#v, want %#v", name, got, want.value)
		}
		semantic := want.th.Semantic()
		if got := theme.Composite(semantic.FieldBackground, semantic.Ring, .3); got != want.value {
			t.Fatalf("%s selection does not follow the field and focus colors: %#v", name, got)
		}
	}
}

func TestSemanticPreservesTransparentColors(t *testing.T) {
	active := theme.Light()
	colors := active.Semantic()
	colors.Overlay = color.NRGBA{}
	colors.ControlSurface = color.NRGBA{}
	colors.FieldBackground = color.NRGBA{}
	colors.Separator = color.NRGBA{}
	colors.OverlayForeground = color.NRGBA{}
	active.UseSemantic(colors)
	if got := active.Semantic(); got != colors {
		t.Fatalf("semantic colors changed during installation: %#v", got)
	}
}
