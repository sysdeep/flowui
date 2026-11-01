package slider

import (
	"math"
	"strconv"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type Widget struct {
	style           style.Style
	key             string
	value           float64
	hasValue        bool
	hasDefault      bool
	upperValue      float64
	hasUpperValue   bool
	hasDefaultUpper bool
	minValue        float64
	maxValue        float64
	step            float64
	label           string
	valueText       string
	hasValueText    bool
	showValue       bool
	rangeMode       bool
	disabled        bool
	orientation     Orientation
	formatValue     func(float64) string
	onChange        func(float64)
	onRangeChange   func(float64, float64)
}

func New(value float64) Widget {
	return Widget{
		key:      core.AutoKey("slider"),
		value:    value,
		maxValue: 100,
		step:     1,
	}
}

func (s Widget) Key(key string) Widget {
	s.key = key
	return s
}

func RangeSlider(lowerValue, upperValue float64) Widget {
	return Widget{
		key:        core.AutoKey("range-slider"),
		value:      lowerValue,
		upperValue: upperValue,
		maxValue:   100,
		step:       1,
		rangeMode:  true,
	}
}

func (s Widget) Range(minValue, maxValue float64) Widget {
	if math.IsNaN(minValue) || math.IsInf(minValue, 0) || math.IsNaN(maxValue) || math.IsInf(maxValue, 0) || maxValue <= minValue {
		panic("flowui: slider maximum must be greater than minimum")
	}
	s.minValue = minValue
	s.maxValue = maxValue
	return s
}

func (s Widget) Step(step float64) Widget {
	if math.IsNaN(step) || math.IsInf(step, 0) || step <= 0 {
		panic("flowui: slider step must be positive")
	}
	s.step = step
	return s
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) ShowValue() Widget {
	s.showValue = true
	return s
}

func (s Widget) ValueText(text string) Widget {
	s.valueText = text
	s.hasValueText = true
	s.showValue = true
	return s
}

func (s Widget) FormatValue(format func(float64) string) Widget {
	s.formatValue = format
	return s
}

func (s Widget) OnChange(fn func(float64)) Widget {
	s.onChange = fn
	return s
}

func (s Widget) Value(value float64) Widget {
	s.value = value
	return s
}

func (s Widget) UpperValue(value float64) Widget {
	s.upperValue = value
	return s
}

func (s Widget) OnRangeChange(fn func(float64, float64)) Widget {
	s.onRangeChange = fn
	return s
}

func (s Widget) Orientation(orientation Orientation) Widget {
	s.orientation = orientation
	return s
}

func (s Widget) Vertical() Widget {
	s.orientation = Vertical
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if s.rangeMode {
		return s.layoutBaseRange(ctx, gtx)
	}
	return s.layoutBase(ctx, gtx)
}

func (s Widget) axis() layout.Axis {
	if s.orientation == Vertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

func (s Widget) outputText(values sliderValues) string {
	if s.hasValueText {
		return s.valueText
	}
	if !s.showValue {
		return ""
	}
	if values.rangeMode {
		return s.format(values.lower) + " - " + s.format(values.upper)
	}
	return s.format(values.lower)
}

func (s Widget) format(value float64) string {
	if s.formatValue != nil {
		return s.formatValue(value)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (s Widget) semanticDescriptionFor(ctx *core.Context, values sliderValues) string {
	label := s.label
	if label == "" {
		label = ctx.Text(i18n.SliderDefault)
	}
	output := s.outputText(values)
	if output == "" {
		if values.rangeMode {
			output = s.format(values.lower) + " - " + s.format(values.upper)
		} else {
			output = s.format(values.lower)
		}
	}
	return ctx.Textf(i18n.SliderValue, label, output)
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (s Widget) Style(values ...style.Style) Widget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}
