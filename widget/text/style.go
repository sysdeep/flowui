package text

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type textVisual struct {
	Defaults  style.Style
	Selection color.NRGBA
}

type textVisualKey struct {
	Foreground color.NRGBA
}

type textVisualCache = computed.VisualCache[textVisualKey, textVisual]

func (t Widget) visualWithCache(ctx *core.Context) textVisual {
	active := core.ActiveTheme(ctx)
	foreground := ctx.ForegroundColor()
	if t.visualCache == nil {
		return resolveTextVisual(active, foreground)
	}
	return t.visualCache.Resolve(active, textVisualKey{Foreground: foreground}, func() textVisual {
		return resolveTextVisual(active, foreground)
	})
}

func resolveTextVisual(active *theme.Theme, foreground color.NRGBA) textVisual {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return textVisual{
		Defaults: style.Style{}.
			FontSize(active.Typography.BodySize).
			Typeface(active.Typography.Typeface).
			TextColor(style.Solid(foreground)),
		Selection: active.Semantic().Selection,
	}
}

func drawTextUnderline(gtx layout.Context, dims layout.Dimensions, col color.NRGBA) {
	if dims.Size.X <= 0 || dims.Size.Y <= 0 || col.A == 0 {
		return
	}
	thickness := max(gtx.Dp(1), 1)
	top := dims.Baseline + thickness
	if top+thickness > dims.Size.Y {
		top = dims.Size.Y - thickness
	}
	if top < 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.Rect{
		Min: image.Point{Y: top},
		Max: image.Point{X: dims.Size.X, Y: top + thickness},
	}.Op())
}
