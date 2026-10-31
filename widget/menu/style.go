package menu

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// PanelMetrics contains the popup motion geometry an owner needs to place
// adjacent overlay chrome around a menu panel. It is the only menu appearance
// data shared across packages; item, text, and indicator drawing stay
// menu-owned.
type PanelMetrics struct {
	EnterScale        float32
	ExitScale         float32
	AnimationDistance unit.Dp
}

// RuntimeVisual is one layout pass's resolved menu popup. A Runtime resolves it
// once and hands it to layout; popup owners can reuse it for adjacent chrome
// instead of rebuilding the menu appearance.
type RuntimeVisual struct {
	Metrics  PanelMetrics
	defaults menuDefaults
	panel    style.ResolvedStyle
}

// PanelColors returns the concrete popup surface and border colors, falling
// back to the theme defaults when the resolved panel uses a brush that adjacent
// chrome cannot reproduce as a solid fill.
func (v RuntimeVisual) PanelColors() (background, border color.NRGBA) {
	background, border = v.defaults.Colors.PanelBackground, v.defaults.Colors.PanelBorder
	if v.panel.Paint == nil {
		return background, border
	}
	if value, ok := v.panel.Paint.Background.(style.SolidColor); ok {
		background = value.Color
	}
	if v.panel.Paint.Border != nil {
		if value, ok := v.panel.Paint.Border.Color.(style.SolidColor); ok {
			border = value.Color
		}
	}
	return background, border
}

// PanelChrome returns the resolved panel corner radius and border width, which
// an owner needs to align adjacent chrome such as an anchored arrow.
func (v RuntimeVisual) PanelChrome() (radius, borderWidth unit.Dp) {
	if v.panel.Paint == nil {
		return 0, 0
	}
	if v.panel.Paint.Radius != nil {
		radius = *v.panel.Paint.Radius
	}
	if v.panel.Paint.Border != nil && v.panel.Paint.Border.Width != nil {
		borderWidth = *v.panel.Paint.Border.Width
	}
	return radius, borderWidth
}

type menuDefaultsCache = computed.VisualCache[bool, menuDefaults]

func (m Widget) defaults(active *theme.Theme, cache *menuDefaultsCache) menuDefaults {
	key := m.compact
	load := func() menuDefaults { return buildMenu(active, m.compact) }
	return cache.Resolve(active, key, load)
}

// visual resolves the default appearance from the active theme.
func (m Widget) visual(ctx *core.Context) RuntimeVisual {
	return m.visualWithCache(ctx, nil)
}

func (m Widget) visualWithCache(ctx *core.Context, cache *menuDefaultsCache) RuntimeVisual {
	active := core.ActiveTheme(ctx)
	defaults := m.defaults(active, cache)
	// The instance style is the last cascade layer, applied after the cached
	// defaults so Style can sit anywhere in the option chain.
	panel := computed.ResolveStatic(ctx, style.State{}, defaults.Root, style.Style{}, style.Style{}, m.style)
	return RuntimeVisual{
		Metrics: PanelMetrics{
			EnterScale:        defaults.Metrics.EnterScale,
			ExitScale:         defaults.Metrics.ExitScale,
			AnimationDistance: defaults.Metrics.AnimationDistance,
		},
		defaults: defaults,
		panel:    panel,
	}
}
