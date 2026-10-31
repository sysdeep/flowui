package nodegraph

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if w.visualCache == nil {
		w.visualCache = &nodeGraphDefaultsCache{}
	}
	active := core.ActiveTheme(ctx)
	defaults := w.resolveDefaults(active)
	resolved := computed.ResolveStatic(ctx, style.State{Disabled: w.disabled}, defaults.root, style.Style{}, style.Style{}, w.style)
	host := &w.visualCache.layoutHost
	host.value = w
	host.defaults = defaults
	return layoutui.Resolved(ctx, gtx, resolved, host)
}

type nodeGraphLayoutHost struct {
	value    Widget
	defaults nodeGraphDefaults
}

func (h *nodeGraphLayoutHost) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return h.value.layout(ctx, gtx, h.defaults)
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, defaults nodeGraphDefaults) layout.Dimensions {
	if w.disabled {
		gtx = gtx.Disabled()
	}
	key := core.ClaimKey(ctx, core.KindNodeGraph, w.key)
	value := core.UseState[graphState](ctx, key, stateSlotNodeGraph)
	resolved := resolveGraphMeasuredWithMetrics(ctx, gtx, w.graph, defaults.metrics, w.key)
	size := gtx.Constraints.Max
	if w.height > 0 {
		size.Y = gtx.Dp(w.height)
	}
	size = gtx.Constraints.Constrain(size)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	viewport, fitChanged := w.resolveViewport(value, resolved, size, graphPixelsPerDP(gtx))
	if fitChanged && w.onViewportChange != nil {
		w.onViewportChange(viewport)
	}
	enabled := gtx.Enabled() && !w.disabled
	selected := w.selectedNodeSet(resolved)
	selectedEdges := w.selectedEdgeSet(resolved)
	if next, changed := w.updateGestures(ctx, gtx, value, resolved, viewport, selected, selectedEdges, size, enabled, defaults.metrics); changed {
		viewport = next
		value.viewport = next
		value.ready = true
		if w.onViewportChange != nil {
			w.onViewportChange(next)
		}
	}

	graphGtx := gtx
	graphGtx.Constraints = layout.Exact(size)
	w.updateDropEvents(graphGtx, value, viewport, graphPixelsPerDP(gtx))
	renderSelected := selected
	if value.selectionCurrentSet != nil {
		renderSelected = value.selectionCurrentSet
	}
	w.updateKeyboard(graphGtx, value, resolved, selected, selectedEdges, enabled)
	if next, changed := ensureFocusedNodeVisible(resolved, value.focusedNode, viewport, size, graphPixelsPerDP(gtx)); changed {
		viewport = normalizeViewport(next, w.minZoom, w.maxZoom)
		value.viewport = viewport
		value.ready = true
		if w.onViewportChange != nil {
			w.onViewportChange(viewport)
		}
	}
	// Register the canvas target before node content so child widgets are the
	// topmost pointer hit inside their own layout bounds.
	w.addInput(graphGtx, value, resolved, enabled)
	focusedNode := ""
	if value.focusVisible {
		focusedNode = value.focusedNode
	}
	drawNodeGraph(ctx, graphGtx, resolved, viewport, renderSelected, selectedEdges, value.hoveredNode, focusedNode, w.key, w.showGrid, w.gridPattern, w.gridSize, w.nodesResizable, w.resizeHandles, w.minimapEnabled, w.controlsEnabled, w.cullingEnabled, w.panels, w.viewportOverlays, w.nodeToolbars, w.edgeToolbars, &w.visualCache.minimapScratch, defaults)
	if value.gesture == gestureSelectionBox && value.dragStarted {
		drawSelectionBox(graphGtx, value.selectionStart, value.selectionCurrent, defaults)
	}
	if value.gesture == gestureConnect {
		drawConnectionPreview(graphGtx, viewport, value, defaults)
	}
	return layout.Dimensions{Size: size}
}

func ensureFocusedNodeVisible(graph resolvedGraph, focused string, viewport Viewport, size image.Point, pixelsPerDP float32) (Viewport, bool) {
	node, exists := graph.byID[focused]
	if !exists || size.X <= 0 || size.Y <= 0 {
		return viewport, false
	}
	scale := worldScale(viewport, pixelsPerDP)
	visibleWidth := float32(size.X) / scale
	visibleHeight := float32(size.Y) / scale
	margin := min(defaultNodeGraphBehavior.focusRevealMarginPx/scale, min(visibleWidth, visibleHeight)/4)
	next := viewport
	if node.position.X < viewport.Origin.X+margin {
		next.Origin.X = node.position.X - margin
	} else if right := node.position.X + node.size.Width; right > viewport.Origin.X+visibleWidth-margin {
		next.Origin.X = right + margin - visibleWidth
	}
	if node.position.Y < viewport.Origin.Y+margin {
		next.Origin.Y = node.position.Y - margin
	} else if bottom := node.position.Y + node.size.Height; bottom > viewport.Origin.Y+visibleHeight-margin {
		next.Origin.Y = bottom + margin - visibleHeight
	}
	return next, next.Origin != viewport.Origin
}

func (w Widget) resolveViewport(value *graphState, graph resolvedGraph, size image.Point, pixelsPerDP float32) (Viewport, bool) {
	if !w.fitView {
		value.fitViewActive = false
	} else if !value.fitViewActive {
		if viewport, ok := fitGraphViewport(graph, size, pixelsPerDP, w.fitViewPadding, w.minZoom, w.maxZoom); ok {
			value.fitViewActive = true
			value.viewport = viewport
			value.ready = true
			return viewport, true
		}
	}
	if w.hasViewport {
		return normalizeViewport(w.viewport, w.minZoom, w.maxZoom), false
	}
	if !value.ready {
		if w.hasDefault {
			value.viewport = normalizeViewport(w.defaultViewport, w.minZoom, w.maxZoom)
		} else {
			value.viewport = Viewport{Zoom: 1}
		}
		value.ready = true
	}
	return normalizeViewport(value.viewport, w.minZoom, w.maxZoom), false
}
