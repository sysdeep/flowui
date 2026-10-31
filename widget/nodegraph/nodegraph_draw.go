package nodegraph

import (
	"image"
	"image/color"
	"math"
	"slices"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
)

func drawNodeGraph(ctx *core.Context, gtx layout.Context, graph resolvedGraph, viewport Viewport, selected, selectedEdges map[string]bool, hoveredNode, focusedNode, graphKey string, showGrid bool, gridPattern GridPattern, gridSize float32, nodesResizable bool, resizeHandles ResizeHandle, minimap, controls, culling bool, panels []Panel, viewportOverlays []ViewportOverlay, nodeToolbars []nodeToolbar, edgeToolbars []edgeToolbar, minimapScratch *nodeGraphMinimapScratch, defaults nodeGraphDefaults) {
	size := gtx.Constraints.Max
	canvas := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(defaults.metrics.CanvasRadius), 0), min(size.X, size.Y)/2)
	drawNodeGraphCanvas(gtx, nodeGraphCanvas{
		Size: size, Radius: radius, BorderWidth: max(gtx.Dp(defaults.metrics.CanvasBorderWidth), 0),
		Background: defaults.colors.CanvasBackground, Border: defaults.colors.CanvasBorder,
	})
	root := clip.UniformRRect(canvas, radius).Push(gtx.Ops)
	semantic.DescriptionOp(ctx.Text(i18n.NodeGraphDescription)).Add(gtx.Ops)
	if showGrid {
		drawGrid(gtx, viewport, gridSize, defaults, gridPattern)
	}
	edgeIndices := graphDrawEdgeIndices(graph, viewport, gtx, culling)
	for _, edgeIndice := range slices.Backward(edgeIndices) {
		edge := graph.edges[edgeIndice]
		if !culling || edgeVisible(gtx, viewport, edge) {
			drawEdge(ctx, gtx, viewport, edge, selectedEdges[edge.edge.ID], graphKey, defaults)
		}
		if edge.edge.isAnimated() && ctx != nil {
			ctx.Invalidate()
		}
	}
	nodeIndices := graphDrawNodeIndices(graph, viewport, gtx, culling)
	for _, nodeIndice := range slices.Backward(nodeIndices) {
		node := graph.nodes[nodeIndice]
		drawNode(ctx, gtx, viewport, node, selected[node.node.ID], node.node.ID == hoveredNode, node.node.ID == focusedNode, node.node.isResizable(nodesResizable), node.node.resizeHandleSet(resizeHandles), graphKey, defaults, culling)
	}
	drawGraphViewportOverlays(ctx, gtx, viewport, graphKey, viewportOverlays)
	drawGraphToolbars(ctx, gtx, graph, viewport, graphKey, nodeToolbars, edgeToolbars, defaults)
	drawGraphPanels(ctx, gtx, graphKey, panels, defaults)
	drawNodeMinimap(gtx, graph, viewport, selected, defaults, minimap, minimapScratch)
	drawNodeControls(ctx, gtx, defaults, controls)
	root.Pop()
}

func drawGraphToolbars(ctx *core.Context, gtx layout.Context, graph resolvedGraph, viewport Viewport, graphKey string, nodeToolbars []nodeToolbar, edgeToolbars []edgeToolbar, defaults nodeGraphDefaults) {
	overlays := make([]ViewportOverlay, 0, len(nodeToolbars)+len(edgeToolbars))
	for _, toolbar := range nodeToolbars {
		if node, ok := graph.byID[toolbar.nodeID]; ok {
			offset := float32(defaults.metrics.ToolbarOffset)
			overlays = append(overlays, ViewportOverlay{Key: "node-toolbar:" + toolbar.key, Position: Point{X: node.position.X + node.size.Width/2, Y: node.position.Y - offset}, Content: toolbar.content})
		}
	}
	for _, toolbar := range edgeToolbars {
		if edge, ok := graph.edgeByID[toolbar.edgeID]; ok {
			mid := graphEdgeMidpoint(edge.edge.edgeTypeValue(), f32.Pt(edge.source.X, edge.source.Y), f32.Pt(edge.target.X, edge.target.Y), float32(defaults.metrics.EdgeMinimumDistance))
			overlays = append(overlays, ViewportOverlay{Key: "edge-toolbar:" + toolbar.key, Position: Point{X: mid.X, Y: mid.Y}, Content: toolbar.content})
		}
	}
	drawGraphViewportOverlays(ctx, gtx, viewport, graphKey, overlays)
}

func graphDrawNodeIndices(graph resolvedGraph, viewport Viewport, gtx layout.Context, culling bool) []int {
	if !culling {
		indices := make([]int, len(graph.nodes))
		for index := range indices {
			indices[index] = index
		}
		return indices
	}
	return graph.spatial.nodeCandidates(graphViewportBounds(viewport, gtx, defaultNodeGraphBehavior.nodeCullingPadding))
}

func graphDrawEdgeIndices(graph resolvedGraph, viewport Viewport, gtx layout.Context, culling bool) []int {
	if !culling {
		indices := make([]int, len(graph.edges))
		for index := range indices {
			indices[index] = index
		}
		return indices
	}
	return graph.spatial.edgeCandidates(graphViewportBounds(viewport, gtx, defaultNodeGraphBehavior.edgeCullingPadding))
}

func graphViewportBounds(viewport Viewport, gtx layout.Context, padding float32) graphRect {
	pixelsPerDP := graphPixelsPerDP(gtx)
	first := screenToWorld(viewport, f32.Point{}, pixelsPerDP)
	last := screenToWorld(viewport, f32.Pt(float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y)), pixelsPerDP)
	return graphRect{min: Point{X: first.X - padding, Y: first.Y - padding}, max: Point{X: last.X + padding, Y: last.Y + padding}}
}

func drawGraphViewportOverlays(ctx *core.Context, gtx layout.Context, viewport Viewport, graphKey string, overlays []ViewportOverlay) {
	if ctx == nil {
		return
	}
	pixelsPerDP := graphPixelsPerDP(gtx)
	for _, overlay := range overlays {
		screen := worldToScreen(viewport, overlay.Position, pixelsPerDP).Round()
		if screen.X < -gtx.Constraints.Max.X || screen.Y < -gtx.Constraints.Max.Y || screen.X > gtx.Constraints.Max.X*2 || screen.Y > gtx.Constraints.Max.Y*2 {
			continue
		}
		overlayGtx := gtx
		overlayGtx.Constraints = layout.Constraints{Max: gtx.Constraints.Max}
		overlayGtx.Metric.PxPerDp = pixelsPerDP * viewport.Zoom
		overlayGtx.Metric.PxPerSp = max(gtx.Metric.PxPerSp, 1) * viewport.Zoom
		macro := op.Record(gtx.Ops)
		restoreGraphKey := core.PushKey(ctx, graphKey)
		restoreOverlayKey := core.PushKey(ctx, "viewport-overlay:"+overlay.Key)
		_, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return overlay.Content.Layout(ctx, overlayGtx)
		})
		placement.PlaceOffset(screen)
		restoreOverlayKey()
		restoreGraphKey()
		call := macro.Stop()
		offset := op.Offset(screen).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
}

func drawGraphPanels(ctx *core.Context, gtx layout.Context, graphKey string, panels []Panel, defaults nodeGraphDefaults) {
	if ctx == nil {
		return
	}
	padding := gtx.Dp(defaults.metrics.PanelPadding)
	for _, panel := range panels {
		panelGtx := gtx
		panelGtx.Constraints = layout.Constraints{Max: gtx.Constraints.Max}
		macro := op.Record(gtx.Ops)
		restoreGraphKey := core.PushKey(ctx, graphKey)
		restorePanelKey := core.PushKey(ctx, "panel:"+panel.Key)
		dimensions, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return panel.Content.Layout(ctx, panelGtx)
		})
		restorePanelKey()
		restoreGraphKey()
		call := macro.Stop()
		position := graphPanelPosition(panel.Position, dimensions.Size, gtx.Constraints.Max, padding)
		placement.PlaceOffset(position)
		offset := op.Offset(position).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
}

func graphPanelPosition(position PanelPosition, size, canvas image.Point, padding int) image.Point {
	x, y := padding, padding
	switch position {
	case PanelTopCenter, PanelBottomCenter:
		x = max((canvas.X-size.X)/2, padding)
	case PanelTopRight, PanelBottomRight:
		x = max(canvas.X-size.X-padding, padding)
	}
	switch position {
	case PanelBottomLeft, PanelBottomCenter, PanelBottomRight:
		y = max(canvas.Y-size.Y-padding, padding)
	}
	return image.Pt(x, y)
}

func edgeVisible(gtx layout.Context, viewport Viewport, edge resolvedEdge) bool {
	pixelsPerDP := graphPixelsPerDP(gtx)
	from := worldToScreen(viewport, edge.source, pixelsPerDP)
	to := worldToScreen(viewport, edge.target, pixelsPerDP)
	padding := max(defaultNodeGraphBehavior.edgeCullingMinPaddingPx, abs32(to.X-from.X)*defaultNodeGraphBehavior.edgeCullingOvershootRatio)
	minX, maxX := min(from.X, to.X)-padding, max(from.X, to.X)+padding
	minY, maxY := min(from.Y, to.Y)-padding, max(from.Y, to.Y)+padding
	canvas := gtx.Constraints.Max
	return maxX >= 0 && maxY >= 0 && minX <= float32(canvas.X) && minY <= float32(canvas.Y)
}

func drawGrid(gtx layout.Context, viewport Viewport, gridSize float32, defaults nodeGraphDefaults, pattern GridPattern) {
	if gridSize <= 0 || !finite(gridSize) {
		return
	}
	scale := worldScale(viewport, graphPixelsPerDP(gtx))
	step := graphGridStep(gridSize, scale, float32(defaults.metrics.GridMinimumStep))
	if step <= 0 {
		return
	}
	startX := -float32(math.Mod(float64(viewport.Origin.X*scale), float64(step)))
	startY := -float32(math.Mod(float64(viewport.Origin.Y*scale), float64(step)))
	if startX > 0 {
		startX -= step
	}
	if startY > 0 {
		startY -= step
	}
	drawNodeGraphGrid(gtx, nodeGraphGrid{
		Size: gtx.Constraints.Max, Pattern: nodeGraphGridPattern(pattern),
		StartX: startX, StartY: startY, Step: step,
		Width: float32(max(gtx.Dp(defaults.metrics.GridWidth), 1)), Color: defaults.colors.Grid,
	})
}

// graphGridStep skips overly dense grid intervals at low zoom while keeping
// the visible grid aligned to an integer multiple of the configured spacing.
func graphGridStep(gridSize, scale float32, minimum ...float32) float32 {
	step := gridSize * scale
	if step <= 0 || !finite(step) {
		return 0
	}
	minimumPixelStep := float32(8)
	if len(minimum) > 0 && minimum[0] > 0 {
		minimumPixelStep = minimum[0]
	}
	if step < minimumPixelStep {
		step *= float32(math.Ceil(float64(minimumPixelStep / step)))
	}
	return step
}

func drawEdge(ctx *core.Context, gtx layout.Context, viewport Viewport, edge resolvedEdge, selected bool, graphKey string, defaults nodeGraphDefaults) {
	pixelsPerDP := graphPixelsPerDP(gtx)
	from := worldToScreen(viewport, edge.source, pixelsPerDP)
	to := worldToScreen(viewport, edge.target, pixelsPerDP)
	colorValue := edge.edge.Color
	if colorValue.A == 0 {
		colorValue = defaults.colors.EdgeColor
	}
	if selected {
		colorValue = defaults.colors.SelectedEdgeColor
	}
	width := edge.edge.width
	if width <= 0 {
		width = float32(defaults.metrics.EdgeWidth)
	}
	width = max(float32(gtx.Dp(unit.Dp(width)))*viewport.Zoom, 1)
	minimumDistance := float32(defaults.metrics.EdgeMinimumDistance) * worldScale(viewport, pixelsPerDP)
	animated := edge.edge.isAnimated()
	phase := float32(0)
	if animated {
		phase = float32(gtx.Now.UnixNano()%1_000_000_000) / 1_000_000_000 * max(width*8, 16)
	}
	points := graphEdgePolyline(edge.edge.edgeTypeValue(), from, to, minimumDistance)
	if edge.edge.label != "" || edge.edge.labelContent != nil {
		midpoint := graphEdgeMidpoint(edge.edge.edgeTypeValue(), from, to, minimumDistance)
		labelWidth := max(int(math.Round(float64(float32(defaults.metrics.EdgeLabelWidth)*viewport.Zoom))), 1)
		labelHeight := max(int(math.Round(float64(float32(defaults.metrics.EdgeLabelHeight)*viewport.Zoom))), 1)
		labelRect := image.Rect(int(math.Round(float64(midpoint.X)))-labelWidth, int(math.Round(float64(midpoint.Y)))-labelHeight, int(math.Round(float64(midpoint.X)))+labelWidth, int(math.Round(float64(midpoint.Y)))+labelHeight)
		if edge.edge.labelContent == nil {
			drawNodeGraphEdge(defaults.material, gtx, nodeGraphEdge{
				EdgeType: nodeGraphEdgeType(edge.edge.edgeTypeValue()), From: from, To: to, Points: points,
				Width: width, MinimumDistance: minimumDistance, Dashed: edge.edge.isDashed() || animated,
				Phase: phase, Color: colorValue, SourceMarker: nodeGraphMarker(edge.edge.sourceMark),
				TargetMarker: nodeGraphMarker(edge.edge.targetMark), MarkerSize: float32(gtx.Dp(defaults.metrics.EdgeMarkerSize)),
				Label: edge.edge.label, LabelRect: labelRect,
				LabelSize:  unit.Sp(min(max(float32(defaults.metrics.NodePortTextSize)*viewport.Zoom, float32(defaults.metrics.NodePortTextSize)*defaults.metrics.NodeTextMinScale), float32(defaults.metrics.NodePortTextSize)*defaults.metrics.NodeTextMaxScale)),
				LabelColor: defaults.colors.NodeForeground, LabelWeight: font.Normal,
			})
			return
		}
		visible := labelRect.Intersect(image.Rectangle{Max: gtx.Constraints.Max})
		if visible.Empty() {
			return
		}
		restoreGraphKey := core.PushKey(ctx, graphKey)
		restoreEdgeKey := core.PushKey(ctx, "edge:"+edge.edge.ID)
		contentGtx := gtx
		contentGtx.Constraints = layout.Exact(labelRect.Size())
		offset := op.Offset(labelRect.Min).Push(gtx.Ops)
		edge.edge.labelContent.Layout(ctx, contentGtx)
		offset.Pop()
		restoreEdgeKey()
		restoreGraphKey()
	}
	drawNodeGraphEdge(defaults.material, gtx, nodeGraphEdge{
		EdgeType: nodeGraphEdgeType(edge.edge.edgeTypeValue()), From: from, To: to, Points: points,
		Width: width, MinimumDistance: minimumDistance, Dashed: edge.edge.isDashed() || animated,
		Phase: phase, Color: colorValue, SourceMarker: nodeGraphMarker(edge.edge.sourceMark),
		TargetMarker: nodeGraphMarker(edge.edge.targetMark), MarkerSize: float32(gtx.Dp(defaults.metrics.EdgeMarkerSize)),
	})
}

func drawConnectionPreview(gtx layout.Context, viewport Viewport, state *graphState, defaults nodeGraphDefaults) {
	from := worldToScreen(viewport, state.connectionSource.point, graphPixelsPerDP(gtx))
	to := state.connectionCurrent
	colorValue := defaults.colors.EdgeColor
	if state.connectionTarget.endpoint.NodeID != "" {
		to = worldToScreen(viewport, state.connectionTarget.point, graphPixelsPerDP(gtx))
	}
	if state.connectionMode == connectionReconnectSource {
		from, to = to, from
	}
	drawNodeGraphConnection(gtx, nodeGraphConnection{
		From: from, To: to, Width: max(float32(gtx.Dp(defaults.metrics.EdgeWidth))*viewport.Zoom, 1),
		Color: colorValue, MinimumDistance: float32(defaults.metrics.EdgeMinimumDistance) * worldScale(viewport, graphPixelsPerDP(gtx)),
	})
}

func graphEdgeMidpoint(edgeType EdgeType, from, to f32.Point, minimumDistance float32) f32.Point {
	if edgeType == EdgeBezier {
		return graphBezierPoint(from, to, minimumDistance, .5)
	}
	return f32.Pt((from.X+to.X)/2, (from.Y+to.Y)/2)
}

func drawSelectionBox(gtx layout.Context, first, second f32.Point, defaults nodeGraphDefaults) {
	rect := image.Rect(
		int(math.Floor(float64(min(first.X, second.X)))),
		int(math.Floor(float64(min(first.Y, second.Y)))),
		int(math.Ceil(float64(max(first.X, second.X)))),
		int(math.Ceil(float64(max(first.Y, second.Y)))),
	)
	if rect.Empty() {
		return
	}
	drawNodeGraphSelection(gtx, nodeGraphSelection{
		Rect: rect, Fill: defaults.colors.SelectionFill, Border: defaults.colors.SelectionBorder,
		Width:      max(gtx.Dp(defaults.metrics.SelectionWidth), 1),
		DashLength: max(gtx.Dp(defaults.metrics.SelectionDashLength), 1),
		DashGap:    max(gtx.Dp(defaults.metrics.SelectionDashGap), 1), Filled: true,
	})
}

func drawNode(ctx *core.Context, gtx layout.Context, viewport Viewport, node resolvedNode, selected, hovered, focused, resizable bool, resizeHandles ResizeHandle, graphKey string, defaults nodeGraphDefaults, culling bool) {
	pixelsPerDP := graphPixelsPerDP(gtx)
	position := worldToScreen(viewport, node.position, pixelsPerDP).Round()
	size := image.Pt(
		max(int(math.Round(float64(node.size.Width*worldScale(viewport, pixelsPerDP)))), 1),
		max(int(math.Round(float64(node.size.Height*worldScale(viewport, pixelsPerDP)))), 1),
	)
	rect := image.Rectangle{Min: position, Max: position.Add(size)}
	if culling && rect.Intersect(image.Rectangle{Max: gtx.Constraints.Max}).Empty() {
		return
	}
	radius := min(max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.NodeRadius))*viewport.Zoom))), 1), min(size.X, size.Y)/2)
	headerHeight := min(max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.NodeHeaderHeight))*viewport.Zoom))), 1), size.Y)
	border := defaults.colors.NodeBorder
	if selected {
		border = defaults.colors.SelectedNodeBorder
	}
	borderWidth := max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.NodeBorderWidth))*viewport.Zoom))), 1)
	padding := max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.NodePadding))*viewport.Zoom))), 4)
	title := node.node.Title
	if title == "" {
		title = node.node.ID
	}
	titleSize := unit.Sp(float32(defaults.metrics.NodeTitleSize) * viewport.Zoom)
	titleSize = unit.Sp(min(max(float32(titleSize), float32(defaults.metrics.NodeTitleSize)*defaults.metrics.NodeTextMinScale), float32(defaults.metrics.NodeTitleSize)*defaults.metrics.NodeTextMaxScale))
	drawNodeGraphNode(defaults.material, gtx, defaults.metrics, nodeGraphNode{
		Rect: rect, Radius: radius, BorderWidth: borderWidth,
		Background: defaults.colors.NodeBackground, Border: border, Shadow: defaults.colors.Shadow,
		DrawShadow: hovered && !selected, Title: title,
		TitleRect: image.Rect(rect.Min.X+padding, rect.Min.Y, rect.Max.X-padding, rect.Min.Y+headerHeight),
		TitleSize: titleSize, TitleColor: defaults.colors.NodeForeground, TitleWeight: font.Normal,
	})

	clipped := clip.Rect(rect).Push(gtx.Ops)
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(node.node.Title).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	if focused {
		semantic.DescriptionOp(ctx.Text(i18n.NodeGraphFocused)).Add(gtx.Ops)
	}
	if node.node.content != nil {
		inner := rect.Inset(padding)
		if !inner.Empty() {
			restoreGraphKey := core.PushKey(ctx, graphKey)
			restoreNodeKey := core.PushKey(ctx, "node:"+node.node.ID)
			contentGtx := gtx
			contentGtx.Constraints = layout.Exact(inner.Size())
			contentScale := max(viewport.Zoom, 0.01)
			contentGtx.Metric.PxPerDp = graphPixelsPerDP(gtx) * contentScale
			contentGtx.Metric.PxPerSp = max(gtx.Metric.PxPerSp, 1) * contentScale
			offset := op.Offset(inner.Min).Push(gtx.Ops)
			node.node.content.Layout(ctx, contentGtx)
			offset.Pop()
			restoreNodeKey()
			restoreGraphKey()
		}
	}
	clipped.Pop()

	// Ports and resize handles intentionally cross the node boundary. The node
	// content above remains clipped, while these interaction targets do not.
	for _, resolvedPort := range node.inputs {
		drawNodePort(ctx, gtx, viewport, resolvedPort, rect, size, padding, defaults)
	}
	for _, resolvedPort := range node.outputs {
		drawNodePort(ctx, gtx, viewport, resolvedPort, rect, size, padding, defaults)
	}
	if focused {
		inset := max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.FocusInset))*viewport.Zoom))), 1)
		drawNodeGraphSelection(gtx, nodeGraphSelection{
			Rect: rect.Inset(inset), Border: defaults.colors.SelectionBorder,
			Width:      max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.FocusWidth))*viewport.Zoom))), 1),
			DashLength: max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.FocusDashLength))*viewport.Zoom))), 1),
			DashGap:    max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.FocusDashGap))*viewport.Zoom))), 1),
		})
	}
	if selected && resizable && resizeHandles != 0 {
		drawNodeGraphResize(gtx, nodeGraphResize{
			Rect: rect, Handles: uint16(resizeHandles),
			Radius:      max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.ResizeHandleRadius))*viewport.Zoom))), 3),
			BorderWidth: max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.ResizeHandleBorderWidth))*viewport.Zoom))), 1),
			Border:      defaults.colors.SelectedNodeBorder, Fill: defaults.colors.ResizeHandleFill,
		})
	}
}

func drawNodePort(ctx *core.Context, gtx layout.Context, viewport Viewport, resolvedPort resolvedPort, rect image.Rectangle, size image.Point, padding int, defaults nodeGraphDefaults) {
	port := resolvedPort.port
	anchor := worldToScreen(viewport, resolvedPort.point, graphPixelsPerDP(gtx))
	label := graphPortLabel(port)
	textSize := unit.Sp(float32(defaults.metrics.NodePortTextSize) * viewport.Zoom)
	textSize = unit.Sp(min(max(float32(textSize), float32(defaults.metrics.NodePortTextSize)*defaults.metrics.NodeTextMinScale), float32(defaults.metrics.NodePortTextSize)*defaults.metrics.NodeTextMaxScale))
	maxWidth := max(size.X/2-padding, 1)
	// Measure the font's full line box; the nominal text size can clip glyphs.
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(maxWidth, math.MaxInt32)}
	labelStyle := material.Label(defaults.material, textSize, label)
	labelStyle.MaxLines = 1
	measurement := op.Record(gtx.Ops)
	textHeight := labelStyle.Layout(textGtx).Size.Y
	measurement.Stop()
	position := image.Pt(rect.Min.X+padding, int(math.Round(float64(anchor.Y)))-textHeight/2)
	switch resolvedPort.position {
	case HandleRight:
		position = image.Pt(rect.Min.X+size.X/2, int(math.Round(float64(anchor.Y)))-textHeight/2)
	case HandleTop:
		position = image.Pt(int(math.Round(float64(anchor.X)))-maxWidth/2, rect.Min.Y+padding)
	case HandleBottom:
		position = image.Pt(int(math.Round(float64(anchor.X)))-maxWidth/2, rect.Max.Y-padding-textHeight)
	}
	labelRect := image.Rect(position.X, position.Y, position.X+maxWidth, position.Y+textHeight)
	if strings.TrimSpace(label) == "" {
		labelRect = image.Rectangle{}
	}
	drawNodeGraphPort(defaults.material, gtx, nodeGraphPort{
		Center:      anchor,
		Radius:      max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.PortRadius))*viewport.Zoom))), 2),
		BorderWidth: max(int(math.Round(float64(float32(gtx.Dp(defaults.metrics.PortBorderWidth))*viewport.Zoom))), 1),
		Color:       portColor(port, defaults.colors.PortColor), Border: defaults.colors.PortBorder,
		Label: label, LabelRect: labelRect, LabelSize: textSize,
		LabelColor: defaults.colors.NodeMutedForeground, LabelWeight: font.Normal,
	})
}

func graphPortLabel(port Port) string {
	if port.Label != "" {
		return port.Label
	}
	return port.ID
}

func portColor(port Port, fallback color.NRGBA) color.NRGBA {
	if port.Color.A != 0 {
		return port.Color
	}
	return fallback
}

func nodeGraphGridPattern(pattern GridPattern) NodeGraphGridPattern {
	if pattern == GridDots {
		return NodeGraphGridDots
	}
	return NodeGraphGridLines
}

func nodeGraphEdgeType(value EdgeType) NodeGraphEdgeType {
	switch value {
	case EdgeBezier:
		return NodeGraphEdgeBezier
	case EdgeStep:
		return NodeGraphEdgeStep
	case EdgeSmoothStep:
		return NodeGraphEdgeSmoothStep
	default:
		return NodeGraphEdgeStraight
	}
}

func nodeGraphMarker(value EdgeMarker) NodeGraphEdgeMarker {
	if value == MarkerArrow {
		return NodeGraphMarkerArrow
	}
	return NodeGraphMarkerNone
}
