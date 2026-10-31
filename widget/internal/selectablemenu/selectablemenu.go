// Package selectablemenu connects selectable text to the optional context-menu
// implementation without making either widget package import the other.
package selectablemenu

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/base"
)

type Action = base.SelectionAction

const (
	Copy      = base.CopySelection
	SelectAll = base.SelectAll
)

// Wrap is installed by widget/menu during initialization. A nil hook leaves
// selectable text without a context menu; it does not disable keyboard copying.
var Wrap func(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	field core.Widget,
	selectable *widget.Selectable,
	queue func(Action),
) layout.Dimensions
