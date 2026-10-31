package combobox

import (
	"image"
	"strings"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/internal/navigation"
	"github.com/qianniancn/flowui/widget/selector"
	"github.com/qianniancn/flowui/widget/text"
)

type ComboboxItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
}

// ComboboxWidget is the base-backed single-value combobox.
type ComboboxWidget struct {
	key           string
	selected      string
	input         string
	items         []ComboboxItem
	open          bool
	hint          string
	emptyText     string
	allowCustom   bool
	disabled      bool
	invalid       bool
	fullWidth     bool
	variant       input.Variant
	size          input.Size
	style         style.Style
	onChange      func(string)
	onInputChange func(string)
	onOpenChange  func(bool)
}

type comboboxState struct {
	highlighted int
	query       string
	open        bool
	list        layout.List
	scrollbar   widget.Scrollbar
	reveal      bool
}

func New(items []ComboboxItem) ComboboxWidget {
	return ComboboxWidget{key: core.AutoKey("combobox"), items: append([]ComboboxItem(nil), items...)}
}

func (c ComboboxWidget) Key(key string) ComboboxWidget          { c.key = key; return c }
func (c ComboboxWidget) SelectedKey(key string) ComboboxWidget  { c.selected = key; return c }
func (c ComboboxWidget) InputValue(value string) ComboboxWidget { c.input = value; return c }
func (c ComboboxWidget) Open(open bool) ComboboxWidget          { c.open = open; return c }
func (c ComboboxWidget) Hint(value string) ComboboxWidget       { c.hint = value; return c }
func (c ComboboxWidget) EmptyText(value string) ComboboxWidget  { c.emptyText = value; return c }
func (c ComboboxWidget) AllowCustomValue() ComboboxWidget       { c.allowCustom = true; return c }
func (c ComboboxWidget) Disabled(value bool) ComboboxWidget     { c.disabled = value; return c }
func (c ComboboxWidget) Invalid(value bool) ComboboxWidget      { c.invalid = value; return c }
func (c ComboboxWidget) FillWidth() ComboboxWidget              { c.fullWidth = true; return c }
func (c ComboboxWidget) Variant(value input.Variant) ComboboxWidget {
	c.variant = value
	return c
}
func (c ComboboxWidget) Size(value input.Size) ComboboxWidget { c.size = value; return c }
func (c ComboboxWidget) Style(values ...style.Style) ComboboxWidget {
	c.style = style.Join(c.style, style.Join(values...))
	return c
}

func (c ComboboxWidget) OnChange(fn func(string)) ComboboxWidget      { c.onChange = fn; return c }
func (c ComboboxWidget) OnInputChange(fn func(string)) ComboboxWidget { c.onInputChange = fn; return c }
func (c ComboboxWidget) OnOpenChange(fn func(bool)) ComboboxWidget    { c.onOpenChange = fn; return c }

func (c ComboboxWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	state := core.UseState[comboboxState](ctx, core.FullKey(ctx, c.key), "combobox-keyboard")
	triggerWidth := 0
	if !c.open {
		state.highlighted = -1
	} else {
		items := c.visibleItems()
		if !state.open || state.query != c.input || state.highlighted < 0 || state.highlighted >= len(items) || items[state.highlighted].Disabled {
			state.highlighted, _ = navigation.First(comboboxNavigation(items))
			state.reveal = true
		}
	}
	state.query, state.open = c.input, c.open
	popup := base.Popup(c.key, c.open).
		Disabled(c.disabled).
		OnDismiss(func() {
			if c.onOpenChange != nil {
				c.onOpenChange(false)
			}
		}).
		Trigger(func(base.PopupState) core.Widget {
			return c.inputField(ctx, state, &triggerWidth)
		}).
		Panel(func(base.PopupState) core.Widget {
			return c.panel(ctx, state, &triggerWidth)
		})
	return popup.Layout(ctx, gtx)
}

func (c ComboboxWidget) inputField(ctx *core.Context, state *comboboxState, triggerWidth *int) core.Widget {
	return base.TextInput(c.key+"-input", c.input).
		Hint(c.hint).
		Disabled(c.disabled).
		Invalid(c.invalid).
		OnKey(func(ctx *core.Context, gtx layout.Context, tag event.Tag) {
			c.handleKeys(gtx, tag, state)
		}).
		OnChange(func(next string) {
			if c.onInputChange != nil {
				c.onInputChange(next)
			}
			if c.allowCustom && c.onChange != nil && !c.hasLabel(next) {
				c.onChange(next)
			}
			if c.onOpenChange != nil && next != "" {
				c.onOpenChange(true)
			}
		}).
		Content(func(state base.TextInputState, editor *widget.Editor) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				fieldStyle := c.resolveFieldStyle(ctx, gtx, state)
				child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					appearance := selector.DefaultAppearance(core.ActiveTheme(ctx), c.variant)
					iconColor := core.ActiveTheme(ctx).Semantic().FieldPlaceholder
					if value, ok := computed.Color(fieldStyle.icon.Text.Color); ok {
						iconColor = value
					}
					iconSize := gtx.Dp(appearance.IndicatorSize)
					indicatorWidth := gtx.Dp(appearance.IndicatorWidth)
					rotation := float32(0)
					if fieldStyle.icon.Trans != nil && fieldStyle.icon.Trans.Rotate != nil {
						rotation = *fieldStyle.icon.Trans.Rotate
					}
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layoutInputEditor(ctx, gtx, editor, c.hint, fieldStyle.content, fieldStyle.placeholder, fieldStyle.selection)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							slot := gtx
							slot.Constraints = layout.Exact(image.Pt(indicatorWidth, iconSize))
							return layout.Center.Layout(slot, func(gtx layout.Context) layout.Dimensions {
								button := base.Button(c.key + "-open").
									Label("Open options").
									Disabled(c.disabled).
									OnClick(func() {
										if c.onOpenChange != nil {
											c.onOpenChange(!c.open)
										}
									}).
									Content(func(base.ButtonState) core.Widget {
										return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
											selector.DrawChevron(gtx, iconSize, appearance.IndicatorStroke, iconColor, rotation)
											return layout.Dimensions{Size: image.Pt(iconSize, iconSize)}
										})
									})
								return button.Layout(ctx, gtx)
							})
						}),
					)
				})
				dims := layoutui.Resolved(ctx, gtx, fieldStyle.root, child)
				*triggerWidth = dims.Size.X
				return dims
			})
		})
}

type comboboxFieldStyle struct {
	root, content, placeholder, selection, icon, indicator style.ResolvedStyle
}

func (c ComboboxWidget) resolveFieldStyle(ctx *core.Context, gtx layout.Context, state base.TextInputState) comboboxFieldStyle {
	active := core.ActiveTheme(ctx)
	appearance := selector.DefaultAppearance(active, c.variant)
	styleState := style.State{
		Hovered: state.Hovered, Focused: state.Focused, FocusVisible: state.FocusVisible,
		Disabled: state.Disabled, Invalid: state.Invalid, Open: c.open,
	}
	size := style.Style{}
	switch c.size {
	case input.Small:
		size = size.Height(24).Padding(1, 6)
	case input.Large:
		size = size.Height(40).Padding(9, 12).Radius(6)
	}
	if c.fullWidth {
		size = size.FillWidth()
	}
	root := computed.Resolve(ctx, gtx, state.Key, styleState, appearance.Trigger, style.Style{}, size, c.style)
	resolve := func(declaration style.Style, part style.Part) style.ResolvedStyle {
		return computed.ResolvePart(ctx, gtx, state.Key, part, styleState, declaration, style.Style{}, style.Style{}, c.style)
	}
	return comboboxFieldStyle{
		root:        root,
		content:     root,
		placeholder: resolve(appearance.Placeholder, style.PartPlaceholder),
		selection:   resolve(style.New().Part(style.PartSelection, style.New().Background(style.Solid(active.Semantic().Selection))), style.PartSelection),
		icon:        resolve(appearance.Icon, style.PartIcon),
		indicator:   resolve(appearance.Indicator, style.PartIndicator),
	}
}

func (c ComboboxWidget) panel(ctx *core.Context, state *comboboxState, triggerWidth *int) core.Widget {
	items := c.visibleItems()
	var content core.Widget
	if len(items) == 0 {
		message := c.emptyText
		if message == "" {
			message = "No results"
		}
		content = layoutui.Box(text.New(message).Color(core.ActiveTheme(ctx).Colors.MutedForeground)).
			Style(style.New().Padding(8))
	} else {
		children := make([]layoutui.Widget, 0, len(items))
		for index, item := range items {
			children = append(children, base.NewSelectItem(c.key+"-item-"+item.Key, item.Label).
				Selected(item.Key == c.selected).
				Highlighted(index == state.highlighted).
				Disabled(item.Disabled).
				OnSelect(func() {
					state.highlighted = index
					c.selectItem(item)
				}).
				Content(func(state base.SelectItemState) core.Widget {
					appearance := selector.DefaultAppearance(core.ActiveTheme(ctx), c.variant)
					return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
						return selector.RenderItem(ctx, gtx, state, appearance, c.style)
					})
				}))
		}
		content = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			state.list.Axis = layout.Vertical
			if state.reveal {
				state.list.ScrollTo(max(state.highlighted, 0))
				state.reveal = false
			}
			return layoutui.TrackedScrollbar(ctx, gtx, &state.list, &state.scrollbar, len(children), c.disabled, true,
				func(gtx layout.Context, index int) layout.Dimensions { return children[index].Layout(ctx, gtx) })
		})
	}
	appearance := selector.DefaultAppearance(core.ActiveTheme(ctx), c.variant)
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		panelGtx := gtx
		if width := *triggerWidth; width > 0 {
			width = min(width, panelGtx.Constraints.Max.X)
			panelGtx.Constraints.Min.X = width
			panelGtx.Constraints.Max.X = width
		}
		resolved := computed.ResolvePart(ctx, panelGtx, c.key, style.PartPanel,
			style.State{Open: c.open, Disabled: c.disabled}, appearance.Panel, style.Style{}, style.Style{}, c.style)
		return layoutui.Resolved(ctx, panelGtx, resolved, content)
	})
}

func (c ComboboxWidget) handleKeys(gtx layout.Context, tag event.Tag, state *comboboxState) {
	if !c.open {
		for {
			e, ok := gtx.Event(key.Filter{Focus: tag, Name: key.NameDownArrow}, key.Filter{Focus: tag, Name: key.NameUpArrow})
			if !ok {
				return
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				state.open = true
				items := comboboxNavigation(c.visibleItems())
				if event.Name == key.NameUpArrow {
					state.highlighted, _ = navigation.Last(items)
				} else {
					state.highlighted, _ = navigation.First(items)
				}
				state.reveal = true
				if c.onOpenChange != nil {
					c.onOpenChange(true)
				}
			}
		}
	}
	for {
		e, ok := gtx.Event(
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameEscape},
		)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		items := c.visibleItems()
		state.reveal = true
		switch event.Name {
		case key.NameEscape:
			if c.onOpenChange != nil {
				c.onOpenChange(false)
			}
		case key.NameDownArrow:
			state.highlighted, _ = navigation.Move(comboboxNavigation(items), state.highlighted, 1, true)
		case key.NameUpArrow:
			state.highlighted, _ = navigation.Move(comboboxNavigation(items), state.highlighted, -1, true)
		case key.NameHome:
			state.highlighted, _ = navigation.First(comboboxNavigation(items))
		case key.NameEnd:
			state.highlighted, _ = navigation.Last(comboboxNavigation(items))
		case key.NameReturn:
			if state.highlighted >= 0 && state.highlighted < len(items) {
				item := items[state.highlighted]
				if !item.Disabled {
					c.selectItem(item)
				}
			}
		}
	}
}

func (c ComboboxWidget) selectItem(item ComboboxItem) {
	if c.onChange != nil {
		c.onChange(item.Key)
	}
	if c.onInputChange != nil {
		c.onInputChange(item.Label)
	}
	if c.onOpenChange != nil {
		c.onOpenChange(false)
	}
}

func comboboxNavigation(items []ComboboxItem) navigation.List {
	return navigation.List{Count: len(items), Disabled: func(index int) bool { return items[index].Disabled }}
}

func (c ComboboxWidget) visibleItems() []ComboboxItem {
	query := strings.TrimSpace(c.input)
	if query == "" {
		return c.items
	}
	for _, item := range c.items {
		if item.Key == c.selected && strings.EqualFold(query, item.Label) {
			return c.items
		}
	}
	visible := make([]ComboboxItem, 0, len(c.items))
	for _, item := range c.items {
		if strings.Contains(strings.ToLower(item.Label), strings.ToLower(query)) {
			visible = append(visible, item)
		}
	}
	return visible
}

func (c ComboboxWidget) hasLabel(value string) bool {
	for _, item := range c.items {
		if strings.EqualFold(item.Label, value) {
			return true
		}
	}
	return false
}
