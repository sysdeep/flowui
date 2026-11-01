# 04 - Layout

FlowUI separates box styling from child arrangement:

| Owner | Controls |
| --- | --- |
| `Style` | Width, height, min/max, padding, margin, overflow, and paint |
| Layout container | Direction, gaps, alignment, flex growth, scrolling, and measurement |

Do not use a style to express sibling spacing or flex growth. Put that policy
on `Row`, `Column`, or another container.

## Common containers

```go
layout.Column(
	text.New("Title").Size(20),
	layout.Row(
		button.New("Cancel"),
		button.New("Save"),
	).Gap(8),
).Gap(12)
```

`Box` gives one child a styled shell. `Center` centers its child in the
available constraints. `Expanded` and `Flexible` distribute remaining space.
`Row` and `Column` also accept `Width`, `Height`, `FillWidth`, `FillHeight`,
`MaxWidth`, and `Padding`. `Padding` follows CSS shorthand: one value for all
sides, two for vertical/horizontal, three for top/horizontal/bottom, or four
for top/right/bottom/left.

```go
layout.Row(
	sidebar.New("home", items).Key("navigation"),
	layout.Expanded(layout.Scroll(layout.Column(/* content */))),
)
```

## Scrolling and grids

`Scroll` requires a stable key. `Grid` uses a fixed column count, while
`AutoGrid` chooses columns from a minimum column width.

```go
layout.Scroll(layout.Column(/* many children */).Gap(8))

layout.Grid(3, cardA, cardB, cardC).Gap(16)
layout.AutoGrid(190, cardA, cardB, cardC).ColumnGap(16).RowGap(12)
```

`Scroll` reserves a scrollbar gutter only when its content overflows. On a
page whose height changes, such as settings tabs, switch to an overlay
scrollbar so the content width stays put when a short page follows a long one:

```go
layout.Scroll(content).Overlay()
// Hide the bar and gutter while keeping wheel and touch scrolling.
layout.Scroll(content).WithoutScrollbar()
```

For an appended chat or log, use `.FollowEnd()` to start at the end and keep
following while the user remains there. Scrolling away pauses following until
the user returns to the end.

`layout.Scroll` is the single viewport entry point; `layout.Scrollbar` was
removed. `.Style(...)` merges root and `PartTrack` / `PartThumb` overrides.
Changing scrollbar mode does not replace or reset the scroll identity.

### Visual outset and clipping

`layout.Scroll`, `layout.List`, and `splitpane.New` automatically reserve space
inside their viewport for shadows and outlines declared by `Style`. Components
with the common Box shell, such as `Input` and `Card`, need no extra margin.
The space is applied at viewport edges, not between every pair of list items.
An explicit `OverflowHidden` box remains a local clip boundary; child paint
already clipped there does not affect an outer viewport.

For custom paint that extends beyond its box, declare a minimum range with
`style.VisualOutset`, or use `layout.VisualOutset` when there is no Style shell:

```go
layout.Box(customCanvas).Style(
	style.New().VisualOutset(6, 8, 10, 8),
)

// top, right, bottom, left for direct custom paint.
layout.VisualOutset(ctx, gtx, customCanvas, 6, 8, 10, 8)
```

When a new extent is discovered, the viewport requests another frame to settle
the safe padding. It never lets that paint escape the scroll viewport or window.

Run `go run ./examples/showcase` from the repository root for a runnable
layout composition. The grid code above is a usage snippet.

Other useful containers include `Wrap`, `Stack`, `Surface`, and `Card`.
The themed resize control is `splitpane.New` in `widget/splitpane`; visual
separators are `divider.New` in `widget/divider`. See the [component guide](06-components.md) for the full list.

## Size and constraints

Declare dimensions on a style and apply it to a box or component:

```go
layout.Box(input.New(m.Name)).Style(style.New().Width(320))

layout.Box(child).Style(
	style.New().Width(200).
		Height(40).
		MinWidth(120).
		MaxWidth(400).
		Padding(12).
		Margin(4),
)
```

Use `AspectRatio` for a fixed width-to-height ratio:

```go
layout.AspectRatio(1, layout.Box(text.New("Square content")))
```

Parents provide maximum constraints. Children choose a size inside those
constraints, and containers decide how children are placed.

## A typical page shell

```go
func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		titlebar.NewPlatform("FlowUI", menu).Key("title"),
		layout.Expanded(layout.Row(
			sidebar.New(m.Selected, items).Key("sidebar"),
			layout.Expanded(layout.Scroll(content)),
		)),
		statusbar.New(left, right),
	)
}
```

Build widgets every frame, keep keys stable, and use business IDs for list
items whose order can change.

