package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

var inputEditorStyleParts = [...]style.Part{
	style.PartRoot,
	style.PartPlaceholder,
	style.PartSelection,
}

var inputEditorRootStyleParts = [...]style.Part{style.PartRoot}

// Text areas and multiline groups resolve a provisional zero-height appearance
// before requesting the row-derived height. Keep both entries so the normal
// frame does not rebuild the defaults twice.
const inputDefaultsCacheCapacity = 8

type inputEditorStyleResolution struct {
	ctx              *core.Context
	gtx              layout.Context
	key              string
	state            style.State
	defaults, custom style.Style
}

func resolveInputEditorPartsInto(request inputEditorStyleResolution, parts []style.Part, destination []style.ResolvedStyle) {
	computed.ResolvePartsInto(
		request.ctx,
		request.gtx,
		request.key,
		request.state,
		parts,
		destination,
		request.defaults,
		style.Style{},
		style.Style{},
		request.custom,
	)
}

func resolveInputEditorStyles(request inputEditorStyleResolution) (root, placeholder, selection style.ResolvedStyle) {
	var resolved [len(inputEditorStyleParts)]style.ResolvedStyle
	resolveInputEditorPartsInto(request, inputEditorStyleParts[:], resolved[:])
	return resolved[0], resolved[1], resolved[2]
}

func resolveInputEditorRootStyle(request inputEditorStyleResolution) style.ResolvedStyle {
	var resolved [len(inputEditorRootStyleParts)]style.ResolvedStyle
	resolveInputEditorPartsInto(request, inputEditorRootStyleParts[:], resolved[:])
	return resolved[0]
}

// inputStyleKey contains only the inputs that change the cached defaults. The
// instance style is resolved separately as the last cascade layer, so it never
// enters the key.
type inputStyleKey struct {
	variant   Variant
	size      Size
	fullWidth bool
	height    unit.Dp
}

func (s *inputState) defaults(i Widget, active *theme.Theme) style.Style {
	key := inputStyleKey{variant: i.variant, size: i.size, fullWidth: i.fullWidth}
	return s.styleCache.Resolve(active, key, func() style.Style {
		return defaultInputRules(active, i.variant, i.size, i.fullWidth)
	})
}

func (s *textAreaState) defaults(t TextAreaWidget, active *theme.Theme, height unit.Dp) style.Style {
	s.styleCache.SetCapacity(inputDefaultsCacheCapacity)
	key := inputStyleKey{variant: t.variant, fullWidth: t.fullWidth, height: height}
	return s.styleCache.Resolve(active, key, func() style.Style {
		return defaultTextAreaRules(active, t.variant, t.fullWidth, height)
	})
}

func (s *inputGroupState) defaults(g GroupWidget, active *theme.Theme, minHeight unit.Dp) style.Style {
	s.styleCache.SetCapacity(inputDefaultsCacheCapacity)
	key := inputStyleKey{variant: g.variant, fullWidth: g.fullWidth, height: minHeight}
	return s.styleCache.Resolve(active, key, func() style.Style {
		return defaultGroupRules(active, g.variant, g.fullWidth, minHeight)
	})
}

const stateSlotInputActionDefaults = "input-action-defaults"

type groupActionDefaultsCache = computed.VisualCache[struct{}, groupActionDefaults]

func groupActionDefaultsFor(ctx *core.Context, key string) groupActionDefaults {
	active := core.ActiveTheme(ctx)
	cache := core.UseState[groupActionDefaultsCache](ctx, core.FullKey(ctx, key), stateSlotInputActionDefaults)
	return cache.Resolve(active, struct{}{}, func() groupActionDefaults {
		return defaultGroupAction(active)
	})
}

type inputGroupResolvedStyle struct {
	root        style.ResolvedStyle
	content     style.ResolvedStyle
	prefix      style.ResolvedStyle
	suffix      style.ResolvedStyle
	divider     style.ResolvedStyle
	selection   style.ResolvedStyle
	placeholder style.ResolvedStyle
	lineHeight  unit.Sp
}

var inputGroupStyleParts = [...]style.Part{
	style.PartRoot,
	style.PartContent,
	style.PartPrefix,
	style.PartSuffix,
	style.PartIndicator,
	style.PartPlaceholder,
	style.PartSelection,
}

func resolveInputGroupStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, defaults, custom style.Style) inputGroupResolvedStyle {
	var resolved [len(inputGroupStyleParts)]style.ResolvedStyle
	computed.ResolvePartsInto(ctx, gtx, key, state, inputGroupStyleParts[:], resolved[:], defaults, style.Style{}, style.Style{}, custom)
	return inputGroupResolvedStyle{
		root:        resolved[0],
		content:     resolved[1],
		prefix:      resolved[2],
		suffix:      resolved[3],
		divider:     resolved[4],
		placeholder: resolved[5],
		selection:   resolved[6],
	}
}
