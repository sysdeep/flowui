package input

import (
	"fmt"
	"strings"

	"github.com/qianniancn/flowui/style"
)

func normalizeResolvedStyle(value style.ResolvedStyle) style.ResolvedStyle {
	for i := range value.Transitions {
		value.Transitions[i].Ease = nil
	}
	return value
}

func partName(part style.Part) string {
	if part == style.PartRoot {
		return "root"
	}
	return string(part)
}

// describeStyle renders a resolved style field by field, so a failure message names
// the property that differs instead of dumping two opaque structs.
func describeStyle(r style.ResolvedStyle) string {
	var b strings.Builder
	b.WriteString("box{")
	if r.Box != nil {
		fmt.Fprintf(&b, "w=%s h=%s minW=%s maxW=%s minH=%s maxH=%s fillW=%s fillH=%s aspect=%s pad=%s margin=%s overflow=%s cursor=%s",
			ptrString(r.Box.Width), ptrString(r.Box.Height), ptrString(r.Box.MinWidth), ptrString(r.Box.MaxWidth),
			ptrString(r.Box.MinHeight), ptrString(r.Box.MaxHeight), ptrString(r.Box.FillWidth), ptrString(r.Box.FillHeight),
			ptrString(r.Box.AspectRatio), ptrString(r.Box.Padding), ptrString(r.Box.Margin),
			ptrString(r.Box.Overflow), ptrString(r.Box.Cursor))
	}
	b.WriteString("} paint{")
	if r.Paint != nil {
		fmt.Fprintf(&b, "bg=%s border=%s bottom=%s radius=%s radii=%s shadows=%s outline=%s outset=%s opacity=%s",
			describePaintSource(r.Paint.Background), describeBorder(r.Paint.Border), describeBorder(r.Paint.BorderBottom),
			ptrString(r.Paint.Radius), ptrString(r.Paint.Radii), describeShadows(r.Paint.Shadows),
			describeOutline(r.Paint.Outline), ptrString(r.Paint.VisualOutset), ptrString(r.Paint.Opacity))
	}
	b.WriteString("} text{")
	if r.Text != nil {
		fmt.Fprintf(&b, "color=%s size=%s weight=%s typeface=%s style=%s lineHeight=%s lineHeightScale=%s maxLines=%s align=%s wrap=%s truncator=%s",
			describeColorSource(r.Text.Color), ptrString(r.Text.FontSize), ptrString(r.Text.FontWeight),
			ptrString(r.Text.Typeface), ptrString(r.Text.FontStyle), ptrString(r.Text.LineHeight),
			ptrString(r.Text.LineHeightScale), ptrString(r.Text.MaxLines), ptrString(r.Text.Align),
			ptrString(r.Text.Wrap), ptrString(r.Text.Truncator))
	}
	b.WriteString("} transform{")
	if r.Trans != nil {
		fmt.Fprintf(&b, "translateX=%s translateY=%s scaleX=%s scaleY=%s rotate=%s",
			ptrString(r.Trans.TranslateX), ptrString(r.Trans.TranslateY),
			ptrString(r.Trans.ScaleX), ptrString(r.Trans.ScaleY), ptrString(r.Trans.Rotate))
	}
	b.WriteString("} transitions[")
	for _, transition := range r.Transitions {
		fmt.Fprintf(&b, "%d/%v/%v ", transition.Property, transition.Duration, transition.Delay)
	}
	b.WriteString("]")
	return b.String()
}

func describeBorder(b *style.BorderStyle) string {
	if b == nil {
		return "-"
	}
	return fmt.Sprintf("%s/%s", describeColorSource(b.Color), ptrString(b.Width))
}

func describeOutline(o *style.OutlineStyle) string {
	if o == nil {
		return "-"
	}
	return fmt.Sprintf("w=%v offset=%v color=%s", o.Width, o.Offset, describeColorSource(o.Color))
}

func describeShadows(shadows []style.Shadow) string {
	if len(shadows) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(shadows))
	for _, shadow := range shadows {
		parts = append(parts, fmt.Sprintf("(%v,%v,%v,%v,%s,%s)",
			shadow.OffsetX, shadow.OffsetY, shadow.Blur, shadow.Spread,
			describeColorSource(shadow.Color), ptrString(shadow.Profile)))
	}
	return strings.Join(parts, ";")
}

func describePaintSource(source style.PaintSource) string {
	switch value := source.(type) {
	case nil:
		return "-"
	case style.SolidColor:
		return fmt.Sprintf("solid%v", value.Color)
	case *style.SolidColor:
		if value == nil {
			return "-"
		}
		return fmt.Sprintf("solid%v*", value.Color)
	case style.Gradient:
		return fmt.Sprintf("gradient%v/%d", value.AngleDegrees, len(value.Stops))
	default:
		return fmt.Sprintf("%T", source)
	}
}

func describeColorSource(source style.ColorSource) string {
	if source == nil {
		return "-"
	}
	return describePaintSource(source)
}

func ptrString[T any](value *T) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(*value)
}
