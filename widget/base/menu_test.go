package base_test

import (
	"image"
	"slices"
	"testing"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestMenuNavigationSubmenuAndRootIntents(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var nav base.MenuNavigation
	var click widget.Clickable
	keys := []string{"leaf", "disabled", "submenu"}
	labels := []string{"Leaf", "Disabled", "Submenu"}
	active, nested, disabled := 0, false, false
	var got []string
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		nav.Update(gtx, base.Collection{
			Count:      len(keys),
			Clickables: func(yield func(int, *widget.Clickable) bool) { yield(active, &click) },
			Key:        func(i int) string { return keys[i] }, Label: func(i int) string { return labels[i] },
			Disabled: func(i int) bool { return i == 1 },
		}, base.MenuPolicy{Nested: nested, HasSubmenu: func(i int) bool { return i == 2 }}, base.MenuActions{
			Focus:        func(next string) { active = slices.Index(keys, next) },
			Activate:     func(next string) { got = append(got, "activate:"+next) },
			Open:         func(next string) { got = append(got, "open:"+next) },
			Close:        func() { got = append(got, "close") },
			NextRoot:     func() { got = append(got, "next-root") },
			PreviousRoot: func() { got = append(got, "previous-root") },
		})
		return click.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(160, 40)} })
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Key(key.NameRightArrow, 0)
	h.Frame(view)
	h.Key(key.NameLeftArrow, 0)
	h.Frame(view)
	h.Key("S", 0)
	h.Frame(view)
	h.Key(key.NameRightArrow, 0)
	h.Frame(view)
	nested = true
	h.Key(key.NameLeftArrow, 0)
	h.Frame(view)
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Press})
	h.Frame(view)
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Release})
	h.Frame(view)
	want := []string{"next-root", "previous-root", "open:submenu", "close", "activate:submenu"}
	if !slices.Equal(got, want) {
		t.Fatalf("menu intents = %v, want %v", got, want)
	}
	disabled = true
	h.Key(key.NameRightArrow, 0)
	h.Frame(view)
	if len(got) != len(want) {
		t.Fatalf("disabled menu emitted intent: %v", got)
	}
}
