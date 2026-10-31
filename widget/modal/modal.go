package modal

import (
	"github.com/qianniancn/flowui/style"
	"image"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
)

// Widget renders an overlay dialog controlled by application state.
type Widget struct {
	style                   style.Style
	key                     string
	open                    bool
	hasOpen                 bool
	defaultOpen             bool
	hasDefaultOpen          bool
	title                   string
	body                    core.Widget
	header                  core.Widget
	footer                  core.Widget
	icon                    core.Widget
	onOpenChange            func(bool)
	size                    Size
	placement               Placement
	backdrop                BackdropVariant
	scroll                  Scroll
	animation               Animation
	dismissable             bool
	hasDismissable          bool
	keyboardDismissDisabled bool
	closeButton             bool
	hasCloseButton          bool
}

// Size selects the dialog width and full-screen behavior.
type Size int

const (
	// Medium uses the default modal width.
	Medium Size = iota
	// XSmall uses the smallest modal width.
	XSmall
	// Small uses a compact modal width.
	Small
	// Large uses a wider modal width.
	Large
	// Cover fills the available safe area while preserving page margins.
	Cover
	// Full fills the entire overlay.
	Full
)

// Placement selects where the dialog is placed inside the overlay.
type Placement int

const (
	// Auto places the modal based on viewport size.
	Auto Placement = iota
	// Top places the modal near the top edge.
	Top
	// Center places the modal in the center of the overlay.
	Center
	// Bottom places the modal near the bottom edge.
	Bottom
)

// BackdropVariant selects the backdrop treatment behind the dialog.
type BackdropVariant int

const (
	// BackdropOpaque draws the default opaque backdrop.
	BackdropOpaque BackdropVariant = iota
	// BackdropBlur draws a translucent blurred backdrop.
	BackdropBlur
	// BackdropTransparent keeps the backdrop visually transparent while still blocking input.
	BackdropTransparent
)

// ModalSize and ModalBackdrop are neutral size and backdrop values for hosts
// that keep their own presets independent of widget/modal. Their order mirrors
// Size and BackdropVariant.
type ModalSize uint8

const (
	ModalSizeMedium ModalSize = iota
	ModalSizeXSmall
	ModalSizeSmall
	ModalSizeLarge
	ModalSizeCover
	ModalSizeFull
)

type ModalBackdrop uint8

const (
	ModalBackdropOpaque ModalBackdrop = iota
	ModalBackdropBlur
	ModalBackdropTransparent
)

// Scroll selects whether the body or the dialog surface owns scrolling.
type Scroll int

const (
	// ScrollInside scrolls the body content inside a fixed dialog surface.
	ScrollInside Scroll = iota
	// ScrollOutside scrolls the entire dialog content inside the surface.
	ScrollOutside
)

// Animation selects the motion preset used when a modal enters and exits.
type Animation int

const (
	// AnimationAuto uses FlowUI's default modal animation.
	AnimationAuto Animation = iota
	// AnimationFadeScale fades the modal while scaling it up on entry and down on exit.
	AnimationFadeScale
	// AnimationSlideDown fades the modal while moving it down from above.
	AnimationSlideDown
	// AnimationSlideUp fades the modal while moving it up from below.
	AnimationSlideUp
	// AnimationFade only fades the modal without moving or scaling it.
	AnimationFade
	// AnimationBounceScale adds a restrained overshoot while opening and a clean scale-down while closing.
	AnimationBounceScale
	// AnimationZoomOut starts slightly larger and settles back to the modal's final size.
	AnimationZoomOut
	// AnimationPop opens with a short pop-style scale overshoot and closes with a clean scale-down.
	AnimationPop
)

// New creates a modal dialog. It starts uncontrolled and closed; Open sets the
// controlled value and DefaultOpen the uncontrolled starting value.
func New(title string, body core.Widget) Widget {
	return Widget{
		key:   core.AutoKey("modal"),
		title: title,
		body:  body,
	}
}

// Key sets an explicit identity for the dialog state.
func (m Widget) Key(key string) Widget {
	m.key = key
	return m
}

// Uncontrolled makes the constructor's open value an internal initial state.
func (m Widget) Uncontrolled() Widget {
	m.defaultOpen = m.open
	m.hasDefaultOpen = true
	m.hasOpen = false
	return m
}

// OnOpenChange registers a callback for close requests.
func (m Widget) OnOpenChange(fn func(bool)) Widget {
	m.onOpenChange = fn
	return m
}

// Open sets the modal to controlled mode with the given open state.
func (m Widget) Open(open bool) Widget {
	m.open = open
	m.hasOpen = true
	return m
}

// DefaultOpen sets the initial open state for uncontrolled mode.
func (m Widget) DefaultOpen(open bool) Widget {
	m.defaultOpen = open
	m.hasDefaultOpen = true
	return m
}

// Header replaces the default title/header area.
func (m Widget) Header(header core.Widget) Widget {
	m.header = header
	return m
}

// Body replaces the modal body content.
func (m Widget) Body(body core.Widget) Widget {
	m.body = body
	return m
}

// Footer sets the modal footer content.
func (m Widget) Footer(footer core.Widget) Widget {
	m.footer = footer
	return m
}

// Icon sets an optional icon shown above the default title.
func (m Widget) Icon(icon core.Widget) Widget {
	m.icon = icon
	return m
}

// Size sets the modal size preset.
func (m Widget) Size(size Size) Widget {
	m.size = size
	return m
}

// Placement sets the modal placement preset.
func (m Widget) Placement(placement Placement) Widget {
	m.placement = placement
	return m
}

// Backdrop sets the modal backdrop variant.
func (m Widget) Backdrop(backdrop BackdropVariant) Widget {
	m.backdrop = backdrop
	return m
}

// Scroll sets how overflowing modal content scrolls.
func (m Widget) Scroll(scroll Scroll) Widget {
	m.scroll = scroll
	return m
}

// Animation sets the modal enter and exit motion preset.
func (m Widget) Animation(animation Animation) Widget {
	m.animation = animation
	return m
}

// Dismissable controls whether clicking outside the dialog requests close.
func (m Widget) Dismissable(dismissable bool) Widget {
	m.dismissable = dismissable
	m.hasDismissable = true
	return m
}

// KeyboardDismissDisabled controls whether Escape can request close.
func (m Widget) KeyboardDismissDisabled(disabled bool) Widget {
	m.keyboardDismissDisabled = disabled
	return m
}

// CloseButton controls whether the default close button is shown.
func (m Widget) CloseButton(show bool) Widget {
	m.closeButton = show
	m.hasCloseButton = true
	return m
}

func (m Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return m.layoutWithStyle(ctx, gtx)
}

func (m Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	fullKey := core.FullKey(ctx, m.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)

	// Check if we need to evaluate open state (peek first to avoid claiming state).
	// A default-open uncontrolled binding is seeded once; after it is closed,
	// keep its state record so the default value is not applied again.
	stateValue, stateExists := core.PeekState[modalState](ctx, fullKey, stateSlotModal)
	needsState := stateExists && (stateValue.open || stateValue.visible() ||
		(!m.hasOpen && m.hasDefaultOpen && m.defaultOpen))
	if !needsState {
		// For uncontrolled mode, check if we have a default that would open it
		if !m.hasOpen && m.hasDefaultOpen && m.defaultOpen && !stateExists {
			needsState = true
		} else if m.hasOpen && m.open {
			needsState = true
		}
	}

	if !needsState {
		return layout.Dimensions{}
	}

	// Now we know we need state, claim it
	state := modalStateFor(ctx, m.key)
	state.bind(m)
	open := state.isOpen(m)
	defaults := m.defaultsWithCache(ctx, &state.visualCache)

	if !open && !state.visible() {
		if m.hasOpen || !m.hasDefaultOpen || !m.defaultOpen {
			deleteModalState(ctx, fullKey)
		}
		return layout.Dimensions{}
	}

	progress := state.progressFor(gtx, open, defaults.metrics.AnimationDuration, defaults.motion)
	if !open && progress <= 0 {
		if m.hasOpen || !m.hasDefaultOpen || !m.defaultOpen {
			deleteModalState(ctx, fullKey)
		}
		return layout.Dimensions{}
	}
	var tail func(layout.Context)
	if !naturallyDisabled {
		tail = func(gtx layout.Context) {
			state.dialog.LayoutEnd(gtx)
		}
	}
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:      fullKey,
		Layer:    core.OverlayLayerModal,
		Disabled: naturallyDisabled,
		Layout: func(gtx layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
			var close *widget.Clickable
			if m.showCloseButton() {
				close = &state.close
			}
			if state.dialog.Update(ctx, gtx, base.DialogPolicy{
				Open: open, Interactive: interactive, Close: close,
				DismissOutside: m.isDismissable(), DismissEscape: !m.keyboardDismissDisabled,
			}) {
				m.requestClose(state)
			}
			return m.layoutOverlay(ctx, gtx, state, progress, open && gtx.Enabled(), defaults)
		},
		Tail: tail,
	})
	state.dialog.ManageFocus(ctx, fullKey, open, !naturallyDisabled)

	return layout.Dimensions{}
}

func (m Widget) requestClose(state *modalState) {
	state.requestOpen(m, false)
}

func (m Widget) isDismissable() bool {
	if !m.hasDismissable {
		return true
	}
	return m.dismissable
}

func (m Widget) showCloseButton() bool {
	if !m.hasCloseButton {
		return true
	}
	return m.closeButton
}

func (m Widget) resolvedAnimation() Animation {
	if m.animation == AnimationAuto {
		return AnimationFadeScale
	}
	return m.animation
}

// Style merges instance declarations into the widget. They are resolved as the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (m Widget) Style(values ...style.Style) Widget {
	m.style = style.Join(m.style, style.Join(values...))
	return m
}
