package menu

import (
	"runtime"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/internal/selectablemenu"
)

func init() {
	selectablemenu.Wrap = wrapSelectableText
}

// wrapSelectableText adds the copy/select-all edit menu to a selectable-text
// field. The menu takes focus while open, so the selection highlight is kept
// painted through PaintSelectionWhenUnfocused.
func wrapSelectableText(ctx *core.Context, gtx layout.Context, key string, field core.Widget, selectable *widget.Selectable, queue func(selectablemenu.Action)) layout.Dimensions {
	return ContextMenu(core.DerivedKey(ctx, key, "context-menu"), field, selectableEditContent(ctx, selectable, queue)).
		FocusTarget(selectable).
		OnOpenChange(func(open bool) { selectable.PaintSelectionWhenUnfocused = open }).
		Layout(ctx, gtx)
}

func selectableEditContent(ctx *core.Context, selectable *widget.Selectable, queue func(selectablemenu.Action)) Widget {
	copyLabel, selectAllLabel := ctx.Text(i18n.ActionCopy), ctx.Text(i18n.ActionSelectAll)
	return New([]Item{
		{
			Key:      "copy",
			Label:    copyLabel,
			Shortcut: selectableShortcutLabel("C"),
			Disabled: selectable.SelectionLen() == 0,
			OnAction: func() { queue(selectablemenu.Copy) },
		},
		{
			Key:      "select-all",
			Label:    selectAllLabel,
			Shortcut: selectableShortcutLabel("A"),
			Disabled: selectable.Text() == "",
			OnAction: func() { queue(selectablemenu.SelectAll) },
		},
	}).Key("selectable-edit")
}

// selectableShortcutLabel renders the platform primary modifier for
// display-only shortcut hints.
func selectableShortcutLabel(keyName string) string {
	if runtime.GOOS == "darwin" {
		return "Command+" + keyName
	}
	return "Ctrl+" + keyName
}
