// Package navigation provides collection navigation helpers and navigation widgets.
package navigation

import (
	"gioui.org/io/key"
	"github.com/qianniancn/flowui/widget/internal/navigation"
)

// List describes an indexed collection.
type List = navigation.List

// Typeahead accumulates incremental type-to-search input.
type Typeahead = navigation.Typeahead

const DefaultTypeaheadTimeout = navigation.DefaultTypeaheadTimeout

// First returns the first enabled index, or (-1, false) if none is enabled.
func First(l List) (int, bool) { return navigation.First(l) }

// Last returns the last enabled index, or (-1, false) if none is enabled.
func Last(l List) (int, bool) { return navigation.Last(l) }

// NearestEnabled searches in delta's direction, then in the opposite direction.
func NearestEnabled(l List, current, delta int) (int, bool) {
	return navigation.NearestEnabled(l, current, delta)
}

// Move advances by delta, skipping disabled rows and optionally wrapping.
// An out-of-range current starts at the first or last enabled row.
func Move(l List, current, delta int, wrap bool) (int, bool) {
	return navigation.Move(l, current, delta, wrap)
}

// Match finds the next enabled label with a case-insensitive query prefix.
func Match(l List, current int, query string) (int, bool) {
	return navigation.Match(l, current, query)
}

// Printable returns a lowercase printable key name, or empty for control keys.
func Printable(name key.Name) string { return navigation.Printable(name) }
