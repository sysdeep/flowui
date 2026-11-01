package button

import (
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestButtonDefaultCacheIgnoresInstanceStyles(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("Save")
	var cache buttonDeclarationCache
	value.declarations(active, &cache)
	styled := value.Style(style.New().Height(50))
	defaults := styled.declarations(active, &cache)
	resolved := styled.staticStyleParts(ctx, style.State{}, defaults, false)
	if *resolved.root.Box.Height != 50 {
		t.Fatalf("instance height = %v, want 50", *resolved.root.Box.Height)
	}
	styled.Style(style.New().Height(60)).declarations(active, &cache)
	styled.Size(Large).declarations(active, &cache)
	replacement := *active
	styled.Size(Large).declarations(&replacement, &cache)
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 3 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("default cache stats = %+v", stats)
	}
}

func TestButtonStyleCopiesRemainIndependent(t *testing.T) {
	ctx := newContext(nil)
	original := New("Save").Style(style.New().Height(40))
	first := original.Style(style.New().Height(50))
	second := original.Style(style.New().Height(60))
	for _, test := range []struct {
		name   string
		value  Widget
		height int
	}{
		{name: "original", value: original, height: 40},
		{name: "first", value: first, height: 50},
		{name: "second", value: second, height: 60},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolved := test.value.staticStyle(ctx, style.State{})
			if int(*resolved.root.Box.Height) != test.height {
				t.Fatalf("height = %v, want %d", *resolved.root.Box.Height, test.height)
			}
		})
	}
}

func TestPreparedButtonStyleResolvesOnlyLiveParts(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("").Content(nil)
	declarations := value.declarations(active, nil)

	prepared := value.resolvePreparedStyleWithDeclarations(
		ctx,
		testLayoutContext(),
		value.key,
		style.State{},
		declarations,
	)
	if !resolvedStyleEmpty(prepared.indicatorPart) {
		t.Fatalf("prepared indicator part = %#v, want unresolved", prepared.indicatorPart)
	}

	static := value.staticContentStyleWithDeclarations(ctx, style.State{}, declarations)
	if !resolvedStyleEmpty(static.indicatorPart) {
		t.Fatalf("non-loading prepared indicator part = %#v, want unresolved", static.indicatorPart)
	}

	loading := value.Loading(true)
	loadingDeclarations := loading.declarations(active, nil)
	loadingStyle := loading.staticContentStyleWithDeclarations(
		ctx,
		style.State{Loading: true},
		loadingDeclarations,
	)
	if resolvedStyleEmpty(loadingStyle.indicatorPart) {
		t.Fatal("loading prepared indicator part was not resolved")
	}
}

func TestButtonStyleResolvesIndicatorOnlyWhileLoading(t *testing.T) {
	active := new(theme.Light())
	ctx := core.New(nil, active, locale.LanguageEnglish)
	gtx := testLayoutContext()
	value := New("").Content(nil)
	declarations := value.declarations(active, nil)

	regular := value.resolveStyleWithDeclarations(ctx, gtx, value.key, style.State{}, declarations)
	if !resolvedStyleEmpty(regular.indicatorPart) {
		t.Fatalf("regular button indicator part = %#v, want unresolved", regular.indicatorPart)
	}

	loading := value.Loading(true)
	loadingStyle := loading.resolveStyleWithDeclarations(
		ctx,
		gtx,
		loading.key,
		style.State{Loading: true},
		declarations,
	)
	if resolvedStyleEmpty(loadingStyle.indicatorPart) {
		t.Fatal("loading button indicator part was not resolved")
	}
}

func resolvedStyleEmpty(value style.ResolvedStyle) bool {
	return value.Box == nil && value.Paint == nil && value.Text == nil && value.Trans == nil && len(value.Transitions) == 0
}

func TestCloseButtonVisualUsesDefaultIconSize(t *testing.T) {
	active := theme.Light()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defaults := NewCloseButton().Key("close").styles(ctx)
	resolved := style.CascadePart(style.State{}, style.PartIcon, defaults.icon)
	if resolved.Box == nil || resolved.Box.Width == nil || *resolved.Box.Width != 16 {
		t.Fatal("close button icon does not use default size")
	}
}

func TestCloseButtonVisualCacheTracksTheme(t *testing.T) {
	first := new(theme.Light())
	second := new(theme.Light())
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := NewCloseButton().Key("close")

	value.styles(firstContext)
	value.styles(firstContext)
	value.styles(secondContext)

	stats := value.styleCache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("close button cache stats = %#v, want one hit, two misses and one theme invalidation", stats)
	}
}
