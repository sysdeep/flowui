package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"sync"

	"gioui.org/font"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/badge"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/image"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
)

func demoText(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New(m.text("用 Go 构建桌面应用")).Size(28).Weight(font.Bold),
		text.New(m.text("正文 · 默认字号与主题前景色")),
		muted(m.text("辅助信息 · 随明暗主题调整对比度")),
		text.Selectable(m.text("这段文字可以用鼠标选择并复制。")).Key("selectable-text"),
		field(m.text("单行省略 · 调整窗口宽度查看截断"),
			text.New(m.text("这是一段较长的项目说明，用于展示有限宽度下的单行省略效果，完整内容可以在详情页继续阅读。")).
				MaxLines(1).Truncator("…")),
		text.New(m.text("斜体说明与自定义行距\n第二行保持舒适的阅读间距。")).FontStyle(font.Italic).LineHeight(26),
		divider.New(),
		text.NewLabel(m.text("项目名称")).For("label-project"),
		input.New(m.name).Key("label-project").FillWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.name })),
		text.NewDescription(m.text("使用易于识别的名称，点击标签可聚焦输入框。")).For("label-project"),
	).Gap(14)
}

func demoIcon(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			icon.New(lucide.Search).Size(16), icon.New(lucide.Plus).Size(20),
			icon.New(lucide.Settings).Size(24), icon.New(lucide.Heart).Size(32),
		).Gap(24).AlignMiddle(),
		layout.Wrap(
			button.New(m.text("添加")).Content(icon.New(lucide.Plus)).IconOnly().OnClick(report(send, m.text("添加"))),
			button.New(m.text("保存")).Content(layout.Row(icon.New(lucide.Save).Size(16), text.New(m.text("保存"))).Gap(8)).
				OnClick(report(send, m.text("已保存"))),
		).Gap(12),
	).Gap(24)
}

var desertImage = sync.OnceValue(func() paint.ImageOp {
	img, err := jpeg.Decode(bytes.NewReader(images.BGDesertJPG))
	if err != nil {
		panic(fmt.Errorf("decode bundled showcase image: %w", err))
	}
	return paint.NewImageOp(img)
})

func demoImage(m demoState, _ ui.Send[Msg]) ui.Widget {
	imageOp := desertImage()
	return layout.Column(
		muted(m.text("Cover · 填充并裁剪")),
		image.New(imageOp).Width(640).Height(210).Fit(image.FitCover).Radius(8).Alt(m.text("沙漠沙丘")),
		muted(m.text("Contain · 保留完整图片")),
		image.New(imageOp).Width(360).Height(200).Fit(image.FitContain).Radius(8).Alt(m.text("完整沙丘图片")),
		muted(m.text("顶部裁剪与半透明图片")),
		layout.Wrap(
			image.New(imageOp).Width(240).Height(130).Fit(image.FitCover).Position(layout.AlignTopStart).
				Radius(8).Alt(m.text("顶部裁剪沙丘")),
			image.New(imageOp).Width(240).Height(130).Fit(image.FitCover).Opacity(0.5).
				Radius(8).Alt(m.text("半透明沙丘")),
		).Gap(12).LineGap(12),
	).Gap(12)
}

func demoAvatar(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		field(m.text("尺寸对比"), layout.Wrap(
			avatar.New("FL").Size(avatar.SizeSmall), avatar.New("UI").Size(avatar.SizeMedium),
			avatar.New("GO").Size(avatar.SizeLarge),
		).Gap(20).AlignMiddle()),
		field(m.text("语义颜色"), layout.Wrap(
			avatar.New("FL"), avatar.New("UI").Color(avatar.ColorAccent),
			avatar.New("GO").Color(avatar.ColorSuccess), avatar.New("WA").Color(avatar.ColorWarning),
			avatar.New("ER").Color(avatar.ColorDanger),
		).Gap(16).LineGap(12)),
		field(m.text("柔和外观"), layout.Wrap(
			avatar.New("FL").Variant(avatar.VariantSoft),
			avatar.New("UI").Color(avatar.ColorAccent).Variant(avatar.VariantSoft),
			avatar.New("GO").Color(avatar.ColorSuccess).Variant(avatar.VariantSoft),
		).Gap(16)),
		field(m.text("图片与自定义回退"), layout.Wrap(
			avatar.New("FL").Image(desertImage()).Alt(m.text("沙漠沙丘")).Size(avatar.SizeLarge),
			avatar.New("").Fallback(icon.New(lucide.User).Size(20)).Alt(m.text("用户头像")),
		).Gap(16)),
	).Gap(20)
}

func demoBadge(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		field(m.text("计数与状态"), layout.Wrap(
			badge.New(button.New(m.text("通知")).Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.clicks++ }))
			}), fmt.Sprint(m.clicks+3)),
			badge.New(avatar.New("FL"), "8"),
			badge.New(avatar.New("UI"), "").Color(badge.ColorSuccess).Alt(m.text("在线")),
		).Gap(32).AlignMiddle()),
		field(m.text("外观变体"), layout.Wrap(
			badge.New(avatar.New("FL"), "8").Variant(badge.VariantPrimary),
			badge.New(avatar.New("UI"), "8").Variant(badge.VariantSecondary),
			badge.New(avatar.New("GO"), "8").Variant(badge.VariantSoft),
		).Gap(32)),
		field(m.text("语义颜色"), layout.Wrap(
			badge.New(avatar.New("FL"), "1").Color(badge.ColorAccent),
			badge.New(avatar.New("UI"), "2").Color(badge.ColorSuccess),
			badge.New(avatar.New("GO"), "3").Color(badge.ColorWarning),
			badge.New(avatar.New("ER"), "4").Color(badge.ColorDanger),
		).Gap(32).LineGap(16)),
		field(m.text("四角定位"), layout.Wrap(
			badge.New(avatar.New("TL"), "1").Placement(badge.PlacementTopLeft),
			badge.New(avatar.New("TR"), "2").Placement(badge.PlacementTopRight),
			badge.New(avatar.New("BL"), "3").Placement(badge.PlacementBottomLeft),
			badge.New(avatar.New("BR"), "4").Placement(badge.PlacementBottomRight),
		).Gap(32).LineGap(16)),
		field(m.text("尺寸对比"), layout.Wrap(
			badge.New(avatar.New("S"), "8").Size(badge.SizeSmall),
			badge.New(avatar.New("M"), "8").Size(badge.SizeMedium),
			badge.New(avatar.New("L"), "8").Size(badge.SizeLarge),
		).Gap(32)),
	).Gap(24)
}

func demoChip(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		field(m.text("语义颜色"), layout.Wrap(
			chip.New(m.text("默认")), chip.New(m.text("强调")).Color(chip.ColorAccent),
			chip.New(m.text("已发布")).Color(chip.ColorSuccess), chip.New(m.text("警告")).Color(chip.ColorWarning),
			chip.New(m.text("错误")).Color(chip.ColorDanger),
		).Gap(10).LineGap(10)),
		field(m.text("外观变体"), layout.Wrap(
			chip.New("Primary").Color(chip.ColorAccent).Variant(chip.VariantPrimary),
			chip.New("Secondary").Color(chip.ColorAccent).Variant(chip.VariantSecondary),
			chip.New("Tertiary").Color(chip.ColorAccent).Variant(chip.VariantTertiary),
			chip.New("Soft").Color(chip.ColorAccent).Variant(chip.VariantSoft),
		).Gap(10).LineGap(10)),
		field(m.text("尺寸对比"), layout.Wrap(
			chip.New("Small").Size(chip.SizeSmall), chip.New("Medium"), chip.New("Large").Size(chip.SizeLarge),
		).Gap(10).AlignMiddle()),
		field(m.text("前后图标"), layout.Wrap(
			chip.New(m.text("已发布")).Color(chip.ColorSuccess).StartContent(icon.New(lucide.Check).Size(14)),
			chip.New(m.text("桌面应用")).StartContent(icon.New(lucide.Monitor).Size(14)).EndContent(text.New("Go")),
		).Gap(10).LineGap(10)),
	).Gap(20)
}

func demoCard(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(surface.NewCard(layout.Column(
		layout.Row(avatar.New("FL"), layout.Column(
			text.New("FlowUI").Size(20).Weight(font.Bold), muted(m.text("原生桌面组件库")),
		).Gap(4)).Gap(12).AlignMiddle(),
		text.New(m.text("一套默认风格，组合组件来构建你的应用。")),
		button.New(m.text("打开项目")).OnClick(report(send, m.text("已打开 FlowUI 项目"))),
	).Gap(20)),
		field(m.text("卡片表面变体"), layout.Column(
			surface.NewCard(text.New(m.text("次级卡片"))).Variant(surface.CardSecondary),
			surface.NewCard(text.New(m.text("三级卡片"))).Variant(surface.CardTertiary),
			surface.NewCard(text.New(m.text("透明卡片"))).Variant(surface.CardTransparent),
		).Gap(12)),
	).Gap(20)
}

func demoSurface(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		surface.New(text.New(m.text("Default · 默认表面"))),
		surface.New(text.New(m.text("Secondary · 次级表面"))).Variant(surface.Secondary),
		surface.New(text.New(m.text("Tertiary · 三级表面"))).Variant(surface.Tertiary),
		surface.New(text.New(m.text("Transparent · 透明表面"))).Variant(surface.Transparent),
	).Gap(12)
}

func demoCode(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New(m.text("语法高亮 · 可拖拽选中")).Size(13).Weight(font.Bold),
		code.New("go", `// Save writes the model to disk.
func (m *Model) Save() error {
	if !m.Dirty {
		return nil
	}
	return os.WriteFile(m.Path, m.Bytes(), 0o644)
}`),
		muted(m.text("右上角按钮复制全部代码；也可拖拽选中后按 Ctrl/Cmd+C 复制。长行按容器宽度换行。")),
		field(m.text("JSON · 隐藏行号"), code.New("json", `{
  "name": "FlowUI",
  "enabled": true,
  "components": ["Button", "Input", "Table"]
}`).LineNumbers(false)),
		field(m.text("纯文本"), code.New("text", "go run ./examples/showcase").Key("demo-plain").LineNumbers(false)),
	).Gap(12)
}
