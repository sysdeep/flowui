# Button

This package provides one built-in appearance for buttons, button groups, and
close buttons. Interaction, focus, keyboard activation, and accessibility are
provided by `widget/base.Button`.

For a completely custom appearance, compose `widget/base.Button` directly.
This package has no appearance provider, recipe, or renderer protocol.

## Files

| File | Responsibility |
| --- | --- |
| `button.go` | Button API and composition with `base.Button` |
| `group.go` | Group options, shared layout, and corner placement |
| `close.go` | Close button API and default close icon |
| `default.go` | Final built-in styles for all variants, sizes, and states |
| `style.go` | Default-style caches and the shared style cascade |
| `layout.go` | Content measurement, prepared group content, and loading layout |
| `draw.go` | Loading indicator and group separator drawing |
| `types.go` | Public variant and size enums |

`Style(...)` applies an instance declaration through the shared style cascade.
It does not replace the component's drawing implementation. Default-style caches
depend on the active theme and component options; instance styles are resolved
separately.
