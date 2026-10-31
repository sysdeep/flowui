package runtime

import (
	"context"
	"errors"
	"fmt"
	goruntime "runtime"
	"sync"
	"testing"
	"time"
)

func TestBatchRunsCommandsConcurrentlyAndCollectsMessages(t *testing.T) {
	const commandCount = 8
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(commandCount)
	finished.Add(commandCount)
	gate := make(chan struct{})
	cmds := make([]Cmd[int], commandCount)
	for id := range commandCount {
		cmds[id] = func(context.Context, Send[int]) error {
			ready.Done()
			<-gate
			finished.Done()
			return nil
		}
	}
	batch := Batch(append(cmds, nil)...)
	if batch == nil {
		t.Fatal("batch with active commands was nil")
	}

	done := make(chan error, 1)
	go func() {
		done <- batch(context.Background(), func(int) {})
	}()
	readyDone := make(chan struct{})
	go func() {
		ready.Wait()
		close(readyDone)
	}()
	select {
	case <-readyDone:
	case <-time.After(time.Second):
		close(gate)
		t.Fatal("timed out waiting for batch commands to start")
	}
	close(gate)
	finishedDone := make(chan struct{})
	go func() {
		finished.Wait()
		close(finishedDone)
	}()
	select {
	case <-finishedDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for batch commands to finish")
	}
	if err := receiveRuntimeTestValue(t, done); err != nil {
		t.Fatalf("batch error = %v", err)
	}
}

func TestBatchReturnsErrorAndPropagatesPanic(t *testing.T) {
	want := errors.New("command failed")
	batch := Batch(
		func(context.Context, Send[int]) error { return want },
		func(context.Context, Send[int]) error { return nil },
	)
	if err := batch(context.Background(), func(int) {}); !errors.Is(err, want) {
		t.Fatalf("batch error = %v, want %v", err, want)
	}

	panicking := Batch(
		func(context.Context, Send[int]) error { return nil },
		func(context.Context, Send[int]) error { panic("batch child") },
	)
	defer func() {
		if recover() == nil {
			t.Fatal("batch did not re-raise child panic")
		}
	}()
	_ = panicking(context.Background(), func(int) {})
}

func TestBatchIgnoresNilCommands(t *testing.T) {
	if Batch[int]() != nil || Batch[int](nil, nil) != nil {
		t.Fatal("empty batch should be nil")
	}
	var ran bool
	single := Batch[int](nil, func(context.Context, Send[int]) error {
		ran = true
		return nil
	}, nil)
	if single == nil {
		t.Fatal("single-command batch was nil")
	}
	if err := single(context.Background(), func(int) {}); err != nil || !ran {
		t.Fatalf("single-command batch err=%v ran=%v", err, ran)
	}
}

func TestStartCmdReportsReturnedError(t *testing.T) {
	want := errors.New("load failed")
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, Send[int]) error {
		return want
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) {
		t.Fatalf("reported error type = %T, want *EffectError", err)
	}
	if effectErr.Kind != EffectCommand || effectErr.Key != "" || !errors.Is(effectErr, want) {
		t.Fatalf("reported command error = %#v", effectErr)
	}
}

func TestStartCmdReportsContextCanceledWhileEffectIsActive(t *testing.T) {
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, Send[int]) error {
		return context.Canceled
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) || !errors.Is(effectErr, context.Canceled) {
		t.Fatalf("reported cancellation = %v, want active command error", err)
	}
}

func TestStartCmdRecoversPanic(t *testing.T) {
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, Send[int]) error {
		panic("broken command")
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) {
		t.Fatalf("reported error type = %T, want *EffectError", err)
	}
	if effectErr.Kind != EffectCommand || effectErr.Panic != "broken command" {
		t.Fatalf("reported panic = %#v", effectErr)
	}
	if len(effectErr.Stack) == 0 {
		t.Fatal("panic report has no stack trace")
	}
}

func TestStartCmdCancelsAndDropsLateMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	finished := make(chan struct{})
	sent := make(chan int, 2)
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, ctx, func(ctx context.Context, send Send[int]) error {
		close(started)
		<-ctx.Done()
		send(1)
		close(finished)
		return ctx.Err()
	}, func(msg int) {
		sent <- msg
	}, func(err error) {
		reported <- err
	})

	receiveRuntimeTestValue(t, started)
	cancel()
	receiveRuntimeTestValue(t, finished)
	select {
	case msg := <-sent:
		t.Fatalf("canceled command sent late message %d", msg)
	default:
	}
	select {
	case err := <-reported:
		t.Fatalf("canceled command reported error %v", err)
	default:
	}
	if !effects.waitFor(time.Second) {
		t.Fatal("canceled command did not stop")
	}
}

func TestLatestCmdCancelsOlderGenerationAndDropsItsMessages(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	started := make(chan struct{})
	canceled := make(chan struct{})
	sent := make(chan int, 2)
	var effects effectGroup
	old := LatestCmd("search", func(ctx context.Context, send Send[int]) error {
		close(started)
		<-ctx.Done()
		send(1)
		close(canceled)
		return ctx.Err()
	})
	newer := LatestCmd("search", func(_ context.Context, send Send[int]) error {
		send(2)
		return nil
	})
	StartCmd(&effects, ctx, old, func(msg int) { sent <- msg }, func(error) {})
	receiveRuntimeTestValue(t, started)
	StartCmd(&effects, ctx, newer, func(msg int) { sent <- msg }, func(error) {})
	receiveRuntimeTestValue(t, canceled)
	if got := receiveRuntimeTestValue(t, sent); got != 2 {
		t.Fatalf("latest command message = %d, want 2", got)
	}
	select {
	case got := <-sent:
		t.Fatalf("stale command message = %d", got)
	default:
	}
	if !effects.waitFor(time.Second) {
		t.Fatal("latest commands did not stop")
	}
}

func TestLatestCmdCanBeReusedAcrossExecutions(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	sent := make(chan int, 2)
	runs := 0
	reused := LatestCmd("preview", func(_ context.Context, send Send[int]) error {
		runs++
		send(runs)
		return nil
	})
	if err := reused(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("first command error: %v", err)
	}
	if err := reused(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("reused command error: %v", err)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 1 {
		t.Fatalf("first latest command message = %d, want 1", got)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 2 {
		t.Fatalf("reused latest command message = %d, want 2", got)
	}
}

func TestLatestCmdUsesDispatchOrderDespiteDelayedExecution(t *testing.T) {
	for _, name := range []string{"latest", "cancel", "batch", "mapped"} {
		t.Run(name, func(t *testing.T) {
			ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
			release := make(chan struct{})
			sent := make(chan int, 4)
			var effects effectGroup
			old := LatestCmd("search", func(_ context.Context, send Send[int]) error {
				send(1)
				return nil
			})
			if name == "cancel" {
				old = CancelLatestCmd[int]("search")
			}
			if name == "batch" {
				old = Batch(old, func(context.Context, Send[int]) error { return nil })
			}
			if name == "mapped" {
				old = MapCmd(old, func(value int) int { return value })
			}
			StartCmd(&effects, ctx, func(ctx context.Context, send Send[int]) error {
				<-release
				return old(ctx, send)
			}, func(value int) { sent <- value }, nil)
			newDone := make(chan struct{})
			StartCmd(&effects, ctx, LatestCmd("search", func(_ context.Context, send Send[int]) error {
				send(2)
				close(newDone)
				return nil
			}), func(value int) { sent <- value }, nil)
			receiveRuntimeTestValue(t, newDone)
			close(release)
			if !effects.waitFor(time.Second) {
				t.Fatal("commands did not finish")
			}
			if got := receiveRuntimeTestValue(t, sent); got != 2 {
				t.Fatalf("result = %d, want newer result 2", got)
			}
			if len(sent) != 0 {
				t.Fatal("older command sent a result after the newer command finished")
			}
		})
	}
}

func TestMapCmdPreservesQueuedLatestValidation(t *testing.T) {
	manager := newLatestCommandManager()
	ctx := withLatestCommandManager(context.Background(), manager)
	var tokens []latestCommandToken
	var values []string
	ctx = context.WithValue(ctx, latestDispatchKey{}, func(token latestCommandToken, value string) {
		tokens = append(tokens, token)
		values = append(values, value)
	})
	cmd := MapCmd(LatestCmd("search", func(_ context.Context, send Send[int]) error {
		send(7)
		return nil
	}), func(value int) string { return fmt.Sprint(value) })
	for range 2 {
		if err := cmd(ctx, func(string) { t.Fatal("mapped latest result bypassed validation") }); err != nil {
			t.Fatal(err)
		}
	}
	if len(tokens) != 2 || values[1] != "7" || manager.accepts(tokens[0]) || !manager.accepts(tokens[1]) {
		t.Fatalf("mapped results did not retain generation: tokens=%v values=%v", tokens, values)
	}
}

func TestBatchLatestCommandsFollowSliceOrder(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	release := make(chan struct{})
	sent := make(chan int, 2)
	older := LatestCmd("search", func(_ context.Context, send Send[int]) error {
		send(1)
		return nil
	})
	newer := LatestCmd("search", func(_ context.Context, send Send[int]) error {
		send(2)
		close(release)
		return nil
	})
	batch := Batch(func(ctx context.Context, send Send[int]) error {
		<-release
		return older(ctx, send)
	}, Batch(newer, func(context.Context, Send[int]) error { return nil }))
	if err := batch(ctx, func(value int) { sent <- value }); err != nil {
		t.Fatal(err)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 2 || len(sent) != 0 {
		t.Fatalf("batch result=%d, additional messages=%d", got, len(sent))
	}
}

func TestLatestCmdFinishCancelsChildContext(t *testing.T) {
	parent := &afterFuncTestContext{Context: context.Background(), done: make(chan struct{})}
	manager := newLatestCommandManager()
	child, token, ok := manager.start(parent, "preview")
	if !ok {
		t.Fatal("LatestCmd did not start")
	}
	manager.finish(token)
	if child == nil || child.Err() != context.Canceled {
		t.Fatalf("finished child context = %v, want canceled", child)
	}
	if !parent.stopped {
		t.Fatal("finished child context was not removed from parent")
	}
}

type afterFuncTestContext struct {
	context.Context
	done    chan struct{}
	stopped bool
}

func (c *afterFuncTestContext) Done() <-chan struct{} { return c.done }

func (c *afterFuncTestContext) AfterFunc(func()) func() bool {
	return func() bool {
		c.stopped = true
		return true
	}
}

func TestLatestCmdGenerationIsKeyLocalAfterHighChurn(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	sent := make(chan int, 2)
	reused := LatestCmd("preview", func(_ context.Context, send Send[int]) error {
		send(7)
		return nil
	})
	if err := reused(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("first preview command error: %v", err)
	}
	for index := range 300 {
		cmd := LatestCmd(fmt.Sprintf("churn-%d", index), func(context.Context, Send[int]) error { return nil })
		if err := cmd(ctx, func(int) {}); err != nil {
			t.Fatalf("churn command %d error: %v", index, err)
		}
	}
	if err := reused(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("reused preview command error: %v", err)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 7 {
		t.Fatalf("first preview message = %d, want 7", got)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 7 {
		t.Fatalf("reused preview message = %d, want 7", got)
	}
}

func TestCancelLatestCmdStopsActiveGeneration(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	started := make(chan struct{})
	canceled := make(chan struct{})
	var effects effectGroup
	active := LatestCmd("load", func(ctx context.Context, _ Send[int]) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return ctx.Err()
	})
	cancel := CancelLatestCmd[int]("load")
	StartCmd(&effects, ctx, active, func(int) {}, nil)
	receiveRuntimeTestValue(t, started)
	StartCmd(&effects, ctx, cancel, func(int) {}, nil)
	receiveRuntimeTestValue(t, canceled)
	if !effects.waitFor(time.Second) {
		t.Fatal("canceled latest command did not stop")
	}
}

func TestLatestCmdInvalidatesQueuedResultsAtDispatch(t *testing.T) {
	for _, action := range []string{"cancel", "replace"} {
		for _, wrapper := range []string{"direct", "batch", "mapped", "nested"} {
			t.Run(action+"/"+wrapper, func(t *testing.T) {
				previous := goruntime.GOMAXPROCS(1)
				defer goruntime.GOMAXPROCS(previous)
				core := newLoopCore(0, func(model *int, msg int) Cmd[int] {
					if msg != 1 {
						*model += msg
						return nil
					}
					cmd := CancelLatestCmd[int]("search")
					if action == "replace" {
						cmd = LatestCmd("search", func(context.Context, Send[int]) error { return nil })
					}
					if wrapper == "batch" || wrapper == "nested" {
						cmd = Batch(cmd, func(context.Context, Send[int]) error { return nil })
					}
					if wrapper == "mapped" || wrapper == "nested" {
						cmd = MapCmd(cmd, func(value int) int { return value })
					}
					return cmd
				})
				ctx := withLatestCommandManager(context.Background(), core.latest)
				ctx = context.WithValue(ctx, latestDispatchKey{}, func(token latestCommandToken, value int) {
					core.sendLatest(token, value)
				})
				core.send(1)
				for _, item := range []struct {
					key   string
					value int
				}{{"search", 99}, {"other", 7}} {
					cmd := LatestCmd(item.key, func(_ context.Context, send Send[int]) error {
						send(item.value)
						return nil
					})
					if err := cmd(ctx, core.send); err != nil {
						t.Fatal(err)
					}
				}
				var effects effectGroup
				_, _, err := core.updateMessages(&effects, ctx, core.send, nil, nil)
				if !effects.waitFor(time.Second) {
					t.Fatal("commands did not finish")
				}
				if err != nil || core.model != 7 {
					t.Fatalf("model=%d err=%v; want only unrelated result 7", core.model, err)
				}
			})
		}
	}
}

func TestStartCmdPreparationLeavesApplicationWorkAsynchronous(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	release := make(chan struct{})
	sent := make(chan int, 1)
	work := MapCmd(Batch(
		LatestCmd("search", func(_ context.Context, send Send[string]) error {
			<-release
			send("result")
			return nil
		}),
		func(context.Context, Send[string]) error { return nil },
	), func(value string) int { return len(value) })
	var effects effectGroup
	StartCmd(&effects, ctx, work, func(value int) { sent <- value }, nil)
	close(release)
	if !effects.waitFor(time.Second) {
		t.Fatal("application command did not finish")
	}
	if got := receiveRuntimeTestValue(t, sent); got != 6 {
		t.Fatalf("mapped result=%d, want 6", got)
	}
}

func TestEffectGroupWaitIsBounded(t *testing.T) {
	release := make(chan struct{})
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, Send[int]) error {
		<-release
		return nil
	}, func(int) {}, nil)
	if effects.waitFor(10 * time.Millisecond) {
		t.Fatal("effect group reported idle while command was still running")
	}
	close(release)
	if !effects.waitFor(time.Second) {
		t.Fatal("effect group did not become idle after command completed")
	}
}

func receiveRuntimeTestValue[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for runtime event")
		var zero T
		return zero
	}
}

func TestStartCmdSendsMessage(t *testing.T) {
	done := make(chan int, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(_ context.Context, send Send[int]) error {
		send(7)
		return nil
	}, func(msg int) {
		done <- msg
	}, nil)

	select {
	case got := <-done:
		if got != 7 {
			t.Fatalf("message = %d, want 7", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command")
	}
}
