package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/text"
)

func (c CheckboxWidget) layoutContent(ctx *core.Context, gtx layout.Context, options CheckboxControlOptions, indicatorState CheckboxIndicatorState, resolved checkboxResolvedStyle) layout.Dimensions {
	row := func(gtx layout.Context) layout.Dimensions {
		return c.layoutRow(ctx, gtx, options, indicatorState, resolved)
	}
	message := c.supportingText()
	if message == "" {
		return row(gtx)
	}
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(resolved.defaults.Metrics.DescriptionGap)}.Layout(gtx,
		layout.Rigid(row),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: resolved.defaults.Metrics.DescriptionIndent}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutui.Resolved(ctx, gtx, resolved.description, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return text.LayoutDescriptionWithResolvedText(ctx, gtx, text.NewDescription(message).
						For(c.key).
						Disabled(options.Disabled), resolved.description.Text)
				}))
			})
		}),
	)
}

func (c CheckboxWidget) layoutRow(ctx *core.Context, gtx layout.Context, options CheckboxControlOptions, indicatorState CheckboxIndicatorState, resolved checkboxResolvedStyle) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return DrawCheckboxControl(ctx, gtx, options)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(resolved.defaults.Metrics.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved.label, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				children := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return text.WithResolvedText(text.New(c.label), resolved.label.Text).Layout(ctx, gtx)
				})}
				if c.required {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						color := resolved.defaults.Colors.RequiredMark
						if indicatorState.Disabled {
							color = resolved.defaults.Colors.DisabledRequiredMark
						}
						return text.New("*").Color(color).Layout(ctx, gtx)
					}))
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline, Gap: gtx.Dp(resolved.defaults.Metrics.RequiredMarkOffset)}.Layout(gtx, children...)
			}))
		}),
	)
}
