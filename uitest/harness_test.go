package uitest_test

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	giowidget "gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/alert"
	"github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/badge"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/collapsible"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/portal"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/window"
)

func anonymousButton() button.Widget { return button.New("Click") }

func TestAnonymousWidgetsKeepPointerStateAcrossFramesAndWindows(t *testing.T) {
	h := uitest.New(image.Pt(240, 120), theme.Light())
	other := uitest.New(image.Pt(240, 120), theme.Light())
	clicks := 0
	fixed := anonymousButton()
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		return flowlayout.Column(fixed, anonymousButton().OnClick(func() { clicks++ })).Layout(ctx, gtx)
	})
	h.Frame(root)
	other.Frame(anonymousButton())
	h.Click(f32.Pt(20, 48))
	h.Frame(root)
	if clicks != 1 {
		t.Fatalf("rebuilt anonymous button clicks=%d, want 1", clicks)
	}
}

func TestAnonymousContainerKeepsChildStateAcrossFrames(t *testing.T) {
	h := uitest.New(image.Pt(320, 160), theme.Light())
	var count int
	child := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		state := ui.UseState[int](ctx, "child")
		*state++
		count = *state
		return layout.Dimensions{Size: image.Pt(80, 20)}
	})
	for want := 1; want <= 3; want++ {
		h.Frame(collapsible.New("Details", child).Expanded(true))
		if count != want {
			t.Fatalf("child state = %d, want %d", count, want)
		}
	}
}

func TestAnonymousRootsKeepChildStateAcrossFrames(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(ui.Widget) ui.Widget
	}{
		{name: "menubar", build: func(child ui.Widget) ui.Widget {
			return menubar.New([]menubar.Item{
				menubar.NewMenu("file", "File", []menu.Item{{Key: "open", Label: "Open"}}).Trigger(child),
			})
		}},
		{name: "toast", build: func(child ui.Widget) ui.Widget {
			return toast.New([]toast.Item{toast.NewItem("saved", "Saved").Timeout(0).Indicator(child)})
		}},
		{name: "titlebar", build: func(child ui.Widget) ui.Widget {
			return titlebar.New("App", nil).Leading(child)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(640, 480), theme.Light())
			var first *int
			layouts := 0
			child := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
				state := ui.UseState[int](ctx, "child")
				if first == nil {
					first = state
				}
				if state != first {
					t.Fatal("rebuilt root replaced its child state")
				}
				*state++
				layouts++
				return layout.Dimensions{Size: image.Pt(24, 16)}
			})
			for frame := range 3 {
				before := layouts
				h.Frame(test.build(child))
				if layouts <= before {
					t.Fatalf("frame %d did not lay out child", frame)
				}
				h.Advance(time.Second)
			}
		})
	}
}

func TestPresentationTransitionsKeepSeparateInstanceState(t *testing.T) {
	declaration := style.New().Opacity(.5).Transition(style.PropOpacity, time.Second)
	for _, test := range []struct {
		name  string
		build func() ui.Widget
	}{
		{name: "card", build: func() ui.Widget { return surface.NewCard(text.New("Card")).Style(declaration) }},
		{name: "surface", build: func() ui.Widget { return surface.New(text.New("Surface")).Style(declaration) }},
		{name: "alert", build: func() ui.Widget { return alert.New("Alert", "").Style(declaration) }},
		{name: "avatar", build: func() ui.Widget { return avatar.New("AB").Style(declaration) }},
		{name: "badge", build: func() ui.Widget { return badge.New(text.New("Anchor"), "2").Style(declaration) }},
		{name: "chip", build: func() ui.Widget { return chip.New("Chip").Style(declaration) }},
		{name: "text", build: func() ui.Widget { return text.New("Text").Style(declaration) }},
		{name: "label", build: func() ui.Widget { return text.NewLabel("Label").Style(declaration) }},
		{name: "description", build: func() ui.Widget { return text.NewDescription("Description").Style(declaration) }},
		{name: "spinner", build: func() ui.Widget { return spinner.New().Style(declaration) }},
		{name: "statusbar", build: func() ui.Widget { return statusbar.New(text.New("Ready"), nil).Style(declaration) }},
		{name: "toolbar", build: func() ui.Widget { return toolbar.New(text.New("Tools")).Style(declaration) }},
		{name: "toolbar separators", build: func() ui.Widget {
			return toolbar.New(toolbar.Separator().Style(declaration))
		}},
		{name: "standalone separator", build: func() ui.Widget { return toolbar.Separator().Style(declaration) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(640, 480), theme.Light())
			root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return flowlayout.Column(test.build(), test.build()).Layout(ctx, gtx)
			})
			h.Frame(root)
			states := core.StateLen(h.Context())
			if states < 2 {
				t.Fatalf("transition states = %d, want at least 2", states)
			}
			h.Advance(100 * time.Millisecond)
			h.Frame(root)
			if got := core.StateLen(h.Context()); got != states {
				t.Fatalf("rebuilt transition states = %d, want %d", got, states)
			}
		})
	}
}

func TestInternalControlKeysDoNotCollideWithApplicationKeys(t *testing.T) {
	h := uitest.New(image.Pt(640, 480), theme.Light())
	clicks := 0
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		return flowlayout.Column(
			alert.New("Notice", "").Key("notice").Closable(true),
			button.New("Close sibling").Key("alert-close-notice"),
			code.New("go", "var x = 1").Key("sample"),
			button.New("Copy sibling").Key("sample-copy").OnClick(func() { clicks++ }),
			combobox.New(nil).Key("search"),
			input.New("").Key("search-input"),
			button.New("Open sibling").Key("search-open"),
		).Layout(ctx, gtx)
	})
	h.Frame(root)
	node, ok := uitest.SemanticByLabel(h.Semantics(), "Copy sibling")
	if !ok {
		t.Fatal("missing sibling button")
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	h.Frame(root)
	if clicks != 1 {
		t.Fatalf("sibling clicks = %d, want 1", clicks)
	}
}

func TestHarnessRoutesPointerAndKeyboardClicks(t *testing.T) {
	harness := uitest.New(image.Pt(240, 120), theme.Light())
	button := new(testButton)

	harness.Frame(button)
	harness.Click(f32.Pt(20, 16))
	harness.Frame(button)
	if button.clicks != 1 {
		t.Fatalf("pointer clicks = %d, want 1", button.clicks)
	}

	harness.Router().Source().Execute(key.FocusCmd{Tag: &button.state})
	if !harness.Router().Source().Focused(&button.state) {
		t.Fatal("button did not gain keyboard focus")
	}
	harness.Key(key.NameReturn, 0)
	harness.Frame(button)
	if button.clicks != 2 {
		t.Fatalf("clicks after keyboard activation = %d, want 2", button.clicks)
	}
	if len(harness.Router().AppendSemantics(nil)) == 0 {
		t.Fatal("completed frame did not expose semantics")
	}
}

func TestHarnessAdvancesTweenAndLaysOutPortal(t *testing.T) {
	harness := uitest.New(image.Pt(240, 120), theme.Light())
	var progress float32
	portalLayouts := 0
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		progress = animation.Tween("progress", 1).
			Initial(0).
			Duration(time.Second).
			Easing(animation.EaseLinear).
			Value(ctx, gtx)
		return portal.New(true, flowlayout.Spacer(20, 10), func(image.Rectangle, bool) ui.Widget {
			return ui.WidgetFunc(func(*ui.Context, layout.Context) layout.Dimensions {
				portalLayouts++
				return layout.Dimensions{}
			})
		}).Layout(ctx, gtx)
	})

	harness.Frame(root)
	// First open: paint pass + policy pass on the visual top.
	if progress != 0 || portalLayouts != 2 {
		t.Fatalf("first frame = progress %v, portal layouts %d; want 0, 2", progress, portalLayouts)
	}
	harness.Advance(500 * time.Millisecond)
	harness.Frame(root)
	// Following frame freezes the previous owner; only one paint layout.
	if math.Abs(float64(progress-.5)) > .001 || portalLayouts != 3 {
		t.Fatalf("second frame = progress %v, portal layouts %d; want .5, 3", progress, portalLayouts)
	}
}

func TestHarnessConfigurationAndResize(t *testing.T) {
	light := theme.Light()
	if language := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 100), Theme: &light}).Context().Language(); language != window.LanguageEnglish {
		t.Fatalf("default language = %q, want English", language)
	}

	activeTheme := theme.Dark()
	activeTheme.Colors.Tones[theme.TonePrimary].Fill = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	harness := uitest.NewWithConfig(uitest.Config{
		Size:     image.Pt(200, 100),
		Theme:    &activeTheme,
		Language: window.LanguageChinese,
		Now:      time.Unix(10, 0),
	})
	if harness.Context().Language() != window.LanguageChinese {
		t.Fatalf("language = %q, want Chinese", harness.Context().Language())
	}
	if harness.Context().Theme().Colors.Background != activeTheme.Colors.Background {
		t.Fatal("harness did not use the configured theme")
	}
	var now time.Time
	probe := ui.WidgetFunc(func(_ *ui.Context, gtx layout.Context) layout.Dimensions {
		now = gtx.Now
		return layout.Dimensions{Size: gtx.Constraints.Min}
	})
	harness.Resize(image.Pt(320, 180))
	dimensions := harness.Frame(probe)
	if dimensions.Size != image.Pt(320, 180) || now != time.Unix(10, 0) {
		t.Fatalf("resized frame = dimensions %v, time %v", dimensions.Size, now)
	}
	harness.Advance(time.Second)
	harness.Frame(probe)
	if now != time.Unix(11, 0) {
		t.Fatalf("time = %v, want 11 seconds", now)
	}
}

func TestHarnessRejectsInvalidTimeAndSize(t *testing.T) {
	activeTheme := theme.Light()
	mustPanic(t, func() { uitest.New(image.Point{}, activeTheme) })
	mustPanic(t, func() { uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 100)}) })
	_ = uitest.New(image.Pt(100, 100), theme.Theme{})
	harness := uitest.New(image.Pt(100, 100), activeTheme)
	mustPanic(t, func() { harness.Resize(image.Pt(-1, 100)) })
	mustPanic(t, func() { harness.Advance(-time.Second) })
}

type testButton struct {
	state  giowidget.Clickable
	clicks int
}

func (button *testButton) Layout(_ *ui.Context, gtx layout.Context) layout.Dimensions {
	for button.state.Clicked(gtx) {
		button.clicks++
	}
	return button.state.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp("Test button").Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(80, 32)}
	})
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
