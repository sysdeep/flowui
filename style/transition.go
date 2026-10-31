package style

import "time"

type PropertyID uint8

const (
	PropBackgroundColor PropertyID = iota
	PropTextColor
	PropBorderColor
	PropOutlineColor
	PropOpacity
	PropRadius
	PropTransform
	propertyIDCount
)

// Animatable reports whether property is in the Transition whitelist.
func (p PropertyID) Animatable() bool {
	return p < propertyIDCount
}

// EaseFunc maps normalized time to animation progress.
type EaseFunc func(float32) float32

type Transition struct {
	Property PropertyID
	Duration time.Duration
	Delay    time.Duration
	Ease     EaseFunc
}

type TransitionOption func(*Transition)

func TransitionDelay(value time.Duration) TransitionOption {
	return func(transition *Transition) {
		transition.Delay = max(value, 0)
	}
}

func TransitionEase(value EaseFunc) TransitionOption {
	return func(transition *Transition) {
		transition.Ease = value
	}
}

func (s Style) Transition(property PropertyID, duration time.Duration, options ...TransitionOption) Style {
	// Soft-ignore properties outside the animation whitelist (spec 07).
	if !property.Animatable() {
		return s
	}
	s.transitions = append([]Transition(nil), s.transitions...)
	transition := transition(&s.transitions, property)
	transition.Duration = max(duration, 0)
	for _, option := range options {
		if option != nil {
			option(transition)
		}
	}
	return s
}

func transition(transitions *[]Transition, property PropertyID) *Transition {
	for index := range *transitions {
		if (*transitions)[index].Property == property {
			return &(*transitions)[index]
		}
	}
	*transitions = append(*transitions, Transition{Property: property})
	return &(*transitions)[len(*transitions)-1]
}
