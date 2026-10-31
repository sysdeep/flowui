package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestTextInputFocusAndChange(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	value, changes := "", 0
	build := func() ui.Widget {
		return base.TextInput("query", value).
			Hint("Search").
			OnChange(func(next string) { value = next; changes++ }).
			Content(func(state base.TextInputState, editor *widget.Editor) ui.Widget {
				return layout.Box(testEditor(editor)).Style(style.New().Height(32).Padding(8).Radius(6))
			})
	}
	h.Frame(build())
	tag := core.FieldFocusTag(h.Context(), "query")
	h.Router().Source().Execute(key.FocusCmd{Tag: tag})
	h.Frame(build())
	h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: 0}, Text: "abc"})
	h.Frame(build())
	if value != "abc" || changes != 1 {
		t.Fatalf("value=%q changes=%d", value, changes)
	}
}

func TestTextInputDisabledIgnoresEditing(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	changes := 0
	view := base.TextInput("query", "").
		Disabled(true).
		OnChange(func(string) { changes++ }).
		Content(func(state base.TextInputState, editor *widget.Editor) ui.Widget {
			return layout.Box(testEditor(editor)).Style(style.New().Height(32))
		})
	h.Frame(view)
	h.Click(f32.Pt(40, 16))
	h.Frame(view)
	if changes != 0 {
		t.Fatalf("disabled changes=%d", changes)
	}
}

func TestTextInputSynchronizesFocusedValue(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	value := "initial"
	var observed base.TextInputState
	build := func() ui.Widget {
		return base.TextInput("query", value).
			OnChange(func(next string) { value = next }).
			Content(func(state base.TextInputState, editor *widget.Editor) ui.Widget {
				observed = state
				return layout.Box(testEditor(editor)).Style(style.New().Height(32))
			})
	}
	h.Frame(build())
	h.Router().Source().Execute(key.FocusCmd{Tag: core.FieldFocusTag(h.Context(), "query")})
	h.Frame(build())
	value = "external"
	h.Frame(build())
	if !observed.Focused || observed.Text != value || observed.Changed {
		t.Fatalf("external synchronization = %+v", observed)
	}
	h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: len(value)}, Text: "local"})
	h.Frame(build())
	if !observed.Changed || observed.Text != "local" || value != "local" {
		t.Fatalf("local edit state = %+v, value = %q", observed, value)
	}
}

func TestTextInputDisabledSkipsKeyHandler(t *testing.T) {
	h := uitest.New(image.Pt(240, 80), theme.Light())
	calls := 0
	h.Frame(base.TextInput("query", "").Disabled(true).
		OnKey(func(*core.Context, giolayout.Context, event.Tag) { calls++ }).
		Content(func(_ base.TextInputState, editor *widget.Editor) ui.Widget { return testEditor(editor) }))
	if calls != 0 {
		t.Fatalf("disabled key handler calls = %d", calls)
	}
}

// Tests deliberately provide their own editor paint instead of relying on base defaults.
func testEditor(editor *widget.Editor) ui.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
		return editor.Layout(gtx, ctx.TextShaper(), font.Font{}, 14, op.CallOp{}, op.CallOp{})
	})
}
