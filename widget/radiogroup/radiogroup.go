package radiogroup

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type RadioItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
	Invalid     bool
}

type Widget struct {
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []RadioItem
	onChange           func(string)
	variant            Variant
	disabled           bool
	invalid            bool
	horizontal         bool
	style              style.Style
}

// New creates a radio group. It starts uncontrolled with nothing selected;
// SelectedKey sets the controlled value and DefaultSelectedKey the initial one.
func New(items []RadioItem) Widget {
	return Widget{
		key:   core.AutoKey("radio-group"),
		items: items,
	}
}

// Key sets an explicit identity for the group state.
func (r Widget) Key(key string) Widget {
	r.key = key
	return r
}

func (r Widget) OnChange(fn func(string)) Widget {
	r.onChange = fn
	return r
}

func (r Widget) SelectedKey(key string) Widget {
	r.selectedKey = key
	r.hasSelectedKey = true
	return r
}

func (r Widget) DefaultSelectedKey(key string) Widget {
	r.defaultSelectedKey = key
	r.hasDefaultSelected = true
	return r
}

func (r Widget) Disabled(disabled bool) Widget {
	r.disabled = disabled
	return r
}

func (r Widget) Invalid(invalid bool) Widget {
	r.invalid = invalid
	return r
}

func (r Widget) Variant(variant Variant) Widget {
	r.variant = variant
	return r
}

func (r Widget) Horizontal() Widget {
	r.horizontal = true
	return r
}

func (r Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return r.layoutWithStyle(ctx, gtx)
}

func (r Widget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	state := radioGroupStateFor(ctx, r.key)
	state.beginFrame()
	defer state.endFrame()
	state.checkItems(r.items)

	// Bind disclosure state
	state.bind(r)
	selectedKey := state.currentSelectedKey(r)

	disabled := r.disabled || !gtx.Enabled()
	if disabled {
		gtx = gtx.Disabled()
	}
	state.updateKeys(ctx, gtx, r.items, selectedKey, func(key string) {
		if key != selectedKey {
			selectedKey = state.requestSelectedKey(r, key)
		}
	})

	defaults := state.defaults(ctx, r.variant)
	children := make([]layout.Widget, 0, len(r.items))
	for _, item := range r.items {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return r.layoutItem(ctx, gtx, state, item, selectedKey, defaults)
		})
	}

	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range r.items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		itemFocused := gtx.Focused(&itemState.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || core.FocusVisible(ctx, &itemState.clickable, itemFocused)
	}
	return layoutui.Styled(ctx, gtx, core.FullKey(ctx, r.key), style.State{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		Selected:     r.selectedKey != "",
		Checked:      r.selectedKey != "",
		Invalid:      r.invalid,
	}, r.style, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Items(ctx, gtx, r.horizontal, gtx.Dp(defaults.Metrics.HorizontalGap), gtx.Dp(defaults.Metrics.VerticalGap), children)
	}))
}

// Style merges instance declarations into the built-in radio appearance. The
// declaration is resolved as the last cascade layer at layout time, so Style
// can sit anywhere in the chain.
func (r Widget) Style(values ...style.Style) Widget {
	r.style = style.Join(r.style, style.Join(values...))
	return r
}
