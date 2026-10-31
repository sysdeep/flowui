package base_test

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestTextSelectionCopiesOnlyFocusedRangeAndHandlesSourceChange(t *testing.T) {
	h := uitest.New(image.Pt(100, 80), theme.Light())
	var selections [2]base.TextSelection
	sources := [2][]rune{[]rune("abc"), []rune("世界好")}
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		for i := range selections {
			offset := op.Offset(image.Pt(0, i*40)).Push(gtx.Ops)
			selection := &selections[i]
			core.AddPointerArea(gtx, selection, image.Rect(0, 0, 100, 40), pointer.CursorText)
			selection.Update(ctx, gtx, selection, sources[i], func(position f32.Point) int {
				return int(position.X / 10)
			})
			offset.Pop()
		}
		return layout.Dimensions{Size: image.Pt(100, 80)}
	})
	h.Frame(view)
	for i := range selections {
		y := float32(i*40 + 10)
		h.TouchDown(f32.Pt(1, y))
		h.Frame(view)
		h.TouchMove(f32.Pt(30, y))
		h.Frame(view)
		h.TouchUp(f32.Pt(30, y))
		h.Frame(view)
	}
	h.Key("C", key.ModShortcut)
	h.Frame(view)
	_, content, ok := h.Router().WriteClipboard()
	if !ok || string(content) != "世界好" {
		t.Fatalf("copy = %q, ok=%v; want focused selection", content, ok)
	}
	sources[1] = []rune("新")
	h.Frame(view)
	h.Key("C", key.ModShortcut)
	h.Frame(view)
	_, content, ok = h.Router().WriteClipboard()
	if !ok || string(content) != "新" {
		t.Fatalf("copy after source change = %q, ok=%v", content, ok)
	}
}
