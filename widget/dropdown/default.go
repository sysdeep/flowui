package dropdown

import (
	"image"
	"image/color"
	"sync"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/theme"
)

// dropdownDefaults is the theme-derived appearance and geometry of one
// dropdown. It is cached and contains no instance declaration; Style is
// resolved separately as the last cascade layer.
type dropdownDefaults struct {
	TriggerPressedScale float32
	PressDuration       time.Duration
	EnterDuration       time.Duration
	ExitDuration        time.Duration
	LongPressDuration   time.Duration
	HoverOpenDuration   time.Duration
	HoverCloseDuration  time.Duration
	PanelGap            unit.Dp
	ArrowSize           unit.Dp
	Motion              theme.MotionTheme
	Focus               dropdownFocusRing
}

const (
	dropdownDefaultTriggerFocusRingWidth  unit.Dp       = 2
	dropdownDefaultTriggerFocusRingOffset unit.Dp       = 2
	dropdownDefaultTriggerFocusRadius     unit.Dp       = 4
	dropdownDefaultTriggerPressedScale    float32       = 1
	dropdownDefaultPressDuration          time.Duration = 250 * time.Millisecond
	dropdownDefaultEnterDuration          time.Duration = 150 * time.Millisecond
	dropdownDefaultExitDuration           time.Duration = 100 * time.Millisecond
	dropdownDefaultLongPressDuration      time.Duration = 500 * time.Millisecond
	dropdownDefaultHoverOpenDuration      time.Duration = 150 * time.Millisecond
	dropdownDefaultHoverCloseDuration     time.Duration = 100 * time.Millisecond
	dropdownDefaultPanelGap               unit.Dp       = 4
	dropdownDefaultArrowSize              unit.Dp       = 12
)

// dropdownFocusRing is the trigger focus indication painted by the dropdown.
type dropdownFocusRing struct {
	Width  unit.Dp
	Offset unit.Dp
	Radius unit.Dp
	Color  color.NRGBA
}

// buildDropdown derives the trigger and popup chrome from the active theme.
// Popup menu appearance stays owned by the menu component.
func buildDropdown(active *theme.Theme) dropdownDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return dropdownDefaults{
		TriggerPressedScale: dropdownDefaultTriggerPressedScale,
		PressDuration:       dropdownDefaultPressDuration,
		EnterDuration:       dropdownDefaultEnterDuration,
		ExitDuration:        dropdownDefaultExitDuration,
		LongPressDuration:   dropdownDefaultLongPressDuration,
		HoverOpenDuration:   dropdownDefaultHoverOpenDuration,
		HoverCloseDuration:  dropdownDefaultHoverCloseDuration,
		PanelGap:            dropdownDefaultPanelGap,
		ArrowSize:           dropdownDefaultArrowSize,
		Motion:              active.Motion,
		Focus: dropdownFocusRing{
			Width:  dropdownDefaultTriggerFocusRingWidth,
			Offset: dropdownDefaultTriggerFocusRingOffset,
			Radius: dropdownDefaultTriggerFocusRadius,
			Color:  active.Semantic().Ring,
		},
	}
}

func (d dropdownDefaults) layoutFocus(gtx layout.Context, size image.Point, opacity float32) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	strokeWidth := max(gtx.Dp(d.Focus.Width), 1)
	inset := max(gtx.Dp(d.Focus.Offset), 0) + (strokeWidth+1)/2
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	cornerRadius := max(gtx.Dp(d.Focus.Radius)-inset, 0)
	focus := d.Focus.Color
	focus.A = byte(float32(focus.A)*opacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, cornerRadius).Path(gtx.Ops),
		Width: float32(strokeWidth),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, focus)
	stroke.Pop()
}

var dropdownTriggerGlyph struct {
	sync.Once
	value *widget.Icon
}

// layoutDropdownGlyph paints the default split-button trigger glyph.
func layoutDropdownGlyph(gtx layout.Context, foreground color.NRGBA) layout.Dimensions {
	const size = 16
	glyphGtx := gtx
	glyphGtx.Constraints = layout.Exact(image.Pt(gtx.Dp(size), gtx.Dp(size)))
	dropdownTriggerGlyphIcon().Layout(glyphGtx, foreground)
	return layout.Dimensions{Size: glyphGtx.Constraints.Min}
}

func dropdownTriggerGlyphIcon() *widget.Icon {
	dropdownTriggerGlyph.Do(func() {
		value, err := widget.NewIcon(lucide.Ellipsis)
		if err == nil {
			dropdownTriggerGlyph.value = value
		}
	})
	if dropdownTriggerGlyph.value == nil {
		panic("flowui/dropdown: missing trigger glyph")
	}
	return dropdownTriggerGlyph.value
}
