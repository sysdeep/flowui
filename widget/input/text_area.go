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

const defaultTextAreaRows = 3

type TextAreaWidget struct {
	style                      style.Style
	key                        string
	value                      string
	hasValue                   bool
	defaultValue               string
	hasDefault                 bool
	hint                       string
	onChange                   func(string)
	variant                    Variant
	disabled                   bool
	invalid                    bool
	fullWidth                  bool
	width                      *unit.Dp
	readOnly                   bool
	maxLength                  int
	rows                       int
	label                      string
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

// NewTextArea creates a multi-line input holding value.
func NewTextArea(value string) TextAreaWidget {
	return TextAreaWidget{
		key:      core.AutoKey("text-area"),
		value:    value,
		hasValue: true,
	}
}

// Key sets an explicit identity for the text area state.
func (t TextAreaWidget) Key(key string) TextAreaWidget {
	t.key = key
	return t
}

func (t TextAreaWidget) Value(value string) TextAreaWidget {
	t.value = value
	t.hasValue = true
	return t
}

func (t TextAreaWidget) DefaultValue(value string) TextAreaWidget {
	t.defaultValue = value
	t.hasDefault = true
	t.hasValue = false
	return t
}

func (t TextAreaWidget) Hint(hint string) TextAreaWidget {
	t.hint = hint
	return t
}

func (t TextAreaWidget) Placeholder(placeholder string) TextAreaWidget {
	t.hint = placeholder
	return t
}

func (t TextAreaWidget) OnChange(fn func(string)) TextAreaWidget {
	t.onChange = fn
	return t
}

func (t TextAreaWidget) Disabled(disabled bool) TextAreaWidget {
	t.disabled = disabled
	return t
}

func (t TextAreaWidget) Invalid(invalid bool) TextAreaWidget {
	t.invalid = invalid
	return t
}

func (t TextAreaWidget) Variant(variant Variant) TextAreaWidget {
	t.variant = variant
	return t
}

func (t TextAreaWidget) FillWidth() TextAreaWidget {
	t.fullWidth = true
	return t
}

// Width sets a fixed width in dp. It overrides FillWidth when both are set.
func (t TextAreaWidget) Width(dp int) TextAreaWidget {
	width := unit.Dp(max(dp, 0))
	t.width = &width
	return t
}

func (t TextAreaWidget) ReadOnly(readOnly bool) TextAreaWidget {
	t.readOnly = readOnly
	return t
}

func (t TextAreaWidget) MaxLength(maxLength int) TextAreaWidget {
	t.maxLength = max(maxLength, 0)
	return t
}

func (t TextAreaWidget) Rows(rows int) TextAreaWidget {
	t.rows = max(rows, 1)
	return t
}

func (t TextAreaWidget) Label(label string) TextAreaWidget {
	t.label = label
	return t
}

// ContextMenu replaces the built-in cut/copy/paste menu shown on right-click.
// The custom menu is responsible for its own labels and actions.
func (t TextAreaWidget) ContextMenu(content menu.Widget) TextAreaWidget {
	t.contextMenu = content
	t.hasContextMenu = true
	return t
}

// DefaultContextMenu enables or disables the built-in edit menu shown on
// right-click. It is enabled by default; a custom menu set through ContextMenu
// takes precedence.
func (t TextAreaWidget) DefaultContextMenu(enabled bool) TextAreaWidget {
	t.defaultContextMenuDisabled = !enabled
	return t
}

func (t TextAreaWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	applyFixedWidth(&gtx, t.width)
	fullKey := core.FullKey(ctx, t.key)
	state := textAreaStateFor(ctx, fullKey)
	state.bind(t)
	currentValue := state.currentValue(t)

	gtx, key, editor, enabled := t.prepareEditor(ctx, gtx, state, currentValue, t.disabled)
	disabled := !enabled
	state.FieldInteraction.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     t.readOnly,
		Invalid:      t.invalid,
	}
	activeTheme := core.ActiveTheme(ctx)
	defaults := state.defaults(t, activeTheme, 0)
	root := resolveInputEditorRootStyle(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: defaults,
		custom:   t.style,
	})
	_, lineHeight := textMetrics(root)
	height := textAreaHeight(gtx, root, lineHeight, t.resolvedRows())
	if height > 0 {
		defaults = state.defaults(t, activeTheme, height)
	}
	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: defaults,
		custom:   t.style,
	})
	editorLayout := t.editorLayoutWithStyle(ctx, key, enabled, editor, activeTheme, root, placeholder, selection)
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return t.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
	})
	if !enabled || (!t.hasContextMenu && t.defaultContextMenuDisabled) {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, t.readOnly, false)
	return EditContextMenu(ctx, gtx, t.key, frame, editor, snapshot, t.contextMenu, t.hasContextMenu, state.queueEdit)
}

func (t TextAreaWidget) prepareEditor(ctx *core.Context, gtx layout.Context, state *textAreaState, currentValue string, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	fieldKey, editor := core.InputEditor(ctx, t.key)
	enabled := gtx.Enabled() && !disabled
	behavior := base.TextEditor{InputHint: key.HintText, ReadOnly: t.readOnly, MaxLength: t.maxLength}
	if !enabled {
		gtx = gtx.Disabled()
	}
	behavior.Update(ctx, gtx, base.TextEditorRequest{
		FieldKey: fieldKey, Editor: editor, Value: currentValue,
		BeforeUpdate: func(gtx layout.Context) { state.takePendingEdit(gtx, editor, t.readOnly) },
		OnChange:     func(value string) { state.requestValue(t, value) },
	})
	return gtx, fieldKey, editor, enabled
}

func (t TextAreaWidget) editorLayoutWithStyle(ctx *core.Context, key string, enabled bool, editor *widget.Editor, active *theme.Theme, content, placeholder, selection style.ResolvedStyle) layout.Widget {
	return t.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, t.hint, active, content, placeholder, selection))
}

func textAreaHeight(gtx layout.Context, resolved style.ResolvedStyle, lineHeight unit.Sp, rows int) unit.Dp {
	if rows <= 0 || lineHeight <= 0 {
		return 0
	}
	height := gtx.Sp(lineHeight) * rows
	if resolved.Box != nil && resolved.Box.Padding != nil {
		padding := resolved.Box.Padding
		height += gtx.Dp(padding.Top + padding.Bottom)
	}
	return gtx.Metric.PxToDp(height)
}

func (t TextAreaWidget) resolvedRows() int {
	if t.rows > 0 {
		return t.rows
	}
	return defaultTextAreaRows
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (t TextAreaWidget) Style(values ...style.Style) TextAreaWidget {
	t.style = style.Join(t.style, style.Join(values...))
	return t
}
