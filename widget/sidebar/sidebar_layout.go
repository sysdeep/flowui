package sidebar

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/internal/textrender"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/tooltip"
)

// sidebarResolved is one layout pass's theme defaults plus the resolved root and
// static label layers. It carries no renderer indirection.
type sidebarResolved struct {
	defaults     sidebarDefaults
	root         style.ResolvedStyle
	emptyLabel   style.ResolvedStyle
	sectionLabel style.ResolvedStyle
	foreground   color.NRGBA
	background   color.NRGBA
	styleKey     string
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, state *sidebarState, entries []entry) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	defaults := state.defaults.resolve(active, w.metrics(active))
	styleKey := core.FullKey(ctx, w.key)
	root := computed.Resolve(
		ctx,
		gtx,
		styleKey,
		style.State{Disabled: w.disabled},
		defaults.style,
		style.Style{},
		style.Style{},
		w.style,
	)
	sem := active.Semantic()
	resolved := sidebarResolved{
		defaults:   defaults,
		root:       root,
		foreground: textrender.ColorOr(root, sem.SurfaceForeground),
		background: sidebarBackground(root, sem.Surface),
		styleKey:   styleKey,
	}
	muted := style.New().TextColor(style.Solid(sem.MutedForeground))
	resolved.emptyLabel = computed.ResolvePartStatic(
		ctx,
		style.PartLabel,
		style.State{},
		defaults.style,
		style.Join(muted, style.New().FontSize(defaults.metrics.ItemTextSize)),
		style.Style{},
		w.style,
	)
	resolved.sectionLabel = computed.ResolvePartStatic(
		ctx,
		style.PartLabel,
		style.State{},
		defaults.style,
		style.Join(muted, style.New().FontSize(defaults.metrics.SectionTextSize)),
		style.Style{},
		w.style,
	)

	metrics := defaults.metrics
	padding := metrics.Padding
	expandedWidth := metrics.Width
	collapsedWidth := metrics.CollapsedWidth
	target := float32(1)
	easing := animation.EasingFor(defaults.expansionEasing)
	if w.collapsed {
		target = 0
		easing = animation.EasingFor(defaults.collapseEasing)
	}
	expansion := animation.Tween("expansion", target).
		Easing(easing).
		Value(ctx, gtx)
	expandedWidthPx := min(gtx.Dp(expandedWidth), gtx.Constraints.Max.X)
	width := int(animation.LerpFloat(float32(gtx.Dp(collapsedWidth)), float32(expandedWidthPx), expansion) + .5)
	size := gtx.Constraints.Constrain(image.Pt(width, gtx.Constraints.Max.Y))
	expandedContentWidth := max(expandedWidthPx-gtx.Dp(padding)*2, 0)
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	if w.alt != "" {
		semantic.DescriptionOp(w.alt).Add(rootGtx.Ops)
	}

	macro := op.Record(rootGtx.Ops)
	var content layout.Dimensions
	func() {
		restore := core.PushColors(ctx, resolved.foreground, resolved.background)
		defer restore()
		content = layoutui.TrackedInset(ctx, rootGtx, layout.UniformInset(padding), func(gtx layout.Context) layout.Dimensions {
			children := make([]core.Widget, 0, 3)
			if w.header != nil {
				children = append(children, sidebarTransitionContent(w.header, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			children = append(children, layoutui.Expanded(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return w.layoutEntries(ctx, gtx, state, entries, expansion, resolved)
			})))
			if w.footer != nil {
				children = append(children, sidebarTransitionContent(w.footer, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			return layoutui.TrackedFlex(ctx, gtx, layout.Vertical, metrics.ContentGap, layout.Start, children...)
		})
	}()
	call := macro.Stop()

	drawSidebarRoot(rootGtx, size, active, root)
	area := clip.Rect{Max: size}.Push(rootGtx.Ops)
	call.Add(rootGtx.Ops)
	area.Pop()
	return layout.Dimensions{Size: size, Baseline: content.Baseline}
}

func (w Widget) layoutEntries(ctx *core.Context, gtx layout.Context, state *sidebarState, entries []entry, expansion float32, resolved sidebarResolved) layout.Dimensions {
	metrics := resolved.defaults.metrics
	if len(entries) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return textrender.Label(core.ActiveTheme(ctx), gtx, w.resolvedEmptyText(ctx), resolved.emptyLabel, metrics.ItemTextSize)
		})
	}
	state.list.Axis = layout.Vertical
	state.list.Gap = gtx.Dp(metrics.ItemGap)
	state.list.Alignment = layout.Start
	state.list.ScrollToEnd = false
	state.list.ScrollAnyAxis = false
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(entries), w.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
		entry := &entries[index]
		if entry.section {
			return w.layoutSection(ctx, gtx, entry.title, expansion, resolved)
		}
		return w.layoutItem(ctx, gtx, state, state.item(entry.item.Key), entry, expansion, resolved)
	})
}

func (w Widget) layoutSection(ctx *core.Context, gtx layout.Context, title string, expansion float32, resolved sidebarResolved) layout.Dimensions {
	metrics := resolved.defaults.metrics
	height := min(gtx.Dp(metrics.SectionHeight), gtx.Constraints.Max.Y)
	sectionGtx := gtx
	sectionGtx.Constraints.Min.Y = height
	sectionGtx.Constraints.Max.Y = height
	fade := paint.PushOpacity(gtx.Ops, w.targetContentOpacity(expansion))
	defer fade.Pop()
	if w.collapsed {
		return layout.Center.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
			width := max(gtx.Constraints.Max.X-gtx.Dp(metrics.SectionSeparatorInset)*2, 0)
			lineHeight := max(gtx.Dp(metrics.BorderWidth), 1)
			rect := image.Rect(0, 0, width, lineHeight)
			paint.FillShape(gtx.Ops, core.ActiveTheme(ctx).Semantic().Border, clip.Rect(rect).Op())
			return layout.Dimensions{Size: rect.Size()}
		})
	}
	return layout.W.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
		return textrender.Label(core.ActiveTheme(ctx), gtx, title, resolved.sectionLabel, metrics.SectionTextSize)
	})
}

func (w Widget) targetContentOpacity(expansion float32) float32 {
	if w.collapsed {
		return 1 - expansion
	}
	return expansion
}

func (w Widget) expandedLayoutWidth(width int) int {
	if w.collapsed {
		return 0
	}
	return width
}

func sidebarTransitionContent(child core.Widget, opacity float32, width int) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if width > 0 {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
		}
		content := recordSidebarContent(ctx, gtx, child)
		drawSidebarContent(gtx, content, image.Point{}, opacity)
		return content.dims
	})
}

// resolveItem resolves one item's surface, label, and default switcher glyph.
// The item part supplies the text color, which is inherited by the label and
// glyph parts so an item-level declaration reaches all of them. The instance
// style is the last layer in both cascades.
func (w Widget) resolveItem(
	ctx *core.Context,
	gtx layout.Context,
	resolved sidebarResolved,
	key string,
	state style.State,
) (item, label, icon style.ResolvedStyle) {
	itemKey := core.DerivedKey(ctx, resolved.styleKey, "item:"+key)
	item = computed.ResolvePart(
		ctx,
		gtx,
		itemKey,
		style.PartItem,
		state,
		resolved.defaults.style,
		style.Style{},
		style.Style{},
		w.style,
	)
	var parts [2]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		itemKey,
		state,
		sidebarContentParts[:],
		parts[:],
		resolved.defaults.style,
		style.TextDeclaration(item.Text),
		style.Style{},
		w.style,
	)
	return item, parts[0], parts[1]
}

func (w Widget) layoutItem(ctx *core.Context, gtx layout.Context, sidebarState *sidebarState, itemState *sidebarItemState, entry *entry, expansion float32, resolved sidebarResolved) layout.Dimensions {
	item := entry.item
	metrics := resolved.defaults.metrics
	disabled := w.itemDisabled(item)
	interaction := base.RowInteraction(ctx, gtx, &itemState.clickable, disabled, func() {
		if len(item.Children) > 0 && !w.collapsed {
			w.requestOpen(item.Key, !entry.expanded)
		} else {
			w.activate(item.Key)
		}
	})
	disabled = interaction.Disabled
	itemHeight := metrics.ItemHeight
	height := min(max(gtx.Dp(itemHeight), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	size := image.Pt(gtx.Constraints.Max.X, height)
	itemGtx := gtx
	itemGtx.Constraints = layout.Exact(size)
	selected := item.Key == w.selectedKey
	focusVisible := interaction.FocusVisible
	itemVisual, labelVisual, iconVisual := w.resolveItem(ctx, gtx, resolved, item.Key, style.State{
		Hovered:      interaction.Hovered,
		Pressed:      interaction.Pressed,
		Focused:      interaction.Focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		Selected:     selected,
	})
	if w.collapsed || w.expandAction != ExpandActionHover || len(item.Children) == 0 || !interaction.Hovered || entry.expanded {
		itemState.hoverOpenRequested = false
	} else if !itemState.hoverOpenRequested {
		itemState.hoverOpenRequested = true
		w.requestOpen(item.Key, true)
	}
	focus := itemState.focus.Opacity(gtx, focusVisible && !disabled, resolved.defaults.motion)
	if w.collapsed && len(item.Children) > 0 {
		return w.layoutCollapsedSubmenu(ctx, itemGtx, entry, itemVisual, labelVisual, iconVisual, selected, disabled, focus, resolved)
	}

	if !w.collapsed || expansion > 0 || item.Label == "" {
		return w.layoutItemTrigger(ctx, itemGtx, size, itemState, entry, itemVisual, labelVisual, iconVisual, selected, disabled, focus, expansion, resolved)
	}
	trigger := core.WidgetFunc(func(ctx *core.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemTrigger(ctx, triggerGtx, size, itemState, entry, itemVisual, labelVisual, iconVisual, selected, disabled, focus, expansion, resolved)
	})
	return tooltip.New(trigger, item.Label).Key("sidebar-item-label:"+item.Key).
		Placement(core.PopoverRight).
		Delay(0).
		Layout(ctx, itemGtx)
}

func (w Widget) layoutItemTrigger(ctx *core.Context, gtx layout.Context, size image.Point, itemState *sidebarItemState, entry *entry, itemVisual, labelVisual, iconVisual style.ResolvedStyle, selected, disabled bool, focus, expansion float32, resolved sidebarResolved) layout.Dimensions {
	gtx.Constraints = layout.Exact(size)
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, gtx, entry, itemVisual, labelVisual, iconVisual, selected, !disabled, focus, expansion, resolved)
	})
}

func (w Widget) layoutItemSurface(ctx *core.Context, gtx layout.Context, entry *entry, itemVisual, labelVisual, iconVisual style.ResolvedStyle, selected, enabled bool, focus, expansion float32, resolved sidebarResolved) layout.Dimensions {
	item := entry.item
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(item.Label).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if len(item.Children) > 0 {
		semantic.DescriptionOp(sidebarGroupDescription(ctx, entry.expanded)).Add(gtx.Ops)
	}

	size := gtx.Constraints.Min
	foreground := textrender.ColorOr(labelVisual, resolved.foreground)
	background := sidebarBackground(itemVisual, color.NRGBA{})
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement core.OverlayPlacement
	func() {
		restore := core.PushColors(ctx, foreground, background)
		defer restore()
		contentGtx := gtx
		contentGtx.Constraints.Min = image.Pt(size.X, 0)
		contentGtx.Constraints.Max = size
		contentDims, contentPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return w.layoutItemContent(ctx, contentGtx, entry, foreground, labelVisual, iconVisual, expansion, resolved)
		})
	}()
	content := macro.Stop()
	contentOffset := image.Pt(0, max((size.Y-contentDims.Size.Y)/2, 0))
	contentPlacement.PlaceOffset(contentOffset)

	opacity := paint.PushOpacity(gtx.Ops, sidebarOpacity(itemVisual, 1))
	drawSidebarItem(gtx, size, itemVisual, focus, resolved.defaults.metrics.ItemRadius)
	offset := op.Offset(contentOffset).Push(gtx.Ops)
	content.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	return layout.Dimensions{Size: size}
}

func sidebarGroupDescription(ctx *core.Context, expanded bool) string {
	if expanded {
		return ctx.Text(i18n.SidebarGroupCollapse)
	}
	return ctx.Text(i18n.SidebarGroupExpand)
}

func (w Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, entry *entry, foreground color.NRGBA, labelVisual, iconVisual style.ResolvedStyle, expansion float32, resolved sidebarResolved) layout.Dimensions {
	item := entry.item
	metrics := resolved.defaults.metrics
	width := gtx.Constraints.Max.X
	itemPadding := metrics.ItemPaddingX
	padding := gtx.Dp(itemPadding)
	gap := gtx.Dp(metrics.ItemContentGap)
	outerPadding := metrics.Padding
	collapsedWidth := metrics.CollapsedWidth
	collapsedContentWidth := min(max(gtx.Dp(collapsedWidth)-gtx.Dp(outerPadding)*2, 0), width)
	indent := 0
	if !w.collapsed {
		indent = gtx.Dp(metrics.InlineIndent) * entry.depth
	}
	contentWidth := max(width-indent, 0)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childGtx.Constraints.Max.X = max(contentWidth-padding*2, 0)

	var switcher sidebarContent
	if !w.collapsed && len(item.Children) > 0 {
		switcher = recordSidebarContent(ctx, childGtx, sidebarItemSwitcher(item, entry.expanded, foreground, iconVisual, metrics.SwitcherSize))
	}

	var leading, initial sidebarContent
	if item.Leading != nil {
		leading = recordSidebarContent(ctx, childGtx, item.Leading)
	} else if expansion < 1 {
		initial = recordSidebarLabel(ctx, childGtx, item.Label, labelVisual, metrics.ItemTextSize, true)
	}

	var trailing sidebarContent
	if expansion > 0 && item.Trailing != nil {
		trailing = recordSidebarContent(ctx, childGtx, item.Trailing)
	}
	switcherX := width - padding - switcher.dims.Size.X
	trailingX := switcherX - trailing.dims.Size.X
	if trailing.dims.Size.X > 0 && switcher.dims.Size.X > 0 {
		trailingX -= gap
	}

	leadingWidth := leading.dims.Size.X
	labelX := indent + padding
	leadingX := 0
	if leadingWidth > 0 {
		collapsedX := (collapsedContentWidth - leadingWidth) / 2
		leadingX = int(animation.LerpFloat(float32(collapsedX), float32(labelX), expansion) + .5)
		labelX = leadingX + leadingWidth + gap
	}

	var label sidebarContent
	if expansion > 0 {
		labelGtx := childGtx
		labelRight := switcherX
		if trailing.dims.Size.X > 0 {
			labelRight = trailingX - gap
		} else if switcher.dims.Size.X > 0 {
			labelRight -= gap
		}
		labelGtx.Constraints.Max.X = max(labelRight-labelX, 0)
		label = recordSidebarLabel(ctx, labelGtx, item.Label, labelVisual, metrics.ItemTextSize, false)
	}

	height := max(leading.dims.Size.Y, initial.dims.Size.Y, label.dims.Size.Y, trailing.dims.Size.Y, switcher.dims.Size.Y)
	centerY := func(content sidebarContent) int { return max((height-content.dims.Size.Y)/2, 0) }
	if leadingWidth > 0 {
		drawSidebarContent(gtx, leading, image.Pt(leadingX, centerY(leading)), 1)
	} else {
		initialX := max((collapsedContentWidth-initial.dims.Size.X)/2, 0)
		drawSidebarContent(gtx, initial, image.Pt(initialX, centerY(initial)), 1-expansion)
	}
	drawSidebarContent(gtx, label, image.Pt(labelX, centerY(label)), expansion)
	drawSidebarContent(gtx, trailing, image.Pt(trailingX, centerY(trailing)), expansion)
	if switcher.dims.Size.X > 0 {
		drawSidebarContent(gtx, switcher, image.Pt(switcherX, centerY(switcher)), 1)
	}
	return layout.Dimensions{Size: image.Pt(width, height)}
}

func (w Widget) layoutCollapsedSubmenu(ctx *core.Context, gtx layout.Context, entry *entry, itemVisual, labelVisual, iconVisual style.ResolvedStyle, selected, disabled bool, focus float32, resolved sidebarResolved) layout.Dimensions {
	trigger := core.WidgetFunc(func(ctx *core.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, triggerGtx, entry, itemVisual, labelVisual, iconVisual, selected, !disabled, focus, 0, resolved)
	})
	menuKey := "sidebar-submenu:" + entry.item.Key + ":menu"
	menuWidget := menu.New(w.sidebarDropdownItems(entry.item.Children)).Key(menuKey).
		AutoWidth().
		SelectionMode(menu.SelectionSingle).
		SelectedKey(w.selectedKey).
		OnActionEvent(func(event menu.ActionEvent) { w.activate(event.Key) })
	if w.hasDataVersion {
		menuWidget = menuWidget.DataVersion(w.dataVersion)
	}
	popup := dropdown.New(trigger, nil).Key("sidebar-submenu:" + entry.item.Key).
		Menu(menuWidget).
		Placement(core.PopoverRightStart).
		// A collapsed navigation item opens its flyout on hover. ExpandAction
		// only controls inline groups while the sidebar is expanded.
		TriggerMode(dropdown.TriggerHover).
		Disabled(disabled || w.disabled)
	return popup.Layout(ctx, gtx)
}

func (w Widget) sidebarDropdownItems(items []Item) []dropdown.Item {
	result := make([]dropdown.Item, 0, len(items))
	for _, item := range items {
		result = append(result, dropdown.Item{
			Key:      item.Key,
			Label:    item.Label,
			Leading:  item.Leading,
			Trailing: item.Trailing,
			Disabled: w.itemDisabled(item),
			Children: w.sidebarDropdownItems(item.Children),
		})
	}
	return result
}

func sidebarItemSwitcher(item Item, expanded bool, foreground color.NRGBA, iconVisual style.ResolvedStyle, fallbackSize unit.Dp) core.Widget {
	value := item.Switcher
	if expanded && item.ExpandedSwitcher != nil {
		value = item.ExpandedSwitcher
	}
	if value != nil {
		return value
	}
	color := textrender.ColorOr(iconVisual, foreground)
	size := sidebarIconSize(iconVisual, fallbackSize)
	return core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutSidebarSwitcher(gtx, expanded, color, size)
	})
}

type sidebarContent struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement core.OverlayPlacement
}

func recordSidebarContent(ctx *core.Context, gtx layout.Context, child core.Widget) sidebarContent {
	if child == nil {
		return sidebarContent{}
	}
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackWidgetPlacement(ctx, gtx, child)
	return sidebarContent{call: macro.Stop(), dims: dims, placement: placement}
}

func drawSidebarContent(gtx layout.Context, content sidebarContent, position image.Point, opacity float32) {
	content.placement.PlaceOffset(position)
	content.placement.SetOpacity(opacity)
	if opacity <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	content.call.Add(gtx.Ops)
	fade.Pop()
	offset.Pop()
}

// recordSidebarLabel records plain text directly. Labels cannot register
// overlays, so a placement transform would only add per-item work.
func recordSidebarLabel(ctx *core.Context, gtx layout.Context, label string, part style.ResolvedStyle, size unit.Sp, initial bool) sidebarContent {
	value := label
	if initial {
		for _, runeValue := range value {
			value = string(runeValue)
			break
		}
	}
	macro := op.Record(gtx.Ops)
	dims := textrender.Label(core.ActiveTheme(ctx), gtx, value, part, size)
	return sidebarContent{call: macro.Stop(), dims: dims}
}
