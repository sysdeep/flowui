package datepicker

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/internal/fieldstyle"
)

// Variant selects the built-in field surface. It is compatible with input.Variant.
type Variant = fieldstyle.Variant

const (
	Primary    = fieldstyle.Primary
	Secondary  = fieldstyle.Secondary
	Filled     = fieldstyle.Filled
	Borderless = fieldstyle.Borderless
	Underlined = fieldstyle.Underlined
)

// DatePart identifies one editable part of a date.
type DatePart uint8

const (
	DatePartYear DatePart = iota
	DatePartMonth
	DatePartDay
)

// Calendar and date field parts are the semantic style customization points.
const (
	PartCalendarPanel      style.Part = style.PartPanel
	PartCalendarNav        style.Part = "datepicker-nav"
	PartCalendarHeader     style.Part = "datepicker-header"
	PartCalendarWeekday    style.Part = "datepicker-weekday"
	PartCalendarCell       style.Part = "datepicker-cell"
	PartCalendarRange      style.Part = "datepicker-range"
	PartCalendarFocus      style.Part = "datepicker-focus"
	PartDateSegment        style.Part = "datepicker-date-segment"
	PartDateLiteral        style.Part = "datepicker-date-literal"
	PartDateRangeSeparator style.Part = "datepicker-range-separator"
	PartDateTrigger        style.Part = "datepicker-date-trigger"
)
