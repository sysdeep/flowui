package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/commandpalette"
	widgetnav "github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/numericinput"
	"github.com/qianniancn/flowui/widget/skeleton"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toggle"
)

func demoNumericInput(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted(m.text("金额 · -100 至 100，每步 0.25，保留两位小数")),
		numericinput.New(m.amount).Key("amount").Range(-100, 100).Step(0.25).Precision(2).Label(m.text("金额")).
			OnChange(func(value float64) { send(demoMsg(func(m *demoState) { m.amount = value })) }),
		text.New(fmt.Sprintf(m.text("已提交：%.2f"), m.amount)),
		muted(m.text("输入时保留草稿；回车或离开输入框时提交，加减按钮立即生效。")),
		field(m.text("整数数量 · 0 至 10，达到边界后停止增减"),
			numericinput.New(m.quantity).Key("quantity").Range(0, 10).Step(1).Precision(0).Label(m.text("数量")).
				OnChange(func(value float64) { send(demoMsg(func(m *demoState) { m.quantity = value })) })),
		layout.Wrap(
			button.New(m.text("设为最小值")).Variant(button.Secondary).
				OnClick(func() { send(demoMsg(func(m *demoState) { m.quantity = 0 })) }),
			button.New(m.text("设为最大值")).Variant(button.Secondary).
				OnClick(func() { send(demoMsg(func(m *demoState) { m.quantity = 10 })) }),
		).Gap(8),
		numericinput.New(42).Key("amount-disabled").Label(m.text("禁用金额")).Disabled(true),
	).Gap(12)
}

func demoSegmented(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []toggle.SegmentedItem{
		{Key: "day", Label: m.text("日")}, {Key: "week", Label: m.text("周")}, {Key: "month", Label: m.text("月")},
		{Key: "year", Label: m.text("年"), Disabled: true},
	}
	changed := setString(send, func(m *demoState) *string { return &m.segment })
	return layout.Column(
		toggle.NewSegmented(items).Key("period").SelectedKey(m.segment).OnChange(changed),
		muted(m.text("占满容器宽度")),
		toggle.NewSegmented(items).Key("period-full").SelectedKey(m.segment).FillWidth().OnChange(changed),
		text.New(m.text("当前周期：")+m.segment),
		field(m.text("整组禁用"), toggle.NewSegmented(items).Key("period-disabled").SelectedKey("week").Disabled(true)),
	).Gap(16)
}

func demoBreadcrumb(m demoState, send ui.Send[Msg]) ui.Widget {
	path := []widgetnav.BreadcrumbItem{
		{Key: "home", Label: m.text("首页")}, {Key: "docs", Label: m.text("文档")}, {Key: "components", Label: m.text("组件")},
	}
	for index, item := range path {
		if item.Key == m.location {
			path = path[:index+1]
			break
		}
	}
	return layout.Column(
		widgetnav.NewBreadcrumb(path).Key("location").OnChange(setString(send, func(m *demoState) *string { return &m.location })),
		text.New(m.text("当前位置：")+m.location),
		button.New(m.text("回到组件页")).Variant(button.Secondary).
			OnClick(func() { send(demoMsg(func(m *demoState) { m.location = "components" })) }),
		muted(m.text("自定义分隔符与禁用层级")),
		widgetnav.NewBreadcrumb([]widgetnav.BreadcrumbItem{
			{Key: "home", Label: m.text("首页")},
			{Key: "archive", Label: m.text("归档"), Disabled: true},
			{Key: "components", Label: m.text("组件")},
		}).Key("location-custom").Separator("›").OnChange(func(key string) { send(logMsg(m.text("面包屑导航：") + key)) }),
		muted(m.text("较长路径 · 窄窗口自动换行")),
		widgetnav.NewBreadcrumb([]widgetnav.BreadcrumbItem{
			{Key: "workspace", Label: m.text("我的工作区")},
			{Key: "projects", Label: m.text("项目")},
			{Key: "flowui", Label: "FlowUI"},
			{Key: "docs", Label: m.text("文档")},
			{Key: "guide", Label: m.text("使用指南")},
			{Key: "navigation", Label: m.text("数据与导航")},
			{Key: "breadcrumb", Label: m.text("面包屑")},
		}).Key("location-long").OnChange(func(key string) { send(logMsg(m.text("面包屑导航：") + key)) }),
	).Gap(16)
}

func demoSkeleton(m demoState, send ui.Send[Msg]) ui.Widget {
	var content ui.Widget = layout.Column(text.New(m.text("FlowUI 项目")), muted(m.text("内容已经加载，可以开始使用。"))).Gap(8)
	if m.loading {
		content = layout.Row(
			skeleton.New().Circle().Height(48),
			layout.Expanded(layout.Column(skeleton.New().Width(160), skeleton.New(), skeleton.New().Width(220)).Gap(10)),
		).Gap(16).AlignMiddle()
	}
	return layout.Column(
		toggle.NewSwitch(m.text("模拟加载中")).Key("loading").Checked(m.loading).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.loading })),
		content,
		muted(m.text("静态占位（关闭动效时，所有骨架屏都会停止刷新）")),
		skeleton.New().Height(80).Animated(false),
	).Gap(20)
}

func demoCommandPalette(m demoState, send ui.Send[Msg]) ui.Widget {
	commands := []command.Command{
		command.New("save", m.text("保存项目")).Description(m.text("保存当前演示内容")).
			Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).OnExecute(report(send, m.text("已保存演示项目"))),
		command.New("theme", m.text("切换明暗主题")).Description(m.text("切换默认主题")).OnExecute(func() { send(themeMsg{}) }),
		command.New("reset", m.text("重置演示")).Description(m.text("恢复所有示例的初始数据")).OnExecute(func() { send(resetMsg{}) }),
		command.New("disabled", m.text("发布项目（未启用）")).Disabled(true),
	}
	return command.Scope(commands, layout.Column(
		button.New(m.text("打开命令面板")).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.palette = true }))
		}),
		muted(m.text("搜索名称或说明；↑ / ↓ 选择，Enter 执行，Esc 关闭。Ctrl+S 与菜单共用保存命令。")),
		commandpalette.New(m.palette, commands).Key("palette").
			OnOpenChange(setBool(send, func(m *demoState) *bool { return &m.palette })),
	).Gap(16))
}
