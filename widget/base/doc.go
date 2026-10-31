// Package base provides headless controls: pointer, keyboard, focus and
// accessibility behavior without a visual design.
//
// A base owns interaction and reports intent; the application keeps the value
// in its model and draws the control in a Content callback. The official
// widgets add a theme look on the same interaction primitives, and custom
// design systems can build on them instead of reimplementing the behavior:
//
//	base.Button("save").
//		Label("Save").
//		OnClick(func() { send(Save{}) }).
//		Content(func(state base.ButtonState) ui.Widget {
//			return card(state.Hovered, state.Pressed, state.FocusVisible)
//		})
//
// Unlike the widgets in sibling packages, a base draws nothing by itself:
// Content is required. State reports interaction bits so the Content callback
// can paint hover, pressed, focus and disabled looks without rebuilding the
// behavior. Content draws the whole control, including its interaction box.
// Slider and RangeSlider use separate Header and Track callbacks so the header
// stays outside the drag area.
//
// Collection controls follow the listbox pattern rather than giving every row
// its own tab stop: Options and RadioGroup own one keyboard focus target for
// the whole group, move a roving active row with the arrow keys, skip disabled
// rows, and report the activated key. Disclosure pairs a trigger with an
// optional expandable region.
//
// Compound widgets can reuse CollectionNavigation and RowInteraction while
// retaining row focus targets and virtual layout. Axis and activation policies
// cover sidebar and tab collections. Targets accepts Toolbar child focus tags;
// UpdateWithEvents retains modifiers and orders Table/Tree domain commands with
// shared navigation. UpdateRow retains click details and child-cell filtering.
// Tree and Tabs rename editors use TextEditor; hierarchy, loading, selection and drag
// policies remain in the components. MenuNavigation adds submenu/root-menu
// intent without owning popup layout; MenubarNavigation shares root movement
// and typeahead while the component keeps global accelerators. TextEditor.Update shares
// editor event handling without prescribing a field's visuals or context menu;
// FieldInteraction shares shell hover and pointer-to-editor focus routing,
// including compound-field action exclusions, without owning any appearance.
// Disclosure.Update returns a frame-local trigger host that separates behavior
// from animated content lifetime. TextInput Content receives a Gio editor to
// lay out with caller-supplied typography and paint; base chooses neither.
// PopupInteraction shares outside/Escape dismissal and pointer-only popup
// surfaces with Popover, Dropdown, menus and ColorPicker. PopupTrigger shares
// secondary-click, Shift+F10 and cancellable long-press input. Components supply
// animated hit geometry and retain open binding and focus-return policy.
// DatePicker reuses panel input and RowInteractionWithFocus while retaining
// date/range rules and focus-loss dismissal. HoverInteraction shares pass-through
// field and Toast hover; EscapePressed reports target-local Escape without
// owning dismissal or focus. Toast retains timers, queue and expansion policy.
// TooltipInteraction handles passive
// hover/focus visibility deadlines; Tooltip keeps coordination and animation.
// DragInteraction shares pointer capture with ColorPicker and TextSelection;
// SliderKeys shares numeric keyboard input without prescribing value geometry.
// SelectableText adapts Gio selection and queued actions, while TextSelection
// supports custom renderers with caller-owned hit testing and selection paint.
// Pagination composes Button for page controls. Auxiliary Tabs and Toast actions
// use UpdateRow with PreserveFocus to retain the current editor or owner focus.
// DialogInteraction composes those surfaces with modal focus boundaries,
// initial/deferred focus and stacked-scope recovery. Modal supplies the visible
// lifetime and overlay Tail; AlertDialog composes the same Modal behavior.
package base
