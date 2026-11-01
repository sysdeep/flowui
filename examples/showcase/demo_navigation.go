package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/collapsible"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/listbox"
	"github.com/qianniancn/flowui/widget/menu"
	widgetnav "github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/timeline"
	"github.com/qianniancn/flowui/widget/toggle"
	"github.com/qianniancn/flowui/widget/tree"
)

func demoList(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted(m.text("单选")),
		listbox.New(languageItems(m)).Key("list-single").SelectedKey(m.combo).FillWidth().MaxHeight(220).
			OnChange(setString(send, func(m *demoState) *string { return &m.combo })),
		muted(m.text("多选")),
		listbox.New(languageItems(m)).Key("list-multiple").SelectedKeys(m.keys).FillWidth().MaxHeight(220).
			OnSelectionChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.keys = keys })) }),
		text.New(m.text("已选：")+strings.Join(m.keys, ", ")),
		muted(m.text("分组与描述 · 含图标、尾部信息和禁用项")),
		listbox.Sections(m.listGrouped, []listbox.Section{
			{Title: m.text("项目"), Items: []listbox.Item{
				{Key: "flowui", Label: "FlowUI", Description: m.text("组件库"),
					Leading: icon.New(lucide.Folder).Size(16), Trailing: muted("12")},
				{Key: "studio", Label: "Studio", Description: m.text("计划中"), Disabled: true},
			}},
			{Title: m.text("资源"), Items: []listbox.Item{
				{Key: "docs", Label: m.text("文档"), Leading: icon.New(lucide.FileText).Size(16)},
			}},
		}).Key("list-grouped").FillWidth().OnChange(setString(send, func(m *demoState) *string { return &m.listGrouped })),
		muted(m.text("允许取消选择 · 再次点击当前项可清空")),
		listbox.New(languageItems(m)).Key("list-optional").SelectedKey(m.listOptional).
			AllowEmptySelection().HideIndicator().FillWidth().MaxHeight(180).
			OnChange(setString(send, func(m *demoState) *string { return &m.listOptional })),
		text.New(m.text("当前选择：")+m.listOptional),
		muted(m.text("仅执行操作 · 不保留选中状态")),
		listbox.New([]listbox.Item{
			{Key: "open", Label: m.text("打开"), Leading: icon.New(lucide.Folder).Size(16)},
			{Key: "delete", Label: m.text("删除"), Variant: listbox.ItemDanger},
		}).Key("list-actions").SelectionMode(listbox.SelectionNone).FillWidth().
			OnAction(func(key string) { send(logMsg(m.text("列表操作：") + key)) }),
		muted(m.text("空列表与整组禁用")),
		listbox.New([]listbox.Item{}).Key("list-empty").FillWidth().EmptyText(m.text("暂无可选语言")),
		listbox.New(languageItems(m)).Key("list-disabled").SelectedKey("go").FillWidth().Disabled(true),
	).Gap(12)
}

func demoTable(m demoState, send ui.Send[Msg]) ui.Widget {
	rows := tableRows(m, send, "main")
	gridRows := tableRows(m, nil, "")
	loadRows := tableRows(m, nil, "")
	footerRows := tableRows(m, nil, "")
	columns := []table.Column{
		{Key: "name", Label: m.text("项目"), Width: m.tableNameWidth, MinWidth: 130, MaxWidth: 300, Resizable: true, Sortable: true, RowHeader: true},
		{Key: "type", Label: m.text("类型"), Width: m.tableTypeWidth, MinWidth: 90, MaxWidth: 220, Resizable: true, Sortable: true, Align: table.AlignCenter},
		{Key: "status", Label: m.text("状态"), Weight: 1, Sortable: true, Align: table.AlignEnd},
	}
	setSort := func(value table.SortDescriptor) {
		send(demoMsg(func(m *demoState) { m.tableSort = value }))
		send(logMsg(m.text("表格排序：") + value.Column + sortDirectionLabel(m, value.Direction)))
	}
	setWidth := func(key string, width int) {
		send(demoMsg(func(m *demoState) {
			if key == "name" {
				m.tableNameWidth = width
			} else if key == "type" {
				m.tableTypeWidth = width
			}
		}))
	}
	selectSingle := table.New(columns, rows).Key("project-table").
		Variant(table.VariantSecondary).Border(true).MinWidth(520).RowHeight(36).HeaderHeight(36).
		SelectionMode(table.SelectionSingle).AllowEmptySelection().SelectedKey(m.selected).
		DisabledKeys([]string{"studio"}).OnChange(selection(m, send)).OnSortChange(setSort).SortDescriptor(m.tableSort).
		OnColumnResize(setWidth).OnRowActivate(func(key string) { send(logMsg(m.text("双击激活：") + key)) }).
		OnAction(func(key string) { send(demoMsg(func(m *demoState) { m.tableRowSelected = key })) })
	selectSingle = selectSingle.RowContextMenu(func(row table.Row) menu.Widget {
		return menu.New([]menu.Item{
			{Key: "open", Label: m.text("打开 ") + row.Key, OnAction: report(send, m.text("右键菜单打开：")+row.Key)},
			{Key: "copy", Label: m.text("复制项目键"), OnAction: report(send, m.text("右键菜单复制：")+row.Key)},
		}).Key("table-row-actions")
	})
	if m.tableEmpty {
		selectSingle = selectSingle.EmptyText(m.text("没有项目")).EmptyContent(text.New(m.text("空状态：可以通过下方按钮恢复示例数据。")))
	}

	grid := table.New(columns, gridRows).Key("project-table-grid").
		Border(true).GridLines(true).Striped(true).MaxHeight(230).RowHeight(36).
		SelectionMode(table.SelectionMultiple).ShowSelectionIndicator().
		SelectedKeys(m.tableSelected).DisabledKeys([]string{"studio"}).
		OnSelectionChange(func(keys []string) {
			send(demoMsg(func(m *demoState) { m.tableSelected = keys }))
			send(logMsg(m.text("多选项目：") + strings.Join(keys, "、")))
		}).OnAction(func(key string) { send(logMsg(m.text("操作行：") + key)) })

	loadMore := table.New(columns, loadRows).Key("project-table-more").
		Border(true).Variant(table.VariantSecondary).MaxHeight(190).RowHeight(36).
		LoadMore(m.tableLoadCount < 2, false, func() {
			send(demoMsg(func(m *demoState) { m.tableLoadCount++ }))
			send(logMsg(m.text("已加载更多项目")))
		}).LoadMoreContent(text.New(m.text("正在加载项目…")))

	return layout.Column(
		muted(m.text("基础表格 · 单选、空选、禁用行、排序、双击、右键菜单和可调整列宽")),
		selectSingle,
		text.New(m.text("当前行：")+m.tableRowSelected+m.text(" · 排序：")+sortLabel(m, m.tableSort)+m.text(" · 拖动列标题右侧边界可调整宽度")),
		muted(m.text("多选与表格线 · 点击表头勾选全部，Shift+点击可连续选择")),
		grid,
		text.New(m.text("已选项目：")+strings.Join(m.tableSelected, "、")),
		muted(m.text("自定义单元格：状态列包含可交互按钮；点击按钮不会触发行选择。")),
		muted(m.text("加载更多 · 滚动到表格底部触发 LoadMore（最多追加两批）")),
		loadMore,
		layout.Wrap(
			button.New(m.text("切换空状态")).Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.tableEmpty = !m.tableEmpty }))
			}),
			button.New(m.text("恢复加载示例")).Variant(button.Ghost).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.tableEmpty, m.tableLoadCount = false, 0 }))
			}),
		).Gap(8),
		text.New(m.text("Footer 与布局参数：")),
		table.New(columns, footerRows[:min(len(footerRows), 3)]).Key("project-table-footer").Variant(table.VariantSecondary).
			Border(true).GridLines(true).Flow().MinWidth(520).HeaderHeight(34).RowHeight(36).
			Footer(text.New(fmt.Sprintf(m.text("共 %d 条 · VariantSecondary · Flow · MinWidth"), len(rows)))),
		muted(m.text("行高对比 · 28dp 紧凑与 44dp 宽松；默认行高为 36dp")),
		table.New(columns, gridRows).Key("table-compact").RowHeight(28).HeaderHeight(28).Border(true),
		table.New(columns, gridRows).Key("table-comfortable").RowHeight(44).Border(true),
		muted(m.text("整组禁用")),
		table.New(columns, gridRows).Key("table-disabled").Border(true).
			SelectionMode(table.SelectionSingle).SelectedKey("flowui").Disabled(true),
	).Gap(16)
}

func tableRows(m demoState, send ui.Send[Msg], keyPrefix string) []table.Row {
	rows := []table.Row{
		{Key: "flowui", Cells: []table.Cell{{Text: "FlowUI"}, {Text: m.text("组件库")}, {Text: m.text("进行中")}}},
		{Key: "studio", Disabled: true, Cells: []table.Cell{{Text: "Studio"}, {Text: m.text("编辑器")}, {Text: m.text("计划中")}}},
		{Key: "docs", Cells: []table.Cell{{Text: "Documentation"}, {Text: m.text("文档")}, {Text: m.text("已发布")}}},
	}
	if keyPrefix != "" {
		statusButton := button.New(m.text("已发布")).Key(keyPrefix + "-docs-status").Variant(button.Ghost)
		if send != nil {
			statusButton = statusButton.OnClick(func() { send(logMsg(m.text("点击了状态按钮：已发布"))) })
		}
		rows[2].Cells[2] = table.Cell{Text: m.text("已发布"), Content: statusButton, Interactive: true}
	}
	for index := 0; index < m.tableLoadCount*3; index++ {
		key := fmt.Sprintf("sample-%02d", index+1)
		rows = append(rows, table.Row{Key: key, Cells: []table.Cell{{Text: "Sample " + fmt.Sprint(index+1)}, {Text: m.text("示例")}, {Text: m.text("已同步")}}})
	}
	if m.tableEmpty {
		return nil
	}
	if m.tableSort.Column != "" {
		columnIndex := map[string]int{"name": 0, "type": 1, "status": 2}[m.tableSort.Column]
		sort.SliceStable(rows, func(i, j int) bool {
			left, right := rows[i].Cells[columnIndex].Text, rows[j].Cells[columnIndex].Text
			if m.tableSort.Direction == table.SortDescending {
				return left > right
			}
			return left < right
		})
	}
	return rows
}

func sortDirectionLabel(m demoState, direction table.SortDirection) string {
	if direction == table.SortDescending {
		return m.text("（降序）")
	}
	return m.text("（升序）")
}

func sortLabel(m demoState, sort table.SortDescriptor) string {
	if sort.Column == "" {
		return m.text("未设置")
	}
	return sort.Column + sortDirectionLabel(m, sort.Direction)
}

func demoTree(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []tree.Item{
		{Key: "src", Label: "src", Children: []tree.Item{
			{Key: "main.go", Label: "main.go"},
			{Key: "widgets", Label: "widgets", Children: []tree.Item{
				{Key: "button.go", Label: "button.go"}, {Key: "input.go", Label: "input.go"},
			}},
		}},
		{Key: "readme", Label: "README.md"},
		{Key: "archive", Label: "archive", Disabled: true},
	}
	return layout.Column(
		fixedHeight(270, tree.New(m.selected, items).Key("file-tree").ExpandedKeys(m.expandedKeys).
			OnExpandedChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.expandedKeys = keys })) }).
			OnChange(selection(m, send))),
		text.New(m.text("当前文件：")+m.selected),
		muted(m.text("多选与右键菜单 · Ctrl/Command 切换选择，Shift 连续选择")),
		fixedHeight(240, tree.New("", items).Key("file-tree-multi").
			SelectionMode(tree.SelectionMultiple).SelectedKeys(m.treeMultiSelected).
			ExpandedKeys([]string{"src", "widgets"}).Guides(true).GuideStyle(tree.GuideDashed).
			OnSelectionChange(func(keys []string) {
				send(demoMsg(func(m *demoState) { m.treeMultiSelected = keys }))
			}).ContextMenu(menu.New([]menu.Item{
			{Key: "open", Label: m.text("打开"), OnAction: report(send, m.text("打开选中文件"))},
			{Key: "copy", Label: m.text("复制项目键"), OnAction: report(send, m.text("复制选中文件"))},
		}).Key("tree-file-menu")).OnContextMenu(func(key string) { send(logMsg(m.text("右键文件：") + key)) })),
		text.New(m.text("选中文件：")+strings.Join(m.treeMultiSelected, ", ")),
		muted(m.text("独立勾选 · 父子节点互不联动 · 紧凑表面变体")),
		fixedHeight(220, tree.New("", items).Key("file-tree-strict").
			Variant(tree.VariantSurface).Size(tree.SizeSmall).SelectionMode(tree.SelectionNone).
			ExpandedKeys([]string{"src", "widgets"}).Checkable(true).CheckStrictly(true).
			CheckedKeys(m.treeStrictChecked).Guides(true).GuideConnectors(true).
			OnCheckedChange(func(checked, _ []string) {
				send(demoMsg(func(m *demoState) { m.treeStrictChecked = checked }))
			})),
		text.New(m.text("独立勾选：")+strings.Join(m.treeStrictChecked, ", ")),
		muted(m.text("空树与整组禁用")),
		tree.New("", []tree.Item{}).Key("file-tree-empty").EmptyText(m.text("暂无文件")),
		fixedHeight(130, tree.New("", items).Key("file-tree-disabled").Disabled(true)),
	).Gap(16)
}

func demoTabs(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []tabs.TabItem{
		{Key: "overview", Label: m.text("概览"), Icon: lucide.House,
			Panel: text.New(m.text("概览：项目运行正常。"))},
		{Key: "activity", Label: m.text("动态"), Trailing: muted("3"),
			Panel: text.New(m.text("最近动态：更新了默认组件主题。"))},
		{Key: "settings", Label: m.text("设置"), Panel: text.New(m.text("项目设置：Go / Gio。"))},
		{Key: "disabled", Label: m.text("不可用"), Disabled: true},
	}
	changed := setString(send, func(m *demoState) *string { return &m.tab })
	documents := slices.Clone(m.navTabs)
	for index := range documents {
		documents[index].Icon = lucide.FileText
		documents[index].Panel = text.New(m.text("当前文档：") + documents[index].Label)
	}
	overflow := []tabs.TabItem{}
	for index := range 12 {
		label := fmt.Sprintf(m.text("文档 %d"), index+1)
		overflow = append(overflow, tabs.TabItem{
			Key: fmt.Sprintf("overflow-%d", index), Label: label, Panel: text.New(label),
		})
	}
	return layout.Column(
		muted(m.text("主要与次要样式 · 图标、数量与禁用项")),
		tabs.New(m.tab, items).Key("tabs-primary").OnChange(changed),
		divider.New(),
		tabs.New(m.tab, items).Key("tabs-secondary").Variant(tabs.Secondary).OnChange(changed),
		muted(m.text("尺寸与强调色")),
		tabs.New(m.tab, items).Key("tabs-small").Size(tabs.Small).OnChange(changed),
		tabs.New(m.tab, items).Key("tabs-large").Size(tabs.Large).Color(tabs.ColorAccent).
			Fit().Separators(true).OnChange(changed),
		muted(m.text("纵向与底部布局")),
		fixedHeight(170, tabs.New(m.tab, items).Key("tabs-vertical").Vertical().OnChange(changed)),
		tabs.New(m.tab, items).Key("tabs-bottom").Placement(tabs.Bottom).Centered(true).OnChange(changed),
		muted(m.text("文档标签 · 新增、关闭与重命名")),
		muted(m.text("点击加号新增；选中标签后按 F2 重命名，回车提交，Escape 取消。")),
		tabs.New(m.navTabSelected, documents).Key("tabs-documents").CloseOnHover(true).
			OnChange(setString(send, func(m *demoState) *string { return &m.navTabSelected })).
			OnAdd(func() { send(demoMsg(addDemoTab)) }).
			OnClose(func(key string) { send(demoMsg(func(m *demoState) { closeDemoTab(m, key) })) }).
			OnEdit(func(key, label string) {
				send(demoMsg(func(m *demoState) {
					m.navTabs = slices.Clone(m.navTabs)
					for index := range m.navTabs {
						if m.navTabs[index].Key == key {
							m.navTabs[index].Label = label
						}
					}
				}))
			}),
		muted(m.text("溢出菜单 · 缩窄窗口查看隐藏标签")),
		tabs.New("overflow-0", overflow).Key("tabs-overflow").Uncontrolled().Overflow(tabs.OverflowMenu).
			MoreLabel(m.text("更多文档")),
		muted(m.text("手动激活 · 方向键移动焦点，回车或空格切换")),
		tabs.New(m.tab, items).Key("tabs-manual").Activation(tabs.ActivationManual).
			PanelTransition(tabs.PanelFade).OnChange(changed),
		muted(m.text("整组禁用")),
		tabs.New("overview", items).Key("tabs-disabled").Disabled(true),
	).Gap(24)
}

func addDemoTab(m *demoState) {
	m.navTabID++
	key := fmt.Sprintf("document-%d", m.navTabID)
	m.navTabs = append(slices.Clone(m.navTabs), tabs.TabItem{
		Key: key, Label: fmt.Sprintf("untitled-%d.go", m.navTabID), Closable: true, Editable: true,
	})
	m.navTabSelected = key
}

func closeDemoTab(m *demoState, key string) {
	index := slices.IndexFunc(m.navTabs, func(item tabs.TabItem) bool { return item.Key == key })
	if index < 0 {
		return
	}
	m.navTabs = slices.Delete(slices.Clone(m.navTabs), index, index+1)
	if m.navTabSelected == key {
		m.navTabSelected = ""
		if len(m.navTabs) > 0 {
			m.navTabSelected = m.navTabs[min(index, len(m.navTabs)-1)].Key
		}
	}
}

func demoPagination(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := func(page int) { send(demoMsg(func(m *demoState) { m.navPage = page })) }
	return layout.Column(
		muted(m.text("基础分页 · 首尾页自动禁用前后按钮")),
		widgetnav.NewPagination(m.page, 12).Key("pagination").Labels(m.text("上一页"), m.text("下一页")).OnChange(func(page int) {
			send(demoMsg(func(m *demoState) { m.page = page }))
		}),
		text.New(fmt.Sprintf(m.text("第 %d 页 / 共 12 页"), m.page)),
		muted(m.text("大量页码 · 省略号与页码汇总")),
		widgetnav.NewPagination(m.navPage, 100).Key("pagination-many").Siblings(1).Boundaries(1).
			Labels(m.text("上一页"), m.text("下一页")).Summary(text.New(fmt.Sprintf(m.text("第 %d 页 / 共 100 页"), m.navPage))).
			OnChange(changed),
		layout.Wrap(
			button.New(m.text("跳到首页")).Variant(button.Secondary).
				OnClick(func() { changed(1) }),
			button.New(m.text("跳到末页")).Variant(button.Secondary).
				OnClick(func() { changed(100) }),
		).Gap(8),
		muted(m.text("小尺寸 · 仅页码")),
		widgetnav.NewPagination(m.navPage, 100).Key("pagination-small").
			Size(widgetnav.PaginationSizeSmall).ShowControls(false).Siblings(0).OnChange(changed),
		muted(m.text("大尺寸")),
		widgetnav.NewPagination(m.page, 12).Key("pagination-large").
			Size(widgetnav.PaginationSizeLarge).Labels(m.text("上一页"), m.text("下一页")).OnChange(func(page int) {
			send(demoMsg(func(m *demoState) { m.page = page }))
		}),
		muted(m.text("单页与禁用状态")),
		widgetnav.NewPagination(1, 1).Key("pagination-one").Labels(m.text("上一页"), m.text("下一页")),
		widgetnav.NewPagination(3, 12).Key("pagination-disabled").Labels(m.text("上一页"), m.text("下一页")).Disabled(true),
	).Gap(20)
}

func demoSidebar(m demoState, send ui.Send[Msg]) ui.Widget {
	sections := []sidebar.Section{
		{Title: m.text("工作区"), Items: []sidebar.Item{
			{Key: "overview", Label: m.text("概览"), Leading: icon.New(lucide.House).Size(16)},
			{Key: "projects", Label: m.text("项目"), Leading: icon.New(lucide.Folder).Size(16),
				Trailing: muted("2"), Children: []sidebar.Item{
					{Key: "flowui", Label: "FlowUI", Leading: icon.New(lucide.FileText).Size(16)},
					{Key: "studio", Label: "Studio", Disabled: true},
				}},
		}},
		{Title: m.text("资源"), Items: []sidebar.Item{
			{Key: "resources", Label: m.text("文档"), Leading: icon.New(lucide.BookOpen).Size(16),
				Children: []sidebar.Item{
					{Key: "guide", Label: m.text("使用指南")},
					{Key: "api", Label: "API"},
				}},
			{Key: "settings", Label: m.text("设置"), Leading: icon.New(lucide.Settings).Size(16)},
		}},
	}
	footer := ui.Widget(muted(m.text("我的工作区")))
	if m.sidebarCollapsed {
		footer = icon.New(lucide.User).Size(18)
	}
	nav := sidebar.NewSections(m.sidebarSelected, sections).Key("sidebar-demo").
		Width(240).CollapsedWidth(56).Collapsed(m.sidebarCollapsed).OpenKeys(m.sidebarOpen).
		Header(icon.New(lucide.PanelLeft).Size(20)).Footer(footer).
		OnOpenChange(func(keys []string) {
			send(demoMsg(func(m *demoState) { m.sidebarOpen = keys }))
		}).OnChange(setString(send, func(m *demoState) *string { return &m.sidebarSelected }))
	if m.sidebarHover {
		nav = nav.ExpandAction(sidebar.ExpandActionHover)
	}
	return layout.Column(
		toggle.NewSwitch(m.text("折叠为图标导航")).Key("sidebar-collapse").Checked(m.sidebarCollapsed).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.sidebarCollapsed })),
		toggle.NewSwitch(m.text("悬停展开分组")).Key("sidebar-hover").Checked(m.sidebarHover).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.sidebarHover })),
		muted(m.text("多级导航 · 分组、数量与禁用项；折叠后悬停图标打开子菜单")),
		layout.Box(nav).Align(layout.AlignTopStart).Style(style.New().FillWidth().Height(500)),
		text.New(m.text("当前选择：")+m.sidebarSelected),
		muted(m.text("空侧栏与整组禁用")),
		layout.Box(sidebar.New("", []sidebar.Item{}).Key("sidebar-empty").
			Width(240).EmptyText(m.text("暂无导航项"))).Align(layout.AlignTopStart).
			Style(style.New().FillWidth().Height(80)),
		layout.Box(sidebar.NewSections("overview", sections).Key("sidebar-disabled").Width(240).Disabled(true)).
			Align(layout.AlignTopStart).Style(style.New().FillWidth().Height(260)),
	).Gap(16)
}

func demoTimeline(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []timeline.Item{
		{Title: text.New(m.text("设计")), Content: text.New(m.text("确定默认主题与组件规范")), Color: timeline.ColorGreen},
		{Title: text.New(m.text("开发")), Content: text.New(m.text("构建全部组件展示"))},
		{Title: text.New(m.text("验证")), Content: text.New(m.text("检查交互与明暗模式"))},
	}
	return layout.Column(
		muted(m.text("起侧排列 · 描边节点")),
		timeline.New(items).Mode(timeline.ModeStart),
		muted(m.text("末侧排列 · 实心节点")),
		timeline.New(items).Mode(timeline.ModeEnd).Variant(timeline.VariantFilled),
		muted(m.text("交错排列")),
		timeline.New(items).Mode(timeline.ModeAlternate),
		muted(m.text("横向排列")),
		timeline.New([]timeline.Item{
			{Title: text.New(m.text("设计")), Color: timeline.ColorGreen},
			{Title: text.New(m.text("开发")), Color: timeline.ColorBlue},
			{Title: text.New(m.text("验证")), Color: timeline.ColorGray},
		}).Orientation(timeline.OrientationHorizontal).Variant(timeline.VariantFilled),
		muted(m.text("事件状态 · 自定义图标、颜色与禁用项")),
		toggle.NewSwitch(m.text("倒序展示")).Key("timeline-reverse").Checked(m.timelineReverse).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.timelineReverse })),
		toggle.NewSwitch(m.text("演示进行中动画")).Key("timeline-loading").Checked(m.timelineLoading).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.timelineLoading })),
		timeline.New([]timeline.Item{
			{Title: text.New(m.text("已发布")), Icon: icon.New(lucide.Check).Size(14), Color: timeline.ColorGreen},
			{Title: text.New(m.text("进行中")), Loading: m.timelineLoading},
			{Title: text.New(m.text("需要处理")), Content: text.New(m.text("等待修复后重试")), Color: timeline.ColorRed},
			{Title: text.New(m.text("已取消")), Disabled: true, Color: timeline.ColorGray},
		}).Reverse(m.timelineReverse).Gap(16).
			Pending(text.New(m.text("等待下一步"))).PendingLoading(m.timelineLoading),
		muted(m.text("整组禁用")),
		timeline.New(items).Disabled(true),
	).Gap(20)
}

func demoCollapsible(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		collapsible.New(m.text("项目详情"), text.New(m.text("FlowUI 是基于 Gio 的 Go 桌面 UI 框架。"))).Key("collapse-single").Expanded(m.expanded).OnExpandedChange(setBool(send, func(m *demoState) *bool { return &m.expanded })),
		collapsible.Group("collapse-group", m.keys, []collapsible.Item{
			{Key: "go", Label: m.text("如何运行？"), Content: text.New(m.text("在仓库根目录运行 go run ./examples/showcase"))},
			{Key: "theme", Label: m.text("如何切换主题？"), Content: text.New(m.text("点击右上角的月亮或太阳图标切换主题。"))},
		}).OnExpandedChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.keys = keys })) }),
		muted(m.text("多项同时展开 · 标题图标与尾部信息")),
		collapsible.Group("collapse-multiple", m.collapseMultiple, []collapsible.Item{
			{Key: "go", Label: m.text("如何运行？"), Leading: icon.New(lucide.Terminal).Size(16),
				Trailing: muted("Go"), Content: text.New("go run ./examples/showcase")},
			{Key: "theme", Label: m.text("如何切换主题？"), Leading: icon.New(lucide.Palette).Size(16),
				Content: text.New(m.text("点击右上角的月亮或太阳图标切换主题。"))},
			{Key: "disabled", Label: m.text("不可用"), Content: text.New(""), Disabled: true},
		}).AllowMultipleExpanded(true).OnExpandedChange(func(keys []string) {
			send(demoMsg(func(m *demoState) { m.collapseMultiple = keys }))
		}),
		muted(m.text("默认展开与禁用状态")),
		collapsible.New(m.text("默认展开"), text.New(m.text("此面板由组件内部维护展开状态。"))).
			Key("collapse-default").DefaultExpanded(true),
		collapsible.New(m.text("不可用"), text.New("")).Key("collapse-disabled").Disabled(true),
		collapsible.Group("collapse-disabled-group", []string{"info"}, []collapsible.Item{
			{Key: "info", Label: m.text("项目详情"), Content: text.New(m.text("组件库"))},
		}).Disabled(true),
	).Gap(20)
}
