package text

import (
	"gioui.org/layout"
	"gioui.org/text"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// DescriptionWidget provides supporting text for a form field or control.
type DescriptionWidget struct {
	override    *theme.DescriptionOverride
	text        string
	forKey      string
	disabled    bool
	visualCache *descriptionDescriptionVisualCache
}

// New creates supporting text for a form field or control.
func NewDescription(text string) DescriptionWidget {
	return DescriptionWidget{text: text, visualCache: &descriptionDescriptionVisualCache{}}
}

// For associates the description with a keyed control.
func (d DescriptionWidget) For(key string) DescriptionWidget {
	d.forKey = key
	return d
}

func (d DescriptionWidget) Disabled(disabled bool) DescriptionWidget {
	d.disabled = disabled
	return d
}

func (d DescriptionWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return d.layoutWithOverride(ctx, gtx, d.override)
}

// LayoutDescriptionWithResolvedText renders a description whose text declaration has
// already been resolved by its compound parent. The parent remains
// responsible for the surrounding box style; this helper preserves the field
// association without resolving the description provider a second time.
func LayoutDescriptionWithResolvedText(ctx *core.Context, gtx layout.Context, d DescriptionWidget, resolved *style.TextStyle) layout.Dimensions {
	if d.forKey != "" {
		d.registerFieldAssociation(ctx)
	}
	return WithResolvedText(
		New(d.text).Wrap(text.WrapHeuristically),
		resolved,
	).Layout(ctx, gtx)
}

func (d DescriptionWidget) layoutWithOverride(
	ctx *core.Context,
	gtx layout.Context,
	override *theme.DescriptionOverride,
) layout.Dimensions {
	if d.forKey != "" {
		d.registerFieldAssociation(ctx)
	}
	state := style.State{Disabled: d.disabled || !gtx.Enabled()}
	visual := d.visualWithCache(ctx, override)
	resolved := computed.ResolveStatic(
		ctx,
		state,
		visual.Defaults,
		style.Style{},
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) != 0 {
		var key string
		if d.forKey != "" {
			key = core.ClaimDerivedKey(ctx, core.KindStyle, d.forKey, "description")
		} else {
			key = core.ClaimKey(ctx, core.KindStyle, "description")
		}
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return New(d.text).
			Wrap(text.WrapHeuristically).
			Style(style.TextDeclaration(resolved.Text)).
			Layout(ctx, gtx)
	}))
}

func (d DescriptionWidget) registerFieldAssociation(ctx *core.Context) {
	core.RegisterFieldDescription(ctx, core.FullKey(ctx, d.forKey), d.text)
}

// PrepareFieldAssociation registers the description before a container
// chooses its child layout order.
func (d DescriptionWidget) PrepareFieldAssociation(ctx *core.Context) bool {
	if d.forKey == "" {
		return false
	}
	core.PrepareFieldDescription(ctx, core.FullKey(ctx, d.forKey), d.text)
	return true
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w DescriptionWidget) Style(values ...style.Style) DescriptionWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.DescriptionOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
