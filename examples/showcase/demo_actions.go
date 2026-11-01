package main

import (
	"fmt"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toggle"
)

func demoButton(m demoState, send ui.Send[Msg]) ui.Widget {
	variants := []struct {
		label string
		value button.Variant
	}{
		{label: m.text("主要操作"), value: button.Primary},
		{label: m.text("次要操作"), value: button.Secondary},
		{label: m.text("低强调"), value: button.Tertiary},
		{label: m.text("描边"), value: button.Outline},
		{label: m.text("轻量"), value: button.Ghost},
		{label: m.text("危险操作"), value: button.Danger},
		{label: m.text("危险柔和"), value: button.DangerSoft},
		{label: m.text("危险描边"), value: button.DangerOutline},
	}
	buttons := []ui.Widget{}
	for index, variant := range variants {
		buttons = append(buttons, button.New(variant.label).Key(fmt.Sprintf("variant-%d", index)).
			Variant(variant.value).OnClick(report(send, fmt.Sprintf(m.text("点击了“%s”"), variant.label))))
	}
	return layout.Column(
		muted(m.text("语义变体")), layout.Wrap(buttons...).Gap(10).LineGap(10),
		muted(m.text("尺寸与状态")),
		layout.Wrap(
			button.New(m.text("小按钮")).Size(button.Small).OnClick(report(send, m.text("小按钮"))),
			button.New(m.text("默认按钮")).OnClick(report(send, m.text("默认按钮"))),
			button.New(m.text("大按钮")).Size(button.Large).OnClick(report(send, m.text("大按钮"))),
			button.New(m.text("保存中")).Loading(m.buttonLoading),
			button.New(m.text("不可用")).Disabled(true),
		).Gap(10).LineGap(10).AlignMiddle(),
		toggle.NewSwitch(m.text("演示按钮加载动画")).Key("button-loading").Checked(m.buttonLoading).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.buttonLoading })),
		muted(m.text("按钮组")),
		button.Group(
			button.New(m.text("左对齐")).OnClick(report(send, m.text("左对齐"))),
			button.New(m.text("居中")).OnClick(report(send, m.text("居中"))),
			button.New(m.text("右对齐")).OnClick(report(send, m.text("右对齐"))),
		),
		field(m.text("等宽描边按钮组"), button.Group(
			button.New(m.text("左对齐")).OnClick(report(send, m.text("左对齐"))),
			button.New(m.text("居中")).OnClick(report(send, m.text("居中"))),
			button.New(m.text("右对齐")).OnClick(report(send, m.text("右对齐"))),
		).Variant(button.Outline).Size(button.Small).FillWidth().Separators(true)),
		field(m.text("垂直按钮组"), button.Group(
			button.New(m.text("保存")).OnClick(report(send, m.text("已保存"))),
			button.New(m.text("取消")).OnClick(report(send, m.text("取消"))),
		).Orientation(button.GroupVertical).Variant(button.Secondary).Separators(true)),
		divider.New(),
		layout.Wrap(
			button.New(m.text("点击计数")).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.clicks++ }))
			}),
			text.New(fmt.Sprintf(m.text("已点击 %d 次"), m.clicks)),
		).Gap(12).AlignMiddle(),
	).Gap(16)
}

func demoToggle(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := setBool(send, func(m *demoState) *bool { return &m.bold })
	return layout.Column(
		field(m.text("外观变体"), layout.Wrap(
			toggle.NewToggleButton(m.text("加粗")).Key("toggle-bold").Selected(m.bold).OnChange(changed),
			toggle.NewToggleButton(m.text("轻量")).Key("toggle-ghost").Variant(toggle.ToggleButtonGhost).
				Selected(m.bold).OnChange(changed),
			toggle.NewToggleButton(m.text("加粗")).Key("toggle-icon").Content(icon.New(lucide.Bold).Size(16)).
				IconOnly().Selected(m.bold).OnChange(changed),
		).Gap(12).LineGap(12).AlignMiddle()),
		field(m.text("尺寸对比"), layout.Wrap(
			toggle.NewToggleButton("Small").Key("toggle-small").Size(toggle.ToggleButtonSmall).Selected(m.bold).OnChange(changed),
			toggle.NewToggleButton("Medium").Key("toggle-medium").Selected(m.bold).OnChange(changed),
			toggle.NewToggleButton("Large").Key("toggle-large").Size(toggle.ToggleButtonLarge).Selected(m.bold).OnChange(changed),
		).Gap(12).LineGap(12).AlignMiddle()),
		field(m.text("禁用状态"), layout.Wrap(
			toggle.NewToggleButton(m.text("已锁定")).Key("toggle-disabled").Selected(true).Disabled(true),
			toggle.NewToggleButton(m.text("不可用")).Key("toggle-disabled-off").Disabled(true),
		).Gap(12)),
		text.New(fmt.Sprintf(m.text("加粗：%t"), m.bold)),
	).Gap(20)
}

func demoClose(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(layout.Row(text.New(m.text("关闭操作会在底部显示反馈")),
		button.NewCloseButton().Key("close-demo").OnClick(report(send, m.text("已点击关闭按钮"))),
	).Gap(16).AlignMiddle(),
		field(m.text("禁用与自定义图标"), layout.Wrap(
			button.NewCloseButton().Key("close-disabled").Disabled(true).Label(m.text("禁用关闭")),
			button.NewCloseButton().Key("close-custom").Icon(icon.New(lucide.Minus).Size(16)).Label(m.text("收起")).
				OnClick(report(send, m.text("已点击关闭按钮"))),
		).Gap(16)),
	).Gap(20)
}

func demoCommand(m demoState, send ui.Send[Msg]) ui.Widget {
	save := command.New("demo-save", m.text("保存文档")).
		Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).
		OnExecute(report(send, m.text("保存命令已执行（按钮与 Ctrl+S 共享操作）")))
	disabled := command.New("demo-command-disabled", m.text("发布项目（未启用）")).Disabled(true)
	return command.Scope([]command.Command{save, disabled}, layout.Column(
		command.Button("command-save", save),
		command.Button("command-disabled", disabled),
		muted(m.text("点击按钮后，按 Ctrl+S / macOS ⌘S 再执行一次。")),
	).Gap(12))
}
