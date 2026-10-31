package base_test

import (
	"image"
	"reflect"
	"testing"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestTextEditorChangeBeforeSubmit(t *testing.T) {
	h := uitest.New(image.Pt(240, 40), theme.Light())
	var editor widget.Editor
	var got []string
	value := ""
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		(base.TextEditor{SingleLine: true}).Update(ctx, gtx, base.TextEditorRequest{FieldKey: "name", Editor: &editor, Value: value,
			OnChange: func(text string) {
				got = append(got, "change:"+text)
				value = "callback value"
			},
			OnSubmit: func(text string) { got = append(got, "submit:"+text) },
		})
		return editor.Layout(gtx, ctx.TextShaper(), font.Font{}, 14, op.CallOp{}, op.CallOp{})
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &editor})
	h.Frame(view)
	h.Router().Queue(key.EditEvent{Range: key.Range{}, Text: "Ada"})
	h.Key(key.NameReturn, 0)
	h.Frame(view)
	want := []string{"change:Ada", "submit:Ada"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("callbacks = %v, want %v", got, want)
	}
}

func TestTextEditorEditingPolicy(t *testing.T) {
	for _, tc := range []struct {
		name     string
		disabled bool
		readOnly bool
	}{
		{name: "disabled", disabled: true},
		{name: "read only", readOnly: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := uitest.New(image.Pt(240, 40), theme.Light())
			var editor widget.Editor
			restricted := false
			changes, pendingEdits := 0, 0
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if restricted && tc.disabled {
					gtx = gtx.Disabled()
				}
				(base.TextEditor{SingleLine: true, ReadOnly: restricted && tc.readOnly}).Update(
					ctx, gtx, base.TextEditorRequest{
						FieldKey: "name", Editor: &editor, Value: "Ada",
						BeforeUpdate: func(layout.Context) { pendingEdits++ },
						OnChange:     func(string) { changes++ },
					},
				)
				return editor.Layout(gtx, ctx.TextShaper(), font.Font{}, 14, op.CallOp{}, op.CallOp{})
			})
			h.Frame(view)
			h.Router().Source().Execute(key.FocusCmd{Tag: &editor})
			h.Frame(view)
			pendingEdits = 0
			restricted = true
			h.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: 3}, Text: "Grace"})
			h.Frame(view)
			if editor.Text() != "Ada" || changes != 0 {
				t.Fatalf("restricted edit = %q, changes = %d", editor.Text(), changes)
			}
			if tc.disabled && pendingEdits != 0 {
				t.Fatalf("disabled pending edits = %d", pendingEdits)
			}
		})
	}
}
