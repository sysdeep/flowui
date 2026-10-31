package listbox

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/text"
)

func newContext(_ any) *core.Context {
	return core.New(nil, new(theme.Light()), locale.LanguageEnglish)
}

func newContextWithTheme(_ any, value *theme.Theme) *core.Context {
	clone := *value
	return core.New(nil, &clone, locale.LanguageEnglish)
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *core.Context, key, slot string, value *T) {
	core.UseStateWith(ctx, key, slot, func() *T { return value })
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
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

func TestListBoxOptions(t *testing.T) {
	box := New(listBoxTestItems()).Key("files").SelectedKey("go").
		EmptyText("Nothing here").
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnAction(func(string) {}).
		SelectionMode(SelectionNone).
		DisabledKeys([]string{"rust"}).
		Disabled(true).
		FillWidth().
		AllowEmptySelection().
		HideIndicator().
		MaxHeight(120)

	if box.key != "files" || box.selectedKey != "go" || len(box.items) != 3 {
		t.Fatal("listbox constructor did not set fields")
	}
	if box.emptyText != "Nothing here" || box.onChange == nil || box.onSelectionChange == nil || box.onAction == nil {
		t.Fatal("listbox callbacks/options were not set")
	}
	if box.selectionMode != SelectionNone || !listBoxSameKeys(box.disabledKeys, []string{"rust"}) {
		t.Fatal("listbox selection options were not set")
	}
	if !box.disabled || !box.fullWidth || !box.allowEmpty || !box.hideIndicator || box.maxHeight != 120 {
		t.Fatal("listbox visual options were not set")
	}
}

func TestListBoxInstanceStyleIsLastCascadeLayer(t *testing.T) {
	active := new(theme.Light())
	ctx := newContextWithTheme(nil, active)
	base := WithPadding(New([]Item{{Key: "one", Label: "One"}}).Key("styled").SelectedKey("one"), 0)
	measure := func(widget Widget) image.Point {
		gtx := testLayoutContext()
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		dims := widget.Layout(ctx, gtx)
		core.EndFrame(ctx)
		return dims.Size
	}
	baseline := measure(base)
	styled := measure(base.Style(style.Style{}.Part(style.PartItem, style.Style{}.MinHeight(64))))
	if styled.Y <= baseline.Y {
		t.Fatalf("styled item height = %d, baseline %d", styled.Y, baseline.Y)
	}
	stats := testComponentState[listBoxState](ctx, "styled", stateSlotListBox).defaultsCache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("instance style entered the defaults cache key: %#v", stats)
	}
}

func TestListBoxResolvedPartCacheTracksScopeAndTheme(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	defaults := state.defaultsFor(ctx)
	itemDefaults := defaults.item(ItemDefault)
	rowState := style.State{Hovered: true}
	gtx := testLayoutContext()
	state.resolvedPartCache.prepare(ctx, active, style.Style{})
	state.resolvedItemPart(ctx, gtx, active, "item-transition", ItemDefault, rowState, itemDefaults, style.Style{})
	state.resolvedItemPart(ctx, gtx, active, "item-transition", ItemDefault, rowState, itemDefaults, style.Style{})

	stats := state.resolvedPartCache.stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.Size != 1 {
		t.Fatalf("warm resolved cache stats = %#v, want one hit/miss and one entry", stats)
	}

	restore := core.PushStyle(ctx, style.Style{}.Padding(1))
	state.resolvedPartCache.prepare(ctx, active, style.Style{})
	state.resolvedItemPart(ctx, gtx, active, "item-transition", ItemDefault, rowState, itemDefaults, style.Style{})
	restore()
	stats = state.resolvedPartCache.stats()
	if stats.ScopeInvalidations != 1 || stats.Misses != 2 {
		t.Fatalf("scope invalidation stats = %#v, want one invalidation and two misses", stats)
	}
	state.resolvedPartCache.prepare(ctx, active, style.Style{})
	instanceStyle := style.Style{}.Padding(1)
	state.resolvedPartCache.prepare(ctx, active, instanceStyle)
	state.resolvedItemPart(ctx, gtx, active, "item-transition", ItemDefault, rowState, itemDefaults, instanceStyle)
	stats = state.resolvedPartCache.stats()
	if stats.OverrideInvalidations != 1 || stats.ScopeInvalidations != 2 || stats.Misses != 3 {
		t.Fatalf("override invalidation stats = %#v, want one override/two scope invalidations and three misses", stats)
	}

	replaced := *active
	core.ReplaceTheme(ctx, replaced)
	next := core.ActiveTheme(ctx)
	nextDefaults := state.defaultsFor(ctx)
	state.resolvedPartCache.prepare(ctx, next, style.Style{})
	nextRowState := style.State{Hovered: true}
	state.resolvedItemPart(ctx, gtx, next, "item-transition", ItemDefault, nextRowState, nextDefaults.item(ItemDefault), style.Style{})
	stats = state.resolvedPartCache.stats()
	if stats.ThemeInvalidations != 1 || stats.Misses != 4 {
		t.Fatalf("theme invalidation stats = %#v, want one invalidation and four misses", stats)
	}
}

func TestListBoxResolvedPartCacheBypassesUnsafeDeclarations(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	unsafeStyle := style.Style{}.Transition(style.PropOpacity, time.Second,
		style.TransitionEase(func(progress float32) float32 { return progress }))
	state.resolvedPartCache.prepare(ctx, active, unsafeStyle)
	if state.resolvedPartCache.enabled {
		t.Fatal("custom easing unexpectedly enabled the resolved part cache")
	}

	defaults := state.defaultsFor(ctx)
	state.resolvedItemPart(ctx, testLayoutContext(), active, "unsafe-item", ItemDefault, style.State{}, defaults.item(ItemDefault), unsafeStyle)
	if stats := state.resolvedPartCache.stats(); stats.Misses != 0 {
		t.Fatalf("unsafe instance declaration populated resolved cache: %#v", stats)
	}
}

func TestMultipleConstructor(t *testing.T) {
	box := New(listBoxTestItems()).Key("files").SelectedKeys([]string{"go", "rust"})

	if box.key != "files" || box.selectionMode != SelectionMultiple || len(box.selectedKeys) != 2 || len(box.items) != 3 {
		t.Fatal("multi listbox constructor did not set fields")
	}
	if box.hasEmptyText {
		t.Fatal("multi listbox constructor should leave the localized empty text unresolved")
	}
}

func TestSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := Sections("files", "go", sections)

	if box.key != "files" || box.selectedKey != "go" || len(box.sections) != 2 {
		t.Fatal("section listbox constructor did not set fields")
	}
	if got := box.allItems(); len(got) != 3 {
		t.Fatalf("all items = %d, want 3", len(got))
	}
}

func TestMultipleSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := MultipleSections("files", []string{"go"}, sections)

	if box.key != "files" || box.selectionMode != SelectionMultiple || len(box.selectedKeys) != 1 || len(box.sections) != 2 {
		t.Fatal("multi section listbox constructor did not set fields")
	}
}

func TestListBoxRejectsDuplicateItemKeys(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]Item{
			{Key: "go", Label: "Go"},
			{Key: "go", Label: "Go again"},
		})
	})
}

func TestListBoxRejectsEmptyItemKey(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]Item{{Label: "Missing key"}})
	})
}

func TestListBoxClickActivatesItem(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "rust" {
		t.Fatalf("changed = %q, want rust", changed)
	}
	if action != "rust" {
		t.Fatalf("action = %q, want rust", action)
	}
}

func TestListBoxPointerClickDoesNotShowKeyboardFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "go"
	widget := func() Widget {
		return New(listBoxTestItems()).Key("languages").SelectedKey(selected).
			FillWidth().
			OnChange(func(key string) { selected = key })
	}
	layoutFrame := func(now time.Time) {
		viewport := image.Pt(300, 200)
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
		core.BeginFrameWithViewport(ctx, viewport)
		widget().Layout(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(&ops)
	}

	now := time.Unix(4, 0)
	layoutFrame(now)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(2 * time.Millisecond))
	layoutFrame(now.Add(3 * time.Millisecond))

	state, _ := core.PeekState[listBoxState](ctx, "languages", stateSlotListBox)
	item := state.items["rust"]
	if selected != "rust" {
		t.Fatalf("selected = %q, want rust", selected)
	}
	if !router.Source().Focused(&item.Clickable) {
		t.Fatal("pointer click did not focus the selected item")
	}
	if core.FocusVisible(ctx, &item.Clickable, true) {
		t.Fatal("pointer click displayed a keyboard focus ring")
	}
}

func TestListBoxClickSelectedItemOnlyRunsAction(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
	if action != "go" {
		t.Fatalf("action = %q, want go", action)
	}
}

func TestListBoxClickSelectedItemCanClearSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := "not-called"
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		AllowEmptySelection().
		OnChange(func(key string) {
			changed = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
}

func TestSelectionModeNoneOnlyRunsAction(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		SelectionMode(SelectionNone).
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
	if action != "rust" {
		t.Fatalf("action = %q, want rust", action)
	}
}

func TestMultipleClickAddsSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed []string
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go"}).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"go", "rust"}) {
		t.Fatalf("changed = %#v, want [go rust]", changed)
	}
}

func TestMultipleClickRemovesSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed []string
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go", "rust"}).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"rust"}) {
		t.Fatalf("changed = %#v, want [rust]", changed)
	}
}

func TestMultipleDoesNotMutateSelectedKeys(t *testing.T) {
	selected := []string{"go", "go", "rust"}
	next := listBoxToggleSelectedKeys(selected, "zig")

	if !listBoxSameKeys(selected, []string{"go", "go", "rust"}) {
		t.Fatalf("selected = %#v, want original slice unchanged", selected)
	}
	if !listBoxSameKeys(next, []string{"go", "rust", "zig"}) {
		t.Fatalf("next = %#v, want deduped selection with zig", next)
	}
}

func TestMultipleDoesNotCallSingleChange(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	called := false
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go"}).
		OnChange(func(string) {
			called = true
		}).
		OnSelectionChange(func([]string) {}).
		Layout(ctx, testLayoutContext())

	if called {
		t.Fatal("multi listbox called single-selection OnChange")
	}
}

func TestListBoxDisabledItemIgnoresClick(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"swift": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := false
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		OnChange(func(string) {
			changed = true
		}).
		Layout(ctx, testLayoutContext())

	if changed {
		t.Fatal("disabled item handled click")
	}
}

func TestListBoxDisabledKeysIgnoreClick(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := false
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		DisabledKeys([]string{"rust"}).
		OnChange(func(string) {
			changed = true
		}).
		Layout(ctx, testLayoutContext())

	if changed {
		t.Fatal("disabled key handled click")
	}
}

// These exercise the item-model → nav wiring (Disabled field + disabledKeys
// feeding navigation.List); the navigation algorithms themselves are unit-tested in
// package navigation.

func TestListBoxMoveIndexSkipsDisabledAndDoesNotWrap(t *testing.T) {
	items := listBoxTestItems()
	if got, ok := navigation.Move(sliceList(items, nil), 0, 1, false); !ok || got != 1 {
		t.Fatalf("move from first = %d %v, want 1 true", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, nil), 1, 1, false); ok || got != 1 {
		t.Fatalf("move past end = %d %v, want current false", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, nil), 1, -1, false); !ok || got != 0 {
		t.Fatalf("move up = %d %v, want 0 true", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, []string{"rust"}), 0, 1, false); ok || got != 0 {
		t.Fatalf("move into disabled key = %d %v, want current false", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, []string{"rust"}), 1, 1, false); !ok || got != 0 {
		t.Fatalf("move away from disabled current = %d %v, want 0 true", got, ok)
	}
}

func TestListBoxFirstLastEnabled(t *testing.T) {
	items := []Item{
		{Key: "a", Label: "A", Disabled: true},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C", Disabled: true},
	}
	if got, ok := FirstEnabled(items, nil); !ok || got != 1 {
		t.Fatalf("first enabled = %d %v, want 1 true", got, ok)
	}
	if got, ok := LastEnabled(items, nil); !ok || got != 1 {
		t.Fatalf("last enabled = %d %v, want 1 true", got, ok)
	}
	if _, ok := FirstEnabled(items, []string{"b"}); ok {
		t.Fatal("disabled key was returned as first enabled")
	}
}

func TestListBoxTypeaheadSkipsDisabledAndWraps(t *testing.T) {
	items := []Item{
		{Key: "alpha", Label: "Alpha"},
		{Key: "beta", Label: "Beta", Disabled: true},
		{Key: "bravo", Label: "Bravo"},
		{Key: "charlie", Label: "Charlie"},
	}

	if index, ok := navigation.Match(sliceList(items, nil), 0, "b"); !ok || index != 2 {
		t.Fatalf("typeahead b = %d %v, want enabled Bravo", index, ok)
	}
	if index, ok := navigation.Match(sliceList(items, nil), 3, "a"); !ok || index != 0 {
		t.Fatalf("wrapped typeahead a = %d %v, want Alpha", index, ok)
	}
	if index, ok := navigation.Match(sliceList(items, []string{"bravo"}), 0, "b"); ok || index != 0 {
		t.Fatalf("disabled typeahead b = %d %v, want no match", index, ok)
	}
}

func TestListBoxLayoutKeepsState(t *testing.T) {
	ctx := newContext(nil)
	dims := New(listBoxTestItems()).Key("languages").SelectedKey("go").Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("listbox returned empty dimensions")
	}
}

func TestSectionsLayout(t *testing.T) {
	ctx := newContext(nil)
	dims := Sections("languages", "go", listBoxTestSections()).Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing section listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("section listbox returned empty dimensions")
	}
}

func TestListBoxDataVersionCachesEntries(t *testing.T) {
	widget := Sections("cached", "one", []Section{{Title: "Group", Items: []Item{{Key: "one", Label: "One"}}}}).DataVersion(1)
	state := new(listBoxState)
	entries, items := state.resolveEntries(widget)
	cachedEntries, cachedItems := state.resolveEntries(widget)
	if &entries[0] != &cachedEntries[0] || &items[0] != &cachedItems[0] {
		t.Fatal("unchanged ListBox data version did not reuse entries")
	}
	updatedEntries, _ := state.resolveEntries(widget.DataVersion(2))
	if &entries[0] == &updatedEntries[0] {
		t.Fatal("changed ListBox data version reused stale entries")
	}
}

func TestListBoxFillWidth(t *testing.T) {
	dims := New(listBoxTestItems()).Key("languages").SelectedKey("go").
		FillWidth().
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("width = %d, want 300", dims.Size.X)
	}
}

func TestListBoxMaxHeight(t *testing.T) {
	items := make([]Item, 12)
	for i := range items {
		items[i] = Item{Key: string(rune('a' + i)), Label: "Item"}
	}

	dims := New(items).Key("items").
		MaxHeight(80).
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y > 80 {
		t.Fatalf("height = %d, want <= 80", dims.Size.Y)
	}
}

func TestListBoxStyleControlsItemHeight(t *testing.T) {
	active := new(theme.Light())
	ctx := newContextWithTheme(nil, active)
	dims := New([]Item{{Key: "go", Label: "Go"}}).Key("languages").SelectedKey("go").
		Style(style.Style{}.Part(style.PartItem, style.Style{}.MinHeight(72))).Layout(ctx, testLayoutContext())
	if dims.Size.Y < 72 {
		t.Fatalf("height = %d, want at least 72", dims.Size.Y)
	}
}

func TestListBoxPartItemResolvesPerItemState(t *testing.T) {
	active := new(theme.Light())
	box := WithPadding(New([]Item{{Key: "go", Label: "Go"}, {Key: "rust", Label: "Rust"}}), 0)
	ctx := newContextWithTheme(nil, active)
	declaration := style.Style{}.Part(style.PartItem, style.Style{}.Height(44).On(style.Selected, style.Style{}.Height(64)))
	dims := box.Key("languages").SelectedKey("go").Style(declaration).Layout(ctx, testLayoutContext())
	if dims.Size.Y != 108 {
		t.Fatalf("listbox height = %d, want selected 64 + default 44", dims.Size.Y)
	}
}

func TestListBoxIndicatorVisibility(t *testing.T) {
	box := New(listBoxTestItems()).Key("languages").SelectedKey("go")
	if !box.showIndicator(listBoxTestItems()[0]) {
		t.Fatal("single-selection listbox should show default indicator")
	}
	if box.SelectionMode(SelectionNone).showIndicator(listBoxTestItems()[0]) {
		t.Fatal("none-selection listbox should hide default indicator")
	}
	if box.HideIndicator().showIndicator(Item{
		Key:       "custom",
		Label:     "Custom",
		Indicator: func(bool) core.Widget { return text.New("!") },
	}) {
		t.Fatal("hide indicator should override custom indicator")
	}
}

func TestListBoxCustomIndicatorReceivesSelectedState(t *testing.T) {
	ctx := newContext(nil)
	state := new(listBoxState)
	defaults := state.defaultsFor(ctx)
	gotSelected := false
	item := Item{
		Key:   "custom",
		Label: "Custom",
		Indicator: func(selected bool) core.Widget {
			gotSelected = selected
			return nil
		},
	}

	New([]Item{item}).Key("custom").SelectedKey("custom").
		layoutIndicator(ctx, testLayoutContext(), defaults, item, listBoxPalette{}, 0, true)

	if !gotSelected {
		t.Fatal("custom indicator did not receive selected=true")
	}
}

func TestListBoxSelectionAnimation(t *testing.T) {
	state := new(listBoxItemState)
	active := new(theme.Light())
	defaults := buildListBoxDefaults(active)
	duration := defaults.selectionDuration
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.Selection(gtx, false, duration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.Selection(gtx, true, duration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.Selection(gtx, true, duration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.Selection(gtx, true, duration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func listBoxTestItems() []Item {
	return []Item{
		{Key: "go", Label: "Go", Description: "Simple deployment"},
		{Key: "rust", Label: "Rust", Description: "Memory safety"},
		{Key: "swift", Label: "Swift", Description: "Unavailable", Disabled: true},
	}
}

func listBoxTestSections() []Section {
	return []Section{
		{
			Title: "Languages",
			Items: []Item{
				{Key: "go", Label: "Go"},
				{Key: "rust", Label: "Rust"},
			},
		},
		{
			Title: "Unavailable",
			Items: []Item{
				{Key: "swift", Label: "Swift", Disabled: true},
			},
		},
	}
}
