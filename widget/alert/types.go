package alert

type Status uint8

const (
	Default Status = iota
	Accent
	Success
	Warning
	Danger
)
