package timeline

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	timelineDefaultTailWidth            unit.Dp = 1
	timelineDefaultDotBorderWidth       unit.Dp = 2
	timelineDefaultDotSize              unit.Dp = 10
	timelineDefaultItemPaddingBottom    unit.Dp = 16
	timelineDefaultTitleContentGap      unit.Dp = 8
	timelineDefaultSectionGap           unit.Dp = 20
	timelineDefaultInlineInset          unit.Dp = 12
	timelineDefaultHorizontalTitleGap   unit.Dp = 8
	timelineDefaultHorizontalContentGap unit.Dp = 8
	timelineDefaultHorizontalItemGap    unit.Dp = 24
	timelineDefaultProcessDashLength    unit.Dp = 4
	timelineDefaultProcessDashGap       unit.Dp = 4
)

// timelineMetrics contains geometry consumed by timeline layout and drawing.
type timelineMetrics struct {
	TailWidth            unit.Dp
	DotBorderWidth       unit.Dp
	DotSize              unit.Dp
	ItemPaddingBottom    unit.Dp
	TitleContentGap      unit.Dp
	SectionGap           unit.Dp
	InlineInset          unit.Dp
	HorizontalTitleGap   unit.Dp
	HorizontalContentGap unit.Dp
	HorizontalItemGap    unit.Dp
	ProcessDashLength    unit.Dp
	ProcessDashGap       unit.Dp
}

// timelineColors contains component-owned colors used by timeline behavior.
type timelineColors struct {
	Background    color.NRGBA
	Tail          color.NRGBA
	DotBackground color.NRGBA
	Title         color.NRGBA
	Content       color.NRGBA
	Muted         color.NRGBA
	Primary       color.NRGBA
	Error         color.NRGBA
	Success       color.NRGBA
	Disabled      color.NRGBA
}

// timelineRailRequest contains behavior-owned geometry for one rail.
type timelineRailRequest struct {
	Vertical      bool
	Center        int
	Start         int
	End           int
	MarkerX       int
	MarkerY       int
	Tail          bool
	Process       bool
	DotSize       int
	TailWidth     int
	BorderWidth   int
	DashLength    int
	DashGap       int
	DotBackground color.NRGBA
	TailColor     color.NRGBA
	MarkerColor   color.NRGBA
	CustomIcon    bool
	Filled        bool
}

// timelineVisual is the resolved default appearance of a timeline.
type timelineVisual struct {
	Defaults style.Style
	Metrics  timelineMetrics
	Colors   timelineColors
}

// buildTimeline returns the default appearance for the active theme.
func buildTimeline(active *theme.Theme) timelineVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	return timelineVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Metrics: timelineMetrics{
			TailWidth:            timelineDefaultTailWidth,
			DotBorderWidth:       timelineDefaultDotBorderWidth,
			DotSize:              timelineDefaultDotSize,
			ItemPaddingBottom:    timelineDefaultItemPaddingBottom,
			TitleContentGap:      timelineDefaultTitleContentGap,
			SectionGap:           timelineDefaultSectionGap,
			InlineInset:          timelineDefaultInlineInset,
			HorizontalTitleGap:   timelineDefaultHorizontalTitleGap,
			HorizontalContentGap: timelineDefaultHorizontalContentGap,
			HorizontalItemGap:    timelineDefaultHorizontalItemGap,
			ProcessDashLength:    timelineDefaultProcessDashLength,
			ProcessDashGap:       timelineDefaultProcessDashGap,
		},
		Colors: timelineColors{
			Background:    semantic.Background,
			Tail:          semantic.Border,
			DotBackground: semantic.Surface,
			Title:         semantic.Foreground,
			Content:       semantic.SurfaceForeground,
			Muted:         semantic.MutedForeground,
			Primary:       semantic.Tone(theme.TonePrimary).Fill,
			Error:         semantic.Tone(theme.ToneDanger).Fill,
			Success:       semantic.Tone(theme.ToneSuccess).Fill,
			Disabled:      semantic.MutedForeground,
		},
	}
}

func timelineRenderRail(gtx layout.Context, request timelineRailRequest) {
	dotSize := max(request.DotSize, 2)
	width := max(request.TailWidth, 1)
	if request.Tail {
		lineStart := request.Start + dotSize/2
		if request.Process {
			if request.Vertical {
				timelinePaintVerticalDashes(gtx, request.Center, lineStart, request.End, width, request.TailColor, max(request.DashLength, 2), max(request.DashGap, 2))
			} else {
				timelinePaintHorizontalDashes(gtx, lineStart, request.End, request.Center, width, request.TailColor, max(request.DashLength, 2), max(request.DashGap, 2))
			}
		} else if request.Vertical {
			paint.FillShape(gtx.Ops, request.TailColor, clip.Rect(image.Rect(
				request.Center-width/2,
				lineStart,
				request.Center+(width+1)/2,
				request.End,
			)).Op())
		} else {
			paint.FillShape(gtx.Ops, request.TailColor, clip.Rect(image.Rect(
				lineStart,
				request.Center-width/2,
				request.End,
				request.Center+(width+1)/2,
			)).Op())
		}
	}
	if request.CustomIcon {
		return
	}
	timelineDrawMarker(gtx, request, dotSize)
}

func timelinePaintVerticalDashes(gtx layout.Context, center, start, end, width int, value color.NRGBA, dashLength, dashGap int) {
	for y := start; y < end; y += dashLength + dashGap {
		dashEnd := min(y+dashLength, end)
		paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(
			center-width/2,
			y,
			center+(width+1)/2,
			dashEnd,
		)).Op())
	}
}

func timelinePaintHorizontalDashes(gtx layout.Context, start, end, center, width int, value color.NRGBA, dashLength, dashGap int) {
	for x := start; x < end; x += dashLength + dashGap {
		dashEnd := min(x+dashLength, end)
		paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(
			x,
			center-width/2,
			dashEnd,
			center+(width+1)/2,
		)).Op())
	}
}

func timelineDrawMarker(gtx layout.Context, request timelineRailRequest, size int) {
	outer := image.Rect(
		request.MarkerX-size/2,
		request.MarkerY-size/2,
		request.MarkerX+(size+1)/2,
		request.MarkerY+(size+1)/2,
	)
	paint.FillShape(gtx.Ops, request.MarkerColor, clip.Ellipse{Min: outer.Min, Max: outer.Max}.Op(gtx.Ops))
	if request.Filled {
		return
	}
	inner := size - max(request.BorderWidth*2, 2)
	if inner <= 0 {
		return
	}
	innerRect := image.Rect(
		request.MarkerX-inner/2,
		request.MarkerY-inner/2,
		request.MarkerX+(inner+1)/2,
		request.MarkerY+(inner+1)/2,
	)
	paint.FillShape(gtx.Ops, request.DotBackground, clip.Ellipse{Min: innerRect.Min, Max: innerRect.Max}.Op(gtx.Ops))
}
