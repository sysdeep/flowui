package text

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type descriptionVisual struct{ Defaults style.Style }
type descriptionVisualCache = computed.VisualCache[struct{}, descriptionVisual]

func (d DescriptionWidget) visualWithCache(ctx *core.Context) descriptionVisual {
	active := core.ActiveTheme(ctx)
	if d.visualCache == nil {
		return buildDescriptionVisual(active)
	}
	return d.visualCache.Resolve(active, struct{}{}, func() descriptionVisual { return buildDescriptionVisual(active) })
}

func buildDescriptionVisual(active *theme.Theme) descriptionVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	muted := active.Semantic().MutedForeground
	defaults := style.Style{}.FontSize(active.Typography.SmallSize).
		TextColor(style.Solid(muted)).Rules(style.Rule{When: style.Disabled, Set: style.Style{}.TextColor(style.Solid(active.DisabledColor(muted)))})
	return descriptionVisual{Defaults: defaults}
}
