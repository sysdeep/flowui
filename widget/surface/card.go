package surface

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

// CardWidget groups related content on a raised semantic surface. The look comes
// from the active theme tokens.
type CardWidget struct {
	key         string
	children    []core.Widget
	variant     CardVariant
	style       style.Style
	padding     *unit.Dp
	gap         *unit.Dp
	visualCache *cardVisualCache
}

func NewCard(children ...core.Widget) CardWidget {
	return CardWidget{key: core.AutoKey("card"), children: children, visualCache: &cardVisualCache{}}
}

// Key sets the stable identity used by style transitions.
func (c CardWidget) Key(key string) CardWidget {
	c.key = key
	return c
}

func (c CardWidget) Variant(variant CardVariant) CardWidget {
	c.variant = variant
	return c
}

func (c CardWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	visual := c.visualWithCache(ctx)
	cardPrepareCardFieldAssociations(ctx, c.children...)
	content := layoutui.Box(
		layoutui.Column(cardNonNilWidgets(c.children)...).Gap(int(visual.Gap)),
	)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		c.style,
	)
	if len(resolved.Transitions) != 0 {
		key := core.ClaimDerivedKey(ctx, core.KindStyle, c.key, "style")
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, content)
}

// Style merges instance declarations into the card. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (c CardWidget) Style(values ...style.Style) CardWidget {
	c.style = style.Join(c.style, style.Join(values...))
	return c
}

// StyleWithMetrics sets the instance declaration together with explicit
// internal padding and gap.
func (c CardWidget) StyleWithMetrics(declaration style.Style, padding, gap unit.Dp) CardWidget {
	padding = max(padding, 0)
	gap = max(gap, 0)
	c.style = style.Join(c.style, declaration)
	c.padding = &padding
	c.gap = &gap
	return c
}

func cardPrepareCardFieldAssociations(ctx *core.Context, widgets ...core.Widget) {
	for _, widget := range widgets {
		switch widget := widget.(type) {
		case CardWidget:
			cardPrepareCardFieldAssociations(ctx, widget.children...)
		case *CardWidget:
			if widget != nil {
				cardPrepareCardFieldAssociations(ctx, widget.children...)
			}
		default:
			layoutui.PrepareFieldAssociations(ctx, widget)
		}
	}
}

func cardNonNilWidgets(widgets []core.Widget) []core.Widget {
	result := make([]core.Widget, 0, len(widgets))
	for _, widget := range widgets {
		if widget != nil {
			result = append(result, widget)
		}
	}
	return result
}
