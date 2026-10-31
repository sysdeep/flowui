// Package textrender contains Gio text adapters shared by component visuals.
package textrender

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Label renders a resolved text part using the active theme's Gio material bridge.
func Label(active *theme.Theme, gtx layout.Context, value string, resolved style.ResolvedStyle, size unit.Sp) layout.Dimensions {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	if resolved.Text == nil {
		panic("flowui/textrender: missing text style")
	}
	if size <= 0 {
		size = active.Typography.BodySize
		if resolved.Text.FontSize != nil && *resolved.Text.FontSize > 0 {
			size = *resolved.Text.FontSize
		}
	}
	label := material.Label(theme.MaterialOf(active), size, value)
	label.SelectionColor = active.Semantic().Selection
	if value, ok := computed.Color(resolved.Text.Color); ok {
		label.Color = value
	} else {
		panic("flowui/textrender: unresolved text color")
	}
	if resolved.Text.FontWeight != nil {
		label.Font.Weight = font.Weight(*resolved.Text.FontWeight)
	}
	if resolved.Text.Typeface != nil {
		label.Font.Typeface = *resolved.Text.Typeface
	}
	if resolved.Text.FontStyle != nil {
		label.Font.Style = *resolved.Text.FontStyle
	}
	if resolved.Text.Align != nil {
		label.Alignment = alignment(*resolved.Text.Align)
	}
	if resolved.Text.MaxLines != nil {
		label.MaxLines = *resolved.Text.MaxLines
	}
	if resolved.Text.LineHeight != nil {
		label.LineHeight = *resolved.Text.LineHeight
	}
	if resolved.Text.LineHeightScale != nil {
		label.LineHeightScale = *resolved.Text.LineHeightScale
	}
	if resolved.Text.Wrap != nil {
		label.WrapPolicy = *resolved.Text.Wrap
	} else {
		label.WrapPolicy = giotext.WrapWords
	}
	if resolved.Text.Truncator != nil {
		label.Truncator = *resolved.Text.Truncator
	}
	return label.Layout(gtx)
}

func alignment(value style.TextAlign) giotext.Alignment {
	switch value {
	case style.TextAlignCenter:
		return giotext.Middle
	case style.TextAlignEnd:
		return giotext.End
	default:
		return giotext.Start
	}
}

// ColorValue reports whether a resolved text Part carries a concrete color.
func ColorValue(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	return computed.Color(resolved.Text.Color)
}

// ColorOr resolves a text Part while retaining a caller-owned fallback for
// declarations that intentionally inherit their parent typography.
func ColorOr(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if value, ok := ColorValue(resolved); ok {
		return value
	}
	return fallback
}
