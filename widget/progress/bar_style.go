package progress

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// progressStyleKey holds only the inputs that change the cached defaults.
// Instance style and the disabled/value state resolve outside the cache.
type progressStyleKey struct {
	color BarColor
	size  BarSize
}

type progressBarStyleCache = computed.VisualCache[progressStyleKey, progressBarDefaults]

type progressCircleStyleCache = computed.VisualCache[progressStyleKey, progressCircleDefaults]

type progressBarResolvedStyle struct {
	root     style.ResolvedStyle
	track    style.ResolvedStyle
	fill     style.ResolvedStyle
	label    style.ResolvedStyle
	defaults progressBarDefaults
}

func (p BarWidget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, cache *progressBarStyleCache) progressBarResolvedStyle {
	active := core.ActiveTheme(ctx)
	declarations := cache.Resolve(active, progressStyleKey{color: p.color, size: p.size}, func() progressBarDefaults {
		return p.defaultStyles(active)
	})
	state := style.State{Disabled: p.disabled}
	parts := [...]style.Part{style.PartRoot, style.PartTrack, style.PartFill, style.PartLabel}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:],
		resolved[:],
		declarations.defaults,
		declarations.variant,
		declarations.size,
		p.style,
	)
	return progressBarResolvedStyle{
		root:     resolved[0],
		track:    resolved[1],
		fill:     resolved[2],
		label:    resolved[3],
		defaults: declarations,
	}
}
