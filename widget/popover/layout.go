package popover

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/text"
)

func (p Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, state *popoverState, trigger image.Rectangle, progress float32, contentEnabled bool, visual popoverVisual) layout.Dimensions {
	if trigger.Empty() {
		return layout.Dimensions{}
	}
	overlaySize := popoverOverlaySize(gtx)
	if overlaySize.X <= 0 || overlaySize.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx
	if !contentEnabled {
		panelGtx = panelGtx.Disabled()
	}
	panelGtx.Constraints = p.panelConstraints(gtx, overlaySize, visual)
	panelCall, panelDims, panelPlacement := p.recordPanel(ctx, panelGtx, visual)
	result := p.resolvedPosition(gtx, trigger, panelDims.Size, overlaySize, p.placement, visual)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position
	panelRect := image.Rectangle{Min: panelPos, Max: panelPos.Add(panelDims.Size)}
	var arrowSize image.Point
	var arrowRect image.Rectangle
	if p.showArrow() {
		arrowSize = image.Pt(gtx.Dp(visual.Metrics.ArrowWidth), gtx.Dp(visual.Metrics.ArrowHeight))
		arrowRect = popoverArrowRect(image.Rectangle{Max: panelDims.Size}, placement, arrowSize)
	}

	panelAffine := popoverPanelAffine(panelRect, progress, placement, visual)
	panelOffset := panelPos.Add(popoverSlideOffset(gtx, progress, placement, visual))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	protected := []image.Rectangle{
		core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform),
	}
	if !arrowRect.Empty() {
		protected = append(protected, core.AffineRectBounds(arrowRect, panelTransform))
	}
	state.popup.LayoutOutside(gtx, overlaySize, false, protected...)
	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	state.popup.LayoutSurface(gtx, image.Rectangle{Max: panelDims.Size})
	panelCall.Add(gtx.Ops)
	if !arrowRect.Empty() {
		resolved := p.resolvePanelStyle(ctx, gtx, style.State{Disabled: !contentEnabled, Open: true}, visual)
		// Paint over the panel edge so the arrow joins its border without a seam.
		popoverArrow(
			gtx,
			placement.Placement().Side,
			panelDims.Size,
			arrowSize,
			resolved,
			visual.Surface,
		)
		state.popup.LayoutArrow(gtx, arrowRect)
	}
	offset.Pop()
	opacity.Pop()
	transform.Pop()

	return layout.Dimensions{Size: overlaySize}
}

func popoverOverlaySize(gtx layout.Context) image.Point {
	size := gtx.Constraints.Max
	if size.X < 0 {
		size.X = 0
	}
	if size.Y < 0 {
		size.Y = 0
	}
	return size
}

func (p Widget) panelConstraints(gtx layout.Context, overlay image.Point, visual popoverVisual) layout.Constraints {
	maxWidth := gtx.Dp(visual.Metrics.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlay.X {
		maxWidth = overlay.X
	}
	return layout.Constraints{
		Max: image.Pt(maxWidth, overlay.Y),
	}
}

func (p Widget) recordPanel(ctx *core.Context, gtx layout.Context, visual popoverVisual) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return p.layoutPanel(ctx, gtx, visual)
	})
	return macro.Stop(), dims, placement
}

func (p Widget) layoutPanel(ctx *core.Context, gtx layout.Context, visual popoverVisual) layout.Dimensions {
	state := style.State{Disabled: !gtx.Enabled(), Open: p.open}
	resolved := p.resolvePanelStyle(ctx, gtx, state, visual)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return p.layoutDialog(ctx, gtx, visual)
	}))
}

func (p Widget) resolvePanelStyle(ctx *core.Context, gtx layout.Context, state style.State, visual popoverVisual) style.ResolvedStyle {
	key := core.DerivedKey(ctx, core.FullKey(ctx, p.key), "style")
	return computed.Resolve(ctx, gtx, key, state, visual.Defaults, style.Style{}, style.Style{}, p.style)
}

func (p Widget) layoutDialog(ctx *core.Context, gtx layout.Context, visual popoverVisual) layout.Dimensions {
	if p.heading == "" {
		return p.layoutContent(ctx, gtx, visual)
	}
	gap := gtx.Dp(visual.Metrics.SectionGap)
	var headingDims layout.Dimensions
	var contentPlacement core.OverlayPlacement
	dims := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			heading := text.WithDefaults(text.New(p.heading), visual.Heading)
			headingDims = heading.Layout(ctx, gtx)
			return headingDims
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var contentDims layout.Dimensions
			contentDims, contentPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return p.layoutContent(ctx, gtx, visual)
			})
			return contentDims
		}),
	)
	contentPlacement.PlaceOffset(image.Pt(0, headingDims.Size.Y+gap))
	return dims
}

func (p Widget) layoutContent(ctx *core.Context, gtx layout.Context, visual popoverVisual) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	return p.styleContent(ctx, p.content, visual).Layout(ctx, gtx)
}

func (p Widget) styleContent(ctx *core.Context, content core.Widget, visual popoverVisual) core.Widget {
	value, ok := content.(text.Widget)
	if ok {
		return text.WithDefaults(value, visual.Content)
	}
	return content
}

func (p Widget) resolvedPosition(gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point, placement core.PopoverPlacement, visual popoverVisual) core.PositionResult {
	return core.ResolvePosition(core.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           p.offsetPx(gtx, visual),
		Placement:        placement.Placement(),
		Flip:             p.flipEnabled(),
		AvoidOverflow:    p.overflowAvoidanceEnabled(),
	})
}

func (p Widget) offsetPx(gtx layout.Context, visual popoverVisual) int {
	if p.hasOffset {
		return gtx.Dp(unit.Dp(p.offset))
	}
	return gtx.Dp(visual.Metrics.Offset)
}

func popoverPanelAffine(rect image.Rectangle, progress float32, placement core.PopoverPlacement, visual popoverVisual) f32.Affine2D {
	origin := popoverTransformOrigin(rect, placement)
	scale := visual.Metrics.AnimationScale
	if scale <= 0 || scale > 1 {
		scale = .90
	}
	scale += (1 - scale) * progress
	return f32.AffineId().Scale(origin, f32.Pt(scale, scale))
}

func popoverTransformOrigin(rect image.Rectangle, placement core.PopoverPlacement) f32.Point {
	return core.TransformOrigin(rect, placement.Placement())
}

func popoverSlideOffset(gtx layout.Context, progress float32, placement core.PopoverPlacement, visual popoverVisual) image.Point {
	delta := gtx.Dp(visual.Metrics.AnimationDistance)
	return core.SlideOffset(delta, progress, placement.Placement())
}

func popoverArrowRect(panel image.Rectangle, placement core.PopoverPlacement, arrow image.Point) image.Rectangle {
	if panel.Empty() || arrow.X <= 0 || arrow.Y <= 0 {
		return image.Rectangle{}
	}
	switch placement.Placement().Side {
	case core.SideTop:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Max.Y, x+arrow.X, panel.Max.Y+arrow.Y)
	case core.SideLeft:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Max.X, y, panel.Max.X+arrow.Y, y+arrow.X)
	case core.SideRight:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Min.X-arrow.Y, y, panel.Min.X, y+arrow.X)
	default:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Min.Y-arrow.Y, x+arrow.X, panel.Min.Y)
	}
}
