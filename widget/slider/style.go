package slider

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type sliderStyle struct {
	track  style.ResolvedStyle
	fill   style.ResolvedStyle
	thumb  style.ResolvedStyle
	label  style.ResolvedStyle
	visual sliderVisual
}

type sliderDefaultsCache = computed.VisualCache[Orientation, sliderVisual]

func resolveSliderDefaults(cache *sliderDefaultsCache, active *theme.Theme, orientation Orientation) sliderVisual {
	return cache.Resolve(active, orientation, func() sliderVisual {
		return buildSlider(active, orientation)
	})
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, visual sliderVisual) sliderStyle {
	parts := [4]style.Part{style.PartTrack, style.PartFill, style.PartThumb, style.PartLabel}
	var resolved [4]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:],
		resolved[:],
		visual.Defaults,
		style.Style{},
		style.Style{},
		s.style,
	)
	return sliderStyle{
		track:  resolved[0],
		fill:   resolved[1],
		thumb:  resolved[2],
		label:  resolved[3],
		visual: visual,
	}
}
