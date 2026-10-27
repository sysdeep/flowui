// Package panel provides lifecycle-aware hosts for mutually exclusive views.
//
// A PanelWidget lays out one visible panel at a time. Hidden panels are deliberately
// kept out of the paint, input, focus, and semantic streams. Callers can opt
// into retaining their transient widget state or initializing them lazily in
// an isolated hidden pass. The host does not render a tab strip; Tabs, a dock
// header, or an application-specific navigation control can own selection.
package surface

import (
	"fmt"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const panelStateSlotPanelHost = "panel-host"

// PanelItem describes one view hosted by a PanelWidget.
type PanelItem struct {
	Key      string
	Content  core.Widget
	Disabled bool
}

// PanelWidget presents one selected panel and owns its hidden-panel lifecycle.
type PanelWidget struct {
	key                string
	selectedKey        string
	hasSelectedKey     bool
	defaultSelectedKey string
	hasDefaultSelected bool
	items              []PanelItem
	onChange           func(string)
	keepAlive          bool
	destroyOnHidden    bool
	forceRender        bool
	disabled           bool
}

type panelHostState struct {
	disclosure      disclosure.Binding[string]
	retainedPanels  map[string]struct{}
	renderedPanels  map[string]struct{}
	itemKeys        map[string]struct{}
	forceRender     bool
	destroyOnHidden bool
}

// New creates a lifecycle-aware panel host. A non-empty selectedKey uses
// controlled mode; an empty selectedKey uses uncontrolled mode and may be
// seeded with DefaultSelectedKey.
func NewPanel(key, selectedKey string, items []PanelItem) PanelWidget {
	return PanelWidget{
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: selectedKey != "",
		items:          append([]PanelItem(nil), items...),
	}
}

func (h PanelWidget) SelectedKey(key string) PanelWidget {
	h.selectedKey = key
	h.hasSelectedKey = true
	return h
}

func (h PanelWidget) DefaultSelectedKey(key string) PanelWidget {
	h.defaultSelectedKey = key
	h.hasDefaultSelected = true
	return h
}

func (h PanelWidget) OnChange(fn func(string)) PanelWidget {
	h.onChange = fn
	return h
}

// BindChange composes an observer with the existing selection callback. It
// lets a Workbench controller link PanelHost selection without replacing the
// application's own model callback.
func (h PanelWidget) BindChange(fn func(string)) PanelWidget {
	if fn == nil {
		return h
	}
	previous := h.onChange
	h.onChange = func(key string) {
		if previous != nil {
			previous(key)
		}
		fn(key)
	}
	return h
}

// KeepAlive retains transient state created by hidden panels without laying
// them out again. It is ignored when DestroyOnHidden is enabled.
func (h PanelWidget) KeepAlive(enabled bool) PanelWidget {
	h.keepAlive = enabled
	return h
}

// ForceRender initializes each panel once in a hidden, disabled layout pass.
// The initialized state is retained until the item is removed or
// DestroyOnHidden is enabled.
func (h PanelWidget) ForceRender(enabled bool) PanelWidget {
	h.forceRender = enabled
	return h
}

// Lazy is the inverse spelling of ForceRender. Lazy(true) is the default.
func (h PanelWidget) Lazy(enabled bool) PanelWidget {
	h.forceRender = !enabled
	return h
}

// DestroyOnHidden releases a panel's retained state whenever it stops being
// selected. It takes precedence over KeepAlive and ForceRender.
func (h PanelWidget) DestroyOnHidden(enabled bool) PanelWidget {
	h.destroyOnHidden = enabled
	return h
}

func (h PanelWidget) Disabled(disabled bool) PanelWidget {
	h.disabled = disabled
	return h
}

func (h PanelWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	rootKey := core.ClaimKey(ctx, core.KindPanelHost, h.key)
	value := core.UseStateWith[panelHostState](ctx, rootKey, panelStateSlotPanelHost, func() *panelHostState {
		return &panelHostState{}
	})
	h.begin(value)
	h.checkItems(value)

	cfg := disclosure.Config[string]{
		Controlled: h.hasSelectedKey,
		Value:      h.selectedKey,
		HasDefault: h.hasDefaultSelected,
		Default:    h.defaultSelectedKey,
		OnChange:   h.onChange,
	}
	value.disclosure.Bind(cfg)
	selectedKey := value.disclosure.Current(cfg)
	selectedKey = h.normalizeSelection(value, cfg, selectedKey)
	if h.forceRender && !h.destroyOnHidden && selectedKey != "" {
		value.renderedPanels[selectedKey] = struct{}{}
	}
	h.syncLifecycle(ctx, value)

	if h.forceRender && !h.destroyOnHidden {
		h.layoutForcedPanels(ctx, gtx, value, selectedKey)
	}

	item, ok := h.item(selectedKey)
	if ok && item.Content != nil {
		h.preparePanelAssociations(ctx, item)
	}
	if !ok || item.Content == nil {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(gtx.Constraints.Min)}
	}
	if h.disabled {
		gtx = gtx.Disabled()
	}
	return h.layoutPanel(ctx, gtx, item)
}

func (h PanelWidget) begin(value *panelHostState) {
	if value.retainedPanels == nil {
		value.retainedPanels = make(map[string]struct{}, len(h.items))
	}
	if value.renderedPanels == nil {
		value.renderedPanels = make(map[string]struct{}, len(h.items))
	}
	if value.itemKeys == nil {
		value.itemKeys = make(map[string]struct{}, len(h.items))
	}
	if value.forceRender != h.forceRender || value.destroyOnHidden != h.destroyOnHidden {
		clear(value.renderedPanels)
		value.forceRender = h.forceRender
		value.destroyOnHidden = h.destroyOnHidden
	}
}

func (h PanelWidget) checkItems(value *panelHostState) {
	clear(value.itemKeys)
	for _, item := range h.items {
		if item.Key == "" {
			panic("flowui: empty panel key")
		}
		if _, exists := value.itemKeys[item.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate panel key %q", item.Key))
		}
		value.itemKeys[item.Key] = struct{}{}
	}
}

func (h PanelWidget) syncLifecycle(ctx *core.Context, value *panelHostState) {
	for _, item := range h.items {
		scope := h.panelStateScope(ctx, item.Key)
		keep := h.keepAlive && !h.destroyOnHidden
		if !keep && h.forceRender && !h.destroyOnHidden {
			_, keep = value.renderedPanels[item.Key]
		}
		if keep {
			core.RetainState(ctx, scope)
			value.retainedPanels[item.Key] = struct{}{}
			continue
		}
		core.ReleaseStateRetention(ctx, scope)
		delete(value.retainedPanels, item.Key)
	}
	for itemKey := range value.retainedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		core.ReleaseStateRetention(ctx, h.panelStateScope(ctx, itemKey))
		delete(value.retainedPanels, itemKey)
	}
	for itemKey := range value.renderedPanels {
		if _, exists := value.itemKeys[itemKey]; exists {
			continue
		}
		delete(value.renderedPanels, itemKey)
	}
}

func (h PanelWidget) normalizeSelection(value *panelHostState, cfg disclosure.Config[string], selected string) string {
	if item, ok := h.item(selected); ok && !item.Disabled {
		return selected
	}
	for _, item := range h.items {
		if item.Disabled {
			continue
		}
		if selected != item.Key {
			value.disclosure.Request(cfg, item.Key)
		}
		if h.hasSelectedKey {
			return item.Key
		}
		return value.disclosure.Current(cfg)
	}
	return ""
}

func (h PanelWidget) item(key string) (PanelItem, bool) {
	for _, item := range h.items {
		if item.Key == key {
			return item, true
		}
	}
	return PanelItem{}, false
}

func (h PanelWidget) panelStateScope(ctx *core.Context, itemKey string) string {
	return core.DerivedKey(ctx, core.FullKey(ctx, h.key), "panel-state:"+itemKey)
}

func (h PanelWidget) layoutForcedPanels(ctx *core.Context, gtx layout.Context, value *panelHostState, selectedKey string) {
	for _, item := range h.items {
		if item.Key == selectedKey || item.Content == nil {
			continue
		}
		if _, rendered := value.renderedPanels[item.Key]; rendered {
			continue
		}
		scope := h.panelStateScope(ctx, item.Key)
		core.LayoutHidden(ctx, gtx, scope, core.WidgetFunc(func(ctx *core.Context, hiddenGtx layout.Context) layout.Dimensions {
			return h.layoutPanel(ctx, hiddenGtx, item)
		}))
		value.renderedPanels[item.Key] = struct{}{}
	}
}

func (h PanelWidget) layoutPanel(ctx *core.Context, gtx layout.Context, item PanelItem) layout.Dimensions {
	scope := h.panelStateScope(ctx, item.Key)
	restoreHost := core.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, item.Key)
	defer restoreItem()
	if (h.keepAlive || h.forceRender) && !h.destroyOnHidden {
		restoreRetention := core.PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	return item.Content.Layout(ctx, gtx)
}

func (h PanelWidget) preparePanelAssociations(ctx *core.Context, item PanelItem) {
	restoreHost := core.PushKey(ctx, h.key)
	defer restoreHost()
	restoreItem := core.PushKey(ctx, item.Key)
	defer restoreItem()
	layoutui.PrepareFieldAssociations(ctx, item.Content)
}
