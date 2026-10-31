package toggle

type CheckboxVariant uint8

const (
	CheckboxPrimary CheckboxVariant = iota
	CheckboxSecondary
)

type SwitchSize uint8

const (
	SwitchMedium SwitchSize = iota
	SwitchSmall
	SwitchLarge
)

type ToggleButtonVariant uint8

const (
	ToggleButtonDefault ToggleButtonVariant = iota
	ToggleButtonGhost
)

type ToggleButtonSize uint8

const (
	ToggleButtonMedium ToggleButtonSize = iota
	ToggleButtonSmall
	ToggleButtonLarge
)
