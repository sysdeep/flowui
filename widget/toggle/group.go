package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type SwitchGroupWidget struct {
	style      style.Style
	children   []core.Widget
	horizontal bool
}

func NewSwitchGroup(children ...core.Widget) SwitchGroupWidget {
	return SwitchGroupWidget{children: children}
}

func (g SwitchGroupWidget) Horizontal() SwitchGroupWidget {
	g.horizontal = true
	return g
}

func (g SwitchGroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		children := make([]layout.Widget, 0, len(g.children))
		for _, child := range g.children {
			children = append(children, func(gtx layout.Context) layout.Dimensions {
				return child.Layout(ctx, gtx)
			})
		}
		return layoutui.Items(ctx, gtx, g.horizontal, gtx.Dp(14), gtx.Dp(10), children)
	})).Style(g.style).Layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w SwitchGroupWidget) Style(values ...style.Style) SwitchGroupWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
