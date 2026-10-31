package style

type State struct {
	Hovered       bool
	Pressed       bool
	Focused       bool
	FocusVisible  bool
	Disabled      bool
	Selected      bool
	Checked       bool
	Indeterminate bool
	ReadOnly      bool
	Invalid       bool
	Loading       bool
	Open          bool
	Expanded      bool
	Dragging      bool
	DropTarget    bool
	Warning       bool
}

// StateMask is a bitmask of interaction states.
//
// StateMask covers the state families that every control shares. It is plain
// data rather than predicates, which buys three properties a predicate tree
// cannot have:
//
//   - precedence is deterministic and documentable (see Rule),
//   - a declaration can be hashed statically, without evaluating anything,
//   - a linter can prove coverage ("every component has a Disabled rule").
//
// The name deliberately avoids "variant": in this library a variant is the
// visual personality a component exposes (button.Variant), while StateMask
// describes the interaction state a component is in.
type StateMask uint16

const (
	Disabled StateMask = 1 << iota
	Hovered
	Pressed
	Focused
	FocusVisible
	Selected
	Checked
	Indeterminate
	ReadOnly
	Invalid
	Loading
	Open
	Expanded
	Dragging
	DropTarget
	Warning
)

// StateMaskAll is the union of every defined state bit.
const StateMaskAll = Disabled |
	Hovered |
	Pressed |
	Focused |
	FocusVisible |
	Selected |
	Checked |
	Indeterminate |
	ReadOnly |
	Invalid |
	Loading |
	Open |
	Expanded |
	Dragging |
	DropTarget |
	Warning

// Bits returns the bits a resolved interaction state contributes.
//
// The mapping is intentionally one-to-one: no bit implies another. Any
// suppression between families (for example "a disabled control is not
// hovered") belongs in the component that produces State, where it can be
// asserted directly, rather than hidden in this conversion.
func (s State) Bits() StateMask {
	var v StateMask
	if s.Disabled {
		v |= Disabled
	}
	if s.Hovered {
		v |= Hovered
	}
	if s.Pressed {
		v |= Pressed
	}
	if s.Focused {
		v |= Focused
	}
	if s.FocusVisible {
		v |= FocusVisible
	}
	if s.Selected {
		v |= Selected
	}
	if s.Checked {
		v |= Checked
	}
	if s.Indeterminate {
		v |= Indeterminate
	}
	if s.ReadOnly {
		v |= ReadOnly
	}
	if s.Invalid {
		v |= Invalid
	}
	if s.Loading {
		v |= Loading
	}
	if s.Open {
		v |= Open
	}
	if s.Expanded {
		v |= Expanded
	}
	if s.Dragging {
		v |= Dragging
	}
	if s.DropTarget {
		v |= DropTarget
	}
	if s.Warning {
		v |= Warning
	}
	return v
}

// Split decomposes a state into independent families. It exists for tests and
// for components that need to reason about one family at a time.
func (v StateMask) Split() State {
	return State{
		Disabled:      v&Disabled != 0,
		Hovered:       v&Hovered != 0,
		Pressed:       v&Pressed != 0,
		Focused:       v&Focused != 0,
		FocusVisible:  v&FocusVisible != 0,
		Selected:      v&Selected != 0,
		Checked:       v&Checked != 0,
		Indeterminate: v&Indeterminate != 0,
		ReadOnly:      v&ReadOnly != 0,
		Invalid:       v&Invalid != 0,
		Loading:       v&Loading != 0,
		Open:          v&Open != 0,
		Expanded:      v&Expanded != 0,
		Dragging:      v&Dragging != 0,
		DropTarget:    v&DropTarget != 0,
		Warning:       v&Warning != 0,
	}
}
