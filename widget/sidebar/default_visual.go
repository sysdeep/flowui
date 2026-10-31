package sidebar

import (
	"image"
	"image/color"
	"sync"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	giowidget "gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// sidebarDefaults is the theme-derived appearance and geometry of one sidebar
// configuration. It is cached by sidebarDefaultsCache; the instance style is
// resolved separately because it is not one of the cache inputs.
type sidebarDefaults struct {
	style           style.Style
	metrics         sidebarMetrics
	motion          theme.MotionTheme
	expansionEasing theme.EaseKind
	collapseEasing  theme.EaseKind
}

type sidebarMetrics struct {
	Width                 unit.Dp
	CollapsedWidth        unit.Dp
	Padding               unit.Dp
	ContentGap            unit.Dp
	ItemGap               unit.Dp
	ItemHeight            unit.Dp
	ItemRadius            unit.Dp
	ItemPaddingX          unit.Dp
	ItemContentGap        unit.Dp
	InlineIndent          unit.Dp
	SwitcherSize          unit.Dp
	ItemTextSize          unit.Sp
	SectionTextSize       unit.Sp
	SectionHeight         unit.Dp
	SectionSeparatorInset unit.Dp
	BorderWidth           unit.Dp
	FocusRingWidth        unit.Dp
}

var defaultSidebarMetrics = sidebarMetrics{
	Width: 248, CollapsedWidth: 64, Padding: 8, ContentGap: 8,
	ItemGap: 4, ItemHeight: 40, ItemRadius: 8, ItemPaddingX: 10,
	ItemContentGap: 12, InlineIndent: 20, SwitcherSize: 16,
	ItemTextSize: 14, SectionTextSize: 12, SectionHeight: 28,
	SectionSeparatorInset: 10, BorderWidth: 1, FocusRingWidth: 2,
}

// sidebarContentParts are resolved together for every visible item. Their text
// color is inherited from the resolved item part, so an item-level declaration
// still colors its label and default switcher.
var sidebarContentParts = [...]style.Part{style.PartLabel, style.PartIcon}

// metrics applies the per-instance geometry overrides to the theme tokens.
// Only the result changes the cached defaults; the instance style does not.
func (w Widget) metrics(active *theme.Theme) sidebarMetrics {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	metrics := defaultSidebarMetrics
	if w.width > 0 {
		metrics.Width = w.width
	}
	if w.collapsedWidth > 0 {
		metrics.CollapsedWidth = w.collapsedWidth
	}
	if w.hasPadding {
		metrics.Padding = w.padding
	}
	if w.hasContentGap {
		metrics.ContentGap = w.contentGap
	}
	if w.hasItemGap {
		metrics.ItemGap = w.itemGap
	}
	if w.itemHeight > 0 {
		metrics.ItemHeight = w.itemHeight
	}
	if w.hasItemPadding {
		metrics.ItemPaddingX = w.itemPaddingX
	}
	if w.hasItemContentGap {
		metrics.ItemContentGap = w.itemContentGap
	}
	if w.hasItemRadius {
		metrics.ItemRadius = w.itemRadius
	}
	if w.hasInlineIndent {
		metrics.InlineIndent = w.inlineIndent
	}
	return metrics
}

// buildSidebarDefaults derives every sidebar layer from the active theme tokens
// and semantics. The collapsed flyout is a dropdown widget that owns its own
// appearance, so its defaults are not duplicated here.
func buildSidebarDefaults(active *theme.Theme, metrics sidebarMetrics) sidebarDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	root := style.New().
		Background(style.Solid(sem.Surface)).
		TextColor(style.Solid(sem.SurfaceForeground)).
		BorderColor(style.Solid(sem.Border)).
		BorderWidth(metrics.BorderWidth)
	item := style.New().
		Background(style.RGBA(0)).
		TextColor(style.Solid(sem.Foreground)).
		Radius(metrics.ItemRadius).
		Outline(metrics.FocusRingWidth, 0, style.Solid(sem.Ring)).
		Opacity(1).
		Rules(
			style.Rule{
				When: style.Hovered,
				Set:  style.New().Background(style.Solid(sem.SurfaceTertiary)),
			},
			style.Rule{
				When: style.Pressed,
				Set:  style.New().Background(style.Solid(sem.SurfacePressed)),
			},
			style.Rule{
				When: style.Selected,
				Set: style.New().
					Background(style.Solid(tone.Soft)).
					TextColor(style.Solid(tone.OnSoft)),
			},
			style.Rule{
				When: style.Selected | style.Hovered,
				Set:  style.New().Background(style.Solid(tone.SoftHover)),
			},
			style.Rule{
				When: style.Selected | style.Pressed,
				Set:  style.New().Background(style.Solid(tone.SoftHover)),
			},
			style.Rule{
				When: style.Disabled,
				Rank: 1,
				Set:  style.New().Opacity(active.DisabledOpacityValue()),
			},
		)
	label := style.New().
		FontSize(metrics.ItemTextSize).
		FontWeight(int(font.Medium)).
		MaxLines(1)
	icon := style.New().
		Width(metrics.SwitcherSize).
		Height(metrics.SwitcherSize)
	return sidebarDefaults{
		style: style.Join(
			root,
			style.New().
				Part(style.PartItem, item).
				Part(style.PartLabel, label).
				Part(style.PartIcon, icon),
		),
		metrics:         metrics,
		motion:          active.Motion,
		expansionEasing: theme.EaseOut,
		collapseEasing:  theme.EaseInOut,
	}
}

// sidebarSwitcherGlyph caches the two chevron icons shared by every sidebar
// instance; icon parsing is expensive and identical for all of them.
type sidebarSwitcherGlyph struct {
	sync.Mutex
	icon *giowidget.Icon
}

var sidebarSwitcherGlyphs = [2]sidebarSwitcherGlyph{}

func layoutSidebarSwitcher(gtx layout.Context, expanded bool, foreground color.NRGBA, size unit.Dp) layout.Dimensions {
	index := 0
	data := lucide.ChevronRight
	if expanded {
		index = 1
		data = lucide.ChevronDown
	}
	glyph := &sidebarSwitcherGlyphs[index]
	glyph.Lock()
	defer glyph.Unlock()
	if glyph.icon == nil {
		icon, err := giowidget.NewIcon(data)
		if err != nil {
			panic(err)
		}
		glyph.icon = icon
	}
	side := max(gtx.Dp(size), 0)
	if side <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(image.Pt(side, side))
	glyph.icon.Layout(gtx, foreground)
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func sidebarSolidColor(source style.PaintSource) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

func sidebarColor(source style.PaintSource, fallback color.NRGBA) color.NRGBA {
	if value, ok := sidebarSolidColor(source); ok {
		return value
	}
	return fallback
}

func sidebarBackground(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Paint == nil {
		return fallback
	}
	return sidebarColor(resolved.Paint.Background, fallback)
}

func sidebarTextColor(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if resolved.Text == nil {
		return fallback
	}
	return sidebarColor(resolved.Text.Color, fallback)
}

func sidebarOpacity(resolved style.ResolvedStyle, fallback float32) float32 {
	if resolved.Paint == nil || resolved.Paint.Opacity == nil {
		return fallback
	}
	return *resolved.Paint.Opacity
}

func sidebarRadius(resolved style.ResolvedStyle, fallback unit.Dp) unit.Dp {
	if resolved.Paint == nil || resolved.Paint.Radius == nil {
		return fallback
	}
	return *resolved.Paint.Radius
}

func sidebarBorderWidth(resolved style.ResolvedStyle, fallback unit.Dp) unit.Dp {
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil {
		return fallback
	}
	return *resolved.Paint.Border.Width
}

func sidebarIconSize(resolved style.ResolvedStyle, fallback unit.Dp) unit.Dp {
	if resolved.Box == nil {
		return fallback
	}
	if resolved.Box.Width != nil && *resolved.Box.Width > 0 {
		return *resolved.Box.Width
	}
	if resolved.Box.Height != nil && *resolved.Box.Height > 0 {
		return *resolved.Box.Height
	}
	return fallback
}

func drawSidebarRoot(gtx layout.Context, size image.Point, active *theme.Theme, root style.ResolvedStyle) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	sem := active.Semantic()
	paint.FillShape(gtx.Ops, sidebarBackground(root, sem.Surface), clip.Rect{Max: size}.Op())
	border := sem.Border
	if root.Paint != nil && root.Paint.Border != nil {
		border = sidebarColor(root.Paint.Border.Color, border)
	}
	width := max(gtx.Dp(sidebarBorderWidth(root, 0)), 0)
	if width == 0 {
		return
	}
	borderRect := image.Rect(max(size.X-width, 0), 0, size.X, size.Y)
	paint.FillShape(gtx.Ops, border, clip.Rect(borderRect).Op())
}

func drawSidebarItem(gtx layout.Context, size image.Point, item style.ResolvedStyle, focus float32, fallbackRadius unit.Dp) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(sidebarRadius(item, fallbackRadius)), 0), min(size.X, size.Y)/2)
	if background := sidebarBackground(item, color.NRGBA{}); background.A != 0 {
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if item.Paint != nil && item.Paint.Border != nil {
		width := max(gtx.Dp(sidebarBorderWidth(item, 0)), 0)
		if width > 0 {
			draw.RoundedBorder(gtx, rect, radius, max(width, 1), sidebarColor(item.Paint.Border.Color, color.NRGBA{}))
		}
	}
	if focus <= 0 || item.Paint == nil || item.Paint.Outline == nil {
		return
	}
	width := max(gtx.Dp(item.Paint.Outline.Width), 0)
	if width == 0 {
		return
	}
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	focusColor := sidebarColor(item.Paint.Outline.Color, color.NRGBA{})
	focusColor.A = byte(float32(focusColor.A)*focus + .5)
	draw.RoundedInsetStroke(gtx, rect, max(radius-inset, 0), width, inset, focusColor)
}
