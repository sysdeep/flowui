package colorpicker

import (
	"time"

	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// buildColorPickerDefaults derives the picker shell from the active theme.
// Instance declarations are resolved separately as the last cascade layer.
func buildColorPickerDefaults(active *theme.Theme) colorPickerDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	return colorPickerDefaults{
		Defaults:          style.Style{},
		TriggerGap:        12,
		PanelGap:          6,
		PanelWidth:        248,
		PanelPadding:      16,
		PanelRadius:       8,
		PanelMaxHeight:    520,
		ContentGap:        12,
		CompactContentGap: 8,
		ChannelGap:        8,
		EnterDuration:     150 * time.Millisecond,
		ExitDuration:      100 * time.Millisecond,
		EnterScale:        1,
		TriggerTextSize:   14,
		TriggerRadius:     4,
		FocusRingWidth:    2,
		PanelForeground:   sem.OverlayForeground,
		PanelBackground:   sem.Overlay,
		Foreground:        sem.Foreground,
		MutedForeground:   sem.MutedForeground,
		FocusColor:        sem.Ring,
		PanelShadow:       draw.ThemeShadow(active.Shadows.Overlay, sem.OverlayShadow, 1),
		DisabledOpacity:   active.DisabledOpacityValue(),
		Motion:            active.Motion,
		Material:          theme.MaterialOf(active),
	}
}

func buildColorWheelDefaults(active *theme.Theme) colorWheelDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return colorWheelDefaults{
		Defaults:         style.Style{},
		Size:             190,
		ThumbSize:        20,
		ThumbBorderWidth: 3,
		FocusRingWidth:   2,
		FocusColor:       active.Semantic().Ring,
		DisabledOpacity:  active.DisabledOpacityValue(),
		Motion:           active.Motion,
	}
}

func buildColorAreaDefaults(active *theme.Theme) colorAreaDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return colorAreaDefaults{
		Defaults:          style.Style{},
		Size:              224,
		Radius:            4,
		ThumbSize:         16,
		DraggingThumbSize: 20,
		ThumbBorderWidth:  3,
		FocusRingWidth:    2,
		DotSize:           2,
		DotGap:            8,
		FocusColor:        active.Semantic().Ring,
		DisabledOpacity:   active.DisabledOpacityValue(),
		Motion:            active.Motion,
	}
}

func buildColorSliderDefaults(active *theme.Theme) colorSliderDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	return colorSliderDefaults{
		Defaults:         style.Style{},
		TextSize:         14,
		HeaderGap:        4,
		TrackHeight:      20,
		ThumbSize:        16,
		ThumbBorderWidth: 3,
		FocusRingWidth:   2,
		Foreground:       sem.OverlayForeground,
		MutedForeground:  sem.MutedForeground,
		FocusColor:       sem.Ring,
		DisabledOpacity:  active.DisabledOpacityValue(),
		Motion:           active.Motion,
	}
}

func buildColorFieldDefaults(active *theme.Theme) colorFieldDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return colorFieldDefaults{
		Defaults:        style.Style{},
		Gap:             4,
		ErrorTextSize:   active.Typography.SmallSize,
		ErrorForeground: active.Semantic().Tone(theme.ToneDanger).Fill,
	}
}

func buildColorSwatchDefaults(active *theme.Theme) colorSwatchDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return colorSwatchDefaults{
		Defaults:       style.Style{},
		ExtraSmallSize: 16,
		SmallSize:      24,
		MediumSize:     32,
		LargeSize:      36,
		ExtraLargeSize: 40,
		SquareRadius:   6,
	}
}

func buildColorSwatchPickerDefaults(active *theme.Theme) colorSwatchPickerDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	return colorSwatchPickerDefaults{
		Defaults:                     style.Style{},
		Gap:                          8,
		ExtraSmallBorderWidth:        1,
		BorderWidth:                  2,
		LargeBorderWidth:             3,
		FocusRingWidth:               2,
		FocusRingGap:                 2,
		CheckStroke:                  1.5,
		SquareItemRadiusExtraSmall:   6,
		SquareItemRadiusSmall:        4,
		SquareItemRadius:             6,
		SquareSwatchRadiusExtraSmall: 6,
		SquareSwatchRadiusSmall:      4,
		SquareSwatchRadius:           4,
		SquareSelectedSmallRadius:    6,
		ShadowOpacity:                0,
		SelectedScale:                .77,
		SelectionDuration:            100 * time.Millisecond,
		CheckDuration:                150 * time.Millisecond,
		HoverDuration:                100 * time.Millisecond,
		FocusColor:                   sem.Ring,
		ShadowColor:                  sem.SurfaceShadow,
		Shadows:                      active.Shadows.Control,
		DisabledOpacity:              active.DisabledOpacityValue(),
		Motion:                       active.Motion,
		Swatch:                       buildColorSwatchDefaults(active),
	}
}
