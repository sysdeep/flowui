package toggle

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// switchMetrics is the geometry consumed around the switch control.
type switchMetrics struct {
	FocusSpace     unit.Dp
	ContentGap     unit.Dp
	DescriptionGap unit.Dp
}

// switchDefaults is the cached appearance and motion of one switch size. It
// holds only theme-derived layers; instance declarations are resolved as the
// last cascade layer and are never part of the cache key.
type switchDefaults struct {
	Defaults          style.Style
	Metrics           switchMetrics
	Motion            theme.MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
}

type switchVisualCache struct {
	cache computed.VisualCache[SwitchSize, switchDefaults]
}

func (c *switchVisualCache) resolve(active *theme.Theme, size SwitchSize) switchDefaults {
	if c == nil {
		return buildSwitchDefaults(active, size)
	}
	return c.cache.Resolve(active, size, func() switchDefaults {
		return buildSwitchDefaults(active, size)
	})
}

func (c *switchVisualCache) stats() computed.VisualCacheStats {
	if c == nil {
		return computed.VisualCacheStats{}
	}
	return c.cache.Stats()
}

type switchResolvedStyle struct {
	trackOff    style.ResolvedStyle
	trackOn     style.ResolvedStyle
	thumbOff    style.ResolvedStyle
	thumbOn     style.ResolvedStyle
	label       style.ResolvedStyle
	description style.ResolvedStyle
	// Keep the theme-derived bundle intact so runtime geometry and motion never
	// drift into a second set of fields.
	defaults switchDefaults
	selected float32
	focus    float32
}

func (s SwitchWidget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	cache *switchVisualCache,
) switchResolvedStyle {
	defaults := cache.resolve(core.ActiveTheme(ctx), s.size)
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	resolvePart := func(part style.Part, partState style.State) style.ResolvedStyle {
		return computed.ResolvePartStatic(ctx, part, partState, defaults.Defaults, style.Style{}, style.Style{}, s.style)
	}
	trackOff := resolvePart(style.PartTrack, offState)
	trackOn := resolvePart(style.PartTrack, onState)
	thumbOff := resolvePart(style.PartThumb, offState)
	thumbOn := resolvePart(style.PartThumb, onState)
	var label, description style.ResolvedStyle
	if s.label != "" {
		label = resolvePart(style.PartLabel, state)
	}
	if s.description != "" {
		description = resolvePart(style.PartDescription, state)
	}
	offOwner := core.DerivedKey(ctx, key, "switch-off")
	onOwner := core.DerivedKey(ctx, key, "switch-on")
	trackOff = computed.ApplyPartTransitions(ctx, gtx, offOwner, style.PartTrack, trackOff)
	trackOn = computed.ApplyPartTransitions(ctx, gtx, onOwner, style.PartTrack, trackOn)
	thumbOff = computed.ApplyPartTransitions(ctx, gtx, offOwner, style.PartThumb, thumbOff)
	thumbOn = computed.ApplyPartTransitions(ctx, gtx, onOwner, style.PartThumb, thumbOn)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)

	return switchResolvedStyle{
		trackOff: trackOff, trackOn: trackOn,
		thumbOff: thumbOff, thumbOn: thumbOn,
		label: label, description: description, defaults: defaults,
	}
}
