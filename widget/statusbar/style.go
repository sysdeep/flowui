package statusbar

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type statusBarMetrics struct{ Height, Gap unit.Dp }
type statusBarVisual struct {
	Defaults, Variant style.Style
	Metrics           statusBarMetrics
}

type statusBarVisualCache = computed.VisualCache[struct{}, statusBarVisual]

func (w Widget) visualWithCache(ctx *core.Context) statusBarVisual {
	active := core.ActiveTheme(ctx)
	load := func() statusBarVisual { return resolveStatusBarVisual(active) }
	if w.visualCache == nil {
		return load()
	}
	return w.visualCache.Resolve(active, struct{}{}, load)
}

func resolveStatusBarVisual(active *theme.Theme) statusBarVisual {
	sem := active.Semantic()
	return statusBarVisual{
		Defaults: style.Style{}.BorderColor(style.Solid(sem.Border)).BorderWidth(1).Opacity(1).Height(28).Padding(0, 10).BorderBottom(1, style.Solid(sem.Separator)),
		Variant:  style.Style{}.TextColor(style.Solid(sem.SurfaceSecondaryForeground)).Background(style.Solid(sem.SurfaceSecondary)),
		Metrics:  statusBarMetrics{Height: 28, Gap: 8},
	}
}
