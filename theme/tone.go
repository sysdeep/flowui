package theme

// Tone is the semantic hue dimension shared by component visual recipes.
type Tone uint8

const (
	ToneDefault Tone = iota
	TonePrimary
	ToneSecondary
	ToneHover
	ToneSuccess
	ToneWarning
	ToneDanger
)

const ToneCount = int(ToneDanger) + 1

// ToneIsHue reports whether a tone draws its fill from a hue ramp.
func ToneIsHue(t Tone) bool {
	switch t {
	case TonePrimary, ToneSecondary, ToneSuccess, ToneWarning, ToneDanger:
		return true
	}
	return false
}

// String names the tone for diagnostics and test failures.
func (t Tone) String() string {
	switch t {
	case ToneDefault:
		return "default"
	case TonePrimary:
		return "primary"
	case ToneSecondary:
		return "secondary"
	case ToneHover:
		return "hover"
	case ToneSuccess:
		return "success"
	case ToneWarning:
		return "warning"
	case ToneDanger:
		return "danger"
	}
	return "tone(?)"
}

// Scale returns the ramp a tone draws from. ToneHover is a neutral surface,
// so it reads the neutral ramp rather than a hue of its own.
func (s Scales) Scale(t Tone) ColorScale {
	switch t {
	case TonePrimary:
		return s.Primary
	case ToneSecondary:
		return s.Secondary
	case ToneSuccess:
		return s.Success
	case ToneWarning:
		return s.Warning
	case ToneDanger:
		return s.Danger
	default:
		return s.Neutral
	}
}
