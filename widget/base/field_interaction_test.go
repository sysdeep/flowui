package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestFieldInteractionDisabledClearsHover(t *testing.T) {
	state := &base.FieldInteraction{Hovered: true}
	active := theme.Light()
	state.Update(core.New(nil, &active, locale.LanguageAuto), fieldTestContext(), true, new(int))
	if state.Hovered {
		t.Fatal("disabled field kept hover state")
	}
}

func TestFieldInteractionHover(t *testing.T) {
	h := uitest.New(image.Pt(100, 60), theme.Light())
	var state base.FieldInteraction
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		state.Update(ctx, gtx, false, nil)
		state.AddPointer(gtx, image.Pt(80, 40), false)
		return layout.Dimensions{Size: image.Pt(80, 40)}
	})
	h.Frame(view)
	h.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(40, 20)})
	h.Frame(view)
	if !state.Hovered {
		t.Fatal("pointer entering the field did not set hover")
	}
	h.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(90, 50)})
	h.Frame(view)
	if state.Hovered {
		t.Fatal("pointer leaving the field kept hover")
	}
}

func TestFieldInteractionPressFocusPolicy(t *testing.T) {
	for _, test := range []struct {
		name      string
		disabled  bool
		filter    func(pointer.Event) bool
		wantFocus bool
	}{
		{name: "default", wantFocus: true},
		{name: "allowed", filter: func(pointer.Event) bool { return true }, wantFocus: true},
		{name: "action slot", filter: func(pointer.Event) bool { return false }},
		{name: "disabled", disabled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(100, 60), theme.Light())
			var state base.FieldInteraction
			var editor widget.Editor
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if test.disabled {
					gtx = gtx.Disabled()
				}
				(base.TextEditor{SingleLine: true}).Update(ctx, gtx, base.TextEditorRequest{
					FieldKey: "field", Editor: &editor,
				})
				state.UpdateWithFocus(ctx, gtx, test.disabled, &editor, test.filter)
				editorContext := gtx
				editorContext.Constraints = layout.Exact(image.Pt(1, 1))
				testEditor(&editor).Layout(ctx, editorContext)
				state.AddPointer(gtx, image.Pt(80, 40), test.disabled)
				return layout.Dimensions{Size: image.Pt(80, 40)}
			})
			h.Frame(view)
			// Press the shell, outside the editor's own one-pixel hit area.
			h.Click(f32.Pt(40, 20))
			h.Frame(view)
			h.Frame(view)
			if got := h.Router().Source().Focused(&editor); got != test.wantFocus {
				t.Fatalf("editor focus = %v, want %v", got, test.wantFocus)
			}
		})
	}
}

func fieldTestContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}
