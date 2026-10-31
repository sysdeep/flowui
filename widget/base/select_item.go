package base

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"

	"github.com/qianniancn/flowui/internal/core"
)

// SelectItemWidget is one option's pointer, keyboard and semantic behavior.
// Select owns the item list; this type owns a single option's interaction.
type SelectItemWidget struct {
	key         string
	label       string
	selected    bool
	highlighted bool
	disabled    bool
	onSelect    func()
	content     func(SelectItemState) core.Widget
}

// NewSelectItem creates an option with an already resolved key.
func NewSelectItem(key, label string) SelectItemWidget {
	return SelectItemWidget{key: key, label: label}
}

func (i SelectItemWidget) Selected(selected bool) SelectItemWidget {
	i.selected = selected
	return i
}
func (i SelectItemWidget) Highlighted(highlighted bool) SelectItemWidget {
	i.highlighted = highlighted
	return i
}
func (i SelectItemWidget) Disabled(disabled bool) SelectItemWidget {
	i.disabled = disabled
	return i
}
func (i SelectItemWidget) OnSelect(fn func()) SelectItemWidget {
	i.onSelect = fn
	return i
}
func (i SelectItemWidget) Content(fn func(SelectItemState) core.Widget) SelectItemWidget {
	i.content = fn
	return i
}

func (i SelectItemWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if i.content == nil {
		panic("flowui: base.SelectItem requires Content")
	}
	enabled := !i.disabled && gtx.Enabled()
	click := core.Begin(ctx, gtx, i.key, nil, enabled, true, i.onSelect)
	state := SelectItemState{
		Interaction: Interaction{
			Hovered:      enabled && click.State.Hovered,
			Pressed:      enabled && click.State.Pressed,
			Focused:      click.State.Focused,
			FocusVisible: click.State.FocusVisible,
			Disabled:     !enabled,
		},
		Key:         i.key,
		Label:       i.label,
		Selected:    i.selected,
		Highlighted: i.highlighted,
		Disabled:    i.disabled,
	}
	content := i.content(state)
	if content == nil {
		panic("flowui: base.SelectItem Content returned nil")
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.SelectedOp(i.selected).Add(gtx.Ops)
		if i.label != "" {
			semantic.LabelOp(i.label).Add(gtx.Ops)
		}
		return content.Layout(ctx, gtx)
	}, i.label)
}
