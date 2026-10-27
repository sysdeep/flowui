package toggle

import "github.com/qianniancn/flowui/internal/core"

const toggleStateSlotToggleButtonVisual = "togglebutton-visual"

type toggleToggleButtonState struct {
	visualCache ToggleButtonVisualCache
}

func toggleToggleButtonStateFor(ctx *core.Context, key string) *toggleToggleButtonState {
	return core.UseState[toggleToggleButtonState](ctx, key, toggleStateSlotToggleButtonVisual)
}
