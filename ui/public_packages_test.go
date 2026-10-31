package ui_test

import (
	"fmt"
	"image/color"
	"testing"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

func TestPublicPackageComposition(t *testing.T) {
	custom := ui.WidgetFunc(func(_ *ui.Context, _ giolayout.Context) giolayout.Dimensions {
		return giolayout.Dimensions{}
	})

	root := flowlayout.Column(
		(button.New("Save")).Style((style.Style{}).Radius(unit.Dp(8))),
		flowlayout.Row(
			flowlayout.Expanded(custom),
			text.New("Ready"),
		),
	)
	var _ ui.Widget = root

	var _ window.Option = window.CenterOnStart()
	var _ window.Option = window.Title("FlowUI")

	brush := draw.LinearGradient(
		draw.GradientStop{Offset: 0, Color: color.NRGBA{R: 0xff, A: 0xff}},
		draw.GradientStop{Offset: 1, Color: color.NRGBA{B: 0xff, A: 0xff}},
	)
	if got := brush.ColorAt(0.5); got.A != 0xff {
		t.Fatalf("draw brush alpha = %d, want 255", got.A)
	}

	_ = (style.Style{}).Padding(unit.Dp(8))
}

func TestMVUProgramUsesPublicWidgetContract(t *testing.T) {
	type model struct{ count int }
	type message int

	program := ui.NewProgram(model{}, func(value *model, msg message) ui.Cmd[message] {
		value.count += int(msg)
		return nil
	}, func(_ *ui.Context, value model, _ ui.Send[message]) ui.Widget {
		return text.New(fmt.Sprintf("count=%d", value.count))
	})

	initial, cmd := program.Init()
	if cmd != nil || initial.count != 0 {
		t.Fatalf("program init = %#v, %v", initial, cmd)
	}
	if update := program.Update(&initial, 2); update != nil || initial.count != 2 {
		t.Fatalf("program update = count %d, cmd=%v", initial.count, update)
	}
	if program.View(nil, initial, func(message) {}) == nil {
		t.Fatal("program view returned a nil widget")
	}
}
