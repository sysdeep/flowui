package input

import (
	"reflect"
	"testing"

	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// TestInputInstanceStyleOverridesFieldBorder pins the styling path where chrome
// enters through the cascade's instance layer, not through a design callback
// slot.
func TestInputInstanceStyleOverridesFieldBorder(t *testing.T) {
	activeTheme := theme.Light()
	defaultCtx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := defaultInputRules(&activeTheme, Primary, Medium, false)

	defaultStyle := computed.ResolveStatic(defaultCtx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	instanceChrome := style.Style{}.
		BorderColor(style.Solid(activeTheme.Colors.Border)).
		BorderWidth(3)
	styledStyle := computed.ResolveStatic(defaultCtx, style.State{}, declaration, style.Style{}, style.Style{}, instanceChrome)

	if defaultStyle.Paint == nil || defaultStyle.Paint.Border == nil || *defaultStyle.Paint.Border.Width != 1 {
		t.Fatal("default input does not have its built-in border")
	}
	if styledStyle.Paint == nil || styledStyle.Paint.Border == nil || *styledStyle.Paint.Border.Width != 3 {
		t.Fatal("instance input style did not override border width")
	}
}

func TestInputEditorStyleAppliesThemeDefaults(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageAuto)
	declaration := defaultInputRules(&active, Primary, Medium, false)
	root := computed.ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	placeholder := computed.ResolvePartStatic(ctx, style.PartPlaceholder, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	selection := computed.ResolvePartStatic(ctx, style.PartSelection, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	editorStyle := material.Editor(core.ActiveMaterial(ctx), &widget.Editor{}, "")
	applyInputEditorStyle(&editorStyle, &active, root, placeholder, selection)

	semantic := active.Semantic()
	if editorStyle.Color != semantic.FieldForeground ||
		editorStyle.HintColor != semantic.FieldPlaceholder ||
		editorStyle.SelectionColor != semantic.Selection {
		t.Fatalf("editor colours = %#v/%#v/%#v, want theme field colours",
			editorStyle.Color, editorStyle.HintColor, editorStyle.SelectionColor)
	}
	if editorStyle.TextSize != active.Typography.ControlSize || editorStyle.LineHeight != unit.Sp(20) {
		t.Fatalf("editor metrics = %v/%v, want %v/%v",
			editorStyle.TextSize, editorStyle.LineHeight, active.Typography.ControlSize, unit.Sp(20))
	}
}

func TestStyleDeclarationCacheLoadAndInvalidation(t *testing.T) {
	firstTheme := theme.Light()
	secondTheme := theme.Light()
	declaration := style.Style{}.Height(unit.Dp(24))
	cache := computed.VisualCache[inputStyleKey, style.Style]{}
	key := inputStyleKey{variant: Primary, size: Medium}
	cache.Resolve(&firstTheme, key, func() style.Style {
		return declaration
	})

	tests := []struct {
		name          string
		theme         *theme.Theme
		variant       Variant
		sizeTier      Size
		fullWidth     bool
		dynamicHeight unit.Dp
		wantHit       bool
	}{
		{name: "same key", theme: &firstTheme, variant: Primary, sizeTier: Medium, wantHit: true},
		{name: "theme pointer", theme: &secondTheme, variant: Primary},
		{name: "variant", theme: &firstTheme, variant: Secondary},
		{name: "size tier", theme: &firstTheme, variant: Primary, sizeTier: Small},
		{name: "full width", theme: &firstTheme, variant: Primary, fullWidth: true},
		{name: "dynamic height", theme: &firstTheme, variant: Primary, dynamicHeight: 24},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cache.Stats()
			cache.Resolve(test.theme, inputStyleKey{
				variant: test.variant, size: test.sizeTier,
				fullWidth: test.fullWidth, height: test.dynamicHeight,
			}, func() style.Style {
				return declaration
			})
			after := cache.Stats()
			gotHit := after.Misses == before.Misses
			if gotHit != test.wantHit {
				t.Fatalf("cache hit = %v, want %v (stats before=%#v after=%#v)", gotHit, test.wantHit, before, after)
			}
		})
	}
}

func TestInputStateStyleDeclarationsCache(t *testing.T) {
	activeTheme := theme.Light()
	otherTheme := theme.Light()
	state := &inputState{}
	widget := New("").Key("name").Variant(Secondary)

	firstDefaults := state.defaults(widget, &activeTheme)
	secondDefaults := state.defaults(widget, &activeTheme)
	if firstDefaults.Hash64() != secondDefaults.Hash64() {
		t.Fatal("repeated input declaration lookup changed the cached layers")
	}
	stats := state.styleCache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("input cache stats after repeated lookup = %#v", stats)
	}
	state.defaults(widget.Size(Small), &activeTheme)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("input size change did not invalidate cache: %#v", stats)
	}

	state.defaults(widget.FillWidth(), &activeTheme)
	stats = state.styleCache.Stats()
	if stats.Misses != 3 || stats.KeyInvalidations < 2 {
		t.Fatalf("input full-width change did not invalidate cache: %#v", stats)
	}
	state.defaults(widget, &otherTheme)
	stats = state.styleCache.Stats()
	if stats.Misses != 4 || stats.ThemeInvalidations == 0 {
		t.Fatalf("input theme change did not invalidate cache: %#v", stats)
	}
}

func TestInputEditorPartsBatchMatchesIndividualResolution(t *testing.T) {
	activeTheme := theme.Light()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	state := style.State{Hovered: true, Focused: true}
	defaults := defaultInputRules(&activeTheme, Primary, Medium, false)
	custom := style.Style{}.
		TextColor(style.RGB(0x123456)).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.RGB(0x654321)))

	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      "batch-input",
		state:    state,
		defaults: defaults,
		custom:   custom,
	})
	wantRoot := computed.Resolve(ctx, gtx, "individual-input", state, defaults, style.Style{}, style.Style{}, custom)
	wantPlaceholder := computed.ResolvePart(ctx, gtx, "individual-input", style.PartPlaceholder, state, defaults, style.Style{}, style.Style{}, custom)
	wantSelection := computed.ResolvePart(ctx, gtx, "individual-input", style.PartSelection, state, defaults, style.Style{}, style.Style{}, custom)
	if !reflect.DeepEqual(root, wantRoot) || !reflect.DeepEqual(placeholder, wantPlaceholder) || !reflect.DeepEqual(selection, wantSelection) {
		t.Fatalf("batch editor parts differ from individual resolution: got %#v/%#v/%#v want %#v/%#v/%#v", root, placeholder, selection, wantRoot, wantPlaceholder, wantSelection)
	}
}

func TestInputEditorRootResolutionMatchesBatchRoot(t *testing.T) {
	activeTheme := theme.Light()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	request := inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      "root-input",
		state:    style.State{Hovered: true, Focused: true},
		defaults: style.Style{}.TextColor(style.RGB(0x123456)),
		custom:   style.Style{}.Part(style.PartPlaceholder, style.Style{}.TextColor(style.RGB(0x654321))),
	}
	want, _, _ := resolveInputEditorStyles(request)
	if got := resolveInputEditorRootStyle(request); !reflect.DeepEqual(got, want) {
		t.Fatalf("root-only editor resolution differs from batch root: got %#v want %#v", got, want)
	}
}

func TestTextAreaStyleDeclarationsCacheTracksHeight(t *testing.T) {
	activeTheme := theme.Light()
	state := &textAreaState{}
	widget := NewTextArea("").Key("notes").Variant(Primary)

	state.defaults(widget, &activeTheme, 40)
	stats := state.styleCache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("textarea first height cache stats = %#v", stats)
	}
	state.defaults(widget, &activeTheme, 80)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("textarea height change did not invalidate cache: %#v", stats)
	}
}

func TestTextAreaVisualCacheRetainsProvisionalAndResolvedHeights(t *testing.T) {
	active := theme.Light()
	state := &textAreaState{}
	widget := NewTextArea("").Key("notes-cache").Rows(4)

	state.defaults(widget, &active, 0)
	state.defaults(widget, &active, 76)
	state.defaults(widget, &active, 0)
	state.defaults(widget, &active, 76)

	stats := state.styleCache.Stats()
	if stats.Misses != 2 || stats.Hits != 2 || stats.Size != 2 {
		t.Fatalf("textarea multi-height cache stats = %#v, want two misses/two hits/two entries", stats)
	}
}

func TestGroupStyleDeclarationsCacheTracksHeight(t *testing.T) {
	activeTheme := theme.Light()
	state := &inputGroupState{}
	widget := Group(New("").Key("query")).Variant(Primary)

	state.defaults(widget, &activeTheme, 36)
	stats := state.styleCache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("input-group first height cache stats = %#v", stats)
	}
	state.defaults(widget, &activeTheme, 76)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("input-group height change did not invalidate cache: %#v", stats)
	}
}

func TestGroupVisualCacheRetainsProvisionalAndResolvedHeights(t *testing.T) {
	active := theme.Light()
	state := &inputGroupState{}
	value := GroupTextArea(NewTextArea("").Key("group-cache").Rows(4))

	state.defaults(value, &active, 0)
	state.defaults(value, &active, 76)
	state.defaults(value, &active, 0)
	state.defaults(value, &active, 76)

	stats := state.styleCache.Stats()
	if stats.Misses != 2 || stats.Hits != 2 || stats.Size != 2 {
		t.Fatalf("group multi-height cache stats = %#v, want two misses/two hits/two entries", stats)
	}
}
