package menubar

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	giowidget "gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/navigation"
)

const (
	stateSlotMenubar = "menubar"
	menubarExclusive = "menubar"
	menubarEnter     = 150 * time.Millisecond
	menubarExit      = 100 * time.Millisecond
)

type menubarState struct {
	key           string
	triggers      map[string]*menubarTriggerState
	frameTriggers map[string]struct{}
	itemKeys      map[string]struct{}
	navigation    base.MenubarNavigation
	openKey       string // cached effective open key, updated by current/requestOpen
	wasOpen       bool
	panelKey      string
	focusPanelKey string
	focusLast     bool
	focusVisible  bool
	hoveredKey    string
	popup         base.PopupInteraction
	transition    animation.FloatTransition
	valueBinding  binding.Binding[string]
	visualCache   menubarDefaultsCache
}

type menubarTriggerState struct {
	clickable giowidget.Clickable
	focus     animation.FocusAnimation
}

func menubarStateFor(ctx *core.Context, key string) *menubarState {
	key = core.ClaimKey(ctx, core.KindMenubar, key)
	value := core.UseStateWith(ctx, key, stateSlotMenubar, func() *menubarState {
		return &menubarState{key: key}
	})
	core.RegisterExclusive(ctx, menubarExclusive, key, value.closeForPeer)
	return value
}

// menubarBindingConfig builds a binding.Config from the widget's open-state fields.
func menubarBindingConfig(widget Widget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasOpenKey,
		Value:      widget.openKey,
		HasDefault: widget.hasDefaultOpenKey,
		Default:    widget.defaultOpenKey,
		OnChange:   widget.onOpenChange,
	}
}

func (s *menubarState) bind(widget Widget) {
	s.valueBinding.Bind(menubarBindingConfig(widget))
}

func (s *menubarState) beginFrame(items []Item) {
	core.BeginFrameMap(&s.frameTriggers)
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.key == "" {
			panic("flowui: empty menubar item key")
		}
		if item.label == "" {
			panic(fmt.Sprintf("flowui: empty menubar item label for key %q", item.key))
		}
		if _, exists := s.itemKeys[item.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate menubar item key %q", item.key))
		}
		s.itemKeys[item.key] = struct{}{}
		s.trigger(item.key)
	}
}

func (s *menubarState) endFrame() {
	core.SweepFrameMap(s.triggers, s.frameTriggers)
}

func (s *menubarState) trigger(key string) *menubarTriggerState {
	return core.UseFrameMap(&s.triggers, &s.frameTriggers, key)
}

func (s *menubarState) current(widget Widget) string {
	key := s.valueBinding.Current(menubarBindingConfig(widget))
	if widget.disabled || !widget.itemEnabled(key) {
		// If the current open key is now disabled, request close to clear it.
		if key != "" {
			s.openKey, _ = s.valueBinding.Request(menubarBindingConfig(widget), "")
			return s.openKey
		}
		key = ""
	}
	s.openKey = key
	return key
}

func (s *menubarState) requestOpen(ctx *core.Context, widget Widget, key string) string {
	if widget.disabled || !widget.itemEnabled(key) {
		key = ""
	}
	if key != "" {
		core.ActivateExclusive(ctx, menubarExclusive, s.key)
	}
	s.openKey, _ = s.valueBinding.Request(menubarBindingConfig(widget), key)
	if s.openKey == "" {
		core.ReleaseExclusive(ctx, menubarExclusive, s.key)
	}
	return s.openKey
}

func (s *menubarState) closeForPeer() {
	if s.valueBinding.PeerClose("") {
		s.openKey = ""
	}
}

func (s *menubarState) observeOpen(ctx *core.Context, key string) {
	if key != "" {
		s.panelKey = key
	}
	open := key != ""
	if open && !s.wasOpen {
		core.ActivateExclusive(ctx, menubarExclusive, s.key)
	} else if !open && s.wasOpen {
		core.ReleaseExclusive(ctx, menubarExclusive, s.key)
	}
	s.wasOpen = open
}

func (s *menubarState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := menubarExit
	if open {
		target = 1
		duration = menubarEnter
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *menubarState) updateInteractions(ctx *core.Context, gtx layout.Context, widget Widget) {
	openKey := s.current(widget)
	for _, item := range widget.items {
		trigger := s.trigger(item.key)
		eventGtx := gtx
		disabled := widget.disabled || item.disabled || !gtx.Enabled()
		if disabled {
			eventGtx = eventGtx.Disabled()
		}
		base.RowInteraction(ctx, eventGtx, &trigger.clickable, disabled, func() {
			visible := core.FocusVisible(ctx, &trigger.clickable, gtx.Focused(&trigger.clickable))
			if openKey == item.key {
				openKey = s.requestOpen(ctx, widget, "")
				s.focusPanelKey = ""
			} else {
				openKey = s.requestOpen(ctx, widget, item.key)
				s.focusPanelKey = item.key
				s.focusLast = false
				s.focusVisible = visible
			}
		})
	}

	hoveredKey := ""
	for _, item := range widget.items {
		if !widget.itemDisabled(item) && s.trigger(item.key).clickable.Hovered() {
			hoveredKey = item.key
			break
		}
	}
	if openKey != "" && hoveredKey != "" && hoveredKey != openKey && hoveredKey != s.hoveredKey {
		trigger := s.trigger(hoveredKey)
		openKey = s.requestOpen(ctx, widget, hoveredKey)
		s.focusPanelKey = ""
		s.focusLast = false
		s.focusVisible = false
		core.RequestFocusVisible(ctx, &trigger.clickable, false)
	}
	s.hoveredKey = hoveredKey

	s.updateGlobalKeys(ctx, gtx, widget, openKey)
	s.updateTriggerKeys(ctx, gtx, widget, openKey)
}

// updateGlobalKeys handles desktop menubar accelerators that are not
// scoped to a focused trigger: F10 focuses the bar; Alt+letter opens AccessKey menus.
func (s *menubarState) updateGlobalKeys(ctx *core.Context, gtx layout.Context, widget Widget, openKey string) {
	if widget.disabled || !gtx.Enabled() {
		return
	}
	filters := []event.Filter{
		key.Filter{Name: key.NameF10, Optional: key.ModAlt},
	}
	for _, item := range widget.items {
		if widget.itemDisabled(item) || item.accessKey == 0 {
			continue
		}
		name := key.Name(strings.ToUpper(string(item.accessKey)))
		filters = append(filters, key.Filter{
			Name:     name,
			Required: key.ModAlt,
		})
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		if event.Name == key.NameF10 {
			first := widget.firstEnabled()
			if first < 0 {
				continue
			}
			if openKey != "" {
				s.requestOpen(ctx, widget, "")
				s.focusPanelKey = ""
			}
			s.focusMovedTrigger(ctx, widget, first, "")
			continue
		}
		if event.Modifiers&key.ModAlt == 0 || event.Modifiers&^key.ModAlt != 0 {
			continue
		}
		text := navigation.Printable(event.Name)
		if text == "" {
			continue
		}
		letter := unicode.ToLower(rune(text[0]))
		for index, item := range widget.items {
			if widget.itemDisabled(item) || item.accessKey == 0 {
				continue
			}
			if item.accessKey == letter {
				// Focus without reusing the previous openKey, then open this item.
				s.focusMovedTrigger(ctx, widget, index, "")
				s.openFromTrigger(ctx, widget, index, false)
				break
			}
		}
	}
}

func (s *menubarState) updateTriggerKeys(ctx *core.Context, gtx layout.Context, widget Widget, openKey string) {
	axis := base.CollectionHorizontal
	if widget.orientation == Vertical {
		axis = base.CollectionVertical
	}
	eventGtx := gtx
	if widget.disabled {
		eventGtx = eventGtx.Disabled()
	}
	s.navigation.Update(eventGtx, base.Collection{
		Count:    len(widget.items),
		Key:      func(i int) string { return widget.items[i].key },
		Label:    func(i int) string { return widget.items[i].label },
		Disabled: func(i int) bool { return widget.itemDisabled(widget.items[i]) },
		Clickables: func(yield func(int, *giowidget.Clickable) bool) {
			for i, item := range widget.items {
				if !yield(i, &s.trigger(item.key).clickable) {
					return
				}
			}
		},
	}, axis, widget.loopFocus, base.MenubarActions{
		Focus: func(key string) { s.focusMovedTrigger(ctx, widget, widget.indexOf(key), openKey) },
		Open:  func(key string, last bool) { s.openFromTrigger(ctx, widget, widget.indexOf(key), last) },
		Close: func() {
			s.requestOpen(ctx, widget, "")
			s.focusPanelKey = ""
			openKey = ""
		},
	})
}

func (s *menubarState) focusMovedTrigger(ctx *core.Context, widget Widget, index int, openKey string) {
	if index < 0 || index >= len(widget.items) {
		return
	}
	item := widget.items[index]
	trigger := s.trigger(item.key)
	core.RequestFocusVisible(ctx, &trigger.clickable, true)
	if openKey != "" {
		s.requestOpen(ctx, widget, item.key)
		s.focusPanelKey = ""
		s.focusLast = false
	}
}

func (s *menubarState) openFromTrigger(ctx *core.Context, widget Widget, index int, last bool) {
	if index < 0 || index >= len(widget.items) || widget.itemDisabled(widget.items[index]) {
		return
	}
	item := widget.items[index]
	s.requestOpen(ctx, widget, item.key)
	s.focusPanelKey = item.key
	s.focusLast = last
	s.focusVisible = true
}

func (s *menubarState) focusTrigger(ctx *core.Context, key string, visible bool) {
	trigger := s.triggers[key]
	if trigger == nil {
		return
	}
	core.RequestFocusVisible(ctx, &trigger.clickable, visible)
}
