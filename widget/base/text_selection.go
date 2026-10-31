package base

import (
	"image"
	"io"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
)

// SelectionAction is a read-only text action, independent of any menu visual.
type SelectionAction uint8

const (
	CopySelection SelectionAction = iota + 1
	SelectAll
)

// SelectableText retains a Gio selection and applies queued menu actions.
// The caller lays out Selectable with its own typography and selection paint.
type SelectableText struct {
	Selectable widget.Selectable
	pressTag   struct{ _ byte }
	pending    SelectionAction
}

func (s *SelectableText) Queue(action SelectionAction) { s.pending = action }

func (s *SelectableText) Update(ctx *core.Context, gtx layout.Context) {
	if gtx.Enabled() {
		switch s.pending {
		case CopySelection:
			copySelection(gtx, s.Selectable.SelectedText())
		case SelectAll:
			s.Selectable.SetCaret(0, len([]rune(s.Selectable.Text())))
		}
	}
	s.pending = 0
	for {
		_, ok := core.NextPointerEvent(gtx, &s.pressTag, pointer.Press)
		if !ok {
			return
		}
		if gtx.Enabled() {
			core.PreserveFocus(ctx)
		}
	}
}

// AddPointer observes presses without intercepting the Gio selectable's input.
func (s *SelectableText) AddPointer(gtx layout.Context, size image.Point) {
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &s.pressTag)
	pass.Pop()
	area.Pop()
}

// TextSelection handles dragging and copying a source-rune range for custom
// text renderers. Hit testing, wrapping and selection painting stay with them.
type TextSelection struct {
	drag          DragInteraction
	anchor, caret int
}

func (s *TextSelection) Range() (start, end int, selected bool) {
	return min(s.anchor, s.caret), max(s.anchor, s.caret), s.anchor != s.caret
}

func (s *TextSelection) Update(ctx *core.Context, gtx layout.Context, target event.Tag, source []rune, hit func(f32.Point) int) {
	for {
		if _, ok := gtx.Event(key.FocusFilter{Target: target}); !ok {
			break
		}
	}
	s.anchor = min(s.anchor, len(source))
	s.caret = min(s.caret, len(source))
	s.drag.Update(ctx, gtx, target, true, func(position f32.Point, started bool) {
		index := min(max(hit(position), 0), len(source))
		if started {
			s.anchor = index
		}
		s.caret = index
	})
	start, end, selected := s.Range()
	if !selected {
		return
	}
	for {
		e, ok := gtx.Event(key.Filter{Focus: target, Name: "C", Required: key.ModShortcut})
		if !ok {
			return
		}
		if press, ok := e.(key.Event); ok && press.State == key.Press && gtx.Enabled() {
			copySelection(gtx, string(source[start:end]))
		}
	}
}

func copySelection(gtx layout.Context, text string) {
	if text != "" {
		gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
	}
}
