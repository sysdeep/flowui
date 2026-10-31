package button

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
)

type GroupOrientation uint8

const (
	GroupHorizontal GroupOrientation = iota
	GroupVertical
)

type buttonGroupPosition uint8

const (
	buttonGroupSingle buttonGroupPosition = iota
	buttonGroupStart
	buttonGroupMiddle
	buttonGroupEnd
)

type buttonGroupItemStyle struct {
	grouped     bool
	orientation GroupOrientation
	position    buttonGroupPosition
}

type GroupWidget struct {
	style       style.Style
	buttons     []Widget
	orientation GroupOrientation
	variant     Variant
	size        Size
	disabled    bool
	fullWidth   bool
	separators  bool
}

func Group(buttons ...Widget) GroupWidget {
	return GroupWidget{buttons: append([]Widget(nil), buttons...)}
}

// GroupStart applies the leading corner treatment used by Group to a
// button rendered by another composite widget.
func GroupStart(value Widget) Widget {
	value.group = buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupStart}
	return value
}

// GroupEnd applies the trailing corner treatment used by Group to a
// button rendered by another composite widget.
func GroupEnd(value Widget) Widget {
	value.group = buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupEnd}
	return value
}

func (g GroupWidget) Orientation(value GroupOrientation) GroupWidget {
	g.orientation = value
	return g
}

func (g GroupWidget) Variant(value Variant) GroupWidget {
	g.variant = value
	return g
}

func (g GroupWidget) Size(value Size) GroupWidget {
	g.size = value
	return g
}

func (g GroupWidget) Disabled(value bool) GroupWidget {
	g.disabled = value
	return g
}

func (g GroupWidget) FillWidth() GroupWidget {
	g.fullWidth = true
	return g
}

func (g GroupWidget) Separators(visible bool) GroupWidget {
	g.separators = visible
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if len(g.buttons) == 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	var items []buttonGroupItemLayout
	if g.separators {
		items = make([]buttonGroupItemLayout, len(g.buttons))
	}
	children := make([]layout.FlexChild, len(g.buttons))
	axis := layout.Horizontal
	if g.orientation == GroupVertical {
		axis = layout.Vertical
	}
	if g.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	// Only the vertical intrinsic-width pass needs prepared values across its
	// measurement and layout passes. Keep them local to this layout call so the
	// behavior Widget does not retain frame-specific visual operations.
	if axis == layout.Vertical && !g.fullWidth {
		type preparedButton struct {
			value   Widget
			content buttonPreparedContent
		}
		buttons := make([]preparedButton, len(g.buttons))
		verticalWidth := 0
		for index := range g.buttons {
			value := g.prepareButton(g.buttons[index], index, len(g.buttons))
			buttons[index] = preparedButton{value: value, content: value.prepareContent(ctx, gtx)}
			verticalWidth = max(verticalWidth, buttons[index].content.width)
		}
		verticalWidth = min(verticalWidth, gtx.Constraints.Max.X)
		for index := range buttons {
			preparedContent := &buttons[index].content
			layoutButton := func(gtx layout.Context) layout.Dimensions {
				if verticalWidth > 0 {
					gtx.Constraints.Min.X = verticalWidth
					gtx.Constraints.Max.X = verticalWidth
				}
				dims, resolved := g.layoutButton(ctx, gtx, buttons[index].value, preparedContent)
				if g.separators {
					items[index] = buttonGroupItemLayout{
						dims:       dims,
						foreground: buttonForeground(resolved.content),
					}
				}
				return dims
			}
			children[index] = layout.Rigid(layoutButton)
		}

		dims := layout.Flex{Axis: axis, Alignment: layout.Middle}.Layout(gtx, children...)
		root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		if g.separators {
			g.drawSeparators(gtx, dims.Size, items)
		}
		root.Pop()
		return dims
	}

	for index := range g.buttons {
		layoutButton := func(gtx layout.Context) layout.Dimensions {
			button := g.prepareButton(g.buttons[index], index, len(g.buttons))
			dims, resolved := g.layoutButton(ctx, gtx, button, nil)
			if g.separators {
				items[index] = buttonGroupItemLayout{
					dims:       dims,
					foreground: buttonForeground(resolved.content),
				}
			}
			return dims
		}
		if g.fullWidth && axis == layout.Horizontal {
			children[index] = layout.Flexed(1, layoutButton)
		} else {
			children[index] = layout.Rigid(layoutButton)
		}
	}

	dims := layout.Flex{Axis: axis, Alignment: layout.Middle}.Layout(gtx, children...)
	root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	if g.separators {
		g.drawSeparators(gtx, dims.Size, items)
	}
	root.Pop()
	return dims
}

func (g GroupWidget) prepareButton(button Widget, index, count int) Widget {
	if !button.variantSet {
		button.variant = g.variant
	}
	if !button.sizeSet {
		button.size = g.size
	}
	if !button.disabledSet {
		button.disabled = g.disabled
	}
	if g.fullWidth {
		button.fullWidth = true
	}
	position := buttonGroupMiddle
	if count == 1 {
		position = buttonGroupSingle
	} else if index == 0 {
		position = buttonGroupStart
	} else if index == count-1 {
		position = buttonGroupEnd
	}
	button.group = buttonGroupItemStyle{grouped: true, orientation: g.orientation, position: position}
	if !g.style.IsEmpty() {
		button.style = g.style
	}
	return button
}

type buttonGroupItemLayout struct {
	dims       layout.Dimensions
	foreground color.NRGBA
}

func (g GroupWidget) drawSeparators(gtx layout.Context, size image.Point, items []buttonGroupItemLayout) {
	if len(items) == 0 {
		return
	}
	main := 0
	for index := 1; index < len(items); index++ {
		previous := g.orientationSize(items[index-1].dims.Size)
		main += previous.X
		drawGroupSeparator(gtx, size, main, g.orientation, items[index].foreground)
	}
}

func (g GroupWidget) orientationSize(value image.Point) image.Point {
	if g.orientation == GroupVertical {
		return image.Pt(value.Y, value.X)
	}
	return value
}

// Style applies instance declarations to every button in the group.
func (g GroupWidget) Style(values ...style.Style) GroupWidget {
	g.style = style.Join(g.style, style.Join(values...))
	return g
}

func (g GroupWidget) layoutButton(ctx *core.Context, gtx layout.Context, button Widget, prepared *buttonPreparedContent) (layout.Dimensions, buttonStyle) {
	var dims layout.Dimensions
	var resolved buttonStyle
	base.Button(button.key).
		Label(button.label).
		Disabled(button.disabled).
		Loading(button.loading).
		OnClick(button.onClick).
		Content(func(state base.ButtonState) core.Widget {
			return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				dims, resolved = button.layoutState(ctx, gtx, state.Key, styleState(state), prepared)
				return dims
			})
		}).
		Layout(ctx, gtx)
	return dims, resolved
}
