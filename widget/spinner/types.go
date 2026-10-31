package spinner

type Color uint8

const (
	Accent Color = iota
	Current
	Success
	Warning
	Danger
)

type Size uint8

const (
	Medium Size = iota
	Small
	Large
	ExtraLarge
)
