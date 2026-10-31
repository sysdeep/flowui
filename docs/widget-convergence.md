# Widget convergence audit

This is the current naming and responsibility baseline for `widget/`. It is
intentionally an audit rather than a bulk rename: the repository is already in
the middle of a large file split, so renaming public types or moving packages
without a compatibility plan would create avoidable churn.

## Style and explicit APIs

`Widget.Style(...)` remains the escape hatch for low-frequency or theme-level
overrides. It is still appropriate for `Part(...)` overrides, state rules,
transitions, and visual properties such as radius, background, and text color.

Stable, frequently used layout and behavior options should have named methods.
The input family now follows this rule:

```go
input.New(value).Width(320)
input.NewTextArea(value).Width(320)
input.Group(...).Width(320)
```

The form example has been migrated to this API. Outer layout remains expressed
through `layout.Box(...).Style(...)`, because width and padding there belong to
the container rather than the child component.

## Components with similar responsibilities

| Components | Decision | Reason |
| --- | --- | --- |
| `selector` / `combobox` | Keep separate | `selector` chooses an existing item; `combobox` edits/filter text and can accept custom values. |
| `input` / `numericinput` | Keep separate | `numericinput` owns numeric parsing, stepping, and bounds in addition to text editing. |
| `text.New` / `text.NewLabel` / `text.NewDescription` | Keep separate | They communicate different semantics and accessibility intent even when their rendering is related. |
| `surface.New` / `surface.NewCard` | Keep separate | `surface.New` is a generic semantic surface; `Card` adds raised-content defaults, variants, padding, and gap. |
| `surface.NewPanel` | Keep for now | This is a lifecycle-aware view host, not a visual panel. It is used by `widget/workbench`, so a rename needs an API migration or alias. |

No additional component was found that is both redundant and safe to delete in
this pass. Existing duplicate-looking implementations are either distinct
interaction models or already being removed by the ongoing refactor.

## Variant review

Variants should describe a small, stable visual vocabulary rather than replace
behavioral methods. The current `surface` variants are scoped to the surface
family and should not be reused by unrelated controls. Button and toggle
variants should remain local to their packages for the same reason.

When a variant only changes one stable behavior or layout rule, prefer a named
method instead of adding another variant value. Keep `Style(...)` available for
one-off visual overrides, but do not require callers to use it for common
component behavior.

## Naming findings

### Public type and package names

- `surface.PanelWidget` is semantically a panel host. A future
  `PanelHostWidget` or `PanelHost` name would be clearer, but the current name
  is retained until `workbench` and downstream users can migrate together.
- `selector.SelectWidget` and `combobox.ComboboxWidget` are explicit enough to
  distinguish their interaction models. Do not merge them into a generic
  `SelectWidget`.
- Packages with constructors named `New` are consistent with the repository's
  component convention; related constructors keep descriptive names such as
  `NewTextArea`, `NewCard`, and `NewCloseButton`.

### File names

The following snake_case names are understandable but inconsistent with the
newer one-concept-per-file naming used elsewhere:

| Current file | Suggested future name | Action now |
| --- | --- | --- |
| `widget/statusbar/status_bar.go` | `statusbar.go` | Defer; public package is already being reorganized. |
| `widget/titlebar/title_bar.go` | `titlebar.go` | Defer; preserve current diff stability. |
| `widget/input/input_group.go` | `group.go` | Defer; the group API is still being split. |
| `widget/colorpicker/color_picker.go` | `picker.go` | Defer; related color files currently use the longer prefix. |
| `widget/datepicker/date_field.go` | `field.go` | Defer; avoid colliding with other date-picker internals. |
| `widget/datepicker/date_range_picker.go` | `range_picker.go` | Defer; rename together with its public constructor plan. |
| `widget/navigation/pagination_visual.go` | `pagination_default_visual.go` | Defer; visual/default split is still in progress. |
| `widget/surface/card_visual.go` | `card_default_visual.go` | Defer; keep the visual cache implementation stable during refactor. |

These are cleanup candidates, not deletion candidates. A file should only be
removed when its declarations have moved and `rg` confirms there are no
remaining imports or references.

## Deletion rule for the next refactor step

Before deleting a component or file, require all of the following:

1. Its public constructor and type have no references outside the package.
2. Its behavior is covered by an existing replacement with the same semantics.
3. A repository-wide `go test ./...` and `go vet ./...` pass after removal.
4. If the name is public, provide a deprecation alias or document the breaking
   migration instead of silently removing it.
