package combobox

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/input"
)

// layoutInputEditor delegates editor rendering to input so compound fields
// inherit the standalone input's typography, placeholder and selection rules.
func layoutInputEditor(ctx *core.Context, gtx layout.Context, editor *widget.Editor, hint string, content, placeholder, selection style.ResolvedStyle) layout.Dimensions {
	return input.EditorLayout(ctx, editor, hint, core.ActiveTheme(ctx), content, placeholder, selection)(gtx)
}
