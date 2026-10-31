# FlowUI Theme Architecture

FlowUI has one default visual design. `theme` constructs its light and dark
semantic tokens; component packages in `widget` own the default appearance
that consumes those tokens. `widget/base` contains behavior without a default
appearance. Shared implementation details used by multiple widgets belong in
`widget/internal`, not in public theme recipe packages.

Applications pass `theme.Light()`, `theme.Dark()`, or a seed-based `theme.New` /
`theme.NewDark` value to the window. Customize shared colors and typography on
the theme, and customize individual components through their widget APIs and
style overrides.

Component dimensions and part declarations live in their widget packages and
can be customized with `Style`. All constructors use the same default appearance, including semantic colors,
installed font families and shadows. `theme.DefaultSeed()` is the seed used by
`theme.Light()` and `theme.Dark()`; change it before calling `theme.New` or
`theme.NewDark` to customize colors and motion.

Component appearance and component vocabulary live with the widgets. Shared
field-shell pointer behavior lives in `widget/base.FieldInteraction`; the shared
built-in field appearance lives in `widget/internal/fieldstyle`. DatePicker owns its
resolved field colors and style snapshots privately.
Component enums are declared by their widget packages, for example
`button.Variant`, `button.Size`, and `progress.BarColor`. Applications customize
components through widget methods and `style.Style`. Missing required visual
layers are component-owned invariant failures; `theme.ErrMissingTheme` remains
a theme error.

Widget-owned visual caches are keyed by the theme pointer and request. Replace
the theme through the framework when changing it; mutating a theme in place
does not automatically invalidate cached visuals.
