package tabs

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
)

func (s *tabsIndicatorState) transition(gtx layout.Context, key string, orientation Orientation, target image.Rectangle, motions ...theme.MotionTheme) image.Rectangle {
	return s.transitionWithDuration(gtx, key, orientation, target, tabsIndicatorDuration, motions...)
}

func (s *tabsItemState) selectionProgress(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	return s.selectionProgressWithDuration(gtx, selected, tabsColorDuration, motions...)
}

// tabsTestSizeStyle is a behavior-test measurement adapter. Production layout
// receives the component-owned tabsGeometry directly.
func tabsTestSizeStyle(size Size) tabsSizeStyle {
	return tabsSizeStyleFor(tabsGeometryFor(Primary), size)
}
