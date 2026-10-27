package surface

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// CardVariant selects the semantic surface prominence of a card.
type CardVariant = theme.CardVariant

const (
	CardDefault     = theme.CardDefault
	CardSecondary   = theme.CardSecondary
	CardTertiary    = theme.CardTertiary
	CardTransparent = theme.CardTransparent
)

// CardWidget groups related content on a raised semantic surface. The look comes
// from the active theme, not from this package.
type CardWidget struct {
	children    []core.Widget
	variant     CardVariant
	visualCache *cardCardVisualCache
	override    *theme.CardOverride
}

func NewCard(children ...core.Widget) CardWidget {
	return CardWidget{children: children, visualCache: &cardCardVisualCache{}}
}

func (c CardWidget) Variant(variant CardVariant) CardWidget {
	c.variant = variant
	return c
}

func (c CardWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithOverride(ctx, gtx, c.override)
}

// Style merges instance declarations into the card. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (c CardWidget) Style(values ...style.Style) CardWidget {
	declaration := style.Join(values...)
	if c.override == nil {
		c.override = &theme.CardOverride{}
	}
	c.override.Declaration = style.Join(c.override.Declaration, declaration)
	return c
}

// StyleWithMetrics sets the instance declaration together with explicit
// internal padding and gap.
func (c CardWidget) StyleWithMetrics(declaration style.Style, padding, gap unit.Dp) CardWidget {
	padding = max(padding, 0)
	gap = max(gap, 0)
	if c.override == nil {
		c.override = &theme.CardOverride{}
	}
	c.override.Declaration = style.Join(c.override.Declaration, declaration)
	c.override.Padding = &padding
	c.override.Gap = &gap
	return c
}

func (c CardWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CardOverride) layout.Dimensions {
	visual := c.visualWithCache(ctx, override)
	cardPrepareCardFieldAssociations(ctx, c.children...)
	content := layoutui.Box(
		layoutui.Column(cardNonNilWidgets(c.children)...).Gap(int(visual.Metrics.Gap)),
	)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) != 0 {
		resolved = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "card"), resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, content)
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
