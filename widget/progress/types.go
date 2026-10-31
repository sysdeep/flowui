package progress

type BarColor uint8

const (
	BarAccent BarColor = iota
	BarDefault
	BarSuccess
	BarWarning
	BarDanger
)

type BarSize uint8

const (
	BarMedium BarSize = iota
	BarSmall
	BarLarge
)
