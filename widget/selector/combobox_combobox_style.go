package selector

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type comboboxComboResolvedStyle struct {
	content     style.ResolvedStyle
	placeholder style.ResolvedStyle
	selection   style.ResolvedStyle
	indicator   style.ResolvedStyle
	icon        style.ResolvedStyle
	visual      theme.ComboVisual
}

func (s *comboboxComboBoxState) partTransitionKey(ctx *core.Context, owner string, part style.Part) string {
	if s.styleKeyOwner != owner {
		s.styleKeyOwner = owner
		s.stylePartKeys = [5]string{}
	}
	index := -1
	switch part {
	case theme.ComboPartTrigger:
		index = 0
	case theme.ComboPartPlaceholder:
		index = 1
	case theme.ComboPartSelection:
		index = 2
	case theme.ComboPartIndicator:
		index = 3
	case theme.ComboPartIcon:
		index = 4
	}
	if index < 0 {
		return core.DerivedKey(ctx, owner, "combo-part:"+string(part))
	}
	if s.stylePartKeys[index] == "" {
		s.stylePartKeys[index] = core.DerivedKey(ctx, owner, "combo-part:"+string(part))
	}
	return s.stylePartKeys[index]
}

func comboboxComboVisualBundle(activeTheme *theme.Theme, variant field.Variant, override *theme.ComboOverride) comboboxComboDeclarationBundle {
	visual := theme.ResolveVisual(activeTheme, theme.ComboRequest{
		Variant:  theme.InputVariant(variant),
		Override: override,
	}, theme.ComponentVisualProvider.Combo, theme.ValidateComboVisual)
	return comboboxComboDeclarationBundle{
		visual: visual,
	}
}

type comboboxComboDeclarationBundle struct {
	visual theme.ComboVisual
}

// comboStyleDeclarationsCached resolves the provider-owned base bundle. An
// instance override deliberately bypasses this cache at the call site so it
// can never pollute the unstyled declaration used by a later frame.
func comboboxComboStyleDeclarationsCached(activeTheme *theme.Theme, variant field.Variant, cache ...*computed.ProviderCache[field.Variant, comboboxComboDeclarationBundle]) comboboxComboDeclarationBundle {
	if len(cache) > 0 && cache[0] != nil {
		return cache[0].Resolve(activeTheme, variant, func() comboboxComboDeclarationBundle {
			return comboboxComboVisualBundle(activeTheme, variant, nil)
		})
	}
	return comboboxComboVisualBundle(activeTheme, variant, nil)
}

func (c ComboboxWidget) resolveStyle(ctx *core.Context, gtx layout.Context, state *comboboxComboBoxState, key string, styleState style.State, instanceStyle style.Style, bundles ...comboboxComboDeclarationBundle) comboboxComboResolvedStyle {
	var declarations comboboxComboDeclarationBundle
	if len(bundles) > 0 {
		declarations = bundles[0]
	} else {
		declarations = comboboxComboStyleDeclarationsCached(core.ActiveTheme(ctx), c.variant)
	}
	defaults := declarations.visual.Defaults
	resolvePart := func(part style.Part) style.ResolvedStyle {
		partKey := state.partTransitionKey(ctx, key, part)
		return computed.ResolvePartWithKey(ctx, gtx, partKey, part, styleState, defaults, style.Style{}, style.Style{}, instanceStyle)
	}
	resolved := comboboxComboResolvedStyle{
		visual:      declarations.visual,
		content:     resolvePart(theme.ComboPartTrigger),
		placeholder: resolvePart(theme.ComboPartPlaceholder),
		selection:   resolvePart(theme.ComboPartSelection),
		indicator:   resolvePart(theme.ComboPartIndicator),
		icon:        resolvePart(theme.ComboPartIcon),
	}
	return resolved
}
