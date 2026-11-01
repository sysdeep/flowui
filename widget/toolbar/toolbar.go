package toolbar

import (
	"image"
	"strconv"

	"gioui.org/io/event"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/menu"
)

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// Widget groups related controls and provides directional keyboard navigation.
type Widget struct {
	style       style.Style
	children    []core.Widget
	commands    []command.Command
	key         string
	commandMode bool
	orientation Orientation
	attached    bool
	disabled    bool
	loopFocus   bool
	alt         string
	visualCache *toolbarVisualCache
}

func New(children ...core.Widget) Widget {
	return Widget{key: core.AutoKey("toolbar"), children: children, loopFocus: true, visualCache: &toolbarVisualCache{}}
}

// Key sets the stable identity used by style transitions and child separators.
func (w Widget) Key(key string) Widget {
	w.key = key
	return w
}

// Commands creates a horizontal command toolbar. Commands that do not fit
// are moved from the end into a More menu. The key identifies its retained
// overflow menu state.
func Commands(key string, commands ...command.Command) Widget {
	if key == "" {
		panic("flowui: empty command toolbar key")
	}
	return Widget{key: key, commands: append([]command.Command(nil), commands...), commandMode: true, loopFocus: true, visualCache: &toolbarVisualCache{}}
}

func (w Widget) Orientation(orientation Orientation) Widget {
	w.orientation = orientation
	return w
}

func (w Widget) Attached(attached bool) Widget {
	w.attached = attached
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) LoopFocus(loop bool) Widget {
	w.loopFocus = loop
	return w
}

func (w Widget) Alt(alt string) Widget {
	w.alt = alt
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithStyle(ctx, gtx)
}

func (w Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if !gtx.Enabled() {
		w.disabled = true
	}
	gtx.Constraints.Min = image.Point{}
	visual := w.visualWithCache(ctx)
	resolved := w.resolveStyle(ctx, gtx, visual)
	group := &core.FocusGroup{}
	dims := layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		restoreGroup := core.PushFocusGroup(ctx, group)
		defer restoreGroup()
		if w.disabled {
			gtx = gtx.Disabled()
		}
		macro := op.Record(gtx.Ops)
		dims := w.layoutChildren(ctx, gtx, visual)
		call := macro.Stop()
		root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.EnabledOp(!w.disabled).Add(gtx.Ops)
		label := w.alt
		if label == "" {
			label = ctx.Text(i18n.ToolbarDefault)
		}
		semantic.DescriptionOp(label).Add(gtx.Ops)
		call.Add(gtx.Ops)
		root.Pop()
		return dims
	}))

	w.updateKeys(ctx, gtx, group.Items)
	return dims
}

func (w Widget) layoutChildren(ctx *core.Context, gtx layout.Context, visual toolbarVisual) layout.Dimensions {
	if w.commandMode {
		return w.layoutCommands(ctx, gtx, visual)
	}
	axis := layout.Horizontal
	align := layout.Middle
	if w.orientation == Vertical {
		axis = layout.Vertical
		align = layout.Start
	}
	children := append([]core.Widget(nil), w.children...)
	for index, child := range children {
		if separator, ok := child.(SeparatorWidget); ok {
			separator.orientation = w.orientation
			separator.key = core.DerivedKey(ctx, core.FullKey(ctx, w.key), "separator:"+strconv.Itoa(index))
			separator.visual = &visual
			children[index] = separator
			continue
		}
	}
	return layoutui.TrackedFlex(
		ctx,
		gtx,
		axis,
		visual.Metrics.Gap,
		align,
		children...,
	)
}

func (w Widget) layoutCommands(ctx *core.Context, gtx layout.Context, visual toolbarVisual) layout.Dimensions {
	if w.orientation == Vertical {
		panic("flowui: command toolbar only supports horizontal orientation")
	}
	commandWidgets := make([]core.Widget, len(w.commands))
	commandSizes := make([]int, len(w.commands))
	measureGtx := gtx
	measureGtx.Constraints.Min = image.Point{}
	for index, item := range w.commands {
		menuItem := command.MenuItemFor(ctx, item)
		commandWidgets[index] = command.Button(core.DerivedKey(ctx, w.key, "command:"+menuItem.Key), item)
		commandSizes[index] = core.MeasureWidget(ctx, measureGtx, commandWidgets[index]).Size.X
	}
	if len(commandWidgets) == 0 {
		return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, visual.Metrics.Gap, layout.Middle)
	}

	more := button.New(ctx.Text(i18n.ToolbarMore)).Key(core.DerivedKey(ctx, w.key, "overflow-trigger")).Variant(button.Tertiary)
	moreWidth := core.MeasureWidget(ctx, measureGtx, more).Size.X
	visible := commandVisibleCount(commandSizes, gtx.Constraints.Max.X, max(gtx.Dp(visual.Metrics.Gap), 0), moreWidth)

	children := append([]core.Widget(nil), commandWidgets[:visible]...)
	if visible < len(commandWidgets) {
		items := make([]dropdown.Item, 0, len(commandWidgets)-visible)
		for _, item := range w.commands[visible:] {
			items = append(items, command.MenuItemFor(ctx, item))
		}
		popup := menu.New(items).Key(core.DerivedKey(ctx, w.key, "overflow:menu")).AutoWidth()
		children = append(children, dropdown.New(more, nil).Key(core.DerivedKey(ctx, w.key, "overflow")).Menu(popup))
	}
	return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, visual.Metrics.Gap, layout.Middle, children...)
}

func commandVisibleCount(widths []int, available, gap, moreWidth int) int {
	total := max(len(widths)-1, 0) * gap
	for _, width := range widths {
		total += width
	}
	if total <= available {
		return len(widths)
	}
	available -= moreWidth + gap
	used, visible := 0, 0
	for visible < len(widths) {
		candidate := widths[visible]
		if visible > 0 {
			candidate += gap
		}
		if used+candidate > available {
			break
		}
		used += candidate
		visible++
	}
	return visible
}

func (w Widget) updateKeys(ctx *core.Context, gtx layout.Context, items []core.FocusGroupItem) {
	if w.disabled {
		gtx = gtx.Disabled()
	}
	axis := base.CollectionHorizontal
	if w.orientation == Vertical {
		axis = base.CollectionVertical
	}
	var navigation base.CollectionNavigation
	navigation.UpdateWithEvents(gtx, base.Collection{
		Count: len(items),
		Targets: func(yield func(int, event.Tag) bool) {
			for index, item := range items {
				if !yield(index, item.Tag) {
					return
				}
			}
		},
	}, -1, base.CollectionPolicy{Axis: axis, Wrap: w.loopFocus}, func(e base.CollectionEvent) int {
		core.RequestFocusVisible(ctx, items[e.Index].Tag, true)
		return e.Index
	})
}

type SeparatorWidget struct {
	orientation Orientation
	key         string
	style       style.Style
	visual      *toolbarVisual
}

func Separator() SeparatorWidget {
	return SeparatorWidget{key: core.AutoKey("toolbar-separator")}
}

func (s SeparatorWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := s.visual
	if visual == nil {
		resolvedVisual := buildToolbar(core.ActiveTheme(ctx), false)
		visual = &resolvedVisual
	}
	declaration := visual.Separator
	if s.orientation == Vertical {
		declaration = visual.VerticalSeparator
	}
	resolved := computed.ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, s.style)
	if len(resolved.Transitions) != 0 {
		key := core.ClaimDerivedKey(ctx, core.KindStyle, s.key, "style")
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, nil)
}

func (w Widget) resolveStyle(ctx *core.Context, gtx layout.Context, visual toolbarVisual) style.ResolvedStyle {
	resolved := computed.ResolveStatic(
		ctx,
		style.State{Disabled: w.disabled},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		w.style,
	)
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	key := core.ClaimDerivedKey(ctx, core.KindStyle, w.key, "style")
	return computed.ApplyTransitions(ctx, gtx, key, resolved)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}

// Style merges instance declarations into the separator.
func (s SeparatorWidget) Style(values ...style.Style) SeparatorWidget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}
