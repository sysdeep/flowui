package core

import "testing"

func TestMemoCachesAndInvalidates(t *testing.T) {
	for _, tc := range []struct {
		name         string
		theme, input uint64
		want         int
	}{
		{"cached", 1, 100, 1},
		{"theme changed", 2, 100, 2},
		{"input changed", 1, 200, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var store Store
			store.BeginFrame()
			calls := 0
			load := func() int { calls++; return calls }
			UseMemo[int](&store, "component", "value", 1, 100).Compute(load)
			memo := UseMemo[int](&store, "component", "value", tc.theme, tc.input)
			if got := memo.Compute(load); got != tc.want {
				t.Fatalf("memo = %d, want %d", got, tc.want)
			}
			if got := memo.Compute(load); got != tc.want || calls != tc.want {
				t.Fatalf("cached memo/calls = %d/%d, want %d/%d", got, calls, tc.want, tc.want)
			}
		})
	}
}

func TestMemoSeparatesIdentities(t *testing.T) {
	var store Store
	store.BeginFrame()
	UseMemo[int](&store, "first", "value", 1, 100).Set(10)
	for _, tc := range []struct {
		name, key, slot string
	}{
		{"different key", "second", "value"},
		{"different slot", "first", "other"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			memo := UseMemo[int](&store, tc.key, tc.slot, 1, 100)
			if got := memo.Compute(func() int { return 20 }); got != 20 {
				t.Fatalf("independent memo = %d, want 20", got)
			}
		})
	}
	if got, ok := UseMemo[int](&store, "first", "value", 1, 100).Get(); !ok || got != 10 {
		t.Fatalf("original memo = %d/%v, want 10/true", got, ok)
	}
}

func BenchmarkMemo_Hit(b *testing.B) {
	store := &Store{}
	store.BeginFrame()

	memo := UseMemo[int](store, "key1", "slot1", 1, 1100)
	memo.Set(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memo := UseMemo[int](store, "key1", "slot1", 1, 1100)
		_ = memo.Compute(func() int {
			return 99 // Should not be called
		})
	}
}

func BenchmarkMemo_Miss(b *testing.B) {
	store := &Store{}
	store.BeginFrame()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memo := UseMemo[int](store, "key1", "slot1", 1, uint64(i))
		_ = memo.Compute(func() int {
			return i
		})
	}
}
