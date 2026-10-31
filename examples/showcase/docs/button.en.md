# Button

Buttons trigger actions such as saving, submitting and cancelling. Pass text directly to the constructor; use `Content` for icons or composed content.

## Basic Usage

```go
import "github.com/qianniancn/flowui/widget/button"

button.New("Save").OnClick(func() { send(Save{}) })
```

## Constructor

```go
button.New(label) // Returns button.Widget
```

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `label` | `string` | Required | Button text and accessible name. Pass a string for ordinary buttons. |

Use `.Key("...")` when an explicit identity is needed, such as in lists or dynamically reordered content. Otherwise identity is derived from the construction site.

Chain methods return a configured value; pass that value to the layout. The application supplies callbacks, state and business data.

## Common Patterns

### Loading and Disabled States

Pass model state to the button to prevent duplicate submissions.

```go
button.New("Submit").
    Loading(m.Saving).
    Disabled(!m.Valid).
    OnClick(func() { send(Save{}) })
```

### Icons and Composed Content

`IconOnly` controls sizing. The constructor text still provides the accessible name. Combine the icon package with the `icon` component.

```go
// import "github.com/qianniancn/flowui-icons-lucide"
// import "github.com/qianniancn/flowui/widget/icon"
button.New("Add project").
    Content(icon.New(lucide.Plus)).IconOnly().
    OnClick(addProject)

button.New("Save").Content(
    layout.Row(icon.New(lucide.Save).Size(16), text.New("Save")).Gap(8),
).OnClick(save)
```

### Width and Local Styling

`FillWidth` fills the available width. Configure behavior first, then append appearance overrides with `.Style(...)` when needed.

```go
button.New("Continue").FillWidth().OnClick(next)

button.New("Save").Primary().Width(120).OnClick(save)

button.New("Save").OnClick(save).Style(style.New().Radius(8).Padding(20))
```

## Parameters and Methods

| Parameter / Method | Type | Default | Description |
| --- | --- | --- | --- |
| `Variant(value)` | `button.Variant` | `Primary` | Sets visual emphasis. Supports the eight variants listed below. |
| `Primary()` | No arguments | Already the default | Shorthand for `.Variant(button.Primary)`. |
| `Size(value)` | `button.Size` | `Medium` | Small / Medium / Large; dimensions come from the active theme. |
| `Disabled(value)` | `bool` | `false` | Prevents OnClick and keyboard activation. |
| `Loading(value)` | `bool` | `false` | Shows a loading indicator and prevents clicks. The application controls the loading state. |
| `FillWidth()` | No arguments | Off | Fills the width allowed by the parent. Otherwise width follows content and theme padding. |
| `Width(dp)` | `int` | Content width | Sets a fixed width, constrained by the parent. Takes precedence over `FillWidth()`. |
| `IconOnly()` | No arguments | Off | Uses square icon-button dimensions. Supply the icon through Content. |
| `Content(value)` | `ui.Widget` | Label text | Replaces visible content with an icon or composed layout while retaining the accessible name. |
| `Label(value)` | `string` | Constructor label | Changes only the accessible name, leaving visible content unchanged. |

## Variants and Sizes

| Variant | Type | Default | Description |
| --- | --- | --- | --- |
| `Primary` | `button.Variant` | Default variant | Solid primary color for main actions such as saving or submitting. |
| `Secondary` | `button.Variant` | — | Neutral background for secondary actions. |
| `Tertiary` | `button.Variant` | — | Subtle background for low-emphasis actions. |
| `Outline` | `button.Variant` | — | Outlined style for cancellation or secondary actions. |
| `Ghost` | `button.Variant` | — | Lightweight style for toolbars and supporting actions. |
| `Danger` | `button.Variant` | — | Solid danger color for destructive actions such as deletion. |
| `DangerSoft` | `button.Variant` | — | Soft danger-colored background. |
| `DangerOutline` | `button.Variant` | — | Danger-colored outline. |

Default theme heights: **Small 24dp, Medium 32dp, Large 40dp**. `IconOnly` uses equal width and height. Custom themes and instance styles can adjust these dimensions.

## Click Events and Keyboard

| Method | Type | Default | Description |
| --- | --- | --- | --- |
| `OnClick(fn)` | `func()` | `nil` | Runs on mouse, touch or keyboard activation. Does not run while disabled or loading. No business action runs without a callback. |

```go
button.New("Save").OnClick(func() { send(Save{}) })
```

Tab / Shift+Tab moves focus. Enter or Space activates a focused button. Give icon-only buttons a clear action name too.

`OnClick` runs synchronously during UI updates. Run slow operations asynchronously, update the model through messages, then pass the new state to `Loading` / `Disabled`.

## Button Groups

```go
button.Group(
    button.New("Previous").OnClick(previous),
    button.New("Next").OnClick(next),
).Variant(button.Outline).
    Size(button.Small).
    Separators(true)
```

| Parameter / Method | Type | Default | Description |
| --- | --- | --- | --- |
| `Group(buttons...)` | `...button.Widget` | Empty group | Combines buttons. Each button still needs its own identity and callback. |
| `Orientation(value)` | `button.GroupOrientation` | `GroupHorizontal` | GroupHorizontal arranges horizontally; GroupVertical arranges vertically. |
| `Variant(value)` | `button.Variant` | `Primary` | Default variant for buttons in the group. |
| `Size(value)` | `button.Size` | `Medium` | Default size for buttons in the group. |
| `Disabled(value)` | `bool` | `false` | Default disabled state for buttons in the group. |
| `FillWidth()` | No arguments | Off | Fills the available width. Horizontal groups divide the width equally. |
| `Separators(value)` | `bool` | `false` | Shows separators between adjacent buttons. |

Explicit `Variant`, `Size` and `Disabled` settings on individual buttons override group settings. For example, `Disabled(false)` overrides a disabled group. Use Segmented or RadioGroup for mutually exclusive choices.

## Usage Notes

- A key identifies the button. Use a stable record ID in lists rather than click counts or frequently changing text.
- `Loading` does not start or finish a task automatically. It can display a loading state even without `OnClick`.
- `Content` changes visible content only. Update `Label` if the action's meaning also changes.
- Ordinary button text does not need `text.New`; reserve `Content` for icons and composed layouts.
- Define `m`, `send`, `Save`, callbacks and model fields in your application. Buttons do not manage business state.
