package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

const stateSlotColorArea = "color-area"

type ColorAreaWidget struct {
	style    style.Style
	key      string
	value    color.NRGBA
	label    string
	onChange func(color.NRGBA)
	disabled bool
	showDots bool
	color    *colorValueState
}

type colorAreaState struct {
	control       colorControlState
	color         colorValueState
	defaultsCache colorAreaDefaultsCache
}

func ColorArea(value color.NRGBA) ColorAreaWidget {
	return ColorAreaWidget{key: core.AutoKey("color-area"), value: value}
}

func (area ColorAreaWidget) Key(key string) ColorAreaWidget {
	area.key = key
	return area
}

func (area ColorAreaWidget) Label(label string) ColorAreaWidget {
	area.label = label
	return area
}

func (area ColorAreaWidget) OnChange(fn func(color.NRGBA)) ColorAreaWidget {
	area.onChange = fn
	return area
}

func (area ColorAreaWidget) Disabled(disabled bool) ColorAreaWidget {
	area.disabled = disabled
	return area
}

func (area ColorAreaWidget) ShowDots(show bool) ColorAreaWidget {
	area.showDots = show
	return area
}

func (area ColorAreaWidget) withColorState(state *colorValueState) ColorAreaWidget {
	area.color = state
	return area
}

func (area ColorAreaWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return area.layoutWithStyle(ctx, gtx)
}

func (area ColorAreaWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorArea, area.key)
	areaState := core.UseState[colorAreaState](ctx, key, stateSlotColorArea)
	defaults := colorAreaDefaultsFor(ctx, &areaState.defaultsCache)
	valueState := &areaState.color
	if area.color != nil {
		valueState = area.color
	} else {
		valueState.sync(area.value)
	}

	enabled := gtx.Enabled() && !area.disabled
	side := min(gtx.Dp(defaults.Size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(side, 0), max(side, 0))
	areaGtx := gtx
	areaGtx.Constraints = layout.Exact(size)
	current := valueState.hsv()
	if next, changed := areaState.control.updateArea(ctx, areaGtx, size, current, enabled); changed {
		nextColor := hsvToNRGBA(next)
		valueState.accept(nextColor, next.h)
		if area.onChange != nil && nextColor != area.value {
			area.onChange(nextColor)
		}
		current = next
	}

	focused := gtx.Focused(&areaState.control)
	state := style.State{
		Pressed:      areaState.control.Dragging,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &areaState.control, focused),
		Disabled:     !enabled,
		Dragging:     areaState.control.Dragging,
	}
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults.Defaults,
		style.Style{},
		style.Style{},
		area.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		thumbSize := defaults.ThumbSize
		if areaState.control.Dragging {
			thumbSize = defaults.DraggingThumbSize
		}
		opacity := paint.PushOpacity(gtx.Ops, colorControlOpacity(enabled, defaults.DisabledOpacity))
		colorDrawColorArea(
			gtx,
			size,
			current,
			max(gtx.Dp(defaults.Radius), 0),
			max(gtx.Dp(thumbSize), 1),
			max(gtx.Dp(defaults.ThumbBorderWidth), 1),
			max(gtx.Dp(defaults.FocusRingWidth), 1),
			areaState.control.focusOpacity(ctx, gtx, defaults.Motion),
			defaults.FocusColor,
			area.showDots,
			max(gtx.Dp(defaults.DotSize), 1),
			max(gtx.Dp(defaults.DotGap), 1),
		)
		opacity.Pop()
		areaState.control.AddInput(gtx, &areaState.control, image.Rectangle{Max: size}, enabled, pointer.CursorCrosshair, area.semanticLabel(ctx), formatHexColor(area.value, area.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (area ColorAreaWidget) semanticLabel(ctx *core.Context) string {
	if area.label != "" {
		return area.label
	}
	return ctx.Text(i18n.ColorPickerArea)
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w ColorAreaWidget) Style(values ...style.Style) ColorAreaWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
