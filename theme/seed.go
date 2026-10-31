package theme

import (
	"image/color"
	"time"
)

// Seed customizes colors and motion within FlowUI's default appearance.
// The zero value is valid: a grey, motionless theme.
type Seed struct {
	Primary color.NRGBA // brand hue, any hue angle
	Neutral color.NRGBA // neutral base; decides the cool / warm cast
	Danger  color.NRGBA // an unset hue takes the primary's ramp
	Success color.NRGBA
	Warning color.NRGBA
	Info    color.NRGBA

	Motion MotionSeed
}

// MotionSeed is the motion half of a Seed. The zero value turns animation off.
type MotionSeed struct {
	Enabled  bool
	Duration time.Duration // state-feedback baseline
	Ease     EaseKind
}

// DefaultSeed returns the colors and motion used by Light and Dark.
func DefaultSeed() Seed {
	return Seed{
		Primary: rgb(0x3e63dd),
		Danger:  rgb(0xe5484d),
		Success: rgb(0x30a46c),
		Warning: rgb(0xffc53d),
		Info:    rgb(0x3e63dd),

		Motion: MotionSeed{Enabled: true, Duration: 140 * time.Millisecond},
	}
}
