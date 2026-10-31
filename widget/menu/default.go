package menu

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/internal/textrender"
)

const menuItemPressDuration = 250 * time.Millisecond

// menuDefaults is the theme-derived appearance and geometry of one menu
// density. It is cached by menuDefaultsCache; the instance style is resolved
// separately because it is not one of the cache inputs.
type menuDefaults struct {
	Root      style.Style
	Items     [2][2]menuItemDefaults
	Metrics   menuMetrics
	Colors    menuColors
	Motion    theme.MotionTheme
	TextWidth *menuTextWidthCache
}

// menuMetrics is the geometry and timing the menu layout reads.
type menuMetrics struct {
	Width                      unit.Dp
	MaxHeight                  unit.Dp
	MaxWidthFraction           float32
	Padding                    unit.Dp
	Radius                     unit.Dp
	BorderWidth                unit.Dp
	ItemGap                    unit.Dp
	ItemMinHeight              unit.Dp
	ItemRadius                 unit.Dp
	ItemPaddingX               unit.Dp
	ItemPaddingY               unit.Dp
	ItemContentGap             unit.Dp
	ItemTextSize               unit.Sp
	ItemDescriptionSize        unit.Sp
	ShortcutTextSize           unit.Sp
	ShortcutHeight             unit.Dp
	ShortcutPaddingX           unit.Dp
	IndicatorSize              unit.Dp
	IndicatorContentGap        unit.Dp
	CheckmarkSize              unit.Dp
	RadioDotSize               unit.Dp
	IndicatorOffsetY           unit.Dp
	SubmenuIndicatorSize       unit.Dp
	FocusRingWidth             unit.Dp
	FocusRingOffset            unit.Dp
	PressedScale               float32
	SectionTextSize            unit.Sp
	SectionPaddingX            unit.Dp
	SectionPaddingTop          unit.Dp
	SectionPaddingBottom       unit.Dp
	SeparatorMarginX           unit.Dp
	SeparatorMarginY           unit.Dp
	SeparatorWidth             unit.Dp
	DescriptionLeadingHeight   unit.Dp
	DescriptionLeadingInsetTop unit.Dp
	SubmenuGap                 unit.Dp
	ContextMenuOffset          unit.Dp
	EnterScale                 float32
	ExitScale                  float32
	EnterDuration              time.Duration
	ExitDuration               time.Duration
	SubmenuOpenDelay           time.Duration
	AnimationDistance          unit.Dp
}

// menuColors is the theme-derived palette every menu item and panel reads.
type menuColors struct {
	PanelBackground color.NRGBA
	PanelForeground color.NRGBA
	PanelBorder     color.NRGBA
	Separator       color.NRGBA
	Foreground      color.NRGBA
	Description     color.NRGBA
	Shortcut        color.NRGBA
	Indicator       color.NRGBA
	Danger          color.NRGBA
}

// menuItemDefaults is the declaration and fallback colors of one item variant in
// one disabled state.
type menuItemDefaults struct {
	Defaults style.Style
	Variant  style.Style
	Colors   menuItemColors
}

// menuItemColors holds the fallback colors of the text and leading slots of an
// item. The resolved item part can still override its foreground.
type menuItemColors struct {
	Foreground  color.NRGBA
	Description color.NRGBA
	Shortcut    color.NRGBA
	Indicator   color.NRGBA
}

// resolvedColors applies the resolved item part to the variant fallbacks.
func (i menuItemDefaults) resolvedColors(part style.ResolvedStyle) menuItemColors {
	result := i.Colors
	if part.Text != nil {
		value := textrender.ColorOr(part, result.Foreground)
		result.Foreground = value
		result.Description = value
		result.Shortcut = value
	}
	return result
}

// menuTextSpec is the semantic text, its role, the fallback color, and the
// resolved item part whose typography it inherits.
type menuTextSpec struct {
	Text  string
	Role  MenuTextRole
	Color color.NRGBA
	Part  style.ResolvedStyle
}

// item returns the precomputed item declaration and colors for a request.
func (d menuDefaults) item(variant ItemVariant, disabled bool) menuItemDefaults {
	variantIndex := 0
	if variant == ItemDanger {
		variantIndex = 1
	}
	disabledIndex := 0
	if disabled {
		disabledIndex = 1
	}
	return d.Items[variantIndex][disabledIndex]
}

// buildMenu derives one menu density from the active theme.
func buildMenu(active *theme.Theme, compact bool) menuDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := defaultMenuMetrics
	if compact {
		tokens = menuCompact(tokens)
	}
	colors := menuColorsOf(active)
	items := [2][2]menuItemDefaults{}
	for variantIndex := range 2 {
		variant := ItemDefault
		if variantIndex == 1 {
			variant = ItemDanger
		}
		for disabledIndex := range 2 {
			items[variantIndex][disabledIndex] = menuItemDefaultsOf(active, tokens, variant, disabledIndex == 1)
		}
	}
	return menuDefaults{
		Root:      menuRoot(tokens, colors),
		Items:     items,
		Metrics:   tokens,
		Colors:    colors,
		Motion:    active.Motion,
		TextWidth: &menuTextWidthCache{},
	}
}

var defaultMenuMetrics = menuMetrics{
	Width: 220, MaxWidthFraction: .48, Padding: 8, Radius: 8, BorderWidth: 1,
	ItemMinHeight: 32, ItemRadius: 4, ItemPaddingX: 12, ItemPaddingY: 5,
	ItemContentGap: 8, ItemTextSize: 14, ItemDescriptionSize: 12,
	ShortcutTextSize: 14, ShortcutHeight: 20, ShortcutPaddingX: 8,
	IndicatorSize: 12, CheckmarkSize: 10, RadioDotSize: 6,
	SubmenuIndicatorSize: 12, SectionTextSize: 14, SectionPaddingX: 12,
	SectionPaddingTop: 8, SectionPaddingBottom: 4, SeparatorMarginX: 12,
	SeparatorMarginY: 8, SeparatorWidth: 1, DescriptionLeadingHeight: 32,
	DescriptionLeadingInsetTop: 1, SubmenuGap: 8, ContextMenuOffset: 2,
	EnterScale: 1, ExitScale: 1, EnterDuration: 150 * time.Millisecond,
	ExitDuration: 100 * time.Millisecond, SubmenuOpenDelay: 100 * time.Millisecond,
	AnimationDistance: 4,
}

func menuCompact(tokens menuMetrics) menuMetrics {
	tokens.Padding = menuMinDp(tokens.Padding, 4)
	tokens.Radius = menuMinDp(tokens.Radius, 8)
	tokens.ItemGap = menuMinDp(tokens.ItemGap, 1)
	tokens.ItemMinHeight = menuMinDp(tokens.ItemMinHeight, 28)
	tokens.ItemRadius = menuMinDp(tokens.ItemRadius, 4)
	tokens.ItemPaddingX = menuMinDp(tokens.ItemPaddingX, 8)
	tokens.ItemPaddingY = menuMinDp(tokens.ItemPaddingY, 3)
	tokens.ItemContentGap = menuMinDp(tokens.ItemContentGap, 8)
	tokens.ItemTextSize = menuMinSp(tokens.ItemTextSize, 13)
	tokens.ItemDescriptionSize = menuMinSp(tokens.ItemDescriptionSize, 11)
	tokens.ShortcutTextSize = menuMinSp(tokens.ShortcutTextSize, 12)
	tokens.ShortcutHeight = menuMinDp(tokens.ShortcutHeight, 20)
	tokens.ShortcutPaddingX = menuMinDp(tokens.ShortcutPaddingX, 4)
	tokens.IndicatorSize = menuMinDp(tokens.IndicatorSize, 14)
	tokens.CheckmarkSize = menuMinDp(tokens.CheckmarkSize, 9)
	tokens.RadioDotSize = menuMinDp(tokens.RadioDotSize, 6)
	tokens.SubmenuIndicatorSize = menuMinDp(tokens.SubmenuIndicatorSize, 12)
	tokens.SectionTextSize = menuMinSp(tokens.SectionTextSize, 11)
	tokens.SectionPaddingX = menuMinDp(tokens.SectionPaddingX, 6)
	tokens.SectionPaddingTop = menuMinDp(tokens.SectionPaddingTop, 4)
	tokens.SectionPaddingBottom = menuMinDp(tokens.SectionPaddingBottom, 3)
	return tokens
}

func menuMinDp(value, limit unit.Dp) unit.Dp {
	return unit.Dp(math.Min(float64(value), float64(limit)))
}

func menuMinSp(value, limit unit.Sp) unit.Sp {
	return unit.Sp(math.Min(float64(value), float64(limit)))
}

func menuColorsOf(active *theme.Theme) menuColors {
	sem := active.Semantic()
	return menuColors{
		PanelBackground: sem.Overlay,
		PanelForeground: sem.OverlayForeground,
		PanelBorder:     sem.Border,
		Separator:       sem.Border,
		Foreground:      sem.OverlayForeground,
		Description:     sem.MutedForeground,
		Shortcut:        sem.MutedForeground,
		Indicator:       sem.OverlayForeground,
		Danger:          sem.Tone(theme.ToneDanger).OnSoft,
	}
}

func menuRoot(tokens menuMetrics, colors menuColors) style.Style {
	return style.Style{}.
		Background(style.Solid(colors.PanelBackground)).
		TextColor(style.Solid(colors.PanelForeground)).
		BorderColor(style.Solid(colors.PanelBorder)).
		BorderWidth(tokens.BorderWidth).
		Radius(tokens.Radius).
		Shadow(style.ShadowMenu).
		Overflow(style.OverflowHidden)
}

func menuItemDefaultsOf(active *theme.Theme, tokens menuMetrics, variant ItemVariant, disabled bool) menuItemDefaults {
	sem := active.Semantic()
	colors := menuColorsOf(active)
	foreground := colors.Foreground
	indicator := colors.Indicator
	if variant == ItemDanger {
		foreground = colors.Danger
		indicator = colors.Danger
	}
	if disabled {
		foreground = active.DisabledColor(foreground)
		indicator = active.DisabledColor(indicator)
	}
	fill, fillText := menuItemHighlight(active, tokens, variant)
	highlighted := style.Style{}.
		Background(style.Solid(fill)).
		TextColor(style.Solid(fillText))
	base := style.Style{}.
		MinHeight(tokens.ItemMinHeight).
		Padding(tokens.ItemPaddingY, tokens.ItemPaddingX).
		Radius(tokens.ItemRadius).
		Background(style.RGBA(0x00000000)).
		TextColor(style.Solid(foreground)).
		BorderWidth(tokens.FocusRingWidth).
		BorderColor(style.WithAlpha(sem.Ring, 0)).
		Scale(1).
		Transition(style.PropTransform, menuItemPressDuration)
	// The state rules live in the variant layer: the menu's text is inherited
	// into an item after its defaults, and would otherwise keep the highlighted
	// text dark.
	states := style.Style{}.
		Part(style.PartItem, style.Style{}.TextColor(style.Solid(foreground))).
		Rules(
			style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.FocusVisible, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.Open, Rank: -1, Set: style.Style{}.Part(style.PartItem, style.Style{}.Background(style.Solid(sem.HoverSurface())))},
			style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.Scale(menuPressedScale(tokens.PressedScale)))},
			style.Rule{When: style.Selected, Set: style.Style{}.Part(style.PartItem, style.Style{}.Background(menuSelectedBackground(active, tokens)))},
			style.Rule{When: style.Selected | style.Hovered, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.Selected | style.FocusVisible, Set: style.Style{}.Part(style.PartItem, highlighted)},
		)
	return menuItemDefaults{
		Defaults: style.Style{}.Part(style.PartItem, base),
		Variant:  states,
		Colors: menuItemColors{
			Foreground:  foreground,
			Description: colors.Description,
			Shortcut:    colors.Shortcut,
			Indicator:   indicator,
		},
	}
}

// menuItemHighlight returns the fill and the text color of a highlighted item.
// The highlighted item is the solid accent; a danger item uses the danger tone,
// and an explicit HoverColor token wins.
func menuItemHighlight(active *theme.Theme, tokens menuMetrics, variant ItemVariant) (fill, text color.NRGBA) {
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	if variant == ItemDanger {
		tone = sem.Tone(theme.ToneDanger)
	}
	fill, text = tone.Fill, tone.OnFill
	// Mid-dark fills pair with white text. The shared luminance rule picks black
	// for red, which reads as a tinted row rather than a destructive action, so
	// prefer white when it stays legible.
	if theme.ContrastRatio(fill, menuWhite) >= 3 {
		text = menuWhite
	}
	return fill, text
}

var menuWhite = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

// label builds the component-owned text primitive for one semantic slot. The
// resolved item part supplies typography when the instance style overrides it.
func (d menuDefaults) label(active *theme.Theme, spec menuTextSpec) material.LabelStyle {
	size := d.Metrics.ItemTextSize
	weight := font.Normal
	switch spec.Role {
	case MenuTextEmpty:
	case MenuTextSection:
		size = d.Metrics.SectionTextSize
	case MenuTextItem:
		weight = font.Medium
	case MenuTextDescription:
		size = d.Metrics.ItemDescriptionSize
	case MenuTextShortcut:
		size = d.Metrics.ShortcutTextSize
		weight = font.Medium
	default:
		panic("flowui/menu: unsupported text role")
	}
	textStyle := spec.Part.Text
	if textStyle != nil {
		if textStyle.FontSize != nil && *textStyle.FontSize > 0 {
			size = *textStyle.FontSize
		}
		if spec.Role == MenuTextItem && textStyle.FontWeight != nil {
			weight = font.Weight(*textStyle.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, spec.Text)
	label.Color = textrender.ColorOr(spec.Part, spec.Color)
	label.SelectionColor = active.Semantic().Selection
	label.Font.Weight = weight
	if textStyle != nil {
		if textStyle.Typeface != nil {
			label.Font.Typeface = *textStyle.Typeface
		}
		if textStyle.FontStyle != nil {
			label.Font.Style = *textStyle.FontStyle
		}
		if textStyle.LineHeight != nil {
			label.LineHeight = *textStyle.LineHeight
		}
		if textStyle.LineHeightScale != nil {
			label.LineHeightScale = *textStyle.LineHeightScale
		}
	}
	return label
}

func (d menuDefaults) layoutText(active *theme.Theme, gtx layout.Context, spec menuTextSpec) layout.Dimensions {
	return d.label(active, spec).Layout(gtx)
}

// textWidth measures the same text primitive the menu paints, without painting
// it. The cache lives on the cached defaults, so a theme or density change
// invalidates it together with the rest of the appearance.
func (d menuDefaults) textWidth(active *theme.Theme, gtx layout.Context, spec menuTextSpec) int {
	if spec.Text == "" {
		return 0
	}
	label := d.label(active, spec)
	key := menuTextWidthKey{
		Text:            spec.Text,
		Role:            spec.Role,
		Constraints:     gtx.Constraints,
		Metric:          gtx.Metric,
		Locale:          gtx.Locale,
		Font:            label.Font,
		Alignment:       label.Alignment,
		MaxLines:        label.MaxLines,
		WrapPolicy:      label.WrapPolicy,
		Truncator:       label.Truncator,
		TextSize:        label.TextSize,
		LineHeight:      label.LineHeight,
		LineHeightScale: label.LineHeightScale,
	}
	cache := d.TextWidth
	if cache == nil {
		cache = &menuTextWidthCache{}
	}
	if width, ok := cache.lookup(key); ok {
		return width
	}
	measureGtx := gtx
	measureGtx.Constraints = layout.Constraints{
		Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024))),
	}
	measureGtx.Ops = new(op.Ops)
	width := label.Layout(measureGtx).Size.X
	cache.store(key, width)
	return width
}

const menuTextWidthCacheCapacity = 32

type menuTextWidthKey struct {
	Text            string
	Role            MenuTextRole
	Constraints     layout.Constraints
	Metric          unit.Metric
	Locale          system.Locale
	Font            font.Font
	Alignment       giotext.Alignment
	MaxLines        int
	WrapPolicy      giotext.WrapPolicy
	Truncator       string
	TextSize        unit.Sp
	LineHeight      unit.Sp
	LineHeightScale float32
}

type menuTextWidthEntry struct {
	key   menuTextWidthKey
	width int
	valid bool
}

type menuTextWidthCache struct {
	entries [menuTextWidthCacheCapacity]menuTextWidthEntry
	next    uint8
}

func (c *menuTextWidthCache) lookup(key menuTextWidthKey) (int, bool) {
	for _, entry := range c.entries {
		if entry.valid && entry.key == key {
			return entry.width, true
		}
	}
	return 0, false
}

func (c *menuTextWidthCache) store(key menuTextWidthKey, width int) {
	index := int(c.next) % len(c.entries)
	c.entries[index] = menuTextWidthEntry{key: key, width: width, valid: true}
	c.next = uint8((index + 1) % len(c.entries))
}

func menuSelectedBackground(active *theme.Theme, tokens menuMetrics) style.SolidColor {
	return style.Solid(active.Semantic().Tone(theme.TonePrimary).Soft)
}

func menuPressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .98
	}
	return value
}

func drawMenuSeparator(gtx layout.Context, size image.Point, foreground color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || foreground.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, foreground, clip.Rect{Max: size}.Op())
}

// menuIndicatorSpec contains the facts needed to paint a checkmark or radio dot.
type menuIndicatorSpec struct {
	Kind          IndicatorType
	Size          image.Point
	DotSize       int
	CheckmarkSize int
	Color         color.NRGBA
}

func drawMenuIndicator(gtx layout.Context, spec menuIndicatorSpec) {
	if spec.Size.X <= 0 || spec.Size.Y <= 0 || spec.Color.A == 0 {
		return
	}
	if spec.Kind == IndicatorDot {
		dot := min(max(spec.DotSize, 1), min(spec.Size.X, spec.Size.Y))
		rect := image.Rect((spec.Size.X-dot)/2, (spec.Size.Y-dot)/2, (spec.Size.X+dot)/2, (spec.Size.Y+dot)/2)
		paint.FillShape(gtx.Ops, spec.Color, clip.Ellipse(rect).Op(gtx.Ops))
		return
	}
	if spec.Kind != IndicatorCheckmark {
		return
	}
	check := min(max(spec.CheckmarkSize, 1), min(spec.Size.X, spec.Size.Y))
	offset := (spec.Size.X - check) / 2
	points := [3]f32.Point{
		f32.Pt(float32(offset)+float32(check)*.05, float32(spec.Size.Y)*.56),
		f32.Pt(float32(offset)+float32(check)*.40, float32(spec.Size.Y)*.86),
		f32.Pt(float32(offset)+float32(check)*.95, float32(spec.Size.Y)*.14),
	}
	path := draw.CheckPath(gtx.Ops, points, 1)
	stroke := clip.Stroke{Path: path, Width: max(draw.DpFloat(gtx, unit.Dp(1.5)), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, spec.Color)
	stroke.Pop()
}

func drawMenuSubmenuIndicator(gtx layout.Context, size image.Point, foreground color.NRGBA) layout.Dimensions {
	if size.X <= 0 || size.Y <= 0 || foreground.A == 0 {
		return layout.Dimensions{Size: size}
	}
	width := float32(size.X)
	height := float32(size.Y)
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(width*.35, height*.18))
	path.LineTo(f32.Pt(width*.68, height*.5))
	path.LineTo(f32.Pt(width*.35, height*.82))
	pathSpec := path.End()
	stroke := clip.Stroke{Path: pathSpec, Width: max(draw.DpFloat(gtx, unit.Dp(1.5)), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
	return layout.Dimensions{Size: size}
}
