package base

import "gioui.org/layout"

// MenubarNavigation shares collection movement and typeahead. Perpendicular
// arrows report panel-opening intent; global Alt/F10 accelerators stay with
// the menu bar, which also owns switching an already open panel.
type MenubarNavigation struct {
	navigation CollectionNavigation
}

type MenubarActions struct {
	Focus func(string)
	Open  func(string, bool)
	Close func()
}

func (s *MenubarNavigation) Update(gtx layout.Context, rows Collection, axis CollectionAxis, wrap bool, actions MenubarActions) {
	open := func(index int, last bool) {
		if index >= 0 && index < rows.Count && !listDisabled(rows, index) && actions.Open != nil {
			actions.Open(rows.Key(index), last)
		}
	}
	commands := collectionActions{focus: actions.Focus, escape: actions.Close}
	if axis == CollectionVertical {
		commands.forward = func(index int) { open(index, false) }
	} else {
		commands.down = func(index int) { open(index, false) }
		commands.up = func(index int) { open(index, true) }
	}
	s.navigation.update(gtx, rows, -1, CollectionPolicy{Axis: axis, Wrap: wrap, Typeahead: true}, commands)
}
