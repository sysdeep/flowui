// Package fieldstyle shares the built-in field appearance between Input and DatePicker.
// This widget-private helper owns no behavior; components supply geometry and layout.
package fieldstyle

type Variant uint8

const (
	Primary Variant = iota
	Secondary
	// Filled is a resting gray surface with no border.
	Filled
	// Borderless is a transparent surface.
	Borderless
	// Underlined draws only a bottom border.
	Underlined
)
