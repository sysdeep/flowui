package statusbar

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

// Widget presents compact application state at the bottom of a window.
type Widget struct {
	key         string
	style       style.Style
	left        core.Widget
	right       core.Widget
	visualCache *statusBarVisualCache
}

func New(left, right core.Widget) Widget {
	return Widget{key: core.AutoKey("status-bar"), left: left, right: right, visualCache: &statusBarVisualCache{}}
}

// Key sets the stable identity used by style transitions.
func (w Widget) Key(key string) Widget {
	w.key = key
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := w.visualWithCache(ctx)
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(visual.Metrics.Height)))
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	children := make([]core.Widget, 0, 3)
	if w.left != nil {
		children = append(children, w.left)
	}
	children = append(children, layoutui.Expanded(layoutui.Spacer(0, 0)))
	if w.right != nil {
		children = append(children, w.right)
	}
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.TrackedDirection(ctx, gtx, layout.Center, func(gtx layout.Context) layout.Dimensions {
			return layoutui.TrackedFlex(ctx, gtx, layout.Horizontal, visual.Metrics.Gap, layout.Middle, children...)
		})
	})
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		w.style,
	)
	if len(resolved.Transitions) != 0 {
		key := core.ClaimDerivedKey(ctx, core.KindStyle, w.key, "style")
		resolved = computed.ApplyTransitions(ctx, rootGtx, key, resolved)
	}
	return layoutui.Resolved(ctx, rootGtx, resolved, content)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
