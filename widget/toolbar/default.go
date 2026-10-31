package toolbar

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	toolbarDefaultGap             unit.Dp = 8
	toolbarDefaultPadding         unit.Dp = 4
	toolbarDefaultRadius          unit.Dp = 24
	toolbarDefaultSeparatorLength unit.Dp = 20
	toolbarDefaultSeparatorWidth  unit.Dp = 1
)

type toolbarAxis uint8

const (
	toolbarHorizontal toolbarAxis = iota
	toolbarVertical
)

// toolbarMetrics contains component visual geometry used by toolbar behavior.
type toolbarMetrics struct {
	Gap unit.Dp
}

// toolbarVisual is the resolved default appearance of a toolbar.
type toolbarVisual struct {
	Defaults          style.Style
	Variant           style.Style
	Separator         style.Style
	VerticalSeparator style.Style
	Metrics           toolbarMetrics
}

// buildToolbar builds the component-owned toolbar appearance.
func buildToolbar(active *theme.Theme, attached bool) toolbarVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	variant := style.Style{}
	if attached {
		variant = style.Style{}.
			Background(style.Solid(sem.Surface)).
			TextColor(style.Solid(sem.SurfaceForeground)).
			Padding(toolbarDefaultPadding).
			Radius(toolbarDefaultRadius).
			Shadow(style.ShadowOverlay)
	}
	separator := style.Style{}.
		Width(toolbarDefaultSeparatorWidth).
		Height(toolbarDefaultSeparatorLength).
		Background(style.Solid(sem.Separator))
	verticalSeparator := style.Style{}.
		Width(toolbarDefaultSeparatorLength * 2).
		Height(toolbarDefaultSeparatorWidth).
		Background(style.Solid(sem.Separator))
	return toolbarVisual{
		Defaults:          style.Style{},
		Variant:           variant,
		Separator:         separator,
		VerticalSeparator: verticalSeparator,
		Metrics:           toolbarMetrics{Gap: toolbarDefaultGap},
	}
}
