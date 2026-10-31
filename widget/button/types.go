package button

type Variant uint8

const (
	Primary Variant = iota
	Secondary
	Tertiary
	Ghost
	Outline
	Danger
	DangerSoft
	DangerOutline
)

type Size uint8

const (
	Medium Size = iota
	Small
	Large
)
