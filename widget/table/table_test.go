package table

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	inputui "github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	selector "github.com/qianniancn/flowui/widget/selector"
)

type tableProbe struct {
	size        image.Point
	layouts     int
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

type tableOverlayProbe struct {
	anchor image.Rectangle
}

type tableInteractiveProbe struct {
	clicks int
	button widget.Clickable
}

func (p *tableInteractiveProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	for p.button.Clicked(gtx) {
		p.clicks++
	}
	return p.button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(80, 28))}
	})
}

func (p *tableOverlayProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       "cell-overlay",
		Anchor:    image.Rect(1, 2, 2, 3),
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			p.anchor = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(80, 20))}
}

func (p *tableProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestTableGridLineModes(t *testing.T) {
	base := New("members", []Column{{Key: "name"}}, nil)
	if !base.showsGridLines() || base.showsFullGrid() || base.bordered {
		t.Fatal("default table line mode changed")
	}
	grid := base.GridLines(true).Border(true)
	if !grid.showsGridLines() || !grid.showsFullGrid() || !grid.bordered {
		t.Fatal("full grid mode was not enabled")
	}
	hidden := base.GridLines(false)
	if hidden.showsGridLines() || hidden.showsFullGrid() {
		t.Fatal("hidden grid mode still shows lines")
	}
	secondary := base.Variant(VariantSecondary)
	if secondary.usesUnifiedFrame() || !secondary.GridLines(true).usesUnifiedFrame() || !secondary.Border(true).usesUnifiedFrame() {
		t.Fatal("secondary grid and border modes did not select the unified frame")
	}
	if !base.Striped(true).striped {
		t.Fatal("striped rows were not enabled")
	}
}

func TestTableSeparatorRectCentersOnColumnBoundary(t *testing.T) {
	if got := tableSeparatorRect(40, 0, 1, 12); got != image.Rect(40, 0, 41, 12) {
		t.Fatalf("one-pixel separator = %v, want %v", got, image.Rect(40, 0, 41, 12))
	}
	if got := tableSeparatorRect(40, 0, 2, 12); got != image.Rect(39, 0, 41, 12) {
		t.Fatalf("two-pixel separator = %v, want %v", got, image.Rect(39, 0, 41, 12))
	}
}

func TestTableRejectsInvalidStructure(t *testing.T) {
	tests := []struct {
		name    string
		columns []Column
		rows    []Row
	}{
		{name: "no columns"},
		{name: "empty column", columns: []Column{{}}},
		{name: "duplicate column", columns: []Column{{Key: "same"}, {Key: "same"}}},
		{name: "empty row", columns: []Column{{Key: "name"}}, rows: []Row{{Cells: []Cell{{Text: "A"}}}}},
		{name: "duplicate row", columns: []Column{{Key: "name"}}, rows: []Row{{Key: "same", Cells: []Cell{{}}}, {Key: "same", Cells: []Cell{{}}}}},
		{name: "cell mismatch", columns: []Column{{Key: "name"}, {Key: "role"}}, rows: []Row{{Key: "one", Cells: []Cell{{}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid Table structure to panic")
				}
			}()
			new(tableState).check(test.columns, test.rows)
		})
	}
}

func TestTableTracksOverlayAnchorInsideCustomCell(t *testing.T) {
	activeTheme := theme.Light()
	ctx := tableTestContext(&activeTheme)
	probe := new(tableOverlayProbe)
	table := New(
		"tracked-cell",
		[]Column{{Key: "name", Label: "Name", Width: 200}},
		[]Row{
			{Key: "first", Label: "First", Cells: []Cell{{Text: "First"}}},
			{Key: "member", Label: "Member", Cells: []Cell{{Content: probe}}},
		},
	).Variant(VariantSecondary)
	gtx := tableLayoutContext(nil, image.Pt(400, 240), time.Unix(1, 0))
	core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	table.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	tokens := defaultTableGeometry
	cellHeight := gtx.Dp(tokens.CellPaddingY)*2 + 20
	rowOffset := max((gtx.Dp(tokens.RowMinHeight)-cellHeight)/2, 0)
	want := image.Rect(
		gtx.Dp(tokens.CellPaddingX)+1,
		gtx.Dp(tokens.HeaderHeight)+gtx.Dp(tokens.RowMinHeight)+rowOffset+gtx.Dp(tokens.CellPaddingY)+2,
		gtx.Dp(tokens.CellPaddingX)+2,
		gtx.Dp(tokens.HeaderHeight)+gtx.Dp(tokens.RowMinHeight)+rowOffset+gtx.Dp(tokens.CellPaddingY)+3,
	)
	if probe.anchor != want {
		t.Fatalf("custom cell overlay anchor = %v, want %v", probe.anchor, want)
	}
}

func TestTableInteractiveCellOwnsPrimaryClick(t *testing.T) {
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	probe := new(tableInteractiveProbe)
	actions := 0
	table := New(
		"interactive-cell",
		[]Column{{Key: "name", Label: "Name", Width: 180}},
		[]Row{{Key: "first", Label: "First", Cells: []Cell{{Content: probe, Interactive: true}}}},
	).OnAction(func(string) { actions++ })
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, table, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(40, 58)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(40, 58)},
	)
	layoutTableFrame(ctx, router, table, start.Add(time.Millisecond))
	if probe.clicks != 1 {
		t.Fatalf("interactive cell clicks = %d, want 1", probe.clicks)
	}
	if actions != 0 {
		t.Fatalf("row actions = %d, want 0", actions)
	}
	tablePeekState(ctx, "interactive-cell").row("first").clickable.Click()
	layoutTableFrame(ctx, router, table, start.Add(2*time.Millisecond))
	if actions != 1 {
		t.Fatalf("keyboard or programmatic row actions = %d, want 1", actions)
	}
}

func TestTableRowContextMenuUsesCompleteRow(t *testing.T) {
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	menuRows := 0
	table := New(
		"row-menu-table",
		[]Column{{Key: "name", Label: "Name", Width: 180}, {Key: "role", Label: "Role", Width: 180}},
		[]Row{{Key: "first", Label: "First", Cells: []Cell{{Text: "First"}, {Text: "Engineer"}}}},
	).RowContextMenu(func(row Row) menu.Widget {
		if row.Key == "first" {
			menuRows++
		}
		return menu.New("row-actions", []menu.Item{{Key: "open", Label: "Open"}})
	})
	start := time.Unix(1, 0)
	layoutTableOverlayFrame(ctx, router, table, start)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(260, 58),
	})
	layoutTableOverlayFrame(ctx, router, table, start.Add(time.Millisecond))
	layoutTableOverlayFrame(ctx, router, table, start.Add(200*time.Millisecond))
	if menuRows < 2 {
		t.Fatalf("row menu callbacks = %d, want one per frame", menuRows)
	}
	if !core.HasTopOverlay(ctx) {
		t.Fatal("secondary click outside the first cell did not open the row menu")
	}
	if focus := tablePeekState(ctx, "row-menu-table").row("first").focus.TargetOpacity(); focus != 0 {
		t.Fatalf("secondary click row focus opacity = %v, want 0", focus)
	}
}

func TestTableRowContextMenuKeepsPointerFocusHiddenAcrossRepeatedOpens(t *testing.T) {
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	table := New(
		"repeated-row-menu",
		[]Column{{Key: "name", Label: "Name", Width: 180}},
		[]Row{{Key: "first", Label: "First", Cells: []Cell{{Text: "First"}}}},
	).RowContextMenu(func(Row) menu.Widget {
		return menu.New("row-actions", []menu.Item{{Key: "open", Label: "Open"}})
	})
	start := time.Unix(1, 0)
	layoutTableOverlayFrame(ctx, router, table, start)
	for cycle := range 3 {
		openAt := start.Add(time.Duration(cycle*250+1) * time.Millisecond)
		router.Queue(pointer.Event{
			Kind: pointer.Press, Source: pointer.Mouse, PointerID: pointer.ID(cycle*2 + 1),
			Buttons: pointer.ButtonSecondary, Position: f32.Pt(120, 58),
		})
		layoutTableOverlayFrame(ctx, router, table, openAt)
		layoutTableOverlayFrame(ctx, router, table, openAt.Add(200*time.Millisecond))
		if focus := tablePeekState(ctx, "repeated-row-menu").row("first").focus.TargetOpacity(); focus != 0 {
			t.Fatalf("cycle %d row focus opacity = %v, want 0", cycle+1, focus)
		}

		router.Queue(
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: pointer.ID(cycle*2 + 2), Buttons: pointer.ButtonPrimary, Position: f32.Pt(620, 300)},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: pointer.ID(cycle*2 + 2), Position: f32.Pt(620, 300)},
		)
		layoutTableOverlayFrame(ctx, router, table, openAt.Add(201*time.Millisecond))
		layoutTableOverlayFrame(ctx, router, table, openAt.Add(202*time.Millisecond))
		if core.HasTopOverlay(ctx) {
			t.Fatalf("cycle %d context menu overlay remained after outside click", cycle+1)
		}
	}
}

func TestTableRowContextMenuShiftF10FromInteractiveContentRestoresFocus(t *testing.T) {
	tests := []struct {
		name    string
		content core.Widget
	}{
		{name: "input", content: inputui.New("Ada").FillWidth()},
		{name: "select", content: selector.New([]selector.SelectItem{{Key: "engineer", Label: "Engineer"}}).Key("row-role").SelectedKey("engineer").FillWidth()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := tableTestContext(new(theme.Light()))
			router := new(input.Router)
			keyValue := "interactive-row-menu-" + test.name
			table := New(
				keyValue,
				[]Column{{Key: "value", Label: "Value", Width: 220}},
				[]Row{{Key: "first", Label: "First", Cells: []Cell{{Content: test.content, Interactive: true}}}},
			).RowContextMenu(func(Row) menu.Widget {
				return menu.New("row-actions", []menu.Item{{Key: "open", Label: "Open"}})
			})
			start := time.Unix(1, 0)
			layoutTableOverlayFrame(ctx, router, table, start)
			targets := tablePeekState(ctx, keyValue).row("first").focusTargets
			if len(targets) == 0 {
				t.Fatalf("interactive %s focus target was not collected", test.name)
			}
			target := targets[0]

			router.Source().Execute(key.FocusCmd{Tag: target})
			layoutTableOverlayFrame(ctx, router, table, start.Add(time.Millisecond))
			if !router.Source().Focused(target) {
				t.Fatalf("interactive %s did not accept focus", test.name)
			}

			router.Queue(key.Event{Name: key.NameF10, Modifiers: key.ModShift, State: key.Press})
			layoutTableOverlayFrame(ctx, router, table, start.Add(2*time.Millisecond))
			layoutTableOverlayFrame(ctx, router, table, start.Add(200*time.Millisecond))
			if !core.HasTopOverlay(ctx) {
				t.Fatalf("Shift+F10 from the interactive %s did not open the row context menu", test.name)
			}

			router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
			layoutTableOverlayFrame(ctx, router, table, start.Add(201*time.Millisecond))
			if !router.Source().Focused(target) {
				t.Fatalf("closing the row context menu did not restore the interactive %s focus", test.name)
			}
			if !core.FocusVisible(ctx, target, true) {
				t.Fatalf("keyboard-opened row context menu did not restore keyboard-visible %s focus", test.name)
			}
		})
	}
}

func TestTableSingleSelectionAndAction(t *testing.T) {
	columns, rows := tableTestData()
	var changed, action string
	table := New("members", columns, rows).
		SelectionMode(SelectionSingle).
		SelectedKey("kate").
		OnChange(func(key string) { changed = key }).
		OnAction(func(key string) { action = key })
	table.activate("john")
	if changed != "john" || action != "john" {
		t.Fatalf("callbacks = changed %q action %q", changed, action)
	}

	changed = "not-called"
	table.AllowEmptySelection().activate("kate")
	if changed != "" {
		t.Fatalf("allow-empty changed = %q, want empty", changed)
	}
}

func TestTableMultipleSelectionTogglesWithoutMutation(t *testing.T) {
	columns, rows := tableTestData()
	selected := []string{"kate"}
	var changed []string
	New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys(selected).
		OnSelectionChange(func(keys []string) { changed = keys }).
		activate("john")
	if !sameKeys(selected, []string{"kate"}) {
		t.Fatalf("selected input mutated: %#v", selected)
	}
	if !sameKeys(changed, []string{"kate", "john"}) {
		t.Fatalf("changed = %#v", changed)
	}
}

func TestTableShiftSelectionSelectsRangeAndSkipsDisabledRows(t *testing.T) {
	columns, rows := tableTestData()
	var changed []string
	stateValue := &tableState{selectionAnchor: "kate"}
	table := New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"kate"}).
		DisabledKeys([]string{"john"}).
		OnSelectionChange(func(keys []string) { changed = keys })
	table.activateWithModifiers(stateValue, "sara", key.ModShift)
	if !sameKeys(changed, []string{"kate", "sara"}) {
		t.Fatalf("range selection = %v, want kate and sara", changed)
	}
}

func TestTableSelectAllSkipsDisabledAndPreservesDisabledSelection(t *testing.T) {
	columns, rows := tableTestData()
	var changed []string
	table := New("members", columns, rows).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"sara"}).
		DisabledKeys([]string{"sara"}).
		OnSelectionChange(func(keys []string) { changed = keys })
	table.toggleAll()
	if !sameKeys(changed, []string{"kate", "john", "sara"}) {
		t.Fatalf("select all = %#v", changed)
	}

	table.selectedKeys = changed
	table.toggleAll()
	if !sameKeys(changed, []string{"sara"}) {
		t.Fatalf("clear all = %#v, want disabled selection preserved", changed)
	}
}

func TestTableSortCyclesAndChangesColumn(t *testing.T) {
	columns, rows := tableTestData()
	var got SortDescriptor
	table := New("members", columns, rows).
		SortDescriptor(SortDescriptor{Column: "name", Direction: SortAscending}).
		OnSortChange(func(sort SortDescriptor) { got = sort })
	table.requestSort("name")
	if got.Column != "name" || got.Direction != SortDescending {
		t.Fatalf("same-column sort = %#v", got)
	}
	table.requestSort("role")
	if got.Column != "role" || got.Direction != SortAscending {
		t.Fatalf("new-column sort = %#v", got)
	}
}

func TestTableRowClickSelectsAndDisabledRowDoesNot(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	stateValue := new(tableState)
	stateValue.beginFrame()
	stateValue.row("john").clickable.Click()
	stateValue.endFrame()
	tableSetState(ctx, "members", stateValue)
	changed := ""
	view := New("members", columns, rows).
		SelectionMode(SelectionSingle).
		OnChange(func(key string) { changed = key })
	layoutTableFrame(ctx, new(input.Router), view, time.Unix(1, 0))
	if changed != "john" {
		t.Fatalf("row click changed = %q", changed)
	}

	ctx = tableTestContext(new(theme.Light()))
	stateValue = new(tableState)
	stateValue.beginFrame()
	stateValue.row("john").clickable.Click()
	stateValue.endFrame()
	tableSetState(ctx, "members", stateValue)
	changed = ""
	view = New("members", columns, rows).
		SelectionMode(SelectionSingle).
		DisabledKeys([]string{"john"}).
		OnChange(func(key string) { changed = key })
	layoutTableFrame(ctx, new(input.Router), view, time.Unix(1, 0))
	if changed != "" {
		t.Fatalf("disabled row changed = %q", changed)
	}
}

func TestTableKeyboardNavigationSkipsDisabledRows(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	view := New("members", columns, rows).DisabledKeys([]string{"john"})
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, view, start)
	state := tablePeekState(ctx, "members")
	router.Source().Execute(key.FocusCmd{Tag: &state.rows["kate"].clickable})
	layoutTableFrame(ctx, router, view, start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutTableFrame(ctx, router, view, start.Add(2*time.Millisecond))
	if !router.Source().Focused(&state.rows["sara"].clickable) {
		t.Fatal("Down did not skip disabled john")
	}
	router.Queue(key.Event{Name: key.NameHome, State: key.Press})
	layoutTableFrame(ctx, router, view, start.Add(3*time.Millisecond))
	if !router.Source().Focused(&state.rows["kate"].clickable) {
		t.Fatal("Home did not focus kate")
	}
}

func TestTableKeyboardEnterSelectsFocusedRow(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	selected := ""
	widget := func() Widget {
		return New("members", columns, rows).
			SelectionMode(SelectionSingle).
			SelectedKey(selected).
			OnChange(func(key string) { selected = key })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(), start)
	stateValue := tablePeekState(ctx, "members")
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.rows["sara"].clickable})
	layoutTableFrame(ctx, router, widget(), start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press}, key.Event{Name: key.NameReturn, State: key.Release})
	layoutTableFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	if selected != "sara" {
		t.Fatalf("selected = %q, want sara", selected)
	}
}

func TestTableKeyboardSkipsDisabledRows(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	selected, action := "kate", ""
	widget := func() Widget {
		return New("keyboard-disabled", columns, rows).
			SelectionMode(SelectionSingle).
			SelectedKey(selected).
			DisabledKeys([]string{"john"}).
			OnChange(func(key string) { selected = key }).
			OnAction(func(key string) { action = key })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(), start)
	stateValue := tablePeekState(ctx, "keyboard-disabled")
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.rows["kate"].clickable})
	layoutTableFrame(ctx, router, widget(), start.Add(time.Millisecond))

	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutTableFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	layoutTableFrame(ctx, router, widget(), start.Add(3*time.Millisecond))
	if selected != "sara" || action != "sara" {
		t.Fatalf("keyboard callbacks selected=%q action=%q, want sara/sara", selected, action)
	}
}

func TestTableKeyboardShiftSelectsRange(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	selected := []string{"kate"}
	widget := func() Widget {
		return New("keyboard-range", columns, rows).
			SelectionMode(SelectionMultiple).
			SelectedKeys(selected).
			OnSelectionChange(func(keys []string) { selected = append([]string(nil), keys...) })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(), start)
	stateValue := tablePeekState(ctx, "keyboard-range")
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.rows["kate"].clickable})
	layoutTableFrame(ctx, router, widget(), start.Add(time.Millisecond))

	router.Queue(key.Event{Name: key.NameDownArrow, Modifiers: key.ModShift, State: key.Press})
	layoutTableFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameDownArrow, Modifiers: key.ModShift, State: key.Press})
	layoutTableFrame(ctx, router, widget(), start.Add(3*time.Millisecond))
	if !sameKeys(selected, []string{"kate", "john", "sara"}) {
		t.Fatalf("shift range = %#v, want all rows", selected)
	}
}

func TestTableKeyboardShiftSpaceRequestsRangeSelection(t *testing.T) {
	columns, rows := tableTestData()
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	selected := []string{"kate"}
	view := func() Widget {
		return New("members", columns, rows).SelectionMode(SelectionMultiple).
			SelectedKeys(selected).
			OnSelectionChange(func(keys []string) { selected = append([]string(nil), keys...) })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, view(), start)
	stateValue := tablePeekState(ctx, "members")
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.rows["john"].clickable})
	layoutTableFrame(ctx, router, view(), start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameSpace, State: key.Press, Modifiers: key.ModShift})
	layoutTableFrame(ctx, router, view(), start.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameSpace, State: key.Release})
	layoutTableFrame(ctx, router, view(), start.Add(3*time.Millisecond))
	if !sameKeys(selected, []string{"kate", "john"}) {
		t.Fatalf("shift-space range = %#v, want kate/john", selected)
	}
}

func TestTableColumnWidthsFillAndOverflow(t *testing.T) {
	columns := []Column{
		{Key: "fixed", Width: 120},
		{Key: "wide", MinWidth: 100, Weight: 2},
		{Key: "narrow", MinWidth: 80, Weight: 1},
	}
	ctx := tableTestContext(new(theme.Light()))
	gtx := tableLayoutContext(nil, image.Pt(600, 300), time.Time{})
	resolved := New("table", columns, nil).resolveColumns(ctx, gtx, nil, 600, tableGeometryFor(ctx))
	if resolved.width != 600 || sumWidths(resolved) != 600 {
		t.Fatalf("resolved widths = %#v total %d", resolved, sumWidths(resolved))
	}
	if resolved.widths[1] <= resolved.widths[2] {
		t.Fatalf("weighted widths = %#v", resolved.widths)
	}

	resolved = New("table", columns, nil).MinWidth(760).resolveColumns(ctx, gtx, nil, 320, tableGeometryFor(ctx))
	if resolved.width != 760 || sumWidths(resolved) != 760 {
		t.Fatalf("overflow widths = %#v total %d", resolved, sumWidths(resolved))
	}
}

func TestTableColumnWidthsRespectMaximums(t *testing.T) {
	columns := []Column{
		{Key: "fixed", Width: 120, MaxWidth: 120},
		{Key: "flex", MinWidth: 100, MaxWidth: 160},
	}
	ctx := tableTestContext(new(theme.Light()))
	gtx := tableLayoutContext(nil, image.Pt(600, 300), time.Time{})
	resolved := New("table", columns, nil).resolveColumns(ctx, gtx, nil, 600, tableGeometryFor(ctx))
	if resolved.width != 600 || resolved.widths[0] != 120 || resolved.widths[1] != 160 {
		t.Fatalf("resolved capped widths = %#v", resolved)
	}
	if sumWidths(resolved) >= resolved.width {
		t.Fatalf("capped columns should leave trailing table space: %#v", resolved)
	}
}

func TestResizableFlexColumnAdaptsUntilDragged(t *testing.T) {
	columns := []Column{
		{Key: "name", MinWidth: 120, Weight: 2, Resizable: true},
		{Key: "role", MinWidth: 100, Weight: 1},
	}
	ctx := tableTestContext(new(theme.Light()))
	stateValue := new(tableState)
	stateValue.beginFrame()
	first := New("table", columns, nil).resolveColumns(ctx, tableLayoutContext(nil, image.Pt(600, 300), time.Time{}), stateValue, 600, tableGeometryFor(ctx))
	stateValue.endFrame()
	stateValue.beginFrame()
	second := New("table", columns, nil).resolveColumns(ctx, tableLayoutContext(nil, image.Pt(780, 300), time.Time{}), stateValue, 780, tableGeometryFor(ctx))
	stateValue.endFrame()
	if sumWidths(first) != 600 || sumWidths(second) != 780 || second.widths[0] <= first.widths[0] {
		t.Fatalf("resizable flex widths did not adapt: first %#v second %#v", first, second)
	}
}

func TestTableColumnResizerHandlesPointerDrag(t *testing.T) {
	columns := []Column{
		{Key: "name", Label: "Name", Width: 160, MinWidth: 100, MaxWidth: 260, Resizable: true},
		{Key: "role", Label: "Role", MinWidth: 120},
	}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}, {Text: "Engineer"}}}}
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	width := 0
	widget := New("resizable", columns, rows).OnColumnResize(func(_ string, value int) { width = value })
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget, start)
	var hit image.Rectangle
	for _, node := range router.AppendSemantics(nil) {
		if node.Desc.Label == ctx.Textf(i18n.TableResizeColumn, "Name") {
			hit = node.Desc.Bounds
			break
		}
	}
	if hit.Empty() {
		t.Fatal("column resizer has no semantic hit region")
	}
	position := f32.Pt(float32(hit.Min.X+hit.Dx()/2), float32(hit.Min.Y+hit.Dy()/2))
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutTableFrame(ctx, router, widget, start.Add(time.Millisecond))
	position.X += 60
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutTableFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if width != 220 {
		t.Fatalf("resized width = %d, want 220", width)
	}
	stateValue := tablePeekState(ctx, "resizable")
	if !stateValue.columns["name"].resize.dragging {
		t.Fatal("column resizer stopped before pointer release")
	}
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutTableFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	if resize := stateValue.columns["name"].resize; resize.dragging {
		t.Fatal("column resizer remained active after pointer release")
	}
}

func TestVirtualTableOnlyRequestsVisibleRows(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}, {Key: "role", Label: "Role"}}
	calls := 0
	table := NewVirtual("virtual", columns, 10_000, func(index int) Row {
		calls++
		return Row{
			Key:   fmt.Sprintf("row-%d", index),
			Cells: []Cell{{Text: fmt.Sprintf("User %d", index)}, {Text: "Engineer"}},
		}
	}).MaxHeight(140).
		RowHeight(42).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"row-0"}).
		DisabledKeys([]string{"row-2"}).
		ShowSelectionIndicator()
	ctx := tableTestContext(new(theme.Light()))
	core.BeginFrame(ctx)
	table.Layout(ctx, tableLayoutContext(nil, image.Pt(640, 200), time.Unix(1, 0)))
	core.EndFrame(ctx)
	if calls == 0 || calls >= 40 {
		t.Fatalf("virtual row callback calls = %d, want only visible rows", calls)
	}
	stateValue := tablePeekState(ctx, "virtual")
	if len(stateValue.rows) >= 20 {
		t.Fatalf("retained virtual row states = %d", len(stateValue.rows))
	}
}

func TestVirtualTableDoesNotResolveSelectedRowWithoutKeyboardEvent(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	calls := 0
	table := NewVirtual("virtual-selected", columns, 10_000, func(index int) Row {
		calls++
		return Row{Key: fmt.Sprintf("row-%d", index), Cells: []Cell{{Text: "User"}}}
	}).MaxHeight(140).
		RowHeight(42).
		SelectionMode(SelectionSingle).
		SelectedKey("row-9999")
	ctx := tableTestContext(new(theme.Light()))
	core.BeginFrame(ctx)
	table.Layout(ctx, tableLayoutContext(nil, image.Pt(640, 200), time.Unix(1, 0)))
	core.EndFrame(ctx)
	if calls == 0 || calls >= 40 {
		t.Fatalf("virtual row callback calls = %d, want only visible rows", calls)
	}
}

func BenchmarkVirtualTableRowIndex(b *testing.B) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := make([]Row, 10_000)
	for index := range rows {
		rows[index] = Row{Key: fmt.Sprintf("row-%d", index), Cells: []Cell{{Text: "User"}}}
	}
	table := NewVirtual("large", columns, len(rows), func(index int) Row { return rows[index] })
	target := rows[len(rows)-1].Key
	b.ReportAllocs()
	for b.Loop() {
		runtime.KeepAlive(table.rowIndex(target))
	}
}

func BenchmarkTableLargeSelection(b *testing.B) {
	keys := make([]string, 10_000)
	for index := range keys {
		keys[index] = fmt.Sprintf("row-%d", index)
	}
	table := Widget{selectionMode: SelectionMultiple, selectedKeys: keys}
	target := keys[len(keys)-1]
	state := new(tableState)
	b.ReportAllocs()
	for b.Loop() {
		table.selectedKeySet = state.selectedKeys.Resolve(keys)
		runtime.KeepAlive(table.isSelected(target))
	}
}

func BenchmarkVirtualTableLargeLayout(b *testing.B) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := make([]Row, 10_000)
	selected := make([]string, len(rows))
	for index := range rows {
		rows[index] = Row{Key: fmt.Sprintf("row-%d", index), Cells: []Cell{{Text: "User"}}}
		selected[index] = rows[index].Key
	}
	table := NewVirtual("large", columns, len(rows), func(index int) Row { return rows[index] }).
		SelectionMode(SelectionMultiple).
		SelectedKeys(selected).
		MaxHeight(140).
		RowHeight(42)
	ctx := tableTestContext(new(theme.Light()))
	var router input.Router
	b.ReportAllocs()
	for b.Loop() {
		layoutTableFrame(ctx, &router, table, time.Time{})
	}
}

func TestTableLoadMoreLatchesUntilRowsGrow(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}}}}
	ctx := tableTestContext(new(theme.Light()))
	router := new(input.Router)
	requests := 0
	widget := func(rows []Row, loading bool) Widget {
		return New("loading", columns, rows).MaxHeight(180).LoadMore(true, loading, func() { requests++ })
	}
	start := time.Unix(1, 0)
	layoutTableFrame(ctx, router, widget(rows, false), start)
	layoutTableFrame(ctx, router, widget(rows, true), start.Add(time.Millisecond))
	layoutTableFrame(ctx, router, widget(rows, false), start.Add(2*time.Millisecond))
	if requests != 1 {
		t.Fatalf("load requests = %d, want one latched request", requests)
	}
	rows = append(rows, Row{Key: "two", Cells: []Cell{{Text: "Two"}}})
	layoutTableFrame(ctx, router, widget(rows, false), start.Add(3*time.Millisecond))
	if requests != 2 {
		t.Fatalf("load requests after growth = %d, want 2", requests)
	}
}

func TestTableLayoutUsesMaxHeightAndHorizontalViewport(t *testing.T) {
	columns, rows := tableTestData()
	activeTheme := theme.Light()
	ctx := tableTestContext(&activeTheme)
	dims := New("members", columns, rows).
		MinWidth(760).
		MaxHeight(150).
		Layout(ctx, tableLayoutContext(nil, image.Pt(420, 300), time.Time{}))
	if dims.Size.X != 420 {
		t.Fatalf("table width = %d, want viewport 420", dims.Size.X)
	}
	if dims.Size.Y > 154 {
		t.Fatalf("table height = %d, want <= 154", dims.Size.Y)
	}
	stateValue := tablePeekState(ctx, "members")
	if stateValue.horizontal.Axis != layout.Horizontal || stateValue.vertical.Axis != layout.Vertical {
		t.Fatalf("scroll axes = %v/%v", stateValue.horizontal.Axis, stateValue.vertical.Axis)
	}
}

func TestTableContentHeightIncludesHeaderAndBody(t *testing.T) {
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "one", Cells: []Cell{{Text: "One"}}}}
	activeTheme := theme.Light()
	ctx := tableTestContext(&activeTheme)
	gtx := tableLayoutContext(nil, image.Pt(320, 200), time.Time{})
	stateValue := new(tableState)
	stateValue.beginFrame()
	defer stateValue.endFrame()
	stateValue.check(columns, rows)
	defaults := tableDefaultsFor(ctx, VariantPrimary)
	resolved := New("members", columns, rows).resolveColumns(ctx, gtx, nil, 320, defaults.Geometry)
	styles := tableStylesFor(ctx, New("members", columns, rows), defaults)
	dims := New("members", columns, rows).layoutContent(ctx, gtx, stateValue, resolved, defaults, styles)
	if dims.Size.Y != 76 {
		t.Fatalf("content height = %d, want default header 40 + row 36", dims.Size.Y)
	}
}

func TestTableCustomCellInheritsSelectedColors(t *testing.T) {
	activeTheme := theme.Light()
	probe := &tableProbe{size: image.Pt(40, 16)}
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{{Key: "selected", Label: "Selected", Cells: []Cell{{Content: probe}}}}
	ctx := tableTestContext(&activeTheme)
	New("members", columns, rows).
		SelectionMode(SelectionSingle).
		SelectedKey("selected").
		Layout(ctx, tableLayoutContext(nil, image.Pt(320, 160), time.Time{}))
	sem := activeTheme.Semantic()
	wantForeground := sem.Foreground
	wantBackground := sem.Tone(theme.TonePrimary).Soft
	if probe.layouts != 1 || probe.foreground != wantForeground || probe.background != wantBackground {
		t.Fatalf("probe = layouts %d colors %#v/%#v, want fg %#v bg %#v", probe.layouts, probe.foreground, probe.background, wantForeground, wantBackground)
	}
}

func TestStripedTableUsesThemedRowBackgrounds(t *testing.T) {
	activeTheme := theme.Light()
	plain := &tableProbe{size: image.Pt(40, 16)}
	stripe := &tableProbe{size: image.Pt(40, 16)}
	columns := []Column{{Key: "name", Label: "Name"}}
	rows := []Row{
		{Key: "plain", Cells: []Cell{{Content: plain}}},
		{Key: "stripe", Cells: []Cell{{Content: stripe}}},
	}
	New("striped", columns, rows).
		Variant(VariantSecondary).
		GridLines(true).
		Border(true).
		Striped(true).
		Layout(tableTestContext(&activeTheme), tableLayoutContext(nil, image.Pt(320, 180), time.Time{}))

	sem := activeTheme.Semantic()
	plainWant := sem.Surface
	if plain.background != plainWant {
		t.Fatalf("plain row background = %#v, want theme body %#v", plain.background, plainWant)
	}
	stripeWant := sem.SurfaceSecondary
	if stripe.background != stripeWant {
		t.Fatalf("stripe row background = %#v, want theme stripe %#v", stripe.background, stripeWant)
	}
}

func TestTableComposedRegionsInheritSurfaceColors(t *testing.T) {
	activeTheme := theme.Light()
	header := &tableProbe{size: image.Pt(40, 16)}
	empty := &tableProbe{size: image.Pt(40, 16)}
	footer := &tableProbe{size: image.Pt(40, 16)}
	columns := []Column{{Key: "name", Label: "Name", Header: header}}
	New("members", columns, nil).
		EmptyContent(empty).
		Footer(footer).
		Layout(tableTestContext(&activeTheme), tableLayoutContext(nil, image.Pt(320, 240), time.Time{}))
	ctx := tableTestContext(&activeTheme)
	sem := activeTheme.Semantic()
	headerBackground := theme.Mix(sem.Surface, sem.SurfaceHover, 50)
	if header.foreground != sem.MutedForeground || header.background != headerBackground {
		t.Fatalf("header colors = %#v/%#v", header.foreground, header.background)
	}
	geometry := tableGeometryFor(ctx)
	if header.constraints.Min.Y != 0 || header.constraints.Max.Y != int(geometry.HeaderHeight) {
		t.Fatalf("header height constraints = %+v", header.constraints)
	}
	if empty.foreground != sem.Foreground || empty.background != sem.Surface {
		t.Fatalf("empty colors = %#v/%#v", empty.foreground, empty.background)
	}
	if footer.foreground != sem.Foreground || footer.background != sem.Surface {
		t.Fatalf("footer colors = %#v/%#v", footer.foreground, footer.background)
	}
}

func tableTestData() ([]Column, []Row) {
	columns := []Column{
		{Key: "name", Label: "Name", Sortable: true, RowHeader: true},
		{Key: "role", Label: "Role"},
	}
	rows := []Row{
		{Key: "kate", Label: "Kate Moore", Cells: []Cell{{Text: "Kate Moore"}, {Text: "CEO"}}},
		{Key: "john", Label: "John Smith", Cells: []Cell{{Text: "John Smith"}, {Text: "CTO"}}},
		{Key: "sara", Label: "Sara Johnson", Cells: []Cell{{Text: "Sara Johnson"}, {Text: "CMO"}}},
	}
	return columns, rows
}

func tableTestContext(activeTheme *theme.Theme) *core.Context {
	return core.New(nil, activeTheme, locale.LanguageEnglish)
}

func tableLayoutContext(router *input.Router, max image.Point, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Constraints: layout.Constraints{Max: max}, Source: source, Ops: new(op.Ops), Now: now}
}

func layoutTableFrame(ctx *core.Context, router *input.Router, table Widget, now time.Time) {
	gtx := tableLayoutContext(router, image.Pt(640, 320), now)
	core.BeginFrame(ctx)
	table.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func layoutTableOverlayFrame(ctx *core.Context, router *input.Router, table Widget, now time.Time) {
	gtx := tableLayoutContext(router, image.Pt(640, 320), now)
	core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	table.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func tableSetState(ctx *core.Context, key string, value *tableState) {
	core.UseStateWith(ctx, key, stateSlotTable, func() *tableState { return value })
}

func tablePeekState(ctx *core.Context, key string) *tableState {
	value, _ := core.PeekState[tableState](ctx, key, stateSlotTable)
	return value
}

func tableDefaultsFor(ctx *core.Context, variant Variant) tableDefaults {
	active := core.ActiveTheme(ctx)
	return buildTableDefaults(active, tableDefaultsKey{variant: variant})
}

func tableGeometryFor(ctx *core.Context) tableGeometry {
	return tableDefaultsFor(ctx, VariantPrimary).Geometry
}

func tableStylesFor(ctx *core.Context, t Widget, defaults tableDefaults) tableStyles {
	gtx := tableLayoutContext(nil, image.Pt(320, 240), time.Time{})
	return t.resolveStyles(ctx, gtx, core.FullKey(ctx, t.key), style.State{}, defaults)
}

func sumWidths(columns tableColumns) int {
	total := columns.selection
	for _, width := range columns.widths {
		total += width
	}
	return total
}

func TestTableFlowRendersEveryRow(t *testing.T) {
	if !New("flow-option", []Column{{Key: "value"}}, nil).Flow().flow {
		t.Fatal("Flow did not enable flow layout")
	}

	run := func(size image.Point, configure func(Widget) Widget) (layout.Dimensions, int) {
		activeTheme := theme.Light()
		ctx := tableTestContext(&activeTheme)
		probes := make([]*tableProbe, 8)
		rows := make([]Row, 0, len(probes))
		for index := range probes {
			probes[index] = &tableProbe{size: image.Pt(40, 20)}
			rows = append(rows, Row{Key: fmt.Sprintf("row-%d", index), Cells: []Cell{{Content: probes[index]}}})
		}
		instance := configure(New("flow-rows", []Column{{Key: "value", Label: "Value", Width: 320}}, rows))
		gtx := tableLayoutContext(nil, size, time.Unix(1, 0))
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		dims := instance.Layout(ctx, gtx)
		core.EndFrame(ctx)
		rendered := 0
		for _, probe := range probes {
			rendered += probe.layouts
		}
		return dims, rendered
	}

	// A bounded table keeps its internal viewport and renders only what fits.
	bounded, boundedRows := run(image.Pt(400, 200), func(value Widget) Widget { return value.MaxHeight(200) })
	if bounded.Size.Y > 200 {
		t.Fatalf("bounded table height = %d, want at most 200", bounded.Size.Y)
	}
	if boundedRows >= 8 {
		t.Fatalf("bounded table rendered %d rows, want a virtual window", boundedRows)
	}

	// Flow mode renders every row and reports the natural height, so the table
	// can grow inside the page scroll instead of clipping itself.
	flowed, flowedRows := run(image.Pt(400, 100000), func(value Widget) Widget { return value.Flow() })
	if flowedRows != 8 {
		t.Fatalf("flow table rendered %d rows, want all 8", flowedRows)
	}
	if flowed.Size.Y <= 200 {
		t.Fatalf("flow table height = %d, want the natural content height", flowed.Size.Y)
	}
}
