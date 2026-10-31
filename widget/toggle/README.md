# Toggle

This package provides one built-in appearance for Checkbox, Switch,
ToggleButton, Segmented, and SwitchGroup. The active light or dark theme
supplies shared colors, dimensions, and motion values.

Checkbox and Switch compose `widget/base.Toggle`; ToggleButton composes
`widget/base.Button`. Segmented composes toggle buttons and SwitchGroup lays
out switches. Use these base behaviors directly for a custom appearance.

| File | Responsibility |
| --- | --- |
| `checkbox.go`, `switch.go`, `togglebutton.go` | Public component APIs and composition |
| `segmented.go`, `group.go` | Compound components and group layout |
| `default.go` | The built-in appearance for every component, size, and state |
| `*_style.go` | Default-style caches and instance style resolution |
| `*_layout.go`, `*_draw.go` | Content layout and component drawing |
| `state.go` | Appearance animation and cache state |
| `types.go` | Public variant and size enums |

`Style(...)` applies instance declarations after the built-in defaults.
Variants and sizes are options within this appearance. They do not select
another theme or renderer.

`DrawCheckboxControl`, `CheckboxVisualCache`, and `CheckboxSelectionAnimation`
also serve compound widgets such as Table, so their checkbox indicators share
the same built-in appearance without creating another interaction target.
