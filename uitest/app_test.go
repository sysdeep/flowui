package uitest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestAppHarnessDrivesInitUpdateAndCommands(t *testing.T) {
	type model struct {
		value   int
		loading bool
	}
	type message struct {
		value int
		load  bool
	}

	release := make(chan struct{})
	initSent := make(chan struct{})
	harness := uitest.NewAppWithConfig(uitest.AppConfig[model, message]{
		Initial: model{},
		Init: ui.Do(func(send ui.Send[message]) {
			send(message{value: 1})
			close(initSent)
		}),
		Update: func(model *model, msg message) ui.Cmd[message] {
			if msg.load {
				model.loading = true
				return ui.DoContext(func(ctx context.Context, send ui.Send[message]) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-release:
						send(message{value: 9})
						return nil
					}
				})
			}
			model.value = msg.value
			model.loading = false
			return nil
		},
	})
	defer func() {
		if err := harness.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}()

	waitTestValue(t, initSent)
	if harness.Model().value != 1 {
		if !harness.Wait(time.Second) {
			t.Fatal("initial command did not invalidate the app")
		}
		harness.Frame()
	}
	if got := harness.Model().value; got != 1 {
		t.Fatalf("initial value = %d, want 1", got)
	}

	harness.Send(message{load: true})
	if got := harness.Frame(); !got.loading || got.value != 1 {
		t.Fatalf("loading model = %#v", got)
	}
	close(release)
	if !harness.Wait(time.Second) {
		t.Fatal("command result did not invalidate the app")
	}
	if got := harness.Frame(); got.loading || got.value != 9 {
		t.Fatalf("loaded model = %#v", got)
	}
}

func TestAppHarnessLatestMappedBatchUsesUpdateOrder(t *testing.T) {
	type message struct {
		request bool
		value   int
	}
	release := make(chan struct{})
	oldDone := make(chan struct{})
	newDone := make(chan struct{})
	h := uitest.NewApp(0, func(model *int, msg message) ui.Cmd[message] {
		if !msg.request {
			*model = msg.value
			return nil
		}
		cmd := ui.MapCmd(ui.LatestCmd("search", func(_ context.Context, send ui.Send[int]) error {
			send(msg.value)
			return nil
		}), func(value int) message { return message{value: value} })
		return ui.Batch(func(ctx context.Context, send ui.Send[message]) error {
			if msg.value == 1 {
				<-release
				defer close(oldDone)
			} else {
				defer close(newDone)
			}
			return cmd(ctx, send)
		}, func(context.Context, ui.Send[message]) error { return nil })
	})
	t.Cleanup(func() { _ = h.Close() })
	h.Send(message{request: true, value: 1})
	h.Send(message{request: true, value: 2})
	h.Frame()
	waitTestValue(t, newDone)
	close(release)
	waitTestValue(t, oldDone)
	if got := h.Frame(); got != 2 {
		t.Fatalf("latest result=%d, want 2", got)
	}
}

func TestAppHarnessReportsEffectErrorsAndCancelsOnClose(t *testing.T) {
	type message struct{ fail bool }
	want := errors.New("failed")
	canceled := make(chan struct{})
	harness := uitest.NewAppWithConfig(uitest.AppConfig[struct{}, message]{
		Init: ui.DoContext(func(ctx context.Context, _ ui.Send[message]) error {
			<-ctx.Done()
			close(canceled)
			return ctx.Err()
		}),
		Update: func(_ *struct{}, msg message) ui.Cmd[message] {
			if !msg.fail {
				return nil
			}
			return ui.DoContext(func(_ context.Context, _ ui.Send[message]) error { return want })
		},
	})

	harness.Send(message{fail: true})
	harness.Frame()
	if !harness.Wait(time.Second) {
		t.Fatal("effect error did not invalidate the app")
	}
	harness.Frame()
	gotErrors := harness.Errors()
	if len(gotErrors) != 1 || !errors.Is(gotErrors[0], want) {
		t.Fatalf("effect errors = %v, want %v", gotErrors, want)
	}
	if err := harness.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	waitTestValue(t, canceled)
}

func TestAppHarnessDrivesSynchronousUpdate(t *testing.T) {
	harness := uitest.NewApp(1, func(model *int, msg int) ui.Cmd[int] {
		*model += msg
		return nil
	})
	defer harness.Close()
	harness.Send(2)
	if got := harness.Frame(); got != 3 {
		t.Fatalf("model = %d, want 3", got)
	}
}

func waitTestValue(t *testing.T, values <-chan struct{}) {
	t.Helper()
	select {
	case <-values:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for app harness")
	}
}
