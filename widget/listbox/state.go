package listbox

import (
	"fmt"

	"gioui.org/layout"
	giowidget "gioui.org/widget"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/optionrow"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/navigation"
)

const stateSlotListBox = "listbox"

func listBoxStateFor(ctx *core.Context, key string) *listBoxState {
	key = core.ClaimKey(ctx, core.KindListBox, key)
	return core.UseState[listBoxState](ctx, key, stateSlotListBox)
}

func (l Widget) stateFor(ctx *core.Context) *listBoxState {
	if l.derivedOwner == "" {
		return listBoxStateFor(ctx, l.key)
	}
	key := core.ClaimDerivedResolvedKey(ctx, core.KindListBox, l.derivedOwner, l.derivedRole)
	return core.UseState[listBoxState](ctx, key, stateSlotListBox)
}

type listBoxState struct {
	scrollToKey   string
	scrollToIndex int
	singleBinding binding.Binding[string]
	// For multiple selection, we manage state manually since []string is not comparable
	multipleValue     []string
	multipleReady     bool
	list              layout.List
	bar               giowidget.Scrollbar
	visualOutset      layoutui.VisualOutsetState
	items             map[string]*listBoxItemState
	frameItems        map[string]struct{}
	itemKeys          map[string]int
	navigation        base.CollectionNavigation
	dataCache         listBoxDataCache
	selectedKeys      core.StringSetCache
	disabledKeys      core.StringSetCache
	defaultsCache     computed.VisualCache[struct{}, listBoxDefaults]
	resolvedPartCache listBoxResolvedPartCache
}

type listBoxDataCache struct {
	ready   bool
	version uint64
	entries []listBoxEntry
	items   []Item
}

type listBoxKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *listBoxState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *listBoxState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *listBoxState) item(key string) *listBoxItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *listBoxState) resolveEntries(widget Widget) ([]listBoxEntry, []Item) {
	if !widget.hasDataVersion {
		entries, items := widget.entriesAndItems()
		s.checkItems(items)
		s.dataCache.ready = false
		return entries, items
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion {
		return s.dataCache.entries, s.dataCache.items
	}
	entries, items := widget.entriesAndItems()
	s.checkItems(items)
	s.dataCache = listBoxDataCache{
		ready:   true,
		version: widget.dataVersion,
		entries: entries,
		items:   items,
	}
	return entries, items
}

func (s *listBoxState) checkItems(items []Item) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]int, len(items))
	} else {
		clear(s.itemKeys)
	}
	for index, item := range items {
		if item.Key == "" {
			panic("flowui: empty listbox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate listbox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = index
	}
}

func (s *listBoxState) updateKeys(gtx layout.Context, items []Item, widget Widget, selectedKey string) listBoxKeyResult {
	result := listBoxKeyResult{}
	selected := -1
	if index, ok := s.itemKeys[selectedKey]; ok {
		selected = index
	}
	s.navigation.Update(gtx, base.Collection{
		Count:    len(items),
		Key:      func(i int) string { return items[i].Key },
		Label:    func(i int) string { return items[i].Label },
		Disabled: func(i int) bool { return widget.itemDisabled(items[i]) },
		Clickables: func(yield func(int, *giowidget.Clickable) bool) {
			for key, row := range s.items {
				if index, ok := s.itemKeys[key]; ok {
					if !yield(index, &row.Clickable) {
						return
					}
				}
			}
		},
	}, selected, base.CollectionPolicy{Activate: true, Typeahead: true},
		func(key string) { result.focusKey = key },
		func(key string) { result.actionKey = key })
	return result
}

func (s *listBoxState) keyboardActiveKey(widget Widget) string {
	if widget.selectionMode != SelectionMultiple {
		return widget.selectedKey
	}
	for _, key := range widget.selectedKeys {
		if _, ok := s.itemKeys[key]; ok {
			return key
		}
	}
	return ""
}

func listBoxIndexByKey(items []Item, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}

func listBoxItemByKey(items []Item, key string) (Item, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func listBoxItemDisabled(item Item, disabledKeys []string) bool {
	return item.Disabled || listBoxContainsKey(disabledKeys, item.Key)
}

// sliceList adapts an item slice + disabled-key slice into a navigation.List for the
// exported one-shot helpers used by Select/ComboBox. These lookups are not on
// the per-keystroke path, so the slice-based disabled check is acceptable.
func sliceList(items []Item, disabledKeys []string) navigation.List {
	return navigation.List{
		Count:    len(items),
		Disabled: func(i int) bool { return listBoxItemDisabled(items[i], disabledKeys) },
		Label:    func(i int) string { return items[i].Label },
	}
}

func FirstEnabled(items []Item, disabledKeys []string) (int, bool) {
	return navigation.First(sliceList(items, disabledKeys))
}

func LastEnabled(items []Item, disabledKeys []string) (int, bool) {
	return navigation.Last(sliceList(items, disabledKeys))
}

func IndexByKey(items []Item, key string) int {
	return listBoxIndexByKey(items, key)
}

func ItemDisabled(item Item, disabledKeys []string) bool {
	return listBoxItemDisabled(item, disabledKeys)
}

func FocusItem(ctx *core.Context, key, itemKey string, visible bool) bool {
	return focusItem(ctx, core.FullKey(ctx, key), itemKey, visible)
}

func FocusDerivedItem(ctx *core.Context, owner, role, itemKey string, visible bool) bool {
	return focusItem(ctx, core.DerivedKey(ctx, owner, role), itemKey, visible)
}

func focusItem(ctx *core.Context, stateKey, itemKey string, visible bool) bool {
	clickable, _, ok := item(ctx, stateKey, itemKey)
	if !ok {
		return false
	}
	core.RequestFocusVisible(ctx, clickable, visible)
	return true
}

type listBoxItemState struct {
	optionrow.FocusableState
	Clickable     giowidget.Clickable
	styleKey      string
	transitionKey string
}

// Binding helpers for single selection mode
func listBoxSingleBindingConfig(widget Widget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultKey,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *listBoxState) currentSingleValue(widget Widget) string {
	return s.singleBinding.Current(listBoxSingleBindingConfig(widget))
}

func (s *listBoxState) bindSingle(widget Widget) {
	s.singleBinding.Bind(listBoxSingleBindingConfig(widget))
}

func (s *listBoxState) requestSingleValue(widget Widget, value string) string {
	newValue, _ := s.singleBinding.Request(listBoxSingleBindingConfig(widget), value)
	return newValue
}

// Multiple-selection binding is maintained here ([]string is not comparable)
func (s *listBoxState) currentMultipleValue(widget Widget) []string {
	if widget.hasSelectedKeys {
		// Controlled mode
		return widget.selectedKeys
	}
	if !s.multipleReady {
		// First frame - initialize from default or empty
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
	return s.multipleValue
}

func (s *listBoxState) bindMultiple(widget Widget) {
	// For controlled mode, just validate
	if widget.hasSelectedKeys {
		return
	}
	// For uncontrolled mode, ensure initialized
	if !s.multipleReady {
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
}

func (s *listBoxState) requestMultipleValue(widget Widget, value []string) []string {
	if widget.hasSelectedKeys {
		// Controlled mode - just call onChange
		if widget.onSelectionChange != nil {
			widget.onSelectionChange(value)
		}
		return widget.selectedKeys
	}
	// Uncontrolled mode - update internal state
	s.multipleValue = append([]string(nil), value...)
	if widget.onSelectionChange != nil {
		widget.onSelectionChange(value)
	}
	return s.multipleValue
}
