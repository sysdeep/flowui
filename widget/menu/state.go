package menu

import (
	"fmt"
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/navigation"
)

const stateSlotMenu = "menu"

type menuState struct {
	key                   string
	list                  layout.List
	bar                   widget.Scrollbar
	visualOutset          layoutui.VisualOutsetState
	items                 map[string]*menuItemState
	frameItems            map[string]struct{}
	itemKeys              map[string]struct{}
	itemIndex             map[string]int
	navigation            base.MenuNavigation
	dataCache             menuDataCache
	widthCache            menuWidthCache
	frameActionable       []entry
	anchors               map[string]image.Rectangle
	openSubmenu           string
	submenuActive         bool
	submenuFocusVisible   bool
	hoverSubmenu          string
	hoverAt               time.Time
	popup                 base.PopupInteraction
	transition            animation.FloatTransition
	submenuWasOpen        bool
	focusPending          bool
	requestedFocusVisible bool
	visualCache           menuDefaultsCache
}

type menuDataCache struct {
	ready                bool
	version              uint64
	autoSeparateSections bool
	entries              []entry
	actionable           []entry
}

type menuItemState struct {
	clickable widget.Clickable
	focus     animation.FocusAnimation
}

type keyResult struct {
	focusKey  string
	actionKey string
	openKey   string
	close     bool
}

func (m Widget) stateFor(ctx *core.Context) *menuState {
	var key string
	if m.derivedOwner == "" {
		key = core.ClaimKey(ctx, core.KindMenu, m.key)
	} else {
		key = core.ClaimDerivedResolvedKey(ctx, core.KindMenu, m.derivedOwner, m.derivedRole)
	}
	return core.UseStateWith(ctx, key, stateSlotMenu, func() *menuState {
		return &menuState{key: key}
	})
}

func (s *menuState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
	if s.anchors == nil {
		s.anchors = make(map[string]image.Rectangle)
	} else {
		clear(s.anchors)
	}
}

func (s *menuState) endFrame() {
	if s.openSubmenu != "" {
		if _, ok := s.frameItems[s.openSubmenu]; !ok {
			s.openSubmenu = ""
		}
	}
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *menuState) item(key string) *menuItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *menuState) resolveEntries(widget Widget) []entry {
	if !widget.hasDataVersion {
		entries := widget.entries()
		s.frameActionable = actionableEntries(entries)
		s.checkEntries(s.frameActionable)
		s.dataCache.ready = false
		return entries
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion &&
		s.dataCache.autoSeparateSections == widget.autoSeparateSections {
		return s.dataCache.entries
	}
	entries := widget.entries()
	actionable := actionableEntries(entries)
	s.checkEntries(actionable)
	s.dataCache = menuDataCache{
		ready:                true,
		version:              widget.dataVersion,
		autoSeparateSections: widget.autoSeparateSections,
		entries:              entries,
		actionable:           actionable,
	}
	return entries
}

func (s *menuState) actionableEntries(entries []entry) []entry {
	if s.dataCache.ready {
		return s.dataCache.actionable
	}
	return s.frameActionable
}

func (s *menuState) checkEntries(entries []entry) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(entries))
	} else {
		clear(s.itemKeys)
	}
	if s.itemIndex == nil {
		s.itemIndex = make(map[string]int, len(entries))
	} else {
		clear(s.itemIndex)
	}
	for index, entry := range entries {
		if entry.item.Key == "" {
			panic("flowui: empty menu item key")
		}
		if _, ok := s.itemKeys[entry.item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate menu item key %q", entry.item.Key))
		}
		s.itemKeys[entry.item.Key] = struct{}{}
		s.itemIndex[entry.item.Key] = index
	}
}

func (s *menuState) updateKeys(gtx layout.Context, menu Widget, entries []entry, nested bool) keyResult {
	result := keyResult{}
	s.navigation.Update(gtx, base.Collection{
		Count:    len(entries),
		Key:      func(i int) string { return entries[i].item.Key },
		Label:    func(i int) string { return entries[i].item.Label },
		Disabled: func(i int) bool { return menu.itemDisabled(entries[i].item) },
		Clickables: func(yield func(int, *widget.Clickable) bool) {
			first := max(s.list.Position.First-1, 0)
			last := min(first+s.list.Position.Count+2, len(entries))
			for index := first; index < last; index++ {
				if !yield(index, &s.item(entries[index].item.Key).clickable) {
					return
				}
			}
		},
	}, base.MenuPolicy{
		Nested: nested, HasSubmenu: func(i int) bool { return itemHasSubmenu(entries[i].item) },
	}, base.MenuActions{
		Focus:    func(key string) { result.focusKey = key },
		Activate: func(key string) { result.actionKey = key },
		Open:     func(key string) { result.openKey = key },
		Close:    func() { result.close = true },
		NextRoot: menu.onRootNext, PreviousRoot: menu.onRootPrevious,
	})
	return result
}

func actionableEntries(entries []entry) []entry {
	result := make([]entry, 0, len(entries))
	for _, entry := range entries {
		if !entry.separator && entry.sectionTitle == "" {
			result = append(result, entry)
		}
	}
	return result
}

// menuNavList adapts menu entries into a navigation.List; itemDisabled also skips
// non-actionable entries (separators/section titles), preserving prior behavior.
func menuNavList(widget Widget, entries []entry) navigation.List {
	return navigation.List{
		Count:    len(entries),
		Disabled: func(i int) bool { return widget.itemDisabled(entries[i].item) },
		Label:    func(i int) string { return entries[i].item.Label },
	}
}

func menuFirstEnabled(widget Widget, entries []entry) int {
	index, _ := navigation.First(menuNavList(widget, entries))
	return index
}

func menuLastEnabled(widget Widget, entries []entry) int {
	index, _ := navigation.Last(menuNavList(widget, entries))
	return index
}

func entryByKey(entries []entry, key string) (entry, bool) {
	for _, entry := range entries {
		if entry.item.Key == key {
			return entry, true
		}
	}
	return entry{}, false
}

func (s *menuState) focus(ctx *core.Context, entry entry, visible bool) {
	item := s.item(entry.item.Key)
	core.RequestFocusVisible(ctx, &item.clickable, visible)
}

func (s *menuState) focusFirstEntry(ctx *core.Context, widget Widget, visible bool) bool {
	entries := widget.actionableEntries()
	index := menuFirstEnabled(widget, entries)
	if index < 0 {
		return false
	}
	s.focus(ctx, entries[index], visible)
	return true
}

func (s *menuState) focusLastEntry(ctx *core.Context, widget Widget, visible bool) bool {
	entries := widget.actionableEntries()
	index := menuLastEnabled(widget, entries)
	if index < 0 {
		return false
	}
	s.focus(ctx, entries[index], visible)
	return true
}

func (s *menuState) reveal(allEntries []entry, key string) {
	entryIndex := -1
	for index, entry := range allEntries {
		if entry.item.Key == key {
			entryIndex = index
			break
		}
	}
	if entryIndex < 0 || s.list.Position.Count == 0 {
		return
	}
	first := s.list.Position.First
	if entryIndex < first || entryIndex >= first+s.list.Position.Count {
		s.list.ScrollTo(entryIndex)
	}
}
