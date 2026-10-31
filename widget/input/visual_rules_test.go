package input

import (
	"reflect"
	"testing"

	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestInputDefaultsReachResolvedParts(t *testing.T) {
	activeTheme := theme.Light()
	tests := []struct {
		name     string
		defaults style.Style
		part     style.Part
	}{
		{name: "input", defaults: defaultInputRules(&activeTheme, Primary, Medium, false), part: style.PartRoot},
		{name: "textarea", defaults: defaultTextAreaRules(&activeTheme, Primary, false, 48), part: style.PartRoot},
		{name: "group", defaults: defaultGroupRules(&activeTheme, Primary, false, 48), part: style.PartContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !inputDeclarationAnswersToState(test.defaults, test.part) {
				t.Fatalf("disabled state does not affect %s", partName(test.part))
			}
		})
	}

	if declaration := defaultGroupAction(&activeTheme).declaration.Resolve(style.State{Disabled: true}).Box; declaration == nil {
		t.Fatal("group action defaults were not consumed")
	}
}

func TestInputBehaviorStateReachesVisualCascade(t *testing.T) {
	activeTheme := theme.Light()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	defaults := (&inputState{}).defaults(New("").Key("name"), &activeTheme)

	enabled := computed.ResolveStatic(ctx, style.State{}, defaults, style.Style{}, style.Style{}, style.Style{})
	focused := computed.ResolveStatic(ctx, style.State{Focused: true}, defaults, style.Style{}, style.Style{}, style.Style{})
	disabled := computed.ResolveStatic(ctx, style.State{Disabled: true}, defaults, style.Style{}, style.Style{}, style.Style{})
	if enabled.Paint == nil || enabled.Paint.Opacity == nil || *enabled.Paint.Opacity != 1 {
		t.Fatalf("enabled opacity is missing: %s", describeStyle(enabled))
	}
	if disabled.Paint == nil || disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != 1 {
		t.Fatalf("disabled style = %s", describeStyle(disabled))
	}
	semantic := activeTheme.Semantic()
	if !reflect.DeepEqual(disabled.Paint.Background, style.Solid(semantic.DisabledBackground)) ||
		disabled.Text == nil || !reflect.DeepEqual(disabled.Text.Color, style.Solid(semantic.DisabledForeground)) {
		t.Fatalf("disabled style does not use disabled semantic colors: %s", describeStyle(disabled))
	}
	if describeStyle(enabled) == describeStyle(focused) {
		t.Fatal("focus state did not reach the cascade")
	}

	groupDefaults := (&inputGroupState{}).defaults(
		Group(New("").Key("query")),
		&activeTheme,
		36,
	)
	groupEnabled := computed.ResolvePartStatic(
		ctx,
		style.PartContent,
		style.State{},
		groupDefaults,
		style.Style{},
		style.Style{},
		style.Style{},
	)
	groupDisabled := computed.ResolvePartStatic(
		ctx,
		style.PartContent,
		style.State{Disabled: true},
		groupDefaults,
		style.Style{},
		style.Style{},
		style.Style{},
	)
	if groupEnabled.Box == nil || groupEnabled.Box.Cursor == nil || *groupEnabled.Box.Cursor != pointer.CursorText {
		t.Fatalf("enabled group cursor = %#v", groupEnabled.Box)
	}
	if groupDisabled.Box == nil || groupDisabled.Box.Cursor == nil || *groupDisabled.Box.Cursor != pointer.CursorNotAllowed {
		t.Fatalf("disabled group cursor = %#v", groupDisabled.Box)
	}
}

func inputDeclarationAnswersToState(declaration style.Style, part style.Part) bool {
	enabled := normalizeResolvedStyle(style.CascadePart(style.State{}, part, declaration))
	disabled := normalizeResolvedStyle(style.CascadePart(style.State{Disabled: true}, part, declaration))
	return !reflect.DeepEqual(enabled, disabled)
}
