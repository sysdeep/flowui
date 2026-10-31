package collapsible

import (
	"image"
	"math"
	"sync"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// collapsibleMetrics contains geometry and animation timing used by layout.
type collapsibleMetrics struct {
	BodyPadding       unit.Dp
	ContentGap        unit.Dp
	ContentDuration   time.Duration
	IndicatorDuration time.Duration
}

// collapsibleVisual is the resolved default appearance of a section.
type collapsibleVisual struct {
	Label   style.Style
	Metrics collapsibleMetrics
	Motion  theme.MotionTheme
}

const (
	collapsibleDefaultBodyPadding       unit.Dp       = 8
	collapsibleDefaultContentGap        unit.Dp       = 8
	collapsibleDefaultIndicatorSize     unit.Dp       = 16
	collapsibleDefaultIndicatorStroke   unit.Dp       = 1.7
	collapsibleDefaultContentDuration   time.Duration = 200 * time.Millisecond
	collapsibleDefaultIndicatorDuration time.Duration = 250 * time.Millisecond
)

func buildCollapsible(active *theme.Theme) collapsibleVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}

	sem := active.Semantic()
	labelColor := sem.Foreground
	disabledLabelColor := active.DisabledColor(labelColor)
	return collapsibleVisual{
		Label: style.Style{}.
			TextColor(style.Solid(labelColor)).
			FontSize(active.Typography.ControlSize).
			FontWeight(int(font.Medium)).
			On(style.Disabled, style.Style{}.TextColor(style.Solid(disabledLabelColor))),
		Metrics: collapsibleMetrics{
			BodyPadding:       collapsibleDefaultBodyPadding,
			ContentGap:        collapsibleDefaultContentGap,
			ContentDuration:   collapsibleDefaultContentDuration,
			IndicatorDuration: collapsibleDefaultIndicatorDuration,
		},
		Motion: active.Motion,
	}
}

var collapsibleIndicatorIcon struct {
	sync.Once
	value *widget.Icon
}

// collapsibleIndicatorRequest contains behavior-owned state for the disclosure
// indicator.
type collapsibleIndicatorRequest struct {
	Progress float32
	Disabled bool
}

func drawCollapsibleIndicator(active *theme.Theme, gtx layout.Context, request collapsibleIndicatorRequest) layout.Dimensions {
	diameter := max(gtx.Dp(collapsibleDefaultIndicatorSize), gtx.Dp(collapsibleDefaultIndicatorStroke*12))
	diameter = min(diameter, min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	if diameter <= 0 {
		return layout.Dimensions{}
	}

	size := image.Pt(diameter, diameter)
	icon := collapsibleDisclosureIcon()
	if icon == nil {
		return layout.Dimensions{Size: size}
	}
	color := active.Semantic().MutedForeground
	if request.Disabled {
		color = active.DisabledColor(color)
	}
	// The reported size is the icon box, so the caller's flex already places
	// it. Centering it again inside the leftover constraints moved the
	// chevron towards the middle of the trigger instead of its end.
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	rotation := op.Affine(f32.AffineId().Rotate(center, -request.Progress*float32(math.Pi))).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(iconGtx, color)
	rotation.Pop()
	return layout.Dimensions{Size: size}
}

func collapsibleDisclosureIcon() *widget.Icon {
	collapsibleIndicatorIcon.Do(func() {
		value, err := widget.NewIcon(lucide.ChevronDown)
		if err == nil {
			collapsibleIndicatorIcon.value = value
		}
	})
	return collapsibleIndicatorIcon.value
}
