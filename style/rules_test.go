package style

import (
	"reflect"
	"testing"
	"time"
)

const (
	testHoverOpacity   = float32(0.90)
	testPressOpacity   = float32(0.80)
	testDisableOpacity = float32(0.38)
)

func TestRulesCaptureCallerOwnedRows(t *testing.T) {
	t.Parallel()
	constructors := []struct {
		name  string
		build func(*Tween) Style
	}{
		{name: "Rules", build: func(tween *Tween) Style {
			return Style{}.Rules(Rule{When: Hovered, Set: Style{}.Opacity(.5), Tween: tween})
		}},
		{name: "On", build: func(tween *Tween) Style {
			return Style{}.On(Hovered, Style{}.Opacity(.5), tween)
		}},
		{name: "OnAny", build: func(tween *Tween) Style {
			return Style{}.OnAny(Hovered|Pressed, Style{}.Opacity(.5), tween)
		}},
		{name: "OnNot", build: func(tween *Tween) Style {
			return Style{}.OnNot(Hovered, Disabled, Style{}.Opacity(.5), tween)
		}},
	}
	for _, tc := range constructors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tween := &Tween{Property: PropOpacity, Duration: time.Second, Delay: time.Millisecond}
			decl := tc.build(tween)
			hash := decl.Hash64()
			tween.Duration = 2 * time.Second
			tween.Delay = 2 * time.Millisecond
			if decl.Hash64() != hash {
				t.Fatal("mutating the input tween changed the declaration hash")
			}
			resolved := decl.Resolve(State{Hovered: true})
			if len(resolved.Transitions) != 1 || resolved.Transitions[0].Duration != time.Second ||
				resolved.Transitions[0].Delay != time.Millisecond {
				t.Fatalf("mutating the input tween changed the resolved transition: %+v", resolved.Transitions)
			}
		})
	}
}

func TestRuleTableReturnsIndependentSnapshot(t *testing.T) {
	t.Parallel()
	input := []Rule{{When: Hovered, Set: Style{}.Opacity(.5), Tween: Quick(PropOpacity, time.Second)}}
	decl := Style{}.Rules(input...)
	derived := decl.On(Disabled, Style{}.Opacity(.2))
	hash, derivedHash := decl.Hash64(), derived.Hash64()
	input[0].When = Pressed

	rows := decl.RuleTable()
	rows[0].When = Disabled
	rows[0].Set = Style{}.Opacity(.1)
	rows[0].Tween.Duration = 2 * time.Second
	if decl.Hash64() != hash || derived.Hash64() != derivedHash {
		t.Fatal("mutating a rule snapshot changed an existing declaration")
	}
	resolved := decl.Resolve(State{Hovered: true})
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != .5 {
		t.Fatalf("hover opacity changed after snapshot mutation: %+v", resolved.Paint)
	}
	fresh := decl.RuleTable()
	if fresh[0].When != Hovered || fresh[0].Tween.Duration != time.Second {
		t.Fatalf("snapshot mutation leaked into a later read: %+v", fresh[0])
	}
}

func TestRuleTweenOverridesOnlyMatchedTransition(t *testing.T) {
	t.Parallel()
	decl := Style{}.Opacity(1).
		Transition(PropOpacity, time.Second).
		Transition(PropTransform, 3*time.Second).
		On(Hovered, Style{}.Opacity(.5), &Tween{
			Property: PropOpacity,
			Duration: 2 * time.Second,
			Delay:    time.Millisecond,
		})
	for _, tc := range []struct {
		name     string
		state    State
		duration time.Duration
		delay    time.Duration
	}{
		{name: "resting", duration: time.Second},
		{name: "hovered", state: State{Hovered: true}, duration: 2 * time.Second, delay: time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolved := decl.Resolve(tc.state)
			if len(resolved.Transitions) != 2 {
				t.Fatalf("transition count = %d, want 2", len(resolved.Transitions))
			}
			for _, transition := range resolved.Transitions {
				switch transition.Property {
				case PropOpacity:
					if transition.Duration != tc.duration || transition.Delay != tc.delay {
						t.Fatalf("opacity transition = %+v", transition)
					}
				case PropTransform:
					if transition.Duration != 3*time.Second {
						t.Fatalf("unrelated transition changed: %+v", transition)
					}
				}
			}
		})
	}
}

func TestRuleTransitionCacheSafety(t *testing.T) {
	t.Parallel()
	ease := func(value float32) float32 { return value * value }
	custom := Style{}.Opacity(.5).Transition(PropOpacity, time.Second, TransitionEase(ease))
	for _, tc := range []struct {
		name   string
		decl   Style
		unsafe bool
	}{
		{
			name: "data-only tween",
			decl: Style{}.On(Hovered, Style{}.Opacity(.5), Quick(PropOpacity, time.Second)),
		},
		{
			name:   "custom easing in rule",
			decl:   Style{}.On(Hovered, custom),
			unsafe: true,
		},
		{
			name:   "custom easing in nested rule part",
			decl:   Style{}.On(Hovered, Style{}.Part(PartLabel, Style{}.On(Pressed, custom))),
			unsafe: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.decl.CacheUnsafe(); got != tc.unsafe {
				t.Fatalf("CacheUnsafe() = %t, want %t", got, tc.unsafe)
			}
		})
	}
	resolved := Style{}.On(Hovered, custom).Resolve(State{Hovered: true})
	if len(resolved.Transitions) != 1 || resolved.Transitions[0].Ease == nil ||
		resolved.Transitions[0].Ease(.5) != .25 {
		t.Fatalf("rule lost its custom easing: %+v", resolved.Transitions)
	}
}

func resolvedOpacity(t *testing.T, s Style, state State) string {
	t.Helper()
	resolved := s.Resolve(state)
	if resolved.Paint == nil || resolved.Paint.Opacity == nil {
		return "nil"
	}
	return formatFloat(*resolved.Paint.Opacity)
}

func formatFloat(v float32) string {
	switch {
	case v == testDisableOpacity:
		return "disabled"
	case v == testPressOpacity:
		return "pressed"
	case v == testHoverOpacity:
		return "hovered"
	default:
		return "other"
	}
}

// TestRuleDeclarationOrderBreaksTies checks that the later row wins when rank
// and specificity are equal.
func TestRuleDeclarationOrderBreaksTies(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: Hovered, Set: Style{}.Opacity(testHoverOpacity)},
		Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)},
	)
	if got := resolvedOpacity(t, decl, State{Hovered: true}); got != "pressed" {
		t.Fatalf("later row must win the tie, got %s", got)
	}
}

// TestRuleMatchedBitsBeatDeclarationOrder checks that a more specific row wins
// even when it was declared first.
func TestRuleMatchedBitsBeatDeclarationOrder(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: Checked | Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Checked, Set: Style{}.Opacity(testHoverOpacity)},
	)
	both := State{Checked: true, Disabled: true}
	if got := resolvedOpacity(t, decl, both); got != "disabled" {
		t.Fatalf("2-bit row must beat the 1-bit row declared after it, got %s", got)
	}
	only := State{Checked: true}
	if got := resolvedOpacity(t, decl, only); got != "hovered" {
		t.Fatalf("1-bit row must still apply alone, got %s", got)
	}
}

// TestRuleRankOrdersOrthogonalFamilies covers the case bit count cannot order.
//
// "Disabled always beats Invalid and Hovered" is a fixed precedence, but the
// two families are orthogonal: with bit-count priority alone, an
// Invalid|Hovered row (2 bits) would outrank a Disabled row (1 bit), which is
// the wrong answer. Expressing the intent through specificity would need one
// row per combination of the losing family's bits.
func TestRuleRankOrdersOrthogonalFamilies(t *testing.T) {
	// Rank at zero: the higher-specificity row wins, which is wrong here.
	unranked := Style{}.Rules(
		Rule{When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Invalid | Hovered, Set: Style{}.Opacity(testHoverOpacity)},
	)
	state := State{Disabled: true, Invalid: true, Hovered: true}
	if got := resolvedOpacity(t, unranked, state); got != "hovered" {
		t.Fatalf("without Rank the high-specificity row should win; got %s", got)
	}

	// Rank on the Disabled family makes the intent explicit and wins.
	ranked := Style{}.Rules(
		Rule{Rank: 1, When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Invalid | Hovered, Set: Style{}.Opacity(testHoverOpacity)},
	)
	if got := resolvedOpacity(t, ranked, state); got != "disabled" {
		t.Fatalf("Rank must beat matched-bit count, got %s", got)
	}
}

// TestRuleNotIsIsolated covers the Not field: a row is skipped when any of its
// Not bits is set, which is how "focus ring, but not while disabled" is said
// without a nested All/Not predicate.
func TestRuleNotIsIsolated(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: FocusVisible, Not: Disabled, Set: Style{}.Opacity(testHoverOpacity)},
	)
	if got := resolvedOpacity(t, decl, State{FocusVisible: true}); got != "hovered" {
		t.Fatalf("row should apply when no Not bit is set, got %s", got)
	}
	if got := resolvedOpacity(t, decl, State{FocusVisible: true, Disabled: true}); got != "nil" {
		t.Fatalf("row must be skipped when a Not bit is set, got %s", got)
	}
}

// TestRuleTableIsCacheSafeAndStaticallyHashable is the payoff claim of the
// design: because a rule holds no closures, a declaration built from rules can
// be hashed and cached without knowing the interaction state.
func TestRuleTableIsCacheSafeAndStaticallyHashable(t *testing.T) {
	build := func() Style {
		return Style{}.Radius(4).Rules(
			Rule{When: Hovered, Set: Style{}.Opacity(testHoverOpacity)},
			Rule{When: Checked | Indeterminate, Set: Style{}.Opacity(testPressOpacity)},
			Rule{Rank: 1, When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		)
	}

	if build().CacheUnsafe() {
		t.Fatal("a declaration built only from Rules must be cache safe")
	}

	want := build().Hash64()
	if got := build().Hash64(); got != want {
		t.Fatalf("rebuilt table hash = %#x, want %#x", got, want)
	}

	// And the hash must actually distinguish tables that differ.
	other := Style{}.Radius(4).Rules(
		Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)},
	)
	if other.Hash64() == want {
		t.Fatal("two different rule tables produced the same static hash")
	}
}

// TestRuleTablesHashApartAndStayCacheSafe pins the caching contract of rule
// rows: the key is the bits themselves, so two tables differing only in When
// never share a hash, and these concrete rule declarations are cache-safe.
func TestRuleTablesHashApartAndStayCacheSafe(t *testing.T) {
	checked := Style{}.Rules(Rule{When: Checked, Set: Style{}.Opacity(testPressOpacity)})
	hovered := Style{}.Rules(Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)})
	if checked.Hash64() == hovered.Hash64() {
		t.Fatal("rule tables differing only in When must not share a hash")
	}
	if checked.CacheUnsafe() || hovered.CacheUnsafe() {
		t.Fatal("rule tables must stay cache safe")
	}
}

// TestOnAnyAttachesTweenToFirstRowOnly checks that the tween lands on the
// first expanded row (the lowest bit), not on every row.
func TestOnAnyAttachesTweenToFirstRowOnly(t *testing.T) {
	set := Style{}.Radius(2)
	rows := Style{}.OnAny(Hovered|Pressed, set, Quick(PropTransform, 90*time.Millisecond)).RuleTable()

	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].Tween == nil || rows[0].Tween.Property != PropTransform || rows[0].Tween.Duration != 90*time.Millisecond {
		t.Fatalf("first row must carry the tween, got %+v", rows[0].Tween)
	}
	if rows[1].Tween != nil {
		t.Fatalf("second row must not carry the tween, got %+v", rows[1].Tween)
	}
}

// TestOnAnyEmptyMaskAppendsNothing keeps the empty mask a no-op rather than an
// accidental unconditional row; the unconditional row is On's job.
func TestOnAnyEmptyMaskAppendsNothing(t *testing.T) {
	decl := Style{}.OnAny(0, Style{}.Opacity(0.5))
	if resolved := decl.Resolve(State{}); resolved.Paint != nil {
		t.Fatalf("OnAny(0) applied an unconditional style: %#v", resolved.Paint)
	}
}

// TestOnZeroMaskIsUnconditionalRow pins the escape hatch checkbox's resting
// row relies on: When 0 matches every state and sorts first.
func TestOnZeroMaskIsUnconditionalRow(t *testing.T) {
	decl := Style{}.On(0, Style{}.Opacity(0.5))
	for _, state := range []State{{}, {Hovered: true}, {Disabled: true, Checked: true}} {
		resolved := decl.Resolve(state)
		if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.5 {
			t.Fatalf("unconditional row must resolve in state %+v", state)
		}
	}
}

// Compare resolved values, including transitions, for the authoring helpers
// and their equivalent raw rules.
func TestOnFormsResolveLikeTheirRawTables(t *testing.T) {
	set := Style{}.Background(RGB(0xff0000)).Scale(0.97)
	alt := Style{}.TextColor(RGB(0x00ff00))
	tween := Quick(PropTransform, 90*time.Millisecond)

	sugared := Style{}.Opacity(1).
		On(Hovered, set).
		OnAny(Pressed|FocusVisible, alt, tween).
		OnNot(Selected, Disabled, Style{}.Opacity(0.5))
	raw := Style{}.Opacity(1).Rules(
		Rule{When: Hovered, Set: set},
		Rule{When: Pressed, Set: alt, Tween: tween},
		Rule{When: FocusVisible, Set: alt},
		Rule{When: Selected, Not: Disabled, Set: Style{}.Opacity(0.5)},
	)

	if sugared.Hash64() != raw.Hash64() {
		t.Fatalf("sugared declaration hashed %#x, raw table %#x", sugared.Hash64(), raw.Hash64())
	}
	for _, tc := range []struct {
		name  string
		state State
	}{
		{name: "resting"},
		{name: "hovered", state: State{Hovered: true}},
		{name: "pressed", state: State{Pressed: true}},
		{name: "focus visible", state: State{FocusVisible: true}},
		{name: "selected", state: State{Selected: true}},
		{name: "disabled selection", state: State{Selected: true, Disabled: true}},
		{name: "combined", state: State{Hovered: true, Pressed: true, FocusVisible: true, Selected: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := sugared.Resolve(tc.state), raw.Resolve(tc.state)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("On forms and raw rules resolved differently:\n got %#v\nwant %#v", a, b)
			}
		})
	}
}
