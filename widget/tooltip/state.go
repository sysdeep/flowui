package tooltip

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
)

const stateSlotTooltip = "tooltip"

const tooltipExclusiveGroup = "tooltip"

func tooltipStateFor(ctx *core.Context, key string) (string, *tooltipState) {
	key = core.ClaimKey(ctx, core.KindTooltip, key)
	value := core.UseState[tooltipState](ctx, key, stateSlotTooltip)
	core.RegisterExclusive(ctx, tooltipExclusiveGroup, key, value.closeForPeer)
	return key, value
}

type tooltipState struct {
	base.TooltipInteraction
	popup       PopupTransition
	visualCache tooltipDefaultsCache
}

func (s *tooltipState) closeForPeer() { s.Close() }

func (s *tooltipState) update(gtx layout.Context, trigger Trigger, disabled bool, delay, closeDelay time.Duration) {
	s.TooltipInteraction.Update(gtx, s, base.TooltipPolicy{
		Focus: trigger == Focus, Disabled: disabled, Visible: s.popup.Value() > 0,
		Delay: delay, CloseDelay: closeDelay,
	})
}

func (s *tooltipState) addInput(gtx layout.Context, size image.Point, disabled bool) {
	s.TooltipInteraction.AddInput(gtx, s, size, disabled)
}

func (s *tooltipState) progress(gtx layout.Context, motion theme.MotionTheme, enter, exit time.Duration) float32 {
	return s.popup.ProgressWithDurations(gtx, s.Open, enter, exit, motion)
}

func (s *tooltipState) exiting() bool {
	return s.popup.Exiting()
}
