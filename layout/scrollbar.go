package layout

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotScrollbarBar = "bar"

type scrollbarMetrics struct {
	TrackWidth     unit.Dp
	ThumbWidth     unit.Dp
	ContentGap     unit.Dp
	MinThumbLength unit.Dp
	MajorPadding   unit.Dp
	Radius         unit.Dp
	ThumbOpacity   float32
	HoverOpacity   float32
}

var defaultScrollbarMetrics = scrollbarMetrics{
	TrackWidth: 10, ThumbWidth: 6, ContentGap: 4, MinThumbLength: 32,
	MajorPadding: 2, Radius: 3, ThumbOpacity: .15, HoverOpacity: .28,
}

type scrollbarState struct {
	bar           widget.Scrollbar
	defaultsCache computed.VisualCache[scrollbarDeclarationKey, style.Style]
}

func derivedScrollbarState(ctx *Context, owner string) *scrollbarState {
	key := core.ClaimDerivedKey(ctx, core.KindScrollbar, owner, "bar")
	return core.UseState[scrollbarState](ctx, key, stateSlotScrollbarBar)
}

func layoutScrollbarList(ctx *Context, gtx layout.Context, key string, list *layout.List, bar *widget.Scrollbar, visual *VisualOutsetState, count int, disabled, overlay bool, custom style.Style, defaultsCache *computed.VisualCache[scrollbarDeclarationKey, style.Style], showBar bool, item layout.ListElement) layout.Dimensions {
	if disabled {
		gtx = gtx.Disabled()
	}
	if !showBar {
		return layoutVisualOutsetList(ctx, gtx, list, count, visual, item)
	}
	disabled = disabled || !gtx.Enabled()
	style := resolvedScrollbarStyle(ctx, gtx, key, bar, disabled, custom, defaultsCache)
	axis := list.Axis
	originalConstraints := gtx.Constraints
	barWidth := gtx.Dp(style.Width())
	reservedWidth := barWidth + gtx.Dp(max(defaultScrollbarMetrics.ContentGap, 0))
	majorAxisLimit := axis.Convert(originalConstraints.Max).X
	// ponytail: overflow changes settle on the next frame to avoid laying out interactive children twice.
	reserveScrollbar := !overlay && count > 0 && list.Position.Length > majorAxisLimit
	if reserveScrollbar {
		maxConstraints := axis.Convert(gtx.Constraints.Max)
		minConstraints := axis.Convert(gtx.Constraints.Min)
		maxConstraints.Y = max(maxConstraints.Y-reservedWidth, 0)
		minConstraints.Y = max(minConstraints.Y-reservedWidth, 0)
		gtx.Constraints.Max = axis.Convert(maxConstraints)
		gtx.Constraints.Min = axis.Convert(minConstraints)
	}

	listDims := layoutVisualOutsetList(ctx, gtx, list, count, visual, item)
	gtx.Constraints = originalConstraints
	majorAxisSize := axis.Convert(listDims.Size).X
	start, end := scrollbarViewport(list.Position, count, majorAxisSize)
	scrollable := start != 0 || end != 1
	if !overlay && reserveScrollbar != scrollable {
		gtx.Execute(op.InvalidateCmd{})
	}

	gtx.Constraints.Min = listDims.Size
	if reserveScrollbar {
		minor := axis.Convert(gtx.Constraints.Min)
		minor.Y += reservedWidth
		gtx.Constraints.Min = axis.Convert(minor)
	}
	anchor := layout.E
	if axis == layout.Horizontal {
		anchor = layout.S
	}
	anchor.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return style.Layout(gtx, axis, start, end)
	})

	if scrollable {
		if distance := bar.ScrollDistance(); distance != 0 {
			list.ScrollBy(distance * float32(count))
		}
	}
	if reserveScrollbar {
		minor := axis.Convert(listDims.Size)
		minor.Y += reservedWidth
		listDims.Size = axis.Convert(minor)
	}
	return listDims
}

// TrackedScrollbar lays out a Gio list with FlowUI's scrollbar style.
func TrackedScrollbar(ctx *Context, gtx layout.Context, list *layout.List, bar *widget.Scrollbar, count int, disabled, overlay bool, item layout.ListElement) layout.Dimensions {
	return layoutScrollbarList(ctx, gtx, "", list, bar, nil, count, disabled, overlay, style.Style{}, nil, true, item)
}

// TrackedScrollbarWithVisualOutset lays out a Gio list with a persistent
// visual-outset state. Composite components use it when they clip arbitrary
// child content inside their own viewport.
func TrackedScrollbarWithVisualOutset(ctx *Context, gtx layout.Context, list *layout.List, bar *widget.Scrollbar, visual *VisualOutsetState, count int, disabled, overlay bool, item layout.ListElement) layout.Dimensions {
	return layoutScrollbarList(ctx, gtx, "", list, bar, visual, count, disabled, overlay, style.Style{}, nil, true, item)
}

type scrollbarDeclarationKey struct {
	disabled bool
	dragging bool
}

func resolvedScrollbarStyle(ctx *Context, gtx layout.Context, key string, bar *widget.Scrollbar, disabled bool, custom style.Style, caches ...*computed.VisualCache[scrollbarDeclarationKey, style.Style]) material.ScrollbarStyle {
	activeTheme := core.ActiveTheme(ctx)
	tokens := defaultScrollbarMetrics
	base := scrollbarStyleFor(activeTheme, bar, disabled)
	state := scrollbarStyleState(bar, disabled)
	declarationKey := scrollbarDeclarationKey{disabled: disabled, dragging: bar.Dragging()}
	var cache *computed.VisualCache[scrollbarDeclarationKey, style.Style]
	if len(caches) > 0 {
		cache = caches[0]
	}
	defaults := cache.Resolve(activeTheme, declarationKey, func() style.Style {
		return scrollbarDefaults(tokens, base)
	})

	track := resolveScrollbarPart(ctx, gtx, key, style.PartTrack, state, defaults, custom)
	thumb := resolveScrollbarPart(ctx, gtx, key, style.PartThumb, state, defaults, custom)
	trackWidth := tokens.TrackWidth
	thumbWidth := tokens.ThumbWidth
	if track.Box != nil && track.Box.Width != nil {
		trackWidth = max(*track.Box.Width, 1)
	}
	if thumb.Box != nil && thumb.Box.Width != nil {
		thumbWidth = max(*thumb.Box.Width, 1)
	}
	base.Indicator.MinorWidth = thumbWidth
	base.Track.MinorPadding = max((trackWidth-thumbWidth)/2, 0)
	if track.Paint != nil {
		if brush, ok := computed.Brush(track.Paint.Background); ok {
			base.Track.Color = computed.ApplyOpacity(brush.ColorAt(.5), track.Paint.Opacity)
		}
	}
	if thumb.Paint != nil {
		if brush, ok := computed.Brush(thumb.Paint.Background); ok {
			indicator := computed.ApplyOpacity(brush.ColorAt(.5), thumb.Paint.Opacity)
			base.Indicator.Color = indicator
			base.Indicator.HoverColor = indicator
		}
		if thumb.Paint.Radius != nil {
			base.Indicator.CornerRadius = *thumb.Paint.Radius
		}
	}
	return base
}

func scrollbarDefaults(tokens scrollbarMetrics, base material.ScrollbarStyle) style.Style {
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.
			Width(tokens.TrackWidth).
			Background(style.SolidColor{Color: base.Track.Color})).
		Part(style.PartThumb, style.Style{}.
			Width(tokens.ThumbWidth).
			Background(style.SolidColor{Color: base.Indicator.Color}).
			Radius(tokens.Radius))
}

func scrollbarStyleState(bar *widget.Scrollbar, disabled bool) style.State {
	dragging := !disabled && bar.Dragging()
	return style.State{Pressed: dragging, Dragging: dragging, Disabled: disabled}
}

func resolveScrollbarPart(ctx *Context, gtx layout.Context, key string, part style.Part, state style.State, defaults, custom style.Style) style.ResolvedStyle {
	if key == "" {
		return computed.ResolvePartStatic(ctx, part, state, defaults, style.Style{}, style.Style{}, custom)
	}
	return computed.ResolvePart(ctx, gtx, key, part, state, defaults, style.Style{}, style.Style{}, custom)
}

func scrollbarStyleFor(activeTheme *theme.Theme, state *widget.Scrollbar, disabled bool) material.ScrollbarStyle {
	sem := activeTheme.Semantic()
	tokens := defaultScrollbarMetrics
	style := material.Scrollbar(theme.MaterialOf(activeTheme), state)
	style.Track.MajorPadding = tokens.MajorPadding
	style.Track.MinorPadding = max((tokens.TrackWidth-tokens.ThumbWidth)/2, 0)
	style.Track.Color = color.NRGBA{}
	style.Indicator.MajorMinLen = tokens.MinThumbLength
	style.Indicator.MinorWidth = tokens.ThumbWidth
	style.Indicator.CornerRadius = tokens.Radius
	style.Indicator.Color = scrollbarColor(sem.Foreground, tokens.ThumbOpacity)
	style.Indicator.HoverColor = scrollbarColor(sem.Foreground, tokens.HoverOpacity)
	if disabled {
		style.Indicator.Color = activeTheme.DisabledColor(style.Indicator.Color)
		style.Indicator.HoverColor = style.Indicator.Color
	} else if state.Dragging() {
		style.Indicator.Color = style.Indicator.HoverColor
	}
	return style
}

func scrollbarColor(value color.NRGBA, opacity float32) color.NRGBA {
	opacity = min(max(opacity, 0), 1)
	value.A = byte(float32(value.A)*opacity + .5)
	return value
}

// scrollbarViewport maps Gio's estimated list position to the normalized range used by widget.Scrollbar.
func scrollbarViewport(position layout.Position, elements, majorAxisSize int) (start, end float32) {
	if elements <= 0 || position.Length <= 0 || majorAxisSize >= position.Length {
		return 0, 1
	}
	length := float32(position.Length)
	elementLength := length / float32(elements)
	start = min(max((float32(position.First)*elementLength+float32(position.Offset))/length, 0), 1)
	end = min(max((float32(position.First+position.Count)*elementLength+float32(position.OffsetLast))/length, 0), 1)
	visibleFraction := min(max(float32(majorAxisSize)/length, 0), 1)
	viewportFraction := end - start
	if viewportFraction < 1 {
		start -= start / (1 - viewportFraction) * (visibleFraction - viewportFraction)
		end += (1 - end) / (1 - viewportFraction) * (visibleFraction - viewportFraction)
	}
	return min(max(start, 0), 1), min(max(end, 0), 1)
}
