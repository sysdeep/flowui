package toggle

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

type ToggleButtonVariant = theme.ToggleButtonVariant

const (
	ToggleButtonDefault = theme.ToggleButtonDefault
	ToggleButtonGhost   = theme.ToggleButtonGhost
)

type ToggleButtonSize = theme.ToggleButtonSize

const (
	ToggleButtonMedium = theme.ToggleButtonMedium
	ToggleButtonSmall  = theme.ToggleButtonSmall
	ToggleButtonLarge  = theme.ToggleButtonLarge
)

type ToggleButtonWidget struct {
	override *theme.ToggleOverride
	key      string
	selected bool
	child    core.Widget
	onChange func(bool)
	variant  ToggleButtonVariant
	size     ToggleButtonSize
	disabled bool
	iconOnly bool
	label    string
}

// New creates a toggle with visible text and a matching accessible label. It
// starts unselected; Selected sets the controlled value.
func NewToggleButton(label string) ToggleButtonWidget {
	return ToggleButtonWidget{key: core.AutoKey(), child: text.New(label), label: label}
}

// Key sets an explicit identity for the toggle state.
func (b ToggleButtonWidget) Key(key string) ToggleButtonWidget {
	b.key = key
	return b
}

// Selected sets the controlled selection state.
func (b ToggleButtonWidget) Selected(selected bool) ToggleButtonWidget {
	b.selected = selected
	return b
}

// Content replaces the visible content, preserving the accessible label.
func (b ToggleButtonWidget) Content(value core.Widget) ToggleButtonWidget {
	b.child = value
	return b
}

func (b ToggleButtonWidget) OnChange(onChange func(bool)) ToggleButtonWidget {
	b.onChange = onChange
	return b
}

func (b ToggleButtonWidget) Variant(variant ToggleButtonVariant) ToggleButtonWidget {
	b.variant = variant
	return b
}

func (b ToggleButtonWidget) Size(size ToggleButtonSize) ToggleButtonWidget {
	b.size = size
	return b
}

func (b ToggleButtonWidget) Disabled(disabled bool) ToggleButtonWidget {
	b.disabled = disabled
	return b
}

func (b ToggleButtonWidget) IconOnly() ToggleButtonWidget {
	b.iconOnly = true
	return b
}

// Label overrides the accessible label without changing the visible content.
func (b ToggleButtonWidget) Label(label string) ToggleButtonWidget {
	b.label = label
	return b
}

func (b ToggleButtonWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return b.layoutWithOverride(ctx, gtx, b.override)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ToggleButtonWidget) Style(values ...style.Style) ToggleButtonWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ToggleOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
