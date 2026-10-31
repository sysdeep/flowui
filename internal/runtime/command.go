package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"slices"
	"sync"
	"time"
)

// EffectKind identifies asynchronous work managed by the runtime.
type EffectKind string

const (
	EffectCommand      EffectKind = "command"
	EffectSubscription EffectKind = "subscription"
)

// EffectError describes an error or panic produced by asynchronous work.
type EffectError struct {
	Kind  EffectKind
	Key   string
	Err   error
	Panic any
	Stack []byte
}

func (e *EffectError) Error() string {
	source := string(e.Kind)
	if e.Key != "" {
		source += fmt.Sprintf(" %q", e.Key)
	}
	if e.Panic != nil {
		return fmt.Sprintf("%s panicked: %v", source, e.Panic)
	}
	return fmt.Sprintf("%s failed: %v", source, e.Err)
}

func (e *EffectError) Unwrap() error {
	return e.Err
}

// Send delivers a message to the runtime, which queues it for the next frame.
// It is safe for concurrent effects to call Send. The queue is bounded; a
// burst beyond that bound drops the newest messages and reports
// QueueOverflowError, so a sender must not assume unlimited buffering.
type Send[Msg any] func(Msg)

// Cmd is asynchronous work started by Update, Init, or a Subscription. It runs
// after the callback that returned it and may send messages back later. Its
// context is canceled when the window closes.
type Cmd[Msg any] func(context.Context, Send[Msg]) error

type latestCommandKey struct{}
type latestDispatchKey struct{}
type commandOrderKey struct{}

type commandPreparationKey struct{}

// Only framework wrappers participate in preparation; application commands must
// never execute on the dispatch goroutine. Store code addresses, not closures,
// so this registry does not retain individual commands or their captures.
var preparableCommands sync.Map

func preparableCmd[Msg any](prepare func(context.Context) Cmd[Msg]) Cmd[Msg] {
	cmd := Cmd[Msg](func(ctx context.Context, send Send[Msg]) error {
		if result, ok := ctx.Value(commandPreparationKey{}).(*Cmd[Msg]); ok {
			*result = prepare(context.WithValue(ctx, commandPreparationKey{}, nil))
			return nil
		}
		run := prepare(ctx)
		return run(ctx, send)
	})
	preparableCommands.Store(reflect.ValueOf(cmd).Pointer(), struct{}{})
	return cmd
}

func prepareCmd[Msg any](ctx context.Context, cmd Cmd[Msg]) Cmd[Msg] {
	if _, ok := preparableCommands.Load(reflect.ValueOf(cmd).Pointer()); !ok {
		return cmd
	}
	var prepared Cmd[Msg]
	_ = cmd(context.WithValue(ctx, commandPreparationKey{}, &prepared), nil)
	return prepared
}

type latestCommandToken struct {
	key        string
	generation uint64
}

type latestCommandState struct {
	generation uint64
	cancel     context.CancelFunc
	order      []uint64
}

type latestCommandManager struct {
	mu       sync.Mutex
	states   map[string]latestCommandState
	sequence uint64
}

func (m *latestCommandManager) orderedContext(ctx context.Context) context.Context {
	m.mu.Lock()
	m.sequence++
	order := []uint64{m.sequence}
	m.mu.Unlock()
	return context.WithValue(ctx, commandOrderKey{}, order)
}

func newLatestCommandManager() *latestCommandManager {
	return &latestCommandManager{states: make(map[string]latestCommandState)}
}

func withLatestCommandManager(ctx context.Context, manager *latestCommandManager) context.Context {
	return context.WithValue(ctx, latestCommandKey{}, manager)
}

func (m *latestCommandManager) start(ctx context.Context, key string) (context.Context, latestCommandToken, bool) {
	order, _ := ctx.Value(commandOrderKey{}).([]uint64)
	if order == nil {
		order = m.orderedContext(ctx).Value(commandOrderKey{}).([]uint64)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.states[key]
	if slices.Compare(order, current.order) < 0 {
		return nil, latestCommandToken{}, false
	}
	if current.cancel != nil {
		current.cancel()
	}
	generation := current.generation + 1
	if generation == 0 {
		// Reserve zero for the manager's untouched state if a process ever
		// executes more than MaxUint64 latest commands for one key.
		generation = 1
	}
	child, cancel := context.WithCancel(ctx)
	m.states[key] = latestCommandState{generation: generation, cancel: cancel, order: order}
	return child, latestCommandToken{key: key, generation: generation}, true
}

func (m *latestCommandManager) accepts(token latestCommandToken) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.states[token.key]
	return ok && current.generation == token.generation
}

func (m *latestCommandManager) finish(token latestCommandToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.states[token.key]
	if ok && current.generation == token.generation {
		// Keep a generation tombstone so older LatestCmd closures cannot start
		// after a newer generation has finished. Cancel the child before
		// releasing its cancel function from the state.
		if current.cancel != nil {
			current.cancel()
		}
		current.cancel = nil
		m.states[token.key] = current
	}
}

// LatestCmd cancels an older command with the same key and drops messages sent
// by an older generation. Outside a FlowUI runtime it behaves like cmd.
//
// Runtime dispatch order determines which execution is newer, independent of
// goroutine scheduling. The same Cmd value may be reused across Update calls.
func LatestCmd[Msg any](key string, cmd Cmd[Msg]) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	if cmd == nil {
		return nil
	}
	return preparableCmd(func(ctx context.Context) Cmd[Msg] {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager == nil {
			return cmd
		}
		child, token, ok := manager.start(ctx, key)
		if !ok {
			return func(context.Context, Send[Msg]) error { return nil }
		}
		run := prepareCmd(child, cmd)
		return func(_ context.Context, send Send[Msg]) error {
			defer manager.finish(token)
			dispatch, _ := ctx.Value(latestDispatchKey{}).(func(latestCommandToken, Msg))
			err := run(child, func(msg Msg) {
				if !manager.accepts(token) || child.Err() != nil {
					return
				}
				if dispatch != nil {
					dispatch(token, msg)
					return
				}
				send(msg)
			})
			if child.Err() != nil {
				return nil
			}
			return err
		}
	})
}

// CancelLatestCmd cancels the active command with key and invalidates its
// queued messages.
//
// Cancellation follows runtime dispatch order, like LatestCmd.
func CancelLatestCmd[Msg any](key string) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	return preparableCmd(func(ctx context.Context) Cmd[Msg] {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager != nil {
			_, token, ok := manager.start(ctx, key)
			if ok {
				manager.finish(token)
			}
		}
		return func(context.Context, Send[Msg]) error { return nil }
	})
}

// Batch combines commands so Update can start several independent effects at
// once. Nil commands are ignored. Remaining commands run concurrently after
// Update returns, share the effect context, and may each call send. The first
// non-nil error is returned after every command finishes; a panic in any
// command is re-raised so the runtime reports it like a single command panic.
func Batch[Msg any](cmds ...Cmd[Msg]) Cmd[Msg] {
	active := make([]Cmd[Msg], 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			active = append(active, cmd)
		}
	}
	switch len(active) {
	case 0:
		return nil
	case 1:
		return active[0]
	}
	return preparableCmd(func(ctx context.Context) Cmd[Msg] {
		if ctx.Value(commandOrderKey{}) == nil {
			if manager, ok := ctx.Value(latestCommandKey{}).(*latestCommandManager); ok {
				ctx = manager.orderedContext(ctx)
			}
		}
		type execution struct {
			ctx context.Context
			cmd Cmd[Msg]
		}
		executions := make([]execution, len(active))
		for index, cmd := range active {
			child := ctx
			if order, ok := ctx.Value(commandOrderKey{}).([]uint64); ok {
				child = context.WithValue(ctx, commandOrderKey{}, append(slices.Clone(order), uint64(index)))
			}
			executions[index] = execution{child, prepareCmd(child, cmd)}
		}
		return func(_ context.Context, send Send[Msg]) error {
			var (
				wg       sync.WaitGroup
				mu       sync.Mutex
				firstErr error
				panicked any
			)
			for _, execution := range executions {
				wg.Go(func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							mu.Lock()
							if panicked == nil {
								panicked = recovered
							}
							mu.Unlock()
						}
					}()
					err := execution.cmd(execution.ctx, send)
					if err == nil || ctx.Err() != nil {
						return
					}
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				})
			}
			wg.Wait()
			if panicked != nil {
				panic(panicked)
			}
			return firstErr
		}
	})
}

// ErrEffectShutdownTimeout reports asynchronous work that did not stop within
// the runtime's bounded shutdown period.
var ErrEffectShutdownTimeout = fmt.Errorf("timed out waiting for asynchronous effects to stop")

type effectGroup struct {
	wait sync.WaitGroup
}

func StartCmd[Msg any](group *effectGroup, ctx context.Context, cmd Cmd[Msg], send func(Msg), report func(error)) {
	if cmd != nil {
		if manager, ok := ctx.Value(latestCommandKey{}).(*latestCommandManager); ok {
			ctx = manager.orderedContext(ctx)
		}
		cmd = prepareCmd(ctx, cmd)
		startEffect(group, ctx, EffectCommand, "", cmd, send, report, nil)
	}
}

// MapCmd preserves latest-result validation when adapting message types.
func MapCmd[ChildMsg any, ParentMsg any](cmd Cmd[ChildMsg], mapMsg func(ChildMsg) ParentMsg) Cmd[ParentMsg] {
	if cmd == nil {
		return nil
	}
	if mapMsg == nil {
		panic("flowui: nil command mapper")
	}
	return preparableCmd(func(ctx context.Context) Cmd[ParentMsg] {
		if dispatch, ok := ctx.Value(latestDispatchKey{}).(func(latestCommandToken, ParentMsg)); ok {
			ctx = context.WithValue(ctx, latestDispatchKey{}, func(token latestCommandToken, msg ChildMsg) {
				dispatch(token, mapMsg(msg))
			})
		}
		run := prepareCmd(ctx, cmd)
		return func(_ context.Context, send Send[ParentMsg]) error {
			return run(ctx, func(msg ChildMsg) { send(mapMsg(msg)) })
		}
	})
}

func startEffect[Msg any](
	group *effectGroup,
	ctx context.Context,
	kind EffectKind,
	key string,
	run Cmd[Msg],
	send Send[Msg],
	report func(error),
	onDone func(),
) <-chan struct{} {
	done := make(chan struct{})
	group.wait.Go(func() {
		defer func() {
			close(done)
			if onDone != nil {
				onDone()
			}
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				reportEffectError(report, &EffectError{
					Kind:  kind,
					Key:   key,
					Panic: recovered,
					Stack: debug.Stack(),
				})
			}
		}()

		err := run(ctx, func(msg Msg) {
			if ctx.Err() == nil {
				send(msg)
			}
		})
		if err == nil || ctx.Err() != nil {
			return
		}
		reportEffectError(report, &EffectError{Kind: kind, Key: key, Err: err})
	})
	return done
}

func reportEffectError(report func(error), err error) {
	if report != nil && err != nil {
		report(err)
	}
}

func (g *effectGroup) waitFor(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		g.wait.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
