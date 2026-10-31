package core

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestNewUsesProvidedTheme(t *testing.T) {
	active := theme.Light()
	active.Colors.Tones[theme.TonePrimary].Fill = color.NRGBA{R: 1, G: 2, B: 3, A: 255}

	ctx := New(nil, &active, locale.LanguageAuto)
	if ActiveTheme(ctx) != &active {
		t.Fatal("context did not use provided theme")
	}
	snapshot := ctx.Theme()
	snapshot.Colors.Tones[theme.TonePrimary].Fill = color.NRGBA{}
	if ActiveTheme(ctx).Colors.Tones[theme.TonePrimary].Fill != active.Colors.Tones[theme.TonePrimary].Fill {
		t.Fatal("mutating the public theme snapshot changed the active theme")
	}
}

func TestNewWithoutThemeStaysNeutral(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	if ActiveTheme(ctx) != nil {
		t.Fatal("theme-less context installed an implicit theme")
	}
	if got := ActiveMaterial(ctx); got != nil {
		t.Fatal("theme-less context installed an implicit material theme")
	}
	if got := ctx.Theme(); got.Colors != (theme.Semantic{}) {
		t.Fatalf("theme-less context snapshot = %#v, want neutral zero theme", got)
	}
	if got := ctx.ForegroundColor(); got != (color.NRGBA{}) {
		t.Fatalf("theme-less foreground = %#v, want transparent zero", got)
	}
	if got := ctx.BackgroundColor(); got != (color.NRGBA{}) {
		t.Fatalf("theme-less background = %#v, want transparent zero", got)
	}
}

func TestCachedThemeHashCachesByThemePointer(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	first := theme.Light()
	second := theme.Light()
	calls := 0
	compute := func() uint64 {
		calls++
		return uint64(calls)
	}

	if got := CachedThemeHash(ctx, &first, compute); got != 1 {
		t.Fatalf("first hash = %d, want 1", got)
	}
	if got := CachedThemeHash(ctx, &first, compute); got != 1 {
		t.Fatalf("same-theme hash = %d, want 1", got)
	}
	if calls != 1 {
		t.Fatalf("same theme computed %d times, want once", calls)
	}
	if got := CachedThemeHash(ctx, &second, compute); got != 2 {
		t.Fatalf("changed-theme hash = %d, want 2", got)
	}
	if calls != 2 {
		t.Fatalf("changed theme computed %d times, want twice", calls)
	}
	if got := CachedThemeHash(ctx, &second, nil); got != 0 {
		t.Fatalf("nil compute result = %d, want 0", got)
	}
	if got := CachedThemeHash(nil, &first, compute); got != 3 {
		t.Fatalf("nil context hash = %d, want 3", got)
	}
}

func TestNewWithCatalogStrictRequiresTheme(t *testing.T) {
	if _, err := NewWithCatalogStrict(nil, nil, locale.LanguageAuto, i18n.Catalog{}); err != theme.ErrMissingTheme {
		t.Fatalf("nil theme error = %v", err)
	}
	active := theme.Light()
	ctx, err := NewWithCatalogStrict(nil, &active, locale.LanguageAuto, i18n.Catalog{})
	if err != nil {
		t.Fatalf("valid theme rejected: %v", err)
	}
	replacement := theme.Light()
	ReplaceTheme(ctx, replacement)
	if ActiveTheme(ctx) == nil {
		t.Fatal("replacement theme missing")
	}
}

func TestContextResolvesWindowTranslationsAndLanguageChanges(t *testing.T) {
	catalog := i18n.NewCatalog(i18n.Translation{
		Language: locale.Language("fr"),
		Key:      i18n.ActionClose,
		Value:    "Fermer",
	})
	ctx := NewWithCatalog(nil, nil, locale.Language("fr-CA"), catalog)
	if got := ctx.Text(i18n.ActionClose); got != "Fermer" {
		t.Fatalf("French regional override = %q, want Fermer", got)
	}
	if got := ctx.Text(i18n.ActionCopy); got != "Copy" {
		t.Fatalf("missing translation fallback = %q, want Copy", got)
	}
	if got := ctx.Textf(i18n.TabsClose, "Overview"); got != "Close Overview" {
		t.Fatalf("formatted translation fallback = %q, want Close Overview", got)
	}

	ReplaceLanguage(ctx, locale.LanguageChinese)
	if got := ctx.Text(i18n.ActionCopy); got != "复制" {
		t.Fatalf("Chinese built-in translation = %q, want 复制", got)
	}
	if got := ctx.Textf(i18n.TableResizeColumn, "名称"); got != "调整名称列大小" {
		t.Fatalf("formatted Chinese translation = %q, want 调整名称列大小", got)
	}
}

func TestPushStyleRestoresCascadeScope(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	first := style.Style{}.Background(style.RGB(0x112233))
	second := style.Style{}.Padding(12)
	restoreFirst := PushStyle(ctx, first)
	restoreSecond := PushStyle(ctx, second)

	active := ActiveStyles(ctx)
	if len(active) != 2 || active[0].Resolve(style.State{}).Paint.Background != style.RGB(0x112233) {
		t.Fatalf("active styles = %#v", active)
	}
	restoreSecond()
	if len(ActiveStyles(ctx)) != 1 {
		t.Fatalf("nested style was not restored")
	}
	restoreFirst()
	if len(ActiveStyles(ctx)) != 0 {
		t.Fatalf("style scope leaked")
	}
}

func TestContextRejectsDuplicateExplicitKeys(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	ctx.Clickable("save")
	mustPanic(t, func() {
		ctx.Editor("save")
	})
}

func TestContextScopesExplicitKeys(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	pop := PushKey(ctx, "todo:1")
	ctx.Clickable("delete")
	pop()
	ctx.Clickable("delete")

	if state, _ := PeekState[widget.Clickable](ctx, "todo:1/delete", stateSlotClickable); state == nil {
		t.Fatal("missing scoped clickable")
	}
	if state, _ := PeekState[widget.Clickable](ctx, "delete", stateSlotClickable); state == nil {
		t.Fatal("missing root clickable")
	}
}

func TestContextRetainsDraggableStateByKey(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	first := ctx.Draggable("component")
	EndFrame(ctx)

	BeginFrame(ctx)
	second := ctx.Draggable("component")
	if first != second {
		t.Fatal("draggable state was not retained by key")
	}
}

func TestContextPublicFocusHelpersPreserveModality(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	BeginFrame(ctx)
	pointerTarget := new(int)
	ctx.RequestFocusVisible(pointerTarget, false)
	if ctx.FocusVisible(pointerTarget, true) {
		t.Fatal("pointer-originated custom focus was visible")
	}

	keyboardTarget := new(int)
	ctx.RequestFocus(keyboardTarget)
	if !ctx.FocusVisible(keyboardTarget, true) {
		t.Fatal("keyboard-originated custom focus was hidden")
	}
	ctx.RequestFocusVisible(keyboardTarget, false)
	ctx.PreserveFocus()
	if ctx.FocusVisible(keyboardTarget, true) {
		t.Fatal("pointer-originated focus remained visible after preservation")
	}
}

func TestSemanticRegistryExcludesHiddenLayout(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	BeginFrame(ctx)
	RegisterSemantic(ctx, SemanticNode{Key: "tabs", Role: SemanticTabList, Label: "Tabs"})
	LayoutHidden(ctx, layout.Context{Ops: new(op.Ops)}, "hidden", WidgetFunc(func(ctx *Context, _ layout.Context) layout.Dimensions {
		RegisterSemantic(ctx, SemanticNode{Key: "hidden", Role: SemanticTabPanel, Label: "Hidden"})
		return layout.Dimensions{}
	}))
	nodes := Semantics(ctx)
	if len(nodes) != 1 || nodes[0].Key != "tabs" {
		t.Fatalf("semantic registry = %#v", nodes)
	}
}

func TestContextScopesKeysWithSeparators(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)

	pop := PushKey(ctx, "a/b")
	parentSlashKey, parentSlash := ClickableWithKey(ctx, "c")
	pop()

	popA := PushKey(ctx, "a")
	popB := PushKey(ctx, "b")
	nestedKey, nested := ClickableWithKey(ctx, "c")
	popB()
	leafSlashKey, leafSlash := ClickableWithKey(ctx, "b/c")
	popA()

	rootKey, root := ClickableWithKey(ctx, "a/b/c")
	keys := []string{parentSlashKey, nestedKey, leafSlashKey, rootKey}
	states := []*widget.Clickable{parentSlash, nested, leafSlash, root}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] == keys[j] {
				t.Fatalf("distinct component paths produced key %q", keys[i])
			}
			if states[i] == states[j] {
				t.Fatalf("distinct component paths reused state for %q and %q", keys[i], keys[j])
			}
		}
	}
}

func TestContextSweepsUnmountedInteractionState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	ctx.Clickable("save")
	UseState[int](ctx, "custom", "test")
	EndFrame(ctx)

	BeginFrame(ctx)
	EndFrame(ctx)
	if StateLen(ctx) != 0 {
		t.Fatalf("interaction states = %d, want 0", StateLen(ctx))
	}
}

func TestLayoutHiddenRetainsStateWithoutVisibleServices(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	BeginFrame(ctx)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	tag := new(int)
	child := WidgetFunc(func(ctx *Context, gtx layout.Context) layout.Dimensions {
		if gtx.Source.Enabled() {
			t.Fatal("hidden layout retained an enabled event source")
		}
		RegisterFieldFocus(ctx, "hidden-field", tag, true)
		RequestFocus(ctx, tag)
		RegisterOverlay(ctx, OverlayRequest{
			Key:    "hidden-overlay",
			Layout: func(layout.Context, image.Rectangle, bool) layout.Dimensions { return layout.Dimensions{} },
		})
		UseState[widget.Clickable](ctx, "hidden-control", "probe")
		return layout.Dimensions{Size: image.Pt(10, 10)}
	})
	LayoutHidden(ctx, gtx, "hidden-scope", child)
	if len(ctx.fieldFocus) != 0 || len(ctx.overlays.requests) != 0 {
		t.Fatalf("hidden services leaked: fields %d overlays %d", len(ctx.fieldFocus), len(ctx.overlays.requests))
	}
	EndFrame(ctx)
	if _, ok := PeekState[widget.Clickable](ctx, "hidden-control", "probe"); !ok {
		t.Fatal("hidden layout state was not retained by its scope")
	}

	BeginFrame(ctx)
	RetainState(ctx, "hidden-scope")
	EndFrame(ctx)
	if _, ok := PeekState[widget.Clickable](ctx, "hidden-control", "probe"); !ok {
		t.Fatal("retained hidden state was released on the next frame")
	}
}

func TestStateRetentionBoundaryKeepsOnlyOuterScopes(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	const (
		outerScope = "outer"
		innerScope = "inner"
	)

	BeginFrame(ctx)
	restoreOuter := PushStateRetention(ctx, outerScope)
	restoreInner := PushStateRetention(ctx, innerScope)
	if depth := StateRetentionDepth(ctx); depth != 2 {
		t.Fatalf("retention depth = %d, want 2", depth)
	}
	restoreBoundary := PushStateRetentionBoundary(ctx, 1)
	UseState[widget.Clickable](ctx, "boundary-control", "probe")
	restoreBoundary()
	restoreInner()
	restoreOuter()
	EndFrame(ctx)

	BeginFrame(ctx)
	RetainState(ctx, outerScope)
	EndFrame(ctx)
	if _, ok := PeekState[widget.Clickable](ctx, "boundary-control", "probe"); !ok {
		t.Fatal("outer retention scope did not keep boundary state")
	}

	BeginFrame(ctx)
	RetainState(ctx, innerScope)
	EndFrame(ctx)
	if _, ok := PeekState[widget.Clickable](ctx, "boundary-control", "probe"); ok {
		t.Fatal("inner retention scope retained state across its boundary")
	}
}

func TestPreparedAssociationTransitionRequestsAnotherFrame(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	ctx.fieldDescriptions = map[string]string{"email": "Supporting text"}
	ctx.previousDescriptions = map[string]string{"email": "Supporting text"}
	ctx.preparedDescriptions = map[string]struct{}{}
	ctx.previousPreparedDescriptions = map[string]struct{}{"email": {}}

	if !fieldAssociationsChanged(ctx) {
		t.Fatal("prepared-to-unprepared transition did not change field associations")
	}
}

func TestContextTracksWindowState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	state := UpdateWindowConfig(ctx, app.Config{
		Size:      image.Pt(640, 480),
		Mode:      app.Maximized,
		Focused:   true,
		Decorated: true,
		TopMost:   true,
	})
	if state != (WindowState{Size: image.Pt(640, 480), Mode: Maximized, Focused: true, Decorated: true, TopMost: true}) {
		t.Fatalf("window state = %#v", state)
	}
	BeginFrameWithViewport(ctx, image.Pt(620, 440))
	if got := ctx.WindowState(); got.Size != image.Pt(620, 440) || got.Mode != Maximized || !got.Focused {
		t.Fatalf("frame window state = %#v", got)
	}
}

func TestContextRoutesWindowCloseRequests(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	if RequestWindowClose(ctx) {
		t.Fatal("close request succeeded without a lifecycle callback")
	}

	requests := 0
	SetWindowCloseRequest(ctx, func() { requests++ })
	if !RequestWindowClose(ctx) || requests != 1 {
		t.Fatalf("close requests = %d, want 1", requests)
	}
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestClickableUsesScopedState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	first := ctx.Clickable("save")
	EndFrame(ctx)
	BeginFrame(ctx)
	second := ctx.Clickable("save")
	if first != second {
		t.Fatal("clickable did not retain state by key")
	}
	if _, ok := PeekState[widget.Clickable](ctx, "save", "clickable"); !ok {
		t.Fatal("clickable state was not retained")
	}
}
