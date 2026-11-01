package progress

import (
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
)

// Meter creates a determinate progress bar with meter semantics.
// Meter intentionally returns BarWidget because a meter shares the bar's
// visual and interaction surface; only its accessibility defaults differ.
func Meter(value float64) BarWidget {
	return Bar(value).Key(core.AutoKey("progress-meter")).meterEnabled()
}

func (p BarWidget) meterEnabled() BarWidget {
	p.meter = true
	p.semanticKey = i18n.MeterDefault
	return p
}
