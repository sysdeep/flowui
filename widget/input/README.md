# Input

This package provides one built-in appearance for Input, TextArea, InputGroup,
and GroupAction. The light or dark theme supplies shared colors, dimensions,
and motion values.

Input and TextArea use `widget/base.TextEditor` for editing behavior.
Input, TextArea and InputGroup use `widget/base.FieldInteraction` for shell
hover and pointer-to-editor focus routing, including action-slot exclusions.
GroupAction uses `widget/base.Button`. InputGroup composes the same editors
inside a field shell with prefix and suffix slots. Use the base behaviors
directly when building a different appearance.

| File | Responsibility |
| --- | --- |
| `input.go`, `text_area.go` | Public editor APIs and behavior composition |
| `input_group.go`, `input_group_action.go` | Compound field and action APIs |
| `default.go` | Built-in appearance for all variants, sizes, parts, and states |
| `style.go` | Theme default caches and instance style resolution |
| `state.go` | Value bindings, pending edits, and field interaction state |
| `editor.go` | Gio editor rendering and accessibility semantics |
| `layout.go`, `input_group_layout.go`, `input_group_draw.go` | Field layout and divider drawing |
| `context_menu.go` | Edit actions and editor context menus |
| `types.go` | Public variant, size, and input type enums |

`Style(...)` resolves after the built-in defaults. Instance declarations stay
outside the default cache. TextArea and multiline InputGroup retain both the
initial style measurement and the height derived from rows in their caches.

`widget/internal/fieldstyle` shares field appearance rules with DatePicker. It is an
internal implementation helper with one fixed border and feedback design;
each component owns its geometry, parts and layout.
