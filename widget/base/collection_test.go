package base_test

import (
	"image"
	"slices"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/base"
)

func TestUpdateRowPreservesAuxiliaryActionFocus(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var owner, action widget.Clickable
	activated, disabled := 0, false
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		base.UpdateRow(ctx, gtx, base.RowInput{
			Clickable: &action, PreserveFocus: true, Disabled: disabled,
			OnClick: func(widget.Click) { activated++ },
		})
		owner.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{} })
		return action.Layout(gtx, func(layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(160, 40)}
		})
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &owner})
	h.Frame(view)
	h.Click(f32.Pt(40, 20))
	h.Frame(view)
	if activated != 1 || !h.Router().Source().Focused(&owner) {
		t.Fatalf("activated=%d owner focused=%v", activated, h.Router().Source().Focused(&owner))
	}
	disabled = true
	action.Click()
	h.Frame(view)
	if activated != 1 {
		t.Fatal("disabled auxiliary action activated")
	}
}

func TestCollectionNavigation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy base.CollectionPolicy
		keys   []key.Name
		want   string
	}{
		{name: "skip disabled", keys: []key.Name{key.NameDownArrow}, want: "c"},
		{name: "stop at edge", keys: []key.Name{key.NameUpArrow}, want: "a"},
		{name: "wrap at edge", policy: base.CollectionPolicy{Wrap: true}, keys: []key.Name{key.NameUpArrow}, want: "d"},
		{name: "home after end", keys: []key.Name{key.NameEnd, key.NameHome}, want: "a"},
		{name: "typeahead fallback", policy: base.CollectionPolicy{Typeahead: true}, keys: []key.Name{"C", "D"}, want: "d"},
		{name: "horizontal ignores down", policy: base.CollectionPolicy{Axis: base.CollectionHorizontal}, keys: []key.Name{key.NameDownArrow}, want: "a"},
		{name: "horizontal skips disabled", policy: base.CollectionPolicy{Axis: base.CollectionHorizontal}, keys: []key.Name{key.NameRightArrow}, want: "c"},
		{name: "vertical ignores right", keys: []key.Name{key.NameRightArrow}, want: "a"},
		{name: "both axes", policy: base.CollectionPolicy{Axis: base.CollectionBothAxes}, keys: []key.Name{key.NameRightArrow, key.NameUpArrow}, want: "a"},
	} {
		for _, groupFocus := range []bool{false, true} {
			name := "rows/" + tc.name
			if groupFocus {
				name = "group/" + tc.name
			}
			t.Run(name, func(t *testing.T) {
				h := uitest.New(image.Pt(160, 40), theme.Light())
				var nav base.CollectionNavigation
				var click widget.Clickable
				active := "a"
				keys := []string{"a", "b", "c", "d"}
				labels := []string{"Alpha", "Beta", "Charlie", "Delta"}
				view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
					rows := base.Collection{
						Count: len(keys), Key: func(i int) string { return keys[i] },
						Label: func(i int) string { return labels[i] }, Disabled: func(i int) bool { return i == 1 },
						Clickables: func(yield func(int, *widget.Clickable) bool) {
							// Only the current row is mounted; destinations can be off-screen.
							yield(slices.Index(keys, active), &click)
						},
					}
					if groupFocus {
						rows.Focus, rows.Clickables = &click, nil
					}
					nav.Update(gtx, rows, slices.Index(keys, active), tc.policy, func(next string) { active = next }, nil)
					return click.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(160, 40)} })
				})
				h.Frame(view)
				h.Router().Source().Execute(key.FocusCmd{Tag: &click})
				h.Frame(view)
				for _, name := range tc.keys {
					h.Key(name, 0)
					h.Frame(view)
				}
				if active != tc.want {
					t.Fatalf("active = %q, want %q", active, tc.want)
				}
			})
		}
	}
}

func TestCollectionShortcutsAndPressActivation(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var nav base.CollectionNavigation
	var click widget.Clickable
	keys := []string{"a", "b", "c"}
	active, selected, actions := 0, "a", 0
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		nav.Update(gtx, base.Collection{
			Count: len(keys), Focus: &click, Key: func(i int) string { return keys[i] },
		}, active, base.CollectionPolicy{
			Axis: base.CollectionHorizontal, Wrap: true, Shortcuts: true,
			Activate: true, ActivateOnPress: true,
		}, func(next string) { active = slices.Index(keys, next) }, func(next string) { selected = next; actions++ })
		return click.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(160, 40)} })
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Key(key.NameTab, key.ModShortcut|key.ModShift)
	h.Frame(view)
	if active != 2 || selected != "a" {
		t.Fatalf("reverse shortcut = active %d selected %q", active, selected)
	}
	h.Key(key.NamePageUp, key.ModShortcut)
	h.Frame(view)
	if active != 1 {
		t.Fatalf("PageUp active = %d, want 1", active)
	}
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Press})
	h.Frame(view)
	if selected != "b" || actions != 1 {
		t.Fatalf("press activation = %q/%d, want b/1", selected, actions)
	}
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Release})
	h.Frame(view)
	if actions != 1 {
		t.Fatalf("release activated again: %d", actions)
	}
}

func TestCollectionGroupWithNoEnabledRows(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var nav base.CollectionNavigation
	var click widget.Clickable
	actions := 0
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		nav.Update(gtx, base.Collection{
			Count: 1, Focus: &click,
			Key: func(int) string { return "a" }, Disabled: func(int) bool { return true },
		}, -1, base.CollectionPolicy{Wrap: true, Activate: true, Typeahead: true}, nil, func(string) { actions++ })
		return click.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(160, 40)} })
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	for _, name := range []key.Name{key.NameDownArrow, key.NameHome, key.NameReturn, "A"} {
		h.Key(name, 0)
		h.Frame(view)
	}
	if actions != 0 {
		t.Fatalf("disabled rows activated %d times", actions)
	}
}

func TestCollectionActivationCancelledByDisable(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var nav base.CollectionNavigation
	var click widget.Clickable
	disabled, actions := false, 0
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		nav.Update(gtx, base.Collection{
			Count: 1, Key: func(int) string { return "a" }, Clickables: func(yield func(int, *widget.Clickable) bool) { yield(0, &click) },
		}, 0, base.CollectionPolicy{Activate: true}, nil, func(string) { actions++ })
		return click.Layout(gtx, func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(160, 40)} })
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Press})
	h.Frame(view)
	if actions != 0 {
		t.Fatal("activated before release")
	}
	disabled = true
	h.Frame(view)
	disabled = false
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Router().Queue(key.Event{Name: key.NameReturn, State: key.Release})
	h.Frame(view)
	if actions != 0 {
		t.Fatal("released key activated after disabling")
	}
	h.Key(key.NameReturn, 0)
	h.Frame(view)
	if actions != 1 {
		t.Fatalf("fresh activation count = %d, want 1", actions)
	}
}

func TestCollectionEventsPreservePressModifiersAndRowIdentity(t *testing.T) {
	for _, change := range []string{"none", "disable", "remove", "reorder"} {
		t.Run(change, func(t *testing.T) {
			h := uitest.New(image.Pt(160, 40), theme.Light())
			var nav base.CollectionNavigation
			var click widget.Clickable
			keys := []string{"a", "b"}
			disabled := false
			var events []base.CollectionEvent
			view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				nav.UpdateWithEvents(gtx, base.Collection{
					Count: len(keys), Focus: &click,
					Key:      func(i int) string { return keys[i] },
					Disabled: func(i int) bool { return disabled && keys[i] == "a" },
				}, 0, base.CollectionPolicy{Activate: true, Modifiers: key.ModShift}, func(e base.CollectionEvent) int {
					events = append(events, e)
					return e.Index
				})
				return click.Layout(gtx, func(layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: image.Pt(160, 40)}
				})
			})
			h.Frame(view)
			h.Router().Source().Execute(key.FocusCmd{Tag: &click})
			h.Frame(view)
			h.Router().Queue(key.Event{Name: key.NameSpace, State: key.Press, Modifiers: key.ModShift})
			h.Frame(view)
			switch change {
			case "disable":
				disabled = true
			case "remove":
				keys = []string{"b"}
			case "reorder":
				keys = []string{"b", "a"}
			}
			h.Router().Queue(key.Event{Name: key.NameSpace, State: key.Release})
			h.Frame(view)
			if change == "disable" || change == "remove" {
				if len(events) != 0 {
					t.Fatalf("invalidated row activated: %+v", events)
				}
				return
			}
			if len(events) != 1 || events[0].Kind != base.CollectionActivate || events[0].Modifiers != key.ModShift || keys[events[0].Index] != "a" {
				t.Fatalf("activation = %+v, want original row and press modifiers", events)
			}
		})
	}
}

func TestCollectionCommandUpdatesSameFrameNavigation(t *testing.T) {
	h := uitest.New(image.Pt(160, 40), theme.Light())
	var nav base.CollectionNavigation
	var click widget.Clickable
	var events []base.CollectionEvent
	view := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		nav.UpdateWithEvents(gtx, base.Collection{Count: 4, Focus: &click}, 0,
			base.CollectionPolicy{Commands: []key.Name{key.NameRightArrow}}, func(e base.CollectionEvent) int {
				events = append(events, e)
				if e.Kind == base.CollectionCommand {
					return 2
				}
				return e.Index
			})
		return click.Layout(gtx, func(layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(160, 40)}
		})
	})
	h.Frame(view)
	h.Router().Source().Execute(key.FocusCmd{Tag: &click})
	h.Frame(view)
	h.Router().Queue(
		key.Event{Name: key.NameRightArrow, State: key.Press},
		key.Event{Name: key.NameDownArrow, State: key.Press},
	)
	h.Frame(view)
	if len(events) != 2 || events[0].Kind != base.CollectionCommand || events[1].Kind != base.CollectionFocus || events[1].Index != 3 {
		t.Fatalf("ordered events = %+v, want command then focus row 3", events)
	}
}
