package nodegraph

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	nodeGraphSelectionOpacity = 0.22
	nodeGraphShadowOpacity    = 20
)

// nodeGraphColors contains the palette decisions used by graph drawing.
// Explicit port and edge colors remain widget inputs.
type nodeGraphColors struct {
	CanvasBackground    color.NRGBA
	CanvasBorder        color.NRGBA
	Grid                color.NRGBA
	NodeBackground      color.NRGBA
	NodeBorder          color.NRGBA
	NodeForeground      color.NRGBA
	NodeMutedForeground color.NRGBA
	PortColor           color.NRGBA
	PortBorder          color.NRGBA
	EdgeColor           color.NRGBA
	SelectedEdgeColor   color.NRGBA
	SelectedNodeBorder  color.NRGBA
	SelectionFill       color.NRGBA
	SelectionBorder     color.NRGBA
	ControlBackground   color.NRGBA
	ControlForeground   color.NRGBA
	ControlBorder       color.NRGBA
	Shadow              color.NRGBA
	ResizeHandleFill    color.NRGBA
}

// nodeGraphMetrics contains the geometry and typography consumed by graph
// layout and drawing. Values are resolved from the active theme once per
// theme and shared by every instance.
type nodeGraphMetrics struct {
	CanvasRadius            unit.Dp
	CanvasBorderWidth       unit.Dp
	GridWidth               unit.Dp
	GridMinimumStep         unit.Dp
	EdgeWidth               unit.Dp
	EdgeMinimumDistance     unit.Dp
	EdgeLabelWidth          unit.Dp
	EdgeLabelHeight         unit.Dp
	NodeHeaderHeight        unit.Dp
	NodeVerticalPadding     unit.Dp
	NodePortRowHeight       unit.Dp
	NodeDefaultWidth        unit.Dp
	NodeDefaultHeight       unit.Dp
	NodePadding             unit.Dp
	NodeRadius              unit.Dp
	NodeBorderWidth         unit.Dp
	NodeTitleSize           unit.Sp
	NodePortTextSize        unit.Sp
	NodeTextMinScale        float32
	NodeTextMaxScale        float32
	PortRadius              unit.Dp
	PortBorderWidth         unit.Dp
	EdgeMarkerSize          unit.Dp
	SelectionWidth          unit.Dp
	SelectionDashLength     unit.Dp
	SelectionDashGap        unit.Dp
	FocusInset              unit.Dp
	FocusWidth              unit.Dp
	FocusDashLength         unit.Dp
	FocusDashGap            unit.Dp
	ResizeHandleRadius      unit.Dp
	ResizeHandleBorderWidth unit.Dp
	ControlWidth            unit.Dp
	ControlHeight           unit.Dp
	ControlMargin           unit.Dp
	ControlRadius           unit.Dp
	ControlDividerInset     unit.Dp
	ControlTextSize         unit.Sp
	ControlFitSpan          float32
	ControlFitArm           float32
	PanelPadding            unit.Dp
	ToolbarOffset           unit.Dp
	MinimapWidth            unit.Dp
	MinimapHeight           unit.Dp
	MinimapMargin           unit.Dp
	MinimapPadding          unit.Dp
	MinimapRadius           unit.Dp
	MinimapBorderWidth      unit.Dp
	MinimapEdgeWidth        unit.Dp
	MinimapViewportWidth    unit.Dp
	ShadowOffsetY           unit.Dp
	ShadowBlur              unit.Dp
	ShadowSpread            unit.Dp
}

// nodeGraphCanvas is the value passed to the canvas drawing function.
type nodeGraphCanvas struct {
	Size        image.Point
	Radius      int
	BorderWidth int
	Background  color.NRGBA
	Border      color.NRGBA
}

// nodeGraphGrid is the value passed to the grid drawing function.
type nodeGraphGrid struct {
	Size    image.Point
	Pattern NodeGraphGridPattern
	StartX  float32
	StartY  float32
	Step    float32
	Width   float32
	Color   color.NRGBA
}

// nodeGraphEdge is the value passed to the edge drawing function.
type nodeGraphEdge struct {
	EdgeType        NodeGraphEdgeType
	From            f32.Point
	To              f32.Point
	Points          []f32.Point
	Width           float32
	MinimumDistance float32
	Dashed          bool
	Phase           float32
	Color           color.NRGBA
	SourceMarker    NodeGraphEdgeMarker
	TargetMarker    NodeGraphEdgeMarker
	MarkerSize      float32
	Label           string
	LabelRect       image.Rectangle
	LabelSize       unit.Sp
	LabelColor      color.NRGBA
	LabelWeight     font.Weight
}

// nodeGraphConnection is the value passed to the connection preview function.
type nodeGraphConnection struct {
	From            f32.Point
	To              f32.Point
	Width           float32
	MinimumDistance float32
	Color           color.NRGBA
}

// nodeGraphNode is the value passed to the node shell drawing function.
type nodeGraphNode struct {
	Rect        image.Rectangle
	Radius      int
	BorderWidth int
	Background  color.NRGBA
	Border      color.NRGBA
	Shadow      color.NRGBA
	DrawShadow  bool
	Title       string
	TitleRect   image.Rectangle
	TitleSize   unit.Sp
	TitleColor  color.NRGBA
	TitleWeight font.Weight
}

// nodeGraphPort is the value passed to the port drawing function.
type nodeGraphPort struct {
	Center      f32.Point
	Radius      int
	BorderWidth int
	Color       color.NRGBA
	Border      color.NRGBA
	Label       string
	LabelRect   image.Rectangle
	LabelSize   unit.Sp
	LabelColor  color.NRGBA
	LabelWeight font.Weight
}

// nodeGraphSelection is the value passed to the selection outline function.
type nodeGraphSelection struct {
	Rect       image.Rectangle
	Fill       color.NRGBA
	Border     color.NRGBA
	Width      int
	DashLength int
	DashGap    int
	Filled     bool
}

// nodeGraphResize is the value passed to the resize handle function.
type nodeGraphResize struct {
	Rect        image.Rectangle
	Handles     uint16
	Radius      int
	BorderWidth int
	Border      color.NRGBA
	Fill        color.NRGBA
}

// nodeGraphControls is the value passed to the controls drawing function.
type nodeGraphControls struct {
	Rect         image.Rectangle
	ButtonWidth  int
	Radius       int
	DividerInset int
	TextSize     unit.Sp
	FitSpan      float32
	FitArm       float32
	Background   color.NRGBA
	Foreground   color.NRGBA
	Border       color.NRGBA
}

// nodeGraphMinimapNode is one node rectangle in minimap coordinates.
type nodeGraphMinimapNode struct {
	Rect     image.Rectangle
	Selected bool
}

// nodeGraphMinimapEdge is one edge segment in minimap coordinates.
type nodeGraphMinimapEdge struct {
	Points []f32.Point
	Color  color.NRGBA
}

// nodeGraphMinimap is the value passed to the minimap drawing function.
type nodeGraphMinimap struct {
	Rect          image.Rectangle
	Nodes         []nodeGraphMinimapNode
	Edges         []nodeGraphMinimapEdge
	Viewport      image.Rectangle
	Radius        int
	BorderWidth   int
	EdgeWidth     float32
	ViewportWidth float32
	Background    color.NRGBA
	Border        color.NRGBA
	Selected      color.NRGBA
}

// defaultNodeGraphDefaults builds the theme-derived appearance shared by every
// node graph instance.
func defaultNodeGraphDefaults(active *theme.Theme) nodeGraphDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	shadow := semantic.OverlayShadow
	if shadow.A == 0 {
		shadow = semantic.SurfaceShadow
	}
	colors := nodeGraphColors{
		CanvasBackground:    semantic.Background,
		CanvasBorder:        semantic.Border,
		Grid:                nodeGraphWithAlpha(semantic.MutedForeground, .35),
		NodeBackground:      semantic.FieldBackground,
		NodeBorder:          semantic.Border,
		NodeForeground:      semantic.SurfaceForeground,
		NodeMutedForeground: semantic.MutedForeground,
		PortColor:           semantic.Foreground,
		PortBorder:          semantic.Surface,
		EdgeColor:           semantic.Tone(theme.TonePrimary).Fill,
		SelectedEdgeColor:   semantic.Tone(theme.TonePrimary).Hover,
		SelectedNodeBorder:  semantic.MutedForeground,
		SelectionFill:       nodeGraphWithAlpha(semantic.Tone(theme.TonePrimary).Soft, nodeGraphSelectionOpacity),
		SelectionBorder:     semantic.Tone(theme.TonePrimary).Fill,
		ControlBackground:   semantic.FieldBackground,
		ControlForeground:   semantic.SurfaceForeground,
		ControlBorder:       semantic.Border,
		Shadow:              nodeGraphWithAlpha(shadow, float32(nodeGraphShadowOpacity)/255),
		ResizeHandleFill:    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	}
	metrics := nodeGraphMetrics{
		CanvasRadius:            8,
		CanvasBorderWidth:       1,
		GridWidth:               1,
		GridMinimumStep:         8,
		EdgeWidth:               1,
		EdgeMinimumDistance:     32,
		EdgeLabelWidth:          60,
		EdgeLabelHeight:         14,
		NodeHeaderHeight:        26,
		NodeVerticalPadding:     8,
		NodePortRowHeight:       20,
		NodeDefaultWidth:        150,
		NodeDefaultHeight:       64,
		NodePadding:             10,
		NodeRadius:              3,
		NodeBorderWidth:         1,
		NodeTitleSize:           12,
		NodePortTextSize:        10,
		NodeTextMinScale:        0.75,
		NodeTextMaxScale:        1.333333,
		PortRadius:              3,
		PortBorderWidth:         1,
		EdgeMarkerSize:          5,
		SelectionWidth:          1,
		SelectionDashLength:     2,
		SelectionDashGap:        2,
		FocusInset:              2,
		FocusWidth:              1,
		FocusDashLength:         2,
		FocusDashGap:            2,
		ResizeHandleRadius:      4,
		ResizeHandleBorderWidth: 1,
		ControlWidth:            108,
		ControlHeight:           36,
		ControlMargin:           12,
		ControlRadius:           5,
		ControlDividerInset:     8,
		ControlTextSize:         16,
		ControlFitSpan:          .18,
		ControlFitArm:           .8,
		PanelPadding:            12,
		ToolbarOffset:           12,
		MinimapWidth:            180,
		MinimapHeight:           120,
		MinimapMargin:           12,
		MinimapPadding:          8,
		MinimapRadius:           5,
		MinimapBorderWidth:      1,
		MinimapEdgeWidth:        .7,
		MinimapViewportWidth:    1,
		ShadowOffsetY:           1,
		ShadowBlur:              4,
		ShadowSpread:            1,
	}
	root := style.Style{}.
		FillWidth().
		Background(style.Solid(colors.CanvasBackground)).
		Opacity(1).
		Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
			Opacity(active.DisabledOpacityValue())})
	return nodeGraphDefaults{
		root:     root,
		metrics:  metrics,
		colors:   colors,
		material: theme.MaterialOf(active),
	}
}

func drawNodeGraphCanvas(gtx layout.Context, request nodeGraphCanvas) {
	if request.Size.X <= 0 || request.Size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: request.Size}
	paint.FillShape(gtx.Ops, request.Background, clip.UniformRRect(rect, request.Radius).Op(gtx.Ops))
	if request.BorderWidth <= 0 || request.Border.A == 0 {
		return
	}
	nodeGraphDrawRoundedStroke(gtx, rect, request.Radius, request.BorderWidth, request.Border)
}

func drawNodeGraphGrid(gtx layout.Context, request nodeGraphGrid) {
	if request.Step <= 0 || request.Size.X <= 0 || request.Size.Y <= 0 || request.Color.A == 0 {
		return
	}
	if request.Pattern == NodeGraphGridDots {
		for y := request.StartY; y < float32(request.Size.Y); y += request.Step {
			for x := request.StartX; x < float32(request.Size.X); x += request.Step {
				center := image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
				paint.FillShape(gtx.Ops, request.Color, clip.Ellipse(image.Rect(center.X-1, center.Y-1, center.X+1, center.Y+1)).Op(gtx.Ops))
			}
		}
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	for x := request.StartX; x < float32(request.Size.X); x += request.Step {
		path.MoveTo(f32.Pt(x, 0))
		path.LineTo(f32.Pt(x, float32(request.Size.Y)))
	}
	for y := request.StartY; y < float32(request.Size.Y); y += request.Step {
		path.MoveTo(f32.Pt(0, y))
		path.LineTo(f32.Pt(float32(request.Size.X), y))
	}
	stroke := clip.Stroke{Path: path.End(), Width: nodeGraphMaxFloat(request.Width, 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Color)
	stroke.Pop()
}

func drawNodeGraphEdge(materialTheme *material.Theme, gtx layout.Context, request nodeGraphEdge) {
	if request.Width <= 0 || request.Color.A == 0 {
		return
	}
	if request.Dashed {
		nodeGraphDrawDashedPolyline(gtx, request.Points, request.Width, request.Color, request.Phase)
	} else {
		switch request.EdgeType {
		case NodeGraphEdgeBezier:
			nodeGraphDrawBezier(gtx, request.From, request.To, request.Width, request.Color, request.MinimumDistance)
		case NodeGraphEdgeSmoothStep:
			nodeGraphDrawSmoothStep(gtx, request.From, request.To, request.Width, request.Color)
		default:
			nodeGraphDrawPolyline(gtx, request.Points, request.Width, request.Color)
		}
	}
	nodeGraphDrawMarker(gtx, request.SourceMarker, request.From, request.To, true, request.Color, request.MarkerSize, request.Points)
	nodeGraphDrawMarker(gtx, request.TargetMarker, request.From, request.To, false, request.Color, request.MarkerSize, request.Points)
	if request.Label != "" && !request.LabelRect.Empty() {
		nodeGraphDrawCenteredText(materialTheme, gtx, request.Label, request.LabelSize, request.LabelWeight, request.LabelColor, request.LabelRect)
	}
}

func drawNodeGraphConnection(gtx layout.Context, request nodeGraphConnection) {
	nodeGraphDrawBezier(gtx, request.From, request.To, request.Width, request.Color, request.MinimumDistance)
}

func drawNodeGraphNode(materialTheme *material.Theme, gtx layout.Context, metrics nodeGraphMetrics, request nodeGraphNode) {
	if request.Rect.Empty() {
		return
	}
	if request.DrawShadow && request.Shadow.A != 0 {
		shadow := draw.BoxShadow{
			OffsetY: float32(gtx.Dp(metrics.ShadowOffsetY)),
			Blur:    float32(gtx.Dp(metrics.ShadowBlur)),
			Spread:  float32(gtx.Dp(metrics.ShadowSpread)),
			Color:   request.Shadow,
		}
		draw.Shadow(gtx, request.Rect, draw.RoundedShadowCorners(metrics.NodeRadius, metrics.NodeRadius, metrics.NodeRadius, metrics.NodeRadius), shadow)
	}
	paint.FillShape(gtx.Ops, request.Background, clip.UniformRRect(request.Rect, request.Radius).Op(gtx.Ops))
	strokeInset := max((request.BorderWidth+1)/2, 1)
	strokeRect := request.Rect
	strokeRadius := request.Radius
	if request.Rect.Dx() > strokeInset*2 && request.Rect.Dy() > strokeInset*2 {
		strokeRect = request.Rect.Inset(strokeInset)
		strokeRadius = max(request.Radius-strokeInset, 0)
	}
	nodeGraphDrawRoundedStroke(gtx, strokeRect, strokeRadius, request.BorderWidth, request.Border)
	if request.Title != "" && !request.TitleRect.Empty() {
		nodeGraphDrawCenteredText(materialTheme, gtx, request.Title, request.TitleSize, request.TitleWeight, request.TitleColor, request.TitleRect)
	}
}

func drawNodeGraphPort(materialTheme *material.Theme, gtx layout.Context, request nodeGraphPort) {
	radius := max(request.Radius, 2)
	center := request.Center.Round()
	rect := image.Rect(center.X-radius, center.Y-radius, center.X+radius, center.Y+radius)
	paint.FillShape(gtx.Ops, request.Border, clip.Ellipse(rect).Op(gtx.Ops))
	inset := max(request.BorderWidth, 1)
	inner := rect.Inset(inset)
	if !inner.Empty() {
		paint.FillShape(gtx.Ops, request.Color, clip.Ellipse(inner).Op(gtx.Ops))
	}
	if request.Label != "" && !request.LabelRect.Empty() {
		nodeGraphDrawText(materialTheme, gtx, request.Label, request.LabelSize, request.LabelWeight, request.LabelColor, request.LabelRect)
	}
}

func drawNodeGraphSelection(gtx layout.Context, request nodeGraphSelection) {
	if request.Rect.Empty() {
		return
	}
	if request.Filled && request.Fill.A != 0 {
		paint.FillShape(gtx.Ops, request.Fill, clip.Rect(request.Rect).Op())
	}
	if request.Border.A == 0 || request.Width <= 0 {
		return
	}
	nodeGraphDrawDottedRect(gtx, request.Rect, request.Border, request.Width, request.DashLength, request.DashGap)
}

func drawNodeGraphResize(gtx layout.Context, request nodeGraphResize) {
	radius := max(request.Radius, 3)
	for bit := uint16(1); bit <= 128; bit <<= 1 {
		if request.Handles&bit == 0 {
			continue
		}
		center := nodeGraphResizeHandlePoint(request.Rect, bit)
		handleRect := image.Rect(center.X-radius, center.Y-radius, center.X+radius, center.Y+radius)
		paint.FillShape(gtx.Ops, request.Fill, clip.Rect(handleRect).Op())
		nodeGraphDrawRoundedStroke(gtx, handleRect, 1, max(request.BorderWidth, 1), request.Border)
	}
}

func drawNodeGraphControls(materialTheme *material.Theme, gtx layout.Context, request nodeGraphControls) {
	if request.Rect.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, request.Background, clip.UniformRRect(request.Rect, request.Radius).Op(gtx.Ops))
	nodeGraphDrawRoundedStroke(gtx, request.Rect, request.Radius, 1, request.Border)
	for index, label := range []string{"-", "+"} {
		button := image.Rect(request.Rect.Min.X+index*request.ButtonWidth, request.Rect.Min.Y, request.Rect.Min.X+(index+1)*request.ButtonWidth, request.Rect.Max.Y)
		if label != "" {
			nodeGraphDrawCenteredText(materialTheme, gtx, label, request.TextSize, font.Normal, request.Foreground, button)
		}
	}
	fit := image.Rect(request.Rect.Min.X+2*request.ButtonWidth, request.Rect.Min.Y, request.Rect.Max.X, request.Rect.Max.Y)
	if !fit.Empty() {
		nodeGraphDrawFitIcon(gtx, fit, request.Foreground, request.FitSpan, request.FitArm)
	}
	for index := 1; index < 3; index++ {
		x := request.Rect.Min.X + index*request.ButtonWidth
		paint.FillShape(gtx.Ops, request.Border, clip.Rect(image.Rect(x, request.Rect.Min.Y+request.DividerInset, x+1, request.Rect.Max.Y-request.DividerInset)).Op())
	}
}

func drawNodeGraphMinimap(gtx layout.Context, request nodeGraphMinimap) {
	if request.Rect.Empty() || len(request.Nodes) == 0 {
		return
	}
	paint.FillShape(gtx.Ops, request.Background, clip.UniformRRect(request.Rect, request.Radius).Op(gtx.Ops))
	nodeGraphDrawRoundedStroke(gtx, request.Rect, request.Radius, request.BorderWidth, request.Border)
	clipStack := clip.Rect(request.Rect.Inset(max(request.BorderWidth, 1))).Push(gtx.Ops)
	for _, edge := range request.Edges {
		nodeGraphDrawPolyline(gtx, edge.Points, request.EdgeWidth, edge.Color)
	}
	for _, node := range request.Nodes {
		fill := request.Border
		if node.Selected {
			fill = request.Selected
		}
		if node.Rect.Empty() {
			continue
		}
		paint.FillShape(gtx.Ops, fill, clip.Rect(node.Rect).Op())
	}
	if !request.Viewport.Empty() {
		nodeGraphDrawRoundedStroke(gtx, request.Viewport, 1, int(math.Ceil(float64(request.ViewportWidth))), request.Selected)
	}
	clipStack.Pop()
}

func nodeGraphDrawBezier(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA, minimumDistance float32) {
	distance := nodeGraphMaxFloat(nodeGraphAbs32(to.X-from.X)*.5, minimumDistance)
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.CubeTo(f32.Pt(from.X+distance, from.Y), f32.Pt(to.X-distance, to.Y), to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func nodeGraphDrawPolyline(gtx layout.Context, points []f32.Point, width float32, col color.NRGBA) {
	if len(points) < 2 || width <= 0 || col.A == 0 {
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(points[0])
	for _, point := range points[1:] {
		path.LineTo(point)
	}
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func nodeGraphDrawDashedPolyline(gtx layout.Context, points []f32.Point, width float32, col color.NRGBA, phase float32) {
	dash, gap := nodeGraphMaxFloat(width*4, 8), nodeGraphMaxFloat(width*2, 5)
	cycle := dash + gap
	offset := float32(math.Mod(float64(phase), float64(cycle)))
	for index := 1; index < len(points); index++ {
		from, to := points[index-1], points[index]
		delta := to.Sub(from)
		length := float32(math.Hypot(float64(delta.X), float64(delta.Y)))
		if length <= 0 {
			continue
		}
		progress := -offset
		for progress < length {
			start := nodeGraphMaxFloat(progress, 0)
			end := nodeGraphMinFloat(progress+dash, length)
			if end > start {
				a := f32.Pt(from.X+delta.X*start/length, from.Y+delta.Y*start/length)
				b := f32.Pt(from.X+delta.X*end/length, from.Y+delta.Y*end/length)
				nodeGraphDrawPolyline(gtx, []f32.Point{a, b}, width, col)
			}
			progress += cycle
		}
		offset = float32(math.Mod(float64(offset+length), float64(cycle)))
	}
}

func nodeGraphDrawSmoothStep(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	deltaX, deltaY := to.X-from.X, to.Y-from.Y
	if nodeGraphAbs32(deltaY) < 1 {
		nodeGraphDrawPolyline(gtx, []f32.Point{from, to}, width, col)
		return
	}
	midX := (from.X + to.X) / 2
	directionX, directionY := float32(1), float32(1)
	if deltaX < 0 {
		directionX = -1
	}
	if deltaY < 0 {
		directionY = -1
	}
	radius := nodeGraphMinFloat(8, nodeGraphMinFloat(nodeGraphAbs32(midX-from.X)/2, nodeGraphAbs32(deltaY)/2))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(f32.Pt(midX-directionX*radius, from.Y))
	path.QuadTo(f32.Pt(midX, from.Y), f32.Pt(midX, from.Y+directionY*radius))
	path.LineTo(f32.Pt(midX, to.Y-directionY*radius))
	path.QuadTo(f32.Pt(midX, to.Y), f32.Pt(midX+directionX*radius, to.Y))
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func nodeGraphDrawMarker(gtx layout.Context, marker NodeGraphEdgeMarker, from, to f32.Point, source bool, col color.NRGBA, size float32, points []f32.Point) {
	if marker != NodeGraphMarkerArrow || len(points) < 2 {
		return
	}
	endpoint, neighbor := to, points[len(points)-2]
	if source {
		endpoint, neighbor = from, points[1]
	}
	direction := endpoint.Sub(neighbor)
	length := float32(math.Hypot(float64(direction.X), float64(direction.Y)))
	if length <= 0 {
		return
	}
	direction = direction.Mul(1 / length)
	perpendicular := f32.Pt(-direction.Y, direction.X)
	size = nodeGraphMaxFloat(size, 5)
	base := endpoint.Sub(direction.Mul(size))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(endpoint)
	path.LineTo(base.Add(perpendicular.Mul(size * .55)))
	path.LineTo(base.Sub(perpendicular.Mul(size * .55)))
	path.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func nodeGraphDrawText(materialTheme *material.Theme, gtx layout.Context, value string, size unit.Sp, weight font.Weight, col color.NRGBA, rect image.Rectangle) {
	if materialTheme == nil || value == "" || rect.Empty() {
		return
	}
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: rect.Size()}
	label := material.Label(materialTheme, size, value)
	label.Color = col
	label.Font.Weight = weight
	label.MaxLines = 1
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	label.Layout(textGtx)
	offset.Pop()
}

func nodeGraphDrawCenteredText(materialTheme *material.Theme, gtx layout.Context, value string, size unit.Sp, weight font.Weight, col color.NRGBA, rect image.Rectangle) {
	if materialTheme == nil || value == "" || rect.Empty() {
		return
	}
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: rect.Size()}
	macro := op.Record(gtx.Ops)
	label := material.Label(materialTheme, size, value)
	label.Color = col
	label.Font.Weight = weight
	label.MaxLines = 1
	dimensions := label.Layout(textGtx)
	call := macro.Stop()
	position := image.Pt(rect.Min.X+max((rect.Dx()-dimensions.Size.X)/2, 0), rect.Min.Y+max((rect.Dy()-dimensions.Size.Y)/2, 0))
	offset := op.Offset(position).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
}

func nodeGraphDrawFitIcon(gtx layout.Context, rect image.Rectangle, col color.NRGBA, span, arm float32) {
	center := f32.Pt(float32(rect.Min.X+rect.Max.X)/2, float32(rect.Min.Y+rect.Max.Y)/2)
	span *= float32(min(rect.Dx(), rect.Dy()))
	arm *= span
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(center.Add(f32.Pt(-span, -arm)))
	path.LineTo(center.Add(f32.Pt(-span, -span)))
	path.LineTo(center.Add(f32.Pt(-arm, -span)))
	path.MoveTo(center.Add(f32.Pt(span, -arm)))
	path.LineTo(center.Add(f32.Pt(span, -span)))
	path.LineTo(center.Add(f32.Pt(arm, -span)))
	path.MoveTo(center.Add(f32.Pt(-span, arm)))
	path.LineTo(center.Add(f32.Pt(-span, span)))
	path.LineTo(center.Add(f32.Pt(-arm, span)))
	path.MoveTo(center.Add(f32.Pt(span, arm)))
	path.LineTo(center.Add(f32.Pt(span, span)))
	path.LineTo(center.Add(f32.Pt(arm, span)))
	stroke := clip.Stroke{Path: path.End(), Width: 1}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func nodeGraphDrawDottedRect(gtx layout.Context, rect image.Rectangle, col color.NRGBA, width, segment, gap int) {
	if rect.Empty() || col.A == 0 || width <= 0 {
		return
	}
	segment, gap = max(segment, 1), max(gap, 1)
	for x := rect.Min.X; x < rect.Max.X; x += segment + gap {
		end := min(x+segment, rect.Max.X)
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, rect.Min.Y, end, min(rect.Min.Y+width, rect.Max.Y))).Op())
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, max(rect.Max.Y-width, rect.Min.Y), end, rect.Max.Y)).Op())
	}
	for y := rect.Min.Y + segment; y < rect.Max.Y-segment; y += segment + gap {
		end := min(y+segment, rect.Max.Y)
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(rect.Min.X, y, min(rect.Min.X+width, rect.Max.X), end)).Op())
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(max(rect.Max.X-width, rect.Min.X), y, rect.Max.X, end)).Op())
	}
}

func nodeGraphDrawRoundedStroke(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func nodeGraphResizeHandlePoint(rect image.Rectangle, bit uint16) image.Point {
	midX, midY := (rect.Min.X+rect.Max.X)/2, (rect.Min.Y+rect.Max.Y)/2
	switch bit {
	case 1:
		return image.Pt(midX, rect.Min.Y)
	case 2:
		return image.Pt(rect.Max.X, midY)
	case 4:
		return image.Pt(midX, rect.Max.Y)
	case 8:
		return image.Pt(rect.Min.X, midY)
	case 16:
		return rect.Min
	case 32:
		return image.Pt(rect.Max.X, rect.Min.Y)
	case 64:
		return rect.Max
	case 128:
		return image.Pt(rect.Min.X, rect.Max.Y)
	default:
		return rect.Min
	}
}

func nodeGraphWithAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A)*factor + .5)
	return value
}

func nodeGraphMaxFloat(first, second float32) float32 {
	if first > second {
		return first
	}
	return second
}

func nodeGraphMinFloat(first, second float32) float32 {
	if first < second {
		return first
	}
	return second
}

func nodeGraphAbs32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}
