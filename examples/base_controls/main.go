package main

import (
	"image"
	"image/color"

	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Sync          bool
	Notifications bool
	Volume        int
}

type Msg interface{ msg() }

type SetSync struct{ Value bool }
type SetNotifications struct{ Value bool }
type SetVolume struct{ Value int }

func (SetSync) msg()          {}
func (SetNotifications) msg() {}
func (SetVolume) msg()        {}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SetSync:
		model.Sync = msg.Value
	case SetNotifications:
		model.Notifications = msg.Value
	case SetVolume:
		model.Volume = msg.Value
	}
	return nil
}

func View(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	active := ctx.Theme()
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Headless controls").Size(22),
				toggleRow(
					"sync",
					"Sync files",
					"Keep this device in step",
					model.Sync,
					active.Colors.Tones[theme.TonePrimary].Fill,
					active.Colors.SurfaceSecondary,
					active.Colors.Border,
					active.Colors.MutedForeground,
					func(next bool) { send(SetSync{next}) },
				),
				toggleRow(
					"notifications",
					"Notifications",
					"Show updates in the corner",
					model.Notifications,
					active.Colors.Tones[theme.TonePrimary].Fill,
					active.Colors.SurfaceSecondary,
					active.Colors.Border,
					active.Colors.MutedForeground,
					func(next bool) { send(SetNotifications{next}) },
				),
				volumeRow(
					model.Volume,
					active.Colors.Tones[theme.TonePrimary].Fill,
					active.Colors.SurfaceTertiary,
					func(next int) { send(SetVolume{next}) },
				),
			).Gap(12),
		).Style(style.New().Padding(24)),
	)
}

func toggleRow(key, label, hint string, checked bool, accent, surface, border, muted color.NRGBA, onChange func(bool)) ui.Widget {
	return base.Toggle(key, checked).
		Label(label).
		Description(hint).
		OnChange(onChange).
		Content(func(state base.ToggleState) ui.Widget {
			mark := layout.Box(layout.Spacer(1, 1)).Style(
				style.New().Width(22).Height(22).Radius(7).
					Background(style.Solid(surface)).
					BorderWidth(1.5).BorderColor(style.Solid(border)),
			)
			if state.Checked {
				mark = layout.Box(layout.Spacer(1, 1)).Style(
					style.New().Width(22).Height(22).Radius(7).
						Background(style.Solid(accent)),
				)
			}
			return layout.Box(
				layout.Row(
					mark,
					layout.Column(
						text.New(label),
						text.New(hint).Size(12).Color(muted),
					).Gap(2),
				).Gap(12).AlignMiddle(),
			).Style(
				style.New().FillWidth().Padding(14).Radius(12).Background(style.Solid(surface)),
			)
		})
}

func volumeRow(value int, accent, track color.NRGBA, onChange func(int)) ui.Widget {
	return base.Slider("volume", float64(value), 0, 100).
		Step(1).
		Label("Volume").
		OnChange(func(next float64) { onChange(int(next)) }).
		Track(func(state base.SliderState) ui.Widget {
			return layout.Box(ui.WidgetFunc(func(_ *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				width := gtx.Constraints.Min.X
				if width <= 0 {
					width = min(gtx.Constraints.Max.X, gtx.Dp(360))
				}
				height := gtx.Constraints.Min.Y
				if height <= 0 {
					height = gtx.Dp(56)
				}
				if width <= 0 || height <= 0 {
					return giolayout.Dimensions{Size: image.Pt(width, height)}
				}
				thickness := gtx.Dp(6)
				y := height/2 - thickness/2
				rr := thickness / 2
				paint.FillShape(gtx.Ops, track, clip.RRect{
					Rect: image.Rect(0, y, width, y+thickness),
					NW:   rr, NE: rr, SW: rr, SE: rr,
				}.Op(gtx.Ops))
				fill := int(float32(width) * state.Ratio)
				if fill > 0 {
					paint.FillShape(gtx.Ops, accent, clip.RRect{
						Rect: image.Rect(0, y, fill, y+thickness),
						NW:   rr, NE: rr, SW: rr, SE: rr,
					}.Op(gtx.Ops))
				}
				diameter := gtx.Dp(18)
				cx := int(float32(width) * state.Ratio)
				paint.FillShape(gtx.Ops, accent, clip.Ellipse(image.Rect(
					cx-diameter/2, height/2-diameter/2,
					cx+diameter/2, height/2+diameter/2,
				)).Op(gtx.Ops))
				return giolayout.Dimensions{Size: image.Pt(width, height)}
			})).Style(style.New().FillWidth().Height(56))
		})
}

func main() {
	ui.Run(
		ui.NewProgram(Model{Volume: 40}, Update, View),
		window.Theme(theme.Light()),
		window.Title("Headless controls"),
		window.Size(520, 380),
	)
}
