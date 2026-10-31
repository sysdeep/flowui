package base

import (
	"iter"
	"slices"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/navigation"
)

// Collection describes indexed rows without requiring the base to own their
// storage or layout. Clickables visits only materialized rows, paired with
// their indices in the full collection. Navigation can target any enabled row.
type Collection struct {
	Count    int
	Key      func(int) string
	Label    func(int) string
	Disabled func(int) bool
	// Fallback resolves an active row only when a key arrives without row focus.
	// Virtual collections can avoid scanning their data on idle frames.
	Fallback func() int
	// Focus is the single roving focus target used by headless groups. When it
	// is set, selected is its active row and Clickables may be nil.
	Focus      *widget.Clickable
	Clickables iter.Seq2[int, *widget.Clickable]
	// Targets supports composite controls whose focus targets are not clickables.
	// When set, it takes precedence over Clickables.
	Targets iter.Seq2[int, event.Tag]
}

// CollectionNavigation owns keyboard navigation for individually focusable
// rows or a single roving group target. The zero value is ready to use.
type CollectionNavigation struct {
	filters   []event.Filter
	pressed   key.Name
	action    string
	modifiers key.Modifiers
	typeahead navigation.Typeahead
}

// CollectionAxis determines which arrow keys move between rows.
type CollectionAxis uint8

const (
	CollectionVertical CollectionAxis = iota
	CollectionHorizontal
	CollectionBothAxes
)

// CollectionPolicy selects the interaction contract of a collection.
type CollectionPolicy struct {
	// PreserveHomeEnd leaves caret movement to an editable composite's editor.
	PreserveHomeEnd bool
	Wrap            bool
	Axis            CollectionAxis
	SelectOnFocus   bool
	Activate        bool
	Typeahead       bool
	// ActivateOnPress preserves controls such as manual tabs that select on
	// key-down instead of waiting for release.
	ActivateOnPress bool
	// Shortcuts enables Ctrl/Cmd+PageUp/PageDown and Ctrl/Cmd+Tab cycling.
	Shortcuts bool
	// Modifiers are optional for movement and activation, such as Shift for ranges.
	Modifiers key.Modifiers
	// Commands are component-owned keys processed in the same event order.
	Commands []key.Name
}

// CollectionEventKind distinguishes navigation, activation and custom commands.
type CollectionEventKind uint8

const (
	CollectionFocus CollectionEventKind = iota
	CollectionActivate
	CollectionCommand
)

// CollectionEvent preserves the key and its modifiers for component policies.
// Activation reports the modifiers captured on key-down, even on key release.
type CollectionEvent struct {
	Kind      CollectionEventKind
	Index     int
	Name      key.Name
	Modifiers key.Modifiers
}

type collectionActions struct {
	focus    func(string)
	activate func(string)
	forward  func(int)
	back     func()
	down     func(int)
	up       func(int)
	escape   func()
	event    func(CollectionEvent) int
}

// Update consumes navigation and activation keys. Focus changes are applied
// through the caller so it can materialize an off-screen destination first.
// selected is the fallback active index, or -1 when there is none. Call Update
// with a disabled context too, so pending activation and typeahead are reset.
func (s *CollectionNavigation) Update(gtx layout.Context, rows Collection, selected int, policy CollectionPolicy, focus, activate func(string)) {
	s.update(gtx, rows, selected, policy, collectionActions{focus: focus, activate: activate})
}

// UpdateWithEvents lets a component apply selection and custom-key policies.
// The callback returns the active index after handling an event so subsequent
// keys in the same frame continue from that destination, even before layout.
func (s *CollectionNavigation) UpdateWithEvents(
	gtx layout.Context,
	rows Collection,
	selected int,
	policy CollectionPolicy,
	handle func(CollectionEvent) int,
) {
	s.update(gtx, rows, selected, policy, collectionActions{event: handle})
}

func (s *CollectionNavigation) update(gtx layout.Context, rows Collection, selected int, policy CollectionPolicy, actions collectionActions) {
	if !gtx.Enabled() {
		s.pressed, s.action = "", ""
		s.modifiers = 0
		s.typeahead.Reset()
		return
	}
	s.filters = s.filters[:0]
	current := selected
	for i, click := range rows.targets(selected) {
		if gtx.Focused(click) && rows.Focus == nil {
			current = i
		}
		enabled := i >= 0 && i < rows.Count && (rows.Disabled == nil || !rows.Disabled(i))
		s.appendFilters(click, enabled, policy, actions)
	}
	list := navigation.List{Count: rows.Count, Disabled: rows.Disabled, Label: rows.Label}
	for len(s.filters) > 0 {
		e, ok := gtx.Event(s.filters...)
		if !ok {
			break
		}
		press, ok := e.(key.Event)
		if !ok {
			continue
		}
		if current < 0 && rows.Fallback != nil {
			current = rows.Fallback()
		}
		if slices.Contains(policy.Commands, press.Name) {
			if press.State == key.Press && current >= 0 && current < rows.Count && !listDisabled(rows, current) && actions.event != nil {
				current = actions.event(CollectionEvent{Kind: CollectionCommand, Index: current, Name: press.Name, Modifiers: press.Modifiers})
			}
			continue
		}
		next, moved := current, false
		switch press.Name {
		case key.NameDownArrow:
			if press.State == key.Press && policy.Axis != CollectionHorizontal {
				next, moved = navigation.Move(list, current, 1, policy.Wrap)
			} else if press.State == key.Press && actions.down != nil {
				actions.down(current)
			}
		case key.NameUpArrow:
			if press.State == key.Press && policy.Axis != CollectionHorizontal {
				next, moved = navigation.Move(list, current, -1, policy.Wrap)
			} else if press.State == key.Press && actions.up != nil {
				actions.up(current)
			}
		case key.NameEscape:
			if press.State == key.Press && actions.escape != nil {
				actions.escape()
			}
		case key.NameRightArrow:
			if press.State == key.Press {
				if actions.forward != nil {
					actions.forward(current)
				} else if policy.Axis != CollectionVertical {
					next, moved = navigation.Move(list, current, 1, policy.Wrap)
				}
			}
		case key.NameLeftArrow:
			if press.State == key.Press {
				if actions.back != nil {
					actions.back()
				} else if policy.Axis != CollectionVertical {
					next, moved = navigation.Move(list, current, -1, policy.Wrap)
				}
			}
		case key.NamePageUp, key.NamePageDown, key.NameTab:
			if policy.Shortcuts && press.State == key.Press && press.Modifiers.Contain(key.ModShortcut) {
				delta := 1
				if press.Name == key.NamePageUp || (press.Name == key.NameTab && press.Modifiers.Contain(key.ModShift)) {
					delta = -1
				}
				next, moved = navigation.Move(list, current, delta, policy.Wrap)
			}
		case key.NameHome:
			if press.State == key.Press {
				next, moved = navigation.First(list)
			}
		case key.NameEnd:
			if press.State == key.Press {
				next, moved = navigation.Last(list)
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			if !policy.Activate {
				continue
			}
			if policy.ActivateOnPress {
				if press.State == key.Press && current >= 0 && current < rows.Count && !listDisabled(rows, current) {
					if actions.activate != nil {
						actions.activate(rows.Key(current))
					}
					if actions.event != nil {
						current = actions.event(CollectionEvent{Kind: CollectionActivate, Index: current, Name: press.Name, Modifiers: press.Modifiers})
					}
				}
				continue
			}
			if press.State == key.Press {
				s.pressed, s.action = press.Name, ""
				s.modifiers = press.Modifiers
				if current >= 0 && current < rows.Count && (rows.Disabled == nil || !rows.Disabled(current)) {
					s.action = rows.Key(current)
				}
			} else {
				action := s.action
				modifiers := s.modifiers
				matched := s.pressed == press.Name
				s.pressed, s.action = "", ""
				s.modifiers = 0
				if matched && action != "" {
					for i := 0; i < rows.Count; i++ {
						if rows.Key(i) == action && (rows.Disabled == nil || !rows.Disabled(i)) {
							if actions.activate != nil {
								actions.activate(action)
							}
							if actions.event != nil {
								current = actions.event(CollectionEvent{Kind: CollectionActivate, Index: i, Name: press.Name, Modifiers: modifiers})
							}
							break
						}
					}
				}
			}
		default:
			if !policy.Typeahead || press.State != key.Press || press.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := navigation.Printable(press.Name)
			if text == "" {
				continue
			}
			query := s.typeahead.Append(gtx.Now, text)
			next, moved = navigation.Match(list, current, query)
			if !moved && query != text {
				s.typeahead.Set(text)
				next, moved = navigation.Match(list, current, text)
			}
		}
		if moved {
			current = next
			if actions.event != nil {
				current = actions.event(CollectionEvent{Kind: CollectionFocus, Index: next, Name: press.Name, Modifiers: press.Modifiers})
			}
			if actions.focus != nil {
				actions.focus(rows.Key(next))
			}
			if policy.SelectOnFocus && actions.activate != nil {
				actions.activate(rows.Key(next))
			}
		}
	}
}

func (rows Collection) targets(active int) iter.Seq2[int, event.Tag] {
	return func(yield func(int, event.Tag) bool) {
		if rows.Focus != nil {
			yield(active, rows.Focus)
			return
		}
		if rows.Targets != nil {
			rows.Targets(yield)
			return
		}
		if rows.Clickables != nil {
			for index, click := range rows.Clickables {
				if !yield(index, click) {
					return
				}
			}
		}
	}
}

func listDisabled(rows Collection, index int) bool {
	return rows.Disabled != nil && rows.Disabled(index)
}

func (s *CollectionNavigation) appendFilters(focus event.Tag, enabled bool, policy CollectionPolicy, actions collectionActions) {
	var names []key.Name
	if !policy.PreserveHomeEnd {
		names = append(names, key.NameHome, key.NameEnd)
	}
	if policy.Axis != CollectionHorizontal {
		names = append(names, key.NameDownArrow, key.NameUpArrow)
	} else {
		if actions.down != nil {
			names = append(names, key.NameDownArrow)
		}
		if actions.up != nil {
			names = append(names, key.NameUpArrow)
		}
	}
	if actions.escape != nil {
		names = append(names, key.NameEscape)
	}
	if policy.Axis != CollectionVertical {
		names = append(names, key.NameRightArrow, key.NameLeftArrow)
	} else {
		if actions.forward != nil {
			names = append(names, key.NameRightArrow)
		}
		if actions.back != nil {
			names = append(names, key.NameLeftArrow)
		}
	}
	for _, name := range names {
		s.filters = append(s.filters, key.Filter{Focus: focus, Name: name, Optional: policy.Modifiers})
	}
	// Own command keys even without an enabled row so they cannot fall through
	// to a composite's trigger. Dispatch still requires an enabled destination.
	for _, name := range policy.Commands {
		s.filters = append(s.filters, key.Filter{Focus: focus, Name: name, Optional: policy.Modifiers})
	}
	if policy.Activate && enabled {
		for _, name := range []key.Name{key.NameEnter, key.NameReturn, key.NameSpace} {
			s.filters = append(s.filters, key.Filter{Focus: focus, Name: name, Optional: policy.Modifiers})
		}
	}
	if policy.Typeahead && enabled {
		s.filters = append(s.filters, key.Filter{Focus: focus, Name: ""})
	}
	if policy.Shortcuts && enabled {
		s.filters = append(s.filters,
			key.Filter{Focus: focus, Name: key.NamePageUp, Required: key.ModShortcut},
			key.Filter{Focus: focus, Name: key.NamePageDown, Required: key.ModShortcut},
			key.Filter{Focus: focus, Name: key.NameTab, Required: key.ModShortcut, Optional: key.ModShift},
		)
	}
}

// RowInteraction consumes pointer activation while retaining caller-owned row
// state. Components can place the hit region inside their transformed box.
func RowInteraction(ctx *core.Context, gtx layout.Context, click *widget.Clickable, disabled bool, activate func()) Interaction {
	return rowInteraction(ctx, gtx, click, click, disabled, activate)
}

// RowInteractionWithFocus routes pointer focus to a composite's retained target
// while the row keeps its own keyboard activation and hit region.
func RowInteractionWithFocus(ctx *core.Context, gtx layout.Context, click, focus *widget.Clickable, disabled bool, activate func()) Interaction {
	return rowInteraction(ctx, gtx, click, focus, disabled, activate)
}

// RowInput retains click details for range selection, double clicks and cell filtering.
// Accept can reject a click owned by an interactive child before row focus changes.
type RowInput struct {
	Clickable *widget.Clickable
	Focus     *widget.Clickable
	Disabled  bool
	// PreserveFocus keeps the current focus when an auxiliary action is clicked.
	PreserveFocus bool
	Accept        func(widget.Click) bool
	OnClick       func(widget.Click)
}

func rowInteraction(ctx *core.Context, gtx layout.Context, click, focus *widget.Clickable, disabled bool, activate func()) Interaction {
	return UpdateRow(ctx, gtx, RowInput{
		Clickable: click, Focus: focus, Disabled: disabled,
		OnClick: func(widget.Click) {
			if activate != nil {
				activate()
			}
		},
	})
}

// UpdateRow processes activation and focus once while retaining the caller's clickable.
func UpdateRow(ctx *core.Context, gtx layout.Context, input RowInput) Interaction {
	click, focus := input.Clickable, input.Focus
	if focus == nil {
		focus = click
	}
	enabled := !input.Disabled && gtx.Enabled()
	presses := core.SnapshotPresses(click.History())
	for {
		activation, ok := click.Update(gtx)
		if !ok {
			break
		}
		if !enabled || (input.Accept != nil && !input.Accept(activation)) {
			continue
		}
		if input.PreserveFocus {
			core.PreserveFocus(ctx)
		} else {
			core.RequestFocusVisible(ctx, focus, presses.ClickFocusVisible(click.History()))
		}
		if input.OnClick != nil {
			input.OnClick(activation)
		}
	}
	if enabled && !input.PreserveFocus {
		core.FocusOnPress(ctx, focus, click.History(), presses.Active())
	}
	focused := gtx.Focused(click)
	return Interaction{Hovered: enabled && click.Hovered(), Pressed: enabled && click.Pressed(), Focused: focused, FocusVisible: enabled && core.FocusVisible(ctx, click, focused), Disabled: !enabled}
}
