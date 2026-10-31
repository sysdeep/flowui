package layout

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// StyleFunc computes an instance-style override from the current interaction
// state. It is evaluated before the normal style cascade.
type StyleFunc func(*Context, layout.Context, style.State) style.Style

// BoxWidget is a layout host whose visual and box-model properties come only
// from Style (including style scopes). Identity, interaction, and content
// alignment stay on the widget.
type BoxWidget struct {
	child       Widget
	key         string
	label       string
	onClick     func()
	disabled    bool
	interactive bool
	customStyle style.Style
	styleFunc   StyleFunc
	hasAlign    bool
	align       Align
}

// Box returns a host for child. Geometry comes only from Style.
func Box(child Widget) BoxWidget {
	return BoxWidget{child: child}
}

// Child returns the hosted widget.
func (b BoxWidget) Child() Widget {
	return b.child
}

// Key sets the stable identity used by an interactive or transitioning Box.
func (b BoxWidget) Key(key string) BoxWidget {
	b.key = key
	return b
}

// Label sets the accessible name of an interactive Box.
func (b BoxWidget) Label(label string) BoxWidget {
	b.label = label
	return b
}

func (b BoxWidget) Disabled(disabled bool) BoxWidget {
	b.disabled = disabled
	return b
}

func (b BoxWidget) OnClick(fn func()) BoxWidget {
	b.onClick = fn
	b.interactive = true
	return b
}

// Align sets how the child is placed inside the box. Alignment is layout
// policy, not Style.
func (b BoxWidget) Align(align Align) BoxWidget {
	b.align = align
	b.hasAlign = true
	return b
}

// Style merges declarations into the instance layer in order. Later properties
// override earlier ones while unspecified properties are preserved.
func (b BoxWidget) Style(values ...style.Style) BoxWidget {
	b.customStyle = style.Join(b.customStyle, style.Join(values...))
	return b
}

// ReplaceStyle replaces the box's complete instance style layer with the given
// declarations joined in order. This is useful when a reusable style should not
// inherit an earlier Box style.
func (b BoxWidget) ReplaceStyle(values ...style.Style) BoxWidget {
	b.customStyle = style.Join(values...)
	return b
}

// StyleFunc adds a state-aware instance-style override and makes the box
// interactive so hover, press, and focus state are available to fn.
func (b BoxWidget) StyleFunc(fn StyleFunc) BoxWidget {
	b.styleFunc = fn
	if fn != nil {
		b.interactive = true
	}
	return b
}

func (b BoxWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	if b.interactive {
		click := core.Begin(ctx, gtx, b.key, nil, !b.disabled, true, b.onClick)
		customStyle := b.customStyle
		if b.styleFunc != nil {
			customStyle = style.Join(customStyle, b.styleFunc(ctx, gtx, click.State))
		}
		resolved := computed.Resolve(
			ctx,
			gtx,
			click.Key,
			click.State,
			style.Style{},
			style.Style{},
			style.Style{},
			customStyle,
		)
		return b.layoutInteractiveResolved(ctx, gtx, resolved, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			return click.Layout(gtx, visual, b.label)
		})
	}
	prepareFieldAssociation(ctx, b.child)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		style.Style{},
		style.Style{},
		style.Style{},
		b.customStyle,
	)
	if len(resolved.Transitions) != 0 {
		key := ""
		if b.key != "" {
			key = core.ClaimKey(ctx, core.KindStyle, b.key)
		}
		resolved = computed.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return b.layoutResolved(ctx, gtx, resolved)
}

// Styled resolves an interactive component root through
// internal/computed, then renders it with Box semantics.
func Styled(ctx *Context, gtx layout.Context, key string, state style.State, value style.Style, child Widget) layout.Dimensions {
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		style.Style{},
		style.Style{},
		style.Style{},
		value,
	)
	return Resolved(ctx, gtx, resolved, child)
}

// Resolved renders a style that has already passed through the common
// cascade. Compound components use it to share Box semantics without resolving
// their default, variant, and instance layers twice.
func Resolved(ctx *Context, gtx layout.Context, resolved style.ResolvedStyle, child Widget) layout.Dimensions {
	prepareFieldAssociation(ctx, child)
	return (BoxWidget{child: child}).layoutResolved(ctx, gtx, resolved)
}

// InteractiveResolved keeps margin outside an interaction host while
// applying every other Box property to the host's visual and hit area.
func InteractiveResolved(
	ctx *Context,
	gtx layout.Context,
	resolved style.ResolvedStyle,
	child Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return (BoxWidget{child: child}).layoutInteractiveResolved(ctx, gtx, resolved, host)
}
