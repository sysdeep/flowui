# FlowUI Theme Architecture

The `theme` tree separates visual contracts, shared component recipes, and
the default design system. Applications normally import `theme/base` and pass
`base.Light()` or `base.Dark()` to `window.Theme`. Other design systems are
provided by external modules. Applications do not need to import the component
recipe packages directly.

## Package Roles

| Package | Responsibility |
| --- | --- |
| `theme` | Defines `Theme`, semantic values, component requests, visual results, and Provider interfaces. |
| `theme/<domain>style` | One recipe domain: `action`, `field`, `feedback`, `navigation`, `overlay`, `surface`, `data`, or `chart`. Per-component files live inside it. |
| `theme/base` | Provides the single default design system: Light/Dark entry points, seed constructors and a complete zero-value Provider. |
| `theme/themetest` | Provides reusable visual-contract acceptance tests for the default and external themes. |
| `theme/internal` | Holds helpers that are implementation details of the theme package. |

The `*style` suffix means **component visual recipe**, not a widget and not a
second component implementation. For example, `theme/actionstyle` groups the
button, close-button, toggle-button, checkbox, switch, and radio recipes. Each
component's code lives in a prefixed file set inside the domain package (for
example `buttonstyle_render.go`), so unexported helpers stay private to the
domain without a second package. A recipe maps semantic colors, variants,
sizes, motion, and theme options to a visual; it does not handle pointer
events, focus state, business state, or component layout ownership.

### Migrating a per-component import

The Build function names did not change; replace the import path and package
selector with the domain:

| Old package | New package |
| --- | --- |
| `theme/buttonstyle`, `theme/closebuttonstyle`, `theme/togglestyle`, `theme/checkboxstyle`, `theme/switchstyle`, `theme/radiostyle` | `theme/actionstyle` |
| `theme/inputactionstyle`, `theme/combostyle`, `theme/selectstyle`, `theme/datepickerstyle`, `theme/colorpickerstyle`, `theme/sliderstyle` | `theme/fieldstyle` |
| `theme/alertstyle`, `theme/alertdialogstyle`, `theme/toaststyle`, `theme/progressstyle`, `theme/spinnerstyle`, `theme/skeletonstyle`, `theme/badgestyle`, `theme/chipstyle` | `theme/feedbackstyle` |
| `theme/tabsstyle`, `theme/sidebarstyle`, `theme/paginationstyle`, `theme/segmentedstyle`, `theme/menubarstyle`, `theme/menustyle`, `theme/dropdownstyle`, `theme/collapsiblestyle` | `theme/navigationstyle` |
| `theme/modalstyle`, `theme/popoverstyle`, `theme/tooltipstyle` | `theme/overlaystyle` |
| `theme/cardstyle`, `theme/surfacestyle`, `theme/textstyle`, `theme/labelstyle`, `theme/descriptionstyle`, `theme/codestyle`, `theme/imagestyle`, `theme/iconstyle`, `theme/avatarstyle` | `theme/surfacestyle` |
| `theme/tablestyle`, `theme/treestyle`, `theme/listboxstyle`, `theme/toolbarstyle`, `theme/statusbarstyle`, `theme/titlebarstyle` | `theme/datastyle` |
| `theme/barchartstyle`, `theme/linechartstyle`, `theme/piechartstyle`, `theme/candlestickstyle`, `theme/heatmapstyle`, `theme/ganttstyle`, `theme/nodegraphstyle`, `theme/timelinestyle` | `theme/chartstyle` |

`theme/surfacestyle` is both a domain and an old component name; the old
component files now live inside the domain package alongside the other surface
recipes.

## Runtime Flow

```text
widget/button
    -> theme.ComponentProvider
    -> base.Provider.Button
    -> actionstyle.BuildButton
    -> theme.ButtonVisual
    -> widget/button resolves declarations and lays out the result
```

Most widgets use `theme.ResolveVisual` with a Provider method expression and
the matching validator. Button retains a specialized validation cache for
immutable Providers. Shared components, such as Toolbar, follow this path:

```text
widget/toolbar
    -> theme.ResolveVisual
    -> base.Provider.Toolbar
    -> datastyle.BuildToolbar
    -> validated theme.ToolbarVisual
```

A widget owns behavior, identity, interaction state, accessibility, and domain
layout. The Provider
owns the visual recipe and any theme-specific render strategy. The `style`
package resolves the returned declarations against the current state.

## Why Recipes Are Separate

Keeping recipes separate from concrete themes avoids copying every component
implementation into each design system. The core owns one complete default
Provider. External Providers reuse it and supply their own palette, component
metrics or recipe configuration where needed.

External Providers can live in separate modules and reuse the core Provider.
The core has no dependency on them. Generic colour ramps and seed derivation
remain in `theme`.

## Custom Providers

Applications that only need to customize an existing design system should use
the public `style` declarations, theme customization APIs, and component
overrides. To override a component's entire visual recipe, embed the existing
`theme.ComponentVisualProvider` and implement that component's method on the
outer type. Install the outer type as the active theme's `VisualProvider`.

For a new design system, start with a complete light or dark theme:

```go
active := base.New(theme.DefaultSeed())
dark := base.NewDark(theme.DefaultSeed())
```

These constructors install `base.Provider{}`, which supplies every component,
including Button, and reads the active theme's semantic values and component
metrics. Unlike these constructors, `theme.New` and `theme.NewDark` only build
theme values; they still require an explicit visual Provider.

Embed `base.Provider` and override only the recipes your design changes:

```go
package customtheme

import (
    "github.com/qianniancn/flowui/theme"
    "github.com/qianniancn/flowui/theme/base"
    "github.com/qianniancn/flowui/theme/actionstyle"
)

type Provider struct{ base.Provider }

var _ theme.ComponentVisualProvider = Provider{}

func (Provider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
    config := actionstyle.DefaultConfig()
    config.Radius = 0
    return actionstyle.BuildButton(active, request, config)
}

func New(seed theme.Seed) theme.Theme {
    active := base.New(seed)
    active.VisualProvider = Provider{}
    return active
}
```

`actionstyle.DefaultConfig()` returns a complete configuration, including
spinner, pressed-scale, and group separator values. Change only the fields
you need. `BuildButton` uses supplied values directly: valid zeros such as
radius, border width, and content gap stay zero. It does not merge defaults
over explicit values. [The executable example](base/provider_test.go) also
shows a Toolbar metric override. External design systems can supply their own
Button configurations without copying component behavior.

`base.Provider` deliberately does not implement `theme.ImmutableVisualProvider`.
Only add `VisualsAreImmutable()` to your own Provider when all returned visuals
are deterministic for an unchanged theme and request. Embedding shared recipes
alone cannot guarantee that custom methods or captured state are immutable.

Every running window must provide one complete `theme.ComponentVisualProvider`.
Composition happens inside that Provider. Missing methods and incomplete visual
results remain errors; widgets do not select an implicit fallback theme.

## Diagnostics

`ValidateVisualProvider` and `ComponentProvider` return an error listing only
the missing or incompatible methods, for example `Provider.DatePicker`.
Component validators keep their fail-fast panic behavior, but now identify
the component, field, and reason:

```text
flowui: theme visual provider is incomplete: Button.SpinnerSize: must be finite and greater than 0
```

These errors have type `*theme.VisualValidationError`, with `Component`,
`Field`, and `Reason` fields. Both returned errors and recovered validator
errors support `errors.Is(err, theme.ErrIncompleteVisualProvider)` and
`errors.As(err, &detail)`. Code comparing errors directly with the sentinel
should use `errors.Is` to preserve compatibility with the added detail.
Missing Providers still return `theme.ErrMissingVisualProvider`.

## Theme Acceptance Tests

Use the same acceptance suite for built-in and custom themes:

```go
func TestMyTheme(t *testing.T) {
    t.Run("light", func(t *testing.T) {
        themetest.Check(t, New(theme.DefaultSeed()))
    })
    t.Run("dark", func(t *testing.T) {
        active := base.NewDark(theme.DefaultSeed())
        active.VisualProvider = Provider{}
        themetest.Check(t, active)
    })
}
```

Import `github.com/qianniancn/flowui/theme/themetest` in your test file. The
suite validates every Provider method, required renderers, resolved geometry,
paint, text, and motion. It also exercises Button variants, sizes, icon-only
and grouped requests, instance overrides, Input variants/sizes/modes, related
field variants, and Tree variants/sizes. Provider or validator panics fail the
named subtest. The suite's protocol coverage guard fails when a new component
method has no acceptance case.

The core runs this suite for default themes, seed-based constructors and a
custom Provider in light and dark modes. External Providers run the same suite.

## Adding a Component

1. Define its request, visual result, Provider method, and validator together
   in `theme/<component>_visual.go`. Include its capability in
   `ComponentVisualProvider` in `theme/provider.go`; shared scalar and style
   validation helpers stay in `theme/visual_validation.go`.
2. Implement the recipe in the matching `theme/<domain>style` package, in a
   `<component>_*.go` file set. Prefix its unexported helpers with the component
   name so they stay unique inside the domain, then add its forwarding method
   once to `base.Provider`. Concrete themes inherit it; only themes needing a
   different recipe add an override.
3. Keep interaction and layout in `widget/<component>`. Build the visual with
   `theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Component,
   theme.ValidateComponentVisual)` so lookup and validation stay together.
4. If caching is needed, use the existing instance-owned `ProviderCache` or
   `ProviderCacheMap` in `internal/computed`. Include every visual request input
   in the key, or bypass caching for instance overrides that cannot be keyed
   safely. Add behavior tests and update independent test Providers as needed.
5. Add a component case to `theme/themetest/check.go`. Its protocol coverage
   guard checks that every Provider method has a case; the same acceptance
   checks then run against all supported themes.

## Cache Lifecycle

Both Provider caches invalidate when the Theme pointer, its VisualProvider, or
the request key changes. A Provider replacement on the same Theme therefore
rebuilds the visual and validates the new Provider before reuse. Failed loads
leave the previous successful entry intact. Map caches discard all old request
entries after a successful theme or Provider change.

Provider identity must be safely comparable. Non-comparable Provider values
(including interface fields holding slices) reload instead of risking stale
results or comparison panics. Changes to state behind an unchanged Provider
pointer cannot be detected: use the context's theme replacement API when
changing appearance. The `ThemeInvalidations` counter includes Provider changes.

## Directory Guidance

The public `theme/<domain>style` packages are stable. The per-component files
inside a domain are free to move within it without changing public import
paths. If a domain boundary changes later, update this document and the
Provider and architecture tests in the same change.

Component contracts and validators live in separate `*_visual.go` files inside
the existing `theme` package. This organization changes where declarations are
maintained while keeping public type names and import paths stable.
