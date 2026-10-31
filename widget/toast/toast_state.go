package toast

import (
	"fmt"
	"image"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotToast = "toast"

func toastStateFor(ctx *core.Context, key string) (string, *toastHostState) {
	key = core.ClaimKey(ctx, core.KindToast, key)
	return key, core.UseState[toastHostState](ctx, key, stateSlotToast)
}

type toastHostState struct {
	entries        map[string]*toastEntryState
	order          []string
	region         toastRegionTag
	regionHover    base.HoverInteraction
	touchMode      bool
	expansion      animation.FloatTransition
	expansionValue float32
}

type toastEntryState struct {
	item               Item
	present            bool
	closeRequested     bool
	root               toastRootTag
	hover              base.HoverInteraction
	close              widget.Clickable
	action             widget.Clickable
	remaining          time.Duration
	configuredTimeout  time.Duration
	deadline           time.Time
	timerRunning       bool
	value              float32
	progressTransition animation.FloatTransition
	stack              animation.FloatTransition
	expandedOffset     animation.FloatTransition
}

type toastRootTag struct{ _ byte }
type toastRegionTag struct{ _ byte }

func (s *toastHostState) sync(gtx layout.Context, items []Item, defaultTimeout time.Duration) {
	if s.entries == nil {
		s.entries = make(map[string]*toastEntryState)
	}
	seen := make(map[string]struct{}, len(items))
	nextOrder := make([]string, 0, len(items)+len(s.order))
	for _, item := range items {
		if item.key == "" {
			panic("flowui: empty toast key")
		}
		if _, duplicate := seen[item.key]; duplicate {
			panic(fmt.Sprintf("flowui: duplicate toast key %q", item.key))
		}
		seen[item.key] = struct{}{}
		nextOrder = append(nextOrder, item.key)
		entry := s.entries[item.key]
		timeout := defaultTimeout
		if item.hasTimeout {
			timeout = item.timeout
		}
		timeout = max(timeout, 0)
		if entry == nil {
			entry = &toastEntryState{
				item:              item,
				present:           true,
				remaining:         timeout,
				configuredTimeout: timeout,
			}
			if timeout > 0 {
				entry.deadline = gtx.Now.Add(timeout)
				entry.timerRunning = true
			}
			s.entries[item.key] = entry
			continue
		}
		restartTimer := !entry.present || entry.configuredTimeout != timeout || entry.item.loading != item.loading
		entry.item = item
		entry.present = true
		if restartTimer && !entry.closeRequested {
			entry.configuredTimeout = timeout
			entry.remaining = timeout
			entry.timerRunning = timeout > 0
			if entry.timerRunning {
				entry.deadline = gtx.Now.Add(timeout)
			}
		}
	}

	for _, key := range s.order {
		if _, current := seen[key]; current {
			continue
		}
		entry := s.entries[key]
		if entry == nil {
			continue
		}
		entry.present = false
		entry.timerRunning = false
		nextOrder = append(nextOrder, key)
	}
	s.order = nextOrder
}

func (s *toastHostState) visible() bool {
	for _, entry := range s.entries {
		if entry.value > 0 || (entry.present && !entry.closeRequested) {
			return true
		}
	}
	return false
}

func (s *toastHostState) resetRegion() {
	s.regionHover.Hovered = false
	s.expansion.Reset()
	s.expansionValue = 0
}

func (s *toastHostState) cleanup() {
	if len(s.entries) == 0 {
		return
	}
	nextOrder := s.order[:0]
	for _, key := range s.order {
		entry := s.entries[key]
		if entry == nil {
			continue
		}
		if !entry.present && entry.value <= 0 {
			delete(s.entries, key)
			continue
		}
		nextOrder = append(nextOrder, key)
	}
	s.order = nextOrder
}

func (s *toastHostState) entry(key string) *toastEntryState {
	return s.entries[key]
}

func (s *toastHostState) paused(gtx layout.Context) bool {
	for _, key := range s.order {
		entry := s.entries[key]
		if entry != nil && (entry.hover.Hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close) || gtx.Focused(&entry.action)) {
			return true
		}
	}
	return false
}

func (s *toastHostState) updateRegionEvents(gtx layout.Context) {
	s.regionHover.Update(gtx, &s.region, true)
	for {
		e, ok := core.NextPointerEvent(gtx, &s.region, pointer.Press)
		if !ok {
			return
		}
		if gtx.Enabled() && e.Source == pointer.Touch {
			s.touchMode = true
		}
	}
}

func (s *toastHostState) addRegionInput(gtx layout.Context, bounds image.Rectangle) {
	s.regionHover.AddInput(gtx, &s.region, bounds, true, nil)
}

func (s *toastHostState) expansionProgress(gtx layout.Context, expanded bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if expanded {
		target = 1
	}
	s.expansionValue = s.expansion.Value(gtx, target, max(duration, time.Millisecond), animation.EaseSmoothstep, motions...)
	return s.expansionValue
}

func (e *toastEntryState) updateRootEvents(gtx layout.Context) {
	e.hover.Update(gtx, &e.root, true)
}

func (e *toastEntryState) addRootInput(gtx layout.Context, size image.Point) {
	e.hover.AddInput(gtx, &e.root, image.Rectangle{Max: size}, true, nil)
}

func (s *toastHostState) updateTimers(gtx layout.Context, paused bool, close func(*toastEntryState)) {
	for _, key := range s.order {
		entry := s.entries[key]
		if entry == nil || !entry.present || entry.closeRequested || entry.remaining <= 0 {
			continue
		}
		if paused {
			entry.pauseTimer(gtx.Now)
			continue
		}
		entry.resumeTimer(gtx.Now)
		if !gtx.Now.Before(entry.deadline) {
			close(entry)
			continue
		}
		gtx.Execute(op.InvalidateCmd{At: entry.deadline})
	}
}

func (e *toastEntryState) pauseTimer(now time.Time) {
	if !e.timerRunning {
		return
	}
	e.remaining = max(e.deadline.Sub(now), 0)
	e.timerRunning = false
}

func (e *toastEntryState) resumeTimer(now time.Time) {
	if e.timerRunning || e.remaining <= 0 {
		return
	}
	e.deadline = now.Add(e.remaining)
	e.timerRunning = true
}

func (e *toastEntryState) requestClose(onClose func(string)) {
	if e.closeRequested {
		return
	}
	e.closeRequested = true
	e.timerRunning = false
	if onClose != nil {
		onClose(e.item.key)
	}
}

func (e *toastEntryState) progress(gtx layout.Context, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if e.present && !e.closeRequested {
		target = 1
	}
	if !e.progressTransition.Ready() {
		e.progressTransition.Initialize(0, gtx.Now)
	}
	e.value = e.progressTransition.Value(gtx, target, max(duration, time.Millisecond), animation.EaseSmoothstep, motions...)
	return e.value
}

func (e *toastEntryState) stackPosition(gtx layout.Context, target float32, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return e.stack.Value(gtx, target, max(duration, time.Millisecond), animation.EaseSmoothstep, motions...)
}

func (e *toastEntryState) expandedPosition(gtx layout.Context, target float32, duration time.Duration, motions ...theme.MotionTheme) float32 {
	return e.expandedOffset.Value(gtx, target, max(duration, time.Millisecond), animation.EaseSmoothstep, motions...)
}
