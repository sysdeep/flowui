package progress

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type CircleWidget struct {
	style         style.Style
	key           string
	value         float64
	minValue      float64
	maxValue      float64
	label         string
	valueText     string
	hasValueText  bool
	indeterminate bool
	color         BarColor
	size          BarSize
	disabled      bool
}

const stateSlotProgressCircle = "progress-circle"

func Circle(key string, value float64) CircleWidget {
	return CircleWidget{key: key, value: value, maxValue: 100}
}

func (p CircleWidget) Label(label string) CircleWidget {
	p.label = label
	return p
}

func (p CircleWidget) ValueText(text string) CircleWidget {
	p.valueText = text
	p.hasValueText = true
	return p
}

func (p CircleWidget) Range(minValue, maxValue float64) CircleWidget {
	p.minValue = minValue
	p.maxValue = maxValue
	return p
}

func (p CircleWidget) Indeterminate() CircleWidget {
	p.indeterminate = true
	return p
}

func (p CircleWidget) Color(circleColor BarColor) CircleWidget {
	p.color = circleColor
	return p
}

func (p CircleWidget) Size(size BarSize) CircleWidget {
	p.size = size
	return p
}

func (p CircleWidget) Disabled(disabled bool) CircleWidget {
	p.disabled = disabled
	return p
}

func (p CircleWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindProgressCircle, p.key)
	progressState := progressCircleStateFor(ctx, key)
	resolved := p.resolveStyle(ctx, gtx, key, &progressState.circleStyleCache)
	progress := progressState.progress(gtx, p.ratio(), p.indeterminate, resolved.defaults.valueDuration, resolved.defaults.motion)
	return layoutui.Resolved(ctx, gtx, resolved.root, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		diameter := max(min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y), 1)
		size := gtx.Constraints.Constrain(image.Pt(diameter, diameter))
		iconSize := min(size.X, size.Y)
		if iconSize > 0 {
			offset := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
			period := time.Duration(0)
			if !p.disabled {
				period = theme.ResolveMotionDuration(resolved.defaults.motion, resolved.defaults.spinDuration)
			}
			progressDrawProgressCircle(
				gtx,
				iconSize,
				resolved.defaults.strokeRatio,
				resolved.track,
				resolved.fill,
				progress,
				p.indeterminate,
				period,
			)
			if p.indeterminate && period > 0 {
				gtx.Execute(op.InvalidateCmd{})
			}
			offset.Pop()
		}
		clipped := clip.Rect{Max: size}.Push(gtx.Ops)
		semantic.EnabledOp(!p.disabled).Add(gtx.Ops)
		semantic.DescriptionOp(p.semanticDescriptionFor(ctx)).Add(gtx.Ops)
		clipped.Pop()
		return layout.Dimensions{Size: size}
	}))
}

func (p CircleWidget) ratio() float32 {
	return progressRatio(p.value, p.minValue, p.maxValue, p.indeterminate)
}

func (p CircleWidget) semanticDescription() string {
	return p.semanticDescriptionFor(nil)
}

func (p CircleWidget) semanticDescriptionFor(ctx *core.Context) string {
	label := p.label
	if label == "" {
		label = ctx.Text(i18n.ProgressDefault)
	}
	if p.indeterminate {
		return ctx.Textf(i18n.ProgressInProgress, label)
	}
	value := p.valueText
	if !p.hasValueText || value == "" {
		value = fmt.Sprintf("%.0f%%", p.ratio()*100)
	}
	return ctx.Textf(i18n.ProgressValue, label, value)
}

func progressCircleStateFor(ctx *core.Context, key string) *progressBarState {
	return core.UseState[progressBarState](ctx, key, stateSlotProgressCircle)
}

type progressCircleResolvedStyle struct {
	root     style.ResolvedStyle
	track    color.NRGBA
	fill     color.NRGBA
	defaults progressCircleDefaults
}

func (p CircleWidget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, cache *progressCircleStyleCache) progressCircleResolvedStyle {
	active := core.ActiveTheme(ctx)
	declarations := cache.Resolve(active, progressStyleKey{color: p.color, size: p.size}, func() progressCircleDefaults {
		return p.defaultStyles(active)
	})
	state := style.State{Disabled: p.disabled}
	parts := [...]style.Part{style.PartRoot, style.PartTrack, style.PartFill}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:],
		resolved[:],
		declarations.defaults,
		declarations.variant,
		declarations.size,
		p.style,
	)
	track, ok := progressColor(resolved[1])
	if !ok {
		panic("flowui/progress.Circle: unresolved track color")
	}
	fill, ok := progressColor(resolved[2])
	if !ok {
		panic("flowui/progress.Circle: unresolved fill color")
	}
	return progressCircleResolvedStyle{
		root:     resolved[0],
		track:    track,
		fill:     fill,
		defaults: declarations,
	}
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w CircleWidget) Style(values ...style.Style) CircleWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
