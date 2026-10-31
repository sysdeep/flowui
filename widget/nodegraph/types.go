package nodegraph

// NodeGraphGridPattern identifies the grid primitive used by graph drawing.
type NodeGraphGridPattern uint8

const (
	NodeGraphGridLines NodeGraphGridPattern = iota
	NodeGraphGridDots
)

// NodeGraphEdgeType identifies the route geometry used by graph drawing.
type NodeGraphEdgeType uint8

const (
	NodeGraphEdgeBezier NodeGraphEdgeType = iota
	NodeGraphEdgeStraight
	NodeGraphEdgeStep
	NodeGraphEdgeSmoothStep
)

// NodeGraphEdgeMarker identifies an optional endpoint marker.
type NodeGraphEdgeMarker uint8

const (
	NodeGraphMarkerNone NodeGraphEdgeMarker = iota
	NodeGraphMarkerArrow
)
