package radiogroup

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// radioMetrics is the geometry consumed by radio layout.
type radioMetrics struct {
	FocusSpace      unit.Dp
	ContentGap      unit.Dp
	VerticalGap     unit.Dp
	HorizontalGap   unit.Dp
	DescriptionGap  unit.Dp
	DotScale        float32
	DotPressedScale float32
	PressedScale    float32
}

// radioDefaults is the cached appearance and motion of one radio variant. It
// holds only theme-derived layers; instance declarations are resolved as the
// last cascade layer and are never part of the cache key.
type radioDefaults struct {
	Defaults          style.Style
	Metrics           radioMetrics
	Motion            theme.MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
	PressInDuration   time.Duration
	PressOutDuration  time.Duration
}

type radioDefaultsCache = computed.VisualCache[Variant, radioDefaults]

func resolveRadioDefaults(cache *radioDefaultsCache, active *theme.Theme, variant Variant) radioDefaults {
	return cache.Resolve(active, variant, func() radioDefaults {
		return buildRadioDefaults(active, variant)
	})
}

type radioStyle struct {
	indicatorOff style.ResolvedStyle
	indicatorOn  style.ResolvedStyle
	label        style.ResolvedStyle
	description  style.ResolvedStyle
	selected     float32
	focus        float32
	pressed      bool
}

func (r Widget) resolveItemStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, defaults radioDefaults) radioStyle {
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults.Defaults, style.Style{}, style.Style{}, r.style)
	on := computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults.Defaults, style.Style{}, style.Style{}, r.style)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, defaults.Defaults, style.Style{}, style.Style{}, r.style)
	description := computed.ResolvePartStatic(ctx, style.PartDescription, state, defaults.Defaults, style.Style{}, style.Style{}, r.style)
	off = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "off"), style.PartIndicator, off)
	on = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "on"), style.PartIndicator, on)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)
	return radioStyle{
		indicatorOff: off, indicatorOn: on,
		label: label, description: description,
		pressed: state.Pressed,
	}
}
