package progress

import "github.com/qianniancn/flowui/i18n"

// Meter creates a determinate progress bar with meter semantics.
// Meter intentionally returns BarWidget because a meter shares the bar's
// visual and interaction surface; only its accessibility defaults differ.
func Meter(key string, value float64) BarWidget {
	return Bar(key, value).meterEnabled()
}

func (p BarWidget) meterEnabled() BarWidget {
	p.meter = true
	p.semanticKey = i18n.MeterDefault
	return p
}
