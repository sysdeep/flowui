package modal

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

func (m Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, state *modalState, progress float32, contentEnabled bool, defaults modalDefaults) layout.Dimensions {
	size := modalOverlaySize(gtx)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(size)

	styleState := style.State{Disabled: !contentEnabled, Open: state.open}
	backdrop := computed.ResolvePartWithKey(
		ctx, gtx, core.FullKey(ctx, m.key), style.PartBackdrop, styleState,
		defaults.defaults, style.Style{}, style.Style{}, m.style,
	)
	if backdrop.Paint == nil {
		backdrop.Paint = &style.PaintStyle{}
	} else {
		// ResolveStatic may return cache-owned immutable data. Opacity is the
		// only field changed for the overlay pass.
		paintCopy := *backdrop.Paint
		backdrop.Paint = &paintCopy
	}
	backdropOpacity := progress
	if backdrop.Paint.Opacity != nil {
		backdropOpacity *= *backdrop.Paint.Opacity
	}
	backdrop.Paint.Opacity = &backdropOpacity
	layoutui.Resolved(ctx, gtx, backdrop, nil)
	contentGtx := gtx
	if !contentEnabled {
		contentGtx = contentGtx.Disabled()
	}
	var last event.Tag
	if m.showCloseButton() && contentEnabled {
		last = &state.close
	}
	state.dialog.LayoutStart(ctx, gtx, last)

	dialogGtx := contentGtx
	dialogGtx.Constraints = m.dialogConstraints(ctx, gtx, size, defaults)
	macro := op.Record(gtx.Ops)
	dialogDims, dialogPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutDialogFrame(ctx, dialogGtx, state, defaults)
	})
	dialogCall := macro.Stop()

	dialogPos := m.dialogPosition(ctx, gtx, size, dialogDims.Size, defaults)
	dialogRect := image.Rectangle{
		Min: dialogPos,
		Max: dialogPos.Add(dialogDims.Size),
	}
	motion := m.dialogMotion(ctx, gtx, dialogRect, progress, state.opening(), defaults)
	dialogPlacement.PlaceTransform(motion.transform.Mul(f32.AffineId().Offset(f32.Pt(float32(dialogPos.X), float32(dialogPos.Y)))))
	dialogPlacement.SetOpacity(motion.opacity)
	state.dialog.LayoutOutside(gtx, size, false, core.AffineRectBounds(dialogRect, motion.transform))
	transform := op.Affine(motion.transform).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, motion.opacity)
	offset := op.Offset(dialogPos).Push(gtx.Ops)
	state.dialog.LayoutSurface(gtx, image.Rectangle{Max: dialogDims.Size})
	dialogCall.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	transform.Pop()
	return layout.Dimensions{Size: size}
}

func modalOverlaySize(gtx layout.Context) image.Point {
	size := gtx.Constraints.Max
	if size.X < 0 {
		size.X = 0
	}
	if size.Y < 0 {
		size.Y = 0
	}
	return size
}

func (m Widget) dialogConstraints(ctx *core.Context, gtx layout.Context, size image.Point, defaults modalDefaults) layout.Constraints {
	margin := m.dialogMargin(ctx, gtx, size, defaults)
	available := image.Pt(max(size.X-margin*2, 0), max(size.Y-margin*2, 0))
	if m.size == Full {
		return layout.Exact(size)
	}
	if m.size == Cover {
		return layout.Exact(available)
	}

	width := min(modalWidth(gtx, m.size, defaults), available.X)
	if width <= 0 {
		width = available.X
	}
	return layout.Constraints{
		Min: image.Pt(width, 0),
		Max: image.Pt(width, available.Y),
	}
}

func modalWidth(gtx layout.Context, size Size, defaults modalDefaults) int {
	metrics := defaults.metrics
	switch size {
	case XSmall:
		return gtx.Dp(metrics.XSmallWidth)
	case Small:
		return gtx.Dp(metrics.SmallWidth)
	case Large:
		return gtx.Dp(metrics.LargeWidth)
	default:
		return gtx.Dp(metrics.MediumWidth)
	}
}

func (m Widget) dialogMargin(_ *core.Context, gtx layout.Context, size image.Point, defaults modalDefaults) int {
	metrics := defaults.metrics
	margin := gtx.Dp(metrics.Margin)
	if size.X >= gtx.Dp(metrics.DesktopBreakpoint) {
		margin = gtx.Dp(metrics.DesktopMargin)
	}
	if m.size == Full {
		return 0
	}
	return min(max(margin, 0), min(size.X, size.Y)/2)
}

func (m Widget) dialogPosition(ctx *core.Context, gtx layout.Context, overlay, dialog image.Point, defaults modalDefaults) image.Point {
	margin := m.dialogMargin(ctx, gtx, overlay, defaults)
	x := max((overlay.X-dialog.X)/2, 0)
	placement := m.resolvedPlacement(gtx, overlay, defaults)
	var y int
	switch placement {
	case Top:
		y = margin
	case Bottom:
		y = overlay.Y - margin - dialog.Y
	default:
		y = (overlay.Y - dialog.Y) / 2
	}
	if y < margin && m.size != Full {
		y = margin
	}
	if y+dialog.Y > overlay.Y-margin && m.size != Full {
		y = max(overlay.Y-margin-dialog.Y, 0)
	}
	return image.Pt(x, max(y, 0))
}

func (m Widget) resolvedPlacement(gtx layout.Context, size image.Point, defaults modalDefaults) Placement {
	if m.placement != Auto {
		return m.placement
	}
	if size.X < gtx.Dp(defaults.metrics.DesktopBreakpoint) {
		return Bottom
	}
	return Center
}

type modalDialogMotion struct {
	transform f32.Affine2D
	opacity   float32
}

func (m Widget) dialogMotion(_ *core.Context, gtx layout.Context, rect image.Rectangle, progress float32, opening bool, defaults modalDefaults) modalDialogMotion {
	metrics := defaults.metrics
	transform := f32.AffineId()
	center := f32.Pt(float32(rect.Min.X)+float32(rect.Dx())/2, float32(rect.Min.Y)+float32(rect.Dy())/2)
	scale := float32(1)

	switch m.resolvedAnimation() {
	case AnimationFade:
	case AnimationSlideDown:
		distance := float32(gtx.Dp(metrics.AnimationDistance))
		transform = transform.Offset(f32.Pt(0, -distance*(1-progress)))
	case AnimationSlideUp:
		distance := float32(gtx.Dp(metrics.AnimationDistance))
		transform = transform.Offset(f32.Pt(0, distance*(1-progress)))
	case AnimationBounceScale:
		if opening {
			scale = modalDialogBounceScale(
				metrics.AnimationScale,
				metrics.AnimationBounceScale,
				metrics.AnimationBounceSettleScale,
				progress,
			)
		} else {
			scale = modalDialogScale(metrics.AnimationScale, progress)
		}
	case AnimationZoomOut:
		if opening {
			scale = modalDialogZoomOutScale(metrics.AnimationZoomScale, progress)
		} else {
			scale = modalDialogScale(metrics.AnimationScale, progress)
		}
	case AnimationPop:
		if opening {
			scale = modalDialogPopScale(
				metrics.AnimationPopStartScale,
				metrics.AnimationPopOvershootScale,
				metrics.AnimationPopSettleScale,
				progress,
			)
		} else {
			scale = modalDialogScale(metrics.AnimationScale, progress)
		}
	default:
		scale = modalDialogScale(metrics.AnimationScale, progress)
	}
	if scale != 1 {
		transform = transform.Scale(center, f32.Pt(scale, scale))
	}
	return modalDialogMotion{
		transform: transform,
		opacity:   progress,
	}
}

func modalDialogScale(startScale, progress float32) float32 {
	if startScale <= 0 || startScale > 1 {
		startScale = 1
	}
	return startScale + (1-startScale)*progress
}

func modalDialogBounceScale(startScale, overshoot, settleScale, progress float32) float32 {
	if startScale <= 0 || startScale > 1 {
		startScale = 1
	}
	if overshoot <= 1 {
		overshoot = 1
	}
	if settleScale <= 0 || settleScale > 1 {
		settleScale = 1
	}
	progress = clampProgress(progress)
	switch {
	case progress < 0.65:
		return animation.LerpFloat(startScale, overshoot, modalEaseOutCubic(progress/0.65))
	case progress < 0.86:
		return animation.LerpFloat(overshoot, settleScale, animation.EaseSmoothstep((progress-0.65)/0.21))
	default:
		return animation.LerpFloat(settleScale, 1, modalEaseOutCubic((progress-0.86)/0.14))
	}
}

func modalDialogZoomOutScale(startScale, progress float32) float32 {
	if startScale <= 1 {
		startScale = 1
	}
	return animation.LerpFloat(startScale, 1, modalEaseOutCubic(progress))
}

func modalDialogPopScale(startScale, overshoot, settleScale, progress float32) float32 {
	if startScale <= 0 || startScale > 1 {
		startScale = 1
	}
	if overshoot <= 1 {
		overshoot = 1
	}
	if settleScale <= 0 || settleScale > 1 {
		settleScale = 1
	}
	progress = clampProgress(progress)
	switch {
	case progress < 0.58:
		return animation.LerpFloat(startScale, overshoot, modalEaseOutCubic(progress/0.58))
	case progress < 0.82:
		return animation.LerpFloat(overshoot, settleScale, animation.EaseSmoothstep((progress-0.58)/0.24))
	default:
		return animation.LerpFloat(settleScale, 1, modalEaseOutCubic((progress-0.82)/0.18))
	}
}

func modalEaseOutCubic(progress float32) float32 {
	progress = clampProgress(progress)
	inverse := 1 - progress
	return 1 - inverse*inverse*inverse
}

func clampProgress(progress float32) float32 {
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func (m Widget) layoutDialogFrame(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) layout.Dimensions {
	return m.layoutDialog(ctx, gtx, state, defaults)
}

func (m Widget) layoutDialog(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) layout.Dimensions {
	focused := gtx.Focused(&state.close)
	styleState := style.State{
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &state.close, focused),
		Disabled:     !gtx.Enabled(),
		Open:         state.open,
	}
	resolved := computed.ResolvePartWithKey(
		ctx, gtx, core.FullKey(ctx, m.key), style.PartRoot, styleState,
		defaults.defaults, style.Style{}, style.Style{}, m.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return m.layoutDialogSurface(ctx, gtx, state, resolved, defaults)
	}))
}

func (m Widget) layoutDialogSurface(ctx *core.Context, gtx layout.Context, state *modalState, resolved style.ResolvedStyle, defaults modalDefaults) layout.Dimensions {
	dims := m.layoutDialogContent(ctx, gtx, state, defaults)
	if m.showCloseButton() {
		padding := dialogPadding(resolved)
		left := gtx.Dp(padding.Left)
		right := gtx.Dp(padding.Right)
		top := gtx.Dp(padding.Top)
		bottom := gtx.Dp(padding.Bottom)
		outerSize := dims.Size.Add(image.Pt(left+right, top+bottom))
		buttonGtx := gtx
		stack := op.Offset(image.Pt(-left, -top)).Push(gtx.Ops)
		m.layoutCloseButton(ctx, buttonGtx, state, outerSize, defaults)
		stack.Pop()
	}
	return dims
}

func (m Widget) layoutDialogContent(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) layout.Dimensions {
	if m.scroll == ScrollOutside {
		state.outsideList.Axis = layout.Vertical
		state.outsideList.ScrollAnyAxis = false
		return layoutui.TrackedListWithVisualOutset(ctx, gtx, &state.outsideList, &state.outsideVisualOutset, 1, func(gtx layout.Context, _ int) layout.Dimensions {
			return m.layoutDialogSections(ctx, gtx, state, defaults)
		})
	}
	return m.layoutDialogSections(ctx, gtx, state, defaults)
}

func (m Widget) layoutDialogSections(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) layout.Dimensions {
	sectionGtx := gtx
	sectionGtx.Constraints.Min = image.Point{}

	headerCall, headerDims, headerPlacement := m.recordHeader(ctx, sectionGtx, defaults)
	footerCall, footerDims, footerPlacement := m.recordFooter(ctx, sectionGtx)
	bodyGap, footerGap := m.dialogGaps(sectionGtx, headerDims.Size.Y > 0, footerDims.Size.Y > 0, defaults)

	bodyGtx := sectionGtx
	maxBodyY := max(gtx.Constraints.Max.Y-headerDims.Size.Y-footerDims.Size.Y-bodyGap-footerGap, 0)
	bodyGtx.Constraints.Max.Y = maxBodyY
	bodyCall, bodyDims, bodyPlacement := m.recordBody(ctx, bodyGtx, state, defaults)

	width := max(headerDims.Size.X, max(bodyDims.Size.X, footerDims.Size.X))
	height := headerDims.Size.Y + bodyDims.Size.Y + footerDims.Size.Y + bodyGap + footerGap
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	y := 0
	if headerDims.Size.Y > 0 {
		headerPlacement.PlaceOffset(image.Point{})
		headerCall.Add(gtx.Ops)
		y += headerDims.Size.Y
	}
	if bodyDims.Size.Y > 0 {
		y += bodyGap
		bodyPlacement.PlaceOffset(image.Pt(0, y))
		stack := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
		bodyCall.Add(gtx.Ops)
		stack.Pop()
		y += bodyDims.Size.Y
	}
	if footerDims.Size.Y > 0 {
		y += footerGap
		pos := m.dialogFooterPosition(size, footerDims.Size, y)
		footerPlacement.PlaceOffset(pos)
		stack := op.Offset(pos).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		stack.Pop()
	}
	if headerDims.Size.Y == 0 {
		headerPlacement.PlaceOffset(image.Point{})
	}
	if bodyDims.Size.Y == 0 {
		bodyPlacement.PlaceOffset(image.Pt(0, y))
	}
	if footerDims.Size.Y == 0 {
		footerPlacement.PlaceOffset(image.Pt(0, y))
	}
	return layout.Dimensions{Size: size}
}

func (m Widget) dialogFooterPosition(content, footer image.Point, y int) image.Point {
	if m.size == Cover || m.size == Full {
		y = max(y, content.Y-footer.Y)
	}
	return image.Pt(max(content.X-footer.X, 0), y)
}

func (m Widget) dialogGaps(gtx layout.Context, hasHeader, hasFooter bool, defaults modalDefaults) (bodyGap, footerGap int) {
	if hasHeader && m.body != nil {
		bodyGap = gtx.Dp(defaults.metrics.BodyGap)
	}
	if hasFooter {
		if m.body != nil || hasHeader {
			footerGap = gtx.Dp(defaults.metrics.SectionGap)
		} else {
			footerGap = gtx.Dp(defaults.metrics.FooterGap)
		}
	}
	return bodyGap, footerGap
}

func (m Widget) recordHeader(ctx *core.Context, gtx layout.Context, defaults modalDefaults) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutHeader(ctx, gtx, defaults)
	})
	return macro.Stop(), dims, placement
}

func (m Widget) layoutHeader(ctx *core.Context, gtx layout.Context, defaults modalDefaults) layout.Dimensions {
	if m.header != nil {
		return m.header.Layout(ctx, gtx)
	}
	if m.title == "" && m.icon == nil {
		return layout.Dimensions{}
	}
	children := make([]layout.FlexChild, 0, 2)
	if m.icon != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutIcon(ctx, gtx, defaults)
		}))
	}
	if m.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			resolved := computed.ResolveStatic(ctx, style.State{}, defaults.header, style.Style{}, style.Style{}, m.style)
			value := text.WithResolvedText(text.New(m.title), resolved.Text)
			return value.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(defaults.metrics.HeaderGap),
	}.Layout(gtx, children...)
}

func (m Widget) layoutIcon(ctx *core.Context, gtx layout.Context, defaults modalDefaults) layout.Dimensions {
	size := image.Pt(gtx.Dp(defaults.metrics.IconSize), gtx.Dp(defaults.metrics.IconSize))
	gtx.Constraints = layout.Exact(size)
	if size.X > 0 && size.Y > 0 {
		radius := min(size.X, size.Y) / 2
		paint.FillShape(gtx.Ops, defaults.iconBackground, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
	}
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return m.icon.Layout(ctx, gtx)
	})
}

func (m Widget) recordBody(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutBody(ctx, gtx, state, defaults)
	})
	return macro.Stop(), dims, placement
}

func (m Widget) layoutBody(ctx *core.Context, gtx layout.Context, state *modalState, defaults modalDefaults) layout.Dimensions {
	if m.body == nil {
		return layout.Dimensions{}
	}
	body := m.styleBody(ctx, m.body, defaults)
	if m.scroll == ScrollOutside {
		return body.Layout(ctx, gtx)
	}
	state.bodyList.Axis = layout.Vertical
	state.bodyList.ScrollAnyAxis = false
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.bodyList, &state.bodyBar, &state.bodyVisualOutset, 1, !gtx.Enabled(), false, func(gtx layout.Context, _ int) layout.Dimensions {
		return body.Layout(ctx, gtx)
	})
}

func (m Widget) styleBody(ctx *core.Context, body core.Widget, defaults modalDefaults) core.Widget {
	value, ok := body.(text.Widget)
	if !ok {
		return body
	}
	resolved := computed.ResolveStatic(ctx, style.State{}, defaults.body, style.Style{}, style.Style{}, m.style)
	return text.WithResolvedText(value, resolved.Text)
}

func (m Widget) recordFooter(ctx *core.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutFooter(ctx, gtx)
	})
	return macro.Stop(), dims, placement
}

func (m Widget) layoutFooter(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if m.footer == nil {
		return layout.Dimensions{}
	}
	return m.footer.Layout(ctx, gtx)
}

func (m Widget) layoutCloseButton(ctx *core.Context, gtx layout.Context, modalStateValue *modalState, dialogSize image.Point, defaults modalDefaults) {
	closeSize := gtx.Dp(defaults.metrics.CloseButtonSize)
	size := image.Pt(closeSize, closeSize)
	inset := gtx.Dp(defaults.metrics.CloseInset)
	pos := image.Pt(max(dialogSize.X-inset-size.X, 0), min(inset, max(dialogSize.Y-size.Y, 0)))

	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(size)
	stack := op.Offset(pos).Push(gtx.Ops)
	button.LayoutCloseButtonWithClickableNoEvents(
		*modalStateValue.closeButtonWidget(),
		ctx,
		buttonGtx,
		&modalStateValue.close,
	)
	stack.Pop()
}
