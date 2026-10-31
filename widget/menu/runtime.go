package menu

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// Runtime binds a Menu widget to one owner-specific state instance.
type Runtime struct {
	widget Widget
	state  *menuState
}

func (m Widget) Runtime(ctx *core.Context, owner, role string, onRequestClose func(bool)) Runtime {
	return m.RuntimeWithStyle(ctx, owner, role, onRequestClose, style.Style{})
}

// RuntimeWithStyle binds a Menu widget to an owner-specific state instance and
// merges the owner's menu declaration into the widget's instance style.
func (m Widget) RuntimeWithStyle(ctx *core.Context, owner, role string, onRequestClose func(bool), declaration style.Style) Runtime {
	m = m.withDerivedIdentity(owner, role).withClose(onRequestClose).Style(declaration)
	return Runtime{widget: m, state: m.stateFor(ctx)}
}

func (r Runtime) Layout(ctx *core.Context, gtx layout.Context, interactive bool) layout.Dimensions {
	return r.LayoutWithVisual(ctx, gtx, interactive, r.Visual(ctx))
}

// LayoutWithVisual renders the runtime with a visual already resolved by its
// owner. This keeps overlay geometry, menu content, and instance style on the
// same resolved appearance within one layout pass.
func (r Runtime) LayoutWithVisual(ctx *core.Context, gtx layout.Context, interactive bool, visual RuntimeVisual) layout.Dimensions {
	return r.widget.layoutRootWithVisual(ctx, gtx, r.state, interactive, visual)
}

// Visual resolves the runtime's popup appearance, including its instance style.
// Owners that draw adjacent popup chrome can reuse the result.
func (r Runtime) Visual(ctx *core.Context) RuntimeVisual {
	return r.widget.visualWithCache(ctx, &r.state.visualCache)
}

// RootNavigation hands horizontal navigation at the root menu boundary to its
// owner. Nested submenus retain Left Arrow for returning to their parent.
func (r Runtime) RootNavigation(previous, next func()) Runtime {
	r.widget.onRootPrevious = previous
	r.widget.onRootNext = next
	return r
}

func (r Runtime) HasActiveSubmenu() bool {
	return r.state != nil && (r.state.openSubmenu != "" || r.state.submenuActive)
}

// HoveredWithSubmenus includes the visible descendants of a root menu. It is
// used by hover-triggered owners that need to know whether the pointer remains
// inside a nested popup, rather than merely whether a submenu is open.
func (r Runtime) HoveredWithSubmenus(ctx *core.Context) bool {
	if r.state == nil {
		return false
	}
	return menuHovered(ctx, r.widget, r.state)
}

func menuHovered(ctx *core.Context, widget Widget, state *menuState) bool {
	entries := state.resolveEntries(widget)
	actionable := state.actionableEntries(entries)
	for _, item := range state.items {
		if item.clickable.Hovered() {
			return true
		}
	}
	if state.openSubmenu == "" {
		return false
	}
	for _, entry := range actionable {
		if entry.item.Key != state.openSubmenu || !itemHasSubmenu(entry.item) {
			continue
		}
		child := widget.submenu(state, entry.item)
		childState := child.peekState(ctx)
		if childState == nil {
			continue
		}
		if childState.popup.Hovered() || menuHovered(ctx, child, childState) {
			return true
		}
	}
	return false
}

// peekState reads an existing menu state without claiming its key for the
// current layout pass. Hover inspection runs before the menu is laid out and
// must not make the later layout claim the same derived submenu key twice.
func (m Widget) peekState(ctx *core.Context) *menuState {
	var key string
	if m.derivedOwner == "" {
		key = core.FullKey(ctx, m.key)
	} else {
		key = core.DerivedKey(ctx, m.derivedOwner, m.derivedRole)
	}
	state, _ := core.PeekState[menuState](ctx, key, stateSlotMenu)
	return state
}

func (r Runtime) CloseSubmenus() {
	if r.state != nil {
		r.state.openSubmenu = ""
	}
}

func (r Runtime) FocusFirst(ctx *core.Context, visible bool) bool {
	return r.state != nil && r.state.focusFirstEntry(ctx, r.widget, visible)
}

func (r Runtime) FocusLast(ctx *core.Context, visible bool) bool {
	return r.state != nil && r.state.focusLastEntry(ctx, r.widget, visible)
}
