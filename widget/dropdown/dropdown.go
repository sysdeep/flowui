package dropdown

import (
	"github.com/qianniancn/flowui/style"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/menu"
)

type SelectionMode = menu.SelectionMode
type ItemKind = menu.ItemKind
type ItemVariant = menu.ItemVariant
type IndicatorType = menu.IndicatorType
type Item = menu.Item
type Section = menu.Section
type ActionEvent = menu.ActionEvent

const (
	ItemAction     = menu.ItemAction
	ItemCheckbox   = menu.ItemCheckbox
	ItemRadio      = menu.ItemRadio
	ItemSubmenu    = menu.ItemSubmenu
	ItemSeparator  = menu.ItemSeparator
	ItemGroupLabel = menu.ItemGroupLabel

	SelectionNone     = menu.SelectionNone
	SelectionSingle   = menu.SelectionSingle
	SelectionMultiple = menu.SelectionMultiple

	ItemDefault = menu.ItemDefault
	ItemDanger  = menu.ItemDanger

	IndicatorNone      = menu.IndicatorNone
	IndicatorCheckmark = menu.IndicatorCheckmark
	IndicatorDot       = menu.IndicatorDot
)

type TriggerMode uint8

const (
	TriggerPress TriggerMode = iota
	TriggerLongPress
	TriggerHover
	TriggerContextMenu
)

// OpenChangeSource identifies the interaction that requested a Dropdown state
// change.
type OpenChangeSource uint8

const (
	OpenChangeProgrammatic OpenChangeSource = iota
	OpenChangeTrigger
	OpenChangeMenu
	OpenChangeOutside
	OpenChangeKeyboard
	OpenChangePeer
	OpenChangeContextMenu
)

// OpenChangeEvent contains the state transition and its interaction source.
type OpenChangeEvent struct {
	Open   bool
	Source OpenChangeSource
}

type Widget struct {
	style              style.Style
	menuStyle          style.Style
	key                string
	trigger            core.Widget
	menu               menu.Widget
	open               bool
	hasOpen            bool
	defaultOpen        bool
	hasDefaultOpen     bool
	triggerMode        TriggerMode
	placement          core.PopoverPlacement
	offset             unit.Dp
	hasOffset          bool
	shouldFlip         bool
	hasShouldFlip      bool
	avoidOverflow      bool
	hasAvoidOverflow   bool
	matchTriggerWidth  bool
	arrow              bool
	hoverOpenDelay     time.Duration
	hasHoverOpenDelay  bool
	hoverCloseDelay    time.Duration
	hasHoverCloseDelay bool
	longPressDelay     time.Duration
	hasLongPressDelay  bool
	disabled           bool
	onOpenChangeEvent  func(OpenChangeEvent)
}

func New(key string, trigger core.Widget, items []Item) Widget {
	return Widget{
		key:       key,
		trigger:   trigger,
		menu:      menu.New(key+":menu", items),
		placement: core.PopoverBottomStart,
	}
}

func NewSections(key string, trigger core.Widget, sections []Section) Widget {
	return Widget{
		key:       key,
		trigger:   trigger,
		menu:      menu.Sections(key+":menu", sections).AutoSeparateSections(false),
		placement: core.PopoverBottomStart,
	}
}

// Menu replaces the popup menu configuration. Configure selection, actions,
// sizing, and content on menu.Widget instead of mirroring the Menu API here.
func (d Widget) Menu(value menu.Widget) Widget {
	d.menu = value
	return d
}

func Separator() Item {
	return menu.Separator()
}

// GroupLabel returns a non-selectable heading for a dropdown section.
func GroupLabel(label string) Item {
	return menu.GroupLabel(label)
}

func (d Widget) Open(open bool) Widget {
	d.open = open
	d.hasOpen = true
	return d
}

func (d Widget) DefaultOpen(open bool) Widget {
	d.defaultOpen = open
	d.hasDefaultOpen = true
	return d
}

// OnOpenChangeEvent reports state changes together with their source.
func (d Widget) OnOpenChangeEvent(fn func(OpenChangeEvent)) Widget {
	d.onOpenChangeEvent = fn
	return d
}

func (d Widget) TriggerMode(mode TriggerMode) Widget {
	d.triggerMode = mode
	return d
}

// HoverOpenDelay sets the delay before a hover-triggered popup opens.
func (d Widget) HoverOpenDelay(delay time.Duration) Widget {
	d.hoverOpenDelay = max(delay, 0)
	d.hasHoverOpenDelay = true
	return d
}

// HoverCloseDelay sets the delay before a hover-triggered popup closes.
func (d Widget) HoverCloseDelay(delay time.Duration) Widget {
	d.hoverCloseDelay = max(delay, 0)
	d.hasHoverCloseDelay = true
	return d
}

// LongPressDelay sets the delay before a long-press-triggered popup opens.
func (d Widget) LongPressDelay(delay time.Duration) Widget {
	d.longPressDelay = max(delay, 0)
	d.hasLongPressDelay = true
	return d
}

func (d Widget) Placement(placement core.PopoverPlacement) Widget {
	d.placement = placement
	return d
}

func (d Widget) Offset(dp int) Widget {
	d.offset = unit.Dp(max(dp, 0))
	d.hasOffset = true
	return d
}

func (d Widget) ShouldFlip(shouldFlip bool) Widget {
	d.shouldFlip = shouldFlip
	d.hasShouldFlip = true
	return d
}

func (d Widget) AvoidOverflow(avoid bool) Widget {
	d.avoidOverflow = avoid
	d.hasAvoidOverflow = true
	return d
}

func (d Widget) Disabled(disabled bool) Widget {
	d.disabled = disabled
	d.menu = d.menu.Disabled(disabled)
	return d
}

// MatchTriggerWidth makes the popup at least as wide as its trigger.
func (d Widget) MatchTriggerWidth(match bool) Widget {
	d.matchTriggerWidth = match
	return d
}

// Arrow controls whether the popup draws an anchor arrow.
func (d Widget) Arrow(show bool) Widget {
	d.arrow = show
	return d
}

func (d Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return d.layoutRoot(ctx, gtx)
}

func (d Widget) flipEnabled() bool {
	return !d.hasShouldFlip || d.shouldFlip
}

func (d Widget) overflowAvoidanceEnabled() bool {
	return !d.hasAvoidOverflow || d.avoidOverflow
}

// Style merges instance declarations into the trigger. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (d Widget) Style(values ...style.Style) Widget {
	d.style = style.Join(d.style, style.Join(values...))
	return d
}

// StyleMenu merges instance declarations into the popup menu.
func (d Widget) StyleMenu(values ...style.Style) Widget {
	d.menuStyle = style.Join(d.menuStyle, style.Join(values...))
	return d
}

// StyleWithMenu sets the trigger and popup declarations together.
func (d Widget) StyleWithMenu(trigger, menu style.Style) Widget {
	return d.Style(trigger).StyleMenu(menu)
}
