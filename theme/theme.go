package theme

import (
	"image/color"
	"math"
	"time"

	"gioui.org/font"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Theme contains FlowUI semantic tokens and framework-level values.
//
// The Gio material theme is an unexported substrate bridge for text/editor
// helpers. Applications theme through Colors, Typography, Fonts,
// Motion and component style overrides — never through material.Theme.
type Theme struct {
	Colors     Semantic
	Typography Typography
	Fonts      FontConfig
	Shadows    ShadowsTheme
	Motion     MotionTheme
	// Code is the syntax-highlighting palette. Assign it to customize token
	// colors; leave it zero to use the palette-derived default.
	Code            CodeColors
	DisabledOpacity float32
	material        *material.Theme
	managedShaper   *text.Shaper
	managedFace     font.Typeface
}

type MotionTheme struct {
	// Enabled controls all framework animations.
	Enabled bool
	// DefaultDuration is used by Tween when no duration is specified.
	DefaultDuration time.Duration
	// DurationScale scales animation time globally. Zero disables motion.
	DurationScale float32
	// ControlDuration is the baseline for state feedback
	// (color, focus, press). Zero lets each component use its own constant.
	ControlDuration time.Duration
	// Ease is the curve for state transitions.
	Ease EaseKind
}

// EaseKind names a transition curve. The mapping to an
// actual easing function lives resolver-side (animation.EasingFor) so rule
// tables stay data.
type EaseKind uint8

const (
	EaseDefault EaseKind = iota
	EaseLinear
	EaseOut
	EaseInOut
	EaseBack
)

// ResolveMotionDuration applies the active motion policy to a transition.
// A non-positive result means the transition should snap to its target.
func ResolveMotionDuration(motion MotionTheme, duration time.Duration) time.Duration {
	if !MotionEnabled(motion) || duration <= 0 {
		return 0
	}
	scaled := time.Duration(float64(duration) * float64(motion.DurationScale))
	if scaled <= 0 {
		return 0
	}
	return scaled
}

// MotionEnabled reports whether the global motion policy permits animation.
func MotionEnabled(motion MotionTheme) bool {
	return motion.Enabled && motion.DurationScale > 0 &&
		!math.IsNaN(float64(motion.DurationScale)) && !math.IsInf(float64(motion.DurationScale), 0)
}

// GradientStop is one color stop of a Gradient. An Offset outside [0, 1]
// is normalized by the brush, like style.Gradient.
type GradientStop struct {
	Offset float32
	Color  color.NRGBA
}

// Gradient is the theme-side gradient value. It mirrors style.Gradient in
// shape; components convert it at the declaration boundary.
type Gradient struct {
	AngleDegrees float32
	Stops        []GradientStop
}

type Typography struct {
	// Typeface is the default family list used by framework text.
	Typeface font.Typeface
	// MonoTypeface is the recommended family list for code and tabular text.
	MonoTypeface font.Typeface
	BodySize     unit.Sp
	ControlSize  unit.Sp
	SmallSize    unit.Sp
}

// FontConfig controls the font faces available to each window's text shaper.
// Collection can contain faces parsed from TTF, OTF, or TTC data. SystemFonts
// keeps the platform font fallback enabled when true.
type FontConfig struct {
	Collection  []font.FontFace
	SystemFonts bool
}

const ShadowLayerCount = 3

// ShadowLayerTheme describes one theme-controlled soft shadow pass.
type ShadowLayerTheme struct {
	OffsetX unit.Dp
	OffsetY unit.Dp
	// Blur is capped at 96 physical pixels during rasterization.
	Blur   unit.Dp
	Spread unit.Dp
	// Opacity is clamped to [0, 1]. Zero disables the layer.
	Opacity float32
}

// ShadowTheme stores shadow layers from tightest to broadest.
type ShadowTheme struct {
	Layers [ShadowLayerCount]ShadowLayerTheme
}

// ShadowsTheme groups the shadow profiles used by framework surfaces.
type ShadowsTheme struct {
	Surface ShadowTheme
	Overlay ShadowTheme
	Menu    ShadowTheme
	Control ShadowTheme
}

// DefaultShadows returns the framework shadow profiles.
func DefaultShadows() ShadowsTheme {
	return ShadowsTheme{
		Surface: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 1, Blur: 2, Opacity: .72},
			{OffsetY: 2, Blur: 4, Opacity: .48},
		}},
		Overlay: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{OffsetY: 2, Blur: 6, Opacity: .72},
			{OffsetY: 8, Blur: 22, Spread: 4, Opacity: .56},
		}},
		Menu: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 6, Spread: 1, Opacity: .20},
			{OffsetY: 4, Blur: 12, Opacity: .10},
			{OffsetY: 12, Blur: 28, Spread: 2, Opacity: .10},
		}},
		Control: ShadowTheme{Layers: [ShadowLayerCount]ShadowLayerTheme{
			{Blur: 1, Opacity: 15.0 / 104.0},
			{OffsetY: 1, Blur: 2, Opacity: 15.0 / 104.0},
			{OffsetY: 2, Blur: 4, Opacity: 10.0 / 104.0},
		}},
	}
}

// MaterialOf returns the internal Gio material bridge for text/editor helpers.
// Callers outside this package must not treat it as part of the public theme API.
func MaterialOf(theme *Theme) *material.Theme {
	if theme == nil {
		return nil
	}
	if theme.material != nil && theme.material.Shaper == nil {
		theme.material.Shaper = newTextShaper(theme.Fonts)
		theme.managedShaper = theme.material.Shaper
	}
	return theme.material
}

// DetachMaterial gives theme a private material bridge copy so later mutations
// do not share state with another Theme value.
func DetachMaterial(theme *Theme) {
	if theme == nil || theme.material == nil {
		return
	}
	clone := *theme.material
	if clone.Shaper == theme.managedShaper {
		// Most snapshots only read tokens. Allocate their independent shaper
		// only if a caller actually requests the material bridge.
		clone.Shaper = nil
		theme.managedShaper = nil
	}
	if clone.Face == theme.managedFace {
		clone.Face = theme.Typography.Typeface
		theme.managedFace = clone.Face
	}
	theme.material = &clone
}

// SyncMaterialTheme updates the Gio material bridge from FlowUI tokens.
func SyncMaterialTheme(theme *Theme) {
	if theme == nil {
		return
	}
	if theme.material == nil {
		theme.material = material.NewTheme()
	}
	if theme.material.Shaper == nil || theme.managedShaper == nil || theme.material.Shaper == theme.managedShaper {
		theme.material.Shaper = newTextShaper(theme.Fonts)
		theme.managedShaper = theme.material.Shaper
	}
	if theme.material.Face == theme.managedFace {
		theme.material.Face = theme.Typography.Typeface
		theme.managedFace = theme.material.Face
	}
	theme.material.TextSize = theme.Typography.BodySize
	theme.material.Palette.Fg = theme.Colors.Foreground
	theme.material.Palette.Bg = theme.Colors.Background
	theme.material.Palette.ContrastBg = theme.Colors.Tones[TonePrimary].Fill
	theme.material.Palette.ContrastFg = theme.Colors.Tones[TonePrimary].OnFill
}

func newTextShaper(fonts FontConfig) *text.Shaper {
	// Keep the common system-only path lazy. Gio's zero-value Shaper loads the
	// platform font map on first use; eagerly scanning every installed font for
	// every temporary theme copy can otherwise cause a large startup spike.
	if fonts.SystemFonts && len(fonts.Collection) == 0 {
		return new(text.Shaper)
	}
	options := make([]text.ShaperOption, 0, 2)
	if !fonts.SystemFonts {
		options = append(options, text.NoSystemFonts())
	}
	if len(fonts.Collection) > 0 {
		options = append(options, text.WithCollection(fonts.Collection))
	}
	return text.NewShaper(options...)
}

func (theme *Theme) DisabledColor(c color.NRGBA) color.NRGBA {
	opacity := theme.DisabledOpacityValue()
	c.A = byte(float32(c.A) * opacity)
	return c
}

// DisabledOpacityValue returns the configured disabled opacity clamped to [0, 1].
func (theme *Theme) DisabledOpacityValue() float32 {
	opacity := float32(0.5)
	if theme != nil {
		opacity = theme.DisabledOpacity
	}
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	return opacity
}
