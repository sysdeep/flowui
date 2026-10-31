package surface

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// Widget provides a semantic content surface around a child. Its visual
// declaration uses the active theme tokens.
type Widget struct {
	style       style.Style
	child       core.Widget
	variant     Variant
	visualCache *surfaceVisualCache
}

func New(child core.Widget) Widget {
	return Widget{child: child, visualCache: &surfaceVisualCache{}}
}

func (s Widget) Variant(variant Variant) Widget {
	s.variant = variant
	return s
}

// Layout resolves the default surface and scopes its background and
// foreground around the child while preserving the child's behavior.
func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return s.layoutWithStyle(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (s Widget) Style(values ...style.Style) Widget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}
