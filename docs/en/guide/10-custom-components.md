# 10 - Custom components

Start with composition. Only drop to custom layout or drawing when the public
components and styles cannot express the behavior.

## Path A: composition

Build a reusable function from existing widgets and return a value-like widget:

```go
func StatusPill(label string, selected bool) ui.Widget {
	value := style.New().Padding(12, 6).Radius(999).
		On(style.Hovered, style.New().Background(style.Solid(activeTheme.Colors.SurfaceTertiary))).
		On(style.Selected, style.New().Background(style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill)))
	return layout.Box(text.New(label)).Style(value)
}
```

Composition keeps interaction, semantics, and theme behavior in the existing
components.

## Path B: custom interactive layout

For a new interactive control, retain state with a stable key, resolve a style,
and use the public layout helpers. Keep the widget's event host around the
whole interactive area and leave painting to a small child widget.

The public sequence is:

```text
UseState -> resolve.Resolve / resolve.ResolvePart ->
layout.InteractiveResolved (or layout.Resolved)
```

The resolvers live in `github.com/qianniancn/flowui/style/resolve`, not in `style`
itself — `style` is a pure declaration mechanism with no framework dependency.
The theme supplies values without importing `style`; widgets own appearance. `Resolve` and `ResolvePart`
resolve and apply transitions in one call. A component that also resolves during
measurement uses `ResolveStatic` / `ResolvePartStatic` plus `HasTransitions` and
`ApplyTransitions` / `ApplyPartTransitions` instead, so transition state is
applied exactly once per frame; both call shapes share the same widget key.

The snippets here illustrate the public extension APIs. Do not use
internal component packages; they are implementation details rather than
application APIs.

Use an explicit widget type for a complete custom control. `ui.WidgetFunc` is useful when only a
small layout callback is needed.

## Path C: custom drawing

Use Gio operations for the content of a chart or illustration, but keep the
outer box, hit target, focus behavior, and theme tokens in FlowUI components.
Canvas code should not duplicate button or field semantics.

## Path D: headless base

When the structure itself must differ — a whole-row settings card, a gauge
instead of a track, a custom trigger — but pointer, keyboard, focus and
accessibility must stay correct, use `widget/base`:

```go
base.Toggle("sync", model.Sync).
	Label("Sync files").
	OnChange(func(next bool) { send(SetSync{next}) }).
	Style(style.New().FillWidth().Padding(14).Radius(12).
		Background(style.Solid(activeTheme.Colors.SurfaceSecondary))).
	Content(func(state base.ToggleState) ui.Widget {
		return layout.Row(
			indicator(state.Checked, state.Hovered),
			text.New("Sync files"),
		)
	})
```

`base.Toggle` and `base.Slider` own behavior, state and semantics and paint
nothing. `Content` receives the interaction state (`Checked`, `Hovered`,
`Dragging`, `FocusVisible`, and so on) and draws the control. `Content` is
required and `Style` sizes the interaction box; the model still owns the value
and `OnChange` reports intent.

The styled `Checkbox` and `Slider` remain available. Bases are for design
systems that change the structure too; when only colors, radii or shadows
change, Path A and theme tokens are enough.

For common drawing and measurement tasks, reuse the public helpers instead of
duplicating theme resolution or Gio event boilerplate:

```go
brush, ok := resolve.Brush(style.LinearGradient(
	style.ColorStop(0, style.Solid(activeTheme.Colors.Tones[theme.TonePrimary].Fill)),
	style.ColorStop(1, style.Solid(activeTheme.Colors.Tones[theme.ToneDanger].Fill)),
))
if ok {
	draw.Fill(gtx, image.Rect(0, 0, 160, 32), 6, brush)
}

size := text.Measure(ctx, gtx, text.New("Preview").Size(14).MaxLines(1))
_ = size
```

For intrinsic sizing of an arbitrary child, implement the optional
`ui.Measurable` interface or call `ui.MeasureWidget`. The measurement pass
uses a separate ops list and an empty input source, so it cannot consume
pointer or keyboard input:

An implementation of `Measure` must only report dimensions; it must not read or
consume input from the current frame.

```go
type measuredBadge struct{}

func (measuredBadge) Measure(_ *ui.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(gtx.Dp(48), gtx.Dp(20))}
}
```

Low-level pointer widgets can use `AddPointerArea`, `NextPointerEvent`,
`IsPrimaryPointerPress`, and `GrabPointer`. Use `layout.VisualOverflow` for a
local focus decoration that must cross the child's own clip. It is
not a replacement for `Popover`, `Portal`, or another overlay host. When that
decoration also needs room inside an ancestor scroll or split viewport, report
its top/right/bottom/left extent with `layout.VisualOutset`; a component with a
Style shell can use `style.VisualOutset` instead.

## Rules of thumb

1. Use a public component before copying its implementation.
2. Keep business state in `Model` and interaction state under a stable key.
3. Resolve inherited and state styles rather than hard-coding colors.
4. Give custom controls keyboard focus and semantic labels where appropriate.
5. Test layout and pointer behavior with `uitest.Harness`.

Continue with [Multiple windows](11-multiple-windows.md).

