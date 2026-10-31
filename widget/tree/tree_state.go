package tree

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/toggle"
)

const stateSlotTree = "tree"

const treeDragMIMEPrefix = "application/x-flowui-tree-item/"

const (
	treeExpandDuration  = 200 * time.Millisecond
	treeDragExpandDelay = 600 * time.Millisecond
	treeDragScrollSpeed = 8
)

type flatItem struct {
	item          Item
	depth         int
	parentKey     string
	isLast        bool
	ancestorsLast []bool
}

func flattenVisibleItems(items []Item, expanded map[string]struct{}) []flatItem {
	result := make([]flatItem, 0)
	var walk func([]Item, int, string, []bool)
	walk = func(children []Item, depth int, parent string, ancestorsLast []bool) {
		for index, item := range children {
			isLast := index == len(children)-1
			result = append(result, flatItem{
				item:          item,
				depth:         depth,
				parentKey:     parent,
				isLast:        isLast,
				ancestorsLast: append([]bool(nil), ancestorsLast...),
			})
			if _, ok := expanded[item.Key]; ok && len(item.Children) > 0 {
				walk(item.Children, depth+1, item.Key, append(append([]bool(nil), ancestorsLast...), isLast))
			}
		}
	}
	walk(items, 0, "", nil)
	return result
}

type treeState struct {
	list               layout.List
	bar                widget.Scrollbar
	visualOutset       layoutui.VisualOutsetState
	items              map[string]*treeItemState
	frameItems         map[string]struct{}
	itemKeys           map[string]struct{}
	navigation         base.CollectionNavigation
	selectionAnchor    string
	dragMIME           string
	dragSource         string
	dragSources        []string
	dropTarget         treeDropTarget
	dragEventActive    bool
	dragEventSources   []string
	dragEventTarget    treeDropTarget
	dragHoverKey       string
	dragHoverAt        time.Time
	dragHoverRequested bool
	dragScrollAt       time.Time
	renameKey          string
	renameOriginal     string
	renameEditor       widget.Editor
	renameFocused      bool
	renameRequestKey   string
	renameRequest      uint64
	renameRequestReady bool
	dataCache          treeDataCache
	dragGeometry       treeDragGeometry
	selectedKeys       core.StringSetCache
	checkedKeys        core.StringSetCache
	halfCheckedKeys    core.StringSetCache
	disabledKeys       core.StringSetCache
	checkboxVisual     toggle.CheckboxVisualCache
	visualCache        treeVisualCache
}

type treeDataCache struct {
	ready         bool
	version       uint64
	filterQuery   string
	customFilter  bool
	filterVersion uint64
	expandedKeys  []string
	visible       []flatItem
}

// treeDragGeometry keeps the stable row geometry needed while an item is
// dragged. It avoids rebuilding a full height table for every pointer core.
type treeDragGeometry struct {
	ready             bool
	cacheable         bool
	dataVersion       uint64
	filterQuery       string
	customFilter      bool
	filterVersion     uint64
	expandedKeys      []string
	rowHeight         int
	descriptionHeight int
	gap               int
	heights           []int
	starts            []int
	subtreeEnd        []int
	indexByKey        map[string]int
}

func (s *treeState) dragGeometryFor(tree Widget, visible []flatItem, rowHeight, descriptionHeight, gap int) *treeDragGeometry {
	cacheable := tree.hasDataVersion && (tree.filterFunc == nil || tree.hasFilterVersion)
	geometry := &s.dragGeometry
	if geometry.ready && geometry.cacheable && cacheable &&
		geometry.dataVersion == tree.dataVersion &&
		geometry.filterQuery == tree.filterQuery &&
		geometry.customFilter == (tree.filterFunc != nil) &&
		geometry.filterVersion == tree.filterVersion &&
		geometry.rowHeight == rowHeight && geometry.descriptionHeight == descriptionHeight && geometry.gap == gap &&
		len(geometry.heights) == len(visible) && treeKeysEqual(geometry.expandedKeys, tree.expandedKeys) {
		return geometry
	}

	geometry.ready = true
	geometry.cacheable = cacheable
	geometry.dataVersion = tree.dataVersion
	geometry.filterQuery = tree.filterQuery
	geometry.customFilter = tree.filterFunc != nil
	geometry.filterVersion = tree.filterVersion
	geometry.expandedKeys = append(geometry.expandedKeys[:0], tree.expandedKeys...)
	geometry.rowHeight = rowHeight
	geometry.descriptionHeight = descriptionHeight
	geometry.gap = gap
	geometry.heights = slices.Grow(geometry.heights[:0], len(visible))
	geometry.starts = slices.Grow(geometry.starts[:0], len(visible)+1)
	geometry.subtreeEnd = slices.Grow(geometry.subtreeEnd[:0], len(visible))
	geometry.starts = append(geometry.starts, 0)
	if geometry.indexByKey == nil {
		geometry.indexByKey = make(map[string]int, len(visible))
	} else {
		clear(geometry.indexByKey)
	}
	for index, entry := range visible {
		height := rowHeight
		if treeItemDescription(entry.item) != "" {
			height = descriptionHeight
		}
		geometry.heights = append(geometry.heights, height)
		geometry.starts = append(geometry.starts, geometry.starts[index]+height+gap)
		geometry.subtreeEnd = append(geometry.subtreeEnd, index)
		geometry.indexByKey[entry.item.Key] = index
	}
	stack := make([]int, 0)
	for index, entry := range visible {
		for len(stack) > 0 && entry.depth <= visible[stack[len(stack)-1]].depth {
			parent := stack[len(stack)-1]
			geometry.subtreeEnd[parent] = index - 1
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, index)
	}
	for _, index := range stack {
		geometry.subtreeEnd[index] = len(visible) - 1
	}
	return geometry
}

func (s *treeState) resetDragGeometry() {
	s.dragGeometry = treeDragGeometry{}
}

func treeStateFor(ctx *core.Context, key string) *treeState {
	key = core.ClaimKey(ctx, core.KindTree, key)
	value := core.UseState[treeState](ctx, key, stateSlotTree)
	value.dragMIME = treeDragMIMEPrefix + key
	return value
}

func (s *treeState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *treeState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *treeState) item(key string) *treeItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *treeState) resolveVisible(tree Widget) []flatItem {
	filterActive := tree.filterFunc != nil || tree.filterQuery != ""
	if !tree.hasDataVersion || (tree.filterFunc != nil && !tree.hasFilterVersion) {
		s.dataCache = treeDataCache{}
		s.checkItems(tree.items)
		items, expanded := filteredTreeData(tree, filterActive)
		return flattenVisibleItems(items, treeKeySet(expanded))
	}
	cache := &s.dataCache
	if cache.ready && cache.version == tree.dataVersion && cache.filterQuery == tree.filterQuery && cache.customFilter == (tree.filterFunc != nil) && cache.filterVersion == tree.filterVersion && treeKeysEqual(cache.expandedKeys, tree.expandedKeys) {
		return cache.visible
	}
	s.checkItems(tree.items)
	items, expanded := filteredTreeData(tree, filterActive)
	cache.visible = flattenVisibleItems(items, treeKeySet(expanded))
	cache.expandedKeys = append(cache.expandedKeys[:0], tree.expandedKeys...)
	cache.version = tree.dataVersion
	cache.filterQuery = tree.filterQuery
	cache.customFilter = tree.filterFunc != nil
	cache.filterVersion = tree.filterVersion
	cache.ready = true
	return cache.visible
}

func filteredTreeData(tree Widget, filterActive bool) ([]Item, []string) {
	items := tree.items
	expanded := tree.expandedKeys
	if !filterActive {
		return items, expanded
	}
	if tree.filterFunc != nil {
		items = filterTreeItemsBy(items, tree.filterFunc)
	} else {
		items = filterTreeItems(items, tree.filterQuery)
	}
	// Keep filter matches reachable by expanding every retained branch.
	expanded = append([]string(nil), expanded...)
	expanded = append(expanded, collectExpandableKeys(items)...)
	return items, expanded
}

func filterTreeItems(items []Item, query string) []Item {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return items
	}
	return filterTreeItemsBy(items, func(item Item) bool {
		return strings.Contains(strings.ToLower(item.Label), query) ||
			strings.Contains(strings.ToLower(item.Description), query)
	})
}

func filterTreeItemsBy(items []Item, matches func(Item) bool) []Item {
	if matches == nil {
		return items
	}
	var walk func([]Item) []Item
	walk = func(children []Item) []Item {
		out := make([]Item, 0, len(children))
		for _, item := range children {
			filteredChildren := walk(item.Children)
			if matches(item) || len(filteredChildren) > 0 {
				copyItem := item
				copyItem.Children = filteredChildren
				out = append(out, copyItem)
			}
		}
		return out
	}
	return walk(items)
}

func collectExpandableKeys(items []Item) []string {
	var keys []string
	var walk func([]Item)
	walk = func(children []Item) {
		for _, item := range children {
			if len(item.Children) > 0 {
				keys = append(keys, item.Key)
				walk(item.Children)
			}
		}
	}
	walk(items)
	return keys
}

func (s *treeState) visibleIndex(visible []flatItem, key string, item *treeItemState) int {
	if item.index >= 0 && item.index < len(visible) && visible[item.index].item.Key == key {
		return item.index
	}
	item.index = treeVisibleIndex(visible, key)
	return item.index
}

func (s *treeState) checkItems(items []Item) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{})
	} else {
		clear(s.itemKeys)
	}
	var walk func([]Item)
	walk = func(children []Item) {
		for _, item := range children {
			if item.Key == "" {
				panic("flowui: empty tree item key")
			}
			if _, ok := s.itemKeys[item.Key]; ok {
				panic(fmt.Sprintf("flowui: duplicate tree item key %q", item.Key))
			}
			s.itemKeys[item.Key] = struct{}{}
			walk(item.Children)
		}
	}
	walk(items)
}

type treeKeyResult struct {
	focusKey        string
	actionKey       string
	checkKey        string
	actionModifiers key.Modifiers
	rangeKey        string
	toggleKey       string
	loadEvent       LoadEvent
	renameKey       string
}

func (s *treeState) updateKeys(gtx layout.Context, tree Widget, visible []flatItem) treeKeyResult {
	if tree.disabled {
		gtx = gtx.Disabled()
	}
	result := treeKeyResult{}
	var expanded map[string]struct{}
	s.navigation.UpdateWithEvents(gtx, base.Collection{
		Count:    len(visible),
		Key:      func(i int) string { return visible[i].item.Key },
		Label:    func(i int) string { return visible[i].item.Label },
		Disabled: func(i int) bool { return tree.itemDisabled(visible[i].item) },
		Fallback: func() int { return tree.keyboardActiveIndex(visible) },
		Clickables: func(yield func(int, *widget.Clickable) bool) {
			for itemKey, itemState := range s.items {
				if index := s.visibleIndex(visible, itemKey, itemState); index >= 0 {
					if !yield(index, &itemState.clickable) {
						return
					}
				}
			}
		},
	}, -1, base.CollectionPolicy{
		Activate: true, Typeahead: true, Modifiers: key.ModShift,
		Commands: []key.Name{key.NameRightArrow, key.NameLeftArrow, key.NameF2},
	}, func(e base.CollectionEvent) int {
		current := e.Index
		entry := visible[current]
		switch e.Kind {
		case base.CollectionFocus:
			result.focusKey = entry.item.Key
			rangeMovement := e.Name == key.NameDownArrow || e.Name == key.NameUpArrow || e.Name == key.NameHome || e.Name == key.NameEnd
			if rangeMovement && e.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
				result.rangeKey = entry.item.Key
			}
		case base.CollectionActivate:
			if e.Name == key.NameSpace && tree.checkable {
				result.checkKey = entry.item.Key
			} else if e.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
				result.rangeKey = entry.item.Key
			} else {
				result.actionKey, result.actionModifiers = entry.item.Key, e.Modifiers
			}
		case base.CollectionCommand:
			if e.Name == key.NameF2 {
				if entry.item.Renamable && tree.onRename != nil {
					result.renameKey = entry.item.Key
				}
				return current
			}
			if expanded == nil {
				expanded = treeKeySet(tree.expandedKeys)
			}
			_, isExpanded := expanded[entry.item.Key]
			switch e.Name {
			case key.NameRightArrow:
				if treeItemExpandable(entry.item) {
					if !isExpanded {
						result.toggleKey = entry.item.Key
						if entry.item.ChildrenState == ChildrenUnloaded || entry.item.ChildrenState == ChildrenError {
							result.loadEvent = treeLoadEvent(entry.item, false)
						}
					} else if entry.item.ChildrenState == ChildrenError {
						result.loadEvent = treeLoadEvent(entry.item, true)
					} else if child := treeFirstEnabledChild(visible, tree, current); child >= 0 {
						current = child
						result.focusKey = visible[child].item.Key
					}
				}
			case key.NameLeftArrow:
				if treeItemExpandable(entry.item) && isExpanded {
					result.toggleKey = entry.item.Key
				} else if parent := treeEnabledAncestor(visible, tree, current); parent >= 0 {
					current = parent
					result.focusKey = visible[parent].item.Key
				}
			}
		}
		return current
	})
	return result
}

func (s *treeState) ensureVisible(index int) {
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

type treeItemState struct {
	index       int
	clickable   widget.Clickable
	check       widget.Clickable
	toggle      core.ClickArea
	focus       animation.FocusAnimation
	expansion   animation.FloatTransition
	drag        core.DragState
	dropTargets [3]core.DropTarget
}

type treeDropTarget struct {
	key      string
	drawKey  string
	depth    int
	position DropPosition
	index    int
}

func (s *treeItemState) updateDrag(gtx layout.Context, mime string, sourceKeys []string) bool {
	payload, _ := json.Marshal(sourceKeys)
	return s.drag.Update(gtx, mime, payload, float32(gtx.Dp(3)))
}

func (t Widget) updateDropEvents(gtx layout.Context, state *treeState, itemState *treeItemState, visible []flatItem, entry flatItem) {
	for index := range itemState.dropTargets {
		for {
			data, ok := itemState.dropTargets[index].Next(gtx, state.dragMIME, core.DefaultMaxDropData)
			if !ok {
				break
			}
			sourceKeys, ok := treeDropSource(data)
			position := DropPosition(index)
			if ok && t.dropAllowed(visible, sourceKeys, entry.item.Key, position) {
				t.onDrop(treeDropEvent(sourceKeys, entry.item.Key, position))
			}
		}
	}
}

func treeDropSource(data []byte) ([]string, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var keys []string
	if json.Unmarshal(data, &keys) != nil {
		return nil, false
	}
	unique := keys[:0]
	for _, itemKey := range keys {
		if itemKey != "" && !treeContainsKey(unique, itemKey) {
			unique = append(unique, itemKey)
		}
	}
	return unique, len(unique) > 0
}

func treeDropAllowed(tree Widget, visible []flatItem, sourceKeys []string, targetKey string, position DropPosition) bool {
	targetIndex := treeVisibleIndex(visible, targetKey)
	if len(sourceKeys) == 0 || targetIndex < 0 {
		return false
	}
	if tree.itemDisabled(visible[targetIndex].item) || visible[targetIndex].item.DropDisabled {
		return false
	}
	if position == DropInside && !treeItemAcceptsChildren(visible[targetIndex].item) {
		return false
	}
	for _, sourceKey := range sourceKeys {
		sourceIndex := treeVisibleIndex(visible, sourceKey)
		if sourceIndex < 0 || sourceIndex == targetIndex || tree.itemDisabled(visible[sourceIndex].item) || visible[sourceIndex].item.DragDisabled {
			return false
		}
		if targetIndex <= sourceIndex {
			continue
		}
		sourceDepth := visible[sourceIndex].depth
		outsideSource := false
		for index := sourceIndex + 1; index <= targetIndex; index++ {
			if visible[index].depth <= sourceDepth {
				outsideSource = true
				break
			}
		}
		if !outsideSource {
			return false
		}
	}
	return true
}

func (t Widget) dropAllowed(visible []flatItem, sourceKeys []string, targetKey string, position DropPosition) bool {
	if !treeDropAllowed(t, visible, sourceKeys, targetKey, position) {
		return false
	}
	return t.canDrop == nil || t.canDrop(treeDropEvent(sourceKeys, targetKey, position))
}

func (t Widget) dragDropAllowed(visible []flatItem, geometry *treeDragGeometry, sourceKeys []string, target treeDropTarget) bool {
	if geometry == nil || target.index < 0 || target.index >= len(visible) || len(sourceKeys) == 0 {
		return false
	}
	targetItem := visible[target.index].item
	if t.itemDisabled(targetItem) || targetItem.DropDisabled || target.position == DropInside && !treeItemAcceptsChildren(targetItem) {
		return false
	}
	for _, sourceKey := range sourceKeys {
		sourceIndex, ok := geometry.indexByKey[sourceKey]
		if !ok || sourceIndex == target.index || t.itemDisabled(visible[sourceIndex].item) || visible[sourceIndex].item.DragDisabled {
			return false
		}
		if target.index > sourceIndex && target.index <= geometry.subtreeEnd[sourceIndex] {
			return false
		}
	}
	return t.canDrop == nil || t.canDrop(treeDropEvent(sourceKeys, target.key, target.position))
}

func treeDropEvent(sourceKeys []string, targetKey string, position DropPosition) DropEvent {
	keys := append([]string(nil), sourceKeys...)
	return DropEvent{SourceKey: keys[0], SourceKeys: keys, TargetKey: targetKey, Position: position}
}

func (t Widget) updateDragLifecycle(state *treeState, sourceKeys []string, target treeDropTarget) {
	if !state.dragEventActive || !treeKeysEqual(state.dragEventSources, sourceKeys) {
		if state.dragEventActive {
			t.endDragLifecycle(state)
		}
		state.dragEventActive = true
		state.dragEventSources = append(state.dragEventSources[:0], sourceKeys...)
		if t.onDragStart != nil {
			t.onDragStart(treeDragEvent(state.dragEventSources, treeDropTarget{}))
		}
	}
	previous := state.dragEventTarget
	changed := previous.key != target.key || previous.position != target.position
	if changed && previous.key != "" && t.onDragLeave != nil {
		t.onDragLeave(treeDragEvent(state.dragEventSources, previous))
	}
	if changed {
		state.dragEventTarget = target
	}
	if changed && target.key != "" && t.onDragEnter != nil {
		t.onDragEnter(treeDragEvent(state.dragEventSources, target))
	}
	if target.key != "" && t.onDragOver != nil {
		t.onDragOver(treeDragEvent(state.dragEventSources, target))
	}
}

func (t Widget) endDragLifecycle(state *treeState) {
	if !state.dragEventActive {
		return
	}
	if state.dragEventTarget.key != "" && t.onDragLeave != nil {
		t.onDragLeave(treeDragEvent(state.dragEventSources, state.dragEventTarget))
	}
	if t.onDragEnd != nil {
		t.onDragEnd(treeDragEvent(state.dragEventSources, state.dragEventTarget))
	}
	state.dragEventActive = false
	state.dragEventSources = state.dragEventSources[:0]
	state.dragEventTarget = treeDropTarget{}
}

func treeDragEvent(sourceKeys []string, target treeDropTarget) DragEvent {
	keys := append([]string(nil), sourceKeys...)
	event := DragEvent{SourceKeys: keys, TargetKey: target.key, Position: target.position}
	if len(keys) > 0 {
		event.SourceKey = keys[0]
	}
	return event
}

func treeDropTargetAt(visible []flatItem, sourceIndex int, pointerY float32, heights []int, gap int) treeDropTarget {
	if sourceIndex < 0 || sourceIndex >= len(visible) || len(heights) != len(visible) {
		return treeDropTarget{}
	}
	top := float32(0)
	for index := sourceIndex - 1; index >= 0; index-- {
		top -= float32(heights[index] + gap)
	}
	for index, entry := range visible {
		height := heights[index]
		bottom := top + float32(height)
		if pointerY >= top && pointerY < bottom {
			return treeDropTarget{
				key: entry.item.Key, drawKey: entry.item.Key, depth: entry.depth,
				position: treeDropPositionAt(pointerY-top, height, treeItemAcceptsChildren(entry.item)), index: index,
			}
		}
		top = bottom + float32(gap)
	}
	return treeDropTarget{}
}

func (g *treeDragGeometry) dropTargetAt(visible []flatItem, sourceIndex int, pointerY float32) treeDropTarget {
	if g == nil || sourceIndex < 0 || sourceIndex >= len(g.heights) || len(g.heights) != len(visible) {
		return treeDropTarget{}
	}
	contentY := pointerY + float32(g.starts[sourceIndex])
	index := sort.Search(len(g.heights), func(index int) bool {
		return contentY < float32(g.starts[index]+g.heights[index])
	})
	if index >= len(visible) || contentY < float32(g.starts[index]) {
		return treeDropTarget{}
	}
	entry := visible[index]
	return treeDropTarget{
		key: entry.item.Key, drawKey: entry.item.Key, depth: entry.depth,
		position: treeDropPositionAt(contentY-float32(g.starts[index]), g.heights[index], treeItemAcceptsChildren(entry.item)), index: index,
	}
}

func treeDropIndicatorTarget(visible []flatItem, target treeDropTarget) treeDropTarget {
	if target.position != DropAfter {
		return target
	}
	index := treeVisibleIndex(visible, target.key)
	if index < 0 {
		return treeDropTarget{}
	}
	for next := index + 1; next < len(visible) && visible[next].depth > target.depth; next++ {
		target.drawKey = visible[next].item.Key
	}
	return target
}

func (g *treeDragGeometry) dropIndicatorTarget(visible []flatItem, target treeDropTarget) treeDropTarget {
	if target.position != DropAfter || target.index < 0 || target.index >= len(visible) || len(g.subtreeEnd) != len(visible) {
		return target
	}
	target.drawKey = visible[g.subtreeEnd[target.index]].item.Key
	return target
}

func treeDropPositionAt(localY float32, height int, acceptsChildren bool) DropPosition {
	if !acceptsChildren {
		if localY < float32(height)/2 {
			return DropBefore
		}
		return DropAfter
	}
	edge := float32(max(height/4, 1))
	if localY < edge {
		return DropBefore
	}
	if localY >= float32(height)-edge {
		return DropAfter
	}
	return DropInside
}

func treeItemAcceptsChildren(item Item) bool {
	return item.AcceptsChildren || treeItemExpandable(item)
}

func treeItemExpandable(item Item) bool {
	return len(item.Children) > 0 || item.ChildrenState != ChildrenLoaded
}

func treeDragViewportY(heights []int, gap int, position layout.Position, sourceIndex int, sourceY float32) float32 {
	top := float32(-position.Offset)
	if sourceIndex >= position.First {
		for index := position.First; index < sourceIndex && index < len(heights); index++ {
			top += float32(heights[index] + gap)
		}
	} else {
		for index := position.First - 1; index >= sourceIndex && index >= 0; index-- {
			top -= float32(heights[index] + gap)
		}
	}
	return top + sourceY
}

func (g *treeDragGeometry) viewportY(position layout.Position, sourceIndex int, sourceY float32) float32 {
	if g == nil || sourceIndex < 0 || sourceIndex >= len(g.heights) || position.First < 0 || position.First >= len(g.heights) {
		return sourceY
	}
	return float32(g.starts[sourceIndex]-g.starts[position.First]-position.Offset) + sourceY
}

func treeListPositionContains(position layout.Position, index int) bool {
	return index >= position.First && index < position.First+position.Count
}

func treeAutoScrollDirection(pointerY float32, viewportHeight, edge int) int {
	if viewportHeight <= 0 || edge <= 0 {
		return 0
	}
	edge = min(edge, viewportHeight/2)
	if pointerY < float32(edge) {
		return -1
	}
	if pointerY > float32(viewportHeight-edge) {
		return 1
	}
	return 0
}

func (s *treeState) updateDragScroll(gtx layout.Context, pointerY float32, viewportHeight, edge, itemCount int) {
	direction := treeAutoScrollDirection(pointerY, viewportHeight, edge)
	if direction == 0 || s.list.Position.Count >= itemCount {
		s.dragScrollAt = time.Time{}
		return
	}
	delta := time.Second / 60
	if !s.dragScrollAt.IsZero() && !gtx.Now.Before(s.dragScrollAt) {
		delta = min(gtx.Now.Sub(s.dragScrollAt), 50*time.Millisecond)
	}
	s.dragScrollAt = gtx.Now
	s.list.ScrollBy(float32(direction) * treeDragScrollSpeed * float32(delta) / float32(time.Second))
	gtx.Execute(op.InvalidateCmd{})
}

func (s *treeState) updateDragHover(gtx layout.Context, key string, eligible bool) bool {
	if !eligible || key == "" {
		s.dragHoverKey = ""
		s.dragHoverAt = time.Time{}
		s.dragHoverRequested = false
		return false
	}
	if s.dragHoverKey != key {
		s.dragHoverKey = key
		s.dragHoverAt = gtx.Now
		s.dragHoverRequested = false
	}
	if s.dragHoverRequested {
		return false
	}
	deadline := s.dragHoverAt.Add(treeDragExpandDelay)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return false
	}
	s.dragHoverRequested = true
	return true
}

func (s *treeState) resetDragAssist() {
	s.dragScrollAt = time.Time{}
	s.dragHoverKey = ""
	s.dragHoverAt = time.Time{}
	s.dragHoverRequested = false
	s.resetDragGeometry()
}

func (s *treeState) beginRename(item Item) {
	s.renameKey = item.Key
	s.renameOriginal = item.Label
	s.renameFocused = false
	s.renameEditor.SingleLine = true
	s.renameEditor.Submit = true
	s.renameEditor.SetText(item.Label)
	s.renameEditor.SetCaret(0, len([]rune(item.Label)))
}

func (s *treeState) applyRenameRequest(ctx *core.Context, tree Widget, visible []flatItem) {
	if tree.renameRequestKey == "" {
		s.renameRequestReady = false
		return
	}
	if s.renameRequestReady && s.renameRequestKey == tree.renameRequestKey && s.renameRequest == tree.renameRequest {
		return
	}
	index := treeVisibleIndex(visible, tree.renameRequestKey)
	if index < 0 {
		return
	}
	item := visible[index].item
	if !item.Renamable || tree.onRename == nil || tree.itemDisabled(item) {
		return
	}
	s.renameRequestKey = tree.renameRequestKey
	s.renameRequest = tree.renameRequest
	s.renameRequestReady = true
	s.beginRename(item)
	core.RequestFocus(ctx, &s.renameEditor)
}

func (s *treeState) updateRename(ctx *core.Context, gtx layout.Context, tree Widget) {
	if s.renameKey == "" {
		return
	}
	events := (base.TextEditor{SingleLine: true}).Update(ctx, gtx, base.TextEditorRequest{
		FieldKey: core.DerivedKey(ctx, tree.key, "rename"),
		Editor:   &s.renameEditor,
		Value:    s.renameEditor.Text(),
	})
	if !gtx.Enabled() {
		return
	}
	if events.Submitted {
		s.finishRename(tree, true)
		return
	}
	for {
		eventValue, ok := gtx.Event(key.Filter{Focus: &s.renameEditor, Name: key.NameEscape})
		if !ok {
			break
		}
		if eventValue, ok := eventValue.(key.Event); ok && eventValue.State == key.Press {
			s.finishRename(tree, false)
			return
		}
	}
	focused := gtx.Focused(&s.renameEditor)
	if focused {
		s.renameFocused = true
	} else if s.renameFocused {
		s.finishRename(tree, true)
	}
}

func (s *treeState) finishRename(tree Widget, commit bool) {
	itemKey := s.renameKey
	original := s.renameOriginal
	value := strings.TrimSpace(s.renameEditor.Text())
	s.renameKey = ""
	s.renameOriginal = ""
	s.renameFocused = false
	if commit && value != "" && value != original && tree.onRename != nil {
		tree.onRename(itemKey, value)
	}
}

func treeVisibleIndex(visible []flatItem, key string) int {
	for index, entry := range visible {
		if entry.item.Key == key {
			return index
		}
	}
	return -1
}

func treeFirstEnabledChild(visible []flatItem, tree Widget, current int) int {
	depth := visible[current].depth + 1
	for index := current + 1; index < len(visible) && visible[index].depth >= depth; index++ {
		if visible[index].depth == depth && !tree.itemDisabled(visible[index].item) {
			return index
		}
	}
	return -1
}

func treeEnabledAncestor(visible []flatItem, tree Widget, current int) int {
	parent := visible[current].parentKey
	for parent != "" {
		index := treeVisibleIndex(visible, parent)
		if index < 0 {
			return -1
		}
		if !tree.itemDisabled(visible[index].item) {
			return index
		}
		parent = visible[index].parentKey
	}
	return -1
}

func treeKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key != "" {
			set[key] = struct{}{}
		}
	}
	return set
}

func treeContainsKey(keys []string, key string) bool {
	return slices.Contains(keys, key)
}

func treeKeysEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func toggleTreeKey(keys []string, key string) []string {
	next := make([]string, 0, len(keys)+1)
	found := false
	seen := make(map[string]struct{}, len(keys))
	for _, current := range keys {
		if current == "" {
			continue
		}
		if current == key {
			found = true
			continue
		}
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		next = append(next, current)
	}
	if !found {
		next = append(next, key)
	}
	return next
}
