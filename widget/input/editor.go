package input

import (
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func editorLayoutFor(ctx *core.Context, editor *widget.Editor, hint string, active *theme.Theme, content, placeholder, selection style.ResolvedStyle) layout.Widget {
	editorStyle := material.Editor(core.ActiveMaterial(ctx), editor, hint)
	applyInputEditorStyle(&editorStyle, active, content, placeholder, selection)
	return editorStyle.Layout
}

// EditorLayout renders a Gio editor using the same resolved field parts as an
// input. Compound fields such as Combobox use this bridge so their text editor
// stays visually aligned with the standalone input component.
func EditorLayout(ctx *core.Context, editor *widget.Editor, hint string, active *theme.Theme, content, placeholder, selection style.ResolvedStyle) layout.Widget {
	return editorLayoutFor(ctx, editor, hint, active, content, placeholder, selection)
}

// applyInputEditorStyle bridges the already-cascaded field parts to Gio's
// editor. Theme semantics supply the fallback colours, while explicit part
// declarations take precedence, so color and typography interpretation stays on
// the component side of the cascade.
func applyInputEditorStyle(target *material.EditorStyle, active *theme.Theme, content, placeholder, selection style.ResolvedStyle) {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	target.Color = semantic.FieldForeground
	target.HintColor = semantic.FieldPlaceholder
	target.SelectionColor = semantic.Selection
	if content.Text != nil {
		if value, ok := computed.Color(content.Text.Color); ok {
			target.Color = value
		}
		textSize, lineHeight := textMetrics(content)
		if textSize > 0 {
			target.TextSize = textSize
		}
		if lineHeight > 0 {
			target.LineHeight = lineHeight
		}
		if content.Text.LineHeightScale != nil {
			target.LineHeightScale = *content.Text.LineHeightScale
		}
		if content.Text.Typeface != nil {
			target.Font.Typeface = *content.Text.Typeface
		}
		if content.Text.FontStyle != nil {
			target.Font.Style = *content.Text.FontStyle
		}
		if content.Text.FontWeight != nil {
			target.Font.Weight = font.Weight(*content.Text.FontWeight)
		}
	}
	if placeholder.Text != nil {
		if value, ok := computed.Color(placeholder.Text.Color); ok {
			target.HintColor = value
		}
	}
	if selection.Paint != nil {
		if brush, ok := computed.Brush(selection.Paint.Background); ok {
			target.SelectionColor = brush.ColorAt(.5)
		}
	}
}

func textMetrics(resolved style.ResolvedStyle) (size, lineHeight unit.Sp) {
	if resolved.Text == nil {
		return 0, 0
	}
	if resolved.Text.FontSize != nil {
		size = *resolved.Text.FontSize
	}
	if resolved.Text.LineHeight != nil {
		lineHeight = *resolved.Text.LineHeight
	}
	return size, lineHeight
}

func (i Widget) withSemantics(ctx *core.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, i.label, enabled, child)
}

func (t TextAreaWidget) withSemantics(ctx *core.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, t.label, enabled, child)
}

func withEditorSemantics(ctx *core.Context, key, label string, enabled bool, child layout.Widget) layout.Widget {
	if label == "" {
		label = core.FieldLabel(ctx, key)
	}
	description := core.FieldDescription(ctx, key)
	return func(gtx layout.Context) layout.Dimensions {
		// Gio suppresses the editor class when its event source is disabled.
		if !enabled {
			semantic.Editor.Add(gtx.Ops)
		}
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		return child(gtx)
	}
}
