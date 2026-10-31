package code

import (
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// codeMetrics is the geometry of a code block.
type codeMetrics struct {
	Padding       unit.Dp
	Radius        unit.Dp
	HeaderGap     unit.Dp
	LineNumberGap unit.Dp
	TextSize      unit.Sp
}

// codeVisual is the resolved appearance of a code block.
type codeVisual struct {
	Root    style.Style
	Metrics codeMetrics
}

func buildCodeVisual(active *theme.Theme) codeVisual {
	tokens := codeMetrics{
		Padding: 14, Radius: 6, HeaderGap: 10, LineNumberGap: 12,
		TextSize: 14,
	}
	surface := style.New().
		FillWidth().
		Padding(tokens.Padding).
		Radius(tokens.Radius).
		Background(style.Solid(active.CodeColors().Background))
	return codeVisual{
		Root:    surface,
		Metrics: tokens,
	}
}
