package base

import (
	giokey "gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
)

// SelectItem is one option in a single-value Select.
type SelectItem struct {
	Key      string
	Label    string
	Disabled bool
	Group    string
}

// SelectState is the interaction state a Select reports to its callbacks.
type SelectState struct {
	Interaction
	Key         string
	Open        bool
	SelectedKey string
	Highlighted int
}

// SelectItemState is the interaction state of one Select option.
type SelectItemState struct {
	Interaction
	Key         string
	Label       string
	Selected    bool
	Highlighted bool
	Disabled    bool
}

// SelectWidget is a controlled single-value select. Popup owns anchoring and
// dismissal; Select owns trigger activation, highlight and item selection.
type SelectWidget struct {
	key          string
	selected     string
	label        string
	items        []SelectItem
	open         bool
	disabled     bool
	placement    core.PopoverPlacement
	gap          int
	onChange     func(string)
	onOpenChange func(bool)
	trigger      func(SelectState) core.Widget
	panel        func(SelectState, core.Widget) core.Widget
	item         func(SelectItemState) core.Widget
	group        func(string) core.Widget
}

// Group renders the heading before the first option in a section.
func (s SelectWidget) Group(fn func(string) core.Widget) SelectWidget {
	s.group = fn
	return s
}

func Select(key, selected string, items []SelectItem) SelectWidget {
	return SelectWidget{key: key, selected: selected, items: append([]SelectItem(nil), items...)}
}

func (s SelectWidget) Label(label string) SelectWidget { s.label = label; return s }
func (s SelectWidget) Open(open bool) SelectWidget     { s.open = open; return s }
func (s SelectWidget) Disabled(disabled bool) SelectWidget {
	s.disabled = disabled
	return s
}
func (s SelectWidget) Placement(value core.PopoverPlacement) SelectWidget {
	s.placement = value
	return s
}
func (s SelectWidget) Gap(value int) SelectWidget            { s.gap = value; return s }
func (s SelectWidget) OnChange(fn func(string)) SelectWidget { s.onChange = fn; return s }
func (s SelectWidget) OnOpenChange(fn func(bool)) SelectWidget {
	s.onOpenChange = fn
	return s
}
func (s SelectWidget) Trigger(fn func(SelectState) core.Widget) SelectWidget {
	s.trigger = fn
	return s
}
func (s SelectWidget) Panel(fn func(SelectState, core.Widget) core.Widget) SelectWidget {
	s.panel = fn
	return s
}
func (s SelectWidget) Item(fn func(SelectItemState) core.Widget) SelectWidget {
	s.item = fn
	return s
}

func (s SelectWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if s.trigger == nil {
		panic("flowui: base.Select requires Trigger")
	}
	if s.item == nil {
		panic("flowui: base.Select requires Item")
	}
	key := core.ClaimKey(ctx, core.KindBaseSelect, s.key)
	state := core.UseState[selectState](ctx, key, "base-select")
	if s.open {
		if !state.open || state.highlighted < 0 {
			state.highlighted = s.selectedIndex()
			state.reveal = true
		}
	} else {
		state.highlighted = -1
	}
	state.open = s.open
	public := SelectState{
		Interaction: Interaction{Disabled: s.disabled || !gtx.Enabled()},
		Key:         key,
		Open:        s.open,
		SelectedKey: s.selected,
		Highlighted: state.highlighted,
	}
	popup := Popup(core.DerivedKey(ctx, key, "popup"), s.open).
		Placement(s.placement).
		Gap(unit.Dp(s.gap)).
		Disabled(public.Disabled).
		OnDismiss(func() {
			if s.onOpenChange != nil {
				s.onOpenChange(false)
			}
		}).
		Trigger(func(PopupState) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return s.layoutTrigger(ctx, gtx, key, state, public)
			})
		}).
		Panel(func(PopupState) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return s.layoutPanel(ctx, gtx, key, state, public)
			})
		})
	return popup.Layout(ctx, gtx)
}

func (s SelectWidget) layoutTrigger(ctx *core.Context, gtx layout.Context, key string, state *selectState, public SelectState) layout.Dimensions {
	triggerKey := core.DerivedKey(ctx, key, "trigger")
	_, clickable := core.ClickableWithKey(ctx, triggerKey)
	eventGtx := gtx
	if public.Disabled {
		eventGtx = gtx.Disabled()
	}
	s.handleKeys(eventGtx, clickable, state)
	click := core.Begin(ctx, gtx, triggerKey, clickable, !public.Disabled, true, func() {
		if s.onOpenChange != nil {
			s.onOpenChange(!s.open)
		}
	})
	core.RegisterFieldFocus(ctx, key, click.Clickable, !public.Disabled)
	if s.open && !gtx.Focused(click.Clickable) {
		core.RequestFocusVisible(ctx, click.Clickable, true)
		gtx.Execute(giokey.FocusCmd{Tag: click.Clickable})
	}
	triggerState := public
	triggerState.Hovered = !public.Disabled && click.State.Hovered
	triggerState.Pressed = !public.Disabled && click.State.Pressed
	triggerState.Focused = click.State.Focused
	triggerState.FocusVisible = click.State.FocusVisible
	content := s.trigger(triggerState)
	if content == nil {
		panic("flowui: base.Select Trigger returned nil")
	}
	return click.LayoutVisualOverflow(gtx, func(gtx layout.Context) layout.Dimensions {
		return content.Layout(ctx, gtx)
	}, s.label)
}

func (s SelectWidget) layoutPanel(ctx *core.Context, gtx layout.Context, key string, state *selectState, public SelectState) layout.Dimensions {
	items := make([]layoutui.Widget, 0, len(s.items))
	for index := range s.items {
		items = append(items, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			if s.items[index].Group != "" && s.group != nil {
				return layoutui.Column(s.group(s.items[index].Group), core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					return s.layoutItem(ctx, gtx, key, state, public, index)
				})).Layout(ctx, gtx)
			}
			return s.layoutItem(ctx, gtx, key, state, public, index)
		}))
	}
	var content core.Widget = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		state.list.Axis = layout.Vertical
		if state.reveal {
			state.list.ScrollTo(max(state.highlighted, 0))
			state.reveal = false
		}
		return layoutui.TrackedScrollbar(ctx, gtx, &state.list, &state.scrollbar, len(items), public.Disabled, true,
			func(gtx layout.Context, index int) layout.Dimensions { return items[index].Layout(ctx, gtx) })
	})
	if s.panel != nil {
		content = s.panel(public, content)
	}
	return content.Layout(ctx, gtx)
}

func (s SelectWidget) layoutItem(ctx *core.Context, gtx layout.Context, owner string, state *selectState, public SelectState, index int) layout.Dimensions {
	item := s.items[index]
	itemKey := core.DerivedKey(ctx, owner, item.Key)
	return NewSelectItem(itemKey, item.Label).
		Selected(item.Key == s.selected).
		Highlighted(index == state.highlighted).
		Disabled(item.Disabled || public.Disabled).
		OnSelect(func() {
			state.highlighted = index
			if s.onChange != nil {
				s.onChange(item.Key)
			}
			if s.onOpenChange != nil {
				s.onOpenChange(false)
			}
		}).
		Content(s.item).
		Layout(ctx, gtx)
}
func (s SelectWidget) handleKeys(gtx layout.Context, tag *widget.Clickable, state *selectState) {
	if !s.open {
		for {
			e, ok := gtx.Event(
				giokey.Filter{Focus: tag, Name: giokey.NameReturn},
				giokey.Filter{Focus: tag, Name: giokey.NameEnter},
				giokey.Filter{Focus: tag, Name: giokey.NameDownArrow},
				giokey.Filter{Focus: tag, Name: giokey.NameUpArrow},
			)
			if !ok {
				return
			}
			if press, ok := e.(giokey.Event); ok && press.State == giokey.Press && s.onOpenChange != nil {
				s.onOpenChange(true)
			}
		}
	}
	current := state.highlighted
	if current < 0 {
		current = s.selectedIndex()
	}
	rows := Collection{
		Count: len(s.items), Focus: tag,
		Key:      func(index int) string { return s.items[index].Key },
		Disabled: func(index int) bool { return s.items[index].Disabled },
	}
	state.navigation.update(gtx, rows, current,
		CollectionPolicy{Wrap: true, Commands: []giokey.Name{giokey.NameReturn}},
		collectionActions{
			escape: func() {
				if s.open && s.onOpenChange != nil {
					s.onOpenChange(false)
				}
			},
			event: func(event CollectionEvent) int {
				if event.Kind == CollectionFocus {
					state.highlighted = event.Index
					state.reveal = true
				} else if event.Kind == CollectionCommand && s.open {
					s.selectHighlighted(state)
				}
				return event.Index
			},
		})
}

func (s SelectWidget) selectedIndex() int {
	for index, item := range s.items {
		if item.Key == s.selected && !item.Disabled {
			return index
		}
	}
	return 0
}

type selectState struct {
	open        bool
	highlighted int
	navigation  CollectionNavigation
	list        layout.List
	scrollbar   widget.Scrollbar
	reveal      bool
}

func (s SelectWidget) selectHighlighted(state *selectState) {
	if state.highlighted < 0 || state.highlighted >= len(s.items) {
		return
	}
	item := s.items[state.highlighted]
	if item.Disabled {
		return
	}
	if s.onChange != nil {
		s.onChange(item.Key)
	}
	if s.onOpenChange != nil {
		s.onOpenChange(false)
	}
}
