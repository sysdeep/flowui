package spinner

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// spinnerDefaults is the cached, theme-derived appearance of one spinner
// configuration. The instance style is applied separately as the last cascade
// layer and never stored here.
type spinnerDefaults struct {
	root    style.Style
	metrics spinnerMetrics
	motion  theme.MotionTheme
}

type spinnerResolvedStyle struct {
	root    style.ResolvedStyle
	color   color.NRGBA
	metrics spinnerMetrics
	motion  theme.MotionTheme
}

type spinnerVisualKey struct {
	color Color
	size  Size
}

type spinnerVisualCache = computed.VisualCache[spinnerVisualKey, spinnerDefaults]

func (s Widget) visualDefaults(ctx *core.Context) spinnerDefaults {
	active := core.ActiveTheme(ctx)
	load := func() spinnerDefaults {
		return buildSpinnerDefaults(active, s.color, s.size)
	}
	if s.visualCache == nil {
		return load()
	}
	return s.visualCache.Resolve(active, spinnerVisualKey{color: s.color, size: s.size}, load)
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context) spinnerResolvedStyle {
	defaults := s.visualDefaults(ctx)
	root := computed.ResolveStatic(ctx, style.State{}, defaults.root, style.Style{}, style.Style{}, s.style)
	indicator := computed.ResolvePartStatic(
		ctx,
		style.PartIndicator,
		style.State{},
		defaults.root,
		style.Style{},
		style.Style{},
		s.style,
	)
	if computed.HasTransitions(root, indicator) {
		key := core.ClaimKey(ctx, core.KindStyle, "spinner")
		root = computed.ApplyTransitions(ctx, gtx, key, root)
		indicator = computed.ApplyPartTransitions(ctx, gtx, key, style.PartIndicator, indicator)
	}
	resolvedColor, ok := spinnerIndicatorColor(indicator)
	if !ok {
		panic("flowui/spinner: unresolved indicator color")
	}
	return spinnerResolvedStyle{
		root:    root,
		color:   resolvedColor,
		metrics: defaults.metrics,
		motion:  defaults.motion,
	}
}
