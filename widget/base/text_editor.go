package base

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
)

// TextEditor configures editor behavior without prescribing its typography,
// materials, placeholder or surrounding field layout.
type TextEditor struct {
	SingleLine bool
	ReadOnly   bool
	MaxLength  int
	Mask       rune
	InputHint  key.InputHint
	Filter     string
}

// TextEditorRequest supplies an editor identity, value, pending edits and
// callbacks for one frame. Disabled state comes only from the layout context.
type TextEditorRequest struct {
	FieldKey     string
	Editor       *widget.Editor
	Value        string
	BeforeUpdate func(layout.Context)
	OnChange     func(string)
	OnSubmit     func(string)
}

// TextEditorEvents preserves change-before-submit ordering and the submitted
// text snapshot even if a callback changes the current model value.
type TextEditorEvents struct {
	Changed    bool
	Submitted  bool
	SubmitText string
}

func (e *TextEditorEvents) Add(event widget.EditorEvent) {
	switch event := event.(type) {
	case widget.ChangeEvent:
		e.Changed = true
	case widget.SubmitEvent:
		e.Submitted, e.SubmitText = true, event.Text
	}
}

// Update registers focus, applies pending edits, drains editor events, then
// synchronizes the controlled value only when no local edit took precedence.
// BeforeUpdate applies queued clipboard/context-menu actions when enabled.
func (t TextEditor) Update(ctx *core.Context, gtx layout.Context, request TextEditorRequest) TextEditorEvents {
	enabled := gtx.Enabled()
	core.RegisterFieldFocus(ctx, request.FieldKey, request.Editor, enabled)
	editor := request.Editor
	editor.SingleLine, editor.Submit = t.SingleLine, t.SingleLine
	editor.ReadOnly, editor.MaxLen = t.ReadOnly, t.MaxLength
	editor.Mask, editor.InputHint, editor.Filter = t.Mask, t.InputHint, t.Filter
	if enabled && request.BeforeUpdate != nil {
		request.BeforeUpdate(gtx)
	}
	var events TextEditorEvents
	for {
		event, ok := editor.Update(gtx)
		if !ok {
			break
		}
		events.Add(event)
	}
	if events.Changed && request.OnChange != nil {
		request.OnChange(editor.Text())
	}
	if events.Submitted && request.OnSubmit != nil {
		request.OnSubmit(events.SubmitText)
	}
	if !events.Changed && !events.Submitted && editor.Text() != request.Value {
		editor.SetText(request.Value)
	}
	return events
}
