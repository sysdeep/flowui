package style

func clone(source Style) Style {
	result := source
	result.box = cloneBox(source.box)
	result.paint = clonePaint(source.paint)
	result.text = cloneText(source.text)
	result.trans = cloneTransform(source.trans)
	// Pre-allocate slices with exact capacity
	if len(source.transitions) > 0 {
		result.transitions = make([]Transition, len(source.transitions))
		copy(result.transitions, source.transitions)
	} else {
		result.transitions = nil
	}
	result.dim = clonePtr(source.dim)
	result.parts = cloneParts(source.parts)
	if len(source.rules) > 0 {
		result.rules = make([]Rule, len(source.rules))
		copy(result.rules, source.rules)
	} else {
		result.rules = nil
	}
	return result
}

func cloneParts(source map[Part]Style) map[Part]Style {
	if len(source) == 0 {
		return nil
	}
	result := make(map[Part]Style, len(source))
	for part, declaration := range source {
		result[part] = clone(declaration)
	}
	return result
}

func cloneBox(source *BoxStyle) *BoxStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Width = clonePtr(source.Width)
	result.Height = clonePtr(source.Height)
	result.MinWidth = clonePtr(source.MinWidth)
	result.MaxWidth = clonePtr(source.MaxWidth)
	result.MinHeight = clonePtr(source.MinHeight)
	result.MaxHeight = clonePtr(source.MaxHeight)
	result.FillWidth = clonePtr(source.FillWidth)
	result.FillHeight = clonePtr(source.FillHeight)
	result.AspectRatio = clonePtr(source.AspectRatio)
	result.Padding = clonePtr(source.Padding)
	result.Margin = clonePtr(source.Margin)
	result.Overflow = clonePtr(source.Overflow)
	result.Cursor = clonePtr(source.Cursor)
	result.paddingMask = source.paddingMask
	result.marginMask = source.marginMask
	return &result
}

func clonePaint(source *PaintStyle) *PaintStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Background = clonePaintSource(source.Background)
	result.backgroundSet = source.backgroundSet
	if source.Border != nil {
		border := *source.Border
		border.Color = cloneColorSource(source.Border.Color)
		border.Width = clonePtr(source.Border.Width)
		border.LineStyle = clonePtr(source.Border.LineStyle)
		result.Border = &border
	}
	if source.BorderBottom != nil {
		border := *source.BorderBottom
		border.Color = cloneColorSource(source.BorderBottom.Color)
		border.Width = clonePtr(source.BorderBottom.Width)
		border.LineStyle = clonePtr(source.BorderBottom.LineStyle)
		result.BorderBottom = &border
	}
	result.Radius = clonePtr(source.Radius)
	result.Radii = clonePtr(source.Radii)
	result.radiusMask = source.radiusMask
	result.Shadows = cloneShadows(source.Shadows)
	result.shadowsSet = source.shadowsSet
	if source.Outline != nil {
		outline := *source.Outline
		outline.Color = cloneColorSource(source.Outline.Color)
		result.Outline = &outline
	}
	result.VisualOutset = clonePtr(source.VisualOutset)
	result.visualOutsetSet = source.visualOutsetSet
	result.Opacity = clonePtr(source.Opacity)
	return &result
}

//go:inline
func cloneShadows(source []Shadow) []Shadow {
	if len(source) == 0 {
		return nil
	}
	result := make([]Shadow, len(source))
	for index, shadow := range source {
		result[index] = shadow
		result[index].Color = cloneColorSource(shadow.Color)
		result[index].Profile = clonePtr(shadow.Profile)
	}
	return result
}

func cloneText(source *TextStyle) *TextStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Color = cloneColorSource(source.Color)
	result.FontSize = clonePtr(source.FontSize)
	result.FontWeight = clonePtr(source.FontWeight)
	result.Typeface = clonePtr(source.Typeface)
	result.FontStyle = clonePtr(source.FontStyle)
	result.LineHeight = clonePtr(source.LineHeight)
	result.LineHeightScale = clonePtr(source.LineHeightScale)
	result.MaxLines = clonePtr(source.MaxLines)
	result.Align = clonePtr(source.Align)
	result.Wrap = clonePtr(source.Wrap)
	result.Truncator = clonePtr(source.Truncator)
	result.Underline = clonePtr(source.Underline)
	return &result
}

func cloneTransform(source *TransformStyle) *TransformStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.TranslateX = clonePtr(source.TranslateX)
	result.TranslateY = clonePtr(source.TranslateY)
	result.ScaleX = clonePtr(source.ScaleX)
	result.ScaleY = clonePtr(source.ScaleY)
	result.Rotate = clonePtr(source.Rotate)
	return &result
}

func clonePaintSource(source PaintSource) PaintSource {
	switch value := source.(type) {
	case Gradient:
		value.Stops = cloneGradientStops(value.Stops)
		return value
	case *Gradient:
		if value == nil {
			return nil
		}
		copy := *value
		copy.Stops = cloneGradientStops(value.Stops)
		return &copy
	case *SolidColor:
		if value == nil {
			return nil
		}
		copy := *value
		return &copy
	default:
		return source
	}
}

//go:inline
func cloneGradientStops(source []GradientStop) []GradientStop {
	if len(source) == 0 {
		return nil
	}
	result := make([]GradientStop, len(source))
	for index, stop := range source {
		result[index] = GradientStop{Offset: stop.Offset, Color: cloneColorSource(stop.Color)}
	}
	return result
}

func cloneColorSource(source ColorSource) ColorSource {
	if source == nil {
		return nil
	}
	cloned := clonePaintSource(source)
	if cloned == nil {
		return nil
	}
	return cloned.(ColorSource)
}

//go:inline
func clonePtr[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

//go:inline
func pick[T any](current, incoming *T) *T {
	if incoming == nil {
		return current
	}
	return clonePtr(incoming)
}
