package navigation

import (
	"testing"
	"time"
)

func TestTypeaheadBufferExpires(t *testing.T) {
	state := new(Typeahead)
	start := time.Unix(1, 0)
	if got := state.Append(start, "n"); got != "n" {
		t.Fatalf("first typeahead = %q, want n", got)
	}
	if got := state.Append(start.Add(100*time.Millisecond), "e"); got != "ne" {
		t.Fatalf("continued typeahead = %q, want ne", got)
	}
	expired := start.Add(100*time.Millisecond + DefaultTypeaheadTimeout + time.Millisecond)
	if got := state.Append(expired, "g"); got != "g" {
		t.Fatalf("expired typeahead = %q, want g", got)
	}
}
