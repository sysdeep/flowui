package input

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/binding"
)

type inputState struct {
	base.FieldInteraction
	valueBinding binding.Binding[string]
	value        string
	styleCache   computed.VisualCache[inputStyleKey, style.Style]
	pendingEdit  EditAction
}

func inputStateFor(ctx *core.Context, key string) *inputState {
	return core.UseState[inputState](ctx, key, stateSlotInput)
}

// inputBindingConfig builds a binding.Config from the widget's value fields.
func inputBindingConfig(widget Widget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *inputState) currentValue(widget Widget) string {
	s.value = s.valueBinding.Current(inputBindingConfig(widget))
	return s.value
}

func (s *inputState) bind(widget Widget) {
	s.valueBinding.Bind(inputBindingConfig(widget))
}

func (s *inputState) requestValue(widget Widget, value string) string {
	s.value, _ = s.valueBinding.Request(inputBindingConfig(widget), value)
	return s.value
}

// queueEdit stores an edit action activated from the context menu; the next
// Layout consumes it.
func (s *inputState) queueEdit(action EditAction) {
	s.pendingEdit = action
}

func (s *inputState) takePendingEdit(gtx layout.Context, editor *widget.Editor, readOnly bool) {
	action := s.pendingEdit
	if action == editNone {
		return
	}
	s.pendingEdit = editNone
	RunEditAction(gtx, editor, action, readOnly)
}

type textAreaState struct {
	base.FieldInteraction
	valueBinding binding.Binding[string]
	value        string
	styleCache   computed.VisualCacheMap[inputStyleKey, style.Style]
	pendingEdit  EditAction
}

func textAreaStateFor(ctx *core.Context, key string) *textAreaState {
	return core.UseState[textAreaState](ctx, key, "textAreaState")
}

func textAreaBindingConfig(widget TextAreaWidget) binding.Config[string] {
	return binding.Config[string]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *textAreaState) currentValue(widget TextAreaWidget) string {
	s.value = s.valueBinding.Current(textAreaBindingConfig(widget))
	return s.value
}

func (s *textAreaState) bind(widget TextAreaWidget) {
	s.valueBinding.Bind(textAreaBindingConfig(widget))
}

func (s *textAreaState) requestValue(widget TextAreaWidget, value string) string {
	s.value, _ = s.valueBinding.Request(textAreaBindingConfig(widget), value)
	return s.value
}

// queueEdit stores an edit action activated from the context menu; the next
// Layout consumes it.
func (s *textAreaState) queueEdit(action EditAction) {
	s.pendingEdit = action
}

func (s *textAreaState) takePendingEdit(gtx layout.Context, editor *widget.Editor, readOnly bool) {
	action := s.pendingEdit
	if action == editNone {
		return
	}
	s.pendingEdit = editNone
	RunEditAction(gtx, editor, action, readOnly)
}

type inputGroupState struct {
	// The group owns shell hover/focus state. Its editor state is kept by the
	// nested Input or TextArea, so embedding inputState here would duplicate
	// value bindings and editor-cache storage without sharing any behavior.
	base.FieldInteraction
	actionBounds []image.Rectangle
	styleCache   computed.VisualCacheMap[inputStyleKey, style.Style]
}

func inputGroupStateFor(ctx *core.Context, key string) *inputGroupState {
	return core.UseState[inputGroupState](ctx, key, stateSlotInputGroup)
}

func (s *inputGroupState) setActionBounds(bounds ...image.Rectangle) {
	s.actionBounds = s.actionBounds[:0]
	for _, bound := range bounds {
		if !bound.Empty() {
			s.actionBounds = append(s.actionBounds, bound)
		}
	}
}

func (s *inputGroupState) shouldFocusPress(event pointer.Event) bool {
	if event.Kind != pointer.Press {
		return true
	}
	for _, bound := range s.actionBounds {
		if pointInRect(bound, event.Position) {
			return false
		}
	}
	return true
}

func pointInRect(rect image.Rectangle, point f32.Point) bool {
	return point.X >= float32(rect.Min.X) && point.X < float32(rect.Max.X) &&
		point.Y >= float32(rect.Min.Y) && point.Y < float32(rect.Max.Y)
}
