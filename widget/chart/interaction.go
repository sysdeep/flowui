package chart

import (
	"image"
	"image/color"
	"sort"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// Datum describes one visible value in a chart selection.
type Datum struct {
	SeriesKey   string
	SeriesLabel string
	X           float64
	Y           float64
	// Percent is populated by proportional charts such as PieChart.
	Percent float64
	// Open, Close, Low, and High are populated by CandlestickChart.
	Open  float64
	Close float64
	Low   float64
	High  float64
	Color color.NRGBA
}

// Selection describes the current chart selection.
type Selection struct {
	Label string
	// Index is the source data or category index, or -1 when not applicable.
	Index int
	X     float64
	Items []Datum
}

func UpdatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, window DataWindow, verticalWindow bool, onWindowChange func(DataWindow), tag event.Tag, hovered *bool, position *f32.Point, windowGesture *DataWindowGesture) {
	var onPointer func(pointer.Event)
	if onWindowChange != nil {
		onPointer = func(e pointer.Event) {
			if next, changed := windowGesture.Update(e, plot, window, verticalWindow); changed {
				window = next
				onWindowChange(next)
			}
		}
	}
	updatePointer(gtx, enabled, tag, hovered, position, onPointer)
	if onWindowChange == nil || !enabled || !gtx.Enabled() {
		windowGesture.Cancel()
	}
}

// UpdateHoverPointer observes chart hover and pointer position without a data
// window. Shape and plot-boundary hit testing remain with the chart component.
func UpdateHoverPointer(gtx layout.Context, enabled bool, tag event.Tag, hovered *bool, position *f32.Point) {
	updatePointer(gtx, enabled, tag, hovered, position, nil)
}

func updatePointer(gtx layout.Context, enabled bool, tag event.Tag, hovered *bool, position *f32.Point, onPointer func(pointer.Event)) {
	enabled = enabled && gtx.Enabled()
	kinds := pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Release | pointer.Cancel
	scrollX, scrollY := pointer.ScrollRange{}, pointer.ScrollRange{}
	if onPointer != nil {
		kinds |= pointer.Scroll
		scrollX = pointer.ScrollRange{Min: -100000, Max: 100000}
		scrollY = pointer.ScrollRange{Min: -100000, Max: 100000}
	}
	if !enabled {
		*hovered = false
	}
	for {
		value, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: kinds, ScrollX: scrollX, ScrollY: scrollY})
		if !ok {
			return
		}
		e, ok := value.(pointer.Event)
		if !ok || !enabled {
			continue
		}
		if onPointer != nil {
			onPointer(e)
		}
		switch e.Kind {
		case pointer.Enter, pointer.Move, pointer.Drag, pointer.Press, pointer.Scroll:
			*hovered = true
			*position = e.Position
		case pointer.Leave, pointer.Cancel:
			*hovered = false
		}
	}
}

func AddPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool, tag event.Tag) {
	AddPointerInputWithCursor(gtx, plot, enabled, pointer.CursorCrosshair, tag)
}

// AddPointerInputWithCursor registers chart pointer input with an explicit
// cursor for visualizations whose interaction is not crosshair-based.
func AddPointerInputWithCursor(gtx layout.Context, plot image.Rectangle, enabled bool, cursor pointer.Cursor, tag event.Tag) {
	if !enabled || !gtx.Enabled() || plot.Empty() {
		return
	}
	area := clip.Rect(plot).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	cursor.Add(gtx.Ops)
	event.Op(gtx.Ops, tag)
	pass.Pop()
	area.Pop()
}

func UpdateClicks(gtx layout.Context, enabled bool, click *gesture.Click) (activated, reset bool) {
	for {
		value, ok := click.Update(gtx.Source)
		if !ok {
			break
		}
		if value.Kind == gesture.KindClick {
			activated = true
			reset = reset || value.NumClicks >= 2
		}
	}
	return activated && enabled && gtx.Enabled(), reset && enabled && gtx.Enabled()
}

func AddClickInput(gtx layout.Context, size image.Point, enabled bool, click *gesture.Click) {
	AddClickInputRect(gtx, image.Rectangle{Max: size}, enabled, click)
}

// AddClickInputRect registers pointer-only clicks in caller-supplied bounds,
// including chart plots whose origin is not the widget origin.
func AddClickInputRect(gtx layout.Context, bounds image.Rectangle, enabled bool, click *gesture.Click) {
	if !enabled || !gtx.Enabled() || bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	click.Add(gtx.Ops)
	pass.Pop()
	area.Pop()
}

func SelectedIndex(position f32.Point, hovered bool, start, end int, plot image.Rectangle, horizontal bool) (int, bool) {
	if !hovered || end <= start || plot.Empty() {
		return 0, false
	}
	ratio := (float64(position.X) - float64(plot.Min.X)) / float64(plot.Dx())
	if horizontal {
		ratio = (float64(position.Y) - float64(plot.Min.Y)) / float64(plot.Dy())
	}
	index := start + int(ratio*float64(end-start))
	return min(max(index, start), end-1), true
}

func NearestX(position f32.Point, hovered bool, values []float64, minimum, maximum float64, plot image.Rectangle) (float64, bool) {
	if !hovered || len(values) == 0 || plot.Empty() {
		return 0, false
	}
	ratio := (float64(position.X) - float64(plot.Min.X)) / float64(plot.Dx())
	ratio = min(max(ratio, 0), 1)
	target := minimum + ratio*(maximum-minimum)
	index := sort.SearchFloat64s(values, target)
	if index <= 0 {
		return values[0], true
	}
	if index >= len(values) {
		return values[len(values)-1], true
	}
	if target-values[index-1] <= values[index]-target {
		return values[index-1], true
	}
	return values[index], true
}

func TooltipAnchor(position f32.Point) image.Rectangle {
	point := position.Round()
	return image.Rectangle{Min: point, Max: point.Add(image.Pt(1, 1))}
}
