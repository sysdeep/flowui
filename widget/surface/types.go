package surface

type Variant uint8

const (
	Default Variant = iota
	Secondary
	Tertiary
	Transparent
)

type CardVariant uint8

const (
	CardDefault CardVariant = iota
	CardSecondary
	CardTertiary
	CardTransparent
)
