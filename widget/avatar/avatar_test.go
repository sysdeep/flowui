package avatar

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type avatarStyle struct {
	background, foreground     color.NRGBA
	diameter, radius, iconSize unit.Dp
	textSize                   unit.Sp
}

func avatarStyleFor(activeTheme *theme.Theme, c Color, v Variant, s Size) avatarStyle {
	visual := buildAvatarVisual(activeTheme, c, v, s)
	variant := style.Cascade(style.State{}, visual.Variant)
	size := style.Cascade(style.State{}, visual.Size)
	result := avatarStyle{}
	if variant.Paint != nil {
		if value, ok := variant.Paint.Background.(style.SolidColor); ok {
			result.background = value.Color
		}
	}
	if variant.Text != nil {
		if value, ok := variant.Text.Color.(style.SolidColor); ok {
			result.foreground = value.Color
		}
	}
	if size.Box != nil && size.Box.Width != nil {
		result.diameter = *size.Box.Width
	}
	if size.Paint != nil && size.Paint.Radius != nil {
		result.radius = *size.Paint.Radius
	}
	if size.Text != nil && size.Text.FontSize != nil {
		result.textSize = *size.Text.FontSize
	}
	part := style.CascadePart(style.State{}, style.PartIcon, visual.Defaults, visual.Variant, visual.Size)
	if part.Box != nil && part.Box.Width != nil {
		result.iconSize = *part.Box.Width
	}
	return result
}

type avatarProbe struct {
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *avatarProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(16, 16))}
}

func TestAvatarSizesStaySquare(t *testing.T) {
	ctx := avatarTestContext()
	for _, test := range []struct {
		size Size
		want int
	}{{SizeSmall, 32}, {SizeMedium, 40}, {SizeLarge, 48}} {
		gtx := avatarLayoutContext(image.Pt(100, 100))
		if dims := New("AB").Size(test.size).Layout(ctx, gtx); dims.Size != image.Pt(test.want, test.want) {
			t.Fatalf("Avatar size %d = %v", test.size, dims.Size)
		}
	}
	gtx := avatarLayoutContext(image.Pt(24, 60))
	if dims := New("AB").Layout(ctx, gtx); dims.Size != image.Pt(24, 24) {
		t.Fatalf("constrained Avatar = %v", dims.Size)
	}
}

func TestAvatarCustomFallbackReceivesSemanticColors(t *testing.T) {
	probe := new(avatarProbe)
	activeTheme := theme.Light()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	New("").Fallback(probe).Color(ColorSuccess).Variant(VariantSoft).Layout(ctx, avatarLayoutContext(image.Pt(100, 100)))
	style := avatarStyleFor(&activeTheme, ColorSuccess, VariantSoft, SizeMedium)
	if probe.layouts != 1 || probe.foreground != style.foreground || probe.background != style.background {
		t.Fatalf("fallback context = layouts %d foreground %v background %v", probe.layouts, probe.foreground, probe.background)
	}
}

func TestAvatarImageSuppressesFallback(t *testing.T) {
	probe := new(avatarProbe)
	img := image.NewNRGBA(image.Rect(0, 0, 80, 40))
	dims := New("").Fallback(probe).Image(paint.NewImageOp(img)).Layout(avatarTestContext(), avatarLayoutContext(image.Pt(100, 100)))
	if dims.Size != image.Pt(40, 40) || probe.layouts != 0 {
		t.Fatalf("image Avatar = size %v fallback layouts %d", dims.Size, probe.layouts)
	}
}

func TestAvatarExposesAccessibleLabel(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(100, 100)}}
	New("AM").Alt("Alex Morgan").Layout(avatarTestContext(), gtx)
	router.Frame(&ops)
	if !avatarSemanticTreeContains(router.AppendSemantics(nil), "Alex Morgan") {
		t.Fatal("Avatar semantics did not expose alt text")
	}
}

func TestAvatarSoftColorsUseSemanticPalette(t *testing.T) {
	activeTheme := theme.Light()
	tests := []struct {
		color Color
		bg    color.NRGBA
		fg    color.NRGBA
	}{
		{ColorAccent, activeTheme.Colors.Tones[theme.TonePrimary].Soft, activeTheme.Colors.Tones[theme.TonePrimary].OnSoft},
		{ColorSuccess, activeTheme.Colors.Tones[theme.ToneSuccess].Soft, activeTheme.Colors.Tones[theme.ToneSuccess].OnSoft},
		{ColorWarning, activeTheme.Colors.Tones[theme.ToneWarning].Soft, activeTheme.Colors.Tones[theme.ToneWarning].OnSoft},
		{ColorDanger, activeTheme.Colors.Tones[theme.ToneDanger].Soft, activeTheme.Colors.Tones[theme.ToneDanger].OnSoft},
	}
	for _, test := range tests {
		style := avatarStyleFor(&activeTheme, test.color, VariantSoft, SizeMedium)
		if style.background != test.bg || style.foreground != test.fg {
			t.Fatalf("Avatar color %d = background %v foreground %v", test.color, style.background, style.foreground)
		}
	}
}

func TestAvatarStyleResolvesAsLastLayer(t *testing.T) {
	probe := new(avatarProbe)
	background := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	foreground := color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}
	New("").Fallback(probe).Color(ColorAccent).Style(
		style.New().Background(style.Solid(background)).TextColor(style.Solid(foreground)),
	).Layout(avatarTestContext(), avatarLayoutContext(image.Pt(100, 100)))
	if probe.background != background || probe.foreground != foreground {
		t.Fatalf("instance style = background %v foreground %v", probe.background, probe.foreground)
	}
}

func TestAvatarStyleStaysOutsideDefaultsCache(t *testing.T) {
	ctx := avatarTestContext()
	value := New("AB").Style(style.New().Background(style.Solid(color.NRGBA{R: 1, A: 0xff})))
	value.visualWithCache(ctx)
	if stats := value.visualCache.Stats(); stats.Misses != 1 {
		t.Fatalf("styled first resolve stats = %+v", stats)
	}
	value.visualWithCache(ctx)
	if stats := value.visualCache.Stats(); stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("styled second resolve stats = %+v", stats)
	}
}

func avatarTestContext() *core.Context {
	return core.New(nil, new(theme.Light()), locale.LanguageEnglish)
}

func avatarLayoutContext(maximum image.Point) layout.Context {
	return layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: maximum}}
}

func avatarSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || avatarSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
