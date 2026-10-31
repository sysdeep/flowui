package collapsible

import (
	"fmt"
	"image"
	"slices"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

// Item describes one entry in a Group.
type Item struct {
	Key      string
	Label    string
	Leading  core.Widget
	Trailing core.Widget
	Content  core.Widget
	Disabled bool
}

// Widget presents one controlled expandable section.
type Widget struct {
	style              style.Style
	triggerStyle       style.Style
	key                string
	expanded           bool
	hasExpanded        bool
	defaultExpanded    bool
	hasDefaultExpanded bool
	label              string
	content            core.Widget
	leading            core.Widget
	trailing           core.Widget
	disabled           bool
	onExpandedChange   func(bool)
}

// New creates one expandable section. When expanded is true, it starts
// in controlled mode with the section expanded. When expanded is false, it starts
// in uncontrolled mode with the section collapsed (use DefaultExpanded(true) for
// uncontrolled mode starting expanded).
// New creates a collapsible section. It starts uncontrolled and collapsed;
// Expanded sets the controlled value and DefaultExpanded the starting one.
func New(label string, content core.Widget) Widget {
	return Widget{
		key:     core.AutoKey("collapsible"),
		label:   label,
		content: content,
	}
}

// Key sets an explicit identity for the section state.
func (w Widget) Key(key string) Widget {
	w.key = key
	return w
}

func (w Widget) Leading(leading core.Widget) Widget {
	w.leading = leading
	return w
}

func (w Widget) Trailing(trailing core.Widget) Widget {
	w.trailing = trailing
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) OnExpandedChange(fn func(bool)) Widget {
	w.onExpandedChange = fn
	return w
}

// Expanded sets the collapsible to controlled mode with the given expanded state.
func (w Widget) Expanded(expanded bool) Widget {
	w.expanded = expanded
	w.hasExpanded = true
	return w
}

// DefaultExpanded sets the initial expanded state for uncontrolled mode.
func (w Widget) DefaultExpanded(expanded bool) Widget {
	w.defaultExpanded = expanded
	w.hasDefaultExpanded = true
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithStyle(ctx, gtx)
}

// Style merges instance declarations into the section. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}

// StyleTrigger styles the internal disclosure button.
func (w Widget) StyleTrigger(values ...style.Style) Widget {
	w.triggerStyle = style.Join(w.triggerStyle, style.Join(values...))
	return w
}

// StyleWithTrigger styles the section and its disclosure button together.
func (w Widget) StyleWithTrigger(root, trigger style.Style) Widget {
	w.style = style.Join(w.style, root)
	w.triggerStyle = style.Join(w.triggerStyle, trigger)
	return w
}

// Style merges instance declarations into the group.
func (g GroupWidget) Style(values ...style.Style) GroupWidget {
	g.style = style.Join(g.style, style.Join(values...))
	return g
}

// StyleWithTrigger styles every item root and its disclosure button.
func (g GroupWidget) StyleWithTrigger(root, trigger style.Style) GroupWidget {
	g.style = style.Join(g.style, root)
	g.triggerStyle = style.Join(g.triggerStyle, trigger)
	return g
}

func (w Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	rootKey := core.FullKey(ctx, w.key)
	state := collapsibleStateFor(ctx, w.key)
	visual := state.visual(ctx)
	state.item.bind(w)
	expanded := state.item.isExpanded(w)

	disabled := w.disabled || !gtx.Enabled()
	trigger := base.Disclosure(w.key, expanded).Label(w.label).Disabled(disabled).Clickable(&state.item.clickable).
		OnToggle(func(next bool) { state.item.requestExpanded(w, next) }).Update(ctx, gtx)

	restore := core.PushKey(ctx, w.key)
	defer restore()
	focused := trigger.Focused
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:      trigger.Hovered,
		Pressed:      trigger.Pressed,
		Focused:      focused,
		FocusVisible: trigger.FocusVisible,
		Disabled:     disabled,
		Expanded:     expanded,
	}, w.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutItem(ctx, gtx, trigger, Item{
			Key:      w.key,
			Label:    w.label,
			Leading:  w.leading,
			Trailing: w.trailing,
			Content:  w.content,
			Disabled: w.disabled,
		}, expanded, disabled, visual, w.triggerStyle)
	}))
}

// GroupWidget coordinates a controlled collection of collapsible sections.
type GroupWidget struct {
	style                 style.Style
	triggerStyle          style.Style
	key                   string
	expandedKeys          []string
	items                 []Item
	allowMultipleExpanded bool
	disabled              bool
	onExpandedChange      func([]string)
}

// Group creates a controlled group that allows one expanded item by default.
func Group(key string, expandedKeys []string, items []Item) GroupWidget {
	return GroupWidget{
		key:          key,
		expandedKeys: append([]string(nil), expandedKeys...),
		items:        append([]Item(nil), items...),
	}
}

func (g GroupWidget) AllowMultipleExpanded(allow bool) GroupWidget {
	g.allowMultipleExpanded = allow
	return g
}

func (g GroupWidget) Disabled(disabled bool) GroupWidget {
	g.disabled = disabled
	return g
}

func (g GroupWidget) OnExpandedChange(fn func([]string)) GroupWidget {
	g.onExpandedChange = fn
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return g.layoutWithStyle(ctx, gtx)
}

func (g GroupWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	rootKey := core.FullKey(ctx, g.key)
	state := collapsibleStateFor(ctx, g.key)
	visual := state.visual(ctx)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(g.items)

	triggers := make([]base.DisclosureState, len(g.items))
	for index, item := range g.items {
		itemState := state.itemFor(item.Key)
		disabled := g.disabled || item.Disabled || !gtx.Enabled()
		triggers[index] = base.Disclosure(core.DerivedKey(ctx, rootKey, item.Key), slices.Contains(g.expandedKeys, item.Key)).
			Label(item.Label).
			Disabled(disabled).Clickable(&itemState.clickable).
			OnToggle(func(bool) {
				if g.onExpandedChange != nil {
					g.onExpandedChange(toggleExpandedKeys(g.expandedKeys, item.Key, g.allowMultipleExpanded))
				}
			}).Update(ctx, gtx)
	}
	if !g.disabled && gtx.Enabled() {
		if focusKey := state.updateKeys(gtx, g.items); focusKey != "" {
			core.RequestFocus(ctx, &state.itemFor(focusKey).clickable)
		}
	}

	restore := core.PushKey(ctx, g.key)
	defer restore()
	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, trigger := range triggers {
		hovered = hovered || trigger.Hovered
		pressed = pressed || trigger.Pressed
		focused = focused || trigger.Focused
		focusVisible = focusVisible || trigger.FocusVisible
	}
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     g.disabled || !gtx.Enabled(),
		Expanded:     len(g.expandedKeys) > 0,
	}, g.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return g.layout(ctx, gtx, triggers, visual)
	}))
}

func (g GroupWidget) layout(ctx *core.Context, gtx layout.Context, triggers []base.DisclosureState, visual collapsibleVisual) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	children := make([]layout.FlexChild, len(g.items))
	placements := make([]core.OverlayPlacement, len(g.items))
	dimensions := make([]layout.Dimensions, len(g.items))
	for index, item := range g.items {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := core.PushKey(ctx, item.Key)
			defer restore()
			dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				disabled := g.disabled || item.Disabled || !gtx.Enabled()
				return layoutItem(ctx, gtx, triggers[index], item, slices.Contains(g.expandedKeys, item.Key), disabled, visual, g.triggerStyle)
			})
			dimensions[index], placements[index] = dims, placement
			return dims
		})
	}
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	y := 0
	for index := range placements {
		placements[index].PlaceOffset(image.Pt(0, y))
		y += dimensions[index].Size.Y
	}
	return dims
}

func toggleExpandedKeys(keys []string, key string, allowMultiple bool) []string {
	if slices.Contains(keys, key) {
		next := make([]string, 0, len(keys)-1)
		for _, current := range keys {
			if current != key {
				next = append(next, current)
			}
		}
		return next
	}
	if !allowMultiple {
		return []string{key}
	}
	next := append([]string(nil), keys...)
	return append(next, key)
}

func validateItems(items []Item) {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty collapsible item key")
		}
		if _, ok := seen[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate collapsible item key %q", item.Key))
		}
		seen[item.Key] = struct{}{}
	}
}
