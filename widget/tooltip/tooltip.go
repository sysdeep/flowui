package tooltip

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/text"
)

type Trigger uint8

const (
	Hover Trigger = iota
	Focus
)

type Widget struct {
	style            style.Style
	key              string
	trigger          core.Widget
	content          core.Widget
	placement        core.PopoverPlacement
	triggerMode      Trigger
	offset           int
	hasOffset        bool
	delay            time.Duration
	hasDelay         bool
	closeDelay       time.Duration
	hasCloseDelay    bool
	shouldFlip       bool
	hasShouldFlip    bool
	avoidOverflow    bool
	hasAvoidOverflow bool
	arrow            bool
	disabled         bool
}

// New attaches a text tooltip to trigger. Use Content for rich tooltip content.
func New(trigger core.Widget, label string) Widget {
	return Widget{
		key:       core.AutoKey("tooltip"),
		trigger:   trigger,
		content:   text.New(label),
		placement: core.PopoverTop,
	}
}

// Key sets a stable identity for tooltip hover and focus state.
func (t Widget) Key(key string) Widget { t.key = key; return t }

// Content replaces the tooltip text with custom content.
func (t Widget) Content(value core.Widget) Widget {
	t.content = value
	return t
}

func (t Widget) Placement(placement core.PopoverPlacement) Widget {
	t.placement = placement
	return t
}

func (t Widget) Trigger(trigger Trigger) Widget {
	t.triggerMode = trigger
	return t
}

func (t Widget) Offset(dp int) Widget {
	t.offset = dp
	t.hasOffset = true
	return t
}

func (t Widget) Delay(delay time.Duration) Widget {
	t.delay = max(delay, 0)
	t.hasDelay = true
	return t
}

func (t Widget) CloseDelay(delay time.Duration) Widget {
	t.closeDelay = max(delay, 0)
	t.hasCloseDelay = true
	return t
}

func (t Widget) ShouldFlip(shouldFlip bool) Widget {
	t.shouldFlip = shouldFlip
	t.hasShouldFlip = true
	return t
}

func (t Widget) AvoidOverflow(avoidOverflow bool) Widget {
	t.avoidOverflow = avoidOverflow
	t.hasAvoidOverflow = true
	return t
}

func (t Widget) Arrow(show bool) Widget {
	t.arrow = show
	return t
}

func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

func (t Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return t.layoutWithStyle(ctx, gtx)
}

// layoutWithStyle resolves the built-in appearance and applies the instance
// declaration as the last cascade layer.
func (t Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	coordinator := tooltipCoordinatorFor(ctx)
	coordinator.update(gtx)
	fullKey, state := tooltipStateFor(ctx, t.key)
	coordinator.register(ctx, gtx, fullKey)
	visual := tooltipVisualWithCache(ctx, &state.visualCache)
	disabled := t.disabled || !gtx.Enabled()
	wasActive := state.Active
	wasOpen := state.Open
	delay := t.showDelay(visual)
	if coordinator.warmed {
		delay = 0
	}
	closeDelay := t.hideDelay(visual)
	state.update(gtx, t.triggerMode, disabled, delay, closeDelay)
	if wasActive && !state.Active {
		coordinator.beginCooldown(gtx, fullKey, closeDelay)
	}
	if state.Open && !wasOpen {
		coordinator.open(fullKey, closeDelay)
		core.ActivateExclusive(ctx, tooltipExclusiveGroup, fullKey)
	} else if wasOpen && !state.Open {
		core.ReleaseExclusive(ctx, tooltipExclusiveGroup, fullKey)
	}

	triggerDims := t.layoutTrigger(ctx, gtx)
	state.addInput(gtx, triggerDims.Size, disabled)

	progress := state.progress(gtx, visual.Motion, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration)
	if progress <= 0 && !state.Open {
		return triggerDims
	}

	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       fullKey,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  true,
		Passive:   true,
		Layout: func(gtx layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			popup := t.popup(progress, state.exiting())
			popup.styleKey = fullKey
			popup.styleState = style.State{Disabled: disabled, Open: state.Open}
			popup.style = t.style
			return popup.layoutWithVisual(ctx, gtx, anchor, visual)
		},
	})
	return triggerDims
}

func (t Widget) layoutTrigger(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.trigger == nil {
		return layout.Dimensions{}
	}
	return t.trigger.Layout(ctx, gtx)
}

func (t Widget) showDelay(visual tooltipVisual) time.Duration {
	if t.hasDelay {
		return t.delay
	}
	return max(visual.Metrics.Delay, 0)
}

func (t Widget) hideDelay(visual tooltipVisual) time.Duration {
	if t.hasCloseDelay {
		return t.closeDelay
	}
	return max(visual.Metrics.CloseDelay, 0)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (t Widget) Style(values ...style.Style) Widget {
	t.style = style.Join(t.style, style.Join(values...))
	return t
}
