package base

import (
	"gioui.org/io/event"
	giokey "gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

// TextInputState is the state a TextInput reports to its Content callback.
type TextInputState struct {
	Interaction
	Key     string
	Text    string
	Hint    string
	Changed bool
}

// TextInputWidget owns an editor's focus, IME, selection, controlled text
// synchronization and disabled/read-only behavior. Content draws the field and
// lays out the provided Gio editor with caller-supplied typography and materials.
type TextInputWidget struct {
	key        string
	value      string
	hint       string
	disabled   bool
	readOnly   bool
	invalid    bool
	singleLine bool
	onChange   func(string)
	onKey      func(*core.Context, layout.Context, event.Tag)
	content    func(TextInputState, *widget.Editor) core.Widget
}

func TextInput(key, value string) TextInputWidget {
	return TextInputWidget{key: key, value: value, singleLine: true}
}

func (t TextInputWidget) Hint(value string) TextInputWidget        { t.hint = value; return t }
func (t TextInputWidget) Disabled(value bool) TextInputWidget      { t.disabled = value; return t }
func (t TextInputWidget) ReadOnly(value bool) TextInputWidget      { t.readOnly = value; return t }
func (t TextInputWidget) Invalid(value bool) TextInputWidget       { t.invalid = value; return t }
func (t TextInputWidget) Multiline() TextInputWidget               { t.singleLine = false; return t }
func (t TextInputWidget) OnChange(fn func(string)) TextInputWidget { t.onChange = fn; return t }
func (t TextInputWidget) OnKey(fn func(*core.Context, layout.Context, event.Tag)) TextInputWidget {
	t.onKey = fn
	return t
}

func (t TextInputWidget) Content(fn func(TextInputState, *widget.Editor) core.Widget) TextInputWidget {
	t.content = fn
	return t
}

func (t TextInputWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if t.content == nil {
		panic("flowui: base.TextInput requires Content")
	}
	key := core.ClaimKey(ctx, core.KindBaseTextInput, t.key)
	state := core.UseState[textInputState](ctx, key, "base-text-input")
	editor := &state.editor
	disabled := t.disabled || !gtx.Enabled()
	state.interaction.Update(ctx, gtx, disabled, editor)
	if !disabled && t.onKey != nil {
		t.onKey(ctx, gtx, editor)
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	events := (TextEditor{SingleLine: t.singleLine, ReadOnly: t.readOnly, InputHint: giokey.HintText}).Update(
		ctx, gtx, TextEditorRequest{FieldKey: key, Editor: editor, Value: t.value, OnChange: t.onChange})
	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)

	public := TextInputState{
		Interaction: Interaction{
			Focused:      focused,
			FocusVisible: focusVisible,
			Disabled:     disabled,
			ReadOnly:     t.readOnly,
			Invalid:      t.invalid,
		},
		Key:     key,
		Text:    editor.Text(),
		Hint:    t.hint,
		Changed: events.Changed,
	}
	content := t.content(public, editor)
	if content == nil {
		panic("flowui: base.TextInput Content returned nil")
	}
	dims := content.Layout(ctx, gtx)
	state.interaction.AddPointer(gtx, dims.Size, disabled)
	return dims
}

type textInputState struct {
	editor      widget.Editor
	interaction FieldInteraction
}
