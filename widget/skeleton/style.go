package skeleton

import (
	"image/color"

	"gioui.org/unit"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type skeletonDefaultsCache = computed.VisualCache[struct{}, skeletonDefaults]

type skeletonResolvedStyle struct {
	root      style.ResolvedStyle
	highlight style.ResolvedStyle
}

// defaultsWithCache resolves the theme-owned placeholder declaration. The
// instance declaration is deliberately absent from the key: it is the last
// cascade layer at resolve time and must not invalidate the cached defaults.
func (s Widget) defaultsWithCache(ctx *core.Context) skeletonDefaults {
	active := core.ActiveTheme(ctx)
	load := func() skeletonDefaults {
		return buildSkeletonDefaults(active)
	}
	if s.defaultsCache == nil {
		return load()
	}
	return s.defaultsCache.Resolve(active, struct{}{}, load)
}

func (s Widget) resolveStyle(ctx *core.Context, defaults skeletonDefaults) skeletonResolvedStyle {
	state := style.State{}
	pulse := style.Join(s.style, defaults.highlight)
	return skeletonResolvedStyle{
		root:      computed.ResolveStatic(ctx, state, defaults.root, style.Style{}, style.Style{}, s.style),
		highlight: computed.ResolveStatic(ctx, state, defaults.root, style.Style{}, style.Style{}, pulse),
	}
}

// skeletonFill extracts the concrete fill used for the pulse. Gradients are
// sampled at their midpoint, matching how other single-colour surfaces derive
// a representative colour from a paint declaration.
func skeletonFill(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Paint == nil {
		return color.NRGBA{}, false
	}
	brush, ok := computed.Brush(resolved.Paint.Background)
	if !ok {
		return color.NRGBA{}, false
	}
	return computed.ApplyOpacity(brush.ColorAt(.5), resolved.Paint.Opacity), true
}

// skeletonRadius returns the resolved corner radius. The placeholder clips
// with a uniform rounded rectangle, so per-corner declarations collapse to
// their largest corner.
func skeletonRadius(resolved style.ResolvedStyle) (unit.Dp, bool) {
	if resolved.Paint == nil {
		return 0, false
	}
	if radii := resolved.Paint.Radii; radii != nil {
		return max(radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft), true
	}
	if resolved.Paint.Radius != nil {
		return *resolved.Paint.Radius, true
	}
	return 0, false
}
