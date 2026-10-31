package input

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/menu"
)

const stateSlotInputGroup = "input-group"

// GroupWidget combines an Input or TextArea with optional prefix and suffix content
// inside one field shell with the built-in appearance.
type GroupWidget struct {
	style              style.Style
	input              Widget
	textArea           TextAreaWidget
	multiline          bool
	prefix             core.Widget
	suffix             core.Widget
	prefixAction       bool
	suffixAction       bool
	focusOnActionPress bool
	variant            Variant
	disabled           bool
	invalid            bool
	fullWidth          bool
	width              *unit.Dp
}

func Group(input Widget) GroupWidget {
	return GroupWidget{
		input:     input,
		variant:   input.variant,
		disabled:  input.disabled,
		invalid:   input.invalid,
		fullWidth: input.fullWidth,
		width:     input.width,
	}
}

func GroupTextArea(textArea TextAreaWidget) GroupWidget {
	return GroupWidget{
		textArea:  textArea,
		multiline: true,
		variant:   textArea.variant,
		disabled:  textArea.disabled,
		invalid:   textArea.invalid,
		fullWidth: textArea.fullWidth,
		width:     textArea.width,
	}
}

func (g GroupWidget) Prefix(prefix core.Widget) GroupWidget {
	g.prefix = prefix
	g.prefixAction = false
	return g
}

func (g GroupWidget) Suffix(suffix core.Widget) GroupWidget {
	g.suffix = suffix
	g.suffixAction = false
	return g
}

// PrefixAction adds a compact interactive control before the editor. The
// default appearance supplies its slot padding; use Style with a PartPrefix
// declaration to change the inset for this group.
func (g GroupWidget) PrefixAction(action core.Widget) GroupWidget {
	g.prefix = action
	g.prefixAction = true
	return g
}

// SuffixAction adds a compact interactive control after the editor. The
// default appearance supplies its slot padding; use Style with a PartSuffix
// declaration to change the inset for this group.
func (g GroupWidget) SuffixAction(action core.Widget) GroupWidget {
	g.suffix = action
	g.suffixAction = true
	return g
}

// FocusOnActionPress controls whether pressing a prefix or suffix action also
// focuses the editor. It is disabled by default for action slots.
func (g GroupWidget) FocusOnActionPress(enabled bool) GroupWidget {
	g.focusOnActionPress = enabled
	return g
}

func (g GroupWidget) Variant(variant Variant) GroupWidget {
	g.variant = variant
	return g
}

func (g GroupWidget) Disabled(disabled bool) GroupWidget {
	g.disabled = disabled
	return g
}

func (g GroupWidget) Invalid(invalid bool) GroupWidget {
	g.invalid = invalid
	return g
}

func (g GroupWidget) FillWidth() GroupWidget {
	g.fullWidth = true
	return g
}

// Width sets a fixed width in dp. It overrides FillWidth when both are set.
func (g GroupWidget) Width(dp int) GroupWidget {
	width := unit.Dp(max(dp, 0))
	g.width = &width
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	applyFixedWidth(&gtx, g.width)
	var key string
	var editor *widget.Editor
	var enabled bool
	var state *inputGroupState
	var queueEdit func(EditAction)
	var readOnly, secret, hasCustomMenu, menuDisabled bool
	var customMenu menu.Widget

	if g.multiline {
		fullKey := core.FullKey(ctx, g.textArea.key)
		textAreaState := textAreaStateFor(ctx, fullKey)
		textAreaState.bind(g.textArea)
		currentValue := textAreaState.currentValue(g.textArea)
		gtx, key, editor, enabled = g.textArea.prepareEditor(ctx, gtx, textAreaState, currentValue, g.disabled)
		queueEdit = textAreaState.queueEdit
		readOnly = g.textArea.readOnly
		customMenu, hasCustomMenu = g.textArea.contextMenu, g.textArea.hasContextMenu
		menuDisabled = !g.textArea.hasContextMenu && g.textArea.defaultContextMenuDisabled
	} else {
		fullKey := core.FullKey(ctx, g.input.key)
		inputState := inputStateFor(ctx, fullKey)
		inputState.bind(g.input)
		currentValue := inputState.currentValue(g.input)
		gtx, key, editor, enabled = g.input.prepareEditor(ctx, gtx, inputState, currentValue, g.disabled)
		queueEdit = inputState.queueEdit
		readOnly = g.input.readOnly
		secret = g.input.inputType == Password
		customMenu, hasCustomMenu = g.input.contextMenu, g.input.hasContextMenu
		menuDisabled = !g.input.hasContextMenu && g.input.defaultContextMenuDisabled
	}
	disabled := !enabled
	state = inputGroupStateFor(ctx, key)
	var shouldFocus func(pointer.Event) bool
	if !g.focusOnActionPress && (g.prefixAction || g.suffixAction) {
		shouldFocus = state.shouldFocusPress
	}
	state.FieldInteraction.UpdateWithFocus(ctx, gtx, disabled, editor, shouldFocus)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered: state.Hovered, Focused: focused, FocusVisible: focusVisible,
		Disabled: disabled, ReadOnly: editor.ReadOnly, Invalid: g.invalid,
	}
	activeTheme := core.ActiveTheme(ctx)
	var actionDefaults groupActionDefaults
	if g.prefixAction || g.suffixAction {
		actionDefaults = groupActionDefaultsFor(ctx, key)
	}
	defaults := state.defaults(g, activeTheme, 0)
	resolved := resolveInputGroupStyle(ctx, gtx, key, styleState, defaults, g.style)
	_, lineHeight := textMetrics(resolved.content)
	if g.multiline {
		minHeight := inputGroupTextAreaHeight(gtx, resolved.content, lineHeight, g.textArea.resolvedRows())
		if minHeight > 0 {
			defaults = state.defaults(g, activeTheme, minHeight)
			resolved = resolveInputGroupStyle(ctx, gtx, key, styleState, defaults, g.style)
			_, lineHeight = textMetrics(resolved.content)
		}
	}
	resolved.lineHeight = lineHeight
	var editorLayout layout.Widget
	if g.multiline {
		editorLayout = g.textArea.editorLayoutWithStyle(ctx, key, enabled, editor, activeTheme, resolved.content, resolved.placeholder, resolved.selection)
	} else {
		editorLayout = g.input.editorLayoutWithStyle(ctx, key, enabled, editor, activeTheme, resolved.content, resolved.placeholder, resolved.selection)
	}
	if g.focusOnActionPress {
		g.prefix = inputGroupActionWithFocus(g.prefix, editor)
		g.suffix = inputGroupActionWithFocus(g.suffix, editor)
	}
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return g.layoutFrame(ctx, gtx, state, resolved, enabled, editorLayout, actionDefaults)
	})
	if !enabled || menuDisabled {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, readOnly, secret)
	return EditContextMenu(ctx, gtx, key, frame, editor, snapshot, customMenu, hasCustomMenu, queueEdit)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (g GroupWidget) Style(values ...style.Style) GroupWidget {
	g.style = style.Join(g.style, style.Join(values...))
	return g
}
