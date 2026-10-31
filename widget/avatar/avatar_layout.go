package avatar

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	imagewidget "github.com/qianniancn/flowui/widget/image"
)

func (a Widget) layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return a.layoutWithVisual(ctx, gtx, a.visualWithCache(ctx))
}

func (a Widget) layoutWithVisual(ctx *core.Context, gtx layout.Context, visual avatarVisual) layout.Dimensions {
	resolved := a.resolveStyle(ctx, gtx, visual)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if label := a.semanticLabel(); label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if a.image.Size() != (image.Point{}) {
			return a.layoutImage(ctx, gtx)
		}
		return a.layoutFallback(ctx, gtx, resolved)
	})
	return layoutui.Resolved(ctx, gtx, resolved.root, content)
}

func (a Widget) layoutImage(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return imagewidget.New(a.image).
		Fit(imagewidget.FitCover).
		Alt(a.semanticLabel()).
		Layout(ctx, gtx)
}

func (a Widget) layoutFallback(ctx *core.Context, gtx layout.Context, resolved avatarResolvedStyle) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.fallback != nil {
			return layoutui.Resolved(ctx, gtx, resolved.label, a.fallback)
		}
		if a.fallbackText != "" {
			return layoutui.Resolved(ctx, gtx, resolved.label, a.fallbackWidget)
		}
		return layoutui.Resolved(ctx, gtx, resolved.icon, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return avatarIcon(gtx, ctx.ForegroundColor())
		}))
	})
}

func (a Widget) semanticLabel() string {
	if a.alt != "" {
		return a.alt
	}
	return a.fallbackText
}
