package style

import "maps"

func (s Style) resolve(state State) Style {
	result := clone(s)
	result.rules = nil
	// State rule rows first: they are a table, so their precedence comes from
	// Rule.Rank and matched-bit count rather than from nesting depth.
	for _, rule := range orderedRules(s.rules, state.Bits()) {
		override := rule.Set.resolve(state)
		if rule.Tween != nil {
			override = override.Transition(
				rule.Tween.Property,
				rule.Tween.Duration,
				TransitionDelay(rule.Tween.Delay),
			)
		}
		result = merge(result, override)
	}
	return result
}

// Resolve applies this declaration's matching conditional rules.
func (s Style) Resolve(state State) ResolvedStyle {
	return resolvedStyle(s.resolve(state))
}

func Cascade(state State, layers ...Style) ResolvedStyle {
	var result Style
	for _, layer := range layers {
		result = merge(result, layer.resolve(state))
	}
	return resolvedStyle(result)
}

// Join combines declarations without resolving conditions.
// It is used internally when component defaults are assembled in stages.
func Join(styles ...Style) Style {
	var result Style
	for _, declaration := range styles {
		result = merge(result, declaration)
	}
	return result
}

// JoinPart combines a base declaration with an override that only targets one
// named part. Unlike Join, it keeps unrelated part declarations shared and
// clones only the selected part. This is useful for per-cell recipes in
// compound widgets where copying every part would dominate frame allocations.
func JoinPart(base, override Style, part Part) Style {
	if len(override.parts) == 0 {
		return base
	}
	overridePart, ok := override.parts[part]
	if !ok {
		return base
	}
	result := base
	result.parts = make(map[Part]Style, len(base.parts))
	maps.Copy(result.parts, base.parts)
	result.parts[part] = merge(base.parts[part], overridePart)
	return result
}

// CascadePart resolves a named component part. Text properties inherit from
// each root layer; box, paint, and transform properties do not.
func CascadePart(state State, part Part, layers ...Style) ResolvedStyle {
	if part == PartRoot {
		return Cascade(state, layers...)
	}
	// Root text is an inherited fallback for a named part. Resolve all root
	// text layers first, then apply the part declarations on top so a part's
	// explicit state color (for example a selected date cell) cannot be
	// overwritten by a later root foreground.
	var inherited Style
	var partResult Style
	for _, layer := range layers {
		resolved := layer.resolve(state)
		if resolved.text != nil {
			rootText := Style{text: cloneText(resolved.text)}
			// Pre-allocate for text color transitions if any exist
			for _, transition := range resolved.transitions {
				if transition.Property == PropTextColor {
					if rootText.transitions == nil {
						rootText.transitions = make([]Transition, 0, 1)
					}
					rootText.transitions = append(rootText.transitions, transition)
				}
			}
			// Keep root text as the fallback layer; named-part text is merged
			// after this loop and therefore has higher precedence.
			inherited = merge(inherited, rootText)
		}
		if declaration, ok := resolved.parts[part]; ok {
			partResult = merge(partResult, declaration.resolve(state))
		}
	}
	result := merge(inherited, partResult)
	result.parts = nil
	return resolvedStyle(result)
}

// TextDeclaration copies resolved text properties into an inheritable
// declaration without exposing Style internals.
func TextDeclaration(value *TextStyle) Style {
	return Style{text: cloneText(value)}
}

func resolvedStyle(source Style) ResolvedStyle {
	// Pre-allocate transitions slice with exact capacity
	var transitions []Transition
	if len(source.transitions) > 0 {
		transitions = make([]Transition, len(source.transitions))
		copy(transitions, source.transitions)
	}
	result := ResolvedStyle{
		Box:         cloneBox(source.box),
		Paint:       clonePaint(source.paint),
		Text:        cloneText(source.text),
		Trans:       cloneTransform(source.trans),
		Transitions: transitions,
	}
	// Dim is applied here rather than inside resolve because this is the
	// resolution boundary: resolve also runs for a rule row's Set, and a row
	// that asks to dim must dim the merged result, not only its own fields.
	if source.dim != nil {
		dimResolved(&result, *source.dim)
	}
	return result
}

// dimResolved scales the alpha of every color in a resolved snapshot. Shadow
// layers are skipped on purpose -- see Style.Dim.
func dimResolved(resolved *ResolvedStyle, factor float32) {
	if paint := resolved.Paint; paint != nil {
		paint.Background = dimPaintSource(paint.Background, factor)
		if paint.Border != nil {
			paint.Border.Color = dimColorSource(paint.Border.Color, factor)
		}
		if paint.BorderBottom != nil {
			paint.BorderBottom.Color = dimColorSource(paint.BorderBottom.Color, factor)
		}
		if paint.Outline != nil {
			paint.Outline.Color = dimColorSource(paint.Outline.Color, factor)
		}
	}
	if text := resolved.Text; text != nil {
		text.Color = dimColorSource(text.Color, factor)
	}
}

func dimPaintSource(source PaintSource, factor float32) PaintSource {
	switch value := source.(type) {
	case nil:
		return nil
	case Gradient:
		value.Stops = dimGradientStops(value.Stops, factor)
		return value
	case *Gradient:
		if value == nil {
			return nil
		}
		copied := *value
		copied.Stops = dimGradientStops(value.Stops, factor)
		return &copied
	case ColorSource:
		return dimColorSource(value, factor)
	default:
		return source
	}
}

func dimGradientStops(stops []GradientStop, factor float32) []GradientStop {
	if len(stops) == 0 {
		return stops
	}
	result := make([]GradientStop, len(stops))
	for index, stop := range stops {
		result[index] = GradientStop{Offset: stop.Offset, Color: dimColorSource(stop.Color, factor)}
	}
	return result
}

// dimColorSource multiplies a color source's alpha.
//
// A concrete color is scaled in place and stays concrete. That is not an
// optimization: it is what makes Dim(DisabledOpacity) on a solid color
// indistinguishable from the DisabledColor arithmetic the built-in components
// already perform, so the two can be compared field by field.
//
// ColorSource has exactly one implementation (SolidColor); the default branch
// is unreachable and only satisfies the compiler.
func dimColorSource(source ColorSource, factor float32) ColorSource {
	switch value := source.(type) {
	case nil:
		return nil
	case SolidColor:
		value.Color.A = ScaleAlpha(value.Color.A, factor)
		return value
	case *SolidColor:
		if value == nil {
			return nil
		}
		copied := *value
		copied.Color.A = ScaleAlpha(copied.Color.A, factor)
		return &copied
	default:
		return source
	}
}
