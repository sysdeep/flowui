package toggle

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// checkboxMetrics is the geometry and spacing consumed by checkbox layout.
type checkboxMetrics struct {
	FocusSpace          unit.Dp
	IndicatorSize       unit.Dp
	IndeterminateStroke unit.Dp
	CheckStroke         unit.Dp
	DescriptionGap      unit.Dp
	DescriptionIndent   unit.Dp
	LabelGap            unit.Dp
	RequiredMarkOffset  unit.Dp
}

// checkboxColors contains component-owned colors needed when the required mark
// is rendered without an explicit declaration.
type checkboxColors struct {
	RequiredMark         color.NRGBA
	DisabledRequiredMark color.NRGBA
}

// checkboxDefaults is the cached appearance and motion of one checkbox
// variant. It holds only theme-derived layers; instance declarations are
// resolved as the last cascade layer and are never part of the cache key.
type checkboxDefaults struct {
	Defaults          style.Style
	Metrics           checkboxMetrics
	Colors            checkboxColors
	Motion            theme.MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
}

// CheckboxVisualCache retains theme-derived checkbox defaults for compound
// widgets that draw many checkboxes without owning a CheckboxWidget state
// object. Its zero value is ready to use.
type CheckboxVisualCache struct {
	cache computed.VisualCacheMap[CheckboxVariant, checkboxDefaults]
}

// CheckboxVisualCacheStats makes cache behavior measurable in compound widget
// benchmarks without exposing declaration internals.
type CheckboxVisualCacheStats = computed.VisualCacheStats

// Stats reports a snapshot of cache hits, misses, and theme invalidations.
func (c *CheckboxVisualCache) Stats() CheckboxVisualCacheStats {
	if c == nil {
		return CheckboxVisualCacheStats{}
	}
	return c.cache.Stats()
}

// resolve returns the cached defaults for the active theme and variant.
func (c *CheckboxVisualCache) resolve(active *theme.Theme, variant CheckboxVariant) checkboxDefaults {
	if c == nil {
		return buildCheckboxDefaults(active, variant)
	}
	return c.cache.Resolve(active, variant, func() checkboxDefaults {
		return buildCheckboxDefaults(active, variant)
	})
}

type checkboxIndicatorStyle struct {
	off style.ResolvedStyle
	on  style.ResolvedStyle
}

// checkboxResolvedStyle is the appearance of one checkbox for a single
// frame. The theme-derived defaults are kept alongside the resolved parts so
// layout reads geometry and motion from one place.
type checkboxResolvedStyle struct {
	defaults    checkboxDefaults
	indicator   checkboxIndicatorStyle
	label       style.ResolvedStyle
	description style.ResolvedStyle
}

func (c CheckboxWidget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	cache *CheckboxVisualCache,
) checkboxResolvedStyle {
	defaults := cache.resolve(core.ActiveTheme(ctx), c.variant)
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults.Defaults, style.Style{}, style.Style{}, c.style)
	on := computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults.Defaults, style.Style{}, style.Style{}, c.style)
	label := computed.ResolvePartStatic(ctx, style.PartLabel, state, defaults.Defaults, style.Style{}, style.Style{}, c.style)
	description := computed.ResolvePartStatic(ctx, style.PartDescription, state, defaults.Defaults, style.Style{}, style.Style{}, c.style)
	off = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "checkbox-indicator-off"), style.PartIndicator, off)
	on = computed.ApplyPartTransitions(ctx, gtx, core.DerivedKey(ctx, key, "checkbox-indicator-on"), style.PartIndicator, on)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)

	return checkboxResolvedStyle{
		defaults:    defaults,
		indicator:   checkboxIndicatorStyle{off: off, on: on},
		label:       label,
		description: description,
	}
}
