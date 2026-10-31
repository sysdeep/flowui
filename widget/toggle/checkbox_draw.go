package toggle

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// CheckboxControlOptions describes the shared Checkbox control used by Checkbox and Table.
type CheckboxControlOptions struct {
	Variant         CheckboxVariant
	Selection       float32
	Indeterminate   bool
	Hovered         bool
	Pressed         bool
	Focused         float32
	Disabled        bool
	Invalid         bool
	CustomIndicator bool
	Indicator       core.Widget
	// Style optionally customizes the shared control when it is drawn without
	// a CheckboxWidget's already-resolved appearance.
	Style style.Style
	// CheckboxVisualCache optionally reuses theme-derived defaults for compound widgets.
	VisualCache *CheckboxVisualCache

	resolvedStyle    *checkboxIndicatorStyle
	resolvedDefaults *checkboxDefaults
}

// DrawCheckboxControl renders the framework checkbox control without owning interaction
// state. Its appearance comes from the active theme.
func DrawCheckboxControl(ctx *core.Context, gtx layout.Context, options CheckboxControlOptions) layout.Dimensions {
	activeTheme := core.ActiveTheme(ctx)
	defaults := checkboxControlDefaults(activeTheme, options)
	indicator := options.resolvedStyle
	if indicator == nil {
		resolved := resolveCheckboxControlStyle(ctx, options, defaults)
		indicator = &resolved
	}
	return drawCheckbox(ctx, gtx, defaults, *indicator, min(max(options.Selection, 0), 1), min(max(options.Focused, 0), 1), options.Indeterminate, options.CustomIndicator, options.Indicator)
}

func checkboxControlDefaults(active *theme.Theme, options CheckboxControlOptions) checkboxDefaults {
	if options.resolvedDefaults != nil {
		return *options.resolvedDefaults
	}
	if options.VisualCache != nil {
		return options.VisualCache.resolve(active, options.Variant)
	}
	return buildCheckboxDefaults(active, options.Variant)
}

func resolveCheckboxControlStyle(ctx *core.Context, options CheckboxControlOptions, defaults checkboxDefaults) checkboxIndicatorStyle {
	state := style.State{
		Hovered: options.Hovered, Pressed: options.Pressed,
		FocusVisible: options.Focused > 0, Disabled: options.Disabled,
		Selected: options.Selection > 0, Checked: options.Selection > 0,
		Indeterminate: options.Indeterminate, Invalid: options.Invalid,
	}
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true
	return checkboxIndicatorStyle{
		off: computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults.Defaults, style.Style{}, style.Style{}, options.Style),
		on:  computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults.Defaults, style.Style{}, style.Style{}, options.Style),
	}
}

func drawCheckbox(ctx *core.Context, gtx layout.Context, defaults checkboxDefaults, resolved checkboxIndicatorStyle, selected, focused float32, indeterminate, customIndicator bool, indicator core.Widget) layout.Dimensions {
	metrics := defaults.Metrics
	controlSize := checkboxIndicatorSize(gtx, resolved)
	focusSpace := max(gtx.Dp(metrics.FocusSpace), 1)
	controlSize.X = min(controlSize.X, max(gtx.Constraints.Max.X-focusSpace*2, 0))
	controlSize.Y = min(controlSize.Y, max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := controlSize.Add(image.Pt(focusSpace*2, focusSpace*2))
	dims := gtx.Constraints.Constrain(bounds)
	if controlSize.X <= 0 || controlSize.Y <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-controlSize.X)/2, (dims.Y-controlSize.Y)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(controlSize),
	}
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		local := image.Rectangle{Max: gtx.Constraints.Min}
		var custom layout.Widget
		if customIndicator && indicator != nil {
			custom = func(indicatorGtx layout.Context) layout.Dimensions {
				return layoutCheckboxIndicator(ctx, indicatorGtx, metrics, local, indicator)
			}
		}
		if customIndicator {
			if custom != nil {
				return custom(gtx)
			}
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}
		// The checkmark is painted directly rather than through a text widget, so
		// read the resolved indicator foreground explicitly. This keeps the mark
		// aligned with the selected-state declaration in both light and dark themes.
		foreground := checkboxIndicatorForeground()
		return drawCheckboxIndicator(gtx, foreground, selected, indeterminate, metrics.IndeterminateStroke, metrics.CheckStroke)
	})
	layoutCheckboxLayer(ctx, gtx, rect, resolved.off, 1-selected, focused, nil)
	layoutCheckboxLayer(ctx, gtx, rect, resolved.on, selected, focused, content)

	return layout.Dimensions{Size: dims}
}

func checkboxIndicatorForeground() color.NRGBA {
	// The default checkmark is a glyph over the accent fill. Keep it opaque
	// white even when the surrounding control or its inherited foreground is
	// disabled/dimmed.
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}

func checkboxIndicatorSize(gtx layout.Context, resolved checkboxIndicatorStyle) image.Point {
	var size image.Point
	for _, endpoint := range [...]style.ResolvedStyle{resolved.off, resolved.on} {
		if endpoint.Box == nil || endpoint.Box.Width == nil || endpoint.Box.Height == nil {
			panic("flowui/toggle.Checkbox: missing indicator dimensions")
		}
		size.X = max(size.X, gtx.Dp(*endpoint.Box.Width))
		size.Y = max(size.Y, gtx.Dp(*endpoint.Box.Height))
	}
	if size.X <= 0 || size.Y <= 0 {
		panic("flowui/toggle.Checkbox: indicator has no pixel area")
	}
	return size
}

func layoutCheckboxLayer(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity, focused float32, child core.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, layerGtx, computed.ApplyOutlineOpacity(resolved, focused), child)
	fade.Pop()
	stack.Pop()
}

func layoutCheckboxIndicator(ctx *core.Context, gtx layout.Context, metrics checkboxMetrics, rect image.Rectangle, indicator core.Widget) layout.Dimensions {
	if indicator == nil {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	diameter := min(gtx.Dp(metrics.IndicatorSize), min(rect.Dx(), rect.Dy()))
	if diameter <= 0 {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	indicatorRect := image.Rectangle{Min: image.Pt((rect.Min.X+rect.Max.X-diameter)/2, (rect.Min.Y+rect.Max.Y-diameter)/2)}
	indicatorRect.Max = indicatorRect.Min.Add(image.Pt(diameter, diameter))
	stack := op.Offset(indicatorRect.Min).Push(gtx.Ops)
	indicatorGtx := gtx
	indicatorGtx.Constraints = layout.Exact(indicatorRect.Size())
	clipStack := clip.Rect{Max: indicatorRect.Size()}.Push(gtx.Ops)
	dims := layout.Center.Layout(indicatorGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
	clipStack.Pop()
	stack.Pop()
	return dims
}

// drawCheckboxIndicator owns the default checkmark and indeterminate-bar
// geometry. The caller supplies the resolved foreground and selection
// progress; the indicator surface and border come from the resolved part that
// hosts this drawing.
func drawCheckboxIndicator(
	gtx layout.Context,
	foreground color.NRGBA,
	selection float32,
	indeterminate bool,
	indeterminateStroke, checkStroke unit.Dp,
) layout.Dimensions {
	size := gtx.Constraints.Min
	selection = min(max(selection, 0), 1)
	if selection <= 0 || size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	value := foreground
	value.A = byte(float32(value.A)*selection + 0.5)
	if value.A == 0 {
		return layout.Dimensions{Size: size}
	}
	if indeterminate {
		length := max(int(float32(size.X)*0.58*selection+0.5), 1)
		thickness := max(gtx.Dp(indeterminateStroke), 1)
		center := image.Pt(size.X/2, size.Y/2)
		line := image.Rect(center.X-length/2, center.Y-thickness/2,
			center.X+(length+1)/2, center.Y+(thickness+1)/2)
		paint.FillShape(gtx.Ops, value, clip.UniformRRect(line, max(thickness/2, 1)).Op(gtx.Ops))
		return layout.Dimensions{Size: size}
	}

	width := float32(size.X)
	height := float32(size.Y)
	points := [3]f32.Point{
		f32.Pt(width*0.27, height*0.52),
		f32.Pt(width*0.43, height*0.68),
		f32.Pt(width*0.75, height*0.31),
	}
	path := draw.CheckPath(gtx.Ops, points, selection)
	stroke := clip.Stroke{Path: path, Width: float32(max(gtx.Dp(checkStroke), 1))}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, value)
	stroke.Pop()
	return layout.Dimensions{Size: size}
}
