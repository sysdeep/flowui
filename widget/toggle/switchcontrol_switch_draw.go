package toggle

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

type switchSwitchDrawResult struct {
	dims      layout.Dimensions
	trackRect image.Rectangle
	thumbRect image.Rectangle
}

func switchSwitchRects(gtx layout.Context, focusSpaceDp unit.Dp, style switchSwitchStyle) switchSwitchDrawResult {
	focusSpace := max(gtx.Dp(focusSpaceDp), 1)
	trackWidthDp, trackHeightDp := switchSwitchTrackDpSize(style)
	thumbWidthDp, thumbHeightDp := switchSwitchThumbDpSize(style)
	trackWidth := min(gtx.Dp(trackWidthDp), max(gtx.Constraints.Max.X-focusSpace*2, 0))
	trackHeight := min(gtx.Dp(trackHeightDp), max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := image.Pt(trackWidth+focusSpace*2, trackHeight+focusSpace*2)
	dims := layout.Dimensions{Size: gtx.Constraints.Constrain(bounds)}
	if trackWidth <= 0 || trackHeight <= 0 {
		return switchSwitchDrawResult{dims: dims}
	}

	trackOrigin := image.Pt((dims.Size.X-trackWidth)/2, (dims.Size.Y-trackHeight)/2)
	track := image.Rectangle{
		Min: trackOrigin,
		Max: trackOrigin.Add(image.Pt(trackWidth, trackHeight)),
	}
	thumbWidth := min(gtx.Dp(thumbWidthDp), trackWidth)
	thumbHeight := min(gtx.Dp(thumbHeightDp), trackHeight)
	padding := max((trackHeight-thumbHeight)/2, 0)
	travel := max(trackWidth-padding*2-thumbWidth, 0)
	thumbX := track.Min.X + padding + int(float32(travel)*style.selected+0.5)
	thumbY := track.Min.Y + (trackHeight-thumbHeight)/2
	thumb := image.Rectangle{
		Min: image.Pt(thumbX, thumbY),
		Max: image.Pt(thumbX+thumbWidth, thumbY+thumbHeight),
	}
	return switchSwitchDrawResult{
		dims:      dims,
		trackRect: track,
		thumbRect: thumb,
	}
}

func switchSwitchTrackDpSize(style switchSwitchStyle) (unit.Dp, unit.Dp) {
	return switchSwitchPartDpSize(style.trackOff, style.trackOn)
}

func switchSwitchThumbDpSize(style switchSwitchStyle) (unit.Dp, unit.Dp) {
	return switchSwitchPartDpSize(style.thumbOff, style.thumbOn)
}

func switchSwitchPartDpSize(first, second style.ResolvedStyle) (unit.Dp, unit.Dp) {
	var width, height unit.Dp
	for _, value := range [...]style.ResolvedStyle{first, second} {
		if value.Box == nil {
			continue
		}
		if value.Box.Width != nil {
			width = max(width, *value.Box.Width)
		}
		if value.Box.Height != nil {
			height = max(height, *value.Box.Height)
		}
	}
	return width, height
}
