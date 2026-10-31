package radiogroup

import (
	"fmt"
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/widget/internal/optionrow"
)

const stateSlotRadioGroup = "radio-group"

const radioSelectDuration = 200 * time.Millisecond

func radioGroupStateFor(ctx *core.Context, key string) *radioGroupState {
	key = core.ClaimKey(ctx, core.KindRadioGroup, key)
	return core.UseState[radioGroupState](ctx, key, stateSlotRadioGroup)
}

type radioGroupState struct {
	valueBinding binding.Binding[string]
	selectedKey  string
	items        map[string]*radioItemState
	frameItems   map[string]struct{}
	itemKeys     map[string]struct{}
	navigation   base.CollectionNavigation
	visualCache  radioDefaultsCache
}

// radioGroupBindingConfig builds a binding.Config from the widget's selected-key fields.
func radioGroupBindingConfig(widget Widget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *radioGroupState) currentSelectedKey(widget Widget) string {
	s.selectedKey = s.valueBinding.Current(radioGroupBindingConfig(widget))
	return s.selectedKey
}

func (s *radioGroupState) bind(widget Widget) {
	s.valueBinding.Bind(radioGroupBindingConfig(widget))
}

func (s *radioGroupState) requestSelectedKey(widget Widget, key string) string {
	s.selectedKey, _ = s.valueBinding.Request(radioGroupBindingConfig(widget), key)
	return s.selectedKey
}

func (s *radioGroupState) beginFrame() {
	core.BeginFrameMap(&s.frameItems)
}

func (s *radioGroupState) endFrame() {
	core.SweepFrameMap(s.items, s.frameItems)
}

func (s *radioGroupState) item(key string) *radioItemState {
	return core.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *radioGroupState) defaults(ctx *core.Context, variant Variant) radioDefaults {
	return resolveRadioDefaults(&s.visualCache, core.ActiveTheme(ctx), variant)
}

func (s *radioGroupState) checkItems(items []RadioItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty radio item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate radio item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *radioGroupState) updateKeys(ctx *core.Context, gtx layout.Context, items []RadioItem, selectedKey string, selectKey func(string)) {
	selected := -1
	for i, item := range items {
		if item.Key == selectedKey {
			selected = i
			break
		}
	}
	s.navigation.Update(gtx, base.Collection{
		Count:    len(items),
		Key:      func(i int) string { return items[i].Key },
		Disabled: func(i int) bool { return items[i].Disabled },
		Clickables: func(yield func(int, *widget.Clickable) bool) {
			for i, item := range items {
				if !yield(i, &s.item(item.Key).clickable) {
					return
				}
			}
		},
	}, selected, base.CollectionPolicy{Wrap: true, Axis: base.CollectionBothAxes, SelectOnFocus: true},
		func(key string) { core.RequestFocus(ctx, &s.item(key).clickable) }, selectKey)
}

type radioItemState struct {
	clickable widget.Clickable
	selected  animation.FloatTransition
	focus     animation.FocusAnimation
}

func (s *radioItemState) selection(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	return s.selectionFor(gtx, selected, radioSelectDuration, motions...)
}

func (s *radioItemState) selectionFor(gtx layout.Context, selected bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selected.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *radioItemState) focusOpacity(gtx layout.Context, focused bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return s.focus.OpacityFor(gtx, focused, duration, motions...)
}

func radioPressScale(gtx layout.Context, history []widget.Press, defaults radioDefaults, disabled bool) float32 {
	target := defaults.Metrics.PressedScale
	if target <= 0 || target > 1 {
		panic("flowui/radiogroup: pressed scale must be in (0, 1]")
	}
	return optionrow.PressScale(gtx, history, disabled, target, defaults.PressInDuration, defaults.PressOutDuration, defaults.Motion)
}
