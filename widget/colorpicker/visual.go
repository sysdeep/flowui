package colorpicker

import (
	"image/color"
	"time"

	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// visualDefaultsCache memoizes one theme-derived defaults bundle. The instance
// style is resolved separately as the last cascade layer, so it never enters
// the cache key.
type visualDefaultsCache[V any] struct {
	cache computed.VisualCache[struct{}, V]
}

func (c *visualDefaultsCache[V]) resolve(active *theme.Theme, load func() V) V {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, struct{}{}, load)
}

func (c *visualDefaultsCache[V]) stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}

// colorPickerDefaults is the cached, theme-derived appearance and geometry of
// the picker shell.
type colorPickerDefaults struct {
	Defaults style.Style

	TriggerGap        unit.Dp
	PanelGap          unit.Dp
	PanelWidth        unit.Dp
	PanelPadding      unit.Dp
	PanelRadius       unit.Dp
	PanelMaxHeight    unit.Dp
	ContentGap        unit.Dp
	CompactContentGap unit.Dp
	ChannelGap        unit.Dp
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	EnterScale        float32

	TriggerTextSize unit.Sp
	TriggerRadius   unit.Dp
	FocusRingWidth  unit.Dp

	PanelForeground color.NRGBA
	PanelBackground color.NRGBA
	Foreground      color.NRGBA
	MutedForeground color.NRGBA
	FocusColor      color.NRGBA
	PanelShadow     draw.BoxShadow
	DisabledOpacity float32
	Motion          theme.MotionTheme
	Material        *material.Theme
}

// colorWheelDefaults is the cached, theme-derived appearance and geometry of
// one color wheel.
type colorWheelDefaults struct {
	Defaults style.Style

	Size             unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
	FocusColor       color.NRGBA
	DisabledOpacity  float32
	Motion           theme.MotionTheme
}

// colorAreaDefaults is the cached, theme-derived appearance and geometry of
// one saturation/value area.
type colorAreaDefaults struct {
	Defaults style.Style

	Size              unit.Dp
	Radius            unit.Dp
	ThumbSize         unit.Dp
	DraggingThumbSize unit.Dp
	ThumbBorderWidth  unit.Dp
	FocusRingWidth    unit.Dp
	DotSize           unit.Dp
	DotGap            unit.Dp
	FocusColor        color.NRGBA
	DisabledOpacity   float32
	Motion            theme.MotionTheme
}

// colorSliderDefaults is the cached, theme-derived appearance and geometry of
// one hue or alpha slider.
type colorSliderDefaults struct {
	Defaults style.Style

	TextSize         unit.Sp
	HeaderGap        unit.Dp
	TrackHeight      unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
	Foreground       color.NRGBA
	MutedForeground  color.NRGBA
	FocusColor       color.NRGBA
	DisabledOpacity  float32
	Motion           theme.MotionTheme
}

// colorFieldDefaults is the cached, theme-derived appearance and geometry of
// the hex color field.
type colorFieldDefaults struct {
	Defaults style.Style

	Gap             unit.Dp
	ErrorTextSize   unit.Sp
	ErrorForeground color.NRGBA
}

// colorSwatchDefaults is the cached, theme-derived geometry of one standalone
// color swatch.
type colorSwatchDefaults struct {
	Defaults style.Style

	ExtraSmallSize unit.Dp
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	ExtraLargeSize unit.Dp
	SquareRadius   unit.Dp
}

// colorSwatchPickerDefaults is the cached, theme-derived appearance and
// geometry of a color swatch grid.
type colorSwatchPickerDefaults struct {
	Defaults style.Style

	Gap                          unit.Dp
	ExtraSmallBorderWidth        unit.Dp
	BorderWidth                  unit.Dp
	LargeBorderWidth             unit.Dp
	FocusRingWidth               unit.Dp
	FocusRingGap                 unit.Dp
	CheckStroke                  unit.Dp
	SquareItemRadiusExtraSmall   unit.Dp
	SquareItemRadiusSmall        unit.Dp
	SquareItemRadius             unit.Dp
	SquareSwatchRadiusExtraSmall unit.Dp
	SquareSwatchRadiusSmall      unit.Dp
	SquareSwatchRadius           unit.Dp
	SquareSelectedSmallRadius    unit.Dp
	ShadowOpacity                float32
	SelectedScale                float32
	SelectionDuration            time.Duration
	CheckDuration                time.Duration
	HoverDuration                time.Duration
	FocusColor                   color.NRGBA
	ShadowColor                  color.NRGBA
	Shadows                      theme.ShadowTheme
	DisabledOpacity              float32
	Motion                       theme.MotionTheme
	Swatch                       colorSwatchDefaults
}

type (
	colorPickerDefaultsCache       = visualDefaultsCache[colorPickerDefaults]
	colorWheelDefaultsCache        = visualDefaultsCache[colorWheelDefaults]
	colorAreaDefaultsCache         = visualDefaultsCache[colorAreaDefaults]
	colorSliderDefaultsCache       = visualDefaultsCache[colorSliderDefaults]
	colorFieldDefaultsCache        = visualDefaultsCache[colorFieldDefaults]
	colorSwatchDefaultsCache       = visualDefaultsCache[colorSwatchDefaults]
	colorSwatchPickerDefaultsCache = visualDefaultsCache[colorSwatchPickerDefaults]
)

const (
	standaloneColorSwatchDefaultsKey  = "colorpicker:standalone"
	standaloneColorSwatchDefaultsSlot = "defaults"
)

func colorPickerDefaultsFor(ctx *core.Context, cache *colorPickerDefaultsCache) colorPickerDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorPickerDefaults {
		return buildColorPickerDefaults(active)
	})
}

func colorWheelDefaultsFor(ctx *core.Context, cache *colorWheelDefaultsCache) colorWheelDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorWheelDefaults {
		return buildColorWheelDefaults(active)
	})
}

func colorAreaDefaultsFor(ctx *core.Context, cache *colorAreaDefaultsCache) colorAreaDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorAreaDefaults {
		return buildColorAreaDefaults(active)
	})
}

func colorSliderDefaultsFor(ctx *core.Context, cache *colorSliderDefaultsCache) colorSliderDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorSliderDefaults {
		return buildColorSliderDefaults(active)
	})
}

func colorFieldDefaultsFor(ctx *core.Context, cache *colorFieldDefaultsCache) colorFieldDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorFieldDefaults {
		return buildColorFieldDefaults(active)
	})
}

func colorSwatchPickerDefaultsFor(ctx *core.Context, cache *colorSwatchPickerDefaultsCache) colorSwatchPickerDefaults {
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorSwatchPickerDefaults {
		return buildColorSwatchPickerDefaults(active)
	})
}

// standaloneColorSwatchDefaults shares one defaults cache across all standalone
// swatches, which have no identity of their own.
func standaloneColorSwatchDefaults(ctx *core.Context) colorSwatchDefaults {
	cache := core.UseState[colorSwatchDefaultsCache](ctx, standaloneColorSwatchDefaultsKey, standaloneColorSwatchDefaultsSlot)
	active := core.ActiveTheme(ctx)
	return cache.resolve(active, func() colorSwatchDefaults {
		return buildColorSwatchDefaults(active)
	})
}

func colorControlOpacity(enabled bool, disabledOpacity float32) float32 {
	if enabled {
		return 1
	}
	return disabledOpacity
}
