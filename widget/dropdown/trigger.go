package dropdown

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/menu"
)

// ButtonWidget combines a primary action button with a Dropdown trigger. It
// keeps the two controls visually joined while reusing the normal Button and
// Dropdown interaction paths.
type ButtonWidget struct {
	menuStyle      style.Style
	key            string
	action         button.Widget
	trigger        button.Widget
	dropdown       Widget
	disabled       bool
	hasDisabled    bool
	loading        bool
	defaultTrigger bool
}

// Button creates a split button with a component-owned default menu trigger.
func Button(key string, action button.Widget, items []Item) ButtonWidget {
	trigger := button.New("").Key(key + ":trigger").Content(nil).IconOnly()
	return ButtonWidget{
		key:            key,
		action:         action,
		trigger:        trigger,
		dropdown:       New(key+":dropdown", trigger, items),
		defaultTrigger: true,
	}
}

// Trigger replaces the menu trigger button.
func (b ButtonWidget) Trigger(value button.Widget) ButtonWidget {
	b.trigger = value
	b.defaultTrigger = false
	return b
}

// OnClick handles the primary action half.
func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.action = b.action.OnClick(fn)
	return b
}

// Variant applies the same button variant to both halves.
func (b ButtonWidget) Variant(value button.Variant) ButtonWidget {
	b.action = b.action.Variant(value)
	b.trigger = b.trigger.Variant(value)
	return b
}

// Size applies the same button size to both halves.
func (b ButtonWidget) Size(value button.Size) ButtonWidget {
	b.action = b.action.Size(value)
	b.trigger = b.trigger.Size(value)
	return b
}

// Disabled disables both halves.
func (b ButtonWidget) Disabled(value bool) ButtonWidget {
	b.disabled = value
	b.hasDisabled = true
	return b
}

// Loading puts the primary action into its loading state and disables the
// menu half until loading finishes.
func (b ButtonWidget) Loading(value bool) ButtonWidget {
	b.action = b.action.Loading(value)
	b.loading = value
	return b
}

// Menu replaces the popup menu configuration.
func (b ButtonWidget) Menu(value menu.Widget) ButtonWidget {
	b.dropdown = b.dropdown.Menu(value)
	return b
}

// OnOpenChangeEvent reports open-state requests with their source.
func (b ButtonWidget) OnOpenChangeEvent(fn func(OpenChangeEvent)) ButtonWidget {
	b.dropdown = b.dropdown.OnOpenChangeEvent(fn)
	return b
}

func (b ButtonWidget) Open(open bool) ButtonWidget {
	b.dropdown = b.dropdown.Open(open)
	return b
}

func (b ButtonWidget) DefaultOpen(open bool) ButtonWidget {
	b.dropdown = b.dropdown.DefaultOpen(open)
	return b
}

func (b ButtonWidget) TriggerMode(mode TriggerMode) ButtonWidget {
	b.dropdown = b.dropdown.TriggerMode(mode)
	return b
}

func (b ButtonWidget) Offset(value int) ButtonWidget {
	b.dropdown = b.dropdown.Offset(value)
	return b
}

func (b ButtonWidget) ShouldFlip(value bool) ButtonWidget {
	b.dropdown = b.dropdown.ShouldFlip(value)
	return b
}

func (b ButtonWidget) AvoidOverflow(value bool) ButtonWidget {
	b.dropdown = b.dropdown.AvoidOverflow(value)
	return b
}

func (b ButtonWidget) HoverOpenDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.HoverOpenDelay(value)
	return b
}

func (b ButtonWidget) HoverCloseDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.HoverCloseDelay(value)
	return b
}

func (b ButtonWidget) LongPressDelay(value time.Duration) ButtonWidget {
	b.dropdown = b.dropdown.LongPressDelay(value)
	return b
}

func (b ButtonWidget) Placement(value core.PopoverPlacement) ButtonWidget {
	b.dropdown = b.dropdown.Placement(value)
	return b
}

func (b ButtonWidget) MatchTriggerWidth(value bool) ButtonWidget {
	b.dropdown = b.dropdown.MatchTriggerWidth(value)
	return b
}

func (b ButtonWidget) Arrow(value bool) ButtonWidget {
	b.dropdown = b.dropdown.Arrow(value)
	return b
}

func (b ButtonWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	dropdown := b.dropdown
	dropdown.menuStyle = style.Join(dropdown.menuStyle, b.menuStyle)
	action := button.GroupStart(b.action)
	trigger := b.trigger
	if b.defaultTrigger {
		trigger = trigger.Content(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return layoutDropdownGlyph(gtx, ctx.ForegroundColor())
		}))
	}
	if b.hasDisabled {
		action = action.Disabled(b.disabled)
		trigger = trigger.Disabled(b.disabled)
	}
	if action.DisabledState() {
		trigger = trigger.Disabled(true)
	}
	if b.loading {
		trigger = trigger.Disabled(true)
	}
	trigger = button.GroupEnd(trigger)
	effectiveDisabled := b.loading || action.DisabledState() || (b.hasDisabled && b.disabled) || dropdown.disabled
	if effectiveDisabled {
		dropdown = dropdown.Disabled(true)
	}
	dropdown.trigger = trigger
	children := layoutui.Row(action, dropdown).Gap(0).AlignMiddle()
	return children.Layout(ctx, gtx)
}

// StyleMenu merges instance declarations into the popup menu.
func (b ButtonWidget) StyleMenu(values ...style.Style) ButtonWidget {
	b.menuStyle = style.Join(b.menuStyle, style.Join(values...))
	return b
}
