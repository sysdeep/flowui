package input

import (
	"image/color"
	"testing"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

func TestDefaultInputStylesResolve(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		for _, variant := range []Variant{Primary, Secondary, Filled, Borderless, Underlined} {
			for _, size := range []Size{Small, Medium, Large} {
				for _, defaults := range []style.Style{
					defaultInputRules(&active, variant, size, false),
					defaultTextAreaRules(&active, variant, false, 48),
					defaultGroupRules(&active, variant, false, 48),
				} {
					resolved := defaults.Resolve(style.State{})
					if resolved.Box == nil || resolved.Text == nil || resolved.Paint == nil {
						t.Fatalf("%v/%v defaults do not resolve a complete field", variant, size)
					}
				}
			}
		}
	}
}

func TestInputFieldStateFace(t *testing.T) {
	for _, test := range []struct {
		name  string
		theme theme.Theme
	}{{"light", theme.Light()}, {"dark", theme.Dark()}} {
		t.Run(test.name, func(t *testing.T) {
			active := test.theme
			sem := active.Semantic()
			defaults := defaultInputRules(&active, Primary, Medium, false)
			resolved := func(state style.State) style.ResolvedStyle {
				return defaults.Resolve(state)
			}
			colorOf := func(source style.ColorSource) color.NRGBA {
				value, ok := resolve.Color(source)
				if !ok {
					t.Fatalf("field colour %v is not solid", source)
				}
				return value
			}
			paintColor := func(source style.PaintSource) color.NRGBA {
				brush, ok := resolve.Brush(source)
				if !ok {
					t.Fatalf("field paint %v is empty", source)
				}
				return brush.ColorAt(.5)
			}
			stroke := func(state style.State) (unit.Dp, color.NRGBA, style.ResolvedStyle) {
				value := resolved(state)
				if value.Paint == nil || value.Paint.Border == nil || value.Paint.Border.Width == nil {
					t.Fatalf("state %+v has no field stroke", state)
				}
				return *value.Paint.Border.Width, colorOf(value.Paint.Border.Color), value
			}

			restingWidth, restingColor, _ := stroke(style.State{})
			if restingWidth != unit.Dp(1) {
				t.Fatalf("resting width = %v, want %v", restingWidth, unit.Dp(1))
			}
			if restingColor != sem.Border {
				t.Fatalf("resting colour = %v, want border %v", restingColor, sem.Border)
			}

			// A bordered field darkens its stroke on hover; only focus and
			// validation change the stroke colour.
			hoverWidth, hoverColor, _ := stroke(style.State{Hovered: true})
			if hoverWidth != restingWidth || hoverColor != theme.Mix(sem.Border, sem.Foreground, 18) {
				t.Fatalf("hover stroke = %v/%v, want darkened border", hoverWidth, hoverColor)
			}

			// Focus draws an inset stroke in the ring colour plus the soft glow
			// declared by the component; it never paints an outline outside
			// the field.
			focusWidth, focusColor, focus := stroke(style.State{Focused: true})
			if focus.Paint.Outline != nil {
				t.Fatal("focused field paints an outline outside its box")
			}
			if len(focus.Paint.Shadows) == 0 {
				t.Fatal("focused field has no glow")
			}
			for _, shadow := range focus.Paint.Shadows {
				if !shadow.Transient {
					t.Fatal("focus glow must not reserve layout space")
				}
			}
			if focusWidth != unit.Dp(1) {
				t.Fatalf("focus width = %v, want ring width %v", focusWidth, unit.Dp(1))
			}
			if focusColor != sem.Ring {
				t.Fatalf("focus colour = %v, want ring %v", focusColor, sem.Ring)
			}

			// Validation keeps the resting stroke width and turns the stroke itself
			// into the tone colour; focus adopts the focus ring token.
			invalidWidth, invalidColor, invalid := stroke(style.State{Invalid: true})
			if invalid.Paint.Outline != nil {
				t.Fatal("invalid field paints an outline outside its box")
			}
			if invalidWidth != max(unit.Dp(1), 1) {
				t.Fatalf("invalid width = %v, want the resting width", invalidWidth)
			}
			if want := sem.Tone(theme.ToneDanger).Fill; invalidColor != want {
				t.Fatalf("invalid colour = %v, want danger %v", invalidColor, want)
			}
			invalidFocusWidth, invalidFocusColor, _ := stroke(style.State{Invalid: true, Focused: true})
			if invalidFocusWidth != unit.Dp(1) || invalidFocusColor != invalidColor {
				t.Fatalf("invalid focus stroke = %v/%v, want %v/%v",
					invalidFocusWidth, invalidFocusColor, unit.Dp(1), invalidColor)
			}

			// Read-only and disabled fields share the disabled face so both read as
			// non-editable instead of looking like an ordinary field.
			for _, state := range []style.State{{ReadOnly: true}, {Disabled: true}} {
				value := resolved(state)
				if got := paintColor(value.Paint.Background); got != sem.DisabledBackground {
					t.Fatalf("state %+v background = %v, want disabled %v", state, got, sem.DisabledBackground)
				}
				if got := colorOf(value.Text.Color); got != sem.DisabledForeground {
					t.Fatalf("state %+v text = %v, want disabled %v", state, got, sem.DisabledForeground)
				}
			}
		})
	}
}
