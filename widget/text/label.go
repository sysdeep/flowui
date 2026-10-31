package text

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// LabelWidget identifies and describes a form field.
type LabelWidget struct {
	style       style.Style
	text        string
	forKey      string
	required    bool
	disabled    bool
	invalid     bool
	visualCache *labelVisualCache
}

// New creates a form label.
func NewLabel(text string) LabelWidget {
	return LabelWidget{text: text, visualCache: &labelVisualCache{}}
}

// For associates the label with a keyed form control.
func (l LabelWidget) For(key string) LabelWidget {
	l.forKey = key
	return l
}

func (l LabelWidget) Required(required bool) LabelWidget {
	l.required = required
	return l
}

func (l LabelWidget) Disabled(disabled bool) LabelWidget {
	l.disabled = disabled
	return l
}

func (l LabelWidget) Invalid(invalid bool) LabelWidget {
	l.invalid = invalid
	return l
}

// Layout lays out the label and preserves its field focus and accessibility
// behavior while obtaining all visual declarations from the active recipe.
func (l LabelWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return l.layoutWithStyle(ctx, gtx)
}

func (l LabelWidget) resolveLayoutStyle(
	ctx *core.Context,
	gtx layout.Context,
	state style.State,
	visual labelVisual,
) labelResolvedStyle {
	resolved := l.resolveStyleStatic(ctx, state, visual)
	if len(resolved.root.Transitions) == 0 {
		return resolved
	}
	resolved.root = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "label"), resolved.root)
	return resolved
}

func (l LabelWidget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	visual labelVisual,
) labelResolvedStyle {
	resolved := l.resolveStyleStatic(ctx, state, visual)
	if len(resolved.root.Transitions) == 0 {
		return resolved
	}
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	return resolved
}

func (l LabelWidget) resolveStyleStatic(
	ctx *core.Context,
	state style.State,
	visual labelVisual,
) labelResolvedStyle {
	return labelResolvedStyle{
		root: computed.ResolveStatic(
			ctx,
			state,
			visual.Defaults,
			style.Style{},
			style.Style{},
			l.style,
		),
		required: computed.ResolveStatic(
			ctx,
			state,
			visual.Required,
			style.Style{},
			style.Style{},
			style.Style{},
		),
	}
}

func (l LabelWidget) visualWithCache(ctx *core.Context) labelVisual {
	active := core.ActiveTheme(ctx)
	foreground := ctx.ForegroundColor()
	if l.visualCache == nil {
		return labelResolveLabelVisual(active, foreground)
	}
	key := labelVisualKey{Foreground: foreground}
	return l.visualCache.Resolve(active, key, func() labelVisual {
		return labelResolveLabelVisual(active, foreground)
	})
}

func (l LabelWidget) registerFieldAssociation(ctx *core.Context) string {
	fieldKey := core.FullKey(ctx, l.forKey)
	core.RegisterFieldLabel(ctx, fieldKey, l.text)
	return fieldKey
}

// PrepareFieldAssociation registers the label before a container chooses its
// child layout order.
func (l LabelWidget) PrepareFieldAssociation(ctx *core.Context) bool {
	if l.forKey == "" {
		return false
	}
	core.PrepareFieldLabel(ctx, core.FullKey(ctx, l.forKey), l.text)
	return true
}

// Style merges instance declarations into the widget. The declaration
// is resolved as the last cascade layer at layout time, so Style can sit
// anywhere in the chain.
func (w LabelWidget) Style(values ...style.Style) LabelWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
