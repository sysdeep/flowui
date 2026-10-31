package layout

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestScrollbarResolvesTrackAndThumbParts(t *testing.T) {
	track := color.NRGBA{R: 1, A: 0xff}
	thumb := color.NRGBA{G: 2, A: 0xff}
	custom := style.Style{}.
		Part(style.PartTrack, style.Style{}.Width(14).Background(style.SolidColor{Color: track})).
		Part(style.PartThumb, style.Style{}.Width(4).Background(style.SolidColor{Color: thumb}).Radius(2))

	style := resolvedScrollbarStyle(newContext(nil), testLayoutContext(), "scrollbar", new(widget.Scrollbar), false, custom)
	if style.Width() != 14 || style.Track.Color != track || style.Indicator.Color != thumb || style.Indicator.MinorWidth != 4 || style.Indicator.CornerRadius != 2 {
		t.Fatalf("parts = width %v track %#v thumb %#v/%v radius %v", style.Width(), style.Track.Color, style.Indicator.Color, style.Indicator.MinorWidth, style.Indicator.CornerRadius)
	}
}

func TestScrollbarDefaultsCacheMatchesUncachedStyle(t *testing.T) {
	cache := new(computed.VisualCache[scrollbarDeclarationKey, style.Style])
	bar := new(widget.Scrollbar)
	for _, test := range []struct {
		name     string
		active   theme.Theme
		disabled bool
	}{
		{name: "light", active: theme.Light()},
		{name: "dark", active: theme.Dark()},
		{name: "disabled", active: theme.Dark(), disabled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := core.New(nil, &test.active, "")
			gtx := testLayoutContext()
			want := resolvedScrollbarStyle(ctx, gtx, "", bar, test.disabled, style.Style{})
			for range 2 {
				got := resolvedScrollbarStyle(ctx, gtx, "", bar, test.disabled, style.Style{}, cache)
				if got.Indicator.Color != want.Indicator.Color || got.Indicator.HoverColor != want.Indicator.HoverColor || got.Width() != want.Width() {
					t.Fatalf("cached scrollbar differs from uncached: got %#v, want %#v", got.Indicator, want.Indicator)
				}
			}
		})
	}
}

func TestScrollbarOnlyReservesSpaceWhenScrollable(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := &scrollbarConstraintWidget{height: 20}
	value := Scroll(probe)
	start := time.Unix(1, 0)

	layoutScrollFrame(ctx, router, value, start)
	if probe.constraints.Max.X != 100 {
		t.Fatalf("non-scrollable content max width = %d, want 100", probe.constraints.Max.X)
	}

	probe.height = 300
	layoutScrollFrame(ctx, router, value, start.Add(time.Millisecond))
	layoutScrollFrame(ctx, router, value, start.Add(2*time.Millisecond))
	if probe.constraints.Max.X != 86 {
		t.Fatalf("scrollable content max width = %d, want 86", probe.constraints.Max.X)
	}

	probe.height = 20
	layoutScrollFrame(ctx, router, value, start.Add(3*time.Millisecond))
	layoutScrollFrame(ctx, router, value, start.Add(4*time.Millisecond))
	if probe.constraints.Max.X != 100 {
		t.Fatalf("content max width after overflow = %d, want 100", probe.constraints.Max.X)
	}
}

func TestScrollbarStyleChangesReservedWidth(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := &scrollbarConstraintWidget{height: 300}
	value := Scroll(probe).Style(style.Style{}.Part(style.PartTrack, style.Style{}.Width(12)))
	start := time.Unix(1, 0)
	layoutScrollFrame(ctx, router, value, start)
	layoutScrollFrame(ctx, router, value, start.Add(time.Millisecond))
	if probe.constraints.Max.X != 84 {
		t.Fatalf("styled scrollbar content width = %d, want 84", probe.constraints.Max.X)
	}
}

func TestOverlayScrollbarNeverReservesSpace(t *testing.T) {
	for _, test := range []struct {
		name string
		bar  ScrollWidget
		want int
	}{
		{name: "short", bar: Scroll(&scrollbarConstraintWidget{height: 20}).Overlay(), want: 100},
		{name: "long", bar: Scroll(&scrollbarConstraintWidget{height: 300}).Overlay(), want: 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe := test.bar.child.(*scrollbarConstraintWidget)
			ctx := newContext(nil)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutScrollFrame(ctx, router, test.bar, start)
			layoutScrollFrame(ctx, router, test.bar, start.Add(time.Millisecond))
			if probe.constraints.Max.X != test.want {
				t.Fatalf("content max width = %d, want %d", probe.constraints.Max.X, test.want)
			}
		})
	}
}

type scrollbarConstraintWidget struct {
	height      int
	constraints layout.Constraints
}

func (w *scrollbarConstraintWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, w.height))}
}

func TestScrollbarViewportUsesListPosition(t *testing.T) {
	start, end := scrollbarViewport(layout.Position{
		First:      0,
		Offset:     150,
		OffsetLast: -50,
		Count:      1,
		Length:     300,
	}, 1, 100)
	if math.Abs(float64(start-.5)) > 1e-6 || math.Abs(float64(end-5.0/6.0)) > 1e-6 {
		t.Fatalf("viewport = %v..%v, want .5..%v", start, end, 5.0/6.0)
	}
	if start, end := scrollbarViewport(layout.Position{Length: 80}, 1, 100); start != 0 || end != 1 {
		t.Fatalf("non-scrollable viewport = %v..%v", start, end)
	}
}

func TestScrollbarDoesNotReuseDistanceAfterContentShrinks(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := &scrollbarConstraintWidget{height: 300}
	value := Scroll(probe).Key("body")
	start := time.Unix(1, 0)
	layoutScrollFrame(ctx, router, value, start)
	layoutScrollFrame(ctx, router, value, start.Add(time.Millisecond))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(95, 80)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(95, 80)},
	)
	layoutScrollFrame(ctx, router, value, start.Add(2*time.Millisecond))

	probe.height = 20
	layoutScrollFrame(ctx, router, value, start.Add(3*time.Millisecond))
	state := testComponentState[layout.List](ctx, "body", stateSlotScroll)
	if state.Position.First != 0 || state.Position.Offset != 0 {
		t.Fatalf("short content position = %#v, want start", state.Position)
	}
}
