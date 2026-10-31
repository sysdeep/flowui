package base

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
)

// EscapePressed drains Escape for a mounted focus target and reports a close
// intent. It does not move focus, own open state or listen globally.
func EscapePressed(gtx layout.Context, target event.Tag) bool {
	pressed := false
	for {
		e, ok := gtx.Event(key.Filter{Focus: target, Name: key.NameEscape})
		if !ok {
			return pressed
		}
		if press, ok := e.(key.Event); ok && press.State == key.Press && gtx.Enabled() {
			pressed = true
		}
	}
}
