package datepicker

import (
	"image"
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/internal/binding"
	"github.com/qianniancn/flowui/widget/internal/optionrow"
)

const stateSlotDatePicker = "datepicker"

func datePickerStateFor(ctx *core.Context, key string) *datePickerState {
	key = core.ClaimKey(ctx, core.KindDatePicker, key)
	return core.UseState[datePickerState](ctx, key, stateSlotDatePicker)
}

type datePickerState struct {
	valueBinding      binding.Binding[time.Time]
	segments          dateSegmentsState
	hover             base.HoverInteraction
	trigger           widget.Clickable
	popup             base.PopupInteraction
	header            widget.Clickable
	prev              widget.Clickable
	next              widget.Clickable
	open              bool
	viewMode          datePickerViewMode
	viewMonth         time.Time
	syncedValue       time.Time
	monthReady        bool
	popoverTransition animation.FloatTransition
	prevBackground    animation.ColorTransition
	nextBackground    animation.ColorTransition
	yearList          layout.List
	yearBar           widget.Scrollbar
	yearVisualOutset  layoutui.VisualOutsetState
	yearScrollYear    int
	yearScrollReady   bool
	days              map[string]*datePickerCellState
	months            map[string]*datePickerCellState
	years             map[string]*datePickerCellState
	frameDays         map[string]struct{}
	frameMonths       map[string]struct{}
	frameYears        map[string]struct{}
	calendarDays      [42]datePickerDay
	calendarDaysYear  int
	calendarDaysMonth time.Month
	calendarDaysWeek  time.Weekday
	calendarDaysLoc   *time.Location
	calendarDaysValid bool
	calendarStyle     style.Style
	partDefaults      datePickerPartDefaultsCache
	resolvedParts     datePickerResolvedPartCache
	localeCache       datePickerLocaleCache
	calendarLayout    datePickerCalendarLayoutCache
}

// datePickerCalendarLayoutCache owns the fixed Gio child descriptors used by
// the six-week calendar grid. The descriptors only need their slot index;
// frame-specific date, theme-independent geometry, and picker values are
// refreshed before the grid is laid out. Keeping the descriptors on state
// avoids rebuilding 42 escaping closures on every open calendar frame.
type datePickerCalendarLayoutCache struct {
	ctx         *core.Context
	picker      Widget
	state       *datePickerState
	days        []datePickerDay
	cell        int
	metrics     datePickerGeometry
	now         time.Time
	rows        [6]layout.FlexChild
	dayChildren [6][7]layout.FlexChild
	initialized bool
}

func (c *datePickerCalendarLayoutCache) prepare(ctx *core.Context, picker Widget, state *datePickerState, days []datePickerDay, cell int, metrics datePickerGeometry, now time.Time) {
	if !c.initialized {
		c.initialize()
	}
	c.ctx = ctx
	c.picker = picker
	c.state = state
	c.days = days
	c.cell = cell
	c.metrics = metrics
	c.now = now
}

func (c *datePickerCalendarLayoutCache) initialize() {
	for row := range c.rows {
		c.rows[row] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return c.layoutWeek(gtx, row)
		})
		for column := range c.dayChildren[row] {
			c.dayChildren[row][column] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return c.layoutDay(gtx, row, column)
			})
		}
	}
	c.initialized = true
}

func (c *datePickerCalendarLayoutCache) layoutWeek(gtx layout.Context, row int) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, c.dayChildren[row][:]...)
}

func (c *datePickerCalendarLayoutCache) layoutDay(gtx layout.Context, row, column int) layout.Dimensions {
	gtx.Constraints = layout.Exact(image.Pt(c.cell, c.cell))
	index := row*7 + column
	day := c.days[index]
	previousOutside := column > 0 && c.days[index-1].outside
	nextOutside := column < 6 && c.days[index+1].outside
	return c.picker.layoutDay(c.ctx, gtx, c.state, day, column, previousOutside, nextOutside, c.metrics, c.now)
}

func (s *datePickerState) monthDays(month time.Time, weekStart time.Weekday) []datePickerDay {
	month = firstOfMonth(month)
	year, valueMonth, _ := month.Date()
	if s.calendarDaysValid && s.calendarDaysYear == year && s.calendarDaysMonth == valueMonth && s.calendarDaysWeek == weekStart && s.calendarDaysLoc == month.Location() {
		return s.calendarDays[:]
	}
	datePickerMonthDaysInto(s.calendarDays[:], month, weekStart)
	s.calendarDaysYear = year
	s.calendarDaysMonth = valueMonth
	s.calendarDaysWeek = weekStart
	s.calendarDaysLoc = month.Location()
	s.calendarDaysValid = true
	return s.calendarDays[:]
}

func (s *datePickerState) beginFrame() {
	core.BeginFrameMap(&s.frameDays)
	core.BeginFrameMap(&s.frameMonths)
	core.BeginFrameMap(&s.frameYears)
}

func (s *datePickerState) endFrame() {
	core.SweepFrameMap(s.days, s.frameDays)
	core.SweepFrameMap(s.months, s.frameMonths)
	core.SweepFrameMap(s.years, s.frameYears)
}

func (s *datePickerState) sync(value, initialMonth time.Time) {
	value = dateOnly(value)
	s.segments.sync(value)
	if !s.monthReady {
		s.viewMonth = firstOfMonth(initialMonth)
		s.syncedValue = value
		s.monthReady = true
		return
	}
	if sameDate(value, s.syncedValue) {
		return
	}
	s.syncedValue = value
	if !value.IsZero() {
		s.viewMonth = firstOfMonth(value)
	}
}

func (s *datePickerState) toggleYearPicker(year int) {
	if s.viewMode == datePickerViewYears {
		s.viewMode = datePickerViewDays
		return
	}
	s.viewMode = datePickerViewYears
	s.yearScrollYear = year
	s.yearScrollReady = true
}

func (s *datePickerState) openCalendar() {
	if !s.open {
		s.viewMode = datePickerViewDays
	}
	s.open = true
}

func (s *datePickerState) move(delta int) {
	switch s.viewMode {
	case datePickerViewMonths:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(delta, 0, 0))
	case datePickerViewYears:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(delta*datePickerYearSpan, 0, 0))
	default:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(0, delta, 0))
	}
}

func (s *datePickerState) updateFocus(focused, disabled bool) {
	if disabled {
		s.open = false
		return
	}
	if !focused {
		s.open = false
	}
}

func (s *datePickerState) updateKeys(gtx layout.Context, focus event.Tag) {
	filters := []event.Filter{
		key.Filter{Focus: focus, Name: key.NameReturn},
		key.Filter{Focus: focus, Name: key.NameEnter},
		key.Filter{Focus: focus, Name: key.NameSpace},
	}
	if s.open {
		filters = append(filters, key.Filter{Focus: focus, Name: key.NameEscape})
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameReturn, key.NameEnter, key.NameSpace:
			s.openCalendar()
		case key.NameEscape:
			s.open = false
		}
	}
}

func (s *datePickerState) day(key string) *datePickerCellState {
	return core.UseFrameMap(&s.days, &s.frameDays, key)
}

func (s *datePickerState) month(key string) *datePickerCellState {
	return core.UseFrameMap(&s.months, &s.frameMonths, key)
}

func (s *datePickerState) year(key string) *datePickerCellState {
	return core.UseFrameMap(&s.years, &s.frameYears, key)
}

func (s *datePickerState) popoverProgress(gtx layout.Context, open bool, motion datePickerMotionDefaults) float32 {
	target := float32(0)
	duration := motion.PopoverOut
	if open {
		target = 1
		duration = motion.PopoverIn
	}
	return s.popoverTransition.Value(gtx, target, duration, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

func (s *datePickerState) navBackground(gtx layout.Context, delta int, target color.NRGBA, motion datePickerMotionDefaults) color.NRGBA {
	if delta < 0 {
		return s.prevBackground.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
	}
	return s.nextBackground.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

type datePickerCellState struct {
	clickable            widget.Clickable
	date                 time.Time
	backgroundTransition animation.ColorTransition
}

func (s *datePickerCellState) background(gtx layout.Context, target color.NRGBA, motion datePickerMotionDefaults) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

func (s *datePickerState) hoveredDay() time.Time {
	for _, day := range s.days {
		if day.clickable.Hovered() {
			return day.date
		}
	}
	return time.Time{}
}

func (s *datePickerState) calendarFocused(gtx layout.Context) bool {
	return gtx.Focused(&s.header) ||
		gtx.Focused(&s.prev) ||
		gtx.Focused(&s.next) ||
		datePickerCellsFocused(gtx, s.days) ||
		datePickerCellsFocused(gtx, s.months) ||
		datePickerCellsFocused(gtx, s.years)
}

func (s *datePickerState) focusVisible(ctx *core.Context, gtx layout.Context) bool {
	return s.segments.focusVisible(ctx, gtx) ||
		core.FocusVisible(ctx, &s.trigger, gtx.Focused(&s.trigger)) ||
		core.FocusVisible(ctx, &s.header, gtx.Focused(&s.header)) ||
		core.FocusVisible(ctx, &s.prev, gtx.Focused(&s.prev)) ||
		core.FocusVisible(ctx, &s.next, gtx.Focused(&s.next)) ||
		datePickerCellsFocusVisible(ctx, gtx, s.days) ||
		datePickerCellsFocusVisible(ctx, gtx, s.months) ||
		datePickerCellsFocusVisible(ctx, gtx, s.years)
}

func (s *datePickerState) calendarEscapePressed(gtx layout.Context) bool {
	return base.EscapePressed(gtx, &s.header) ||
		base.EscapePressed(gtx, &s.prev) ||
		base.EscapePressed(gtx, &s.next) ||
		datePickerCellsEscapePressed(gtx, s.days) ||
		datePickerCellsEscapePressed(gtx, s.months) ||
		datePickerCellsEscapePressed(gtx, s.years)
}

func datePickerCellsFocused(gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if gtx.Focused(&cell.clickable) {
			return true
		}
	}
	return false
}

func datePickerCellsFocusVisible(ctx *core.Context, gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if core.FocusVisible(ctx, &cell.clickable, gtx.Focused(&cell.clickable)) {
			return true
		}
	}
	return false
}

func datePickerCellsEscapePressed(gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if base.EscapePressed(gtx, &cell.clickable) {
			return true
		}
	}
	return false
}

func datePickerPressScale(gtx layout.Context, history []widget.Press, disabled bool, motion datePickerMotionDefaults) float32 {
	target := motion.PressScale
	if target <= 0 || target > 1 {
		target = 1
	}
	return optionrow.PressScaleWithEasing(
		gtx,
		history,
		disabled,
		target,
		motion.PressIn,
		motion.PressOut,
		animation.EasingFor(motion.Policy.Ease),
		motion.Policy,
	)
}

// Disclosure helpers
func datePickerBindingConfig(widget Widget) binding.Config[time.Time] {
	return binding.Config[time.Time]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *datePickerState) currentValue(widget Widget) time.Time {
	return s.valueBinding.Current(datePickerBindingConfig(widget))
}

func (s *datePickerState) bind(widget Widget) {
	s.valueBinding.Bind(datePickerBindingConfig(widget))
}

func (s *datePickerState) requestValue(widget Widget, value time.Time) time.Time {
	newValue, _ := s.valueBinding.Request(datePickerBindingConfig(widget), value)
	return newValue
}
