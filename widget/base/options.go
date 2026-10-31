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

// Option is one row in an OptionList.
type Option struct {
	Key      string
	Label    string
	Disabled bool
}

// OptionState is the interaction and selection state of one option row.
// Content callbacks must treat it as read-only and must not retain it.
type OptionState struct {
	Interaction
	Key      string
	Label    string
	Index    int
	Selected bool
	// Focused reports the list's active row. The list itself owns keyboard
	// focus, so rows are not separate tab stops.
	Focused  bool
	Disabled bool
}

type optionListState struct {
	navigation CollectionNavigation
	focused    int
	focus      *widget.Clickable
}

// OptionListWidget is a headless list of selectable rows. It owns the list's
// keyboard focus, roving active row, arrow/Home/End navigation, disabled-row
// skipping and activation, and reports selection intent. It paints nothing.
type OptionListWidget struct {
	key      string
	label    string
	options  []Option
	selected map[string]bool
	disabled bool
	onSelect func(string)
	item     func(OptionState) core.Widget
}

// Options creates a headless option list with the given identity.
func Options(key string, options []Option) OptionListWidget {
	return OptionListWidget{
		key:      key,
		options:  append([]Option(nil), options...),
		selected: make(map[string]bool, len(options)),
	}
}

// Label sets the accessible name of the list itself.
func (o OptionListWidget) Label(label string) OptionListWidget {
	o.label = label
	return o
}

// Selected marks the controlled selected rows.
func (o OptionListWidget) Selected(keys ...string) OptionListWidget {
	o.selected = make(map[string]bool, len(keys))
	for _, key := range keys {
		o.selected[key] = true
	}
	return o
}

// Disabled turns the whole list off; no row can be focused or activated.
func (o OptionListWidget) Disabled(disabled bool) OptionListWidget {
	o.disabled = disabled
	return o
}

// OnSelect reports the key of an activated row. Update the selection in the
// model and pass it back through Selected.
func (o OptionListWidget) OnSelect(fn func(string)) OptionListWidget {
	o.onSelect = fn
	return o
}

// Item builds the visual for one row. It is required.
func (o OptionListWidget) Item(fn func(OptionState) core.Widget) OptionListWidget {
	o.item = fn
	return o
}

func (o OptionListWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if o.item == nil {
		panic("flowui: base.Options requires Item")
	}
	listKey := core.ClaimKey(ctx, core.KindBaseOptions, o.key)
	state := core.UseState[optionListState](ctx, listKey, "base-options")

	enabled := !o.disabled && gtx.Enabled()
	if !enabled {
		state.focused = -1
	} else if state.focused < 0 || state.focused >= len(o.options) || o.options[state.focused].Disabled {
		state.focused = o.firstSelectable()
	}
	_, focusClick := core.ClickableWithKey(ctx, core.DerivedKey(ctx, listKey, "focus"))
	if state.focus != focusClick {
		state.focus = focusClick
	}
	focused := enabled && gtx.Focused(state.focus)
	focusVisible := focused && core.FocusVisible(ctx, state.focus, focused)
	keyGtx := gtx
	if !enabled {
		keyGtx = keyGtx.Disabled()
	}
	state.navigation.Update(keyGtx, Collection{
		Count: len(o.options), Focus: state.focus,
		Key:      func(i int) string { return o.options[i].Key },
		Disabled: func(i int) bool { return o.options[i].Disabled },
	}, state.focused, CollectionPolicy{Wrap: true, Activate: true},
		func(next string) {
			for i, item := range o.options {
				if item.Key == next {
					state.focused = i
					break
				}
			}
		},
		func(next string) {
			for i, item := range o.options {
				if item.Key == next {
					o.report(state, i)
					break
				}
			}
		},
	)
	layoutCollectionFocus(ctx, keyGtx, state.focus)

	rows := make([]layoutui.Widget, 0, len(o.options))
	for index := range o.options {
		option := o.options[index]
		rowEnabled := enabled && !option.Disabled
		_, click := core.ClickableWithKey(ctx, core.DerivedKey(ctx, listKey, option.Key))
		interaction := rowInteraction(ctx, gtx, click, state.focus, !rowEnabled, func() {
			o.report(state, index)
		})
		row := OptionState{
			Interaction: Interaction{
				Hovered:      interaction.Hovered,
				Pressed:      interaction.Pressed,
				Focused:      focused && index == state.focused,
				FocusVisible: focusVisible && index == state.focused,
				Disabled:     !rowEnabled,
			},
			Key:      option.Key,
			Label:    option.Label,
			Index:    index,
			Selected: o.selected[option.Key],
			Focused:  enabled && index == state.focused,
			Disabled: !rowEnabled,
		}
		content := o.item(row)
		if content == nil {
			panic("flowui: base.Options Item returned nil")
		}
		rows = append(rows, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			draw := func(gtx layout.Context) layout.Dimensions {
				semantic.Button.Add(gtx.Ops)
				semantic.EnabledOp(rowEnabled).Add(gtx.Ops)
				semantic.SelectedOp(row.Selected).Add(gtx.Ops)
				if option.Label != "" {
					semantic.LabelOp(option.Label).Add(gtx.Ops)
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
	if o.label != "" {
		semantic.LabelOp(o.label).Add(gtx.Ops)
	}
	call.Add(gtx.Ops)
	stack.Pop()
	return dims
}

func (o OptionListWidget) report(state *optionListState, index int) {
	if !o.selectable(index) {
		return
	}
	state.focused = index
	if o.onSelect != nil {
		o.onSelect(o.options[index].Key)
	}
}

func (o OptionListWidget) selectable(index int) bool {
	return index >= 0 && index < len(o.options) && !o.options[index].Disabled
}

func (o OptionListWidget) navigationList() navigation.List {
	return navigation.List{Count: len(o.options), Disabled: func(i int) bool { return o.options[i].Disabled }}
}

func (o OptionListWidget) firstSelectable() int {
	index, _ := navigation.First(o.navigationList())
	return index
}
