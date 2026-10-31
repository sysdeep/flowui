package style

import (
	"testing"

	"gioui.org/unit"
)

func TestTransientShadowChangesHash(t *testing.T) {
	normal := New().BoxShadow(0, 0, 5, 0, RGB(0x3e63dd))
	transient := New().BoxShadowTransient(0, 0, 5, 0, RGB(0x3e63dd))
	if normal.Hash64() == transient.Hash64() {
		t.Fatal("transient shadow was not included in the style hash")
	}
}

func TestPartHashIsIndependentOfInsertionOrder(t *testing.T) {
	for _, count := range []int{1, 8, 9} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			forward, reverse := Style{}, Style{}
			for i := range count {
				forward = forward.Part(Part(string(rune('a'+i))), Style{}.Width(unit.Dp(i+1)))
			}
			for i := count - 1; i >= 0; i-- {
				reverse = reverse.Part(Part(string(rune('a'+i))), Style{}.Width(unit.Dp(i+1)))
			}
			want := forward.Hash64()
			for range 20 {
				if got := reverse.Hash64(); got != want {
					t.Fatalf("%d parts hash = %#x, want %#x", count, got, want)
				}
			}
			changed := reverse.Part(Part("a"), Style{}.Width(100))
			if changed.Hash64() == want {
				t.Fatal("changing a Part did not change the hash")
			}
		})
	}
}
