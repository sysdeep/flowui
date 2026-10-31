package badge

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestStyledPreservesBadgeBehavior(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)

	baseAnchor := &badgeProbe{size: image.Pt(48, 32)}
	baseContent := &badgeProbe{size: image.Pt(14, 14)}
	base := New(baseAnchor, "7").Content(baseContent).Alt("Unread messages")
	baseRouter := new(input.Router)
	baseDims := layoutStyledBadgeFrame(ctx, baseRouter, base, image.Pt(240, 120))

	styledAnchor := &badgeProbe{size: image.Pt(48, 32)}
	styledContent := &badgeProbe{size: image.Pt(14, 14)}
	styled := (New(styledAnchor, "7").Content(styledContent).Alt("Unread messages")).Style(style.Style{}.Radius(3))
	styledRouter := new(input.Router)
	styledDims := layoutStyledBadgeFrame(ctx, styledRouter, styled, image.Pt(240, 120))

	if styledDims != baseDims {
		t.Fatalf("styled dimensions = %v, want base dimensions %v", styledDims, baseDims)
	}
	if baseAnchor.layouts != 1 || baseContent.layouts != 1 || styledAnchor.layouts != 1 || styledContent.layouts != 1 {
		t.Fatalf("layouts = base anchor/content %d/%d, styled anchor/content %d/%d; want one each", baseAnchor.layouts, baseContent.layouts, styledAnchor.layouts, styledContent.layouts)
	}
	if !badgeSemanticTreeContains(styledRouter.AppendSemantics(nil), "Unread messages") {
		t.Fatal("styled Badge did not preserve Alt semantics")
	}
}

func TestStyledBadgeUsesCachedDefaults(t *testing.T) {
	ctx := badgeTestContext()
	value := New(nil, "7").Style(style.Style{}.Radius(3).
		Part(style.PartLabel, style.Style{}.FontSize(13)))

	value.visualWithCache(ctx)
	value.visualWithCache(ctx)
	stats := value.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("styled Badge visual cache stats = %+v, want 1 hit and 1 miss", stats)
	}

	resolved := value.resolveStyle(ctx, badgeLayoutContext(nil), value.visualWithCache(ctx))
	if resolved.root.Paint == nil || resolved.root.Paint.Radius == nil {
		t.Fatal("styled Badge root has no radius")
	}
	if got := *resolved.root.Paint.Radius; got != 3 {
		t.Fatalf("styled Badge root radius = %v, want 3", got)
	}
	if resolved.label.Text == nil || resolved.label.Text.FontSize == nil {
		t.Fatal("styled Badge label has no font size")
	}
	if got := *resolved.label.Text.FontSize; got != 13 {
		t.Fatalf("styled Badge label font size = %v, want 13", got)
	}
}

func layoutStyledBadgeFrame(ctx *core.Context, router *input.Router, value core.Widget, max image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: max},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	dims := value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
