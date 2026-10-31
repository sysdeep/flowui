// Package skeleton provides passive placeholders while content is loading.
package skeleton

import (
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	width, height    unit.Dp
	radius           unit.Dp
	circle, animated bool
	hasHeight        bool
	hasRadius        bool
	style            style.Style
	defaultsCache    *skeletonDefaultsCache
}

// New creates a full-width placeholder with a gentle pulse. Its depth and
// corner radius come from the active theme.
func New() Widget { return Widget{animated: true, defaultsCache: &skeletonDefaultsCache{}} }
func (s Widget) Width(value unit.Dp) Widget {
	s.width = max(0, value)
	return s
}

// Height sets the placeholder depth; without it the theme's default applies.
func (s Widget) Height(value unit.Dp) Widget {
	s.height = max(0, value)
	s.hasHeight = true
	return s
}

// Radius sets the corner radius; without it the theme's default applies.
func (s Widget) Radius(value unit.Dp) Widget {
	s.radius = max(0, value)
	s.hasRadius = true
	return s
}

func (s Widget) Circle() Widget             { s.circle = true; return s }
func (s Widget) Animated(value bool) Widget { s.animated = value; return s }

// Style merges instance declarations into the placeholder. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (s Widget) Style(values ...style.Style) Widget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	defaults := s.defaultsWithCache(ctx)
	resolved := s.resolveStyle(ctx, defaults)

	// A placeholder keeps its requested shape even in a stretching parent.
	gtx.Constraints.Min = image.Point{}
	width := gtx.Constraints.Max.X
	if s.width > 0 {
		width = gtx.Dp(s.width)
	}
	height := s.height
	if !s.hasHeight {
		height = defaults.height
	}
	heightPx := gtx.Dp(height)
	if s.circle {
		heightPx = min(heightPx, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
		width = heightPx
	}
	size := gtx.Constraints.Constrain(image.Pt(width, heightPx))
	radius := s.radius
	if !s.hasRadius {
		radius = defaults.radius
		if value, ok := skeletonRadius(resolved.root); ok {
			radius = value
		}
	}
	radiusPx := min(gtx.Dp(radius), min(size.X, size.Y)/2)
	if s.circle {
		radiusPx = min(size.X, size.Y) / 2
	}
	fill, ok := skeletonFill(resolved.root)
	if !ok {
		panic("flowui/skeleton: unresolved fill color")
	}
	period := theme.ResolveMotionDuration(defaults.motion, defaults.period)
	if s.animated && period > 0 {
		highlight, ok := skeletonFill(resolved.highlight)
		if !ok {
			panic("flowui/skeleton: unresolved highlight color")
		}
		phase := float64(gtx.Now.UnixNano()%int64(period)) / float64(period)
		fill = animation.LerpColor(fill, highlight, float32((1-math.Cos(phase*2*math.Pi))/2))
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 60)})
	}
	paint.FillShape(gtx.Ops, fill, clip.UniformRRect(image.Rectangle{Max: size}, radiusPx).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

var _ core.Widget = Widget{}
