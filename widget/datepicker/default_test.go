package datepicker

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

func TestDatePickerPanelUsesMenuElevation(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		defaults := buildDatePickerDefaults(&active, Primary)
		panel := style.CascadePart(style.State{}, PartCalendarPanel, defaults.calendar)
		if panel.Paint == nil || panel.Paint.Border == nil || panel.Paint.Border.Width == nil {
			t.Fatal("date picker panel has no border")
		}
		if got := *panel.Paint.Border.Width; got != datePickerPopoverBorderWidth {
			t.Fatalf("date picker panel border = %v, want %v", got, datePickerPopoverBorderWidth)
		}
		if len(panel.Paint.Shadows) == 0 || panel.Paint.Shadows[0].Profile == nil ||
			*panel.Paint.Shadows[0].Profile != style.ShadowMenu {
			t.Fatal("date picker panel does not use the menu shadow profile")
		}
	}
}

func TestDatePickerOutsideCellIsMuted(t *testing.T) {
	active := theme.Light()
	current := dateCellDeclaration(&active, dateCellState{}).Resolve(style.State{})
	outside := dateCellDeclaration(&active, dateCellState{Outside: true}).Resolve(style.State{})
	currentColor, ok := resolve.Color(current.Text.Color)
	if !ok {
		t.Fatal("current-month cell has no text colour")
	}
	outsideColor, ok := resolve.Color(outside.Text.Color)
	if !ok {
		t.Fatal("outside-month cell has no text colour")
	}
	if outsideColor != active.DisabledColor(active.Semantic().MutedForeground) ||
		currentColor != active.Semantic().Foreground {
		t.Fatalf("day colours = current %v / outside %v, want semantic foregrounds", currentColor, outsideColor)
	}
	if theme.ContrastRatio(outsideColor, active.Semantic().Background) >=
		theme.ContrastRatio(currentColor, active.Semantic().Background) {
		t.Fatalf("outside-month text = %v, want a muted colour than %v", outsideColor, currentColor)
	}
}

func TestDatePickerSelectedCellUsesWhiteOnPrimary(t *testing.T) {
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		cell := dateCellDeclaration(&active, dateCellState{Selected: true}).Resolve(style.State{Selected: true})
		foreground, ok := resolve.Color(cell.Text.Color)
		if !ok {
			t.Fatal("selected cell has no text colour")
		}
		if foreground != white {
			t.Fatalf("selected date text = %v, want white", foreground)
		}
		primary := active.Semantic().Tone(theme.TonePrimary)
		if ratio := theme.ContrastRatio(foreground, primary.Fill); ratio < 4.5 {
			t.Fatalf("selected date contrast = %.2f, want at least 4.5", ratio)
		}
	}
}

func TestDatePickerResolvedSelectedCellUsesOnFill(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	active := theme.Light()
	defaults := buildDatePickerDefaults(&active, Primary)
	base := defaults.calendar
	cellDefaults := style.JoinPart(base,
		style.Style{}.Part(PartCalendarCell, dateCellDeclaration(&active, dateCellState{Selected: true})),
		PartCalendarCell,
	)
	resolved := New(testDate(2026, 7, 1)).Key("date").resolvePartWithDefaults(
		ctx, gtx, PartCalendarCell, style.State{Selected: true}, cellDefaults,
	)
	foreground, ok := resolve.Color(resolved.Text.Color)
	if !ok {
		t.Fatal("resolved selected cell has no text colour")
	}
	want := active.Semantic().Tone(theme.TonePrimary).OnFill
	if foreground != want {
		t.Fatalf("resolved selected date text = %v, want %v", foreground, want)
	}
}

func TestDatePickerSelectedCellBeatsInheritedForeground(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	active := theme.Light()
	defaults := buildDatePickerDefaults(&active, Primary)
	restore := core.PushInheritedStyle(ctx, style.Style{}.TextColor(style.RGB(0x111111)))
	defer restore()
	cellDefaults := style.JoinPart(defaults.calendar,
		style.Style{}.Part(PartCalendarCell, dateCellDeclaration(&active, dateCellState{Selected: true})),
		PartCalendarCell,
	)

	resolved := New(testDate(2026, 7, 1)).Key("date").resolvePartWithDefaults(
		ctx, gtx, PartCalendarCell, style.State{Selected: true}, cellDefaults,
	)
	foreground, ok := resolve.Color(resolved.Text.Color)
	if !ok {
		t.Fatal("resolved selected cell has no text colour")
	}
	want := active.Semantic().Tone(theme.TonePrimary).OnFill
	if foreground != want {
		t.Fatalf("inherited foreground overrode selected date text = %v, want %v", foreground, want)
	}
}
