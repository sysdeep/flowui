package candlestick

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func (w Widget) drawTooltip(ctx *core.Context, selection chartSelection, interval float64, defaults candlestickDefaults, anchor image.Rectangle, progress float32, exiting bool) {
	content := w.tooltipWidget(ctx, selection, interval, defaults)
	if content == nil {
		return
	}
	tooltip.NewPopup(content).
		Placement(core.PopoverRightStart).
		Offset(max(defaults.Metrics.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Overlay(ctx, core.FullKey(ctx, "tooltip:"+w.key), anchor)
}

func (w Widget) tooltipWidget(ctx *core.Context, selection chartSelection, interval float64, defaults candlestickDefaults) core.Widget {
	if w.tooltipContent != nil {
		return w.tooltipContent(w.publicSelectionFor(ctx, selection))
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutTooltipContent(ctx, gtx, selection, interval, defaults)
	})
}

func (w Widget) layoutTooltipContent(ctx *core.Context, gtx layout.Context, selection chartSelection, interval float64, defaults candlestickDefaults) layout.Dimensions {
	textColor := defaults.Colors.OverlayForeground
	gap := max(gtx.Dp(defaults.Metrics.TooltipRowGap), 0)
	values := w.candlestickTooltipRows(ctx, selection, interval)
	rows := make([]chart.TooltipRow, len(values))
	for index, value := range values {
		rows[index].Text = value
	}
	return chart.LayoutTooltipRows(ctx, gtx, w.categoryLabel(selection.index), rows, defaults.Metrics.TooltipTextSize, textColor, 0, gap, chart.TooltipMarkerNone)
}

func (w Widget) candlestickTooltipRows(ctx *core.Context, selection chartSelection, interval float64) []string {
	return []string{
		ctx.Textf(i18n.ChartCandlestickOpen, w.yLabel(selection.candle.open, interval)),
		ctx.Textf(i18n.ChartCandlestickHigh, w.yLabel(selection.candle.high, interval)),
		ctx.Textf(i18n.ChartCandlestickLow, w.yLabel(selection.candle.low, interval)),
		ctx.Textf(i18n.ChartCandlestickClose, w.yLabel(selection.candle.close, interval)),
	}
}
