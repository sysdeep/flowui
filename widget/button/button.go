package button

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	widgetbase "github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
)

type Widget struct {
	key         string
	child       core.Widget
	label       string
	onClick     func()
	variant     Variant
	variantSet  bool
	size        Size
	sizeSet     bool
	disabled    bool
	disabledSet bool
	loading     bool
	fullWidth   bool
	width       *unit.Dp
	iconOnly    bool
	group       buttonGroupItemStyle
	style       style.Style
}

// New creates a button with visible text and a matching accessible label.
// Use Content to replace the text with an icon or composite widget.
// Its identity is derived from the enclosing scope; call
// Key when the same position may hold different buttons.
func New(label string) Widget {
	return Widget{key: core.AutoKey("button"), child: text.New(label), label: label}
}

// Key sets an explicit identity for the button state.
func (b Widget) Key(key string) Widget {
	b.key = key
	return b
}

// Content replaces the visible content, preserving the accessible label.
func (b Widget) Content(value core.Widget) Widget {
	b.child = value
	return b
}

// Style merges instance declarations into the built-in button appearance.
func (b Widget) Style(values ...style.Style) Widget {
	b.style = style.Join(b.style, style.Join(values...))
	return b
}

func (b Widget) OnClick(fn func()) Widget {
	b.onClick = fn
	return b
}

// Label overrides the accessible label without changing the visible content.
func (b Widget) Label(value string) Widget {
	b.label = value
	return b
}

func (b Widget) Disabled(disabled bool) Widget {
	b.disabled = disabled
	b.disabledSet = true
	return b
}

func (b Widget) Loading(loading bool) Widget {
	b.loading = loading
	return b
}

func (b Widget) Variant(variant Variant) Widget {
	b.variant = variant
	b.variantSet = true
	return b
}

// Primary selects the primary action appearance.
func (b Widget) Primary() Widget {
	return b.Variant(Primary)
}

func (b Widget) Size(size Size) Widget {
	b.size = size
	b.sizeSet = true
	return b
}

func (b Widget) FillWidth() Widget {
	b.fullWidth = true
	return b
}

// Width sets a fixed width in dp. It overrides FillWidth when both are set.
func (b Widget) Width(dp int) Widget {
	width := unit.Dp(max(dp, 0))
	b.width = &width
	return b
}

func (b Widget) IconOnly() Widget {
	b.iconOnly = true
	return b
}

// DisabledState reports the resolved instance disabled flag to composite
// widgets inside the internal component tree.
func (b Widget) DisabledState() bool {
	return b.disabled
}

func (b Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return widgetbase.Button(b.key).
		Label(b.label).
		Disabled(b.disabled).
		Loading(b.loading).
		OnClick(b.onClick).
		Content(func(state widgetbase.ButtonState) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				dims, _ := b.layoutState(ctx, gtx, state.Key, styleState(state), nil)
				return dims
			})
		}).
		Layout(ctx, gtx)
}

// Visual draws only the button appearance for an existing behavior state.
// It neither consumes events nor registers a hit target or button semantics;
// the owning base must provide those around the returned widget.
func (b Widget) Visual(state widgetbase.ButtonState) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		dims, _ := b.layoutState(ctx, gtx, state.Key, styleState(state), nil)
		return dims
	})
}

// LayoutWithClickable renders a button with caller-owned interaction state.
func LayoutWithClickable(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return b.layoutWithClickable(ctx, gtx, clickable, true)
}

// LayoutWithClickableNoEvents renders a button with caller-owned state and event handling.
func LayoutWithClickableNoEvents(b Widget, ctx *core.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return b.layoutWithClickable(ctx, gtx, clickable, false)
}

func (b Widget) layoutWithClickable(ctx *core.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool) layout.Dimensions {
	return widgetbase.Button(b.key).
		Label(b.label).
		Disabled(b.disabled).
		Loading(b.loading).
		OnClick(b.onClick).
		Clickable(clickable).
		Events(handleEvents).
		Content(func(state widgetbase.ButtonState) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				dims, _ := b.layoutState(ctx, gtx, state.Key, styleState(state), nil)
				return dims
			})
		}).
		Layout(ctx, gtx)
}

// layoutState draws a button for an interaction state already owned by a base.
func (b Widget) layoutState(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	styleState style.State,
	prepared *buttonPreparedContent,
) (layout.Dimensions, buttonStyle) {
	animGtx := gtx
	layoutGtx := gtx
	if styleState.Disabled {
		layoutGtx = layoutGtx.Disabled()
	}
	var resolvedStyle buttonStyle
	if prepared != nil {
		resolvedStyle = b.resolvePreparedStyleWithDeclarations(ctx, animGtx, key, styleState, prepared.declarations)
	} else {
		resolvedStyle = b.resolveStyle(ctx, animGtx, key, styleState)
	}

	if b.width != nil {
		width := min(max(layoutGtx.Dp(*b.width), layoutGtx.Constraints.Min.X), layoutGtx.Constraints.Max.X)
		layoutGtx.Constraints.Min.X = width
		layoutGtx.Constraints.Max.X = width
	} else if b.fullWidth {
		layoutGtx.Constraints.Min.X = layoutGtx.Constraints.Max.X
	}

	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		motion := resolvedStyle.defaults.motion
		if b.loading && theme.MotionEnabled(motion) {
			animGtx.Execute(op.InvalidateCmd{})
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if prepared != nil {
				prepared.call.Add(gtx.Ops)
				return prepared.dims
			}
			return b.layoutContent(ctx, gtx, resolvedStyle)
		})
	})

	dims := flowlayout.InteractiveResolved(ctx, layoutGtx, resolvedStyle.root, content, nil)
	return dims, resolvedStyle
}
