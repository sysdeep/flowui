package surface

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type surfaceVisual struct{ Defaults, Variant style.Style }

type surfaceVisualCache = computed.VisualCache[Variant, surfaceVisual]

func (s Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := s.visualWithCache(ctx)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		s.style,
	)
	if len(resolved.Transitions) != 0 {
		resolved = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "surface"), resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, s.child)
}

func (s Widget) visualWithCache(ctx *core.Context) surfaceVisual {
	active := core.ActiveTheme(ctx)
	load := func() surfaceVisual { return buildSurfaceVisual(active, s.variant) }
	if s.visualCache == nil {
		return load()
	}
	return s.visualCache.Resolve(active, s.variant, load)
}

func buildSurfaceVisual(active *theme.Theme, variant Variant) surfaceVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}

	sem := active.Semantic()
	background, foreground := surfaceSurfaceColors(sem, variant)
	value := style.Style{}.TextColor(style.Solid(foreground))
	if variant != Transparent {
		value = value.Background(style.Solid(background))
	}
	return surfaceVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(1).
			Opacity(1),
		Variant: value,
	}
}

func surfaceSurfaceColors(sem theme.Semantic, variant Variant) (background, foreground color.NRGBA) {
	switch variant {
	case Secondary:
		return sem.SurfaceSecondary, sem.SurfaceSecondaryForeground
	case Tertiary:
		return sem.SurfaceTertiary, sem.SurfaceTertiaryForeground
	case Transparent:
		return color.NRGBA{}, sem.Foreground
	default:
		return sem.Surface, sem.SurfaceForeground
	}
}
