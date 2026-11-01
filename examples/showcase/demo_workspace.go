package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/dock"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/splitpane"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/toggle"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/widget/workbench"
)

func demoToolbar(m demoState, send ui.Send[Msg]) ui.Widget {
	commands := []command.Command{
		command.New("save", m.text("保存")).OnExecute(report(send, m.text("工具栏：已保存"))),
		command.New("undo", m.text("撤销")).OnExecute(report(send, m.text("工具栏：已撤销"))),
		command.New("copy", m.text("复制")).OnExecute(report(send, m.text("复制操作"))),
		command.New("export", m.text("导出")).OnExecute(report(send, m.text("已选择导出 JSON"))),
		command.New("publish", m.text("发布项目（未启用）")).Disabled(true),
	}
	return layout.Column(toolbar.New(
		button.New(m.text("保存")).OnClick(report(send, m.text("工具栏：已保存"))),
		button.New(m.text("撤销")).Variant(button.Ghost).OnClick(report(send, m.text("工具栏：已撤销"))),
		toolbar.Separator(),
		toggle.NewToggleButton(m.text("加粗")).Key("tool-bold").Selected(m.bold).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.bold })),
	),
		field(m.text("命令工具栏 · 宽度不足时折叠到更多菜单"),
			layout.Box(toolbar.Commands("toolbar-overflow", commands...)).Style(style.New().MaxWidth(240))),
		field(m.text("垂直工具栏"), toolbar.New(
			button.New(m.text("保存")).OnClick(report(send, m.text("工具栏：已保存"))),
			toolbar.Separator(),
			button.New(m.text("复制")).Variant(button.Ghost).OnClick(report(send, m.text("复制操作"))),
		).Orientation(toolbar.Vertical)),
		field(m.text("整组禁用"), toolbar.Commands("toolbar-disabled", commands[:2]...).Disabled(true)),
	).Gap(20)
}

func demoStatusbar(m demoState, _ ui.Send[Msg]) ui.Widget {
	return statusbar.New(text.New(m.text("● 已同步 · 3 个文件")), text.New("UTF-8 · Go · 100%"))
}

func demoTitlebar(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(titlebar.New("FlowUI Studio", demoMenubar(m, send)).Key("titlebar-demo").
		ShowMinimize(false).ShowMaximize(false).ShowClose(false),
		muted(m.text("自定义标题栏内容")),
		titlebar.New("FlowUI Studio", nil).Key("titlebar-custom").Center(text.New("README.md")).
			Trailing(button.New(m.text("保存")).Size(button.Small).Variant(button.Secondary).
				OnClick(report(send, m.text("已保存")))).
			ShowMinimize(false).ShowMaximize(false).ShowClose(false),
	).Gap(16)
}

func demoPanel(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			button.New(m.text("预览")).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.panel = "preview" }))
			}),
			button.New(m.text("编辑")).Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.panel = "edit" }))
			}),
		).Gap(8),
		surface.NewPanel("panel-demo", m.panel, []surface.PanelItem{
			{Key: "preview", Content: text.New(m.text("当前预览：") + m.name)},
			{Key: "edit", Content: input.New(m.name).Key("panel-name").FillWidth().
				OnChange(setString(send, func(m *demoState) *string { return &m.name }))},
		}).KeepAlive(true),
	).Gap(20)
}

func demoDock(m demoState, _ ui.Send[Msg]) ui.Widget {
	root := dock.Split("dock-main", dock.Horizontal,
		dock.Panel("files", pane(m.text("资源"), "main.go\nview.go\nmodel.go")),
		dock.Split("dock-right", dock.Vertical,
			dock.Panel("editor", pane(m.text("编辑区"), m.text("拖动分隔条调整各区域大小。"))),
			dock.Panel("output", pane(m.text("输出"), "Build completed successfully.")),
		).Ratio(0.7),
	).Ratio(0.28)
	return fixedHeight(340, dock.New("dock-demo", root))
}

func pane(title, content string) ui.Widget {
	return surface.NewCard(layout.Column(text.New(title).Size(16), text.New(content)).Gap(16))
}

func initialWorkbench() workbench.State {
	return workbench.NewState([]workbench.Group{{
		Key: "editor", ActiveKey: "main.go", Tabs: []workbench.Tab{
			{Key: "main.go", Title: "main.go", Closable: true},
			{Key: "README.md", Title: "README.md", Closable: true},
		},
	}})
}

func demoWorkbench(m demoState, send ui.Send[Msg]) ui.Widget {
	controller := workbench.NewController(m.workbench)
	controller.OnEvent(func(event workbench.Event) {
		state := controller.State()
		send(demoMsg(func(m *demoState) { m.workbench = state }))
		send(logMsg(fmt.Sprintf(m.text("工作台：%s"), event.TabKey)))
	})
	items := []tabs.TabItem{}
	panels := []surface.PanelItem{}
	for _, group := range m.workbench.Groups {
		for _, tab := range group.Tabs {
			items = append(items, tabs.TabItem{Key: tab.Key, Label: tab.Title, Closable: true})
			panels = append(panels, surface.PanelItem{Key: tab.Key, Content: pane(tab.Title, m.text("此内容由工作台状态选择。"))})
		}
	}
	tabstrip := tabs.New("", items).Key("workbench-tabs").OnClose(func(key string) {
		controller.CloseTab("editor", key)
	})
	return controller.CommandScope(layout.Column(
		button.New(m.text("重新打开文档")).Variant(button.Secondary).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.workbench = initialWorkbench() }))
		}),
		controller.BindTabs("editor", tabstrip),
		controller.BindPanel("editor", surface.NewPanel("workbench-panel", "", panels).KeepAlive(true)),
	).Gap(16))
}

func demoLayout(m demoState, send ui.Send[Msg]) ui.Widget {
	blocks := []ui.Widget{}
	for _, title := range []string{m.text("项目"), m.text("组件"), m.text("主题"), m.text("窗口"), m.text("状态"), m.text("布局")} {
		blocks = append(blocks, surface.NewCard(text.New(title)))
	}
	return layout.Column(
		muted(m.text("Wrap · 宽度不足时自动换行")), layout.Wrap(blocks...).Gap(8).LineGap(8),
		muted(m.text("Grid · 两列网格")), layout.Grid(2, blocks...).Gap(8),
		muted(m.text("Stack · 叠放内容")),
		layout.Stack(layout.Stacked(pane(m.text("基础层"), m.text("叠放布局"))),
			layout.Overlay(text.New(m.text("浮层标注"))).Offset(180, 12)),
		muted(m.text("SplitPane · 可拖动分割条")),
		fixedHeight(180, splitpane.New(pane(m.text("左侧"), m.text("拖动分隔条")), pane(m.text("右侧"), m.text("比例自动更新"))).
			DefaultRatio(0.4)),
		muted(m.text("纵向受控分割 · 两侧最小高度 60dp")),
		fixedHeight(260, splitpane.New(pane(m.text("预览"), "FlowUI"), pane(m.text("输出"), "Build completed.")).
			Key("split-vertical").Vertical().Ratio(m.splitRatio).MinFirst(60).MinSecond(60).
			OnRatioChange(func(value float32) { send(demoMsg(func(m *demoState) { m.splitRatio = value })) })),
		layout.Wrap(
			text.New(fmt.Sprintf(m.text("当前比例：%.0f%%"), m.splitRatio*100)),
			button.New(m.text("均分面板")).Variant(button.Secondary).
				OnClick(func() { send(demoMsg(func(m *demoState) { m.splitRatio = 0.5 })) }),
		).Gap(12).AlignMiddle(),
		muted(m.text("Scroll · 独立滚动区域")),
		layout.Box(layout.Scroll(layout.Column(blocks...).Gap(8)).Vertical()).
			Style(style.New().Height(160).FillWidth()),
	).Gap(16)
}

func demoBuildingBlocks(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New(m.text("chart → 折线、柱状、饼图等图表的选择和坐标协议")),
		text.New(m.text("disclosure → 折叠面板、弹层与内容面板的展开状态")),
		text.New(m.text("navigation → 侧栏、树与分页的导航协议")),
		text.New(m.text("optionrow → 菜单和选择列表共用的选项行布局")),
		muted(m.text("这些包服务于组件组合；日常使用优先从具体 widget 包开始。")),
	).Gap(16)
}
