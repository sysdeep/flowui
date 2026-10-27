package selector

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
)

type ComboboxItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
}

type ComboboxWidget struct {
	override                   *theme.ComboOverride
	key                        string
	selectedKey                string
	hasSelectedKey             bool
	defaultSelectedKey         string
	hasDefaultSelected         bool
	items                      []ComboboxItem
	dataVersion                uint64
	hasDataVersion             bool
	hint                       string
	inputValue                 string
	emptyText                  string
	hasEmptyText               bool
	onChange                   func(string)
	onInputChange              func(string)
	variant                    field.Variant
	disabled                   bool
	invalid                    bool
	fullWidth                  bool
	hasInputValue              bool
	allowCustomValue           bool
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

// New creates a combo box. It starts uncontrolled with nothing selected;
// SelectedKey sets the controlled value and DefaultSelectedKey the initial one.
func NewCombobox(items []ComboboxItem) ComboboxWidget {
	return ComboboxWidget{
		key:   core.AutoKey(),
		items: items,
	}
}

// Key sets an explicit identity for the combo box state.
func (c ComboboxWidget) Key(key string) ComboboxWidget {
	c.key = key
	return c
}

func (c ComboboxWidget) Hint(hint string) ComboboxWidget {
	c.hint = hint
	return c
}

// DataVersion enables item validation and filtered-index reuse. Increase
// version whenever the item data changes.
func (c ComboboxWidget) DataVersion(version uint64) ComboboxWidget {
	c.dataVersion = version
	c.hasDataVersion = true
	return c
}

func (c ComboboxWidget) InputValue(value string) ComboboxWidget {
	c.inputValue = value
	c.hasInputValue = true
	return c
}

func (c ComboboxWidget) EmptyText(text string) ComboboxWidget {
	c.emptyText = text
	c.hasEmptyText = true
	return c
}

func (c ComboboxWidget) OnChange(fn func(string)) ComboboxWidget {
	c.onChange = fn
	return c
}

func (c ComboboxWidget) SelectedKey(key string) ComboboxWidget {
	c.selectedKey = key
	c.hasSelectedKey = true
	return c
}

func (c ComboboxWidget) DefaultSelectedKey(key string) ComboboxWidget {
	c.defaultSelectedKey = key
	c.hasDefaultSelected = true
	return c
}

func (c ComboboxWidget) OnInputChange(fn func(string)) ComboboxWidget {
	c.onInputChange = fn
	return c
}

func (c ComboboxWidget) Disabled(disabled bool) ComboboxWidget {
	c.disabled = disabled
	return c
}

func (c ComboboxWidget) Invalid(invalid bool) ComboboxWidget {
	c.invalid = invalid
	return c
}

func (c ComboboxWidget) Variant(variant field.Variant) ComboboxWidget {
	c.variant = variant
	return c
}

func (c ComboboxWidget) FullWidth() ComboboxWidget {
	c.fullWidth = true
	return c
}

func (c ComboboxWidget) AllowCustomValue() ComboboxWidget {
	c.allowCustomValue = true
	return c
}

// ContextMenu replaces the built-in edit menu shown on right click.
func (c ComboboxWidget) ContextMenu(content menu.Widget) ComboboxWidget {
	c.contextMenu = content
	c.hasContextMenu = true
	return c
}

// DefaultContextMenu toggles the built-in cut/copy/paste menu on right click.
func (c ComboboxWidget) DefaultContextMenu(enabled bool) ComboboxWidget {
	c.defaultContextMenuDisabled = !enabled
	return c
}

func (c ComboboxWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithOverride(ctx, gtx, c.override)
}

func (c ComboboxWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.ComboOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	state := comboboxComboBoxStateFor(ctx, c.key)
	key := core.FullKey(ctx, c.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	eventGtx := gtx
	editor := &state.editor
	core.RegisterFieldFocus(ctx, key, editor, eventGtx.Enabled() && !c.disabled)
	editor.SingleLine = true
	editor.Submit = true
	state.beginFrame()
	state.checkItems(c.items, c.hasDataVersion, c.dataVersion)

	// Bind disclosure state
	state.bind(c)
	selectedKey := state.currentSelectedKey(c)
	// Create a modified widget with the current selectedKey for internal use
	activeWidget := c
	activeWidget.selectedKey = selectedKey

	state.syncEditor(editor, activeWidget)
	state.takePendingEdit(eventGtx, editor)
	state.input.Update(ctx, eventGtx, c.disabled, editor)

	if c.disabled {
		eventGtx = eventGtx.Disabled()
	}

	query := editor.Text()
	selectedLabel, _ := state.selectedLabel(activeWidget)
	visible := state.visibleItems(activeWidget, query, selectedLabel)
	processMainEvents := !state.open
	state.updateFocus(gtx.Focused(editor), c.disabled)
	if !c.disabled && processMainEvents {
		state.highlight = -1
		if index, ok := state.updateKeys(gtx, editor, c.items, visible, false); ok {
			activeWidget.selectItem(editor, state, c.items[visible[index]])
			selectedKey = state.selectedKey
			activeWidget.selectedKey = selectedKey
		}
		activeWidget.updateEditor(editor, state, gtx)
		query = editor.Text()
		visible = state.visibleItems(activeWidget, query, selectedLabel)
		state.clampHighlight(c.items, visible)
	}
	// Register navigation filters before laying out the editor semantics. When
	// the popup is already open, the editor's semantic handler would otherwise
	// consume Arrow/Enter events before the overlay can observe them.
	if !c.disabled && state.open {
		if index, ok := state.updateKeys(eventGtx, editor, c.items, visible, false); ok {
			activeWidget.selectItem(editor, state, c.items[visible[index]])
			selectedKey = state.selectedKey
			activeWidget.selectedKey = selectedKey
		}
	}

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.input.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     c.disabled || !gtx.Enabled(),
		Selected:     selectedKey != "",
		Invalid:      c.invalid,
		Open:         state.open,
	}
	activeTheme := core.ActiveTheme(ctx)
	var declarations comboboxComboDeclarationBundle
	if instanceOverride == nil {
		declarations = comboboxComboStyleDeclarationsCached(activeTheme, c.variant, &state.styleCache)
	} else {
		declarations = comboboxComboVisualBundle(activeTheme, c.variant, instanceOverride)
	}
	override := declarations.visual.Override
	resolved := c.resolveStyle(ctx, gtx, state, key, styleState, override, declarations)
	editorStyle := material.Editor(core.ActiveMaterial(ctx), editor, c.hint)
	providerEditorStyle := resolved.visual.EditorStyle(resolved.content, resolved.placeholder, resolved.selection)
	editorStyle.TextSize = providerEditorStyle.TextSize
	editorStyle.LineHeight = providerEditorStyle.LineHeight
	editorStyle.Color = providerEditorStyle.Foreground
	editorStyle.HintColor = providerEditorStyle.Placeholder
	editorStyle.SelectionColor = providerEditorStyle.Selection

	field := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutInput(ctx, gtx, state, editor, resolved, core.WithFieldSemantics(ctx, key, editorStyle.Layout))
	})
	var dims layout.Dimensions
	if !c.disabled && (c.hasContextMenu || !c.defaultContextMenuDisabled) {
		snapshot := input.SnapshotEditor(editor, false, false)
		dims = input.EditContextMenu(ctx, gtx, key, field, editor, snapshot, c.contextMenu, c.hasContextMenu, state.queueEdit)
	} else {
		dims = field(ctx, gtx)
	}
	progress := state.popoverProgress(gtx, state.open && !c.disabled, resolved.visual.Metrics.AnimationDuration, resolved.visual.Motion)
	if progress == 0 {
		state.endFrame()
		return dims
	}

	return c.layoutOpenWithBundle(ctx, eventGtx, state, editor, dims, visible, progress, naturallyDisabled, override, declarations)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ComboboxWidget) Style(values ...style.Style) ComboboxWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ComboOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
