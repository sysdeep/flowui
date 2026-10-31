package style

import (
	"testing"
)

func TestUnderlineIsHashedAndResolved(t *testing.T) {
	plain := New()
	underlined := New().Underline(true)

	if plain.Hash64() == underlined.Hash64() {
		t.Fatal("Underline(true) hashed identically to an unset declaration")
	}
	if New().Underline(false).Hash64() == underlined.Hash64() {
		t.Fatal("Underline(false) hashed identically to Underline(true)")
	}

	resolved := underlined.Resolve(State{})
	if resolved.Text == nil || resolved.Text.Underline == nil || !*resolved.Text.Underline {
		t.Fatal("Underline did not survive resolution")
	}
}

func TestUnderlineComposesWithStateRules(t *testing.T) {
	declaration := New().
		Underline(true).
		On(Disabled, New().Underline(false))

	resolved := declaration.Resolve(State{Disabled: true})
	if resolved.Text == nil || resolved.Text.Underline == nil {
		t.Fatal("underline missing after state resolution")
	}
	if *resolved.Text.Underline {
		t.Fatal("state row did not turn the underline off")
	}
}
