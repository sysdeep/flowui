package menu

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/binding"
)

const stateSlotContextMenu = "context-menu"
const contextMenuLongPressDelay = 500 * time.Millisecond

type contextMenuState struct {
	key          string
	trigger      base.PopupTrigger
	popup        base.PopupInteraction
	open         bool // cached effective open, updated by isOpen/requestOpen
	wasOpen      bool
	anchor       image.Rectangle
	hasAnchor    bool
	triggerSize  image.Point
	transition   animation.FloatTransition
	skipRestore  bool
	focusVisible bool
	focusTarget  event.Tag
	valueBinding binding.Binding[bool]
	visualCache  menuDefaultsCache
}

func contextMenuStateFor(ctx *core.Context, key string) *contextMenuState {
	key = core.ClaimKey(ctx, core.KindContextMenu, key)
	value := core.UseStateWith(ctx, key, stateSlotContextMenu, func() *contextMenuState {
		return &contextMenuState{key: key}
	})
	core.RegisterExclusive(ctx, "context-menu", key, value.closeForPeer)
	return value
}

func activateContextMenu(ctx *core.Context, value *contextMenuState) {
	if value != nil && value.key != "" {
		core.ActivateExclusive(ctx, "context-menu", value.key)
	}
}

func releaseContextMenu(ctx *core.Context, value *contextMenuState) {
	if value != nil {
		core.ReleaseExclusive(ctx, "context-menu", value.key)
	}
}

// contextMenuBindingConfig builds a binding.Config from the widget's open-state fields.
func contextMenuBindingConfig(widget ContextMenuWidget) binding.Config[bool] {
	return binding.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *contextMenuState) bind(widget ContextMenuWidget) {
	s.valueBinding.Bind(contextMenuBindingConfig(widget))
}

func (s *contextMenuState) isOpen(widget ContextMenuWidget) bool {
	s.open = s.valueBinding.Current(contextMenuBindingConfig(widget))
	return s.open
}

func (s *contextMenuState) requestOpen(ctx *core.Context, widget ContextMenuWidget, open bool) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.skipRestore = false
		activateContextMenu(ctx, s)
	}
	s.open, _ = s.valueBinding.Request(contextMenuBindingConfig(widget), open)
	if !s.open {
		releaseContextMenu(ctx, s)
	}
	return s.open
}

func (s *contextMenuState) closeForPeer() {
	s.skipRestore = true
	if s.valueBinding.PeerClose(false) {
		s.open = false
	}
}

func (s *contextMenuState) observeOpen(open bool) {
	if !open && s.wasOpen {
		s.trigger.Reset()
	}
	s.wasOpen = open
}

func (s *contextMenuState) progressFor(gtx layout.Context, open bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := exit
	if open {
		target = 1
		duration = enter
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}
