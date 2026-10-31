package fieldstyle

import (
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func surfaceColors(activeTheme *theme.Theme, variant Variant) (background, hover style.ColorSource) {
	semantic := activeTheme.Semantic()
	neutral := semantic.Tone(theme.ToneDefault)
	background = style.Solid(semantic.FieldBackground)
	hover = style.Solid(semantic.FieldHover)
	switch variant {
	case Secondary, Filled:
		background, hover = style.Solid(neutral.Fill), style.Solid(neutral.Hover)
	case Borderless:
		background, hover = style.RGBA(0), style.RGBA(0)
	case Underlined:
		hover = background
	}
	return background, hover
}

func disabledSet(activeTheme *theme.Theme, variant Variant) style.Style {
	sem := activeTheme.Semantic()
	if sem.DisabledBackground.A == 0 {
		return style.Style{}.Opacity(activeTheme.DisabledOpacityValue())
	}

	set := style.Style{}.
		TextColor(style.Solid(sem.DisabledForeground)).
		Cursor(pointer.CursorNotAllowed).
		Opacity(1)
	if variant == Borderless {
		return set.Background(style.RGBA(0))
	}
	return set.Background(style.Solid(sem.DisabledBackground)).
		BorderColor(style.Solid(sem.DisabledBorder))
}

// readOnlySet is the disabled face without the disabled interaction: a read-only
// field stays focusable and selectable, so it keeps the text cursor and the
// focus ring takes over while it is focused.
func readOnlySet(activeTheme *theme.Theme, variant Variant) style.Style {
	sem := activeTheme.Semantic()
	set := style.Style{}.
		Background(style.Solid(sem.DisabledBackground)).
		TextColor(style.Solid(sem.DisabledForeground)).
		BorderColor(style.Solid(sem.DisabledBorder)).
		Cursor(pointer.CursorText).
		Opacity(1)
	if variant == Borderless {
		return set.Background(style.RGBA(0))
	}
	return set
}
