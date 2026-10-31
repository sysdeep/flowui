package toast

// Variant identifies a toast's semantic visual treatment.
type Variant uint8

const (
	Default Variant = iota
	Accent
	Success
	Warning
	Danger
)

// Placement identifies where the toast stack anchors inside the viewport.
type Placement uint8

const (
	Bottom Placement = iota
	BottomStart
	BottomEnd
	Top
	TopStart
	TopEnd
)
