package selector

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

// SelectWidget is a single-value Select built on base.Select.
type SelectWidget struct {
	key          string
	selected     string
	items        []SelectItem
	groups       map[int]string
	open         bool
	label        string
	description  string
	errorMessage string
	placeholder  string
	variant      SelectVariant
	disabled     bool
	invalid      bool
	fullWidth    bool
	onChange     func(string)
	onOpenChange func(bool)
	style        style.Style
}

func New(items []SelectItem) SelectWidget {
	return SelectWidget{key: core.AutoKey("select"), items: append([]SelectItem(nil), items...)}
}

func (s SelectWidget) Key(key string) SelectWidget              { s.key = key; return s }
func (s SelectWidget) SelectedKey(key string) SelectWidget      { s.selected = key; return s }
func (s SelectWidget) Open(open bool) SelectWidget              { s.open = open; return s }
func (s SelectWidget) OnChange(fn func(string)) SelectWidget    { s.onChange = fn; return s }
func (s SelectWidget) OnOpenChange(fn func(bool)) SelectWidget  { s.onOpenChange = fn; return s }
func (s SelectWidget) Placeholder(value string) SelectWidget    { s.placeholder = value; return s }
func (s SelectWidget) Label(value string) SelectWidget          { s.label = value; return s }
func (s SelectWidget) Description(value string) SelectWidget    { s.description = value; return s }
func (s SelectWidget) ErrorMessage(value string) SelectWidget   { s.errorMessage = value; return s }
func (s SelectWidget) Variant(value SelectVariant) SelectWidget { s.variant = value; return s }
func (s SelectWidget) Disabled(value bool) SelectWidget         { s.disabled = value; return s }
func (s SelectWidget) Invalid(value bool) SelectWidget          { s.invalid = value; return s }
func (s SelectWidget) FillWidth() SelectWidget                  { s.fullWidth = true; return s }

// Style merges instance declarations into the built-in select appearance. The
// declaration is resolved as the last cascade layer at layout time, so Style
// can sit anywhere in the chain.
func (s SelectWidget) Style(values ...style.Style) SelectWidget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}

func (s SelectWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	active := core.ActiveTheme(ctx)
	defaults := s.declarations(active, core.UseState[selectDeclarationCache](
		ctx,
		core.FullKey(ctx, s.key),
		stateSlotSelectDefaults,
	))
	items := make([]base.SelectItem, 0, len(s.items))
	for index, item := range s.items {
		items = append(items, base.SelectItem{Key: item.Key, Label: item.Label, Disabled: item.Disabled, Group: s.groups[index]})
	}
	triggerWidth := 0
	return base.Select(s.key, s.selected, items).
		Label(s.label).
		Gap(defaults.panelGap).
		Open(s.open).
		Disabled(s.disabled).
		OnChange(s.onChange).
		OnOpenChange(s.onOpenChange).
		Trigger(func(state base.SelectState) core.Widget {
			return s.trigger(state, defaults, &triggerWidth)
		}).
		Panel(func(state base.SelectState, content core.Widget) core.Widget {
			return s.panel(state, content, defaults, &triggerWidth)
		}).
		Item(func(state base.SelectItemState) core.Widget {
			return s.item(state, defaults)
		}).
		Group(func(title string) core.Widget {
			return s.groupTitle(ctx, title)
		}).
		Layout(ctx, gtx)
}

func (s SelectWidget) displayValue() (string, bool) {
	for _, item := range s.items {
		if item.Key == s.selected {
			return item.Label, true
		}
	}
	return s.placeholder, false
}

func NewSelectSections(sections []SelectSection) SelectWidget {
	var items []SelectItem
	groups := make(map[int]string)
	for _, section := range sections {
		if len(section.Items) > 0 {
			groups[len(items)] = section.Title
		}
		items = append(items, section.Items...)
	}
	result := New(items)
	result.groups = groups
	return result
}
