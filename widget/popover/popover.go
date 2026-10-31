package popover

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

type Widget struct {
	key                     string
	open                    bool
	hasOpen                 bool
	defaultOpen             bool
	hasDefaultOpen          bool
	trigger                 core.Widget
	content                 core.Widget
	heading                 string
	onOpenChange            func(bool)
	placement               core.PopoverPlacement
	offset                  int
	hasOffset               bool
	shouldFlip              bool
	hasShouldFlip           bool
	avoidOverflow           bool
	hasAvoidOverflow        bool
	arrow                   bool
	hasArrow                bool
	dismissable             bool
	hasDismissable          bool
	keyboardDismissDisabled bool
	style                   style.Style
}

// New creates a popover. It starts uncontrolled and closed; Open sets the
// controlled value and DefaultOpen the uncontrolled starting value.
func New(trigger core.Widget, content core.Widget) Widget {
	return Widget{
		key:     core.AutoKey("popover"),
		trigger: trigger,
		content: content,
	}
}

// Key sets an explicit identity for the popover state.
func (p Widget) Key(key string) Widget {
	p.key = key
	return p
}

// Uncontrolled makes the constructor's open value an internal initial state.
func (p Widget) Uncontrolled() Widget {
	p.defaultOpen = p.open
	p.hasDefaultOpen = true
	p.hasOpen = false
	return p
}

func (p Widget) OnOpenChange(fn func(bool)) Widget {
	p.onOpenChange = fn
	return p
}

func (p Widget) Open(open bool) Widget {
	p.open = open
	p.hasOpen = true
	return p
}

func (p Widget) DefaultOpen(open bool) Widget {
	p.defaultOpen = open
	p.hasDefaultOpen = true
	return p
}

func (p Widget) Heading(heading string) Widget {
	p.heading = heading
	return p
}

func (p Widget) Placement(placement core.PopoverPlacement) Widget {
	p.placement = placement
	return p
}

func (p Widget) Offset(dp int) Widget {
	p.offset = dp
	p.hasOffset = true
	return p
}

func (p Widget) ShouldFlip(shouldFlip bool) Widget {
	p.shouldFlip = shouldFlip
	p.hasShouldFlip = true
	return p
}

func (p Widget) AvoidOverflow(avoidOverflow bool) Widget {
	p.avoidOverflow = avoidOverflow
	p.hasAvoidOverflow = true
	return p
}

func (p Widget) Arrow(show bool) Widget {
	p.arrow = show
	p.hasArrow = true
	return p
}

func (p Widget) Dismissable(dismissable bool) Widget {
	p.dismissable = dismissable
	p.hasDismissable = true
	return p
}

func (p Widget) KeyboardDismissDisabled(disabled bool) Widget {
	p.keyboardDismissDisabled = disabled
	return p
}

func (p Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithStyle(ctx, gtx)
}

func (p Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	fullKey := core.FullKey(ctx, p.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	triggerDims := p.layoutTrigger(ctx, gtx)

	// Check if we need to evaluate open state (peek first to avoid claiming state).
	// A default-open uncontrolled binding is seeded once; after it is closed,
	// keep its state record so the default value is not applied again.
	stateValue, stateExists := core.PeekState[popoverState](ctx, fullKey, stateSlotPopover)
	needsState := stateExists && (stateValue.open || stateValue.visible() ||
		(!p.hasOpen && p.hasDefaultOpen && p.defaultOpen))
	if !needsState {
		// For uncontrolled mode, check if we have a default that would open it
		if !p.hasOpen && p.hasDefaultOpen && p.defaultOpen && !stateExists {
			needsState = true
		} else if p.hasOpen && p.open {
			needsState = true
		}
	}

	if !needsState {
		return triggerDims
	}

	// Now we know we need state, claim it
	state := popoverStateFor(ctx, p.key)
	state.bind(p)
	open := state.isOpen(p)

	if !open && !state.visible() {
		if p.hasOpen || !p.hasDefaultOpen || !p.defaultOpen {
			deletePopoverState(ctx, fullKey)
		}
		return triggerDims
	}

	visual := p.visualWithCache(ctx, &state.visualCache)
	progress := state.progressFor(gtx, open, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration, visual.Motion)
	if !open && progress <= 0 {
		if p.hasOpen || !p.hasDefaultOpen || !p.defaultOpen {
			deletePopoverState(ctx, fullKey)
		}
		return triggerDims
	}

	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       fullKey,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: triggerDims.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			if state.popup.Update(ctx, gtx, base.PopupPolicy{
				Open: open, Interactive: interactive,
				DismissOutside: p.isDismissable(), DismissEscape: !p.keyboardDismissDisabled,
			}) != base.PopupNotDismissed {
				p.requestClose(state)
			}
			return p.layoutOverlay(ctx, gtx, state, anchor, progress, open && gtx.Enabled(), visual)
		},
	})

	return triggerDims
}

func (p Widget) layoutTrigger(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if p.trigger == nil {
		return layout.Dimensions{}
	}
	return p.trigger.Layout(ctx, gtx)
}

func (p Widget) requestClose(state *popoverState) {
	state.requestOpen(p, false)
}

func (p Widget) isDismissable() bool {
	if !p.hasDismissable {
		return true
	}
	return p.dismissable
}

func (p Widget) showArrow() bool {
	return p.hasArrow && p.arrow
}

func (p Widget) flipEnabled() bool {
	if !p.hasShouldFlip {
		return true
	}
	return p.shouldFlip
}

func (p Widget) overflowAvoidanceEnabled() bool {
	if !p.hasAvoidOverflow {
		return true
	}
	return p.avoidOverflow
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (p Widget) Style(values ...style.Style) Widget {
	p.style = style.Join(p.style, style.Join(values...))
	return p
}
