package tree

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
	"gioui.org/widget/material"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

// treeGeometry is the theme-derived geometry consumed by tree layout.
type treeGeometry struct {
	Padding              unit.Dp
	Gap                  unit.Dp
	MaxHeight            unit.Dp
	RowHeight            unit.Dp
	DescriptionRowHeight unit.Dp
	RowRadius            unit.Dp
	RowPaddingX          unit.Dp
	RowPaddingY          unit.Dp
	Indent               unit.Dp
	ChevronSlotSize      unit.Dp
	ChevronIconSize      unit.Dp
	ContentGap           unit.Dp
	ItemTextSize         unit.Sp
	ItemDescriptionSize  unit.Sp
	ItemTextWeight       font.Weight
	FocusRingWidth       unit.Dp
	SurfaceRadius        unit.Dp
	DragPreviewOffset    unit.Dp
	DragPreviewMaxWidth  unit.Dp
	DragPreviewPaddingX  unit.Dp
	DragPreviewPaddingY  unit.Dp
	DragPreviewRadius    unit.Dp
}

// treePalette is the component-owned appearance derived from the active theme.
type treePalette struct {
	RootBackground        color.NRGBA
	RootForeground        color.NRGBA
	RowHover              color.NRGBA
	RowSelected           color.NRGBA
	RowSelectedHover      color.NRGBA
	Foreground            color.NRGBA
	Description           color.NRGBA
	SelectedForeground    color.NRGBA
	SelectedDescription   color.NRGBA
	Chevron               color.NRGBA
	Focus                 color.NRGBA
	Guides                color.NRGBA
	DropIndicator         color.NRGBA
	Retry                 color.NRGBA
	ToggleBackground      color.NRGBA
	DragPreviewBackground color.NRGBA
	DragPreviewForeground color.NRGBA
	EditorSelection       color.NRGBA
}

// treeDefaults is the cached theme-derived appearance of one tree variant and
// size. Instance declarations are resolved as the last cascade layer at layout
// time and are never part of this value.
type treeDefaults struct {
	root            style.Style
	metrics         treeGeometry
	colors          treePalette
	disabledOpacity float32
	motion          theme.MotionTheme
}

// treeRowPaint is the concrete row appearance for one interaction state.
type treeRowPaint struct {
	Background  color.NRGBA
	Foreground  color.NRGBA
	Description color.NRGBA
	Chevron     color.NRGBA
	Focus       color.NRGBA
	Opacity     float32
}

func (d treeDefaults) rowPaint(selected, hovered, disabled bool) treeRowPaint {
	value := treeRowPaint{
		Foreground: d.colors.Foreground, Description: d.colors.Description,
		Chevron: d.colors.Chevron, Focus: d.colors.Focus, Opacity: 1,
	}
	if hovered {
		value.Background = d.colors.RowHover
	}
	if selected {
		value.Background = d.colors.RowSelected
		value.Foreground = d.colors.SelectedForeground
		value.Description = d.colors.SelectedDescription
		value.Chevron = d.colors.SelectedForeground
		if hovered {
			value.Background = d.colors.RowSelectedHover
		}
	}
	if disabled {
		value.Opacity = d.disabledOpacity
		value.Focus = color.NRGBA{}
	}
	return value
}

// buildTreeDefaults derives the complete tree appearance from the active theme.
// Every field is read on this call so copied themes and runtime replacement
// cannot retain stale values.
func buildTreeDefaults(active *theme.Theme, key treeVisualKey) treeDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := treeGeometry{
		Padding: 4, Gap: 2, MaxHeight: 320,
		RowHeight: 32, DescriptionRowHeight: 52, RowRadius: 4,
		RowPaddingX: 8, RowPaddingY: 5, Indent: 20,
		ChevronSlotSize: 20, ChevronIconSize: 16, ContentGap: 8,
		ItemTextSize: 14, ItemDescriptionSize: 12, FocusRingWidth: 2,
		SurfaceRadius: 8, DragPreviewOffset: 12, DragPreviewMaxWidth: 240,
		DragPreviewPaddingX: 10, DragPreviewPaddingY: 6, DragPreviewRadius: 6,
	}
	if key.Size == SizeSmall {
		tokens.Padding, tokens.Gap = 2, 1
		tokens.RowHeight, tokens.DescriptionRowHeight = 24, 40
		tokens.RowRadius, tokens.RowPaddingX, tokens.RowPaddingY = 4, 4, 2
		tokens.Indent, tokens.ChevronSlotSize, tokens.ChevronIconSize = 12, 16, 12
		tokens.ContentGap, tokens.ItemTextSize, tokens.ItemDescriptionSize = 5, 13, 11
	}
	semantic := active.Semantic()
	rootBackground := color.NRGBA{}
	rootForeground := semantic.Foreground
	if key.Variant == VariantSurface {
		rootBackground = semantic.Surface
		rootForeground = semantic.SurfaceForeground
	}
	selectedForeground := semantic.Tone(theme.TonePrimary).OnSoft
	selectedDescription := selectedForeground
	selectedDescription.A = byte(float32(selectedDescription.A)*0.78 + 0.5)
	return treeDefaults{
		root: style.New().
			FillWidth().
			Background(style.Solid(rootBackground)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		metrics: treeGeometry{
			Padding: tokens.Padding, Gap: tokens.Gap, MaxHeight: tokens.MaxHeight,
			RowHeight: tokens.RowHeight, DescriptionRowHeight: tokens.DescriptionRowHeight,
			RowRadius: tokens.RowRadius, RowPaddingX: tokens.RowPaddingX, RowPaddingY: tokens.RowPaddingY,
			Indent: tokens.Indent, ChevronSlotSize: tokens.ChevronSlotSize,
			ChevronIconSize: tokens.ChevronIconSize, ContentGap: tokens.ContentGap,
			ItemTextSize: tokens.ItemTextSize, ItemDescriptionSize: tokens.ItemDescriptionSize, ItemTextWeight: font.Medium,
			FocusRingWidth: tokens.FocusRingWidth, SurfaceRadius: tokens.SurfaceRadius,
			DragPreviewOffset: tokens.DragPreviewOffset, DragPreviewMaxWidth: tokens.DragPreviewMaxWidth,
			DragPreviewPaddingX: tokens.DragPreviewPaddingX, DragPreviewPaddingY: tokens.DragPreviewPaddingY,
			DragPreviewRadius: tokens.DragPreviewRadius,
		},
		colors: treePalette{
			RootBackground: rootBackground, RootForeground: rootForeground,
			RowHover: semantic.SurfaceTertiary, RowSelected: semantic.Tone(theme.TonePrimary).Soft,
			RowSelectedHover: semantic.Tone(theme.TonePrimary).SoftHover,
			Foreground:       semantic.Foreground, Description: semantic.MutedForeground,
			SelectedForeground: selectedForeground, SelectedDescription: selectedDescription,
			Chevron: semantic.MutedForeground, Focus: semantic.Ring, Guides: semantic.MutedForeground,
			DropIndicator: semantic.Tone(theme.TonePrimary).Fill,
			Retry:         semantic.Tone(theme.ToneDanger).Fill, ToggleBackground: semantic.SurfaceTertiary,
			DragPreviewBackground: semantic.Overlay, DragPreviewForeground: semantic.OverlayForeground,
			EditorSelection: semantic.Selection,
		},
		disabledOpacity: active.DisabledOpacityValue(),
		motion:          active.Motion,
	}
}

func treeClampRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func treeDrawRoot(active *theme.Theme, gtx layout.Context, size image.Point, radiusDp unit.Dp, col color.NRGBA, shadow bool) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	radius := treeClampRadius(gtx, size, radiusDp)
	if shadow {
		sem := active.Semantic()
		draw.Shadow(gtx, image.Rectangle{Max: size}, draw.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp), draw.ThemeShadow(active.Shadows.Surface, sem.SurfaceShadow, 1))
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func treeDrawRow(gtx layout.Context, size image.Point, background, focusColor color.NRGBA, focus float32, radius, focusWidth unit.Dp) {
	rect := image.Rectangle{Max: size}
	rowRadius := min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
	if background.A != 0 {
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, rowRadius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := focusColor
	col.A = byte(float32(col.A)*focus + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, max(rowRadius-inset, 0)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func treeDrawDropIndicator(gtx layout.Context, size image.Point, position DropPosition, depth int, metrics treeGeometry, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(1)), 1)
	if position == DropInside {
		inset := max(width/2+1, 1)
		rect := image.Rectangle{Max: size}.Inset(inset)
		if rect.Empty() {
			return
		}
		radius := min(max(gtx.Dp(metrics.RowRadius)-inset, 0), min(rect.Dx(), rect.Dy())/2)
		stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		stroke.Pop()
		return
	}
	y := 0
	if position == DropAfter {
		y = size.Y - width
	}
	x := max(gtx.Dp(metrics.RowPaddingX+unit.Dp(depth)*metrics.Indent), 0)
	rect := image.Rect(min(x, size.X), max(y, 0), size.X, min(max(y+width, 0), size.Y))
	if !rect.Empty() {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, width/2).Op(gtx.Ops))
	}
}

func treeDrawToggle(gtx layout.Context, size image.Point, metrics treeGeometry, expansion float32, connectors bool, foreground, background color.NRGBA) {
	if connectors {
		treeDrawConnectorToggle(gtx, size, metrics, expansion, foreground, background)
		return
	}
	diameter := min(gtx.Dp(metrics.ChevronIconSize), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	rotation := op.Affine(f32.AffineId().Rotate(center, expansion*float32(math.Pi/2))).Push(gtx.Ops)
	offset := op.Offset(position).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronRight, iconGtx, foreground)
	offset.Pop()
	rotation.Pop()
}

func treeDrawRetry(gtx layout.Context, size image.Point, metrics treeGeometry, col color.NRGBA) {
	iconSize := min(gtx.Dp(metrics.ChevronIconSize), min(size.X, size.Y))
	if iconSize <= 0 {
		return
	}
	stack := op.Offset(image.Pt((size.X-iconSize)/2, (size.Y-iconSize)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
	icon.Layout(lucide.RefreshCw, iconGtx, col)
	stack.Pop()
}

func treeDrawConnectorToggle(gtx layout.Context, size image.Point, metrics treeGeometry, expansion float32, foreground, background color.NRGBA) {
	diameterDp := max(metrics.ChevronIconSize, unit.Dp(14))
	diameter := min(gtx.Dp(diameterDp), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	if background.A == 0 {
		background = foreground
		background.A = 32
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	offset := op.Offset(position).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(image.Rect(0, 0, diameter, diameter), max(gtx.Dp(unit.Dp(3)), 1)).Op(gtx.Ops))
	thickness := max(gtx.Dp(unit.Dp(1)), 1)
	horizontal, vertical := treeConnectorToggleBars(diameter, thickness)
	paint.FillShape(gtx.Ops, foreground, clip.Rect(horizontal).Op())
	if expansion < 1 {
		opacity := paint.PushOpacity(gtx.Ops, max(1-expansion, 0))
		paint.FillShape(gtx.Ops, foreground, clip.Rect(vertical).Op())
		opacity.Pop()
	}
	offset.Pop()
}

func treeDrawGuides(gtx layout.Context, segments []treeGuideSegment, guideStyle GuideStyle, col color.NRGBA, gap unit.Dp) {
	if len(segments) == 0 || col.A == 0 {
		return
	}
	col.A = byte(float32(col.A)*0.5 + 0.5)
	width := max(gtx.Dp(unit.Dp(1)), 1)
	for _, segment := range segments {
		to := segment.to
		if segment.extend && segment.from.X == to.X {
			to.Y += gtx.Dp(gap)
		}
		treeDrawGuideSegment(gtx, segment.from, to, width, guideStyle, col)
	}
}

func treeDrawGuideSegment(gtx layout.Context, from, to image.Point, width int, guideStyle GuideStyle, col color.NRGBA) {
	if guideStyle == GuideSolid {
		var rect image.Rectangle
		switch {
		case from.X == to.X && from.Y < to.Y:
			rect = image.Rect(from.X-width/2, from.Y, from.X+(width+1)/2, to.Y)
		case from.Y == to.Y && from.X < to.X:
			rect = image.Rect(from.X, from.Y-width/2, to.X, from.Y+(width+1)/2)
		default:
			return
		}
		paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
		return
	}
	dash := max(gtx.Dp(unit.Dp(1)), 1)
	period := max(gtx.Dp(unit.Dp(3)), dash+1)
	if from.X == to.X && from.Y < to.Y {
		for _, part := range treeGuideDashes(from.Y, to.Y, dash, period) {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(from.X-width/2, part[0], from.X+(width+1)/2, part[1])).Op())
		}
	} else if from.Y == to.Y && from.X < to.X {
		for _, part := range treeGuideDashes(from.X, to.X, dash, period) {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(part[0], from.Y-width/2, part[1], from.Y+(width+1)/2)).Op())
		}
	}
}

func treeDrawText(active *theme.Theme, gtx layout.Context, value string, size unit.Sp, col color.NRGBA, weight font.Weight) layout.Dimensions {
	label := material.Label(theme.MaterialOf(active), size, value)
	label.Color = col
	label.Font.Weight = weight
	label.MaxLines = 1
	label.Truncator = "..."
	return label.Layout(gtx)
}

func treeDragPreview(active *theme.Theme, gtx layout.Context, label string, defaults treeDefaults) layout.Dimensions {
	metrics := defaults.metrics
	opacity := paint.PushOpacity(gtx.Ops, 0.5)
	defer opacity.Pop()
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(metrics.DragPreviewMaxWidth))
	gtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y, gtx.Dp(metrics.RowHeight+unit.Dp(8)))
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: metrics.DragPreviewPaddingY, Right: metrics.DragPreviewPaddingX, Bottom: metrics.DragPreviewPaddingY, Left: metrics.DragPreviewPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return treeDrawText(active, gtx, label, metrics.ItemTextSize, active.Semantic().OverlayForeground, font.Medium)
	})
	content := macro.Stop()
	sem := active.Semantic()
	radius := min(max(gtx.Dp(metrics.DragPreviewRadius), 0), min(dims.Size.X, dims.Size.Y)/2)
	draw.Surface(gtx, image.Rectangle{Max: dims.Size}, radius, sem.Overlay, draw.ThemeShadow(active.Shadows.Overlay, sem.OverlayShadow, 1))
	content.Add(gtx.Ops)
	return dims
}
