package theme

import (
	"image/color"
	"time"
)

// Light returns FlowUI's default light theme, ready for window.Theme.
func Light() Theme { return New(DefaultSeed()) }

// Dark returns FlowUI's default dark theme, ready for window.Theme.
func Dark() Theme { return NewDark(DefaultSeed()) }

// New creates the default light theme with seed-based colors and motion.
func New(seed Seed) Theme { return newTheme(seed, false) }

// NewDark creates the default dark theme with seed-based colors and motion.
func NewDark(seed Seed) Theme { return newTheme(seed, true) }

func newTheme(seed Seed, dark bool) Theme {
	t := Theme{
		Typography:      defaultTypography(),
		Fonts:           FontConfig{SystemFonts: true},
		Shadows:         DefaultShadows(),
		Motion:          motionFromSeed(seed.Motion),
		DisabledOpacity: 0.5,
	}
	scales, opts := ScalesFromSeed(seed), DefaultDeriveOptions()
	if dark {
		scales, opts = DarkScalesFromSeed(seed), DarkDeriveOptions()
	}
	t.UseSemantic(scales.Derive(opts))
	t.Code = defaultCodeColors(&t)
	configureDefaults(&t, seed, dark)
	SyncMaterialTheme(&t)
	return t
}

// ScalesFromSeed turns the seed's hues into ten-step ramps. An unset hue takes
// the primary's ramp rather than a hard-coded brand colour, so a partially
// filled seed stays consistent with itself instead of importing someone else's
// red or green.
func ScalesFromSeed(seed Seed) Scales { return scalesFromSeed(seed, Ramp) }

// DarkScalesFromSeed is ScalesFromSeed against a dark canvas. Every ramp is a
// dark one, because a light ramp's pale end stays pale however the neutral roles
// are composited.
func DarkScalesFromSeed(seed Seed) Scales { return scalesFromSeed(seed, DarkRamp) }

func scalesFromSeed(seed Seed, ramp func(color.NRGBA) ColorScale) Scales {
	primary := ramp(seedOr(seed.Primary, seedGrey))
	fallback := primary.At(Step500)
	return Scales{
		Primary:   primary,
		Secondary: primary, // one brand hue; the neutral tones cover the rest
		Neutral:   ramp(seedOr(seed.Neutral, seedGrey)),
		Danger:    ramp(seedOr(seed.Danger, fallback)),
		Success:   ramp(seedOr(seed.Success, fallback)),
		Warning:   ramp(seedOr(seed.Warning, fallback)),
		Info:      ramp(seedOr(seed.Info, fallback)),
	}
}

// seedGrey is the fallback for every unset hue: a mid grey, so a zero Seed
// yields a theme with no brand colour instead of borrowing one.
var seedGrey = color.NRGBA{R: 0x8c, G: 0x8c, B: 0x8c, A: 0xff}

// seedOr falls back when a colour was left unset. Ramps are opaque, so alpha is
// normalised rather than compared.
func seedOr(c, fallback color.NRGBA) color.NRGBA {
	if c.R == 0 && c.G == 0 && c.B == 0 {
		return fallback
	}
	c.A = 0xff
	return c
}

// motionFromSeed maps the seed's motion policy onto the theme's. A missing
// duration falls back to the framework's 200ms baseline.
func motionFromSeed(m MotionSeed) MotionTheme {
	d := m.Duration
	if d <= 0 {
		d = 200 * time.Millisecond
	}
	return MotionTheme{
		Enabled:         m.Enabled,
		DefaultDuration: d,
		DurationScale:   1,
		ControlDuration: d,
		Ease:            m.Ease,
	}
}

// defaultTypography is the framework's type ladder. It is a constant, not a
// derivation: nothing in a Seed changes type size yet.
func defaultTypography() Typography {
	typography := Typography{
		BodySize:    14,
		ControlSize: 14,
		SmallSize:   12,
	}
	configureTypography(&typography)
	return typography
}
