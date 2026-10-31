package runtime

import (
	"testing"
)

func TestQueueDrainsInOrder(t *testing.T) {
	var q Queue[int]
	q.Push(1)
	q.Push(2)

	var got []int
	q.Drain(func(msg int) {
		got = append(got, msg)
	})
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("drained messages = %v, want [1 2]", got)
	}
	q.Drain(func(msg int) {
		t.Fatalf("unexpected message %d", msg)
	})
}

func TestQueueDropsWhenFull(t *testing.T) {
	q := Queue[int]{limit: 2}
	if !q.Push(1) || !q.Push(2) || q.Push(3) {
		t.Fatal("queue capacity was not enforced")
	}
	var got []int
	if dropped := q.Drain(func(msg int) { got = append(got, msg) }); dropped != 1 {
		t.Fatalf("dropped = %d, want 1", dropped)
	}
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("drained messages = %v, want [1 2]", got)
	}
}

func TestQueueReplacesMatchingEntryWhenFull(t *testing.T) {
	type message struct {
		key   string
		value int
	}
	q := Queue[message]{limit: 2}
	q.Push(message{key: "other", value: 1})
	q.Push(message{key: "stream", value: 2})
	if !q.PushOrReplace(message{key: "stream", value: 3}, func(current message) bool {
		return current.key == "stream"
	}) {
		t.Fatal("matching entry was not replaced")
	}
	var got []message
	q.Drain(func(msg message) { got = append(got, msg) })
	if len(got) != 2 || got[1].value != 3 {
		t.Fatalf("drained messages = %#v, want latest stream value", got)
	}
}
