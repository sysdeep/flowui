package piechart

import (
	"image"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

const stateSlotPieChart = "pie-chart"

type chartState struct {
	click             gesture.Click
	pointerTag        struct{}
	hovered           bool
	pointer           f32.Point
	animation         chart.DataAnimation[chartData]
	dataCache         chartDataCache
	geometryCache     chart.GeometryCache[chartGeometry, geometryKey]
	legendItems       map[string]*chart.LegendItem
	legendFrame       map[string]struct{}
	tooltipTransition tooltip.PopupTransition
	tooltipSlice      resolvedSlice
}

func chartStateFor(ctx *core.Context, key string) *chartState {
	key = core.ClaimKey(ctx, core.KindPieChart, key)
	return core.UseState[chartState](ctx, key, stateSlotPieChart)
}

func (s *chartState) beginLegendFrame() {
	if s.legendFrame == nil {
		s.legendFrame = make(map[string]struct{})
	} else {
		clear(s.legendFrame)
	}
}

func (s *chartState) endLegendFrame() {
	for key := range s.legendItems {
		if _, ok := s.legendFrame[key]; !ok {
			delete(s.legendItems, key)
		}
	}
}

func (s *chartState) legendItem(key string) *chart.LegendItem {
	if s.legendItems == nil {
		s.legendItems = make(map[string]*chart.LegendItem)
	}
	s.legendFrame[key] = struct{}{}
	if item := s.legendItems[key]; item != nil {
		return item
	}
	item := new(chart.LegendItem)
	s.legendItems[key] = item
	return item
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool) {
	chart.UpdateHoverPointer(gtx, enabled, &s.pointerTag, &s.hovered, &s.pointer)
}

func (s *chartState) addPointerInput(gtx layout.Context, area image.Rectangle, enabled bool) {
	chart.AddPointerInputWithCursor(gtx, area, enabled, pointer.CursorPointer, &s.pointerTag)
}

func (s *chartState) updateClick(gtx layout.Context, enabled bool) bool {
	activated, _ := chart.UpdateClicks(gtx, enabled, &s.click)
	return activated
}

func (s *chartState) addClickInput(gtx layout.Context, area image.Rectangle, enabled bool) {
	chart.AddClickInputRect(gtx, area, enabled, &s.click)
}

func (s *chartState) clearTooltip() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSlice = resolvedSlice{}
}
