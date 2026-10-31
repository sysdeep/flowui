package input

import "github.com/qianniancn/flowui/widget/internal/fieldstyle"

type Variant = fieldstyle.Variant

const (
	Primary    = fieldstyle.Primary
	Secondary  = fieldstyle.Secondary
	Filled     = fieldstyle.Filled
	Borderless = fieldstyle.Borderless
	Underlined = fieldstyle.Underlined
)

// Size selects the built-in appearance's height, typography, and spacing tier.
type Size uint8

const (
	Medium Size = iota
	Small
	Large
)

type Type uint8

const (
	Text Type = iota
	Email
	Number
	Password
)
