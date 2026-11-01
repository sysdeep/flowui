# 02 - Core concepts

FlowUI keeps three responsibilities separate:

| Responsibility | Owner |
| --- | --- |
| Business state and side effects | `Model`, messages, and `Update` |
| Box geometry and appearance | `Style` and component variants |
| Child measurement and placement | Layout containers such as `Row`, `Column`, and `Grid` |

## Model, View, and Update

- **Model** is application data.
- **Msg** is a user intent or an asynchronous result.
- **Update** changes the model and can return a `Cmd`.
- **View** reads a model value and returns a `Widget`.

```go
type Model struct {
	Name string
}

type NameChanged struct{ Value string }

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Value
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		input.New(m.Name).OnChange(func(value string) {
			send(NameChanged{Value: value})
		}),
		text.New("Hello, "+m.Name),
	).Gap(12)
}
```

View code must not mutate the model. Event callbacks send a message; the next
Update call applies it.

## Keys

Interactive widgets retain Gio state between frames by key. Constructors such
as `button.New` and `input.New` provide an automatic identity for fixed layouts:

```go
input.New(m.Email)
button.New("Save")
```

Automatic identity depends on component type, constructor call site, enclosing
key scope, and the order in which widgets from that site first resolve their
keys in a frame. Each window tracks this independently. Prebuilt widgets and
widgets rebuilt each frame follow the same rule; visible text is not an identity.

For dynamic lists, conditional siblings created by a shared helper, and widgets
that move between positions, set an explicit business key:

```go
input.New(item.Name).Key(item.ID)
```

Use `layout.Key(item.ID, child)` to give an entire repeated subtree a stable
scope. A normal `Row` or `Column` does not introduce a key scope. Containers
such as `table.New` generate automatic root keys; row and item keys still
identify business data.
Composite and lifecycle-host components such as `menu.ContextMenu`,
`dropdown.Button`, `toolbar.Commands`, and `surface.NewPanel` keep explicit
keys because they own child scopes or retained host state. Root hosts such as
`menubar.New`, `toast.New`, and `titlebar.New` generate automatic identities;
use `.Key(...)` when their identity must remain stable across construction sites.

Low-level primitives without a constructor token can fall back to an identity
based on state kind and position within the scope when passed an empty key.
This is a positional fallback, not a business identity.

Changing an explicit key starts a new interaction identity. Reusing the same key
for two widgets in one scope is an error. Keys are scoped runtime identities,
not persistent IDs for storage. The automatic token format is internal;
application keys must not use its reserved NUL-delimited `\x00@` marker.

## Context

`*ui.Context` provides frame services such as theme snapshots, language,
focus, transient widget state, and invalidation. It is not a business data
store. Keep application data in `Model`; use `ui.UseState` only for short-lived
interaction state owned by a widget.

## Widgets and layout

Widgets are value-like declarations rebuilt on each frame. Layout containers
measure and place their children. Styles describe the child's box and visual
properties; they do not replace layout policy.

See [Layout](04-layout.md) and [Styling and themes](05-styling-and-themes.md)
for the two halves of rendering.
