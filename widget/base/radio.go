package base

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/navigation"
	layoutui "github.com/qianniancn/flowui/layout"
)

// RadioItem is one choice in a RadioGroup.
type RadioItem struct {
	Key      string
	Label    string
	Disabled bool
}

// RadioState is the interaction and selection state of one radio row.
// Content callbacks must treat it as read-only and must not retain it.
type RadioState struct {
	Interaction
	Key      string
	Label    string
	Index    int
	Selected bool
	Focused  bool
	Disabled bool
}

type radioGroupState struct {
	navigation CollectionNavigation
	focused    int
	focus      *widget.Clickable
}

// RadioGroupWidget is a headless single-choice group. It owns the group's
// keyboard focus, arrow/Home/End navigation (which also selects), disabled-row
// skipping and radio semantics, and reports the requested choice. It paints
// nothing.
type RadioGroupWidget struct {
	key      string
	label    string
	items    []RadioItem
	selected string
	disabled bool
	onChange func(string)
	item     func(RadioState) core.Widget
}

// RadioGroup creates a headless radio group with the given identity and
// current selection.
func RadioGroup(key, selected string, items []RadioItem) RadioGroupWidget {
	return RadioGroupWidget{key: key, selected: selected, items: append([]RadioItem(nil), items...)}
}

// Label sets the accessible name of the group.
func (r RadioGroupWidget) Label(label string) RadioGroupWidget {
	r.label = label
	return r
}

// Disabled turns the whole group off; no row can be focused or activated.
func (r RadioGroupWidget) Disabled(disabled bool) RadioGroupWidget {
	r.disabled = disabled
	return r
}

// OnChange reports the key of the newly requested choice. Update the model and
// pass the value back through the constructor.
func (r RadioGroupWidget) OnChange(fn func(string)) RadioGroupWidget {
	r.onChange = fn
	return r
}

// Item builds the visual for one row. It is required.
func (r RadioGroupWidget) Item(fn func(RadioState) core.Widget) RadioGroupWidget {
	r.item = fn
	return r
}

func (r RadioGroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if r.item == nil {
		panic("flowui: base.RadioGroup requires Item")
	}
	groupKey := core.ClaimKey(ctx, core.KindBaseRadio, r.key)
	state := core.UseState[radioGroupState](ctx, groupKey, "base-radio")
	enabled := !r.disabled && gtx.Enabled()
	if !enabled {
		state.focused = -1
	} else if state.focused < 0 || state.focused >= len(r.items) || r.items[state.focused].Disabled {
		state.focused = r.selectedIndex()
		if state.focused < 0 || r.items[state.focused].Disabled {
			state.focused = r.firstSelectable()
		}
	}

	_, focusClick := core.ClickableWithKey(ctx, core.DerivedKey(ctx, groupKey, "focus"))
	state.focus = focusClick
	focused := enabled && gtx.Focused(focusClick)
	focusVisible := focused && core.FocusVisible(ctx, focusClick, focused)
	keyGtx := gtx
	if !enabled {
		keyGtx = keyGtx.Disabled()
	}
	state.navigation.Update(keyGtx, Collection{
		Count: len(r.items), Focus: state.focus,
		Key:      func(i int) string { return r.items[i].Key },
		Disabled: func(i int) bool { return r.items[i].Disabled },
	}, state.focused, CollectionPolicy{Wrap: true, Axis: CollectionBothAxes, Activate: true, SelectOnFocus: true},
		func(next string) {
			for i, item := range r.items {
				if item.Key == next {
					state.focused = i
					break
				}
			}
		},
		func(next string) {
			for i, item := range r.items {
				if item.Key == next {
					r.report(state, i)
					break
				}
			}
		},
	)
	layoutCollectionFocus(ctx, keyGtx, state.focus)

	rows := make([]layoutui.Widget, 0, len(r.items))
	for index := range r.items {
		item := r.items[index]
		rowEnabled := enabled && !item.Disabled
		_, click := core.ClickableWithKey(ctx, core.DerivedKey(ctx, groupKey, item.Key))
		interaction := rowInteraction(ctx, gtx, click, state.focus, !rowEnabled, func() {
			r.report(state, index)
		})
		row := RadioState{
			Interaction: Interaction{
				Hovered:      interaction.Hovered,
				Pressed:      interaction.Pressed,
				Focused:      focused && index == state.focused,
				FocusVisible: focusVisible && index == state.focused,
				Disabled:     !rowEnabled,
			},
			Key:      item.Key,
			Label:    item.Label,
			Index:    index,
			Selected: item.Key == r.selected,
			Focused:  enabled && index == state.focused,
			Disabled: !rowEnabled,
		}
		content := r.item(row)
		if content == nil {
			panic("flowui: base.RadioGroup Item returned nil")
		}
		rows = append(rows, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			draw := func(gtx layout.Context) layout.Dimensions {
				semantic.RadioButton.Add(gtx.Ops)
				semantic.EnabledOp(rowEnabled).Add(gtx.Ops)
				semantic.SelectedOp(row.Selected).Add(gtx.Ops)
				if item.Label != "" {
					semantic.LabelOp(item.Label).Add(gtx.Ops)
				}
				return content.Layout(ctx, gtx)
			}
			if !rowEnabled {
				return draw(gtx.Disabled())
			}
			return click.Layout(gtx, draw)
		}))
	}

	macro := op.Record(gtx.Ops)
	dims := layoutui.Column(rows...).Layout(ctx, gtx)
	call := macro.Stop()
	stack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	if r.label != "" {
		semantic.LabelOp(r.label).Add(gtx.Ops)
	}
	call.Add(gtx.Ops)
	stack.Pop()
	return dims
}

func (r RadioGroupWidget) report(state *radioGroupState, index int) {
	if index < 0 || index >= len(r.items) || r.items[index].Disabled {
		return
	}
	state.focused = index
	if r.onChange != nil && r.items[index].Key != r.selected {
		r.onChange(r.items[index].Key)
	}
}

func (r RadioGroupWidget) selectedIndex() int {
	for index := range r.items {
		if r.items[index].Key == r.selected {
			return index
		}
	}
	return -1
}

func (r RadioGroupWidget) navigationList() navigation.List {
	return navigation.List{Count: len(r.items), Disabled: func(i int) bool { return r.items[i].Disabled }}
}

func (r RadioGroupWidget) firstSelectable() int {
	index, _ := navigation.First(r.navigationList())
	return index
}
