package splitpane

import (
	"image"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotSplitPane = "split-pane"

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

type Widget struct {
	key             string
	first           core.Widget
	second          core.Widget
	orientation     Orientation
	ratio           float32
	defaultRatio    float32
	hasRatio        bool
	hasDefaultRatio bool
	minFirst        unit.Dp
	minSecond       unit.Dp
	label           string
	onRatioChange   func(float32)
	disabled        bool
	customStyle     style.Style
}

type splitPaneState struct {
	ratio              float32
	initialized        bool
	interaction        base.SplitPaneInteraction
	focus              animation.FocusAnimation
	firstVisualOutset  layoutui.VisualOutsetState
	secondVisualOutset layoutui.VisualOutsetState
}

// PrepareFieldAssociation registers relationships before parent layout reorders children.
func (s Widget) PrepareFieldAssociation(ctx *core.Context) bool {
	layoutui.PrepareFieldAssociations(ctx, s.first, s.second)
	return true
}

// New creates two panes with a draggable divider.
func New(first, second core.Widget) Widget {
	return Widget{
		key:    core.AutoKey("split-pane"),
		first:  first,
		second: second,
		label:  "Resize panels",
	}
}

// Key sets an explicit identity for the divider state.
func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

func (s Widget) Orientation(orientation Orientation) Widget {
	s.orientation = orientation
	return s
}

func (s Widget) Horizontal() Widget {
	s.orientation = Horizontal
	return s
}

func (s Widget) Vertical() Widget {
	s.orientation = Vertical
	return s
}

func (s Widget) Ratio(ratio float32) Widget {
	s.ratio = splitPaneRatio(ratio)
	s.hasRatio = true
	return s
}

func (s Widget) DefaultRatio(ratio float32) Widget {
	s.defaultRatio = splitPaneRatio(ratio)
	s.hasDefaultRatio = true
	return s
}

func (s Widget) MinFirst(dp int) Widget {
	s.minFirst = unit.Dp(max(dp, 0))
	return s
}

func (s Widget) MinSecond(dp int) Widget {
	s.minSecond = unit.Dp(max(dp, 0))
	return s
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) OnRatioChange(fn func(float32)) Widget {
	s.onRatioChange = fn
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

// Style replaces the pane's instance style with the given declarations joined
// in order.
func (s Widget) Style(values ...style.Style) Widget {
	s.customStyle = style.Join(values...)
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	layoutui.PrepareFieldAssociations(ctx, s.first, s.second)
	key := core.ClaimKey(ctx, core.KindSplitPane, s.key)
	value := core.UseState[splitPaneState](ctx, key, stateSlotSplitPane)
	disabled := s.disabled || !gtx.Enabled()
	focused := gtx.Focused(&value.interaction)
	styleState := style.State{
		Hovered:      value.interaction.Hovered && !disabled,
		Pressed:      value.interaction.Dragging && !disabled,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &value.interaction, focused),
		Disabled:     disabled,
		Dragging:     value.interaction.Dragging && !disabled,
	}
	root, style := s.resolveStyle(ctx, gtx, key, styleState)
	return layoutui.Resolved(ctx, gtx, root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return s.layoutPane(ctx, gtx, value, style, disabled)
	}))
}

func (s Widget) layoutPane(ctx *core.Context, gtx layout.Context, value *splitPaneState, style splitPaneStyle, disabled bool) layout.Dimensions {
	size := gtx.Constraints.Constrain(gtx.Constraints.Max)
	axis := s.axis()
	axisSize := axis.Convert(size)
	tokens := defaultSplitPaneMetrics
	divider := min(max(gtx.Dp(style.dividerWidth), 1), axisSize.X)
	available := max(axisSize.X-divider, 0)
	minFirst, maxFirst := s.bounds(gtx, available)
	first := splitPaneFirstSize(s.resolveRatio(value), available, minFirst, maxFirst)
	ratio := splitPaneSizeRatio(first, available)

	enabled := gtx.Enabled() && !s.disabled && available > 0 && minFirst < maxFirst
	if next, changed := value.interaction.Update(ctx, gtx, base.SplitPaneGeometry{
		Axis: axis, Available: available, First: first, Divider: divider, Minimum: minFirst, Maximum: maxFirst,
	}, !enabled); changed {
		first = splitPaneFirstSize(next, available, minFirst, maxFirst)
		next = splitPaneSizeRatio(first, available)
		if next != ratio && s.onRatioChange != nil {
			s.onRatioChange(next)
		}
		ratio = next
	}
	value.ratio = ratio
	value.initialized = true

	second := available - first
	firstSize := axis.Convert(image.Pt(first, axisSize.Y))
	secondSize := axis.Convert(image.Pt(second, axisSize.Y))
	secondOffset := axis.Convert(image.Pt(first+divider, 0))
	s.layoutChild(ctx, gtx, s.first, firstSize, image.Point{}, &value.firstVisualOutset)
	s.layoutChild(ctx, gtx, s.second, secondSize, secondOffset, &value.secondVisualOutset)

	focusVisible := core.FocusVisible(ctx, &value.interaction, gtx.Focused(&value.interaction))
	focusOpacity := value.focus.Opacity(gtx, focusVisible && enabled, core.ActiveTheme(ctx).Motion)
	drawSplitPaneDivider(ctx, gtx, axis, size, first, divider, value.interaction.Hovered, value.interaction.Dragging, focusOpacity, disabled, style)
	value.interaction.AddHandle(gtx, splitPaneHitRect(axis, size, first, divider, max(gtx.Dp(tokens.HitSize), divider)), s.label, !enabled, style.cursor)
	return layout.Dimensions{Size: size}
}

func (s Widget) axis() layout.Axis {
	if s.orientation == Vertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

func (s Widget) resolveRatio(value *splitPaneState) float32 {
	if s.hasRatio {
		return s.ratio
	}
	if value.initialized {
		return value.ratio
	}
	if s.hasDefaultRatio {
		return s.defaultRatio
	}
	return 0.5
}

func (s Widget) bounds(gtx layout.Context, available int) (int, int) {
	minFirst := min(max(gtx.Dp(s.minFirst), 0), available)
	minSecond := min(max(gtx.Dp(s.minSecond), 0), available)
	if minFirst+minSecond <= available {
		return minFirst, available - minSecond
	}
	total := minFirst + minSecond
	if total == 0 {
		return 0, available
	}
	fixed := int(math.Round(float64(available) * float64(minFirst) / float64(total)))
	return fixed, fixed
}

func (s Widget) layoutChild(ctx *core.Context, gtx layout.Context, child core.Widget, size, offset image.Point, visual *layoutui.VisualOutsetState) {
	childGtx := gtx
	childGtx.Constraints = layout.Exact(size)
	stack := op.Offset(offset).Push(gtx.Ops)
	defer stack.Pop()
	_, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layoutui.Viewport(ctx, childGtx, visual, child)
	})
	placement.PlaceOffset(offset)
}

func splitPaneHitRect(axis layout.Axis, size image.Point, first, divider, hitSize int) image.Rectangle {
	center := first + divider/2
	start := center - hitSize/2
	end := start + hitSize
	if axis == layout.Horizontal {
		return image.Rect(max(start, 0), 0, min(end, size.X), size.Y)
	}
	return image.Rect(0, max(start, 0), size.X, min(end, size.Y))
}

func splitPaneFirstSize(ratio float32, available, minimum, maximum int) int {
	first := int(math.Round(float64(splitPaneRatio(ratio) * float32(available))))
	return min(max(first, minimum), maximum)
}

func splitPaneSizeRatio(size, available int) float32 {
	if available <= 0 {
		return 0.5
	}
	return float32(size) / float32(available)
}

func splitPaneRatio(ratio float32) float32 {
	if math.IsNaN(float64(ratio)) || math.IsInf(float64(ratio), 0) {
		return 0.5
	}
	return min(max(ratio, 0), 1)
}
