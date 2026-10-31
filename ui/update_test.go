package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	internalexplorer "github.com/qianniancn/flowui/internal/explorer"
)

func TestDoAdaptsContextFreeCommand(t *testing.T) {
	var got int
	cmd := Do(func(send Send[int]) {
		send(7)
	})
	if err := cmd(context.Background(), func(msg int) { got = msg }); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got != 7 {
		t.Fatalf("message = %d, want 7", got)
	}
}

func TestEventWindowDecoratesEffectContextWithExplorerService(t *testing.T) {
	service := internalexplorer.New(nil)
	window := &platformEventWindow{explorer: service}
	ctx := window.DecorateEffectContext(context.Background())
	if got := internalexplorer.FromContext(ctx); got != service {
		t.Fatalf("explorer service = %p, want %p", got, service)
	}
}

func TestMapCmdMapsMessagesAndPreservesContextAndError(t *testing.T) {
	type contextKey struct{}
	type parentMessage struct{ value int }
	wantErr := errors.New("failed")
	ctx := context.WithValue(context.Background(), contextKey{}, "value")
	child := func(cmdCtx context.Context, send Send[int]) error {
		if cmdCtx != ctx || cmdCtx.Value(contextKey{}) != "value" {
			t.Fatal("mapped command did not preserve context")
		}
		send(2)
		send(3)
		return wantErr
	}
	mapped := MapCmd(child, func(msg int) parentMessage { return parentMessage{value: msg} })

	var got []int
	err := mapped(ctx, func(msg parentMessage) { got = append(got, msg.value) })
	if !errors.Is(err, wantErr) {
		t.Fatalf("mapped error = %v, want %v", err, wantErr)
	}
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("mapped messages = %v, want [2 3]", got)
	}
}

type batchContextKey struct{}

func TestBatchRunsCommandsAndPreservesErrors(t *testing.T) {
	ctx := context.WithValue(context.Background(), batchContextKey{}, "batch")
	var got []int
	var mu sync.Mutex
	batch := Batch(
		DoContext(func(gotCtx context.Context, send Send[int]) error {
			if gotCtx.Value(batchContextKey{}) != "batch" {
				t.Errorf("batch context was not preserved")
			}
			send(1)
			return nil
		}),
		DoContext(func(_ context.Context, send Send[int]) error {
			send(2)
			return errors.New("batch failed")
		}),
	)
	err := batch(ctx, func(value int) {
		mu.Lock()
		got = append(got, value)
		mu.Unlock()
	})
	if err == nil || err.Error() != "batch failed" {
		t.Fatalf("batch error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("batch messages = %v, want two messages", got)
	}
}

func TestMapSubscriptionMapsMessagesAndPreservesKey(t *testing.T) {
	child := Subscribe("events", func(_ context.Context, send Send[int]) error {
		send(7)
		return errors.New("subscription failed")
	})
	parent := MapSubscription(child, func(value int) string { return fmt.Sprintf("value=%d", value) })
	if parent.Key != "events" {
		t.Fatalf("mapped subscription key = %q", parent.Key)
	}
	var got string
	err := parent.Run(context.Background(), func(value string) { got = value })
	if got != "value=7" || err == nil || err.Error() != "subscription failed" {
		t.Fatalf("mapped subscription = %q, %v", got, err)
	}
}

func TestWriteEffectErrorIncludesPanicStack(t *testing.T) {
	var output bytes.Buffer
	writeEffectError(&output, &EffectError{
		Kind:  EffectSubscription,
		Key:   "events",
		Panic: "broken",
		Stack: []byte("stack trace\n"),
	})
	if got := output.String(); got != "flowui: subscription \"events\" panicked: broken\nstack trace\n" {
		t.Fatalf("error output = %q", got)
	}
}

func TestWriteRuntimePanicIncludesStack(t *testing.T) {
	var output bytes.Buffer
	writeEffectError(&output, &RuntimePanicError{
		Phase: RuntimePhaseView,
		Panic: "broken",
		Stack: []byte("stack trace\n"),
	})
	if got := output.String(); got != "flowui: view panicked: broken\nstack trace\n" {
		t.Fatalf("runtime panic output = %q", got)
	}
}
