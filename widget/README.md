# Components and shared primitives

Applications import components directly, for example
`github.com/qianniancn/flowui/widget/button`. There is no Go package at the
`widget` directory itself. The `ui` package supplies the MVU/runtime contract.

The main constructor in each component package is `New`, returning `Widget`.
Related controls keep descriptive names, such as `button.Group`,
`input.TextArea`, and `progress.Bar`.

Text controls accept strings directly:

```go
button.New("Save").Key("save").OnClick(save)
toggle.NewToggleButton("Bold").Key("bold").Selected(selected).OnChange(setBold)
tooltip.New("save-tip", saveButton, "Save changes")
input.GroupAction("clear", "Clear").OnClick(clear)
```

Constructors such as `button.New` provide automatic keys for fixed layouts.
Identity follows component type, call site, enclosing key scope, and the order
of first key resolution at that site each frame. Dynamic lists and conditional
siblings built by a shared helper need `.Key(businessID)` or a stable
`layout.Key(businessID, child)` scope; ordinary layout containers do not create
key scopes. See [Keys](../docs/en/guide/02-core-concepts.md#keys).

Use `Content` for icons, styled text, or composite widgets. For buttons, toggle
buttons, and input actions, the constructor text remains the accessible label;
`Label` overrides that label without changing visible content.

```go
button.New("Save").Key("save").Content(icon.New(lucide.Save)).IconOnly()
button.New("Save").Key("save-large").Content(text.New("Save").Size(20))
tooltip.New("details", trigger, "").Content(details)
```

Containers such as cards, toolbars, and modal bodies accept widgets so callers
can compose their contents. Standalone text still uses `text.New`.

Each component owns its domain layout and default appearance, composing
appearance-free interaction and accessibility behavior from `widget/base`. Shared theme tokens live in `theme`. See the
[theme package map](../theme/README.md).

Some public packages are building blocks for components rather than complete
controls:

| Package | Purpose |
| --- | --- |
| `navigation` | Keyboard movement, disabled-item skipping, and typeahead |
| `chart` | Shared chart data, annotations, scales, caching, and interaction helpers |

These packages retain their public import paths so custom components can reuse
them. In particular, types such as `chart.Selection` and `chart.DataWindow`
appear in chart components' public APIs.

Generic controlled/uncontrolled value bindings and shared row animation/drawing
helpers live in `widget/internal`. Expand/collapse interaction is
`base.Disclosure`; it is distinct from generic value binding.

Put helpers that are private to the widget tree under `widget/internal/` when
needed. Helpers shared across framework layers belong in root `internal/`.
Tests use the production light/dark themes. Configure any test-specific motion
timings locally rather than maintaining a separate test theme.

## Layout-component boundary

`layout` owns arrangement, the generic Style/Box adapter, overlay placement and
clipping viewports. Default visual separators are `divider.New`; draggable
split panes are `splitpane.New`. SplitPane composes `base.SplitPaneInteraction`
for pointer capture, keyboard resizing and handle semantics. The base owns no
colors, theme values or focus animation. Pure clipping/overflow reservation is
`layout.Viewport`, so layout never imports a visual widget package.

Use `divider.New`, `splitpane.New` and `layout.Scroll` for separators,
draggable split panes and scroll viewports respectively. `Box.Style` merges
styles, while `ReplaceStyle` replaces them. Custom visuals compose through the
common `Resolved` / `InteractiveResolved` shell and normal widget composition.
