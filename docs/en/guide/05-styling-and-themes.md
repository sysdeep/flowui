# 05 - Styling and themes

`style.Style` is an immutable declaration of box geometry and appearance. Start
with its zero value and chain the properties you need:

```go
primary := style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill)).
	TextColor(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].OnFill)).
	Radius(8).
	Padding(6, 12).
	Cursor(style.CursorPointer)
```

Apply it to a widget:

```go
button.New("Save").Style(primary)
layout.Box(child).Style(style.New().Padding(24))
```

There is no separate builder step. Each method returns a new style value.

## Cascade order

The effective style is resolved in this order:

```text
component defaults
  -> inherited text
  -> variant
  -> size
  -> resolve.Scope ancestors
  -> instance Style
  -> transitions
```

Each layer applies its own matching state rules before merging; later layers
override earlier layers. Within a layer, rules apply in ascending `Rank`, required
state-bit count, then declaration order. Use `On` with state masks for runtime
states (`On` requires all bits, `OnAny` requires any bit, and `OnNot` excludes bits):

```go
primary := style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill)).
	TextColor(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].OnFill)).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Hover))).
	On(style.Pressed, style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Pressed))).
	On(style.Disabled, style.New().Opacity(0.5))
```

`resolve.Scope` supplies defaults to a subtree without changing the global theme.
It lives in `github.com/qianniancn/flowui/style/resolve`, together with the other
context-bound resolvers — `style` itself is a pure declaration mechanism, so
anything that needs the active context or theme values sits beside it.
`Part` targets a named part inside a compound control, such as
`PartContent`, `PartTrack`, or `PartPanel`.

## Button variants

Button semantic colors, hover/pressed/disabled feedback, and focus rings are
handled by the built-in variants. Choose the action hierarchy with `Variant`;
geometry remains owned by the button component:

```go
button.New("Save").Variant(button.Primary)
button.New("Cancel").Variant(button.Secondary)
button.New("Delete").Variant(button.Danger)
button.New("More").Variant(button.Ghost)
```

Add a local declaration on top of the selected variant when needed:

```go
button.New("Delete").Variant(button.Danger).Style(style.New().Radius(4))
```

## Tokens and colors

Prefer semantic theme tokens so the same style works in light and dark themes:

```go
style.New().Background(style.Solid(activeTheme.Colors.Surface))
style.New().TextColor(style.Solid(activeTheme.Colors.Foreground))
style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill))
```

Use literal colors when a design requires them:

```go
style.RGB(0x0078d4)          // opaque 0xRRGGBB
style.RGBA(0x00000030)       // 0xRRGGBBAA
style.WithAlpha(activeTheme.Colors.Ring, 0.5)
```

## Gradients and custom drawing

When a custom widget needs a gradient, resolve it against the active theme and
reuse the public drawing helpers:

```go
brush, ok := resolve.Brush(style.LinearGradient(
	style.ColorStop(0, style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill)),
	style.ColorStop(1, style.Solid(activeTheme.Colors.Tones[theme.ToneDanger].Fill)),
))
if ok {
	draw.Fill(gtx, image.Rect(0, 0, 160, 32), 6, brush)
}
```

Use `resolve.Color` for a single theme-resolved color and `draw.FillRRect` when
you already have a `clip.RRect`. This keeps theme resolution in one place and
does not turn custom drawing into an overlay system.

## Themes

Import `github.com/qianniancn/flowui/theme` for the default theme.
Use `theme.Light()` / `theme.Dark()` directly, or customize its seed and metrics:

The default design uses Slate greys, Indigo accents, 1dp borders and 32dp
standard controls. See [Default theme and external styles](18-design-styles.md)
for screenshots and geometry. `Light` / `Dark` supply the default Indigo accent;
`New` / `NewDark` use the supplied seed colours.

```go
seed := theme.DefaultSeed()
seed.Primary = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}

ui.Run(ui.NewProgram(Model{}, Update, View),
	window.Theme(theme.NewDark(seed)),
)
```

Adjust component-specific geometry and appearance through the widget's instance
style. The global theme contains shared semantic colors, typography, shadows,
and motion.

For an `Application`, use `SetTheme` and `SetLanguage` so the replacement is
applied on the target window's event loop.

## Fonts

The default theme uses `sans-serif` and `monospace` and permits Gio's system
font fallback. This is convenient for multilingual text, but the system font
index can add startup and memory cost. For deterministic rendering, parse a
bundled TTF, OTF, or TTC and disable system fonts:

```go
// import "gioui.org/font/opentype"
//go:embed fonts/Inter-Regular.ttf
var interRegular []byte

faces, err := opentype.ParseCollection(interRegular)
if err != nil {
	panic(err)
}

t := theme.Light()
t.Typography.Typeface = "Inter"
t.Fonts.Collection = faces
t.Fonts.SystemFonts = false
```

Prepare the collection before opening windows. Shared face data can be reused,
while each window receives its own text shaper. If system fallback is disabled,
include every language and symbol the application needs. Supply the font file
referenced by the snippet in your own application.

## Geometry and transitions

```go
style.New().Width(200).Height(40)
style.New().MinWidth(100).MaxWidth(400)
style.New().Padding(12).Margin(8).Radius(8)
style.New().BorderWidth(1).BorderColor(style.Solid(activeTheme.Colors.Border))
style.New().BorderBottom(1, style.Solid(activeTheme.Colors.Border))
style.New().BoxShadow(0, 6, 18, 0, style.RGBA(0x00000030))
style.New().Opacity(0.9)
style.New().Overflow(style.OverflowHidden)
```

Use `Transition` for state-driven properties and give the interactive widget a
stable key:

```go
style.New().Background(style.Solid(activeTheme.Colors.Surface)).
	Transition(style.PropBackgroundColor, 120*time.Millisecond).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Colors.SurfaceRaised)))
```

Continue with [Animation](12-animation.md).

## Instance styles vs. theme appearance

Each widget owns its default appearance, using light or dark semantic tokens
from `theme`. `.Style(...)` changes one instance; shared semantic colors affect
widgets that consume them.

```go
button.New("Save").Style(primary).Variant(button.Primary)
```
