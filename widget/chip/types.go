package chip

type Color uint8

const (
	ColorDefault Color = iota
	ColorAccent
	ColorSuccess
	ColorWarning
	ColorDanger
)

type Variant uint8

const (
	VariantSecondary Variant = iota
	VariantPrimary
	VariantTertiary
	VariantSoft
)

type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)
