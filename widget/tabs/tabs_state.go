package tabs

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/navigation"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const (
	stateSlotTabs = "tabs"
)

type tabsState struct {
	list             layout.List
	items            map[string]*tabsItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]struct{}
	navigation       base.CollectionNavigation
	previous         widget.Clickable
	next             widget.Clickable
	add              widget.Clickable
	indicator        tabsIndicatorState
	valueBinding     binding.Binding[string]
	selectedKey      string
	selectionSet     bool
	selectionPending bool
	listLayoutSet    bool
	lastListSize     image.Point
	lastListAxis     Orientation
	lastListCount    int
	lastListWidths   []int
	retainedPanels   map[string]struct{}
	renderedPanels   map[string]struct{}
	forceRender      bool
	destroyOnHidden  bool
	editingKey       string
	panelKey         string
	panelOpacity     animation.FloatTransition
}

// tabsBindingConfig builds a binding.Config from the widget's selected-key fields.
func tabsBindingConfig(widget Widget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *tabsState) currentSelectedKey(widget Widget) string {
	s.selectedKey = s.valueBinding.Current(tabsBindingConfig(widget))
	return s.selectedKey
}

func (s *tabsState) bind(widget Widget) {
	s.valueBinding.Bind(tabsBindingConfig(widget))
}

func (s *tabsState) requestSelectedKey(widget Widget, key string) string {
	s.selectedKey, _ = s.valueBinding.Request(tabsBindingConfig(widget), key)
	return s.selectedKey
}

func (s *tabsState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *tabsState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *tabsState) item(key string) *tabsItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *tabsState) checkItems(items []TabItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty tab item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate tab item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
		s.item(item.Key)
	}
}

// normalizeSelection keeps the effective selection usable when a controlled
// value points at a removed or disabled item. The requested fallback is still
// sent through the valueBinding binding so controlled callers can update their
// model; the fallback is used optimistically for the current core.
func (s *tabsState) normalizeSelection(widget Widget, selectedKey string) string {
	index := tabsIndexByKey(widget.items, selectedKey)
	if index >= 0 && !widget.items[index].Disabled {
		return selectedKey
	}
	fallback, ok := tabsFirstEnabled(widget.items)
	if !ok {
		s.selectionPending = false
		s.selectedKey = ""
		return ""
	}
	if index >= 0 {
		fallback, ok = tabsCloseFallback(widget.items, index)
		if !ok {
			fallback, _ = tabsFirstEnabled(widget.items)
		}
	}
	effective := s.requestSelectedKey(widget, widget.items[fallback].Key)
	if widget.hasSelectedKey {
		s.selectedKey = widget.items[fallback].Key
		return s.selectedKey
	}
	return effective
}

func (s *tabsState) syncSelection(items []TabItem, selectedKey string) {
	if s.selectionSet && s.selectedKey == selectedKey {
		return
	}
	s.selectionSet = true
	s.selectedKey = selectedKey
	s.selectionPending = true
	index := tabsIndexByKey(items, selectedKey)
	if index < 0 {
		s.selectionPending = false
		return
	}
}

func (s *tabsState) ensureSelectionVisible(items []TabItem, selectedKey string) bool {
	if !s.selectionPending {
		return false
	}
	index := tabsIndexByKey(items, selectedKey)
	if index < 0 || s.list.Position.Count == 0 {
		return false
	}
	if s.selectionFullyVisible(index) {
		s.selectionPending = false
		return false
	}
	previous := s.list.Position
	s.ensureVisible(index)
	return previous.First != s.list.Position.First || previous.Offset != s.list.Position.Offset
}

// noteListLayout remembers the geometry that controls which items are visible.
// A changed geometry should re-run selection visibility once, while ordinary
// scroll-button movement must remain under the user's control.
func (s *tabsState) noteListLayout(size image.Point, axis Orientation, widths []int, count int) bool {
	changed := !s.listLayoutSet || s.lastListSize != size || s.lastListAxis != axis || s.lastListCount != count || !sameTabWidths(s.lastListWidths, widths)
	s.listLayoutSet = true
	s.lastListSize = size
	s.lastListAxis = axis
	s.lastListCount = count
	s.lastListWidths = append(s.lastListWidths[:0], widths...)
	return changed
}

func sameTabWidths(first, second []int) bool {
	if len(first) != len(second) {
		return false
	}
	for index, width := range first {
		if second[index] != width {
			return false
		}
	}
	return true
}

func (s *tabsState) selectionFullyVisible(index int) bool {
	position := s.list.Position
	if position.Count == 0 || index < position.First || index >= position.First+position.Count {
		return false
	}
	if index == position.First && position.Offset > 0 {
		return false
	}
	if position.Count == 1 && index == position.First && position.Offset == 0 {
		return true
	}
	return index != position.First+position.Count-1 || position.OffsetLast >= 0
}

func (s *tabsState) updateKeys(gtx layout.Context, items []TabItem, selectedKey string, orientation Orientation, activation ActivationMode) (string, string, bool) {
	axis := base.CollectionHorizontal
	if orientation == Vertical {
		axis = base.CollectionVertical
	}
	selectionKey, focusKey := "", ""
	s.navigation.Update(
		gtx,
		base.Collection{
			Count:    len(items),
			Key:      func(i int) string { return items[i].Key },
			Disabled: func(i int) bool { return items[i].Disabled },
			Clickables: func(yield func(int, *widget.Clickable) bool) {
				for i, item := range items {
					if !item.Disabled && !yield(i, &s.item(item.Key).clickable) {
						return
					}
				}
			},
		},
		tabsIndexByKey(items, selectedKey),
		base.CollectionPolicy{
			Wrap: true, Axis: axis, Shortcuts: true,
			SelectOnFocus: activation == ActivationAutomatic,
			Activate:      activation == ActivationManual, ActivateOnPress: true,
		},
		func(key string) { focusKey = key },
		func(key string) { selectionKey = key },
	)
	if focusKey != "" {
		s.ensureVisible(tabsIndexByKey(items, focusKey))
	}
	if selectionKey == selectedKey {
		selectionKey = ""
	}
	return selectionKey, focusKey, selectionKey != "" || focusKey != ""
}

func (s *tabsState) ensureVisible(index int) {
	position := &s.list.Position
	if position.Count == 0 {
		return
	}
	if index < position.First {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First && position.Offset > 0 {
		position.Offset = 0
		return
	}
	if index >= position.First+position.Count {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First+position.Count-1 && position.OffsetLast < 0 {
		position.Offset -= position.OffsetLast
	}
}

func (s *tabsState) scrollPage(direction, count int) {
	step := max(s.list.Position.Count-1, 1)
	first := s.list.Position.First + direction*step
	s.list.Position.First = min(max(first, 0), max(count-1, 0))
	s.list.Position.Offset = 0
}

func (s *tabsState) canScrollPrevious() bool {
	return s.list.Position.First > 0 || s.list.Position.Offset > 0
}

func (s *tabsState) canScrollNext(count int) bool {
	position := s.list.Position
	return position.First+position.Count < count || position.OffsetLast < 0
}

type tabsItemState struct {
	clickable   widget.Clickable
	close       widget.Clickable
	editor      widget.Editor
	editReady   bool
	editFocused bool
	interaction animation.FocusAnimation
	selection   animation.FloatTransition
}

func tabsStateFor(ctx *core.Context, key string) *tabsState {
	key = core.ClaimKey(ctx, core.KindTabs, key)
	return core.UseState[tabsState](ctx, key, stateSlotTabs)
}

type tabsIndicatorState struct {
	key         string
	orientation Orientation
	from        image.Rectangle
	to          image.Rectangle
	at          time.Time
	set         bool
}

func (s *tabsIndicatorState) transitionWithDuration(gtx layout.Context, key string, orientation Orientation, target image.Rectangle, duration time.Duration, motions ...theme.MotionTheme) image.Rectangle {
	if key == "" || target.Empty() {
		s.set = false
		return image.Rectangle{}
	}
	if !s.set || s.orientation != orientation {
		s.key = key
		s.orientation = orientation
		s.from = target
		s.to = target
		s.at = gtx.Now
		s.set = true
		return target
	}

	if s.key != key {
		s.from = s.currentWithDuration(gtx, duration, motions...)
		s.to = target
		s.at = gtx.Now
		s.key = key
	} else if s.to != target {
		if s.to.Size() == target.Size() {
			delta := target.Min.Sub(s.to.Min)
			s.from = s.from.Add(delta)
			s.to = target
		} else {
			s.from = target
			s.to = target
			s.at = gtx.Now
		}
	}
	return s.currentWithDuration(gtx, duration, motions...)
}

func (s *tabsIndicatorState) currentWithDuration(gtx layout.Context, duration time.Duration, motions ...theme.MotionTheme) image.Rectangle {
	if s.from == s.to {
		return s.to
	}
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if duration <= 0 {
		s.from = s.to
		return s.to
	}
	progress := animation.EaseSmoothstep(animation.Progress(gtx.Now.Sub(s.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	} else {
		s.from = s.to
		return s.to
	}
	return image.Rect(
		int(math.Round(float64(animation.LerpFloat(float32(s.from.Min.X), float32(s.to.Min.X), progress)))),
		int(math.Round(float64(animation.LerpFloat(float32(s.from.Min.Y), float32(s.to.Min.Y), progress)))),
		int(math.Round(float64(animation.LerpFloat(float32(s.from.Max.X), float32(s.to.Max.X), progress)))),
		int(math.Round(float64(animation.LerpFloat(float32(s.from.Max.Y), float32(s.to.Max.Y), progress)))),
	)
}

func (s *tabsItemState) selectionProgressWithDuration(gtx layout.Context, selected bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selection.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *tabsState) panelOpacityFor(gtx layout.Context, key string, transition PanelTransition, duration time.Duration, motion theme.MotionTheme) float32 {
	if transition != PanelFade || key == "" {
		s.panelKey = key
		s.panelOpacity.Reset()
		return 1
	}
	if s.panelKey != key {
		if s.panelKey == "" {
			s.panelOpacity.Initialize(1, gtx.Now)
		} else {
			s.panelOpacity.Set(0, 1, gtx.Now)
		}
		s.panelKey = key
	}
	return s.panelOpacity.Value(gtx, 1, duration, animation.EaseCubicOut, motion)
}

func tabsIndexByKey(items []TabItem, key string) int {
	for index, item := range items {
		if item.Key == key {
			return index
		}
	}
	return -1
}

func tabsFirstEnabled(items []TabItem) (int, bool) {
	return navigation.First(navigation.List{
		Count: len(items), Disabled: func(i int) bool { return items[i].Disabled },
	})
}

// tabsCloseFallback prefers the next enabled item and then the previous one.
// It returns false when no other enabled item remains.
func tabsCloseFallback(items []TabItem, closed int) (int, bool) {
	if closed < 0 || closed >= len(items) {
		return tabsFirstEnabled(items)
	}
	for index := closed + 1; index < len(items); index++ {
		if !items[index].Disabled {
			return index, true
		}
	}
	for index := closed - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index, true
		}
	}
	return -1, false
}
