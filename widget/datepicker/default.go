package datepicker

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/internal/fieldstyle"
)

const (
	dateDatePickerPopoverInDuration  = 150 * time.Millisecond
	dateDatePickerPopoverOutDuration = 100 * time.Millisecond
	dateDatePickerCellColorDuration  = 100 * time.Millisecond
	dateDatePickerPressInDuration    = 90 * time.Millisecond
	dateDatePickerPressOutDuration   = 160 * time.Millisecond
	dateDatePickerPressScale         = 0.95
	dateDatePickerPopoverScale       = 0.95
)

// datePickerGeometry is the theme-derived geometry and typography consumed by
// date-picker layout. It is part of the cached defaults rather than a request
// type: behavior reads these values to measure and hit-test, while the default
// declarations own the visual scale.
type datePickerGeometry struct {
	Height             unit.Dp
	Radius             unit.Dp
	RangeRadius        unit.Dp
	TextSize           unit.Sp
	FieldGap           unit.Dp
	FieldPaddingX      unit.Dp
	FocusRingWidth     unit.Dp
	SegmentHeight      unit.Dp
	SegmentRadius      unit.Dp
	YearSegmentWidth   unit.Dp
	SegmentWidth       unit.Dp
	SeparatorWidth     unit.Dp
	RangeSeparatorSize unit.Dp
	TriggerWidth       unit.Dp
	IconSize           unit.Dp
	IconRadius         unit.Dp
	IconStroke         unit.Dp
	PopoverGap         unit.Dp
	PopoverPadding     unit.Dp
	PopoverRadius      unit.Dp
	PopoverMaxHeight   unit.Dp
	CalendarWidth      unit.Dp
	CellSize           unit.Dp
	MonthCellHeight    unit.Dp
	YearCellHeight     unit.Dp
	YearGridGap        unit.Dp
	NavButtonSize      unit.Dp
	NavChevronStroke   unit.Dp
	HeaderGap          unit.Dp
	HeaderBottom       unit.Dp
	CalendarTopGap     unit.Dp
	WeekdayGap         unit.Dp
	HeaderIconSize     unit.Dp
	HeaderTextSize     unit.Sp
	WeekdayTextSize    unit.Sp
	CellTextSize       unit.Sp
	CellStrikeWidth    unit.Dp
	CellStrikeHalfSize unit.Dp
}

var defaultDatePickerGeometry = datePickerGeometry{
	Height: 32, Radius: 4, RangeRadius: 4, TextSize: 14, FieldGap: 4,
	FieldPaddingX: 8, SegmentHeight: 24, SegmentRadius: 4,
	YearSegmentWidth: 38, SegmentWidth: 24, SeparatorWidth: 8,
	RangeSeparatorSize: 20, TriggerWidth: 28, IconSize: 16, IconRadius: 3,
	IconStroke: 1.4, PopoverGap: 4, PopoverPadding: 10, PopoverRadius: 8,
	PopoverMaxHeight: 320, CalendarWidth: 224, CellSize: 32, MonthCellHeight: 40,
	YearCellHeight: 30, YearGridGap: 4, NavButtonSize: 22, NavChevronStroke: 1.8,
	HeaderGap: 4, HeaderBottom: 8, CalendarTopGap: 6, WeekdayGap: 2,
	HeaderIconSize: 14, HeaderTextSize: 15, WeekdayTextSize: 11,
	CellTextSize: 14, CellStrikeWidth: 1, CellStrikeHalfSize: 7,
}

const datePickerPopoverBorderWidth unit.Dp = 1

// datePickerMotionDefaults is the transition policy used by date-picker
// animation. Durations are raw theme values; consumers apply the motion policy
// through the animation helpers so a disabled or scaled theme is respected.
type datePickerMotionDefaults struct {
	Policy       theme.MotionTheme
	PopoverIn    time.Duration
	PopoverOut   time.Duration
	CellColor    time.Duration
	PressIn      time.Duration
	PressOut     time.Duration
	PressScale   float32
	PopoverScale float32
}

// datePickerDefaults is the theme-derived appearance and geometry of one date
// picker variant. It contains no instance declaration: Style is resolved
// separately as the last cascade layer.
type datePickerDefaults struct {
	field    style.Style
	calendar style.Style
	metrics  datePickerGeometry
	motion   datePickerMotionDefaults
}

// buildDatePickerDefaults derives the field face, calendar declarations,
// geometry, and motion policy from the active theme.
func buildDatePickerDefaults(active *theme.Theme, variant Variant) datePickerDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	tokens := defaultDatePickerGeometry
	return datePickerDefaults{
		field: fieldstyle.Defaults(active, variant, fieldstyle.Geometry{
			TargetPart: style.PartContent,
			Radius:     tokens.Radius,
		}),
		calendar: dateCalendarDeclaration(active),
		metrics:  datePickerGeometryFor(active),
		motion: datePickerMotionDefaults{
			Policy:       active.Motion,
			PopoverIn:    dateDatePickerPopoverInDuration,
			PopoverOut:   dateDatePickerPopoverOutDuration,
			CellColor:    dateDatePickerCellColorDuration,
			PressIn:      dateDatePickerPressInDuration,
			PressOut:     dateDatePickerPressOutDuration,
			PressScale:   dateDatePickerPressScale,
			PopoverScale: dateDatePickerPopoverScale,
		},
	}
}

func dateResolveTextColor(value style.ResolvedStyle) color.NRGBA {
	if value.Text == nil {
		panic("flowui/datepicker: missing text style")
	}
	resolved, ok := computed.Color(value.Text.Color)
	if !ok {
		panic("flowui/datepicker: unresolved text color")
	}
	return resolved
}

func dateResolvePaintColor(value style.ResolvedStyle) color.NRGBA {
	if value.Paint == nil {
		panic("flowui/datepicker: missing paint style")
	}
	resolved, ok := computed.Brush(value.Paint.Background)
	if !ok {
		panic("flowui/datepicker: unresolved background paint")
	}
	return resolved.ColorAt(.5)
}

// dateResolveOutlineColor returns the accent a focus ring draws with. A field
// family may express its focus ring as an outer outline or as an inset border,
// so the accent is whichever stroke the resolved part actually carries; the
// calendar's own focus part always provides the outline.
func dateResolveOutlineColor(value style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if value.Paint == nil {
		return fallback
	}
	if value.Paint.Outline != nil {
		if resolved, ok := computed.Color(value.Paint.Outline.Color); ok {
			return resolved
		}
	}
	if value.Paint.Border != nil {
		if resolved, ok := computed.Color(value.Paint.Border.Color); ok {
			return resolved
		}
	}
	return fallback
}

func dateResolveFieldColors(content, placeholder, selection style.ResolvedStyle) dateFieldColors {
	return dateFieldColors{
		Foreground:  dateResolveTextColor(content),
		Placeholder: dateResolveTextColor(placeholder),
		Selection:   dateResolvePaintColor(selection),
	}
}

// datePickerGeometryFor returns the measurements used by date-picker layout.
func datePickerGeometryFor(active *theme.Theme) datePickerGeometry {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	return defaultDatePickerGeometry
}

// dateCalendarDeclaration contains all static calendar and date-part defaults.
func dateCalendarDeclaration(active *theme.Theme) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	tokens := defaultDatePickerGeometry
	danger := sem.Tone(theme.ToneDanger).Fill
	return style.Style{}.
		Part(PartCalendarPanel, datePickerPanel(active, tokens)).
		Part(style.PartDescription, style.Style{}.
			FontSize(active.Typography.SmallSize).
			TextColor(style.Solid(sem.MutedForeground)).
			On(style.Invalid, style.Style{}.TextColor(style.Solid(danger))).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(sem.MutedForeground)))).
			On(style.Invalid|style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(danger))))).
		Part(PartCalendarHeader, style.Style{}.
			FontSize(tokens.HeaderTextSize).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.Foreground)).
			On(style.Selected, style.Style{}.
				TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnSoft)))).
		Part(PartCalendarWeekday, style.Style{}.
			FontSize(tokens.WeekdayTextSize).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.MutedForeground))).
		Part(PartCalendarNav, dateNavStyle(active)).
		Part(PartCalendarCell, dateCellStyle(active)).
		Part(PartCalendarRange, style.Style{}.
			Background(style.Solid(sem.Tone(theme.TonePrimary).Soft))).
		Part(PartCalendarFocus, style.Style{}.
			TextColor(style.Solid(sem.Ring)).
			Outline(1, 0, style.Solid(sem.Ring))).
		Part(PartDateSegment, style.Style{}.
			FontSize(tokens.TextSize).
			FontWeight(int(font.Medium)).
			Background(style.RGBA(0)).
			TextColor(style.Solid(sem.FieldForeground))).
		Part(PartDateLiteral, style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(PartDateRangeSeparator, style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(PartDateTrigger, style.Style{}.
			TextColor(style.Solid(sem.FieldPlaceholder)))
}

func datePickerPanel(active *theme.Theme, metrics datePickerGeometry) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Background(style.Solid(sem.Overlay)).
		TextColor(style.Solid(sem.OverlayForeground)).
		BorderColor(style.Solid(sem.Border)).
		BorderWidth(datePickerPopoverBorderWidth).
		Radius(metrics.PopoverRadius).
		Padding(metrics.PopoverPadding).
		Shadow(style.ShadowMenu).
		Overflow(style.OverflowHidden)
}

// dateCellDeclaration applies the calendar-domain flags of one cell to its
// dynamic declaration. Interaction state remains in the resolver's style.State.
func dateCellDeclaration(active *theme.Theme, cell dateCellState) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	sem := active.Semantic()
	background := color.NRGBA{}
	foreground := sem.Foreground
	if cell.Hovered || cell.Pressed {
		background = sem.SurfaceRaised
	}
	if cell.Today {
		background = sem.Tone(theme.TonePrimary).Soft
		foreground = sem.Tone(theme.TonePrimary).OnSoft
	}
	if cell.Selected {
		background = sem.Tone(theme.TonePrimary).Fill
		foreground = sem.Tone(theme.TonePrimary).OnFill
	}
	if cell.Outside {
		foreground = active.DisabledColor(sem.MutedForeground)
		if cell.Selected {
			background = sem.SurfaceRaised
		}
	}
	if cell.Active && !cell.Selected && !cell.Disabled {
		background = sem.SurfaceRaised
		if !cell.Today {
			foreground = sem.Tone(theme.TonePrimary).OnSoft
		}
	}
	if cell.Range && !cell.Selected {
		background = color.NRGBA{}
		if !cell.Outside {
			foreground = sem.Tone(theme.TonePrimary).OnSoft
		}
	}
	if cell.Disabled {
		background = active.DisabledColor(background)
		foreground = active.DisabledColor(foreground)
	}
	return style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		FontWeight(int(font.Medium))
}

// dateRangeDeclaration returns the range-track declaration for a calendar cell.
func dateRangeDeclaration(active *theme.Theme, outside bool) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	value := active.Semantic().Tone(theme.TonePrimary).Soft
	if outside {
		value = active.Semantic().SurfaceRaised
		value.A = byte(uint16(value.A) / 5)
	}
	return style.Style{}.Background(style.Solid(value))
}

// dateSegmentDeclaration returns the field-derived date segment declaration.
func dateSegmentDeclaration(active *theme.Theme, base color.NRGBA) style.Style {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	danger := active.Semantic().Tone(theme.ToneDanger)
	return style.Style{}.
		TextColor(style.Solid(base)).
		On(style.Focused, style.Style{}.
			TextColor(style.Solid(active.Semantic().Tone(theme.TonePrimary).OnSoft)).
			Background(style.Solid(active.Semantic().Tone(theme.TonePrimary).Soft))).
		On(style.Invalid, style.Style{}.
			TextColor(style.Solid(danger.Fill))).
		On(style.Invalid|style.Focused, style.Style{}.
			TextColor(style.Solid(danger.OnSoft)).
			Background(style.Solid(danger.Soft)))
}

// dateTextPartDeclaration returns the simple text color used by date literals
// and the range separator.
func dateTextPartDeclaration(base color.NRGBA) style.Style {
	return style.Style{}.TextColor(style.Solid(base))
}

func dateNavStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	rest := theme.Composite(sem.Overlay, sem.Foreground, .04)
	return style.Style{}.
		Radius(defaultDatePickerGeometry.NavButtonSize/2).
		TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnSoft)).
		Background(style.Solid(rest)).
		On(style.Hovered, style.Style{}.Background(style.Solid(sem.SurfaceRaised))).
		On(style.Pressed, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
		On(style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(sem.Tone(theme.TonePrimary).OnSoft))).
			Background(style.Solid(active.DisabledColor(rest))))
}

func dateCellStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		FontSize(defaultDatePickerGeometry.CellTextSize).
		FontWeight(int(font.Medium)).
		Radius(defaultDatePickerGeometry.CellSize/2).
		Background(style.RGBA(0)).
		TextColor(style.Solid(sem.Foreground)).
		On(style.Hovered, style.Style{}.Background(style.Solid(sem.SurfaceRaised))).
		On(style.Pressed, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
		On(style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(sem.Foreground))).
			Background(style.Solid(active.DisabledColor(sem.SurfaceRaised))))
}
