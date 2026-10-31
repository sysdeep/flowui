package colorpicker

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotColorField = "color-field"

type ColorFieldWidget struct {
	style        style.Style
	key          string
	value        color.NRGBA
	label        string
	description  string
	errorMessage string
	onChange     func(color.NRGBA)
	variant      input.Variant
	disabled     bool
	invalid      bool
	required     bool
	fullWidth    bool
	alpha        bool
	swatch       bool
}

type colorFieldState struct {
	text          string
	syncedColor   color.NRGBA
	syncedAlpha   bool
	ready         bool
	valid         bool
	defaultsCache colorFieldDefaultsCache
}

func ColorField(key string, value color.NRGBA) ColorFieldWidget {
	return ColorFieldWidget{key: key, value: value}
}

func (field ColorFieldWidget) Label(label string) ColorFieldWidget {
	field.label = label
	return field
}

func (field ColorFieldWidget) Description(description string) ColorFieldWidget {
	field.description = description
	return field
}

func (field ColorFieldWidget) ErrorMessage(message string) ColorFieldWidget {
	field.errorMessage = message
	return field
}

func (field ColorFieldWidget) OnChange(fn func(color.NRGBA)) ColorFieldWidget {
	field.onChange = fn
	return field
}

func (field ColorFieldWidget) Variant(variant input.Variant) ColorFieldWidget {
	field.variant = variant
	return field
}

func (field ColorFieldWidget) Disabled(disabled bool) ColorFieldWidget {
	field.disabled = disabled
	return field
}

func (field ColorFieldWidget) Invalid(invalid bool) ColorFieldWidget {
	field.invalid = invalid
	return field
}

func (field ColorFieldWidget) Required(required bool) ColorFieldWidget {
	field.required = required
	return field
}

func (field ColorFieldWidget) FillWidth() ColorFieldWidget {
	field.fullWidth = true
	return field
}

func (field ColorFieldWidget) Alpha(enabled bool) ColorFieldWidget {
	field.alpha = enabled
	return field
}

func (field ColorFieldWidget) Swatch(show bool) ColorFieldWidget {
	field.swatch = show
	return field
}

func (field ColorFieldWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return field.layoutWithStyle(ctx, gtx)
}

func (field ColorFieldWidget) layoutWithStyle(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorField, field.key)
	fieldState := core.UseState[colorFieldState](ctx, key, stateSlotColorField)
	defaults := colorFieldDefaultsFor(ctx, &fieldState.defaultsCache)
	fieldState.sync(field.value, field.alpha)

	restore := core.PushKey(ctx, field.key)
	defer restore()
	inputKey := "input"
	invalid := field.invalid || !fieldState.valid
	control := input.New(fieldState.text).Key(inputKey).
		Placeholder("#000000").
		Label(field.resolvedLabel(ctx)).
		Variant(field.variant).
		Disabled(field.disabled).
		Invalid(invalid).
		MaxLength(func() int {
			if field.alpha {
				return 9
			}
			return 7
		}()).
		OnChange(func(value string) {
			fieldState.text = value
			parsed, ok := parseHexColor(value, field.value.A)
			fieldState.valid = ok
			if ok && field.onChange != nil && parsed != field.value {
				field.onChange(parsed)
			}
		})
	group := input.Group(control).
		Variant(field.variant).
		Disabled(field.disabled).
		Invalid(invalid)
	var groupLayout core.Widget = group
	if field.fullWidth {
		group = group.FillWidth()
		groupLayout = group
	}
	if field.swatch {
		group = group.Prefix(ColorSwatch(field.value).Size(ColorSwatchExtraSmall))
		groupLayout = (group).Style(style.Style{}.
			Part(style.PartPrefix, style.Style{}.Padding(0, 8, 0, 10)))
	}

	children := make([]layout.FlexChild, 0, 3)
	if field.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.NewLabel(field.label).
				For(inputKey).
				Required(field.required).
				Disabled(field.disabled).
				Invalid(invalid).
				Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return groupLayout.Layout(ctx, gtx)
	}))
	if invalid && field.errorMessage != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(field.errorMessage).
				Size(float32(defaults.ErrorTextSize)).
				Color(defaults.ErrorForeground).
				Layout(ctx, gtx)
		}))
	} else if field.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.NewDescription(field.description).
				For(inputKey).
				Disabled(field.disabled).
				Layout(ctx, gtx)
		}))
	}

	state := style.State{
		Disabled: field.disabled,
		Invalid:  invalid,
	}
	resolved := computed.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults.Defaults,
		style.Style{},
		style.Style{},
		field.style,
	)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
			Gap:  gtx.Dp(defaults.Gap),
		}.Layout(gtx, children...)
	}))
}

func (state *colorFieldState) sync(value color.NRGBA, alpha bool) {
	if state.ready && state.syncedColor == value && state.syncedAlpha == alpha {
		return
	}
	state.text = formatHexColor(value, alpha)
	state.syncedColor = value
	state.syncedAlpha = alpha
	state.ready = true
	state.valid = true
}

func (field ColorFieldWidget) resolvedLabel(ctx *core.Context) string {
	if field.label != "" {
		return field.label
	}
	return ctx.Text(i18n.ColorPickerColor)
}

// Style merges instance declarations into the widget. The declaration is the
// last cascade layer at layout time, so Style can sit anywhere in the chain.
func (w ColorFieldWidget) Style(values ...style.Style) ColorFieldWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
