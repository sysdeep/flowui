package navigation

import (
	"image"
	"strconv"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

const paginationStateSlotPagination = "pagination"

type paginationState struct {
	items      map[string]*paginationItemState
	frameItems map[string]struct{}
}

type paginationItemState struct {
	clickable widget.Clickable
	focus     animation.FocusAnimation
}

func (p PaginationWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithStyle(ctx, gtx)
}

func (p PaginationWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindPagination, p.key)
	value := core.UseState[paginationState](ctx, key, paginationStateSlotPagination)
	value.beginFrame()
	defer value.endFrame()
	visual := p.visualWithCache(ctx)

	semantic.LabelOp(ctx.Text(i18n.PaginationLabel)).Add(gtx.Ops)
	content := func(gtx layout.Context) layout.Dimensions {
		return p.layoutContent(ctx, gtx, value, visual)
	}
	layoutContent := func(gtx layout.Context) layout.Dimensions {
		if p.summary == nil {
			return content(gtx)
		}
		if gtx.Constraints.Max.X < gtx.Dp(visual.Metrics.CompactWidth) {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start, Gap: gtx.Dp(visual.Metrics.ContentGap)}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx, visual) }),
				layout.Rigid(content),
			)
		}
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx, visual) }),
			layout.Rigid(content),
		)
	}
	hovered, pressed, focused := false, false, false
	for _, item := range value.items {
		hovered = hovered || item.clickable.Hovered()
		pressed = pressed || item.clickable.Pressed()
		focused = focused || gtx.Focused(&item.clickable)
	}
	state := style.State{Hovered: hovered, Pressed: pressed, Focused: focused, Disabled: p.disabled || !gtx.Enabled(), Selected: true}
	resolved := computed.ResolveStatic(ctx, state, visual.Defaults, style.Style{}, style.Style{}, p.style)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions { return layoutContent(gtx) }))
}

func (p PaginationWidget) layoutSummary(ctx *core.Context, gtx layout.Context, visual paginationVisual) layout.Dimensions {
	resolved := computed.ResolveStatic(ctx, style.State{}, visual.Summary, style.Style{}, style.Style{}, p.style)
	return layoutui.Resolved(ctx, gtx, resolved, p.summary)
}

func (p PaginationWidget) layoutContent(ctx *core.Context, gtx layout.Context, value *paginationState, visual paginationVisual) layout.Dimensions {
	items := paginationPageItems(p.page, p.total, p.boundaries, p.siblings)
	children := make([]layout.FlexChild, 0, len(items)+2)
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("previous"), p.page-1, p.previousText(ctx), false, p.page <= 1, visual)
		}))
	}
	for _, page := range items {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if page == 0 {
				return p.layoutEllipsis(ctx, gtx, visual)
			}
			return p.layoutPage(ctx, gtx, value.item("page:"+strconv.Itoa(page)), page, page == p.page, visual)
		}))
	}
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("next"), p.page+1, p.nextText(ctx), true, p.page >= p.total, visual)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(visual.Metrics.ItemGap)}.Layout(gtx, children...)
}

func (p PaginationWidget) previousText(ctx *core.Context) string {
	if p.hasPreviousLabel {
		return p.previousLabel
	}
	return ctx.Text(i18n.PaginationPrevious)
}

func (p PaginationWidget) nextText(ctx *core.Context) string {
	if p.hasNextLabel {
		return p.nextLabel
	}
	return ctx.Text(i18n.PaginationNext)
}

func (p PaginationWidget) layoutPage(ctx *core.Context, gtx layout.Context, stateValue *paginationItemState, page int, active bool, visual paginationVisual) layout.Dimensions {
	size := gtx.Dp(visual.Metrics.Size)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	label := strconv.Itoa(page)
	return p.layoutButton(ctx, gtx, stateValue, page, active, false, label, PaginationPage, visual)
}

func (p PaginationWidget) layoutNav(ctx *core.Context, gtx layout.Context, stateValue *paginationItemState, target int, label string, next, disabled bool, visual paginationVisual) layout.Dimensions {
	size := min(gtx.Dp(visual.Metrics.Size), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = size
	gtx.Constraints.Max.Y = size
	kind := PaginationPrevious
	if next {
		kind = PaginationNext
	}
	return p.layoutButton(ctx, gtx, stateValue, target, false, disabled, label, kind, visual)
}

func (p PaginationWidget) layoutEllipsis(ctx *core.Context, gtx layout.Context, visual paginationVisual) layout.Dimensions {
	size := min(gtx.Dp(visual.Metrics.Size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	return p.layoutButton(ctx, gtx, nil, 0, false, true, "...", PaginationEllipsis, visual)
}

func (p PaginationWidget) layoutButton(ctx *core.Context, gtx layout.Context, item *paginationItemState, target int, active, disabled bool, label string, kind PaginationButtonKind, visual paginationVisual) layout.Dimensions {
	disabled = disabled || p.disabled || !gtx.Enabled()
	if item == nil {
		return paginationRenderButton(core.ActiveTheme(ctx), gtx, visual.RenderMetrics, paginationButtonRequest{Kind: kind, Label: label, Disabled: true})
	}
	return base.Button("").Clickable(&item.clickable).Label(label).
		Selected(active).Disabled(disabled).
		OnClick(func() {
			if target != p.page && p.onChange != nil {
				p.onChange(target)
			}
		}).
		Content(func(state base.ButtonState) core.Widget {
			return core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
				focus := item.focus.Opacity(gtx, state.FocusVisible && !state.Disabled, visual.Motion)
				return paginationRenderButton(core.ActiveTheme(ctx), gtx, visual.RenderMetrics, paginationButtonRequest{
					Kind: kind, Label: label, Active: active, Disabled: state.Disabled,
					Hovered: state.Hovered, Pressed: state.Pressed, FocusVisible: state.FocusVisible, Focus: focus,
				})
			})
		}).Layout(ctx, gtx)
}

func (s *paginationState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *paginationState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *paginationState) item(key string) *paginationItemState {
	if s.items == nil {
		s.items = make(map[string]*paginationItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(paginationItemState)
	s.items[key] = item
	return item
}
