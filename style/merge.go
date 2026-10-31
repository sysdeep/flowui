package style

func merge(first, second Style) Style {
	// Fast path: second is empty
	if second.box == nil && second.paint == nil && second.text == nil && second.trans == nil &&
		len(second.transitions) == 0 && len(second.parts) == 0 &&
		len(second.rules) == 0 && second.dim == nil {
		return clone(first)
	}

	result := first

	// One slot-merge function per storage slot: box, paint, text, transform,
	// transitions. Non-slot members follow below.
	result.box = mergeBoxSlot(first.box, second.box)
	result.paint = mergePaintSlot(first.paint, second.paint)
	result.text = mergeTextSlot(first.text, second.text)
	result.trans = mergeTransSlot(first.trans, second.trans)
	result.transitions = mergeMotionSlot(first.transitions, second.transitions)

	if len(second.rules) != 0 {
		// Concatenation, not precedence merging: a later layer's rules must be
		// able to override an earlier layer's rules of equal specificity, which
		// is exactly what the declaration-order tiebreak in orderedRules does.
		result.rules = make([]Rule, len(first.rules), len(first.rules)+len(second.rules))
		copy(result.rules, first.rules)
		result.rules = append(result.rules, second.rules...)
	}
	if len(second.parts) != 0 {
		result.parts = mergeParts(first.parts, second.parts)
	}
	if second.dim != nil {
		// A later layer's dim replaces rather than composes. Two multipliers
		// would be a product, and the declaration that says "dim by a half"
		// should mean a half, not "a half of whatever the layer below asked for".
		result.dim = clonePtr(second.dim)
	}
	return result
}

// The five slot-merge functions below share one shape: a nil incoming side
// keeps the current slot untouched, an empty current slot adopts a clone of
// the incoming side, and otherwise the merged result is a fresh clone so the
// two input declarations are never aliased.

func mergeBoxSlot(first, second *BoxStyle) *BoxStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneBox(second)
	}
	merged := cloneBox(first)
	mergeBox(merged, second)
	return merged
}

func mergePaintSlot(first, second *PaintStyle) *PaintStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return clonePaint(second)
	}
	merged := clonePaint(first)
	mergePaint(merged, second)
	return merged
}

func mergeTextSlot(first, second *TextStyle) *TextStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneText(second)
	}
	merged := cloneText(first)
	mergeText(merged, second)
	return merged
}

func mergeTransSlot(first, second *TransformStyle) *TransformStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneTransform(second)
	}
	merged := cloneTransform(first)
	mergeTransform(merged, second)
	return merged
}

func mergeMotionSlot(first, second []Transition) []Transition {
	if len(second) == 0 {
		return first
	}
	return mergeTransitions(first, second)
}

func mergeParts(first, second map[Part]Style) map[Part]Style {
	if len(first) == 0 && len(second) == 0 {
		return nil
	}
	result := cloneParts(first)
	if result == nil {
		result = make(map[Part]Style, len(second))
	}
	for part, declaration := range second {
		result[part] = merge(result[part], declaration)
	}
	return result
}

func mergeBox(dst, src *BoxStyle) {
	dst.Width = pick(dst.Width, src.Width)
	dst.Height = pick(dst.Height, src.Height)
	dst.MinWidth = pick(dst.MinWidth, src.MinWidth)
	dst.MaxWidth = pick(dst.MaxWidth, src.MaxWidth)
	dst.MinHeight = pick(dst.MinHeight, src.MinHeight)
	dst.MaxHeight = pick(dst.MaxHeight, src.MaxHeight)
	dst.FillWidth = pick(dst.FillWidth, src.FillWidth)
	dst.FillHeight = pick(dst.FillHeight, src.FillHeight)
	dst.AspectRatio = pick(dst.AspectRatio, src.AspectRatio)
	mergeInsets(&dst.Padding, &dst.paddingMask, src.Padding, src.paddingMask)
	mergeInsets(&dst.Margin, &dst.marginMask, src.Margin, src.marginMask)
	dst.Overflow = pick(dst.Overflow, src.Overflow)
	dst.Cursor = pick(dst.Cursor, src.Cursor)
}

func mergeInsets(dst **Insets, dstMask *uint8, src *Insets, srcMask uint8) {
	if src == nil {
		return
	}
	if *dst == nil {
		*dst = &Insets{}
	}
	if srcMask == 0 {
		srcMask = sidesAll
	}
	if srcMask&sideTop != 0 {
		(*dst).Top = src.Top
	}
	if srcMask&sideRight != 0 {
		(*dst).Right = src.Right
	}
	if srcMask&sideBottom != 0 {
		(*dst).Bottom = src.Bottom
	}
	if srcMask&sideLeft != 0 {
		(*dst).Left = src.Left
	}
	*dstMask |= srcMask
}

func mergePaint(dst, src *PaintStyle) {
	if src.backgroundSet || src.Background != nil {
		dst.Background = clonePaintSource(src.Background)
		dst.backgroundSet = true
	}
	if src.Border != nil {
		if dst.Border == nil {
			dst.Border = &BorderStyle{}
		}
		if src.Border.Color != nil {
			dst.Border.Color = cloneColorSource(src.Border.Color)
		}
		dst.Border.Width = pick(dst.Border.Width, src.Border.Width)
		dst.Border.LineStyle = pick(dst.Border.LineStyle, src.Border.LineStyle)
	}
	if src.BorderBottom != nil {
		if dst.BorderBottom == nil {
			dst.BorderBottom = &BorderStyle{}
		}
		if src.BorderBottom.Color != nil {
			dst.BorderBottom.Color = cloneColorSource(src.BorderBottom.Color)
		}
		dst.BorderBottom.Width = pick(dst.BorderBottom.Width, src.BorderBottom.Width)
		dst.BorderBottom.LineStyle = pick(dst.BorderBottom.LineStyle, src.BorderBottom.LineStyle)
	}
	if src.Radius != nil {
		dst.Radius = clonePtr(src.Radius)
		dst.Radii = nil
		dst.radiusMask = 0
	}
	mergeRadii(dst, src)
	if src.shadowsSet {
		dst.Shadows = cloneShadows(src.Shadows)
		dst.shadowsSet = true
	}
	if src.Outline != nil {
		outline := *src.Outline
		outline.Color = cloneColorSource(src.Outline.Color)
		dst.Outline = &outline
	}
	if src.visualOutsetSet {
		dst.VisualOutset = clonePtr(src.VisualOutset)
		dst.visualOutsetSet = true
	}
	dst.Opacity = pick(dst.Opacity, src.Opacity)
}

func mergeRadii(dst, src *PaintStyle) {
	if src.Radii == nil {
		return
	}
	base := CornerRadii{}
	if dst.Radii != nil {
		base = *dst.Radii
	} else if dst.Radius != nil {
		base = uniformRadii(*dst.Radius)
	}
	mask := src.radiusMask
	if mask == 0 {
		mask = 0x0f
	}
	if mask&(1<<0) != 0 {
		base.TopLeft = src.Radii.TopLeft
	}
	if mask&(1<<1) != 0 {
		base.TopRight = src.Radii.TopRight
	}
	if mask&(1<<2) != 0 {
		base.BottomRight = src.Radii.BottomRight
	}
	if mask&(1<<3) != 0 {
		base.BottomLeft = src.Radii.BottomLeft
	}
	dst.Radii = &base
	dst.radiusMask |= mask
}

func mergeText(dst, src *TextStyle) {
	if src.Color != nil {
		dst.Color = cloneColorSource(src.Color)
	}
	dst.FontSize = pick(dst.FontSize, src.FontSize)
	dst.FontWeight = pick(dst.FontWeight, src.FontWeight)
	dst.Typeface = pick(dst.Typeface, src.Typeface)
	dst.FontStyle = pick(dst.FontStyle, src.FontStyle)
	dst.LineHeight = pick(dst.LineHeight, src.LineHeight)
	dst.LineHeightScale = pick(dst.LineHeightScale, src.LineHeightScale)
	dst.MaxLines = pick(dst.MaxLines, src.MaxLines)
	dst.Align = pick(dst.Align, src.Align)
	dst.Wrap = pick(dst.Wrap, src.Wrap)
	dst.Truncator = pick(dst.Truncator, src.Truncator)
	dst.Underline = pick(dst.Underline, src.Underline)
}

func mergeTransform(dst, src *TransformStyle) {
	dst.TranslateX = pick(dst.TranslateX, src.TranslateX)
	dst.TranslateY = pick(dst.TranslateY, src.TranslateY)
	dst.ScaleX = pick(dst.ScaleX, src.ScaleX)
	dst.ScaleY = pick(dst.ScaleY, src.ScaleY)
	dst.Rotate = pick(dst.Rotate, src.Rotate)
}

func mergeTransitions(first, second []Transition) []Transition {
	// Pre-allocate with capacity for worst case (no duplicates)
	result := make([]Transition, len(first), len(first)+len(second))
	copy(result, first)
	for _, incoming := range second {
		found := false
		for index := range result {
			if result[index].Property == incoming.Property {
				result[index] = incoming
				found = true
				break
			}
		}
		if !found {
			result = append(result, incoming)
		}
	}
	return result
}
