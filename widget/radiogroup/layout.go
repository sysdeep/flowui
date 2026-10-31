package radiogroup

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

func (r Widget) layoutItem(ctx *core.Context, gtx layout.Context, group *radioGroupState, item RadioItem, selectedKey string, defaults radioDefaults) layout.Dimensions {
	itemState := group.item(item.Key)
	disabled := r.disabled || item.Disabled
	invalid := r.invalid || item.Invalid
	selected := item.Key == selectedKey
	animGtx := gtx

	interaction := base.RowInteraction(ctx, gtx, &itemState.clickable, disabled, func() {
		if !selected {
			group.requestSelectedKey(r, item.Key)
		}
	})
	disabled = interaction.Disabled
	if disabled {
		gtx = gtx.Disabled()
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := interaction.Focused
		focusVisible := interaction.FocusVisible
		styleState := style.State{
			Hovered: interaction.Hovered, Pressed: interaction.Pressed,
			Focused: focused, FocusVisible: focusVisible, Disabled: disabled,
			Selected: selected, Checked: selected, Invalid: invalid,
		}
		key := core.DerivedKey(ctx, core.FullKey(ctx, r.key), item.Key)
		resolved := r.resolveItemStyle(ctx, animGtx, key, styleState, defaults)
		resolved.selected = itemState.selectionFor(animGtx, selected, defaults.SelectionDuration, defaults.Motion)
		resolved.focus = itemState.focusOpacity(animGtx, focusVisible && !disabled, defaults.FocusDuration, defaults.Motion)
		scale := radioPressScale(animGtx, itemState.clickable.History(), defaults, disabled)
		dims := r.layoutItemContent(ctx, gtx, item, resolved, scale, defaults)
		semanticClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.RadioButton.Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semanticClip.Pop()
		return dims
	})
}

func (r Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, item RadioItem, style radioStyle, scale float32, defaults radioDefaults) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(defaults.Metrics.ContentGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRadio(ctx, gtx, defaults, style, scale)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.Description == "" {
				return layoutui.Resolved(ctx, gtx, style.label, text.New(item.Label))
			}
			return layout.Flex{
				Axis: layout.Vertical,
				Gap:  gtx.Dp(defaults.Metrics.DescriptionGap),
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.Resolved(ctx, gtx, style.label, text.New(item.Label))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.Resolved(ctx, gtx, style.description, text.New(item.Description))
				}),
			)
		}),
	)
}
