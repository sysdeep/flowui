package style

import "testing"

// TestStateMaskMappingIsOneToOne pins the design decision that State -> StateMask
// does not suppress any family. Suppression belongs in the component that
// produces State, where it can be asserted; hiding it here would make a
// disabled control silently stop reporting hover, and would put a rule into the
// conversion that no document states.
func TestStateMaskMappingIsOneToOne(t *testing.T) {
	all := State{
		Disabled: true, Hovered: true, Pressed: true, Focused: true, FocusVisible: true,
		Selected: true, Checked: true, Indeterminate: true, ReadOnly: true, Invalid: true,
		Loading: true, Open: true, Expanded: true, Dragging: true, DropTarget: true,
		Warning: true,
	}
	if got := all.Bits(); got != StateMaskAll {
		t.Fatalf("State.Bits() = %#x, want every bit %#x", got, StateMaskAll)
	}
	if round := all.Bits().Split(); round != all {
		t.Fatalf("StateMask.Split() did not round-trip:\n got %+v\nwant %+v", round, all)
	}
	// A disabled control keeps reporting hover/press, on purpose.
	disabledHover := State{Disabled: true, Hovered: true}.Bits()
	if disabledHover&Hovered == 0 {
		t.Fatal("Disabled must not clear Hovered inside StateMask")
	}
}
