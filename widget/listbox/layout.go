package listbox

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/optionrow"
	"github.com/qianniancn/flowui/widget/internal/textrender"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

func (l Widget) layout(ctx *core.Context, gtx layout.Context, styleKey string, state *listBoxState, defaults listBoxDefaults, entries []listBoxEntry, hasItems bool, instanceStyle style.Style) layout.Dimensions {
	l.applyConstraints(&gtx, defaults)

	macro := op.Record(gtx.Ops)
	dims := l.layoutContent(ctx, gtx, styleKey, state, defaults, entries, hasItems, instanceStyle)
	call := macro.Stop()

	dims.Size = gtx.Constraints.Constrain(dims.Size)
	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (l Widget) applyConstraints(gtx *layout.Context, defaults listBoxDefaults) {
	if l.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	maxHeight := defaults.maxHeight
	if l.maxHeight > 0 {
		maxHeight = unit.Dp(l.maxHeight)
	}
	if maxHeight <= 0 {
		return
	}
	gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
	gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
}

func (l Widget) layoutContent(ctx *core.Context, gtx layout.Context, styleKey string, state *listBoxState, defaults listBoxDefaults, entries []listBoxEntry, hasItems bool, instanceStyle style.Style) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	padding := defaults.padding
	if l.hasPadding {
		padding = l.padding
	}
	return layout.UniformInset(padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !hasItems {
			state.scrollToKey = ""
			return l.layoutEmpty(ctx, gtx, active, defaults)
		}
		state.list.Axis = layout.Vertical
		state.list.Gap = gtx.Dp(defaults.gap)
		state.list.Alignment = layout.Start
		state.list.ScrollToEnd = false
		state.list.ScrollAnyAxis = false
		if l.scrollToKey == "" {
			state.scrollToKey = ""
		} else {
			for index, entry := range entries {
				if entry.header || entry.item.Key != l.scrollToKey {
					continue
				}
				if l.scrollToKey != state.scrollToKey || index != state.scrollToIndex {
					position := state.list.Position
					if index <= position.First || index >= position.First+position.Count-1 {
						state.list.ScrollTo(index)
					}
					state.scrollToKey, state.scrollToIndex = l.scrollToKey, index
				}
				break
			}
		}
		return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(entries), l.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
			entry := entries[index]
			if entry.header {
				return l.layoutSectionHeader(gtx, active, defaults, entry.title)
			}
			return l.layoutItem(ctx, gtx, styleKey, state, active, defaults, instanceStyle, entry.item)
		})
	})
}

type listBoxEntry struct {
	header bool
	title  string
	item   Item
}

func (l Widget) layoutSectionHeader(gtx layout.Context, active *theme.Theme, defaults listBoxDefaults, title string) layout.Dimensions {
	return layout.Inset{
		Top:    defaults.sectionHeaderPaddingY,
		Right:  defaults.sectionHeaderPaddingX,
		Bottom: defaults.sectionHeaderPaddingY,
		Left:   defaults.sectionHeaderPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return l.layoutText(gtx, active, defaults.sectionHeaderTextSize, font.Normal, active.Semantic().MutedForeground, title)
	})
}

func (l Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, active *theme.Theme, defaults listBoxDefaults) layout.Dimensions {
	height := min(gtx.Dp(defaults.itemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value := l.emptyText
		if !l.hasEmptyText {
			value = ctx.Text(i18n.CollectionNoItems)
		}
		return l.layoutText(gtx, active, defaults.itemTextSize, font.Normal, active.Semantic().MutedForeground, value)
	})
}

func (l Widget) layoutItem(ctx *core.Context, gtx layout.Context, styleKey string, listState *listBoxState, active *theme.Theme, defaults listBoxDefaults, itemStyle style.Style, item Item) layout.Dimensions {
	itemState := listState.item(item.Key)
	disabled := l.itemDisabled(item)
	selected := l.isSelected(item.Key)
	animGtx := gtx
	interaction := base.RowInteraction(ctx, gtx, &itemState.Clickable, disabled, func() { l.activate(listState, item.Key) })
	disabled = interaction.Disabled
	if disabled {
		gtx = gtx.Disabled()
	}
	focused, focusVisible := interaction.Focused, interaction.FocusVisible

	itemState.FocusOpacity(animGtx, focusVisible && !disabled, defaults.motion)
	styleState := style.State{
		Hovered:      interaction.Hovered,
		Pressed:      interaction.Pressed,
		Focused:      focused,
		FocusVisible: !disabled && focusVisible,
		Disabled:     disabled,
		Selected:     selected,
	}
	if itemState.styleKey == "" {
		itemState.styleKey = core.DerivedKey(ctx, styleKey, "item:"+item.Key)
		partKey := core.DerivedKey(ctx, itemState.styleKey, "style-part:"+string(style.PartItem))
		itemState.transitionKey = core.DerivedKey(ctx, partKey, "style-transition")
	}
	if itemState.transitionKey == "" {
		partKey := core.DerivedKey(ctx, itemState.styleKey, "style-part:"+string(style.PartItem))
		itemState.transitionKey = core.DerivedKey(ctx, partKey, "style-transition")
	}
	itemDefaults := defaults.item(item.Variant)
	part := listState.resolvedItemPart(
		ctx,
		animGtx,
		active,
		itemState.transitionKey,
		item.Variant,
		styleState,
		itemDefaults,
		itemStyle,
	)
	palette := listBoxItemPalette(active, item.Variant, disabled, part)
	selectionProgress := itemState.Selection(animGtx, selected, defaults.selectionDuration, defaults.motion)

	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		return l.layoutItemContent(ctx, gtx, active, defaults, item, palette, selectionProgress, selected)
	})
	return layoutui.InteractiveResolved(ctx, gtx, part, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.Clickable.Layout(gtx, visual)
	})
}

func (l Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, active *theme.Theme, defaults listBoxDefaults, item Item, palette listBoxPalette, selectionProgress float32, selected bool) layout.Dimensions {
	var children [4]layout.FlexChild
	count := 0
	if item.Leading != nil {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Leading.Layout(ctx, gtx)
		})
		count++
	}
	children[count] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return l.layoutItemText(gtx, active, defaults, item, palette)
	})
	count++
	if item.Trailing != nil {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Trailing.Layout(ctx, gtx)
		})
		count++
	}
	if l.showIndicator(item) {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.layoutIndicator(ctx, gtx, defaults, item, palette, selectionProgress, selected)
		})
		count++
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(defaults.itemContentGap),
	}.Layout(gtx, children[:count]...)
}

func (l Widget) showIndicator(item Item) bool {
	if l.hideIndicator {
		return false
	}
	return item.Indicator != nil || l.selectionMode != SelectionNone
}

func (l Widget) layoutIndicator(ctx *core.Context, gtx layout.Context, defaults listBoxDefaults, item Item, palette listBoxPalette, selectionProgress float32, selected bool) layout.Dimensions {
	size := image.Pt(gtx.Dp(defaults.itemIndicatorSize), gtx.Dp(defaults.itemIndicatorSize))
	if item.Indicator == nil {
		optionrow.DrawCheck(gtx, size, selectionProgress, defaults.itemIndicatorInset, defaults.itemIndicatorStroke, palette.indicator)
		return layout.Dimensions{Size: size}
	}
	indicator := item.Indicator(selected)
	if indicator == nil {
		return layout.Dimensions{Size: size}
	}
	childGtx := gtx
	childGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(childGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
}

func (l Widget) layoutItemText(gtx layout.Context, active *theme.Theme, defaults listBoxDefaults, item Item, palette listBoxPalette) layout.Dimensions {
	if item.Description == "" {
		return l.layoutText(gtx, active, defaults.itemTextSize, font.Medium, palette.foreground, item.Label)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.layoutText(gtx, active, defaults.itemTextSize, font.Medium, palette.foreground, item.Label)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.layoutText(gtx, active, defaults.itemDescriptionSize, font.Normal, palette.description, item.Description)
		}),
	)
}

func (l Widget) layoutText(gtx layout.Context, active *theme.Theme, size unit.Sp, weight font.Weight, foreground color.NRGBA, value string) layout.Dimensions {
	label := material.Label(theme.MaterialOf(active), size, value)
	label.Color = foreground
	label.SelectionColor = active.Semantic().Selection
	label.Font.Weight = weight
	return label.Layout(gtx)
}

type listBoxPalette struct {
	foreground  color.NRGBA
	description color.NRGBA
	indicator   color.NRGBA
}

// listBoxItemPalette keeps the resolved item text color authoritative and
// falls back to the semantic variant palette only when the declaration does
// not resolve to a solid color.
func listBoxItemPalette(active *theme.Theme, variant ItemVariant, disabled bool, resolved style.ResolvedStyle) listBoxPalette {
	sem := active.Semantic()
	foreground, indicator := sem.Foreground, sem.Foreground
	description := sem.MutedForeground
	if variant == ItemDanger {
		danger := sem.Tone(theme.ToneDanger).Fill
		foreground, indicator = danger, danger
	}
	if disabled {
		foreground = active.DisabledColor(foreground)
		indicator = active.DisabledColor(indicator)
		description = active.DisabledColor(description)
	}
	if value, ok := textrender.ColorValue(resolved); ok {
		return listBoxPalette{foreground: value, description: value, indicator: value}
	}
	return listBoxPalette{foreground: foreground, description: description, indicator: indicator}
}
