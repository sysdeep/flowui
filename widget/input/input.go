package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/menu"
)

type Widget struct {
	style                      style.Style
	key                        string
	value                      string
	hasValue                   bool
	defaultValue               string
	hasDefault                 bool
	hint                       string
	onChange                   func(string)
	onSubmit                   func(string)
	variant                    Variant
	disabled                   bool
	invalid                    bool
	warning                    bool
	size                       Size
	fullWidth                  bool
	width                      *unit.Dp
	inputType                  Type
	readOnly                   bool
	maxLength                  int
	label                      string
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

const stateSlotInput = "input"

// New creates a text input holding value.
func New(value string) Widget {
	return Widget{
		key:      core.AutoKey("input"),
		value:    value,
		hasValue: true,
	}
}

// Key sets an explicit identity for the input state.
func (i Widget) Key(key string) Widget {
	i.key = key
	return i
}

// Uncontrolled changes the initial value into an internal state seed.
// It is useful when the initial value is empty and should not imply control.
func (i Widget) Uncontrolled() Widget {
	i.defaultValue = i.value
	i.hasDefault = true
	i.hasValue = false
	return i
}

func (i Widget) Hint(hint string) Widget {
	i.hint = hint
	return i
}

func (i Widget) Placeholder(placeholder string) Widget {
	i.hint = placeholder
	return i
}

func (i Widget) OnChange(fn func(string)) Widget {
	i.onChange = fn
	return i
}

func (i Widget) Value(value string) Widget {
	i.value = value
	i.hasValue = true
	return i
}

func (i Widget) DefaultValue(value string) Widget {
	i.defaultValue = value
	i.hasDefault = true
	i.hasValue = false
	return i
}

func (i Widget) OnSubmit(fn func(string)) Widget {
	i.onSubmit = fn
	return i
}

func (i Widget) Disabled(disabled bool) Widget {
	i.disabled = disabled
	return i
}

func (i Widget) Invalid(invalid bool) Widget {
	i.invalid = invalid
	return i
}

// Warning marks the field with the warning state: an amber outline instead of
// the danger one. Pair with the theme's Warning palette.
func (i Widget) Warning(warning bool) Widget {
	i.warning = warning
	return i
}

// Size selects the built-in geometry tier; Medium is the default.
func (i Widget) Size(size Size) Widget {
	i.size = size
	return i
}

func (i Widget) Variant(variant Variant) Widget {
	i.variant = variant
	return i
}

func (i Widget) FillWidth() Widget {
	i.fullWidth = true
	return i
}

// Width sets a fixed width in dp. It overrides FillWidth when both are set.
func (i Widget) Width(dp int) Widget {
	width := unit.Dp(max(dp, 0))
	i.width = &width
	return i
}

func (i Widget) Type(inputType Type) Widget {
	i.inputType = inputType
	return i
}

func (i Widget) ReadOnly(readOnly bool) Widget {
	i.readOnly = readOnly
	return i
}

func (i Widget) MaxLength(maxLength int) Widget {
	i.maxLength = max(maxLength, 0)
	return i
}

func (i Widget) Label(label string) Widget {
	i.label = label
	return i
}

// ContextMenu replaces the built-in cut/copy/paste menu shown on right-click.
// The custom menu is responsible for its own labels and actions.
func (i Widget) ContextMenu(content menu.Widget) Widget {
	i.contextMenu = content
	i.hasContextMenu = true
	return i
}

// DefaultContextMenu enables or disables the built-in edit menu shown on
// right-click. It is enabled by default; a custom menu set through ContextMenu
// takes precedence.
func (i Widget) DefaultContextMenu(enabled bool) Widget {
	i.defaultContextMenuDisabled = !enabled
	return i
}

func (i Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	applyFixedWidth(&gtx, i.width)
	state := inputStateFor(ctx, core.FullKey(ctx, i.key))

	state.bind(i)
	currentValue := state.currentValue(i)

	gtx, key, editor, enabled := i.prepareEditor(ctx, gtx, state, currentValue, i.disabled)
	disabled := !enabled
	state.FieldInteraction.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     i.readOnly,
		Invalid:      i.invalid,
		Warning:      i.warning,
	}
	activeTheme := core.ActiveTheme(ctx)
	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: state.defaults(i, activeTheme),
		custom:   i.style,
	})
	editorLayout := i.editorLayoutWithStyle(ctx, key, enabled, editor, activeTheme, root, placeholder, selection)
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return i.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
	})
	if !enabled || (!i.hasContextMenu && i.defaultContextMenuDisabled) {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, i.readOnly, i.inputType == Password)
	return EditContextMenu(ctx, gtx, i.key, frame, editor, snapshot, i.contextMenu, i.hasContextMenu, state.queueEdit)
}

func (i Widget) prepareEditor(ctx *core.Context, gtx layout.Context, state *inputState, currentValue string, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	fieldKey, editor := core.InputEditor(ctx, i.key)
	enabled := gtx.Enabled() && !disabled
	mask, hint, filter := inputTypeConfig(i.inputType)
	behavior := base.TextEditor{SingleLine: true, ReadOnly: i.readOnly, MaxLength: i.maxLength, Mask: mask, InputHint: hint, Filter: filter}
	if !enabled {
		gtx = gtx.Disabled()
	}
	behavior.Update(ctx, gtx, base.TextEditorRequest{
		FieldKey: fieldKey, Editor: editor, Value: currentValue,
		BeforeUpdate: func(gtx layout.Context) { state.takePendingEdit(gtx, editor, i.readOnly) },
		OnChange:     func(value string) { state.requestValue(i, value) },
		OnSubmit:     i.onSubmit,
	})
	return gtx, fieldKey, editor, enabled
}

func (i Widget) editorLayoutWithStyle(ctx *core.Context, key string, enabled bool, editor *widget.Editor, active *theme.Theme, content, placeholder, selection style.ResolvedStyle) layout.Widget {
	return i.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, i.hint, active, content, placeholder, selection))
}

func inputTypeConfig(inputType Type) (rune, key.InputHint, string) {
	switch inputType {
	case Email:
		return 0, key.HintEmail, ""
	case Number:
		return 0, key.HintNumeric, "0123456789+-.eE"
	case Password:
		return '\u2022', key.HintPassword, ""
	default:
		return 0, key.HintText, ""
	}
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (i Widget) Style(values ...style.Style) Widget {
	i.style = style.Join(i.style, style.Join(values...))
	return i
}
