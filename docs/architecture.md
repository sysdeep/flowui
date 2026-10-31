# FlowUI Architecture

Implementation notes for the current tree: dependency direction, style cascade,
interaction host, state ownership, overlay host, and effects. For application
usage, see the [project documentation](https://qianniancn.github.io/flowui/) and the public `ui` package docs.

This document is the source of truth for the current component boundaries,
style cascade, interaction host, and migration results.

The public `i18n` package owns languages, message keys, and catalogs.
`internal/locale` handles platform language detection and framework integration;
DatePicker's date-format locale remains a separate domain concern.

## One-minute orientation

Read this before the sections below; it is the map the rest of the document
fills in.

**What an application imports.** Three packages, and nothing else is required:

```go
import "github.com/qianniancn/flowui/ui"             // MVU kernel + lifecycle
import "github.com/qianniancn/flowui/widget/button"  // one package per component
import "github.com/qianniancn/flowui/window"         // native window options
```

`layout` joins them for composition (`layout.Column`, `layout.Row`), and
`style`/`theme` when an application declares its own look.

**The responsibilities, and where each lives.**

| Responsibility | Package | Shape |
| --- | --- | --- |
| Business state, messages, model updates | the application | `Program[M, Msg]`, `Update`, `View` |
| The MVU kernel and per-window services | `ui` | `Context`, `Cmd`, `Subscription`, `Application` |
| Reusable component behavior | `widget/base` | Input, focus, state, and accessibility semantics without painting |
| A complete component and its domain layout | `widget/<component>` | Constructors, behavior integration, and layout |
| Default component appearance | `widget/<component>` | Private default style declarations and focused tests |
| Shared theme tokens | `theme` | Semantic values and light/dark seed constructors |

`ui` is deliberately small: it re-exports the MVU runtime and the frame
contract, and it contains **no components**. Components do not go through a
`ui` facade — `widget/button` is imported directly. That is the single biggest
difference from the v1 layout, where `ui` re-exported every component.

**The foundation, by responsibility.**

| Package | Responsibility |
| --- | --- |
| `internal/core` | Interaction kernel and the shared Widget contract |
| `theme` | Semantic colors, typography, shadows, and shared motion policy |
| `style` | Declarations, cascade tables, and hashing; a pure leaf |
| `layout` | Box, flex, grid, stack, scroll composition, and the box host |
| `widget/text` | The text primitive used by label-like components |
| `animation` | Easing, tween, spring, timeline, and layout animation |
| `internal/computed` | Style resolution, transitions, and visual caches |
| `draw` | Brushes, borders, shadows, and paths |

`internal/computed` uses one static resolution path for roots and named parts.
Transitions are applied after resolution; shared snapshots are copied only when
animation needs to mutate them. Component-owned declarations and visual results
use `VisualCache` (or `VisualCacheMap` for multiple keys), not separate declaration
cache machinery. The bounded cascade cache separately reuses merged style results.
These mechanisms do not own default component appearances.

**Where to go next.** Dependency rules and the pure-leaf constraint are in
[Dependency Direction](#dependency-direction). How a declaration becomes
pixels is in [Style Cascade](#style-cascade). The frame protocol a custom widget
must follow is in [Interaction Core](#interaction-core) and
[Overlay Host](#overlay-host). Effects (commands, subscriptions, cancellation)
are in [Effects](#effects).

FlowUI exposes a small set of responsibility-oriented application packages:

```go
import "github.com/qianniancn/flowui/ui"
```

`ui` owns MVU, application lifecycle, and the public `Context`/`Widget`
contracts. Official components are in `widget/<component>` (one package per
component, for example `widget/button`); layout composition is in
`layout`; style declarations are in `style`, and each widget package owns its
own default appearance; low-level drawing is in `draw`; reusable motion
primitives are in `animation`; native window options are in `window`.

Optional platform services live in focused public packages: `explorer` for
window-bound native file dialogs, `notify` for native desktop notifications,
and `systray` for tray icons and menus. They do not provide a second component
facade and do not expose `internal` implementation types.

Component tests may additionally import `github.com/qianniancn/flowui/uitest`.
The package owns a deterministic FlowUI context, Gio input router, viewport,
and clock. Its `Harness.Frame` method follows the same main layout, root
overlay, focus-command, state-cleanup, and router order as a real window.
Applications do not depend on `uitest` at runtime.

`uitest.AppHarness` drives the production message queue, `Update`, command
execution, error delivery, and shutdown cancellation without opening a window.
It complements the widget Harness rather than maintaining a second MVU runtime.

Applications should not import `internal` packages, and internal implementation
packages must not import `ui`.

## Directory Ownership

Keep public import paths stable when reorganizing files. The current package
boundaries separate behavior from appearance; the number of directories is not
itself a reason to introduce another grouping layer.

| Location | Ownership |
| --- | --- |
| `widget/<component>` | Public components, private default appearance, and focused tests |
| `widget/base` | Headless behavior, focus, and semantics for custom designs; paints nothing itself |
| `widget/navigation` | Navigation widgets and public keyboard-navigation helpers |
| `widget/internal/binding` | Generic controlled/uncontrolled value state shared by widgets |
| `widget/internal/optionrow` | Shared row animation and drawing; interaction state remains caller-owned |
| `widget/internal/textrender` | Shared Gio text adapters for components |
| `widget/code/internal/highlight` | Code-block syntax tokenizer |
| `widget/internal/navigation` | Keyboard navigation algorithms shared by components and navigation helpers |
| `widget/internal/overlayarrow` | Popup arrow geometry and drawing |
| `widget/internal/selectablemenu` | Selectable-text and context-menu integration hook |
| `widget/chart` | Shared chart types, annotations, scales, and interaction helpers; several types appear in public chart APIs |
| `widget/internal/fieldstyle` | Shared built-in field appearance and variant vocabulary; components supply geometry |
| `examples/` | Runnable applications in the root Go module; keep example-only assets beside their example |
| `docs/assets/` | Images used by the documentation |

Repository tests use the production light/dark themes and configure any
test-specific motion timings locally. Default appearance tests live with the
widgets that own the declarations. See the
[theme package map](https://github.com/qianniancn/flowui/blob/main/theme/README.md)
for token ownership.

Each widget owns its default geometry and appearance; applications customize an
instance through `Style` and part declarations. Shared colors, typography,
shadows and motion remain in `theme`.

## Dependency Direction

```text
application
    +--> ui (MVU entry point and runtime)
    |      +--> internal/runtime, internal/explorer
    |      +--> internal/core
    |      +--> window, theme
    +--> widget/<component> (official component packages)
    |      +--> internal/core, internal/computed, theme, internal/locale
    |      +--> layout, draw, animation, style
    |      +--> other widget/<component> packages (peer imports are allowed)
    +--> layout (composition and box host)
    |      +--> internal/core, internal/computed
    |      +--> style, theme, draw, animation
    +--> style (pure leaf: declarations, cascade tables, hashing)
    +--> style/resolve (context- and theme-bound resolvers)
    |      +--> style, internal/computed, internal/core, draw
    +--> theme (light/dark palettes, metrics, typography, and shared tokens)
    +--> draw
    |      +--> theme
    +--> animation (replay and easing primitives)
    |      +--> internal/core, theme
    +--> window (native window configuration)
    |      +--> internal/locale, theme
    |
    +--> explorer --> internal/explorer
    +--> notify   --> gioui.org/x/notify
    +--> systray  --> platform-native implementations

component tests
    |
    +--> uitest
            |
            +--> ui
            +--> internal/core
            +--> theme
```

The diagram names the edges a reader is likely to ask about, not every edge.
Two consequences are worth stating because they are easy to get wrong:

- **`ui` and `widget/<component>` do not depend on each other.** They meet at
  the `Widget` contract, which lives in `internal/core` and is re-exported as
  `ui.Widget`. So `widget/button` never imports `ui`, and adding a component
  cannot create a cycle through the MVU entry point.
- **`widget/<component>` packages may import each other.** `widget/collapsible`
  builds on `widget/button`; `widget/menubar` on `widget/menu`. Peer imports are
  a normal part of the layer — and so is importing `layout` to arrange their
  own parts. Neither edge points upward:
  no component imports `ui`, so the layer stays below the MVU entry point even
  though it is free to compose with its siblings.

Dependencies only point downward. Public packages **are** the implementation:
the twin-layer facades were removed, so `style` and `theme`
declare their types directly instead of re-exporting `internal/style` /
`internal/theme`. Applications refer to the shared frame contract through the
public `ui.Widget` alias; component packages use the same underlying contract
without importing `ui`.

`style` must stay a **pure leaf** — it may not import `core`, `draw`, or
`internal/computed`. The reason is not an indirect path but a **direct edge**:
`internal/core`, `internal/computed`, `widget/internal/fieldstyle` and `layout` all import
`style`, so any such import would close an immediate 2-cycle with no room to
route around it. The constraint is structural, not an optimization to revisit.
That is why context- and theme-bound resolvers live in the separate
`style/resolve` package.

Structural **style and render data types** that applications read as plain data
(`ResolvedStyle`, `PaintSource`, `ColorSource`, `SolidColor`, `Gradient`,
`Insets`, `BoxStyle`, `PaintStyle`, `CornerRadii`, `Shadow`,
`OutlineStyle`, `BorderStyle`, `TextStyle`, `TransformStyle`, `Transition` in
`style/style.go`; `ShadowCornerRadii`, `ShadowShape`, `ShadowLayer`, `BoxShadow`
in `draw/shadow.go`) are owned by the package that defines them, so their fields
stay directly usable. `style` declares them natively and keeps exactly one
alias, `Cursor = pointer.Cursor`.

Core implementation packages live below `internal/`. The public packages above
are the supported entry points; their internal implementation paths are not
additional application APIs. The focused platform packages expose only their
own stable adapters.

Each widget package owns its default appearance and reads shared semantic
tokens from `theme`. Shared implementation used by related widgets lives under
`internal/`. `theme` provides the default light and dark tokens; generic
ramps and seed derivation remain in `theme`.

Reusable interaction belongs in `widget/base`. Styled RadioGroup, ListBox,
Sidebar and Tabs share `CollectionNavigation` and `RowInteraction`; Collapsible shares
`Disclosure.Update` and collection navigation. Input, TextArea and InputGroup
share `TextEditor.Update`, and GroupAction composes `base.Button`. These adapters
preserve caller-owned clickables and editors so identity, focus requests and
virtualized layout stay stable. Base owns event processing; each component owns
its selection binding, layout, animation, semantics around its parts, and one
private default appearance. `theme` supplies shared tokens rather than component
visual providers.

Toolbar also uses `CollectionNavigation`, supplying arbitrary child focus tags,
orientation and wrapping while children retain activation. Table and Tree use
the same navigation loop with modifier-aware events and `UpdateRow` for click
details. Lazy active-index fallback preserves virtual Table's idle-frame cost.
Table owns sorting, range selection, column sizing and virtualization. Tree owns
parent/child navigation, expansion, lazy loading, checking and drag/drop; its
rename editor uses `TextEditor.Update`, with commit/cancel rules kept in Tree.
Custom hierarchy commands return their active index before the next key is
processed, preserving input order within a frame.

Standalone base Options and RadioGroup use one focus target per group; styled
collections can retain individual row targets. Both use the same `base.CollectionNavigation` event loop and
row activation implementation, with indexed algorithms in `widget/internal/navigation`, which has no dependency on styled widgets. Public
`widget/navigation` helpers delegate to that leaf package to avoid a dependency
cycle through navigation widgets and buttons. Popup, Popover, Dropdown, Submenu,
ContextMenu, Menubar and ColorPicker share
`base.PopupInteraction` for outside/Escape close intent and pointer-only panel,
arrow and outside regions. The overlay host grants policy input to one owner;
components retain open binding, placement, animation and focus-return policy.
Popover dismisses on a completed outside click, while Dropdown dismisses on an
outside press. ContextMenu and Dropdown share `base.PopupTrigger` for secondary
clicks, Shift+F10 and cancellable long presses; each keeps its anchor geometry,
open binding and focus handoff.

Collection navigation has an explicit vertical/horizontal/both-axis policy.
Tabs supplies its automatic/manual activation policy and workbench shortcuts;
scrolling, panel lifetime, rename, close and drag extensions stay in Tabs.
Its inline editor uses `base.TextEditor.Update`, with Escape/blur/commit rules
kept locally. Tabs close buttons and Toast close/action buttons use `UpdateRow`
with `PreserveFocus`. Pagination composes `base.Button` for its page controls.
Select reuses `CollectionNavigation` for highlight movement while retaining
trigger focus and selection binding.
Sidebar keeps its nested open-key model and hover expansion. Menu composes
`base.MenuNavigation`, which adds submenu/root-direction intent to the same
collection keyboard loop, and `RowInteraction` for row pointer behavior. Menu
keeps submenu layout, hover scheduling, binding and action/close callbacks.
Modal composes `base.DialogInteraction` for close intent, close-button input,
focus boundaries and initial/deferred/stacked-scope focus. It reuses the popup
pointer surfaces; focus lifecycle no longer depends on animation progress.
The trailing boundary is mounted by the overlay Tail after nested popups.
Exiting dialogs keep their neutral focus target and blockers until Modal ends
the visual lifetime. Open binding, backdrop, scrolling, sections and animation
remain in Modal. AlertDialog composes Modal with explicit-action defaults and
its own header/status visual, without a second behavior state machine.
Menubar shares `base.MenubarNavigation` for root movement, Home/End, typeahead
and perpendicular-arrow opening, and `RowInteraction` for pointer activation.
Global F10/Alt accelerators and switching open panels stay in Menubar.
ColorPicker retains color conversion and history commits while sharing popup
input and swatch activation. DatePicker shares popup panel focus and
`RowInteractionWithFocus` for calendar cells; date arithmetic, segment editing,
range cancellation and focus-loss dismissal remain component behavior.
Tooltip retains its passive hover/focus timing and exclusive visibility policy;
it has no popup dismissal regions or interactive panel to migrate.

Disclosure Update returns a frame-local trigger host: `LayoutTrigger` owns its
hit target and semantics, and `button.Visual` draws from the supplied state
without a second base interaction. The expandable region remains under the
component's animation lifecycle. Do not retain the Update result across frames.
TextEditor accepts a `TextEditorRequest` and reads enabled state from the layout
context; it returns editor events without modifying that context. TextInput's
Content receives the retained Gio editor and supplies all typography and paint.
Generic value bindings and shared navigation, text, overlay, and selectable-menu
helpers are private to `widget/internal`.

## Style Cascade

`style.Style` is the public immutable declaration owned by package `style`.
Public declarations typically start with `style.New()`, then continue through
value-returning fluent methods such as `.Background`, `.Padding`, and
`.Part`. Components use the same immutable chain; there is no separate builder
or build step. The runtime resolves
declarations in this order (`ResolveStatic` / `Resolve`):

```text
component defaults
  -> inherited text (parent TextDeclaration stack)
  -> variant
  -> size
  -> resolve.Scope ancestors
  -> instance Style
```

`Cascade` applies matching `Rule` rows (later properties win), and optional
transitions run under the component key. Colors are concrete values taken
directly from the active `theme.Theme` at declaration time; the only
resolution-time theme dependency left is shadow-profile expansion.

Rule rows match interaction state through `StateMask` bit masks (the `On` /
`OnAny` / `OnNot` family builds them from `Hovered`, `Pressed`,
`FocusVisible`, `Dragging`, semantic bits like `Disabled`,
`Checked`, `Invalid`, `Loading`). The predicate-based
condition mechanism survives only as an unexported implementation detail
`resolve.Scope` carries declarations down the widget tree; it
never mutates the application theme.

```text
layers -> Cascade(state) -> Transitions(key)
            |
            +-> root + named Parts (CascadePart)
```

Animation is dual-track: Style `Transition` (property whitelist under a stable
key) and imperative `Tween` for floats. Both share `Ease*` curves and honor
`Theme.Motion`. Application-facing animation notes: [动画](guide/12-动画.md).

## Interaction Core

Button-like controls share one internal interaction core for clicks, keyboard
activation, focus, disabled state, and accessibility semantics. `button.New`
adds FlowUI defaults, variants, sizes, loading, and compound content. A custom
control uses `layout.Box(...).Key(...).OnClick(...)`; its interaction state is
available to style rule rows (`On` / `Rule{When: mask}`) without inheriting
Button visuals. Raw Gio events
remain available for interactions such as dragging that do not share button
semantics.

Root `Box`, `Paint`, `Text`, and `Transform` properties always describe the
component's outer box. A compound component exposes internal elements through
`Part`: content, labels, descriptions, icons, tracks, fills, thumbs,
indicators, panels, items, backdrops, placeholders, selections, prefixes, and
suffixes. Compound field controls use `PartContent` for the field face. Part
text inherits root text properties; box, paint, and transform properties do
not. Applications may use a custom `style.Part` name for their own components.

The common renderer owns margin, constraints, aspect ratio, padding, overflow
clipping, cursor, background/gradient, per-corner radius, shadows, border,
outline, opacity, and transforms. It also reports the visual extent of shadows
and outlines to clipping viewports; those viewports reserve internal edge space
without changing normal sibling layout. Transient shadows
(`style.BoxShadowTransient`) are painted but excluded from that reservation, so
state-only decoration such as a focus halo cannot reflow a clipping viewport.
Components must not duplicate a partial
adapter for these properties. They retain only domain-specific geometry such as
progress ratios, slider thumb positions, or chart paths.

Components use the same runtime for root layout and interaction state. A custom
widget owns a stable key and transient interaction state with `ui.UseState`,
resolves its root and parts with `resolve.Resolve` and `resolve.ResolvePart`,
then renders with `layout.Resolved` or
`layout.InteractiveResolved`. The interactive helper keeps margin
outside the input host while making padding and the remaining visual box part
of the hit area. `resolve.Scope` styles descendants. Component-level theme
mutation is intentionally not part of the API.

The themed resolvers live in `github.com/qianniancn/flowui/style/resolve`. The
`style` package itself is a pure declaration mechanism with no framework
dependency. `theme` supplies values without importing `style`; widget packages
own their default appearance. The context-bound resolver facade lives in
`style/resolve` instead.

Custom drawing can convert paint sources into brushes with `resolve.Brush`
and resolve concrete colors with `resolve.Color`, draw them with `draw.Fill`,
and measure text with `text.Measure`. Low-level pointer controls use Gio's pointer
operations directly. `layout.VisualOverflow`
is for local decorations such as a focus ring. Use
`layout.VisualOutset` (or `style.VisualOutset` on a Style shell) to reserve
space for that paint inside a clipping viewport; neither API is a popup or
overlay replacement. The internal components use the corresponding shared
implementations below the public facade.

Transition state needs stable widget identity. Stateful components claim their
component key; each non-interactive transitioning sibling uses a distinct
`layout.Key` scope. A `Box` may use its own key directly.

Parent-owned layout policy stays outside Style: flex growth, sibling gap,
alignment, absolute/portal placement, z-order, and scroll state belong to their
layout containers. This prevents CSS-like syntax from hiding MVU state or
changing ownership boundaries.

`layout.Box` follows the same split. Box-model geometry (width, height, min/max,
fill, padding, margin, overflow) and paint live only in `Style` (including
`resolve.Scope`). The widget itself keeps identity (`Key`), interaction
(`OnClick` / `Disabled` / `Label`), and content alignment (`Align`).

## State Ownership

Application and business state belongs in the user's `Model`. A controlled
component receives values from the model and reports intent through callbacks;
the callback sends a typed message, and `Update` mutates the model.

`ui.Context` owns only per-window Gio services and transient interaction state,
such as clickables, draggables, editors, focus, animation progress, and overlay coordination.
It must not become a second application model.

The public context exposes read-only theme and language snapshots plus the small
set of Gio state and focus helpers needed by custom widgets. Runtime theme and
language changes go through `Application.SetTheme` and `Application.SetLanguage`
so they are applied on the target window's event loop. Frame lifecycle and
component registration remain internal; root overlays are exposed only through
the constrained Portal API.

Custom widgets implement `ui.Widget` directly or use `ui.WidgetFunc`. Transient
Gio interaction state can use `ui.UseState` or `ui.UseStateWith`; it is retained
while its explicit key is rendered and released automatically after removal.
Business values still belong in the Model. Public focus methods on `ui.Context`
keep custom focus rings consistent with pointer and keyboard modality.

`portal.New` is the low-level escape hatch for custom root-level content. It
provides a resolved viewport anchor, stacking group, and preceding-frame input
ownership. It intentionally does not add positioning, dismissal, animation,
backdrops, or focus trapping. Applications should prefer Popover, Modal,
Tooltip, and Menu when those policies fit. Portal content may register nested
FlowUI overlays normally.

Component identity is explicit. A key is scoped by `layout.Key`, claimed with a
component kind, and paired with a typed state slot. Identity must not depend on
render order or an implicit widget-tree path.

FlowUI layout containers pre-register keyed Label and Description associations,
so their semantic relationship is independent of visual layout order. A custom
composite Widget that directly lays out children after the control uses an
invalidated follow-up frame; laying associations before the control gives
same-frame semantics.

## The View Contract

FlowUI takes Elm's function set (`Init` / `Update` / `Subscriptions` / `View`) but
not Elm's purity. The three dimensions are chosen separately and on purpose:

| Dimension | Choice | Why |
| --- | --- | --- |
| `Update`'s model access | `func(*M, Msg) Cmd[Msg]` — mutate in place | Go has no cheap persistent data structures. Copying a large model per message is the price Elm pays for time travel, and FlowUI does not offer time travel |
| `View`'s model access | `func(*Context, M, Send[Msg]) Widget` — by value, read-only | the render pass must not be able to change what it renders |
| Dispatch | `Send[Msg]` handed to `View` (TCA / Redux style) | Gio layout is imperative. A declarative `Html Msg` tree would have to be built and then diffed, while FlowUI renders Gio ops directly |

So the boundary is one-way: **`Update` is the only mutation point; everything else reads.**

What the type system enforces, and what it does not:

- Passing `M` by value to `View` and `Subscriptions` fences **rebinding**: assigning to
  the parameter cannot reach the model the runtime holds.
- It does **not** fence **aliasing**. A struct copy still shares the backing array of a
  slice, the buckets of a map, and the target of a pointer. Mutating through any of those
  from `View` mutates the live model, and no compiler will report it.

Rules that follow, ordered by how easy they are to get wrong:

1. `View` and `Subscriptions` must not mutate the model or anything it references. Copy
   before editing, report intent with `Send`, and let `Update` perform the write.
2. Neither may retain the model, or a pointer into it, past the call. A command that needs
   model data must capture a value snapshot, per the `Cmd` capture rules in `ui.Cmd`.
3. `View` may call `Send`, but every call queues a message and schedules another frame.
   Dispatching unconditionally during render is a render loop; dispatch from an event
   callback or from `Update`.

`uitest` holds applications to the same boundary: `AppHarness.Frame` stores the model
snapshot the view phase received, so a test observes exactly what `View` saw.

## Overlay Host

Popup and modal components register work with the per-window root Overlay Host.
The frame order is:

```text
BeginFrame
    -> layout the main widget tree and register overlays
    -> layout root overlays in viewport coordinates
    -> apply focus commands
EndFrame
```

The host keeps root popup branches below modal branches, preserves an overlay's
opening order across frames, and processes overlays registered by another
overlay in the same frame. Within a modal branch, a descendant popup is painted
above the modal that registered it but below any pending nested modal. Panel
content is still laid out before `EndFrame`, so keyed state, focus requests,
animations, and MVU callbacks retain normal same-frame semantics.

For application-facing overlay behavior, see [浮层与弹出](guide/09-浮层与弹出.md).

Input ownership for overlay *policy* events (dismiss, Escape, exclusive
open/close) is owned by at most one non-passive overlay:

- When the preceding frame had an owner, `BeginFrame` freezes that identity as
  the only overlay that receives `interactive=true` during the paint pass
  (ownership lags the visual top by one frame).
- When there is no preceding owner (first open, or after the stack clears),
  every overlay paints with `interactive=false`, then only the final visual top
  is re-entered with `interactive=true` on a throwaway ops list. Key claims from
  the paint pass are reused; nested `RegisterOverlay` is ignored during that
  policy pass. This prevents a first-frame modal and its nested popup from both
  consuming Escape.

Geometric pointer routing can still use deliberate pass-through behavior. After
dynamic overlay layout completes, the host records the new visual top and the
nearest modal focus scope. Focus work runs against that completed state, while
policy ownership transfers on the next frame when an owner already exists so
open/close races cannot double-consume one event.

Overlay blockers use pointer-only click regions. They keep panel padding,
arrows, backdrops, and exit animations from passing clicks to the main tree,
but never enter Gio's keyboard focus order. A top modal contributes a tail
focus boundary after all of its dynamic descendants, so focus cannot escape
through a nested Select or Popover. Only the nearest focus-scope ancestor of
the visual top contributes that boundary.

Gio does not expose the transform at a widget's current layout position.
FlowUI's layout containers therefore attach a transform node while measuring a
child and fill in that node after the child's final position is known. An
overlay anchor follows this transform chain to the viewport, including nested
Box, Flex, Grid, Stack, Wrap, List, Scroll, and Tabs layouts. Position flipping
and overflow avoidance then use the real viewport-relative anchor.

The same chain carries local clipping and animated opacity. Scroll/List clips
decide whether an anchor is still visible without replacing the full anchor
used for placement. A child overlay inherits the opacity of every animated
overlay surface that owns it, so nested content exits with its parent instead
of remaining fully opaque. Transform nodes live in a frame-local index arena;
layout containers can track every measured child without one heap allocation
per node. Containers that accept repeated items must use their tracked item
layout path so each item's final offset and clip are recorded.

A custom Widget that places a FlowUI child with raw Gio layout, transform,
clip, or `paint.PushOpacity` operations must wrap that child with
`layout.TrackOverlayPlacement` and record the matching offset, transform, clip, and
opacity on the returned frame-local placement. FlowUI layout containers already
do this automatically. Pointer coordinates and private operation decoding must
not be used as substitutes because they fail for keyboard and programmatic
opening. Custom layouts must call `PlaceOffset` or `PlaceTransform` after the
tracked child returns, including when the final transform is the identity.

## Effects

`ui.Program` groups `Init`, `Update`, `Subscriptions`, and `View` when an
application needs the complete lifecycle. `Init` runs once per window instance
and its optional command is managed by the same effect group as commands
returned by `Update`. `WindowStateMessage`, when present, maps Gio
configuration changes into queued messages for the next update.

`Update` is serialized on the event loop and is the only place a command
workflow may mutate the model. Each returned `Cmd` starts in its own goroutine
after `Update` returns, so it can overlap later updates and views. Every
command receives a root `context.Context` that is canceled when the window is
destroyed. Blocking and fallible work should use `ui.DoContext`; `ui.Do` remains
a convenience for short context-free work. `ui.LatestCmd` cancels an older
same-key command and rejects messages from older generations; use it for
search, preview, and autocomplete work. `ui.CancelLatestCmd` cancels a keyed
workflow when the model no longer needs its result. Keys are retained for a
window lifetime so delayed older closures remain rejectable; use bounded,
stable workflow keys rather than request-specific values.

A command must capture only immutable value snapshots prepared during
`Update`. It must not retain or access the model pointer or a `ui.Context`.
Reference-backed values such as slices and maps need an explicit copy before
capture. A command returns data to the application only by calling its `Send`
argument; concurrent sends are queued safely and applied by `Update` in a
later frame. The runtime bounds the message queue to 256 entries. Ordinary
messages beyond the bound are dropped and reported as `QueueOverflowError`; if
a subscription stream is full, a queued value from the same generation is
replaced by the latest value. The error queue is bounded to 64 entries as well.
Sends made after cancellation are discarded.

`ui.MapCmd` lets a parent update reuse a child module's command by mapping each
child message to the parent message type. It preserves the command context and
error unchanged; it does not introduce another goroutine or effect lifecycle.

`Subscription` represents long-lived asynchronous input such as a timer,
filesystem watcher, or server event stream. `Run` with
`Program.Subscriptions` derives the
desired subscription set on the initial frame and after an updated model;
stable desired data is reused on animation-only frames. Keys are
lifecycle identities: a stable key retains the running subscription, removal
cancels it, and a changed key starts a replacement. Duplicate or empty keys are
programming errors. A subscription should remain running until its context is
canceled. After normal or failed completion, its key remains retained and is
not restarted every frame; remove and re-add it or change the key to retry.
When keys change, FlowUI cancels removed runners and waits asynchronously for
them to stop before starting replacements. A 250ms grace bound prevents an
uncooperative runner from blocking replacement indefinitely. Messages carry a
subscription generation and are validated on the event thread, so messages
queued by a stopped generation cannot reach `Update`.

Commands and subscriptions may return errors. The runtime recovers their
panics, attaches a stack trace, and reports both failures as `EffectError`
values to the `OnError` handler on the application event thread. Cancellation
errors are ignored. Without an `OnError` option, FlowUI writes the error and any
panic stack to standard error. The handler is for logging and unexpected
infrastructure failures; expected domain failures should normally be converted
into typed messages so `Update` can represent them in the model.
Panics from `Update`, `Subscriptions`, `View`, or a native close-request
handler are recovered as `RuntimePanicError` values with a phase and stack
trace; the window loop stops rather than continuing with a potentially
partially-mutated model. When the
window can request a native close, FlowUI cancels effects, issues that close,
and drains events until `DestroyEvent` before returning the panic error so the
OS shell is not left as a live zombie. Event sources without a close path
(such as headless harnesses) return the panic error immediately.

Within a single frame's message batch, a panic in message *k* leaves earlier
mutations applied, discards later messages in that batch, and does not start
commands for the panicking or later messages. Commands already started from
earlier messages are canceled when the window session ends.

`ui.Batch` starts several independent commands from one `Update` return.
Each command runs concurrently under the same effect context and capture rules
as a single command.

Window shutdown cancels every command and subscription and waits up to two
seconds for their cleanup before returning. A timeout is reported as
`ErrEffectShutdownTimeout`; the process is still allowed to exit after that
bound, so effects must honor their context promptly.

`View` and a `Subscriptions` function follow the same ownership boundary: the
model and any reference-backed fields it contains are read-only. Event
callbacks and effects send messages instead of mutating captured model data.

## Window Close Lifecycle

Gio delivers `app.ClosingEvent` when the user requests to close a window on
Windows, macOS, Wayland, and X11, and the window closes unless the event is
aborted while it is being handled. FlowUI handles the event on the window event
loop before the platform resumes the close, so a handler can still veto it with
`ClosingEvent.Abort()`.

A native close request resolves in this order:

| Configuration | Result |
| --- | --- |
| One-shot intent from an approved `RequestClose`, `Close`, `CloseAll`, or `Quit` | The intent wins and the handler is skipped |
| `window.OnWindowCloseRequest` handler | The handler's `CloseDecision` decides |
| No handler and `Program.CloseRequestMessage` is non-nil | The request is aborted and the returned message is queued for `Update` |
| Nothing configured | The native close proceeds |

The `CloseDecision` values map as follows:

- `CloseProceed` lets the close happen.
- `CloseCancel` calls `Abort()` and keeps the window open.
- `CloseKeepAlive` lets the close happen and enables application keep-alive.
- An unknown decision fails closed and aborts the close.

`Application.RequestClose` asks the same handler before requesting a native
close; when no handler is configured it requests the native close directly, so
`Program.CloseRequestMessage` applies to that path as well. `Close`, `CloseAll`,
and `Quit` are force paths that mark a one-shot intent so the handler is not
called again for the resulting event.

The close handler runs synchronously on the window event-loop goroutine and has
no `ui.Send`, so it must return promptly and cannot open an MVU confirmation
dialog itself. `Program.CloseRequestMessage` is the MVU-facing hook for that
case: FlowUI aborts the native close and queues the message, `Update` records
that confirmation is pending, and the application finishes with
`Application.Close` after the user decides. A nil `CloseRequestMessage` hook
leaves the native close alone. Because the hook aborts the request, the
application owns the final close; `ui.Run` rejects a program that sets this
hook, so a guard uses `ui.NewApplication` and `Application.Run` directly.

Native close events are not delivered on platforms without a window close
request (for example iOS, Android, and the browser), and macOS application
termination can destroy a window without a `ClosingEvent`; those paths cannot
be vetoed through this contract.

## Adding A Component

1. Put the implementation, its focused tests, and its exported constructors in
   `widget/<component>`. The component package **is** the public surface; there
   is no separate facade file to keep in sync.
2. Depend on `internal/core`, `theme`,
   `internal/locale`, and the public `layout`, `draw`, `animation`, or `style`
   packages as needed. Keep widget packages independent of the `ui` package.
3. Use `New` for a component package's main constructor, returning `Widget`:
   `button.New`, `input.New`, and `text.New`. Related controls keep descriptive
   names, such as `button.Group`, `input.TextArea`, and `progress.Bar`.
   Text actions accept strings directly; use `Content(core.Widget)` for custom
   content and keep `Label` for an accessible label override. Containers still
   accept widgets. Avoid repeating the package name in other exported
   identifiers. In `widget/progress`, `ProgressBarWidget`
   becomes `progress.BarWidget` and `ProgressBarAccent` becomes
   `progress.BarAccent`; in `layout`, `LayoutResolved` becomes `layout.Resolved`.
   Component ownership and naming exceptions:
   - **Component vocabulary belongs to its widget package.** Types such as
     `button.Size` and `avatar.Variant` are declared by their owning component;
     `theme` remains the home of shared design tokens and theme construction.
   - **A name is kept when stripping it would collide or stop being a word.**
     `draw.DrawBrush` cannot become `draw.Brush` (the brush type owns that name)
     and `resolve.ResolveStatic` would collapse into the non-word `resolve.Static`.
     Both keep a distinct name; these exceptions are explained in this section.
   File naming follows the same spirit — `<主词>_<面>.go` where the 主词 is the
   package name or a real type name, never a third word. The rule, the audit
   behind it, and the names that are deliberately **not** changed (stdlib/Gio
   collisions such as `layout`, native API mirrors such as `internal/sys`'s
   `GetMessage`) follow the naming rules documented in this section.
   **The type name follows the same count rule**: a package with a single main
   control names it `Widget` (`button.Widget`, `selector.SelectWidget`); a package
   with several parallel controls names each `XxxWidget`
   (`layout.RowWidget`, `toggle.SwitchWidget`).
4. Keep model values controlled and report changes through callbacks.
5. Keep the default appearance and its state, theme, and instance-style tests in
   the component package. Shared tokens belong in `theme`; shared implementation
   details belong in `internal`.
6. Add runnable usage examples under `examples/` in the root module.
7. Run the affected public packages and `internal/...` with `go test` and
   `go vet`.

Component packages follow ownership boundaries, not a fixed "one directory per
widget" rule. Closely coupled controls may share a domain package when they own
the same interaction primitives. For example, input-like controls share field
state and Select reuses ListBox item behavior. Split a domain only after its
shared primitives can move into a lower package without introducing dependency
cycles. File count alone is not a package boundary.

The root module intentionally contains no Go package. This prevents an
accidental second component facade and keeps `github.com/qianniancn/flowui/ui`
as the MVU/runtime contract while `widget/<component>`, `layout`, `style`, `draw`,
`animation`, and `window` provide focused public APIs.

## Internal Interface Design

Internal packages serve as building blocks for components and must be robust
against misuse. The following principles guide internal API design to prevent
subtle bugs from incorrect usage patterns.

### Type Safety Over Documentation

Encode invariants in types and signatures rather than relying on documentation
or runtime checks:

- **Bundle synchronized values**: If two fields must stay in sync (like a cache
  and its generation), put them in one struct rather than separate variables
- **Use the type system**: Prevent invalid states through types, not runtime
  validation
- **Make dependencies explicit**: If function B needs the result of function A,
  make A's result a parameter to B

Example: `KeyFilterCache` bundles `target`, `names`, and `filters` together.
The cache key includes both `target` and `names`; omitting either would cause
silent routing errors.

### Eliminate Call-Order Dependencies

Methods should work correctly regardless of invocation order unless the
dependency is explicit in the signature:

- **Idempotent by default**: Calling a method twice with the same arguments
  should produce the same result
- **Order independence**: If two methods don't have a data dependency, they
  should work in any order
- **Explicit sequencing**: When order matters, use:
  - A builder pattern with a terminal operation
  - Method chaining where each step returns the next valid state
  - Parameters that carry results from required prior steps

Example: `ClickArea.Layout` previously depended on being called after
`Clicked`, which drained events. Now both methods process events identically,
so call order doesn't matter.

### Separate Queries from Commands

Methods that return values should not have side effects unless the side effect
is the primary purpose:

- **Query methods are pure**: Methods named like `Get*`, `Is*`, `Visible`,
  `Contains`, or returning a value without changing receiver state should have
  no side effects
- **Command methods change state**: Methods named like `Set*`, `Add*`,
  `Commit*`, or clearly mutating should have side effects as their primary job
- **Split when necessary**: If a method needs both, split it into a query phase
  and a command phase

Example: `Focus.Observe` records frame-local focus input while
`Focus.CommitObservations` applies persistent modality changes at the frame
boundary.

### Complete Cache Keys

When designing caches, ensure the cache key includes every input that affects
the result:

- **All dependencies**: If the result depends on `(A, B, C)`, the cache key
  must include all three
- **Explicit versioning**: Use version tokens (`DataVersion`, theme generation)
  to invalidate when indirect dependencies change
- **Test with variation**: Test caches by varying each input independently, not
  just the common case

Example: `ComboBoxState.visibleItems` caches based on `(query, selectedLabel,
DataVersion)`. Omitting any of these would cause stale results when that input
changes.

`Menu.AutoWidth` follows the same rule. Its versioned measurement cache includes
the menu data version, theme generation, DPI metrics, layout constraints,
custom style, density and width options. Menus without `DataVersion` bypass the
cache, and changing section separation or empty-state text invalidates the
relevant result.

### Fail Loudly, Not Silently

When invariants are violated, fail immediately with a clear error rather than
continuing with corrupt state:

- **Panic on impossible states**: Use `panic` for internal invariant violations
  that indicate programming errors
- **Return errors for expected failures**: Use `error` returns for failures
  that callers should handle (I/O, validation, user input)
- **Validate early**: Check preconditions at the start of a method, not after
  partial state changes

Example: `state.Keys.claim` panics on duplicate keys with a clear message
rather than silently overwriting the previous registration.

### Design Review Questions

When reviewing or designing internal APIs, ask:

- Can this be misused by changing call order?
- Does this method name promise a query but perform writes?
- Are all cache dependencies reflected in the cache key?
- Will incorrect usage fail loudly or silently corrupt state?
- Could two callers interfere with each other's state?
- Are there implicit assumptions that should be explicit parameters?

These principles emerged from real issues encountered during development:
- `KeyFilterCache` ignored `target`, causing potential routing errors (#12)
- `ClickArea.Layout` had hidden call-order requirements (#13)
- `Focus.Visible` mixed query and command concerns (#14)

Internal APIs built with these principles are easier to use correctly, harder
to misuse, and fail early when misused.

### Layout and visual container ownership

`layout` owns arrangement, the shared Box style adapter and tracked clipping
viewports. `VisualOutsetState` retains overflow only; scrollbar declaration
caches belong to scrollbar state. `Scroll` is the single public scrolling
viewport (`Overlay` or `WithoutScrollbar` changes bar presentation without
changing identity); `List` remains the virtualized item container.

`widget/splitpane` owns the default divider appearance and composes
`base.SplitPaneInteraction` for input. `widget/divider` owns visual separators.
These widgets depend on `layout`, never the reverse. Field association traversal
uses direct calls and the existing `PrepareFieldAssociation` protocol instead
of global init-time registration. `Items` and `Wrap` share one wrapping engine
while retaining their pixel-gap and Dp-gap contracts respectively.
