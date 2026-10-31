package base

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
)

// DisclosureState is the interaction state a Disclosure reports. It changes
// every frame and must not be retained.
type DisclosureState struct {
	Interaction
	Key       string
	Expanded  bool
	clickable *widget.Clickable
	label     string
}

// DisclosureWidget is a headless expand/collapse control: a trigger that
// reports its expanded state, plus an optional content slot that is laid out
// only while expanded. It paints nothing.
type DisclosureWidget struct {
	key       string
	label     string
	expanded  bool
	disabled  bool
	onToggle  func(bool)
	trigger   func(DisclosureState) core.Widget
	content   func() core.Widget
	clickable *widget.Clickable
}

// Disclosure creates a headless disclosure with the given controlled state.
func Disclosure(key string, expanded bool) DisclosureWidget {
	return DisclosureWidget{key: key, expanded: expanded}
}

// Label sets the accessible name of the trigger.
func (d DisclosureWidget) Label(label string) DisclosureWidget {
	d.label = label
	return d
}

// Disabled stops activation but keeps the current expanded state visible.
func (d DisclosureWidget) Disabled(disabled bool) DisclosureWidget {
	d.disabled = disabled
	return d
}

// OnToggle reports the requested next expanded state.
func (d DisclosureWidget) OnToggle(fn func(bool)) DisclosureWidget {
	d.onToggle = fn
	return d
}

// Clickable supplies retained interaction state for a compound control.
func (d DisclosureWidget) Clickable(value *widget.Clickable) DisclosureWidget {
	d.clickable = value
	return d
}

// Update processes trigger interaction independently of content layout. This
// lets animated disclosures keep their closing region mounted in the visual.
func (d DisclosureWidget) Update(ctx *core.Context, gtx layout.Context) DisclosureState {
	click := core.Begin(ctx, gtx, d.key, d.clickable, !d.disabled, true, func() {
		if d.onToggle != nil {
			d.onToggle(!d.expanded)
		}
	})
	return DisclosureState{
		Interaction: Interaction{Hovered: click.State.Hovered, Pressed: click.State.Pressed, Focused: click.State.Focused, FocusVisible: click.State.FocusVisible, Disabled: !click.Enabled},
		Key:         click.Key, Expanded: d.expanded,
		clickable: click.Clickable, label: d.label,
	}
}

// Trigger builds the visual for the disclosure trigger. It is required.
func (d DisclosureWidget) Trigger(fn func(DisclosureState) core.Widget) DisclosureWidget {
	d.trigger = fn
	return d
}

// Content builds the expandable region. It is laid out only while expanded.
func (d DisclosureWidget) Content(fn func() core.Widget) DisclosureWidget {
	d.content = fn
	return d
}

func (d DisclosureWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if d.trigger == nil {
		panic("flowui: base.Disclosure requires Trigger")
	}
	state := d.Update(ctx, gtx)

	trigger := d.trigger(state)
	if trigger == nil {
		panic("flowui: base.Disclosure Trigger returned nil")
	}

	rows := []layoutui.Widget{
		core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return state.LayoutTrigger(ctx, gtx, trigger)
		}),
	}
	if d.expanded && d.content != nil {
		content := d.content()
		if content != nil {
			rows = append(rows, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return content.Layout(ctx, gtx)
			}))
		}
	}
	return layoutui.Column(rows...).Layout(ctx, gtx)
}

// LayoutTrigger mounts the hit target and semantics for this Update result,
// without processing events again. The caller owns the expandable region and
// may keep it mounted for exit animations. Use only within the current frame.
func (s DisclosureState) LayoutTrigger(ctx *core.Context, gtx layout.Context, visual core.Widget) layout.Dimensions {
	if s.clickable == nil || visual == nil {
		panic("flowui: disclosure trigger requires an Update result and visual")
	}
	draw := func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := visual.Layout(ctx, gtx)
		call := macro.Stop()
		area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(!s.Disabled).Add(gtx.Ops)
		semantic.SelectedOp(s.Expanded).Add(gtx.Ops)
		if s.label != "" {
			semantic.LabelOp(s.label).Add(gtx.Ops)
		}
		call.Add(gtx.Ops)
		area.Pop()
		return dims
	}
	if s.Disabled {
		return draw(gtx.Disabled())
	}
	return s.clickable.Layout(gtx, draw)
}
