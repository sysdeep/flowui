package base

import "gioui.org/layout"

// MenuNavigation shares collection movement, typeahead and activation while
// interpreting horizontal arrows as submenu/root-menu intent, not row movement.
// The zero value is ready to use.
type MenuNavigation struct {
	navigation CollectionNavigation
}

type MenuPolicy struct {
	Nested     bool
	HasSubmenu func(int) bool
}

// MenuActions reports intent without owning popup layout or menu data.
type MenuActions struct {
	Focus        func(string)
	Activate     func(string)
	Open         func(string)
	Close        func()
	NextRoot     func()
	PreviousRoot func()
}

func (s *MenuNavigation) Update(gtx layout.Context, rows Collection, policy MenuPolicy, actions MenuActions) {
	commands := collectionActions{
		focus: actions.Focus, activate: actions.Activate,
		forward: func(index int) {
			enabled := index >= 0 && index < rows.Count && !listDisabled(rows, index)
			if enabled && policy.HasSubmenu != nil && policy.HasSubmenu(index) {
				if actions.Open != nil {
					actions.Open(rows.Key(index))
				}
			} else if actions.NextRoot != nil {
				actions.NextRoot()
			}
		},
	}
	if policy.Nested {
		commands.back = actions.Close
	} else {
		commands.back = actions.PreviousRoot
	}
	s.navigation.update(gtx, rows, -1, CollectionPolicy{Wrap: true, Activate: true, Typeahead: true}, commands)
}
