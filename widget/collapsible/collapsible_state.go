package collapsible

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotCollapsible = "collapsible"

type collapsibleItemState struct {
	clickable    widget.Clickable
	valueBinding binding.Binding[bool]
	expanded     bool // cached effective expanded, updated by isExpanded/requestExpanded
}

type collapsibleState struct {
	item        collapsibleItemState
	items       map[string]*collapsibleItemState
	frameItems  map[string]struct{}
	navigation  base.CollectionNavigation
	visualCache computed.VisualCache[struct{}, collapsibleVisual]
}

// collapsibleBindingConfig builds a binding.Config from the widget's expanded-state fields.
func collapsibleBindingConfig(widget Widget) binding.Config[bool] {
	return binding.Config[bool]{
		Controlled: widget.hasExpanded,
		Value:      widget.expanded,
		HasDefault: widget.hasDefaultExpanded,
		Default:    widget.defaultExpanded,
		OnChange:   widget.onExpandedChange,
	}
}

func (s *collapsibleItemState) isExpanded(widget Widget) bool {
	s.expanded = s.valueBinding.Current(collapsibleBindingConfig(widget))
	return s.expanded
}

func (s *collapsibleItemState) bind(widget Widget) {
	s.valueBinding.Bind(collapsibleBindingConfig(widget))
}

func (s *collapsibleItemState) requestExpanded(widget Widget, expanded bool) bool {
	s.expanded, _ = s.valueBinding.Request(collapsibleBindingConfig(widget), expanded)
	return s.expanded
}

func collapsibleStateFor(ctx *core.Context, key string) *collapsibleState {
	key = core.ClaimKey(ctx, core.KindCollapsible, key)
	return core.UseState[collapsibleState](ctx, key, stateSlotCollapsible)
}

func (s *collapsibleState) visual(ctx *core.Context) collapsibleVisual {
	active := core.ActiveTheme(ctx)
	return s.visualCache.Resolve(active, struct{}{}, func() collapsibleVisual {
		return buildCollapsible(active)
	})
}

func (s *collapsibleState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *collapsibleState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *collapsibleState) itemFor(key string) *collapsibleItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *collapsibleState) checkItems(items []Item) {
	validateItems(items)
	for _, item := range items {
		s.itemFor(item.Key)
	}
}

func (s *collapsibleState) updateKeys(gtx layout.Context, items []Item) string {
	target := ""
	s.navigation.Update(gtx, base.Collection{
		Count:    len(items),
		Key:      func(i int) string { return items[i].Key },
		Disabled: func(i int) bool { return items[i].Disabled },
		Clickables: func(yield func(int, *widget.Clickable) bool) {
			for i, item := range items {
				if !yield(i, &s.itemFor(item.Key).clickable) {
					return
				}
			}
		},
	}, -1, base.CollectionPolicy{Wrap: true}, func(key string) { target = key }, nil)
	return target
}
