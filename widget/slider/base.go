package slider

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

type sliderThumbState struct {
	focus animation.FocusAnimation
	scale animation.FloatTransition
}

type sliderBaseState struct {
	thumb      sliderThumbState
	upperThumb sliderThumbState
	styleCache sliderDefaultsCache
}

func (s Widget) layoutBase(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	values := s.resolvedValues()
	state := core.UseState[sliderBaseState](ctx, core.FullKey(ctx, s.key), "slider-base")
	visual := resolveSliderDefaults(&state.styleCache, core.ActiveTheme(ctx), s.orientation)
	inset := max(visual.Metrics.EdgeInset, visual.Metrics.ThumbLength/2)
	hitCross := max(visual.Metrics.ThumbLength, unit.Dp(24))
	step := s.step
	if step <= 0 {
		step = 1
	}
	description := s.semanticDescriptionFor(ctx, values)
	slider := base.Slider(s.key, s.value, s.minValue, s.maxValue).
		Step(step).
		Label(s.label).
		Description(description).
		Disabled(s.disabled).
		HitCross(hitCross).
		Inset(inset).
		Cursor(func(public base.SliderState) *pointer.Cursor {
			component := s.resolveBaseStyle(ctx, gtx, state, public)
			if component.track.Box != nil {
				return component.track.Box.Cursor
			}
			return nil
		}).
		OnChange(s.onChange).
		Header(func(public base.SliderState) core.Widget {
			return s.layoutBaseHeader(ctx, state, public)
		}).
		Track(func(public base.SliderState) core.Widget {
			return s.layoutBaseTrack(ctx, state, public)
		})
	if s.orientation == Vertical {
		slider = slider.Vertical()
	}
	return slider.Layout(ctx, gtx)
}

func (s Widget) layoutBaseHeader(ctx *core.Context, state *sliderBaseState, public base.SliderState) core.Widget {
	output := s.outputText(sliderValues{
		minValue: s.minValue, maxValue: s.maxValue, step: s.step,
		lower: public.Value, upper: public.Value,
	})
	if s.label == "" && output == "" {
		return core.WidgetFunc(func(*core.Context, layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		})
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		component := s.resolveBaseStyle(ctx, gtx, state, public)
		if s.orientation == Vertical {
			children := make([]layout.FlexChild, 0, 2)
			if output != "" {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return s.layoutOutput(ctx, gtx, component, output)
					})
				}))
			}
			if s.label != "" {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return s.layoutLabelVisual(ctx, gtx, component)
					})
				}))
			}
			return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(component.visual.Metrics.VerticalGap)}.Layout(gtx, children...)
		}
		children := make([]layout.FlexChild, 0, 3)
		if s.label != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutLabelVisual(ctx, gtx, component)
			}))
		}
		if output != "" {
			children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
			}))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutOutput(ctx, gtx, component, output)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutui.Spacer(0, gtx.Dp(component.visual.Metrics.HeaderGap)).Layout(ctx, gtx)
			}),
		)
	})
}

func (s Widget) layoutBaseTrack(ctx *core.Context, state *sliderBaseState, public base.SliderState) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		component := s.resolveBaseStyle(ctx, gtx, state, public)
		tokens := component.visual.Metrics
		return layoutui.Resolved(ctx, gtx, component.track, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			edge := max(gtx.Dp(tokens.EdgeInset), 0)
			minimumLength := 2*edge + max(gtx.Dp(tokens.ThumbExtra), 1)
			size := gtx.Constraints.Constrain(gtx.Constraints.Min)
			axisSize := s.axis().Convert(size)
			axisSize.X = max(axisSize.X, minimumLength)
			size = gtx.Constraints.Constrain(s.axis().Convert(axisSize))
			thumbSize := sliderThumbSize(gtx, component.thumb)
			geometry := newSliderGeometry(size, s.axis(), edge, public.Ratio, public.Ratio, false, thumbSize)
			s.layoutFill(ctx, gtx, component.fill, geometry, false, tokens.TrackRadius)
			s.layoutBaseThumb(ctx, gtx, state, public, component, geometry)
			return layout.Dimensions{Size: size}
		}))
	})
}

func (s Widget) resolveBaseStyle(ctx *core.Context, gtx layout.Context, state *sliderBaseState, public base.SliderState) sliderStyle {
	styleState := style.State{
		Hovered:      public.Hovered,
		Pressed:      public.Pressed,
		Focused:      public.Focused,
		FocusVisible: public.FocusVisible,
		Disabled:     public.Disabled,
		Dragging:     public.Dragging,
	}
	visual := resolveSliderDefaults(&state.styleCache, core.ActiveTheme(ctx), s.orientation)
	return s.resolveStyle(ctx, gtx, public.Key, styleState, visual)
}

func (s Widget) layoutBaseThumb(ctx *core.Context, gtx layout.Context, state *sliderBaseState, public base.SliderState, component sliderStyle, geometry sliderGeometry) {
	rect := geometry.thumbRects[0]
	focus := state.thumb.focus.OpacityFor(gtx, public.FocusVisible, component.visual.FocusDuration, component.visual.Motion)
	targetScale := float32(1)
	if public.Dragging {
		targetScale = component.visual.Metrics.DraggingScale
		if targetScale <= 0 || targetScale > 1 {
			targetScale = 0.9
		}
	}
	scale := state.thumb.scale.Value(gtx, targetScale, component.visual.DraggingDuration, animation.EaseSmoothstep, component.visual.Motion)
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	thumbGtx := gtx
	thumbGtx.Constraints = layout.Exact(rect.Size())
	s.layoutThumbVisual(ctx, thumbGtx, computed.ApplyOutlineOpacity(component.thumb, focus), scale)
	stack.Pop()
}

func (s Widget) layoutLabelVisual(ctx *core.Context, gtx layout.Context, component sliderStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, component.label, text.New(s.label))
}

func (s Widget) layoutBaseRange(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	values := s.resolvedValues()
	state := core.UseState[sliderBaseState](ctx, core.FullKey(ctx, s.key), "slider-base")
	visual := resolveSliderDefaults(&state.styleCache, core.ActiveTheme(ctx), s.orientation)
	inset := max(visual.Metrics.EdgeInset, visual.Metrics.ThumbLength/2)
	hitCross := max(visual.Metrics.ThumbLength, unit.Dp(24))
	step := s.step
	if step <= 0 {
		step = 1
	}
	description := s.semanticDescriptionFor(ctx, values)
	slider := base.RangeSlider(s.key, values.lower, values.upper, s.minValue, s.maxValue).
		Step(step).
		Label(s.label).
		Description(description).
		Disabled(s.disabled).
		HitCross(hitCross).
		Inset(inset).
		Cursor(func(public base.RangeSliderState) *pointer.Cursor {
			component := s.resolveBaseRangeStyle(ctx, gtx, state, public)
			if component.track.Box != nil {
				return component.track.Box.Cursor
			}
			return nil
		}).
		OnRangeChange(s.onRangeChange).
		Header(func(public base.RangeSliderState) core.Widget {
			return s.layoutBaseRangeHeader(ctx, state, public)
		}).
		Track(func(public base.RangeSliderState) core.Widget {
			return s.layoutBaseRangeTrack(ctx, state, public)
		})
	if s.orientation == Vertical {
		slider = slider.Vertical()
	}
	return slider.Layout(ctx, gtx)
}

func (s Widget) layoutBaseRangeHeader(ctx *core.Context, state *sliderBaseState, public base.RangeSliderState) core.Widget {
	output := s.outputText(sliderValues{
		minValue: s.minValue, maxValue: s.maxValue, step: s.step,
		lower: public.Lower, upper: public.Upper, rangeMode: true,
	})
	if s.label == "" && output == "" {
		return core.WidgetFunc(func(*core.Context, layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		})
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		component := s.resolveBaseRangeStyle(ctx, gtx, state, public)
		children := make([]layout.FlexChild, 0, 3)
		if s.label != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutLabelVisual(ctx, gtx, component)
			}))
		}
		if output != "" {
			children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
			}))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutOutput(ctx, gtx, component, output)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutui.Spacer(0, gtx.Dp(component.visual.Metrics.HeaderGap)).Layout(ctx, gtx)
			}),
		)
	})
}

func (s Widget) layoutBaseRangeTrack(ctx *core.Context, state *sliderBaseState, public base.RangeSliderState) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		component := s.resolveBaseRangeStyle(ctx, gtx, state, public)
		tokens := component.visual.Metrics
		return layoutui.Resolved(ctx, gtx, component.track, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			edge := max(gtx.Dp(tokens.EdgeInset), 0)
			minimumLength := 2*edge + max(gtx.Dp(tokens.ThumbExtra), 1)
			size := gtx.Constraints.Constrain(gtx.Constraints.Min)
			axisSize := s.axis().Convert(size)
			axisSize.X = max(axisSize.X, minimumLength)
			size = gtx.Constraints.Constrain(s.axis().Convert(axisSize))
			thumbSize := sliderThumbSize(gtx, component.thumb)
			geometry := newSliderGeometry(size, s.axis(), edge, public.LowerRatio, public.UpperRatio, true, thumbSize)
			s.layoutFill(ctx, gtx, component.fill, geometry, true, tokens.TrackRadius)
			s.layoutBaseRangeThumb(ctx, gtx, state, public, component, geometry, 0)
			s.layoutBaseRangeThumb(ctx, gtx, state, public, component, geometry, 1)
			return layout.Dimensions{Size: size}
		}))
	})
}

func (s Widget) resolveBaseRangeStyle(ctx *core.Context, gtx layout.Context, state *sliderBaseState, public base.RangeSliderState) sliderStyle {
	styleState := style.State{
		Hovered:      public.Hovered,
		Pressed:      public.Pressed,
		Focused:      public.Focused,
		FocusVisible: public.FocusVisible,
		Disabled:     public.Disabled,
		Dragging:     public.Dragging >= 0,
	}
	visual := resolveSliderDefaults(&state.styleCache, core.ActiveTheme(ctx), s.orientation)
	return s.resolveStyle(ctx, gtx, public.Key, styleState, visual)
}

func (s Widget) layoutBaseRangeThumb(ctx *core.Context, gtx layout.Context, state *sliderBaseState, public base.RangeSliderState, component sliderStyle, geometry sliderGeometry, index int) {
	rect := geometry.thumbRects[index]
	thumb := &state.thumb
	if index == 1 {
		thumb = &state.upperThumb
	}
	focused := public.Focused && public.Active == index
	focus := thumb.focus.OpacityFor(gtx, public.FocusVisible && focused, component.visual.FocusDuration, component.visual.Motion)
	targetScale := float32(1)
	if public.Dragging == index {
		targetScale = component.visual.Metrics.DraggingScale
		if targetScale <= 0 || targetScale > 1 {
			targetScale = 0.9
		}
	}
	scale := thumb.scale.Value(gtx, targetScale, component.visual.DraggingDuration, animation.EaseSmoothstep, component.visual.Motion)
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	thumbGtx := gtx
	thumbGtx.Constraints = layout.Exact(rect.Size())
	s.layoutThumbVisual(ctx, thumbGtx, computed.ApplyOutlineOpacity(component.thumb, focus), scale)
	stack.Pop()
}
