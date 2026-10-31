package layout

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

const (
	stateSlotList   = "list"
	stateSlotScroll = "scroll"
)

func newContext(_ any) *core.Context {
	active := theme.Light()
	return core.New(nil, &active, locale.LanguageAuto)
}

type enabledProbeWidget struct {
	enabled bool
}

type styledBoxProbe struct {
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

func (w *enabledProbeWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func (w *styledBoxProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

type constraintWidget struct {
	constraints layout.Constraints
}

func (w *constraintWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

type overflowWidget struct{}

func (overflowWidget) Layout(_ *core.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(500, 500)}
}

type overlayProbeWidget struct {
	key      string
	size     image.Point
	baseline int
	anchor   image.Rectangle
	got      *image.Rectangle
	capture  func(image.Rectangle)
}

func (w *overlayProbeWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       w.key,
		Anchor:    w.anchor,
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			if w.got != nil {
				*w.got = anchor
			}
			if w.capture != nil {
				w.capture(anchor)
			}
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: w.size, Baseline: w.baseline}
}

func resolveOverlayAnchor(t *testing.T, constraints layout.Constraints, widget core.Widget) image.Rectangle {
	t.Helper()
	ctx := newContext(nil)
	core.BeginFrameWithViewport(ctx, constraints.Max)
	gtx := layout.Context{Constraints: constraints, Ops: new(op.Ops)}
	var got image.Rectangle
	probe, ok := findOverlayProbe(widget)
	if !ok {
		t.Fatal("test widget does not contain an overlay probe")
	}
	probe.got = &got
	widget.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
	return got
}

func findOverlayProbe(widget core.Widget) (*overlayProbeWidget, bool) {
	switch widget := widget.(type) {
	case *overlayProbeWidget:
		return widget, true
	case CenterWidget:
		return findOverlayProbe(widget.child)
	case BoxWidget:
		return findOverlayProbe(widget.Child())
	case RowWidget:
		return findOverlayProbeIn(widget.children)
	case ColumnWidget:
		return findOverlayProbeIn(widget.children)
	case WrapWidget:
		return findOverlayProbeIn(widget.children)
	case FlexWidget:
		return findOverlayProbe(widget.child)
	default:
		return nil, false
	}
}

func findOverlayProbeIn(widgets []core.Widget) (*overlayProbeWidget, bool) {
	for _, widget := range widgets {
		if probe, ok := findOverlayProbe(widget); ok {
			return probe, true
		}
	}
	return nil, false
}
