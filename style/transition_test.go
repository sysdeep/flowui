package style

import (
	"testing"
	"time"
)

func TestTransitionIgnoresUnknownPropertyID(t *testing.T) {
	const unknown PropertyID = 255
	if unknown.Animatable() {
		t.Fatal("unknown property should not be animatable")
	}
	built := Style{}.
		Opacity(1).
		Transition(unknown, time.Second).
		Transition(PropOpacity, time.Millisecond).
		Resolve(State{})
	if len(built.Transitions) != 1 || built.Transitions[0].Property != PropOpacity {
		t.Fatalf("transitions = %#v, want only PropOpacity", built.Transitions)
	}
}
