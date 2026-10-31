package sidebar

import (
	"slices"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotSidebar = "sidebar"

type sidebarState struct {
	list         layout.List
	bar          widget.Scrollbar
	visualOutset layoutui.VisualOutsetState
	items        map[string]*sidebarItemState
	frameItems   map[string]struct{}
	itemIndex    map[string]int
	navigation   base.CollectionNavigation
	dataCache    sidebarDataCache
	disabledKeys core.StringSetCache
	defaults     sidebarDefaultsCache
}

// sidebarDefaultsCache memoizes the theme-derived appearance of one metric
// configuration. The instance style is resolved separately as the last cascade
// layer and is deliberately absent from the cache key.
type sidebarDefaultsCache struct {
	cache computed.VisualCache[sidebarMetrics, sidebarDefaults]
}

func (c *sidebarDefaultsCache) resolve(active *theme.Theme, metrics sidebarMetrics) sidebarDefaults {
	load := func() sidebarDefaults { return buildSidebarDefaults(active, metrics) }
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, metrics, load)
}

type sidebarDataCache struct {
	ready     bool
	version   uint64
	openKeys  []string
	collapsed bool
	entries   []entry
	items     []Item
}

type sidebarItemState struct {
	clickable          widget.Clickable
	focus              animation.FocusAnimation
	hoverOpenRequested bool
}

func sidebarStateFor(ctx *core.Context, key string) *sidebarState {
	key = core.ClaimKey(ctx, core.KindSidebar, key)
	return core.UseState[sidebarState](ctx, key, stateSlotSidebar)
}

func (s *sidebarState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *sidebarState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *sidebarState) item(key string) *sidebarItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *sidebarState) resolveEntries(widget Widget) ([]entry, []Item) {
	if !widget.hasDataVersion {
		validateSidebarWidget(widget)
		entries, items := widget.entriesAndItems()
		s.checkItems(items)
		s.dataCache.ready = false
		return entries, items
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion && s.dataCache.collapsed == widget.collapsed && slices.Equal(s.dataCache.openKeys, widget.openKeys) {
		return s.dataCache.entries, s.dataCache.items
	}
	validateSidebarWidget(widget)
	entries, items := widget.entriesAndItems()
	s.checkItems(items)
	s.dataCache = sidebarDataCache{
		ready:     true,
		version:   widget.dataVersion,
		openKeys:  append([]string(nil), widget.openKeys...),
		collapsed: widget.collapsed,
		entries:   entries,
		items:     items,
	}
	return entries, items
}

func (s *sidebarState) checkItems(items []Item) {
	if s.itemIndex == nil {
		s.itemIndex = make(map[string]int, len(items))
	} else {
		clear(s.itemIndex)
	}
	for index, item := range items {
		s.itemIndex[item.Key] = index
	}
}

type sidebarKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *sidebarState) updateKeys(gtx layout.Context, sidebar Widget, items []Item) sidebarKeyResult {
	selected := -1
	if index, ok := s.itemIndex[sidebar.selectedKey]; ok {
		selected = index
	}
	result := sidebarKeyResult{}
	s.navigation.Update(
		gtx,
		base.Collection{
			Count:    len(items),
			Key:      func(i int) string { return items[i].Key },
			Disabled: func(i int) bool { return sidebar.itemDisabled(items[i]) },
			Clickables: func(yield func(int, *widget.Clickable) bool) {
				for key, item := range s.items {
					if index, ok := s.itemIndex[key]; ok && !yield(index, &item.clickable) {
						return
					}
				}
			},
		},
		selected,
		base.CollectionPolicy{Activate: true},
		func(key string) { result.focusKey = key },
		func(key string) { result.actionKey = key },
	)
	return result
}

func (s *sidebarState) ensureVisible(index int) {
	if index < 0 || s.list.Position.Count == 0 {
		return
	}
	position := &s.list.Position
	if index < position.First || index >= position.First+position.Count {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First && position.Offset > 0 {
		position.Offset = 0
	} else if index == position.First+position.Count-1 && position.OffsetLast < 0 {
		position.Offset -= position.OffsetLast
	}
}
