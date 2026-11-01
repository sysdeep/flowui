package sidebar

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"slices"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestSidebarGroupDescriptionUsesCatalog(t *testing.T) {
	english := core.NewWithCatalog(nil, nil, locale.LanguageEnglish, i18n.Catalog{})
	if got := sidebarGroupDescription(english, false); got != "Navigation group. Activate to expand." {
		t.Fatalf("English collapsed group description = %q", got)
	}

	chinese := core.NewWithCatalog(nil, nil, locale.LanguageChinese, i18n.Catalog{})
	if got := sidebarGroupDescription(chinese, true); got != "导航组。激活以折叠。" {
		t.Fatalf("Chinese expanded group description = %q", got)
	}
}

func TestSidebarCopiesMutableInputs(t *testing.T) {
	items := []Item{{Key: "home", Label: "Home"}}
	disabled := []string{"settings"}
	widget := New("home", items).Key("primary").DisabledKeys(disabled)
	items[0].Key = "changed"
	disabled[0] = "changed"

	if widget.items[0].Key != "home" || widget.disabledKeys[0] != "settings" {
		t.Fatal("Sidebar retained caller slices")
	}
}

func TestSidebarSectionsCopyAndCollapseHeaders(t *testing.T) {
	sections := []Section{
		{Title: "Workspace", Items: []Item{{Key: "home", Label: "Home"}}},
		{Title: "Account", Items: []Item{{Key: "settings", Label: "Settings"}}},
	}
	widget := NewSections("home", sections).Key("primary")
	sections[0].Items[0].Key = "changed"
	entries, items := widget.entriesAndItems()
	if len(entries) != 4 || len(items) != 2 || entries[0].title != "Workspace" || items[0].Key != "home" {
		t.Fatalf("expanded entries = %#v items = %#v", entries, items)
	}
	collapsedEntries, _ := widget.Collapsed(true).entriesAndItems()
	if len(collapsedEntries) != 3 || !collapsedEntries[1].section || collapsedEntries[1].title != "" {
		t.Fatalf("collapsed entries = %#v", collapsedEntries)
	}
}

func TestSidebarSectionsNormalizeEmptyItemsToNil(t *testing.T) {
	widget := NewSections("", []Section{{Title: "Empty", Items: []Item{}}}).Key("primary")
	if widget.sections[0].Items != nil {
		t.Fatalf("empty section items = %#v, want nil", widget.sections[0].Items)
	}
}

func TestSidebarDataVersionCachesEntries(t *testing.T) {
	widget := NewSections("home", []Section{{Title: "Main", Items: []Item{{Key: "home", Label: "Home"}}}}).Key("cached").DataVersion(1)
	state := new(sidebarState)
	entries, items := state.resolveEntries(widget)
	cachedEntries, cachedItems := state.resolveEntries(widget)
	if &entries[0] != &cachedEntries[0] || &items[0] != &cachedItems[0] {
		t.Fatal("unchanged Sidebar data version did not reuse entries")
	}
	updatedEntries, _ := state.resolveEntries(widget.DataVersion(2))
	if &entries[0] == &updatedEntries[0] {
		t.Fatal("changed Sidebar data version reused stale entries")
	}
}

func TestSidebarNestedItemsUseControlledOpenKeys(t *testing.T) {
	items := []Item{{
		Key:   "workspace",
		Label: "Workspace",
		Children: []Item{
			{Key: "overview", Label: "Overview"},
			{Key: "reports", Label: "Reports"},
		},
	}}
	widget := New("overview", items).Key("nested").DataVersion(1)
	state := new(sidebarState)
	closed, closedItems := state.resolveEntries(widget)
	if len(closed) != 1 || len(closedItems) != 1 || closed[0].item.Key != "workspace" {
		t.Fatalf("closed entries = %#v items = %#v", closed, closedItems)
	}
	open, openItems := state.resolveEntries(widget.OpenKeys([]string{"workspace"}))
	if len(open) != 3 || len(openItems) != 3 || open[1].depth != 1 || open[1].parentKey != "workspace" {
		t.Fatalf("open entries = %#v items = %#v", open, openItems)
	}
	collapsed, collapsedItems := state.resolveEntries(widget.OpenKeys([]string{"workspace"}).Collapsed(true))
	if len(collapsed) != 1 || len(collapsedItems) != 1 {
		t.Fatalf("collapsed entries = %#v items = %#v", collapsed, collapsedItems)
	}
}

func TestSidebarOpenKeysRequestIsImmutable(t *testing.T) {
	keys := []string{"workspace"}
	var requested []string
	widget := New("", nil).Key("nested").OpenKeys(keys).OnOpenChange(func(next []string) { requested = next })
	keys[0] = "changed"
	widget.requestOpen("reports", true)
	if got, want := requested, []string{"workspace", "reports"}; !slices.Equal(got, want) {
		t.Fatalf("requested keys = %q, want %q", got, want)
	}
	widget.requestOpen("workspace", false)
	if got, want := requested, []string{}; !slices.Equal(got, want) {
		t.Fatalf("requested keys = %q, want %q", got, want)
	}
}

func TestSidebarPointerActivationRequestsOpenKeys(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	var openKeys []string
	widget := func() Widget {
		return New("", []Item{{
			Key:      "workspace",
			Label:    "Workspace",
			Children: []Item{{Key: "overview", Label: "Overview"}},
		}}).Key("nested").OpenKeys(openKeys).OnOpenChange(func(next []string) { openKeys = next })
	}
	now := time.Unix(10, 0)
	layoutSidebarFrame(ctx, router, widget(), now)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(24, 24)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(24, 24)},
	)
	layoutSidebarFrame(ctx, router, widget(), now.Add(time.Millisecond))
	layoutSidebarFrame(ctx, router, widget(), now.Add(2*time.Millisecond))
	if got, want := openKeys, []string{"workspace"}; !slices.Equal(got, want) {
		t.Fatalf("open keys = %q, want %q", got, want)
	}
}

func TestSidebarKeyboardActivationRequestsOpenKeys(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	var openKeys []string
	widget := func() Widget {
		return New("", []Item{{
			Key:      "workspace",
			Label:    "Workspace",
			Children: []Item{{Key: "overview", Label: "Overview"}},
		}}).Key("nested").OpenKeys(openKeys).OnOpenChange(func(next []string) { openKeys = next })
	}
	now := time.Unix(11, 0)
	layoutSidebarFrame(ctx, router, widget(), now)
	state, _ := core.PeekState[sidebarState](ctx, "nested", stateSlotSidebar)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["workspace"].clickable})
	layoutSidebarFrame(ctx, router, widget(), now.Add(time.Millisecond))
	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	layoutSidebarFrame(ctx, router, widget(), now.Add(2*time.Millisecond))
	if got, want := openKeys, []string{"workspace"}; !slices.Equal(got, want) {
		t.Fatalf("open keys = %q, want %q", got, want)
	}
}

func BenchmarkSidebarLargeData(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := Widget{key: "large", items: items}
	for _, benchmark := range []struct {
		name   string
		widget Widget
	}{
		{name: "unversioned", widget: widget},
		{name: "versioned", widget: widget.DataVersion(1)},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			state := new(sidebarState)
			state.resolveEntries(benchmark.widget)
			b.ReportAllocs()
			for b.Loop() {
				entries, resolved := state.resolveEntries(benchmark.widget)
				runtime.KeepAlive(entries)
				runtime.KeepAlive(resolved)
			}
		})
	}
}

func BenchmarkSidebarLargeLayout(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New("", items).Key("large").DataVersion(1)
	ctx := sidebarTestContext(new(theme.Light()))
	b.ReportAllocs()
	for b.Loop() {
		var router input.Router
		layoutSidebarFrame(ctx, &router, widget, time.Time{})
	}
}

func BenchmarkSidebarLargeLayoutSteadyState(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New("", items).Key("large-steady").DataVersion(1)
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	var ops op.Ops
	viewport := image.Pt(500, 320)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
	}
	b.ReportAllocs()
	for b.Loop() {
		ops.Reset()
		core.BeginFrameWithViewport(ctx, viewport)
		widget.Layout(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(&ops)
	}
}

func TestSidebarItemHeightOverridesDefault(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	viewport := image.Pt(248, 100)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: &ops}
	item := Item{Key: "home", Label: "Home"}
	state := new(sidebarState)
	widget := New("home", []Item{item}).Key("primary").ItemHeight(32)
	active := core.ActiveTheme(ctx)
	defaults := state.defaults.resolve(active, widget.metrics(active))
	resolved := sidebarResolved{defaults: defaults, styleKey: core.FullKey(ctx, widget.key)}
	dims := widget.layoutItem(ctx, gtx, state, new(sidebarItemState), &entry{item: item}, 1, resolved)
	if dims.Size != image.Pt(248, 32) {
		t.Fatalf("item dimensions = %v, want (248,32)", dims.Size)
	}
}

func TestSidebarUsesExpandedAndCollapsedWidths(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	widget := New("home", sidebarTestItems()).Key("primary")
	expanded := layoutSidebarFrame(ctx, new(input.Router), widget, time.Unix(1, 0))
	if expanded.Size != image.Pt(248, 320) {
		t.Fatalf("expanded dimensions = %v", expanded.Size)
	}

	start := time.Unix(2, 0)
	for _, test := range []struct {
		name string
		at   time.Time
		want int
	}{
		{"collapse start", start, 248},
		{"collapse midpoint", start.Add(100 * time.Millisecond), 156},
		{"collapse end", start.Add(200 * time.Millisecond), 64},
	} {
		dims := layoutSidebarFrame(ctx, new(input.Router), widget.Collapsed(true), test.at)
		if dims.Size != image.Pt(test.want, 320) {
			t.Fatalf("%s dimensions = %v", test.name, dims.Size)
		}
	}

	start = time.Unix(3, 0)
	for _, test := range []struct {
		name string
		at   time.Time
		want int
	}{
		{"expand start", start, 64},
		{"expand midpoint", start.Add(100 * time.Millisecond), 225},
		{"expand end", start.Add(200 * time.Millisecond), 248},
	} {
		dims := layoutSidebarFrame(ctx, new(input.Router), widget, test.at)
		if dims.Size != image.Pt(test.want, 320) {
			t.Fatalf("%s dimensions = %v", test.name, dims.Size)
		}
	}
}

func TestCollapsedSidebarLaysOutItemIcons(t *testing.T) {
	icon := &sidebarProbe{size: image.Pt(18, 18)}
	widget := New("home", []Item{{Key: "home", Label: "Home", Leading: icon}}).Key("primary").Collapsed(true)
	layoutSidebarFrame(sidebarTestContext(new(theme.Light())), new(input.Router), widget, time.Unix(2, 0))
	if icon.layouts != 1 {
		t.Fatalf("collapsed item icon layouts = %d, want 1", icon.layouts)
	}
}

func TestCollapsedSidebarCompactMetricsPreserveIconSpace(t *testing.T) {
	icon := &sidebarProbe{size: image.Pt(32, 32)}
	widget := New("home", []Item{{Key: "home", Label: "Home", Leading: icon}}).Key("primary").
		Collapsed(true).
		Width(40).
		CollapsedWidth(40).
		Padding(4).
		ItemGap(0).
		ItemHeight(36).
		ItemPaddingX(0).
		ItemRadius(0)
	dims := layoutSidebarFrame(sidebarTestContext(new(theme.Light())), new(input.Router), widget, time.Unix(2, 0))
	if dims.Size.X != 40 {
		t.Fatalf("collapsed width = %d, want 40", dims.Size.X)
	}
	if icon.maxWidth != 32 {
		t.Fatalf("icon max width = %d, want 32", icon.maxWidth)
	}
}

func TestSidebarExpandedHeaderUsesFinalWidthDuringTransition(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	header := &sidebarProbe{size: image.Pt(20, 20)}
	widget := New("home", sidebarTestItems()).Key("primary").Header(header)
	layoutSidebarFrame(ctx, new(input.Router), widget.Collapsed(true), time.Unix(1, 0))
	if header.maxWidth != 48 {
		t.Fatalf("collapsed header max width = %d, want 48", header.maxWidth)
	}
	layoutSidebarFrame(ctx, new(input.Router), widget, time.Unix(2, 0))
	if header.maxWidth != 232 {
		t.Fatalf("expanding header max width = %d, want 232", header.maxWidth)
	}
}

func TestCollapsedSidebarTracksItemOverlayThroughHeaderAndList(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	viewport := image.Pt(500, 320)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: &ops}
	anchor := image.Rectangle{}
	icon := &sidebarOverlayProbe{size: image.Pt(18, 18), anchor: &anchor}
	widget := New("home", []Item{{Key: "home", Label: "Home", Leading: icon}}).Key("primary").
		Header(&sidebarProbe{size: image.Pt(20, 20)}).
		Collapsed(true)

	core.BeginFrameWithViewport(ctx, viewport)
	widget.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
	if anchor != image.Rect(23, 47, 41, 65) {
		t.Fatalf("collapsed item overlay anchor = %v", anchor)
	}
}

func TestSidebarKeyboardNavigationSkipsDisabledAndActivates(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	selected := "home"
	widget := func() Widget {
		return New(selected, sidebarTestItems()).Key("primary").
			DisabledKeys([]string{"projects"}).
			OnChange(func(key string) { selected = key })
	}
	now := time.Unix(3, 0)
	layoutSidebarFrame(ctx, router, widget(), now)
	state, _ := core.PeekState[sidebarState](ctx, "primary", stateSlotSidebar)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["home"].clickable})
	layoutSidebarFrame(ctx, router, widget(), now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutSidebarFrame(ctx, router, widget(), now.Add(2*time.Millisecond))
	if !router.Source().Focused(&state.items["reports"].clickable) {
		t.Fatal("Down did not skip the disabled destination")
	}
	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	layoutSidebarFrame(ctx, router, widget(), now.Add(3*time.Millisecond))
	if selected != "reports" {
		t.Fatalf("selected = %q, want reports", selected)
	}
}

func TestSidebarPointerClickDoesNotShowKeyboardFocus(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	selected := "home"
	widget := func() Widget {
		return New(selected, sidebarTestItems()).Key("primary").OnChange(func(key string) { selected = key })
	}
	now := time.Unix(4, 0)
	layoutSidebarFrame(ctx, router, widget(), now)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 68)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(30, 68)},
	)
	layoutSidebarFrame(ctx, router, widget(), now.Add(time.Millisecond))
	layoutSidebarFrame(ctx, router, widget(), now.Add(2*time.Millisecond))
	state, _ := core.PeekState[sidebarState](ctx, "primary", stateSlotSidebar)
	if selected != "projects" {
		t.Fatalf("selected = %q, want projects", selected)
	}
	if core.FocusVisible(ctx, &state.items["projects"].clickable, router.Source().Focused(&state.items["projects"].clickable)) {
		t.Fatal("pointer click displayed a keyboard focus ring")
	}
}

func TestSidebarDisablingCancelsPendingActivation(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	router := new(input.Router)
	disabled, actions := false, 0
	view := func() Widget {
		return New("home", sidebarTestItems()).Key("primary").Disabled(disabled).
			OnAction(func(string) { actions++ })
	}
	now := time.Unix(10, 0)
	frame := func() {
		layoutSidebarFrame(ctx, router, view(), now)
		now = now.Add(time.Millisecond)
	}
	frame()
	state, _ := core.PeekState[sidebarState](ctx, "primary", stateSlotSidebar)
	focus := &state.items["home"].clickable
	router.Source().Execute(key.FocusCmd{Tag: focus})
	frame()
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press})
	frame()
	if actions != 0 {
		t.Fatal("activated before release")
	}
	disabled = true
	frame()
	disabled = false
	frame()
	router.Source().Execute(key.FocusCmd{Tag: focus})
	frame()
	router.Queue(key.Event{Name: key.NameReturn, State: key.Release})
	frame()
	if actions != 0 {
		t.Fatal("release activated after disabling")
	}
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press}, key.Event{Name: key.NameReturn, State: key.Release})
	frame()
	if actions != 1 {
		t.Fatalf("fresh action count = %d, want 1", actions)
	}
}

func TestSidebarSelectedColorsFollowTheme(t *testing.T) {
	activeTheme := theme.Light()
	probe := new(sidebarColorProbe)
	widget := New("home", []Item{{Key: "home", Label: "Home", Leading: probe}}).Key("primary")
	layoutSidebarFrame(sidebarTestContext(&activeTheme), new(input.Router), widget, time.Unix(2, 0))
	if probe.background != activeTheme.Colors.Tones[theme.TonePrimary].Soft || probe.foreground != activeTheme.Colors.Tones[theme.TonePrimary].OnSoft {
		t.Fatalf("selected colors = %#v/%#v", probe.background, probe.foreground)
	}
}

func TestSidebarDefaultSwitcherPreservesCustomWidget(t *testing.T) {
	ctx := sidebarTestContext(new(theme.Light()))
	custom := &sidebarProbe{size: image.Pt(7, 9)}
	widget := New("", []Item{{
		Key:      "workspace",
		Label:    "Workspace",
		Switcher: custom,
		Children: []Item{{Key: "overview", Label: "Overview"}},
	}}).Key("primary")
	layoutSidebarFrame(ctx, new(input.Router), widget, time.Unix(2, 0))
	if custom.layouts != 1 {
		t.Fatalf("custom switcher layouts = %d, want 1", custom.layouts)
	}

	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(24, 24)}, Ops: &ops}
	if dims := layoutSidebarSwitcher(gtx, true, color.NRGBA{R: 0x12, A: 0xff}, 12); dims.Size != image.Pt(12, 12) {
		t.Fatalf("default switcher dimensions = %v, want (12,12)", dims.Size)
	}
}

func TestSidebarContentInheritsSurfaceColors(t *testing.T) {
	activeTheme := theme.Light()
	probe := new(sidebarColorProbe)
	widget := New("home", []Item{{Key: "home", Label: "Home"}}).Key("primary").Header(probe)
	layoutSidebarFrame(sidebarTestContext(&activeTheme), new(input.Router), widget, time.Unix(5, 0))
	if probe.foreground != activeTheme.Colors.SurfaceForeground || probe.background != activeTheme.Colors.Surface {
		t.Fatalf("Sidebar content colors = %#v/%#v", probe.foreground, probe.background)
	}
}

func TestSidebarStyleIsTheLastCascadeLayer(t *testing.T) {
	activeTheme := theme.Light()
	probe := new(sidebarColorProbe)
	foreground := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	background := color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff}
	widget := New("", []Item{{Key: "home", Label: "Home", Leading: probe}}).Key("primary").
		Style(style.Style{}.
			TextColor(style.Solid(foreground)).
			Part(style.PartItem, style.Style{}.Background(style.Solid(background))))
	layoutSidebarFrame(sidebarTestContext(&activeTheme), new(input.Router), widget, time.Unix(2, 0))
	if probe.foreground != foreground || probe.background != background {
		t.Fatalf("instance style colors = %#v/%#v, want %#v/%#v", probe.foreground, probe.background, foreground, background)
	}
}

func TestSidebarItemStyleTextColorReachesContent(t *testing.T) {
	activeTheme := theme.Light()
	probe := new(sidebarColorProbe)
	foreground := color.NRGBA{R: 0x77, G: 0x88, B: 0x99, A: 0xff}
	widget := New("", []Item{{Key: "home", Label: "Home", Leading: probe}}).Key("primary").
		Style(style.Style{}.Part(style.PartItem, style.Style{}.TextColor(style.Solid(foreground))))
	layoutSidebarFrame(sidebarTestContext(&activeTheme), new(input.Router), widget, time.Unix(2, 0))
	if probe.foreground != foreground {
		t.Fatalf("item style foreground = %#v, want %#v", probe.foreground, foreground)
	}
}

func TestSidebarDefaultsCacheTracksThemeLifecycle(t *testing.T) {
	state := new(sidebarState)
	firstTheme := new(theme.Light())
	secondTheme := new(theme.Light())
	metrics := (Widget{}).metrics(firstTheme)
	changed := metrics
	changed.Width += 10

	state.defaults.resolve(firstTheme, metrics)
	state.defaults.resolve(firstTheme, metrics)
	state.defaults.resolve(firstTheme, changed)
	state.defaults.resolve(secondTheme, changed)
	stats := state.defaults.stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("defaults cache stats = %#v", stats)
	}
}

type sidebarProbe struct {
	size     image.Point
	layouts  int
	maxWidth int
}

type sidebarOverlayProbe struct {
	size   image.Point
	anchor *image.Rectangle
}

type sidebarColorProbe struct {
	foreground color.NRGBA
	background color.NRGBA
}

func (p *sidebarColorProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: image.Pt(1, 1)}
}

func (p *sidebarOverlayProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       "probe",
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: p.size},
		HasAnchor: true,
		Passive:   true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			*p.anchor = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: p.size}
}

func (p *sidebarProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.maxWidth = gtx.Constraints.Max.X
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func sidebarTestItems() []Item {
	return []Item{
		{Key: "home", Label: "Home"},
		{Key: "projects", Label: "Projects"},
		{Key: "reports", Label: "Reports"},
	}
}

func sidebarTestContext(activeTheme *theme.Theme) *core.Context {
	// Sidebar's transition assertions use the historical 200ms control cycle;
	// keep that timing explicit while the theme tokens stay neutral.
	if activeTheme != nil {
		activeTheme.Motion.Enabled = true
		activeTheme.Motion.DefaultDuration = 200 * time.Millisecond
		activeTheme.Motion.ControlDuration = 200 * time.Millisecond
		activeTheme.Motion.DurationScale = 1
	}
	return core.New(nil, activeTheme, locale.LanguageEnglish)
}

func layoutSidebarFrame(ctx *core.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(500, 320)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
	core.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
