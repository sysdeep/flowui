# FlowUI base

Headless controls: pointer, keyboard, focus and accessibility without a visual
design. Use them when a custom design needs a different **structure**, not only
different colors.

```go
base.Toggle("sync", model.Sync).
	Label("Sync files").
	OnChange(func(next bool) { send(SetSync{next}) }).
	Content(func(state base.ToggleState) ui.Widget {
		return row(indicator(state.Checked, state.Hovered), text.New("Sync files"))
	})
```

## Available

| Base | Value | Reports |
| --- | --- | --- |
| `Button(key)` | none | click, Space or Enter; `OnClick` |
| `Toggle(key, checked)` | `bool` | click, Space or Enter; `OnChange` |
| `Slider(key, value, min, max)` | `float64` | drag, arrows, Home/End, Page Up/Down; `OnChange`; `Header`/`Track` |
| `RangeSlider(key, lower, upper, min, max)` | two `float64` | pointer/keyboard; `OnRangeChange`; `Header`/`Track` |
| `Options(key, items)` | selected keys | arrows/Home/End, disabled-row skipping; `OnSelect` |
| `RadioGroup(key, selected, items)` | `string` | arrows/Home/End select; `OnChange` |
| `Disclosure(key, expanded)` | `bool` | trigger activation; `OnToggle` |
| `TextInput(key, value)` | `string` | editor focus, IME, selection; `OnChange` |
| `Popup(key, open)` | none | overlay anchor, placement, outside press and Escape dismissal |
| `Select(key, selected, items)` | `string` | trigger open/close, highlight, item selection |

All expose an interaction state to their callbacks:

- `Interaction`: `Hovered`, `Pressed`, `Focused`, `FocusVisible`, `Disabled`,
  `Invalid`, `ReadOnly`, `Loading`.
- `ButtonState`: `Interaction`, `Key`, `Selected`.
- `ToggleState`: `Interaction`, `Key`, `Checked`, `Indeterminate`.
- `SliderState`: `Interaction`, `Key`, `Value`, `Min`, `Max`, `Ratio`,
  `Orientation`, `Dragging`.
- RangeSliderState: Interaction, Key, Lower, Upper, Min, Max,
  `LowerRatio`, `UpperRatio`, `Orientation`, `Dragging`, `Active`.

## Contract

- **Content is required.** A base paints nothing by itself; it panics when
  `Content`/`Track` is missing or returns nil.
- **Content draws the whole control.** The base measures Content and uses that
  size as its interaction and semantic area. Padding, background, radius and
  parts are the visual layer's job; the base has no `Style`.
- **The model owns the value.** `OnChange` reports the requested value; the
  application writes it in `Update` and builds the next frame with the new
  value. There is no pointer binding, unlike an immediate-mode toolkit.
- **Keys are identities.** Give each base a stable key. A key is claimed with
  `KindBaseToggle` / `KindBaseSlider` / `KindBaseRangeSlider`, so it cannot
  collide with an official widget's key.
- **Semantics follow the role.** `Toggle` reports a check box by default,
  `Role(ToggleSwitch)` a switch, `Role(ToggleButton)` a pressed button. `Slider`
  reports its label and description.
- **Slider geometry stays in the visual layer.** `HitCross` reserves a larger
  pointer target than the painted track, `Inset` keeps the value mapping inside
  the thumb radius, and `Cursor` resolves the cursor over the whole hit area.

`Popup` is the overlay foundation: it records the trigger anchor, registers a
popup-layer panel, places it with the shared placement rules, and dismisses on
an outside pointer press or Escape. Its panel and outside blockers are
pointer-only and do not introduce Tab stops. `Select` combines Popup with trigger and item
callbacks. Its highlight movement uses `CollectionNavigation` with disabled-row
skipping, wrapping and Home/End while focus stays on the trigger.
`widget/selector` builds its single-value select on this base; the
editable `widget/combobox` composes the same popup, text input, and listbox
primitives.

## Relation to the styled widgets

`widget/button` uses `base.Button` for standard buttons, groups and close
buttons. `LayoutWithClickable` / `LayoutWithClickableNoEvents` remain as thin
base adapters for composites that own the clickable. `button.Visual` draws
only the appearance from a supplied `base.ButtonState`, without a second
interaction or semantic host.
`widget/toggle` uses `base.Toggle` for Checkbox and Switch, and `base.Button`
for ToggleButton; Segmented composes those ToggleButtons. `widget/slider` uses
`base.Slider` for single-value sliders and `base.RangeSlider` for range sliders.
Pagination composes `base.Button` for page and previous/next controls, retaining
page calculation and default rendering in `widget/navigation`.
Interactive widgets reuse these bases for pointer, keyboard, focus, and
semantics while keeping domain layout and visual recipes in their own packages.

`widget/radiogroup`, `widget/listbox`, `widget/sidebar`, `widget/tabs`,
`widget/table` and `widget/tree` use
`CollectionNavigation` and `RowInteraction`. `widget/menu` uses
`MenuNavigation` and `RowInteraction`. `widget/collapsible` uses `Disclosure.Update` for its triggers
and `CollectionNavigation` for its group. `widget/input` uses `TextEditor` for
Input, TextArea and their InputGroup variants, and `Button` for GroupAction.

These lower-level behavior adapters let a styled component keep its existing
layout and retained state. `CollectionNavigation` processes the materialized
rows' keyboard targets and navigates the complete indexed collection; the
component materializes and focuses an off-screen destination. `RowInteraction`
handles pointer activation and focus while the component supplies the hit
region and row semantics. `TextEditor.Update` registers editor focus, applies
queued edits, dispatches change before submit, and synchronizes controlled text.
`Disclosure.Update` returns a frame-local state whose `LayoutTrigger` mounts
its pointer target and semantics without processing events twice. Collapsible
uses that host with `button.Visual`; content lifetime and exit animations stay
in the component.

The standalone `Options` and `RadioGroup` use one keyboard focus target for the
whole group. Styled RadioGroup and virtualized ListBox retain their row focus
targets through the lower-level adapters. Both use the same
`CollectionNavigation` event loop and row activation implementation. Set
`Collection.Focus` for a single group target, or supply `Clickables` for retained
row targets. The indexed algorithms live in `widget/internal/navigation`;
`widget/navigation` preserves the public helper API. Generic controlled and
uncontrolled value binding and visual row animation helpers live in
`widget/internal`; neither owns clickables. Virtualization,
field menus, animation and domain semantics remain in their component packages.

## Collection policies

`CollectionPolicy.Axis` selects vertical, horizontal or both-axis arrow keys;
vertical is the zero value. This replaces the former `Horizontal` flag.
`SelectOnFocus` enables automatic selection, while `Activate` enables explicit
Enter/Space activation. `ActivateOnPress` preserves Tabs' manual key-down
activation; other collections activate on matching key release. `Shortcuts`
enables Ctrl/Cmd+PageUp/PageDown and Ctrl/Cmd+Tab, including Shift+Tab reversal.

`Collection.Targets` accepts arbitrary retained focus tags for Toolbar's child
controls. Toolbar supplies orientation and wrapping; children keep activation.
CommandPalette uses the same navigation with its search editor as the target;
`CollectionPolicy.PreserveHomeEnd` leaves caret keys with the editor while
arrow keys move the highlighted result without moving keyboard focus.
`Collection.Fallback` resolves an active index only when keyboard input needs
it, so an idle virtual Table does not scan off-screen data for its selection.
`UpdateWithEvents` reports focus, activation and component commands in input
order. `CollectionPolicy.Modifiers` permits range-selection keys, and activation
retains key-down modifiers until release. The callback returns the active index
so Tree's parent/child commands and subsequent arrow keys share one position.

`UpdateRow` accepts a `RowInput` when a component needs click modifiers,
double-click counts or interactive-cell filtering. `RowInput.PreserveFocus`
keeps the current focus for auxiliary actions such as Tabs close buttons and
Toast close/action buttons. Table keeps sorting, range
selection, column resizing and virtualization; Tree keeps hierarchy, lazy
loading, checking, drag/drop and rename commit rules. Tree's rename editor uses
`TextEditor.Update`; Tabs uses the same adapter for inline rename. Their Escape,
blur and commit policies remain in their components. Neither
component moves its visual recipe or domain model into base.

Sidebar keeps its nested open-key model and hover expansion in the component.
Tabs keeps scroll positioning, panel lifetime, close/rename controls and drag
extensions in its component. Both delegate collection keyboard and row-pointer
interaction to base rather than retaining private event loops.

`MenuNavigation` shares the same collection event loop for wrapping navigation,
disabled skipping, typeahead and activation. It adds `MenuPolicy`/`MenuActions`
for right-arrow submenu opening or next-root intent and left-arrow parent or
previous-root intent. Menu retains submenu placement, hover timing, selection
binding and action/close callbacks. `MenubarNavigation` shares collection movement,
Home/End and typeahead, and reports perpendicular-arrow panel-opening intent.
Menubar retains global F10/Alt accelerators and switching between open panels;
its root pointer activation uses `RowInteraction`.

## Popup interaction

`PopupInteraction` is the retained, lower-level behavior shared by `Popup`,
Popover, Dropdown, Submenu, ContextMenu, Menubar and ColorPicker. `Update`
accepts `PopupPolicy` and returns at most one
outside/Escape close intent; it does not mutate the component's open binding.
Only the overlay host's policy-input owner may accept close events. Pointer-only
panel, arrow and outside regions preserve focus without adding keyboard targets.
`LayoutOutside`, `LayoutSurface` and `LayoutArrow` accept component-supplied
geometry, including animated bounds and protected trigger regions.

Popover retains completed-click outside dismissal and its independent keyboard
dismissal switch. Dropdown retains immediate outside-press dismissal, reasonful
open-change callbacks, exclusivity, menu focus handoff and trigger-focus return.
Its ordinary trigger activation uses `RowInteraction`; component-specific
opening keys are processed before the clickable consumes keyboard events.
Submenu, ContextMenu and Menubar retain their parent/root focus-return and
selection policies. ColorPicker retains color conversion, history commits and
controlled value binding. Open binding, placement, arrows and enter/exit
animation remain in the components.

`PopupTrigger` shares secondary-click, Shift+F10 and cancellable long-press
input between ContextMenu and Dropdown. It observes the existing trigger
without replacing its primary clickable or consuming external focus events.
The component supplies gesture policy and owns anchor geometry, opening
callbacks and focus handoff.

DatePicker shares popup panel input through `PopupInteraction.SurfaceFocus` and
calendar row activation through `RowInteractionWithFocus`, which routes pointer
focus to its retained field trigger. It keeps date arithmetic, segment editing,
disabled dates, range selection/cancellation and focus-loss dismissal locally.
Tooltip composes `TooltipInteraction` for passive hover/focus input and show/close
deadlines. `TooltipPolicy.Visible` lets the visual layer report an unfinished exit
animation so re-entry can open immediately. The component retains exclusive
visibility coordination, placement and animation; base adds no popup dismissal,
pointer blocker or Tab stop.

`HoverInteraction` observes a caller-owned target's Enter/Leave/Cancel events
without consuming presses or changing focus. Call `Update` before layout and
`AddInput` with the measured bounds; a nil cursor preserves the child's cursor.
DateField, DatePicker and DateRangePicker use it for field-shell hover. Toast
uses it for item and stack hover, retaining timer pause/resume, touch policy,
queue lifetime and expansion animation in its own package.

`EscapePressed` reports Escape on an existing focus target, not globally. Date
controls and Toast keep their own close/cancel policies and focus management.
Neither helper introduces a new Tab stop or supplies a default appearance.

Chart pointer positions, clicks and data-window gestures are shared in
`widget/chart`, not base. PieChart and Heatmap reuse those helpers while retaining
sector/cell hit testing, selection and tooltips in their component packages.

## Dialog interaction

`DialogInteraction` composes the same pointer-only surfaces with modal focus
behavior. Modal uses it directly; AlertDialog composes Modal rather than owning
a second state machine. `Update` accepts `DialogPolicy` and returns one close
request from a completed outside click, Escape or the caller-owned close button.
Dismissal switches remain independent. A controlled dialog is not internally
closed just because a close request was reported.

`LayoutStart` mounts a leading boundary and neutral initial focus target before
the content. The component supplies the backward-wrap destination when controls
are enabled. `LayoutEnd` runs as the overlay host's Tail, after nested popups,
so Tab cannot escape into the background. `ManageFocus` runs once per visible
frame: initial focus is deferred while disabled, an underlying dialog regains
its focus scope after a stacked dialog exits, and an exiting dialog retains
neutral focus until the component removes its overlay. Focus timing is based on
open state and overlay ownership, not animation progress. This does not add
restoration to an arbitrary application trigger after the last dialog closes.

Open binding, lazy state lifetime, backdrop drawing, animated geometry, scrolling,
content sections and the default visual remain in Modal. AlertDialog retains
its explicit-action defaults (outside and Escape dismissal disabled), header and
status visuals, and forwards controlled/uncontrolled state without forcing a
controlled value.

## Editor contract

`FieldInteraction` retains field-shell hover and routes pointer presses to a
caller-owned focus target. Call `Update` (or `UpdateWithFocus` to exclude action
slots) before layout, then `AddPointer` with the final shell size. Its hit area
passes through to child controls; it supplies no colors, variants or geometry.
Input, TextArea and InputGroup share this behavior.

`TextEditor.Update(ctx, gtx, TextEditorRequest{...})` accepts a retained Gio
editor, field identity, current value, optional queued edits and event callbacks.
Pass `gtx.Disabled()` to disable editing and focus registration; there is no
second disabled flag and Update returns events rather than a modified context.
Input and TextArea apply this behavior to their existing retained editor.

`TextInput.Content` receives `(TextInputState, *widget.Editor)`. The visual must
lay out that editor with its own shaper, font, text size, text material and
selection material. Base provides no default typography or selection color.
This replaces the previous ready-painted `core.Widget` callback argument and
removes `TextInput.TextSize`. Combobox supplies its editor paint locally.

## Tests and example

- `button_test.go`, `toggle_test.go`, `slider_test.go`: pointer, keyboard,
  disabled/read-only, semantics.
- [`examples/base_controls`](../../examples/base_controls): a settings window
  with a custom row and a gauge drawn from `SliderState.Ratio`.

## Split-pane interaction

`SplitPaneInteraction` is a retained input helper, like `FieldInteraction`.
Call `Update` with `SplitPaneGeometry` in physical pixels, then `AddHandle` with
a caller-owned hit rectangle, label and cursor. Retain the helper's address as
the focus identity. It reports the requested ratio, hover and drag state, and
handles pointer capture plus arrow/Home/End/Page keys. It has no default visual,
Style, theme or focus animation; `widget/splitpane` supplies those separately.

## Dragging and numeric axes

`DragInteraction.Update` captures one primary pointer, reports positions and a
`started` flag, and cancels dragging when disabled. `AddInput` mounts the caller's
rectangle, cursor and semantic text. The target must have a stable retained
identity. ColorPicker uses it for the area, wheel and channel axes; color mapping,
HSV conversion, history and focus animation stay in ColorPicker.

`SliderKeys` shares arrows, Home/End and Page keys between Slider and color axes.
Supply a positive `Step`; `Snap` rounds to that increment and optional
`ShiftScale` enables accelerated arrow steps. `Update` returns the next changed
value, so callers may repeat it to process multiple queued keys in order.

## Read-only text selection

`SelectableText` owns a Gio `Selectable`, queued `CopySelection`/`SelectAll`
actions and pass-through press observation. Call `Update` before laying out
`Selectable` with caller-supplied typography and paint, then `AddPointer` with
the measured size. It has no context-menu dependency or default selection color.

`TextSelection` supports custom text renderers such as Code: it retains a rune
range, uses `DragInteraction` and copies only the focused selection. The caller
registers the input target and supplies a position-to-rune callback; text shaping,
wrapping, gutter offsets and selection drawing remain in the component.
