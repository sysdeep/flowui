package navigation

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// PaginationButtonKind identifies the semantic slot being painted.
type PaginationButtonKind uint8

const (
	PaginationPage PaginationButtonKind = iota
	PaginationPrevious
	PaginationNext
	PaginationEllipsis
)

// paginationMetrics contains only geometry consumed by pagination layout.
type paginationMetrics struct {
	Size         unit.Dp
	ItemGap      unit.Dp
	ContentGap   unit.Dp
	CompactWidth unit.Dp
}

// paginationButtonRequest carries behavior-owned state for one pagination slot.
type paginationButtonRequest struct {
	Kind         PaginationButtonKind
	Label        string
	Active       bool
	Disabled     bool
	Hovered      bool
	Pressed      bool
	FocusVisible bool
	Focus        float32
}

// paginationVisual is the resolved appearance of one pagination configuration.
type paginationVisual struct {
	Defaults      style.Style
	Summary       style.Style
	Metrics       paginationMetrics
	RenderMetrics paginationRenderMetrics
	Motion        theme.MotionTheme
}

type paginationVisualKey struct {
	size PaginationSize
}

type paginationVisualCache = computed.VisualCache[paginationVisualKey, paginationVisual]

func (p PaginationWidget) visualWithCache(ctx *core.Context) paginationVisual {
	active := core.ActiveTheme(ctx)
	key := paginationVisualKey{size: p.size}
	load := func() paginationVisual { return buildPagination(active, p.size) }
	if p.visualCache == nil {
		return load()
	}
	return p.visualCache.Resolve(active, key, load)
}
