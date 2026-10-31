package navigation

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

func TestPaginationPageItems(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		total int
		want  []int
	}{
		{name: "short", page: 2, total: 4, want: []int{1, 2, 3, 4}},
		{name: "start", page: 1, total: 12, want: []int{1, 2, 0, 12}},
		{name: "middle", page: 6, total: 12, want: []int{1, 0, 5, 6, 7, 0, 12}},
		{name: "end", page: 12, total: 12, want: []int{1, 0, 11, 12}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := paginationPageItems(test.page, test.total, 1, 1)
			if !samePages(got, test.want) {
				t.Fatalf("paginationPageItems(%d, %d) = %v, want %v", test.page, test.total, got, test.want)
			}
		})
	}
}

func TestPaginationControlledPageClick(t *testing.T) {
	activeTheme := new(theme.Light())
	ctx := core.New(nil, activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	changed := 0
	widget := NewPagination("pages", 1, 5).Size(PaginationSizeSmall).OnChange(func(page int) { changed = page })
	gtx := paginationTestContext(router, image.Pt(480, 80), time.Unix(1, 0))

	core.BeginFrame(ctx)
	dims := widget.Layout(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
	if dims.Size.Y != gtx.Dp(28) {
		t.Fatalf("Pagination height = %d", dims.Size.Y)
	}
	value, ok := core.PeekState[paginationState](ctx, "pages", paginationStateSlotPagination)
	if !ok || value.items["page:2"] == nil {
		t.Fatal("page item state was not retained")
	}
	value.items["page:2"].clickable.Click()

	gtx = paginationTestContext(router, image.Pt(480, 80), time.Unix(1, int64(time.Millisecond)))
	core.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
	if changed != 2 {
		t.Fatalf("changed page = %d, want 2", changed)
	}
}

func TestPaginationDefaultLabelsUseWindowLanguage(t *testing.T) {
	activeTheme := new(theme.Light())
	ctx := core.New(nil, activeTheme, locale.LanguageChinese)
	pagination := NewPagination("pages", 1, 3)
	if got := pagination.previousText(ctx); got != "上一页" {
		t.Fatalf("Chinese previous label = %q", got)
	}
	if got := pagination.nextText(ctx); got != "下一页" {
		t.Fatalf("Chinese next label = %q", got)
	}

	custom := pagination.Labels("Back", "Forward")
	if got := custom.previousText(ctx); got != "Back" {
		t.Fatalf("custom previous label = %q", got)
	}
	if got := custom.nextText(ctx); got != "Forward" {
		t.Fatalf("custom next label = %q", got)
	}
}

func TestPaginationKeyboardAndDisabledActivation(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	changed := 0
	disabled := false
	start := time.Unix(1, 0)
	frame := func() {
		gtx := paginationTestContext(router, image.Pt(480, 80), start)
		core.BeginFrame(ctx)
		NewPagination("pages", 1, 5).Disabled(disabled).
			OnChange(func(page int) { changed = page }).Layout(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(gtx.Ops)
		start = start.Add(time.Millisecond)
	}
	frame()
	state, _ := core.PeekState[paginationState](ctx, "pages", paginationStateSlotPagination)
	click := &state.items["page:2"].clickable
	router.Source().Execute(key.FocusCmd{Tag: click})
	frame()
	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	frame()
	if changed != 2 || !router.Source().Focused(click) {
		t.Fatalf("keyboard page = %d, focused = %v", changed, router.Source().Focused(click))
	}
	changed = 0
	disabled = true
	click.Click()
	state.items["previous"].clickable.Click()
	frame()
	if changed != 0 {
		t.Fatalf("disabled pagination activated page %d", changed)
	}
	disabled = false
	frame()
	if changed != 0 {
		t.Fatalf("disabled click replayed after enabling: page %d", changed)
	}
}

func paginationTestContext(router *input.Router, max image.Point, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Constraints: layout.Constraints{Max: max}, Source: source, Ops: new(op.Ops), Now: now}
}

func samePages(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
