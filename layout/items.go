package layout

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
)

// Items arranges Gio children in a wrapping row or a vertical column. Gaps are
// physical pixels; Widget-based callers normally use Wrap or Column instead.
func Items(ctx *Context, gtx layout.Context, horizontal bool, columnGap, rowGap int, children []layout.Widget) layout.Dimensions {
	columnGap = max(columnGap, 0)
	rowGap = max(rowGap, 0)
	if !horizontal {
		type trackedChild struct {
			dims      layout.Dimensions
			placement core.OverlayPlacement
		}
		tracked := make([]trackedChild, len(children))
		flexChildren := make([]layout.FlexChild, 0, len(children))
		for index, child := range children {
			flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
					return child(gtx)
				})
				tracked[index] = trackedChild{dims: dims, placement: placement}
				return dims
			}))
		}
		gap := max(rowGap, 0)
		dims := layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, flexChildren...)
		y := 0
		for index, child := range tracked {
			child.placement.PlaceOffset(image.Pt(0, y))
			y += child.dims.Size.Y
			if index < len(tracked)-1 {
				y += gap
			}
		}
		return dims
	}

	return layoutWrap(ctx, gtx, columnGap, rowGap, layout.Start, len(children), func(gtx layout.Context, index int) layout.Dimensions {
		return children[index](gtx)
	})
}
