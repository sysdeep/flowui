package datepicker

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type dateFieldColors struct {
	Foreground  color.NRGBA
	Placeholder color.NRGBA
	Selection   color.NRGBA
}

type resolvedDateField struct {
	Content style.ResolvedStyle
	Colors  dateFieldColors
}

func resolveDatePickerField(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	custom style.Style,
	defaults datePickerDefaults,
) resolvedDateField {
	// The owning widget already resolved and cached the complete defaults.
	// Read the field declaration directly so field resolution does not maintain
	// a second cache for a value that has no independent invalidation boundary.
	content := computed.ResolvePart(ctx, gtx, key, style.PartContent, state, defaults.field, style.Style{}, style.Style{}, custom)
	placeholder := computed.ResolvePart(ctx, gtx, key, style.PartPlaceholder, state, defaults.field, style.Style{}, style.Style{}, custom)
	selection := computed.ResolvePart(ctx, gtx, key, style.PartSelection, state, defaults.field, style.Style{}, style.Style{}, custom)
	return resolvedDateField{
		Content: content,
		Colors:  dateResolveFieldColors(content, placeholder, selection),
	}
}

func (d Widget) resolvePart(ctx *core.Context, gtx layout.Context, part style.Part, state style.State, calendar *style.Style) style.ResolvedStyle {
	return resolvePartFor(ctx, gtx, core.FullKey(ctx, d.key), part, state, d.style, style.Style{}, calendar)
}

func (d Widget) resolvePartWithDefaults(ctx *core.Context, gtx layout.Context, part style.Part, state style.State, defaults style.Style) style.ResolvedStyle {
	return computed.ResolvePart(ctx, gtx, core.FullKey(ctx, d.key), part, state, defaults, style.Style{}, style.Style{}, d.style)
}

func resolvePartFor(ctx *core.Context, gtx layout.Context, key string, part style.Part, state style.State, custom, declaration style.Style, calendar *style.Style) style.ResolvedStyle {
	if calendar == nil {
		panic("flowui/datepicker: missing calendar declaration")
	}
	// Merge only the requested part. Joining the whole declaration deep-clones
	// every calendar part for each cell before the resolver sees this declaration.
	defaults := style.JoinPart(*calendar, declaration, part)
	return computed.ResolvePart(ctx, gtx, key, part, state, defaults, style.Style{}, style.Style{}, custom)
}

func (d Widget) datePickerDescriptionColor(ctx *core.Context, gtx layout.Context, invalid, enabled bool, calendar *style.Style) color.NRGBA {
	state := style.State{Disabled: !enabled, Invalid: invalid}
	resolved := d.resolvePart(ctx, gtx, style.PartDescription, state, calendar)
	return dateResolveTextColor(resolved)
}
