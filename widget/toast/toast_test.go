package toast

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

func TestToastIndicatorNilHidesIndicator(t *testing.T) {
	if !NewItem("default", "Default").showIndicator() {
		t.Fatal("default indicator is hidden")
	}
	if NewItem("hidden", "Hidden").Indicator(nil).showIndicator() {
		t.Fatal("nil custom indicator did not hide the indicator")
	}
}

func TestMessageItemIsCompact(t *testing.T) {
	item := MessageItem("saved", "Saved")
	if !item.compact {
		t.Fatal("MessageItem is not compact")
	}
	if item.timeout != 3*time.Second || !item.hasTimeout {
		t.Fatalf("MessageItem timeout = %v, want 3s", item.timeout)
	}
	if item.description != "" || item.actionLabel != "" {
		t.Fatalf("MessageItem has notification content: %#v", item)
	}
}

func TestNewItemUsesGhostAction(t *testing.T) {
	if got := NewItem("saved", "Saved").actionVariant; got != button.Ghost {
		t.Fatalf("default action variant = %v, want ghost", got)
	}
}

func TestToastStateSyncUsesDefaultAndExplicitTimeouts(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastHostState
	state.sync(toastTestContextAt(start), []Item{
		NewItem("default", "Default"),
		NewItem("persistent", "Persistent").Timeout(0),
	}, 4*time.Second)

	if got := state.entry("default"); got.remaining != 4*time.Second || got.deadline != start.Add(4*time.Second) || !got.timerRunning {
		t.Fatalf("default timer = %#v", got)
	}
	if got := state.entry("persistent"); got.remaining != 0 || got.timerRunning {
		t.Fatalf("persistent timer = %#v", got)
	}
}

func TestToastStateSameKeyLoadingUpdateRestartsTimerOnce(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastHostState
	state.sync(toastTestContextAt(start), []Item{
		NewItem("upload", "Uploading").Loading(true).Timeout(0),
	}, 4*time.Second)
	entry := state.entry("upload")
	if entry.remaining != 0 || entry.timerRunning {
		t.Fatalf("loading timer = remaining %v running %v", entry.remaining, entry.timerRunning)
	}

	updatedAt := start.Add(2 * time.Second)
	done := NewItem("upload", "Uploaded").Variant(Success)
	state.sync(toastTestContextAt(updatedAt), []Item{done}, 4*time.Second)
	wantDeadline := updatedAt.Add(4 * time.Second)
	if entry.item.title != "Uploaded" || entry.configuredTimeout != 4*time.Second || entry.remaining != 4*time.Second || entry.deadline != wantDeadline || !entry.timerRunning {
		t.Fatalf("updated timer = %#v", entry)
	}

	state.sync(toastTestContextAt(updatedAt.Add(time.Second)), []Item{done}, 4*time.Second)
	if entry.deadline != wantDeadline {
		t.Fatalf("unchanged toast reset deadline = %v, want %v", entry.deadline, wantDeadline)
	}
}

func TestToastStateRejectsDuplicateAndEmptyKeys(t *testing.T) {
	for _, items := range [][]Item{
		{NewItem("", "Empty")},
		{NewItem("same", "One"), NewItem("same", "Two")},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid toast keys did not panic")
				}
			}()
			var state toastHostState
			state.sync(toastTestContextAt(time.Unix(1, 0)), items, 4*time.Second)
		}()
	}
}

func TestHostCleansEntryRemovedBeforeAnimation(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, New([]Item{
		NewItem("saved", "Saved").Timeout(0),
	}), start, viewport)
	layoutToastFrame(ctx, router, New(nil).Key("toasts"), start, viewport)

	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	if len(stateValue.entries) != 0 || len(stateValue.order) != 0 {
		t.Fatalf("removed invisible toast was retained: entries=%d order=%v", len(stateValue.entries), stateValue.order)
	}
}

func TestToastTimerPausesAndResumes(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastHostState
	state.sync(toastTestContextAt(start), []Item{NewItem("saved", "Saved")}, 4*time.Second)
	closed := false

	state.updateTimers(toastTestContextAt(start.Add(time.Second)), true, func(*toastEntryState) { closed = true })
	entry := state.entry("saved")
	if entry.remaining != 3*time.Second || entry.timerRunning || closed {
		t.Fatalf("paused timer = remaining %v running %v closed %v", entry.remaining, entry.timerRunning, closed)
	}
	state.updateTimers(toastTestContextAt(start.Add(2*time.Second)), false, func(*toastEntryState) { closed = true })
	if entry.deadline != start.Add(5*time.Second) || !entry.timerRunning || closed {
		t.Fatalf("resumed timer = deadline %v running %v closed %v", entry.deadline, entry.timerRunning, closed)
	}
	state.updateTimers(toastTestContextAt(start.Add(5*time.Second)), false, func(*toastEntryState) { closed = true })
	if !closed {
		t.Fatal("timer did not close at resumed deadline")
	}
}

func TestToastTimerPausesWhileActionIsFocused(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	host := New([]Item{
		NewItem("saved", "Saved").Action("Undo"),
	}).Key("toasts")
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), viewport)

	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil {
		t.Fatal("toast state is missing")
	}
	router.Source().Execute(key.FocusCmd{Tag: &entry.action})
	layoutToastFrame(ctx, router, host, start.Add(time.Second), viewport)
	if entry.remaining != 3*time.Second || entry.timerRunning {
		t.Fatalf("focused action timer = remaining %v running %v", entry.remaining, entry.timerRunning)
	}
}

func TestToastEntryProgressEntersAndExits(t *testing.T) {
	start := time.Unix(1, 0)
	entry := toastEntryState{present: true}
	gtx := toastTestContextAt(start)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	gtx.Now = start.Add(350 * time.Millisecond)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 1 {
		t.Fatalf("entered progress = %v, want 1", got)
	}
	entry.present = false
	entry.progress(gtx, 350*time.Millisecond)
	gtx.Now = start.Add(700 * time.Millisecond)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 0 {
		t.Fatalf("exited progress = %v, want 0", got)
	}
}

func TestToastStackPositionAnimatesBetweenIndexes(t *testing.T) {
	start := time.Unix(1, 0)
	entry := toastEntryState{}
	gtx := toastTestContextAt(start)
	if got := entry.stackPosition(gtx, 1, 350*time.Millisecond); got != 1 {
		t.Fatalf("initial stack position = %v, want 1", got)
	}
	entry.stackPosition(gtx, 0, 350*time.Millisecond)
	gtx.Now = start.Add(175 * time.Millisecond)
	if got := entry.stackPosition(gtx, 0, 350*time.Millisecond); got <= 0 || got >= 1 {
		t.Fatalf("mid-animation stack position = %v, want between 0 and 1", got)
	}
	gtx.Now = start.Add(350 * time.Millisecond)
	if got := entry.stackPosition(gtx, 0, 350*time.Millisecond); got != 0 {
		t.Fatalf("final stack position = %v, want 0", got)
	}
}

func TestToastExpandedMiddleCloseKeepsTrailingPositionOnFirstFrame(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	items := []Item{
		NewItem("front", "Front").Timeout(0),
		NewItem("middle", "Middle").Timeout(0),
		NewItem("back", "Back").Timeout(0),
	}
	host := New(items).Key("toasts").Placement(BottomEnd)
	viewport := image.Pt(900, 600)
	start := time.Unix(1, 0)
	duration := defaultToastMetrics.AnimationDuration
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(duration), viewport)
	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	stateValue.regionHover.Hovered = true
	layoutToastFrame(ctx, router, host, start.Add(duration+time.Millisecond), viewport)
	layoutToastFrame(ctx, router, host, start.Add(2*duration+time.Millisecond), viewport)

	middle, middleOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Middle")
	back, backOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Back")
	if !middleOK || !backOK {
		t.Fatal("expanded toast semantics are missing")
	}
	host = New([]Item{items[0], items[2]}).Key("toasts").Placement(BottomEnd)
	closeAt := start.Add(2*duration + 2*time.Millisecond)
	layoutToastFrame(ctx, router, host, closeAt, viewport)
	afterClose, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Back")
	if !ok || afterClose.Desc.Bounds.Min != back.Desc.Bounds.Min {
		t.Fatalf("trailing toast jumped on middle close: before %v after %v", back.Desc.Bounds, afterClose.Desc.Bounds)
	}

	layoutToastFrame(ctx, router, host, closeAt.Add(duration), viewport)
	settled, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Back")
	if !ok || settled.Desc.Bounds.Min != middle.Desc.Bounds.Min {
		t.Fatalf("trailing toast settled at %v, want former middle position %v", settled.Desc.Bounds, middle.Desc.Bounds)
	}
}

func TestToastRegionXMatchesPlacements(t *testing.T) {
	tests := []struct {
		placement Placement
		want      int
	}{
		{BottomStart, 16},
		{TopStart, 16},
		{Bottom, 70},
		{Top, 70},
		{BottomEnd, 124},
		{TopEnd, 124},
	}
	for _, test := range tests {
		if got := toastRegionX(600, 460, 16, test.placement); got != test.want {
			t.Fatalf("placement %d x = %d, want %d", test.placement, got, test.want)
		}
	}
}

func TestToastStylesUseVariantColors(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	host := New(nil).Key("toasts")
	semantic := activeTheme.Semantic()
	tests := []struct {
		variant Variant
		title   color.NRGBA
	}{
		{Default, semantic.OverlayForeground},
		{Accent, semantic.Tone(theme.TonePrimary).OnSoft},
		{Success, semantic.Tone(theme.ToneSuccess).OnSoft},
		{Warning, semantic.Tone(theme.ToneWarning).OnSoft},
		{Danger, semantic.Tone(theme.ToneDanger).OnSoft},
	}
	for _, test := range tests {
		visual := host.visualWithCache(ctx, test.variant, button.Ghost)
		resolved := host.resolveStyle(ctx, toastTestContextAt(time.Unix(1, 0)), style.State{}, visual)
		title, ok := computed.Color(resolved.title.Text.Color)
		if !ok || title != test.title {
			t.Fatalf("variant %d title color = %v, want %v", test.variant, title, test.title)
		}
		description, ok := computed.Color(resolved.description.Text.Color)
		if !ok || description != semantic.MutedForeground {
			t.Fatalf("variant %d description color = %v, want %v", test.variant, description, semantic.MutedForeground)
		}
		indicator, ok := computed.Color(resolved.indicator.Text.Color)
		if !ok || indicator != test.title && test.variant != Accent {
			t.Fatalf("variant %d indicator color = %v, want %v", test.variant, indicator, test.title)
		}
		if test.variant == Accent && indicator != semantic.OverlayForeground {
			t.Fatalf("accent indicator color = %v, want overlay foreground", indicator)
		}
	}
}

func TestToastStyleIsLastCascadeLayer(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	background := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	host := New(nil).Key("toasts").Style(style.Style{}.Background(style.Solid(background)))
	visual := host.baseVisual(ctx)
	resolved := host.resolveStyle(ctx, toastTestContextAt(time.Unix(1, 0)), style.State{}, visual)
	if resolved.panel.Paint == nil {
		t.Fatal("panel has no paint")
	}
	solid, ok := resolved.panel.Paint.Background.(style.SolidColor)
	if !ok || solid.Color != background {
		t.Fatalf("panel background = %#v, want %v", resolved.panel.Paint.Background, background)
	}
}

func TestVisualRecipeUsesComponentDefaultWidth(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	host := New([]Item{NewItem("saved", "Saved").Timeout(0)}).Key("toasts")
	layoutToastFrame(ctx, router, host, start, image.Pt(600, 400))
	layoutToastFrame(ctx, router, host, start.Add(defaultToastMetrics.AnimationDuration), image.Pt(600, 400))
	node, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok || node.Desc.Bounds.Dx() != int(defaultToastMetrics.Width) {
		t.Fatalf("default toast bounds = %v, found=%v; want width %v", node.Desc.Bounds, ok, defaultToastMetrics.Width)
	}
}

func TestVisualRecipeUsesPlacementOffset(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	host := New([]Item{NewItem("saved", "Saved").Timeout(0)}).Key("toasts").
		Placement(BottomEnd).
		Offset(44)
	viewport := image.Pt(600, 400)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(defaultToastMetrics.AnimationDuration), viewport)
	node, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok || node.Desc.Bounds.Max.Y != viewport.Y-44 {
		t.Fatalf("offset toast bounds = %v, found=%v", node.Desc.Bounds, ok)
	}
}

func TestToastStackExpandsOnHover(t *testing.T) {
	for _, test := range []struct {
		name      string
		placement Placement
		expandsUp bool
	}{
		{name: "bottom", placement: BottomEnd, expandsUp: true},
		{name: "top", placement: TopEnd},
	} {
		t.Run(test.name, func(t *testing.T) {
			activeTheme := toastBehaviorTheme()
			ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
			router := new(input.Router)
			frontIndicator := new(toastProbe)
			backIndicator := new(toastProbe)
			host := New([]Item{
				NewItem("front", "Front").Indicator(frontIndicator).Timeout(0),
				NewItem("back", "Back").Description("A taller toast behind the front notification.").Indicator(backIndicator).Timeout(0),
			}).Key("toasts").Placement(test.placement)
			viewport := image.Pt(900, 600)
			start := time.Unix(1, 0)
			duration := defaultToastMetrics.AnimationDuration
			layoutToastFrame(ctx, router, host, start, viewport)
			layoutToastFrame(ctx, router, host, start.Add(duration), viewport)

			front, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Front")
			if !ok {
				t.Fatal("front toast semantic node is missing")
			}
			center := f32.Pt(
				float32(front.Desc.Bounds.Min.X+front.Desc.Bounds.Max.X)/2,
				float32(front.Desc.Bounds.Min.Y+front.Desc.Bounds.Max.Y)/2,
			)
			router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: center})
			layoutToastFrame(ctx, router, host, start.Add(duration+time.Millisecond), viewport)
			layoutToastFrame(ctx, router, host, start.Add(2*duration+time.Millisecond), viewport)

			stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
			if !stateValue.regionHover.Hovered || stateValue.expansionValue != 1 {
				t.Fatalf("toast region expansion = hovered %v progress %v", stateValue.regionHover.Hovered, stateValue.expansionValue)
			}
			if !frontIndicator.enabled || !backIndicator.enabled {
				t.Fatalf("expanded indicator enabled states = front %v back %v", frontIndicator.enabled, backIndicator.enabled)
			}
			front, frontOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Front")
			back, backOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Back")
			separated := front.Desc.Bounds.Max.Y <= back.Desc.Bounds.Min.Y
			if test.expandsUp {
				separated = back.Desc.Bounds.Max.Y <= front.Desc.Bounds.Min.Y
			}
			if !frontOK || !backOK || !separated {
				t.Fatalf("expanded toast bounds = front %v back %v", front.Desc.Bounds, back.Desc.Bounds)
			}

			router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(1, float32(viewport.Y)/2)})
			layoutToastFrame(ctx, router, host, start.Add(2*duration+2*time.Millisecond), viewport)
			layoutToastFrame(ctx, router, host, start.Add(3*duration+2*time.Millisecond), viewport)
			if stateValue.regionHover.Hovered || stateValue.expansionValue != 0 || backIndicator.enabled {
				t.Fatalf("toast region remained expanded = hovered %v progress %v back enabled %v", stateValue.regionHover.Hovered, stateValue.expansionValue, backIndicator.enabled)
			}
		})
	}
}

func TestToastTouchModeShowsCloseButtonOnWideViewport(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	host := New([]Item{NewItem("saved", "Saved").Timeout(0)}).Key("toasts")
	viewport := image.Pt(900, 500)
	start := time.Unix(1, 0)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), viewport)
	if _, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close"); ok {
		t.Fatal("desktop close button was visible before interaction")
	}
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := f32.Pt(
		float32(root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2,
		float32(root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 1, Position: center})
	layoutToastFrame(ctx, router, host, start.Add(360*time.Millisecond), viewport)

	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	if !stateValue.touchMode {
		t.Fatal("touch input did not enable toast touch mode")
	}
	if _, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close"); !ok {
		t.Fatal("touch mode close button is missing")
	}
}

func TestToastActionReportsItemKey(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	actionKey := ""
	host := New([]Item{
		NewItem("saved", "Saved").Action("Undo").Timeout(0),
	}).OnAction(func(key string) { actionKey = key })
	layoutToastFrame(ctx, router, host, start, image.Pt(900, 500))
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), image.Pt(900, 500))
	node, ok := semanticNodeWithClass(router.AppendSemantics(nil), semantic.Button)
	if !ok {
		t.Fatal("toast action semantic node is missing")
	}
	center := image.Pt(
		(node.Desc.Bounds.Min.X+node.Desc.Bounds.Max.X)/2,
		(node.Desc.Bounds.Min.Y+node.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(float32(center.X), float32(center.Y))},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))},
	)
	layoutToastFrame(ctx, router, host, start.Add(360*time.Millisecond), image.Pt(900, 500))
	if actionKey != "saved" {
		t.Fatalf("action key = %q, want saved", actionKey)
	}
}

func TestToastCloseButtonAppearsOnHoverAndReportsKey(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	closedKey := ""
	host := New([]Item{
		NewItem("saved", "Saved").Timeout(0),
	}).OnClose(func(key string) { closedKey = key })
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), viewport)
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := image.Pt((root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2, (root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))})
	layoutToastFrame(ctx, router, host, start.Add(360*time.Millisecond), viewport)
	closeNode, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close")
	if !ok {
		t.Fatal("hovered toast close button is missing")
	}
	closeCenter := image.Pt(
		(closeNode.Desc.Bounds.Min.X+closeNode.Desc.Bounds.Max.X)/2,
		(closeNode.Desc.Bounds.Min.Y+closeNode.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(float32(closeCenter.X), float32(closeCenter.Y))},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(closeCenter.X), float32(closeCenter.Y))},
	)
	layoutToastFrame(ctx, router, host, start.Add(370*time.Millisecond), viewport)
	if closedKey != "saved" {
		t.Fatalf("closed key = %q, want saved", closedKey)
	}
}

func TestToastTimerStaysPausedOverCloseButton(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	host := New([]Item{NewItem("saved", "Saved")}).Key("toasts")
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), viewport)
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := image.Pt((root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2, (root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))})
	layoutToastFrame(ctx, router, host, start.Add(360*time.Millisecond), viewport)
	closeNode, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close")
	if !ok {
		t.Fatal("hovered toast close button is missing")
	}
	closeCenter := image.Pt(
		(closeNode.Desc.Bounds.Min.X+closeNode.Desc.Bounds.Max.X)/2,
		(closeNode.Desc.Bounds.Min.Y+closeNode.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(closeCenter.X), float32(closeCenter.Y))})
	layoutToastFrame(ctx, router, host, start.Add(370*time.Millisecond), viewport)

	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil || !entry.close.Hovered() || entry.timerRunning {
		t.Fatalf("close hover pause = entry %#v", entry)
	}
}

func TestToastRootAcceptsKeyboardFocus(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	host := New([]Item{
		NewItem("saved", "Saved").Timeout(0),
	}).Key("toasts")
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, host, start, viewport)
	layoutToastFrame(ctx, router, host, start.Add(350*time.Millisecond), viewport)

	stateValue := core.UseState[toastHostState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil {
		t.Fatal("toast state is missing")
	}
	router.Source().Execute(key.FocusCmd{Tag: &entry.root})
	if !router.Source().Focused(&entry.root) {
		t.Fatal("toast root did not accept keyboard focus")
	}
	closed := ""
	host = host.OnClose(func(key string) { closed = key })
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutToastFrame(ctx, router, host, start.Add(351*time.Millisecond), viewport)
	if closed != "saved" {
		t.Fatalf("focused Escape closed %q, want saved", closed)
	}
}

func TestToastCustomIndicatorUsesComponentDefaultSize(t *testing.T) {
	activeTheme := toastBehaviorTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	probe := new(toastProbe)
	host := New(nil).Key("toasts")
	gtx := toastTestContextAt(time.Unix(1, 0))
	visual := host.visualWithCache(ctx, Default, button.Ghost)
	resolved := host.resolveStyle(ctx, gtx, style.State{}, visual)
	host.layoutToastIndicator(ctx, gtx, NewItem("custom", "Custom").Indicator(probe), visual, resolved.indicator)
	want := int(defaultToastMetrics.IndicatorSize)
	if probe.constraints != layout.Exact(image.Pt(want, want)) {
		t.Fatalf("indicator constraints = %#v, want exact %dx%d", probe.constraints, want, want)
	}
}

func TestToastBackIndicatorReceivesDisabledContext(t *testing.T) {
	ctx := core.New(nil, new(theme.Light()), locale.LanguageEnglish)
	router := new(input.Router)
	front := new(toastProbe)
	back := new(toastProbe)
	host := New([]Item{
		NewItem("front", "Front").Indicator(front).Timeout(0),
		NewItem("back", "Back").Indicator(back).Timeout(0),
	})
	layoutToastFrame(ctx, router, host, time.Unix(1, 0), image.Pt(900, 500))
	if !front.enabled || back.enabled {
		t.Fatalf("indicator enabled states = front %v back %v", front.enabled, back.enabled)
	}
}

type toastProbe struct {
	constraints layout.Constraints
	enabled     bool
}

func (p *toastProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	p.enabled = gtx.Enabled()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func toastTestContextAt(now time.Time) layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(600, 400)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
}

func layoutToastFrame(ctx *core.Context, router *input.Router, host Widget, now time.Time, viewport image.Point) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	host.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func semanticNodeWithLabel(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Label == label {
			return node, true
		}
		if child, ok := semanticNodeWithLabel(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticNodeWithClass(nodes []input.SemanticNode, class semantic.ClassOp) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class {
			return node, true
		}
		if child, ok := semanticNodeWithClass(node.Children, class); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

// paletteTheme installs a theme's v1 palette as its colour source, so edits to
// the palette reach the widgets: a widget reads the semantic layer, and a theme
// built from a Seed carries one that no longer follows the palette.
func toastBehaviorTheme() theme.Theme {
	active := theme.Light()

	return active
}

func semanticButtonWithExactLabel(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button && node.Desc.Label == label {
			return node, true
		}
		if child, ok := semanticButtonWithExactLabel(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}
