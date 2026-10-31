package ui

// Handler handles one concrete message type for a Model and message set.
// Create handlers with On and pass them to Match.
type Handler[M any, Msg any] struct {
	try func(*M, Msg) (Cmd[Msg], bool)
}

// On creates a handler for one concrete message type. The model and message
// types are inferred from fn.
func On[M any, Msg any, Event any](fn func(*M, Event) Cmd[Msg]) Handler[M, Msg] {
	if fn == nil {
		panic("flowui: nil message handler")
	}
	return Handler[M, Msg]{
		try: func(model *M, msg Msg) (Cmd[Msg], bool) {
			event, ok := any(msg).(Event)
			if !ok {
				return nil, false
			}
			return fn(model, event), true
		},
	}
}

// Match dispatches msg to the first matching handler and returns its command.
// An unmatched message returns nil. Ordinary Go type switches remain a good
// choice for small update functions.
func Match[M any, Msg any](model *M, msg Msg, handlers ...Handler[M, Msg]) Cmd[Msg] {
	for _, handler := range handlers {
		if handler.try == nil {
			continue
		}
		if cmd, matched := handler.try(model, msg); matched {
			return cmd
		}
	}
	return nil
}
