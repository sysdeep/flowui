package table

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/toggle"
)

type tableColumns struct {
	selection int
	widths    []int
	width     int
}

// tableStyles is one layout pass's resolved appearance. The instance style is
// already part of every surface layer; the utility accents come from the
// cached defaults alone.
type tableStyles struct {
	root      style.ResolvedStyle
	header    style.ResolvedStyle
	body      style.ResolvedStyle
	muted     style.ResolvedStyle
	separator style.ResolvedStyle
	focus     style.ResolvedStyle
	resize    style.ResolvedStyle
	stripe    style.ResolvedStyle
}

func (t Widget) resolveStyles(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	defaults tableDefaults,
) tableStyles {
	empty := style.Style{}
	return tableStyles{
		root:      computed.Resolve(ctx, gtx, key, state, defaults.Root, empty, empty, t.style),
		header:    computed.ResolveStatic(ctx, style.State{}, defaults.Header, empty, empty, t.style),
		body:      computed.ResolveStatic(ctx, style.State{}, defaults.Body, empty, empty, t.style),
		muted:     computed.ResolveStatic(ctx, style.State{}, defaults.Muted, empty, empty, t.style),
		separator: computed.ResolveStatic(ctx, style.State{}, defaults.Separator, empty, empty, empty),
		focus:     computed.ResolveStatic(ctx, style.State{}, defaults.Focus, empty, empty, empty),
		resize:    computed.ResolveStatic(ctx, style.State{}, defaults.Resize, empty, empty, empty),
		stripe:    computed.ResolveStatic(ctx, style.State{}, defaults.Stripe, empty, empty, empty),
	}
}

func tableBackground(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Paint != nil {
		if brush, ok := computed.Brush(resolved.Paint.Background); ok {
			if value := brush.ColorAt(.5); value.A != 0 {
				return value
			}
		}
	}
	return fallback
}

func tableTextColor(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Text != nil {
		if value, ok := computed.Color(resolved.Text.Color); ok {
			return value
		}
	}
	return fallback
}

func tableOpacity(resolved style.ResolvedStyle) float32 {
	if resolved.Paint != nil && resolved.Paint.Opacity != nil {
		return *resolved.Paint.Opacity
	}
	return 1
}

func tableRadius(resolved style.ResolvedStyle, fallback unit.Dp) unit.Dp {
	if resolved.Paint != nil && resolved.Paint.Radius != nil {
		return *resolved.Paint.Radius
	}
	return fallback
}

func tableBorder(resolved style.ResolvedStyle, fallback color.NRGBA) (unit.Dp, color.NRGBA, bool) {
	paintStyle := resolved.Paint
	if paintStyle == nil || paintStyle.Border == nil || paintStyle.Border.Width == nil {
		return 0, color.NRGBA{}, false
	}
	width := *paintStyle.Border.Width
	if width <= 0 {
		return 0, color.NRGBA{}, false
	}
	value := fallback
	if col, ok := computed.Color(paintStyle.Border.Color); ok {
		value = col
	}
	if value.A == 0 {
		return 0, color.NRGBA{}, false
	}
	return width, value, true
}

func tableFillRoot(gtx layout.Context, size image.Point, radius int, resolved style.ResolvedStyle) {
	if size.X <= 0 || size.Y <= 0 || resolved.Paint == nil {
		return
	}
	brush, ok := computed.Brush(resolved.Paint.Background)
	if !ok || brush.ColorAt(.5).A == 0 {
		return
	}
	draw.FillRRect(gtx, clip.UniformRRect(image.Rectangle{Max: size}, radius), brush)
}

func (t Widget) layout(ctx *core.Context, gtx layout.Context, stateValue *tableState, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	geometry := defaults.Geometry

	padding := 0
	if t.variant == VariantPrimary {
		padding = gtx.Dp(geometry.RootPadding)
	}
	viewportWidth := max(gtx.Constraints.Max.X-padding*2, 0)
	columns := t.resolveColumns(ctx, gtx, stateValue, viewportWidth, geometry)
	contentMaxHeight := gtx.Constraints.Max.Y
	if !t.flow {
		maxHeight := geometry.MaxHeight
		if t.maxHeight > 0 {
			maxHeight = unit.Dp(t.maxHeight)
		}
		if maxHeight > 0 {
			contentMaxHeight = min(contentMaxHeight, gtx.Dp(maxHeight))
		}
	}

	macro := op.Record(gtx.Ops)
	innerGtx := gtx
	innerGtx.Constraints.Min.X = viewportWidth
	innerGtx.Constraints.Max.X = viewportWidth
	innerGtx.Constraints.Min.Y = 0
	innerGtx.Constraints.Max.Y = max(contentMaxHeight, 0)
	contentDims, contentPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutHorizontal(ctx, innerGtx, stateValue, columns, defaults, styles)
	})
	call := macro.Stop()

	footerHeight := 0
	var footerCall op.CallOp
	if t.footer != nil {
		footerMacro := op.Record(gtx.Ops)
		footerGtx := gtx
		footerGtx.Constraints.Min.X = viewportWidth
		footerGtx.Constraints.Max.X = viewportWidth
		footerGtx.Constraints.Min.Y = 0
		footerGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-contentDims.Size.Y-padding, 0)
		footerForeground := tableTextColor(styles.root, ctx.ForegroundColor())
		footerBackground := tableBackground(styles.root, ctx.BackgroundColor())
		restore := core.PushColors(ctx, footerForeground, footerBackground)
		footerDims := layout.Inset{Top: geometry.FooterPaddingY, Right: geometry.FooterPaddingX, Bottom: geometry.FooterPaddingY, Left: geometry.FooterPaddingX}.Layout(footerGtx, func(gtx layout.Context) layout.Dimensions {
			return t.footer.Layout(ctx, gtx)
		})
		restore()
		footerHeight = footerDims.Size.Y
		footerCall = footerMacro.Stop()
	}

	size := image.Pt(gtx.Constraints.Max.X, contentDims.Size.Y+footerHeight+padding)
	size = gtx.Constraints.Constrain(size)
	rootRadius := tableClampRadius(gtx, size, tableRadius(styles.root, geometry.RootRadius))
	tableFillRoot(gtx, size, rootRadius, styles.root)
	root := clip.UniformRRect(image.Rectangle{Max: size}, rootRadius).Push(gtx.Ops)
	contentPlacement.PlaceOffset(image.Pt(padding, 0))
	contentPlacement.ClipTo(image.Rectangle{Max: image.Pt(viewportWidth, contentDims.Size.Y)})
	offset := op.Offset(image.Pt(padding, 0)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	if t.footer != nil {
		footerOffset := op.Offset(image.Pt(0, contentDims.Size.Y)).Push(gtx.Ops)
		footerCall.Add(gtx.Ops)
		footerOffset.Pop()
	}
	offset.Pop()
	root.Pop()
	separator := tableBackground(styles.separator, color.NRGBA{})
	if width, col, ok := tableBorder(styles.root, separator); ok {
		tableDrawBorder(gtx, size, rootRadius, max(gtx.Dp(width), 1), col)
	}
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutHorizontal(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	if t.flow {
		// A flow table owns no viewport, so it cannot scroll sideways either;
		// content wider than the container simply overflows.
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, defaults, styles)
	}
	stateValue.horizontal.Axis = layout.Horizontal
	stateValue.horizontal.Gap = 0
	stateValue.horizontal.Alignment = layout.Start
	stateValue.horizontal.ScrollAnyAxis = false
	return layoutui.TrackedScrollbar(ctx, gtx, &stateValue.horizontal, &stateValue.horizontalBar, 1, t.disabled, false, func(gtx layout.Context, _ int) layout.Dimensions {
		gtx.Constraints.Min.X = columns.width
		gtx.Constraints.Max.X = columns.width
		return t.layoutContent(ctx, gtx, stateValue, columns, defaults, styles)
	})
}

func (t Widget) layoutContent(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	geometry := defaults.Geometry
	headerHeightToken := geometry.HeaderHeight
	if t.headerHeight > 0 {
		headerHeightToken = unit.Dp(t.headerHeight)
	}
	headerHeight := min(gtx.Dp(headerHeightToken), gtx.Constraints.Max.Y)
	headerGtx := gtx
	headerGtx.Constraints = layout.Exact(image.Pt(columns.width, headerHeight))
	headerDims := t.layoutHeader(ctx, headerGtx, stateValue, columns, defaults, styles)

	bodyGtx := gtx
	bodyGtx.Constraints.Min.Y = 0
	bodyGtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-headerDims.Size.Y, 0)
	bodyOffset := op.Offset(image.Pt(0, headerDims.Size.Y)).Push(gtx.Ops)
	bodyDims, bodyPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutBody(ctx, bodyGtx, stateValue, columns, defaults, styles)
	})
	bodyOffset.Pop()
	bodyPlacement.PlaceOffset(image.Pt(0, headerDims.Size.Y))
	size := image.Pt(columns.width, headerDims.Size.Y+bodyDims.Size.Y)
	t.drawColumnResizeGuide(ctx, gtx, stateValue, columns, size, defaults, styles)
	return layout.Dimensions{Size: size}
}

func (t Widget) drawColumnResizeGuide(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, size image.Point, defaults tableDefaults, styles tableStyles) {
	width := max(gtx.Dp(defaults.Geometry.ColumnResizerWidth), 1)
	accent := tableBackground(styles.resize, color.NRGBA{})
	x := columns.selection
	for index, column := range t.columns {
		x += columns.widths[index]
		if column.Resizable && stateValue.column(column.Key).resize.dragging {
			tableDrawResizer(gtx, x, size.Y, size.Y, width, color.NRGBA{}, accent, true, 0)
		}
	}
}

func (t Widget) layoutHeader(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	geometry := defaults.Geometry
	size := gtx.Constraints.Max
	radius := tableClampRadius(gtx, size, geometry.HeaderRadius)
	headerBackground := tableBackground(styles.header, ctx.BackgroundColor())
	headerForeground := tableTextColor(styles.header, ctx.ForegroundColor())
	columnSeparator := color.NRGBA{}
	headerSeparator := color.NRGBA{}
	if t.showsGridLines() {
		columnSeparator = tableBackground(styles.separator, ctx.BackgroundColor())
		headerSeparator = columnSeparator
	}
	tableDrawHeader(gtx, size, radius, max(gtx.Dp(geometry.SeparatorWidth), 1), headerBackground, headerSeparator)
	restore := core.PushColors(ctx, headerForeground, headerBackground)
	defer restore()
	separatorWidth := max(gtx.Dp(geometry.SeparatorWidth), 1)
	separatorHeight := max(gtx.Dp(geometry.ColumnSeparatorHeight), 1)
	x := 0
	if columns.selection > 0 {
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(columns.selection, size.Y))
		if t.selectionMode == SelectionMultiple {
			all, some := t.selectionSummary()
			motion := defaults.Motion
			selection := stateValue.selectAllSelection.Progress(gtx, all || some, motion.ControlDuration, motion)
			stateValue.selectAll.Layout(cellGtx, func(gtx layout.Context) layout.Dimensions {
				semantic.CheckBox.Add(gtx.Ops)
				semantic.LabelOp(ctx.Text(i18n.TableSelectAllRows)).Add(gtx.Ops)
				semantic.SelectedOp(all).Add(gtx.Ops)
				focusVisible := core.FocusVisible(ctx, &stateValue.selectAll, gtx.Focused(&stateValue.selectAll))
				focus := stateValue.selectAllFocus.Opacity(gtx, focusVisible && !t.disabled, motion)
				toggle.DrawCheckboxControl(ctx, gtx, toggle.CheckboxControlOptions{
					Variant: toggle.CheckboxPrimary, Selection: selection,
					Indeterminate: some, Hovered: stateValue.selectAll.Hovered(),
					Pressed: stateValue.selectAll.Pressed(), Focused: focus, Disabled: t.disabled,
					VisualCache: &stateValue.checkboxVisual,
				})
				return layout.Dimensions{Size: cellGtx.Constraints.Max}
			})
		}
		x += columns.selection
		if len(t.columns) > 0 && t.showsGridLines() {
			tableDrawHeaderSeparator(gtx, x, size.Y, separatorWidth, separatorHeight, columnSeparator, t.showsFullGrid())
		}
	}
	activeForeground := tableTextColor(styles.body, headerForeground)
	for index, column := range t.columns {
		width := columns.widths[index]
		cellGtx := gtx
		cellGtx.Constraints = layout.Exact(image.Pt(width, size.Y))
		offset := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
		t.layoutHeaderCell(ctx, cellGtx, stateValue, column, defaults, styles, headerForeground, activeForeground, columnSeparator)
		offset.Pop()
		x += width
		if index < len(t.columns)-1 && !column.Resizable && t.showsGridLines() {
			tableDrawHeaderSeparator(gtx, x, size.Y, separatorWidth, separatorHeight, columnSeparator, t.showsFullGrid())
		}
	}
	t.layoutColumnResizers(ctx, gtx, stateValue, columns, defaults, styles, size, columnSeparator)
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutColumnResizers(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles, size image.Point, columnSeparator color.NRGBA) {
	geometry := defaults.Geometry
	hitSize := max(gtx.Dp(geometry.ColumnResizerHitSize), 1)
	activeWidth := max(gtx.Dp(geometry.ColumnResizerWidth), 1)
	baseHeight := max(gtx.Dp(geometry.ColumnSeparatorHeight), 1)
	if t.showsFullGrid() {
		baseHeight = size.Y
	}
	enabled := gtx.Enabled() && !t.disabled
	accent := tableBackground(styles.resize, color.NRGBA{})
	x := columns.selection
	for index, column := range t.columns {
		x += columns.widths[index]
		if !column.Resizable {
			continue
		}
		resize := &stateValue.column(column.Key).resize
		focusVisible := core.FocusVisible(ctx, resize, gtx.Focused(resize))
		focus := resize.focus.Opacity(gtx, focusVisible && enabled, defaults.Motion)
		tableDrawResizer(gtx, x, size.Y, baseHeight, activeWidth, columnSeparator, accent, false, focus)

		hit := image.Rect(max(x-hitSize/2, 0), 0, min(x+(hitSize+1)/2, size.X), size.Y)
		if hit.Empty() {
			continue
		}
		clipped := clip.Rect(hit).Push(gtx.Ops)
		semantic.LabelOp(ctx.Textf(i18n.TableResizeColumn, column.Label)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if enabled {
			pointer.CursorColResize.Add(gtx.Ops)
			event.Op(gtx.Ops, resize)
		}
		clipped.Pop()
	}
}

func (t Widget) layoutHeaderCell(ctx *core.Context, gtx layout.Context, stateValue *tableState, column Column, defaults tableDefaults, styles tableStyles, foreground, activeForeground, separator color.NRGBA) layout.Dimensions {
	geometry := defaults.Geometry
	content := func(gtx layout.Context, child layout.Widget) layout.Dimensions {
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = 0
			return layout.Inset{Left: geometry.CellPaddingX, Right: geometry.CellPaddingX}.Layout(gtx, child)
		})
	}
	if !column.Sortable {
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutHeaderContent(ctx, gtx, column, defaults, foreground)
		})
	}
	columnState := stateValue.column(column.Key)
	return columnState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(column.Label).Add(gtx.Ops)
		semantic.DescriptionOp(ctx.Text(i18n.TableSortColumn)).Add(gtx.Ops)
		labelColor := foreground
		if columnState.clickable.Hovered() || t.sort.Column == column.Key {
			labelColor = activeForeground
		}
		focused := core.FocusVisible(ctx, &columnState.clickable, gtx.Focused(&columnState.clickable))
		focus := columnState.focus.Opacity(gtx, focused && !t.disabled, defaults.Motion)
		focusColor := tableBackground(styles.focus, foreground)
		tableDrawCellFocus(gtx, gtx.Constraints.Max, focus, max(gtx.Dp(geometry.FocusRingWidth), 1), geometry.FocusRadius, focusColor)
		return content(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutSortableHeaderContent(ctx, gtx, column, defaults, labelColor)
		})
	})
}

func (t Widget) layoutHeaderContent(ctx *core.Context, gtx layout.Context, column Column, defaults tableDefaults, foreground color.NRGBA) layout.Dimensions {
	child := column.Header
	if child == nil {
		child = tableCellText{
			text: column.Label, size: defaults.Geometry.HeaderTextSize,
			color: foreground, weight: defaults.Geometry.HeaderTextWeight,
		}
	}
	return alignWidget(ctx, gtx, column.Align, child)
}

func (t Widget) layoutSortableHeaderContent(ctx *core.Context, gtx layout.Context, column Column, defaults tableDefaults, foreground color.NRGBA) layout.Dimensions {
	geometry := defaults.Geometry
	label := column.Header
	if label == nil {
		label = tableCellText{text: column.Label, size: geometry.HeaderTextSize, color: foreground, weight: geometry.HeaderTextWeight}
	}
	children := []layout.FlexChild{
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, label)
		}),
	}
	if t.sort.Column == column.Key {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Dp(geometry.SortIconSize), gtx.Dp(geometry.SortIconSize))
			tableDrawSortGlyph(gtx, size, t.sort.Direction, foreground, defaults.SortIcon)
			return layout.Dimensions{Size: size}
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(geometry.SortGap)}.Layout(gtx, children...)
}

func (t Widget) layoutBody(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	bodyBackground := tableBackground(styles.body, ctx.BackgroundColor())
	bodyForeground := tableTextColor(styles.body, ctx.ForegroundColor())
	restore := core.PushColors(ctx, bodyForeground, bodyBackground)
	defer restore()
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	rowCount := t.count()
	showLoader := t.hasMore || t.loadingMore
	if rowCount == 0 && !showLoader {
		contentDims = t.layoutEmpty(ctx, gtx, columns.width, defaults, styles)
	} else {
		itemCount := rowCount
		if showLoader {
			itemCount++
		}
		if t.flow {
			// Flow mode stacks every row in the page scroll, so there is no
			// virtual window to advance and no scroll position to read.
			rows := make([]core.Widget, 0, itemCount)
			for index := range itemCount {
				rows = append(rows, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
					if index == rowCount {
						return t.layoutLoadMore(ctx, gtx, columns.width, defaults.Geometry)
					}
					row := t.row(index)
					return t.layoutRow(ctx, gtx, stateValue, columns, defaults, styles, bodyBackground, bodyForeground, row, index, index == rowCount-1 && !showLoader)
				}))
			}
			contentDims = layoutui.Column(rows...).Layout(ctx, gtx)
		} else {
			stateValue.vertical.Axis = layout.Vertical
			stateValue.vertical.Gap = 0
			stateValue.vertical.Alignment = layout.Start
			stateValue.vertical.ScrollAnyAxis = false
			contentDims = layoutui.TrackedScrollbar(ctx, gtx, &stateValue.vertical, &stateValue.verticalBar, itemCount, t.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
				if index == rowCount {
					return t.layoutLoadMore(ctx, gtx, columns.width, defaults.Geometry)
				}
				row := t.row(index)
				return t.layoutRow(ctx, gtx, stateValue, columns, defaults, styles, bodyBackground, bodyForeground, row, index, index == rowCount-1 && !showLoader)
			})
			stateValue.updateLoadMore(rowCount, t.hasMore, t.loadingMore, stateValue.vertical.Position.First+stateValue.vertical.Position.Count >= itemCount, t.onLoadMore)
		}
	}
	call := macro.Stop()
	contentDims.Size.X = columns.width
	radius := defaults.Geometry.BodyRadius
	tableDrawBody(gtx, contentDims.Size, tableClampRadius(gtx, contentDims.Size, radius), bodyBackground)
	bodyClip := clip.UniformRRect(image.Rectangle{Max: contentDims.Size}, gtx.Dp(radius)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	bodyClip.Pop()
	return contentDims
}

func (t Widget) layoutLoadMore(ctx *core.Context, gtx layout.Context, width int, geometry tableGeometry) layout.Dimensions {
	height := 1
	if t.loadingMore {
		height = max(gtx.Dp(geometry.LoadMoreHeight), 1)
	}
	size := image.Pt(width, min(height, gtx.Constraints.Max.Y))
	loaderGtx := gtx
	loaderGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(loaderGtx, func(gtx layout.Context) layout.Dimensions {
		if !t.loadingMore {
			return layout.Dimensions{}
		}
		if t.loadMoreContent != nil {
			return t.loadMoreContent.Layout(ctx, gtx)
		}
		return spinner.New().Size(spinner.Medium).Label(ctx.Text(i18n.TableLoadingMore)).Layout(ctx, gtx)
	})
}

func (t Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, width int, defaults tableDefaults, styles tableStyles) layout.Dimensions {
	geometry := defaults.Geometry
	height := min(gtx.Dp(geometry.EmptyHeight), gtx.Constraints.Max.Y)
	size := image.Pt(width, max(height, 0))
	emptyGtx := gtx
	emptyGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(emptyGtx, func(gtx layout.Context) layout.Dimensions {
		if t.emptyContent != nil {
			return t.emptyContent.Layout(ctx, gtx)
		}
		return tableCellText{
			text:   t.resolvedEmptyText(ctx),
			size:   geometry.CellTextSize,
			color:  tableTextColor(styles.muted, tableTextColor(styles.body, ctx.ForegroundColor())),
			weight: font.Normal,
		}.Layout(ctx, gtx)
	})
}

type recordedCell struct {
	call         op.CallOp
	dims         layout.Dimensions
	placement    core.OverlayPlacement
	interactive  bool
	focusTargets []event.Tag
}

type tableRowTrigger func(*core.Context, layout.Context) layout.Dimensions

func (trigger tableRowTrigger) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return trigger(ctx, gtx)
}

func (t Widget) layoutRow(ctx *core.Context, gtx layout.Context, stateValue *tableState, columns tableColumns, defaults tableDefaults, styles tableStyles, bodyBackground, bodyForeground color.NRGBA, row Row, index int, last bool) layout.Dimensions {
	rowState := stateValue.rowAt(row.Key, index)
	disabled := t.rowDisabled(row)
	selected := t.isSelected(row.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min.X = columns.width
	gtx.Constraints.Max.X = columns.width
	trigger := tableRowTrigger(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return rowState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.LabelOp(t.rowLabel(row)).Add(gtx.Ops)
			semantic.SelectedOp(selected).Add(gtx.Ops)
			semantic.EnabledOp(!disabled).Add(gtx.Ops)
			stripe := style.Style{}
			if t.striped && index%2 == 1 {
				stripe = defaults.Stripe
			}
			hovered := rowState.clickable.Hovered() && !disabled
			pressed := rowState.clickable.Pressed() && !disabled
			focused := core.FocusVisible(ctx, &rowState.clickable, gtx.Focused(&rowState.clickable))
			rowStyle := computed.ResolveStatic(ctx, style.State{
				Hovered:      hovered,
				Pressed:      pressed,
				Focused:      focused,
				FocusVisible: focused && !disabled,
				Disabled:     disabled,
				Selected:     selected,
			}, defaults.Row, stripe, style.Style{}, t.style)
			rowBackground := tableBackground(rowStyle, bodyBackground)
			rowForeground := tableTextColor(rowStyle, bodyForeground)
			motion := defaults.Motion
			rowBackground = rowState.background.Value(
				animGtx,
				rowBackground,
				motion.ControlDuration,
				animation.EasingFor(motion.Ease),
				motion,
			)
			restore := core.PushColors(ctx, rowForeground, rowBackground)
			defer restore()

			recorded := make([]recordedCell, 0, len(row.Cells)+1)
			rowHeight := t.rowMinHeight(gtx, defaults.Geometry)
			if columns.selection > 0 {
				selection := rowState.selection.Progress(animGtx, selected, motion.ControlDuration, motion)
				recorded = append(recorded, t.recordSelectionCell(ctx, gtx, stateValue, columns.selection, selection, defaults.Geometry))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			for index, cell := range row.Cells {
				recorded = append(recorded, t.recordCell(ctx, gtx, columns.widths[index], t.columns[index], cell, rowForeground, defaults))
				rowHeight = max(rowHeight, recorded[len(recorded)-1].dims.Size.Y)
			}
			rowHeight = min(rowHeight, gtx.Constraints.Max.Y)
			size := image.Pt(columns.width, rowHeight)
			focus := rowState.focus.Opacity(animGtx, focused && !disabled, motion)
			opacity := paint.PushOpacity(gtx.Ops, tableOpacity(rowStyle))
			showSeparator := t.variant == VariantSecondary || !last
			if t.gridLinesSet {
				showSeparator = t.gridLines && !last
			}
			separatorColor := tableBackground(styles.separator, color.NRGBA{})
			focusColor := tableBackground(styles.focus, rowForeground)
			tableDrawRow(
				gtx,
				size,
				rowBackground,
				separatorColor,
				max(gtx.Dp(defaults.Geometry.SeparatorWidth), 1),
				showSeparator,
				focusColor,
				focus,
				defaults.Geometry.FocusRingWidth,
				defaults.Geometry.FocusRadius,
			)
			if t.showsFullGrid() {
				tableDrawRowSeparators(gtx, columns.selection, columns.widths, rowHeight, max(gtx.Dp(defaults.Geometry.SeparatorWidth), 1), separatorColor)
			}
			rowState.interactiveCells = rowState.interactiveCells[:0]
			rowState.focusTargets = rowState.focusTargets[:0]
			x := 0
			for _, cell := range recorded {
				y := max((rowHeight-cell.dims.Size.Y)/2, 0)
				if cell.interactive {
					rowState.interactiveCells = append(rowState.interactiveCells, image.Rect(x, 0, x+cell.dims.Size.X, rowHeight))
					rowState.focusTargets = append(rowState.focusTargets, cell.focusTargets...)
				}
				cell.placement.PlaceOffset(image.Pt(x, y))
				offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
				cell.call.Add(gtx.Ops)
				offset.Pop()
				x += cell.dims.Size.X
			}
			opacity.Pop()
			return layout.Dimensions{Size: size}
		})
	})
	if t.rowContextMenu == nil {
		return trigger.Layout(ctx, gtx)
	}
	owner := core.FullKey(ctx, t.key)
	key := core.DerivedKey(ctx, owner, "row-context-menu:"+row.Key)
	return menu.ContextMenu(key, trigger, t.rowContextMenu(row)).
		FocusTarget(&rowState.clickable).
		FocusTargets(rowState.focusTargets...).
		Disabled(disabled).
		Layout(ctx, gtx)
}

func (t Widget) recordSelectionCell(ctx *core.Context, gtx layout.Context, stateValue *tableState, width int, selection float32, geometry tableGeometry) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints = layout.Exact(image.Pt(width, t.rowMinHeight(gtx, geometry)))
	toggle.DrawCheckboxControl(ctx, cellGtx, toggle.CheckboxControlOptions{
		Variant: toggle.CheckboxSecondary, Selection: selection,
		VisualCache: &stateValue.checkboxVisual,
	})
	return recordedCell{call: macro.Stop(), dims: layout.Dimensions{Size: cellGtx.Constraints.Max}}
}

func (t Widget) rowMinHeight(gtx layout.Context, geometry tableGeometry) int {
	height := geometry.RowMinHeight
	if t.rowHeight > 0 {
		height = unit.Dp(t.rowHeight)
	}
	return gtx.Dp(height)
}

func (t Widget) recordCell(ctx *core.Context, gtx layout.Context, width int, column Column, cell Cell, foreground color.NRGBA, defaults tableDefaults) recordedCell {
	macro := op.Record(gtx.Ops)
	cellGtx := gtx
	cellGtx.Constraints.Min = image.Point{}
	cellGtx.Constraints.Max.X = width
	content := cell.Content
	if content == nil {
		content = tableCellText{text: cell.Text, size: defaults.Geometry.CellTextSize, color: foreground, weight: font.Normal}
	}
	inset := layout.Inset{
		Top: defaults.Geometry.CellPaddingY, Right: defaults.Geometry.CellPaddingX,
		Bottom: defaults.Geometry.CellPaddingY, Left: defaults.Geometry.CellPaddingX,
	}
	collector := new(core.FocusCollector)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		restore := core.PushFocusCollector(ctx, collector)
		defer restore()
		return layoutui.TrackedInset(ctx, cellGtx, inset, func(gtx layout.Context) layout.Dimensions {
			return alignWidget(ctx, gtx, column.Align, content)
		})
	})
	dims.Size.X = width
	return recordedCell{call: macro.Stop(), dims: dims, placement: placement, interactive: cell.Interactive, focusTargets: collector.Targets}
}

func alignWidget(ctx *core.Context, gtx layout.Context, alignment Alignment, child core.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	direction := layout.W
	switch alignment {
	case AlignCenter:
		direction = layout.Center
	case AlignEnd:
		direction = layout.E
	}
	return layoutui.TrackedDirection(ctx, gtx, direction, func(gtx layout.Context) layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
}

type tableCellText struct {
	text   string
	size   unit.Sp
	color  color.NRGBA
	weight font.Weight
}

func (t tableCellText) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	label := material.Label(theme.MaterialOf(core.ActiveTheme(ctx)), t.size, t.text)
	label.Color = t.color
	label.Font.Weight = t.weight
	label.MaxLines = 1
	label.Truncator = "..."
	return label.Layout(gtx)
}

func (t Widget) resolveColumns(ctx *core.Context, gtx layout.Context, stateValue *tableState, available int, geometry tableGeometry) tableColumns {
	result := tableColumns{widths: make([]int, len(t.columns))}
	fixed := make([]bool, len(t.columns))
	maxWidths := make([]int, len(t.columns))
	if stateValue != nil {
		defer func() {
			for index, column := range t.columns {
				if !column.Resizable {
					continue
				}
				configuredWidth := 0
				if column.Width > 0 {
					configuredWidth = gtx.Dp(unit.Dp(column.Width))
				}
				stateValue.column(column.Key).resize.sync(configuredWidth, result.widths[index])
			}
		}()
	}
	if t.showsSelectionIndicator() {
		result.selection = max(gtx.Dp(geometry.SelectionColumnWidth), 0)
	}
	minimum := result.selection
	lastFlex := -1
	for index, column := range t.columns {
		minWidth := max(gtx.Dp(geometry.MinColumnWidth), 1)
		if column.MinWidth > 0 {
			minWidth = max(gtx.Dp(unit.Dp(column.MinWidth)), 1)
		}
		configuredWidth := 0
		if column.Width > 0 {
			configuredWidth = gtx.Dp(unit.Dp(column.Width))
		}
		width := configuredWidth
		hasWidth := width > 0
		if column.Resizable && stateValue != nil {
			if resized, ok := stateValue.column(column.Key).resize.resolvedWidth(configuredWidth); ok {
				width, hasWidth = resized, true
			}
		}
		maxWidth := 0
		if column.MaxWidth > 0 {
			maxWidth = max(gtx.Dp(unit.Dp(column.MaxWidth)), minWidth)
		}
		maxWidths[index] = maxWidth
		if hasWidth {
			result.widths[index] = max(width, minWidth)
			if maxWidth > 0 {
				result.widths[index] = min(result.widths[index], maxWidth)
			}
			fixed[index] = true
		} else {
			result.widths[index] = minWidth
			lastFlex = index
		}
		minimum += result.widths[index]
	}
	result.width = max(available, minimum)
	if t.minWidth > 0 {
		result.width = max(result.width, gtx.Dp(unit.Dp(t.minWidth)))
	}
	extra := result.width - minimum
	if extra <= 0 {
		return result
	}
	remaining := extra
	for remaining > 0 && lastFlex >= 0 {
		weightRemaining := float32(0)
		lastCandidate := -1
		for index, column := range t.columns {
			if fixed[index] || maxWidths[index] > 0 && result.widths[index] >= maxWidths[index] {
				continue
			}
			weight := column.Weight
			if weight <= 0 {
				weight = 1
			}
			weightRemaining += weight
			lastCandidate = index
		}
		if lastCandidate < 0 {
			break
		}
		passRemaining := remaining
		distributed := 0
		for index, column := range t.columns {
			if fixed[index] || maxWidths[index] > 0 && result.widths[index] >= maxWidths[index] {
				continue
			}
			weight := column.Weight
			if weight <= 0 {
				weight = 1
			}
			share := passRemaining
			if index != lastCandidate {
				share = int(float32(passRemaining) * weight / weightRemaining)
			}
			if maxWidths[index] > 0 {
				share = min(share, maxWidths[index]-result.widths[index])
			}
			result.widths[index] += share
			passRemaining -= share
			remaining -= share
			distributed += share
			weightRemaining -= weight
		}
		if distributed == 0 {
			break
		}
	}
	return result
}
