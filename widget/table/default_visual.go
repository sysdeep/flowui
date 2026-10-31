package table

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// tableDefaultsKey contains only the inputs that change the cached table
// defaults. The instance style is resolved separately and is deliberately
// absent here.
type tableDefaultsKey struct {
	variant  Variant
	unified  bool
	bordered bool
}

// tableGeometry is the theme-derived geometry the table layout reads.
type tableGeometry struct {
	RootPadding           unit.Dp
	RootRadius            unit.Dp
	HeaderRadius          unit.Dp
	BodyRadius            unit.Dp
	HeaderHeight          unit.Dp
	RowMinHeight          unit.Dp
	EmptyHeight           unit.Dp
	MaxHeight             unit.Dp
	MinColumnWidth        unit.Dp
	CellPaddingX          unit.Dp
	CellPaddingY          unit.Dp
	HeaderTextSize        unit.Sp
	HeaderTextWeight      font.Weight
	CellTextSize          unit.Sp
	SeparatorWidth        unit.Dp
	ColumnSeparatorHeight unit.Dp
	ColumnResizerHitSize  unit.Dp
	ColumnResizerWidth    unit.Dp
	ColumnResizeStep      unit.Dp
	FocusRingWidth        unit.Dp
	FocusRadius           unit.Dp
	SelectionColumnWidth  unit.Dp
	SortIconSize          unit.Dp
	SortGap               unit.Dp
	FooterPaddingX        unit.Dp
	FooterPaddingY        unit.Dp
	LoadMoreHeight        unit.Dp
}

var defaultTableGeometry = tableGeometry{
	RootPadding: 0, RootRadius: 6, HeaderRadius: 6, BodyRadius: 6,
	HeaderHeight: 40, RowMinHeight: 36, EmptyHeight: 96, MaxHeight: 480,
	MinColumnWidth: 48, CellPaddingX: 12, CellPaddingY: 8,
	HeaderTextSize: 13, HeaderTextWeight: font.Medium, CellTextSize: 14,
	SeparatorWidth: 1, ColumnSeparatorHeight: 16, ColumnResizerHitSize: 8,
	ColumnResizerWidth: 1, ColumnResizeStep: 8, FocusRingWidth: 2, FocusRadius: 4,
	SelectionColumnWidth: 40, SortIconSize: 14, SortGap: 6,
	FooterPaddingX: 12, FooterPaddingY: 8, LoadMoreHeight: 40,
}

// tableDefaults is the theme-derived appearance and geometry of one table
// configuration. It is cached by tableDefaultsCache; the instance style is
// resolved separately because it is not one of the cache inputs.
type tableDefaults struct {
	Root      style.Style
	Header    style.Style
	Body      style.Style
	Row       style.Style
	Muted     style.Style
	Separator style.Style
	Focus     style.Style
	Resize    style.Style
	Stripe    style.Style
	Geometry  tableGeometry
	Motion    theme.MotionTheme
	SortIcon  *widget.Icon
}

func buildTableDefaults(active *theme.Theme, key tableDefaultsKey) tableDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := defaultTableGeometry
	sem := active.Semantic()
	primary := sem.Tone(theme.TonePrimary)

	rootRadius, headerRadius, bodyRadius := tableRadii(tokens, key)
	tokens.RootRadius, tokens.HeaderRadius, tokens.BodyRadius = rootRadius, headerRadius, bodyRadius
	root := style.New().
		Background(style.Solid(sem.Surface)).
		TextColor(style.Solid(sem.Foreground)).
		Radius(rootRadius)
	if key.bordered {
		root = root.
			BorderWidth(tokens.SeparatorWidth).
			BorderColor(style.Solid(sem.Separator))
	}

	row := style.New().
		TextColor(style.Solid(sem.Foreground)).
		Opacity(1).
		On(style.Hovered, style.New().Background(style.Solid(sem.SurfaceHover))).
		On(style.Pressed, style.New().Background(style.Solid(sem.SurfaceHover))).
		On(style.Selected, style.New().Background(style.Solid(primary.Soft))).
		On(style.Selected|style.Hovered, style.New().Background(style.Solid(primary.SoftHover))).
		On(style.Selected|style.Pressed, style.New().Background(style.Solid(primary.SoftHover))).
		On(style.Disabled, style.New().Opacity(active.DisabledOpacityValue()))

	return tableDefaults{
		Root: root,
		Header: style.New().
			Background(style.Solid(theme.Mix(sem.Surface, sem.SurfaceHover, 50))).
			TextColor(style.Solid(sem.MutedForeground)),
		Body: style.New().
			Background(style.Solid(sem.Surface)).
			TextColor(style.Solid(sem.Foreground)),
		Row: row,
		Muted: style.New().
			TextColor(style.Solid(sem.MutedForeground)),
		Separator: style.New().
			Background(style.Solid(sem.Separator)),
		Focus: style.New().
			Background(style.Solid(sem.Ring)),
		Resize: style.New().
			Background(style.Solid(primary.Fill)),
		Stripe: style.New().
			Background(style.Solid(sem.SurfaceSecondary)),
		Geometry: tokens,
		Motion:   active.Motion,
		SortIcon: tableSortIcon(),
	}
}

func tableRadii(tokens tableGeometry, key tableDefaultsKey) (root, header, body unit.Dp) {
	if key.variant == VariantSecondary {
		if key.unified {
			return tokens.HeaderRadius, 0, 0
		}
		return 0, tokens.HeaderRadius, 0
	}
	return tokens.RootRadius, 0, tokens.BodyRadius
}

func tableSortIcon() *widget.Icon {
	icon, _ := widget.NewIcon(lucide.ChevronUp)
	return icon
}

func tableClampRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func tableDrawBorder(gtx layout.Context, size image.Point, radius, width int, col color.NRGBA) {
	if size.X > 0 && size.Y > 0 && width > 0 && col.A != 0 {
		draw.RoundedBorder(gtx, image.Rectangle{Max: size}, radius, width, col)
	}
}

func tableDrawHeader(gtx layout.Context, size image.Point, radius, separatorWidth int, col, separator color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	clipOp := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	if separator.A != 0 {
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-separatorWidth, 0), size.X, size.Y)).Op())
	}
	clipOp.Pop()
}

func tableDrawBody(gtx layout.Context, size image.Point, radius int, col color.NRGBA) {
	if size.X > 0 && size.Y > 0 && col.A != 0 {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
	}
}

func tableDrawRow(
	gtx layout.Context,
	size image.Point,
	background color.NRGBA,
	separator color.NRGBA,
	separatorWidth int,
	showSeparator bool,
	focusColor color.NRGBA,
	focus float32,
	focusWidth, focusRadius unit.Dp,
) {
	if background.A != 0 {
		paint.FillShape(gtx.Ops, background, clip.Rect(image.Rectangle{Max: size}).Op())
	}
	if showSeparator && separator.A != 0 && size.Y > 0 {
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-separatorWidth, 0), size.X, size.Y)).Op())
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	inset := max(width/2+1, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(focusRadius), 1), min(rect.Dx(), rect.Dy())/2)
	col := focusColor
	col.A = byte(float32(col.A)*focus + .5)
	draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, radius, width, inset, col)
}

func tableDrawHeaderSeparator(gtx layout.Context, x, height, width, lineHeight int, col color.NRGBA, full bool) {
	if full {
		lineHeight = height
	}
	lineHeight = min(lineHeight, height)
	y := max((height-lineHeight)/2, 0)
	paint.FillShape(gtx.Ops, col, clip.Rect(tableSeparatorRect(x, y, width, lineHeight)).Op())
}

func tableDrawRowSeparators(gtx layout.Context, selection int, widths []int, height, width int, col color.NRGBA) {
	if height <= 0 || col.A == 0 {
		return
	}
	x := selection
	if x > 0 && len(widths) > 0 {
		paint.FillShape(gtx.Ops, col, clip.Rect(tableSeparatorRect(x, 0, width, height)).Op())
	}
	for index, columnWidth := range widths {
		x += columnWidth
		if index < len(widths)-1 {
			paint.FillShape(gtx.Ops, col, clip.Rect(tableSeparatorRect(x, 0, width, height)).Op())
		}
	}
}

// tableSeparatorRect centers a vertical separator on the resolved column
// boundary. Resizers already use the same convention; keeping grid lines
// centered prevents high-DPI widths from drifting between header and body.
func tableSeparatorRect(x, y, width, height int) image.Rectangle {
	start := x - width/2
	return image.Rect(start, y, start+width, y+height)
}

func tableDrawResizer(gtx layout.Context, x, height, baseHeight, activeWidth int, base, accent color.NRGBA, active bool, focus float32) {
	if height <= 0 {
		return
	}
	lineHeight, width, value := height, 1, accent
	if !active && focus <= 0 {
		value, lineHeight = base, min(height, baseHeight)
	} else {
		width = activeWidth
		if !active {
			value.A = byte(float32(value.A)*focus + .5)
		}
	}
	y := max((height-lineHeight)/2, 0)
	paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(x-width/2, y, x-width/2+width, y+lineHeight)).Op())
}

func tableDrawCellFocus(gtx layout.Context, size image.Point, opacity float32, width int, focusRadius unit.Dp, col color.NRGBA) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	inset := max(width/2+1, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(focusRadius), 1), min(rect.Dx(), rect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + .5)
	draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, radius, width, inset, col)
}

func tableDrawSortGlyph(gtx layout.Context, bounds image.Point, direction SortDirection, col color.NRGBA, icon *widget.Icon) {
	if icon == nil || bounds.X <= 0 || bounds.Y <= 0 {
		return
	}
	diameter := min(bounds.X, bounds.Y)
	if diameter <= 0 {
		return
	}
	center := f32.Pt(float32(bounds.X)/2, float32(bounds.Y)/2)
	angle := float32(0)
	if direction == SortDescending {
		angle = float32(math.Pi)
	}
	rotation := op.Affine(f32.AffineId().Rotate(center, angle)).Push(gtx.Ops)
	offset := op.Offset(image.Pt((bounds.X-diameter)/2, (bounds.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(iconGtx, col)
	offset.Pop()
	rotation.Pop()
}
