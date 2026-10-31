package selector

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

func (s SelectWidget) groupTitle(ctx *core.Context, title string) core.Widget {
	return layoutui.Box(text.New(title).Color(core.ActiveTheme(ctx).Colors.MutedForeground)).
		Style(style.New().Padding(6, 10))
}

func (s SelectWidget) layoutTriggerContent(
	ctx *core.Context,
	gtx layout.Context,
	state base.SelectState,
	defaults selectDefaults,
	trigger, placeholder, icon, indicator style.ResolvedStyle,
) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	sem := active.Semantic()
	value, selected := s.displayValue()
	foreground := selectResolvedColor(trigger, sem.FieldForeground)
	if !selected {
		foreground = selectResolvedColor(placeholder, sem.FieldPlaceholder)
	}
	if state.Disabled {
		foreground = active.DisabledColor(foreground)
	}
	size, weight := defaults.textSize, font.Normal
	if trigger.Text != nil {
		if trigger.Text.FontSize != nil && *trigger.Text.FontSize > 0 {
			size = *trigger.Text.FontSize
		}
		if trigger.Text.FontWeight != nil {
			weight = font.Weight(*trigger.Text.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, value)
	label.Color = foreground
	label.Font.Weight = weight
	label.MaxLines = 1
	label.Truncator = "…"

	iconColor := selectResolvedColor(icon, sem.FieldPlaceholder)
	if state.Disabled {
		iconColor = active.DisabledColor(iconColor)
	}
	iconSize := selectPartSize(gtx, icon, defaults.indicatorSize)
	slotWidth := min(selectPartWidth(gtx, indicator, defaults.indicatorWidth), gtx.Constraints.Max.X)
	rotation := float32(0)
	if icon.Trans != nil && icon.Trans.Rotate != nil {
		rotation = *icon.Trans.Rotate
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(defaults.contentGap),
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			slotGtx := gtx
			slotGtx.Constraints = layout.Exact(image.Pt(slotWidth, iconSize))
			return layout.Center.Layout(slotGtx, func(gtx layout.Context) layout.Dimensions {
				DrawChevron(gtx, iconSize, defaults.indicatorStroke, iconColor, rotation)
				return layout.Dimensions{Size: image.Pt(iconSize, iconSize)}
			})
		}),
	)
}

func (s SelectWidget) panel(state base.SelectState, content core.Widget, defaults selectDefaults, triggerWidth *int) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		panelGtx := gtx
		if width := *triggerWidth; width > 0 {
			width = min(width, panelGtx.Constraints.Max.X)
			panelGtx.Constraints.Min.X = width
			panelGtx.Constraints.Max.X = width
		}
		resolved := computed.ResolvePart(
			ctx,
			panelGtx,
			state.Key,
			style.PartPanel,
			style.State{Open: state.Open, Disabled: state.Disabled},
			defaults.panel,
			style.Style{},
			style.Style{},
			s.style,
		)
		return layoutui.Resolved(ctx, panelGtx, resolved, content)
	})
}

func (s SelectWidget) item(state base.SelectItemState, defaults selectDefaults) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		active := core.ActiveTheme(ctx)
		resolved := computed.ResolvePart(
			ctx,
			gtx,
			state.Key,
			style.PartItem,
			style.State{
				Hovered:      state.Hovered,
				Pressed:      state.Pressed,
				Focused:      state.Focused,
				FocusVisible: state.FocusVisible,
				Disabled:     state.Disabled,
				Selected:     state.Selected || state.Highlighted,
			},
			defaults.item,
			style.Style{},
			style.Style{},
			s.style,
		)
		foreground := selectResolvedColor(resolved, active.Semantic().Foreground)
		if state.Disabled {
			foreground = active.DisabledColor(foreground)
		}
		size, weight := defaults.itemTextSize, font.Medium
		if resolved.Text != nil {
			if resolved.Text.FontSize != nil && *resolved.Text.FontSize > 0 {
				size = *resolved.Text.FontSize
			}
			if resolved.Text.FontWeight != nil {
				weight = font.Weight(*resolved.Text.FontWeight)
			}
		}
		label := material.Label(theme.MaterialOf(active), size, state.Label)
		label.Color = foreground
		label.Font.Weight = weight
		label.MaxLines = 1
		label.Truncator = "…"
		return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return label.Layout(gtx)
		}))
	})
}

// RenderItem renders one option with the shared Select item appearance. It is
// exported for select-like controls that own a different interaction model.
func RenderItem(ctx *core.Context, gtx layout.Context, state base.SelectItemState, appearance Appearance, custom style.Style) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	resolved := computed.ResolvePart(
		ctx, gtx, state.Key, style.PartItem,
		style.State{
			Hovered: state.Hovered, Pressed: state.Pressed,
			Focused: state.Focused, FocusVisible: state.FocusVisible,
			Disabled: state.Disabled, Selected: state.Selected || state.Highlighted,
		},
		appearance.Item, style.Style{}, style.Style{}, custom,
	)
	foreground := selectResolvedColor(resolved, active.Semantic().Foreground)
	if state.Disabled {
		foreground = active.DisabledColor(foreground)
	}
	size := appearance.ItemTextSize
	weight := font.Medium
	if resolved.Text != nil {
		if resolved.Text.FontSize != nil && *resolved.Text.FontSize > 0 {
			size = *resolved.Text.FontSize
		}
		if resolved.Text.FontWeight != nil {
			weight = font.Weight(*resolved.Text.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, state.Label)
	label.Color = foreground
	label.Font.Weight = weight
	label.MaxLines = 1
	label.Truncator = "..."
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return label.Layout(gtx)
	}))
}
