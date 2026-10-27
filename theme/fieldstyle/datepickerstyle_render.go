package fieldstyle

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

func dateCalendarIconRenderer(gtx layout.Context, request theme.DatePickerIconRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(request.Size)
	icon.Layout(lucide.Calendar, iconGtx, request.Color)
}

func dateYearIndicatorRenderer(gtx layout.Context, request theme.DatePickerYearIndicatorRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	data := lucide.ChevronRight
	if request.Open {
		data = lucide.ChevronDown
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(request.Size)
	icon.Layout(data, iconGtx, request.Color)
}

func dateNavButtonRenderer(gtx layout.Context, request theme.DatePickerNavButtonRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	if request.Background.A != 0 {
		paint.FillShape(gtx.Ops, request.Background,
			clip.UniformRRect(image.Rectangle{Max: request.Size}, request.Radius).Op(gtx.Ops))
	}
	data := lucide.ChevronRight
	if request.Direction < 0 {
		data = lucide.ChevronLeft
	}
	diameter := min(icon.LucideSizeForStroke(gtx, request.Metrics.NavChevronStroke), min(request.Size.X, request.Size.Y))
	if diameter <= 0 {
		return
	}
	offset := op.Offset(image.Pt((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(data, iconGtx, request.Foreground)
	offset.Pop()
}

func dateCellRenderer(gtx layout.Context, request theme.DatePickerCellRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 || request.Background.A == 0 {
		return
	}
	radius := min(request.Radius, min(request.Size.X, request.Size.Y)/2)
	paint.FillShape(gtx.Ops, request.Background,
		clip.UniformRRect(image.Rectangle{Max: request.Size}, radius).Op(gtx.Ops))
}

func dateRangeRenderer(gtx layout.Context, request theme.DatePickerRangeRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 4 {
		return
	}
	rect := image.Rect(0, 2, request.Size.X, request.Size.Y-2)
	startRadius := min(max(request.StartRadius, 0), rect.Dy()/2)
	endRadius := min(max(request.EndRadius, 0), rect.Dy()/2)
	paint.FillShape(gtx.Ops, request.Color, clip.RRect{
		Rect: rect,
		NW:   startRadius,
		SW:   startRadius,
		NE:   endRadius,
		SE:   endRadius,
	}.Op(gtx.Ops))
}

func dateStrikeRenderer(gtx layout.Context, request theme.DatePickerStrikeRenderRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	color := request.Color
	color.A = byte(uint16(color.A) * 3 / 4)
	var path clip.Path
	path.Begin(gtx.Ops)
	y := float32(request.Size.Y) / 2
	halfSize := draw.DpFloat(gtx, request.HalfSize)
	path.MoveTo(f32.Pt(float32(request.Size.X)/2-halfSize, y))
	path.LineTo(f32.Pt(float32(request.Size.X)/2+halfSize, y))
	stroke := clip.Stroke{Path: path.End(), Width: max(draw.DpFloat(gtx, request.Width), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, color)
	stroke.Pop()
}

func dateFocusRenderer(gtx layout.Context, request theme.DatePickerFocusRenderRequest) {
	if !request.Visible || request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: request.Size}
	radius := request.Radius
	if request.Trigger {
		diameter := min(gtx.Dp(request.IconSize+8), min(request.Size.X, request.Size.Y))
		rect = image.Rect((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2,
			(dateSizeCenter(request.Size.X, diameter)), (dateSizeCenter(request.Size.Y, diameter)))
		radius = gtx.Dp(request.SegmentRadius)
	}
	width := max(gtx.Dp(request.FocusWidth), 1)
	rect = rect.Inset((width + 1) / 2)
	if rect.Empty() {
		return
	}
	radius = max(radius-(width+1)/2, 0)
	draw.RoundedInsetStroke(gtx, rect, radius, width, 0, request.Color)
}

func dateSizeCenter(size, span int) int { return (size + span) / 2 }

var _ theme.DatePickerIconRenderer = dateCalendarIconRenderer
var _ theme.DatePickerYearIndicatorRenderer = dateYearIndicatorRenderer
var _ theme.DatePickerNavButtonRenderer = dateNavButtonRenderer
var _ theme.DatePickerCellRenderer = dateCellRenderer
var _ theme.DatePickerRangeRenderer = dateRangeRenderer
var _ theme.DatePickerStrikeRenderer = dateStrikeRenderer
var _ theme.DatePickerFocusRenderer = dateFocusRenderer
