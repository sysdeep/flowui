package base

// Interaction is the pointer, focus and validation state shared by interactive
// bases. Bases embed it in their own state so a Content callback can paint
// hover, press, focus and disabled looks without rebuilding behavior.
type Interaction struct {
	Hovered      bool
	Pressed      bool
	Focused      bool
	FocusVisible bool
	Disabled     bool
	Invalid      bool
	ReadOnly     bool
	Loading      bool
}
