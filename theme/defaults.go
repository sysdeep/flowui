package theme

import "runtime"

// configureTypography names real installed families. Gio does not map CSS
// generic names such as "sans-serif" or "monospace", so a generic name on its
// own falls back to an arbitrary face, and a monospace block can lose its CJK
// glyphs entirely. Each list puts the UI face first and a CJK face after it so
// fallback happens per glyph.
func configureTypography(typography *Typography) {
	switch runtime.GOOS {
	case "windows":
		// Generic sans-serif can choose a CJK face with excessive line gap on
		// Windows, especially at medium weight. Prefer matching UI families for
		// Latin and Chinese so controls center a consistent font line box.
		typography.Typeface = "Segoe UI, Microsoft YaHei UI"
		typography.MonoTypeface = "Cascadia Mono, Consolas, Microsoft YaHei UI"
	case "darwin":
		typography.Typeface = "Helvetica Neue, PingFang SC"
		typography.MonoTypeface = "Menlo, PingFang SC"
	default:
		typography.Typeface = "Noto Sans, DejaVu Sans, Noto Sans CJK SC"
		typography.MonoTypeface = "DejaVu Sans Mono, Noto Sans Mono, Noto Sans CJK SC"
	}
}

func configureDefaults(active *Theme, seed Seed, dark bool) {
	active.UseSemantic(semanticColors(active.Semantic(), seed, dark))
	active.Shadows = ShadowsTheme{
		Overlay: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 4, Blur: 8, Spread: -2, Opacity: .4},
			{OffsetY: 12, Blur: 32, Spread: -4, Opacity: 1},
		}},
		Menu: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 0, Spread: 1, Opacity: .60},
			{OffsetY: 6, Blur: 16, Opacity: .70},
			{OffsetY: 16, Blur: 40, Opacity: .55},
		}},
	}
}
