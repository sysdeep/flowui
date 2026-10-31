package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	"gioui.org/font"
	giolayout "gioui.org/layout"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/linechart"
	"github.com/qianniancn/flowui/widget/modal"
	navwidget "github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/selector"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toggle"
)

type overviewSettings struct {
	name, framework         string
	date                    time.Time
	autoSave, notifications bool
}

type overviewProject struct {
	key, name, owner, status string
	progress                 float64
	date                     time.Time
}

type overviewState struct {
	settings, saved         overviewSettings
	period                  string
	page                    int
	selected                []string
	sort                    table.SortDescriptor
	projects                []overviewProject
	frameworkOpen, creating bool
	newName                 string
	nextID                  int
}

func initialOverview() overviewState {
	date := time.Date(2026, time.October, 31, 0, 0, 0, 0, time.Local)
	settings := overviewSettings{
		name: "FlowUI Desktop", framework: "gio", date: date, autoSave: true, notifications: true,
	}
	names := []string{
		"组件库", "组件展示示例", "主题系统", "使用文档",
		"桌面工作台", "图表面板", "项目设置", "资源管理器",
		"发布工具", "表单设计器", "快捷命令", "数据导入",
	}
	owners := []string{"Alex Chen", "Lin Wang", "Sam Lee", "Mia Zhou"}
	statuses := []string{"进行中", "进行中", "已完成", "审核中"}
	values := []float64{.76, .62, 1, .9}
	projects := make([]overviewProject, len(names))
	for index, name := range names {
		projects[index] = overviewProject{
			key: fmt.Sprintf("project-%d", index+1), name: name, owner: owners[index%4],
			status: statuses[index%4], progress: values[index%4], date: date,
		}
	}
	return overviewState{
		settings: settings, saved: settings, period: "week", page: 1,
		selected: []string{"project-1"}, projects: projects, nextID: len(projects) + 1,
	}
}

func editOverview(send ui.Send[Msg], change func(*overviewSettings)) {
	send(demoMsg(func(m *demoState) {
		change(&m.overview.settings)
		if m.overview.settings.autoSave {
			m.overview.saved = m.overview.settings
		}
	}))
}

func overviewCard(child ui.Widget) ui.Widget {
	return surface.NewCard(child).StyleWithMetrics(style.New().FillWidth(), 20, 0)
}

func overviewHeading(value string) ui.Widget {
	return text.New(value).Size(16).Weight(font.Bold)
}

func demoOverview(m demoState, send ui.Send[Msg]) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		wide := gtx.Constraints.Max.X >= gtx.Dp(unit.Dp(880))
		stat := func(label, value, detail string) ui.Widget {
			return overviewCard(layout.Column(muted(m.text(label)),
				text.New(value).Size(28).Weight(font.Bold), muted(m.text(detail))).Gap(8))
		}
		stats := []ui.Widget{
			stat("活跃项目", fmt.Sprint(len(m.overview.projects)), "本地演示数据"),
			stat("已完成任务", "1,284", "本月增长 18.6%"),
			stat("发布进度", "76%", "下次发布 · 10 月 31 日"),
		}
		var statistics ui.Widget = layout.Column(stats...).Gap(16)
		if gtx.Constraints.Max.X >= gtx.Dp(600) {
			statistics = layout.Row(layout.Expanded(stats[0]), layout.Expanded(stats[1]),
				layout.Expanded(stats[2])).Gap(16)
		}
		var middle ui.Widget = layout.Column(overviewActivity(m, send), overviewForm(m, send)).Gap(16)
		if wide {
			middle = layout.Row(layout.Flexible(1.25, overviewActivity(m, send)),
				layout.Flexible(1, overviewForm(m, send))).Gap(16)
		}
		intro := layout.Column(
			text.New(m.text("构建你的原生应用。")).Size(30).Weight(font.Bold),
			muted(m.text("用可组合的 Go 组件，构建风格一致的桌面界面。")),
		).Gap(8)
		newProject := button.New(m.text("新建项目")).Key("overview-new").Content(layout.Row(
			icon.New(lucide.Plus).Size(16), text.New(m.text("新建项目")),
		).Gap(8).AlignMiddle()).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.overview.creating, m.overview.newName = true, "" }))
		})
		var title ui.Widget = layout.Column(intro, newProject).Gap(12)
		if wide {
			title = layout.Row(layout.Expanded(intro), newProject).Gap(16).AlignMiddle()
		}
		return layout.Column(title, statistics, middle, overviewTable(m, send),
			overviewCreateDialog(m, send)).Gap(20).FillWidth().Layout(ctx, gtx)
	})
}

func overviewActivity(m demoState, send ui.Send[Msg]) ui.Widget {
	completed := []float64{22, 36, 29, 48, 42, 62, 76}
	planned := []float64{16, 22, 24, 30, 38, 42, 52}
	labels := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	if m.overview.period == "month" {
		completed, planned = []float64{110, 180, 230, 320}, []float64{100, 160, 220, 280}
		labels = []string{"第一周", "第二周", "第三周", "第四周"}
	}
	for index := range labels {
		labels[index] = m.text(labels[index])
	}
	chart := linechart.New("overview-activity", []linechart.Series{
		linechart.Values("completed", m.text("已完成"), completed).Smooth(true).Area(true),
		linechart.Values("planned", m.text("计划"), planned).Smooth(true),
	}).Categories(labels).Height(220).Legend(true).Animation(false).OnDataClick(chartSelection(m, send))
	period := toggle.NewSegmented([]toggle.SegmentedItem{
		{Key: "week", Label: m.text("本周")}, {Key: "month", Label: m.text("本月")},
	}).Key("overview-period").SelectedKey(m.overview.period).OnChange(func(value string) {
		send(demoMsg(func(m *demoState) { m.overview.period = value }))
	})
	return overviewCard(layout.Column(layout.Row(layout.Expanded(overviewHeading(m.text("项目动态"))),
		period).Gap(8).AlignMiddle(), chart).Gap(12))
}

func overviewForm(m demoState, send ui.Send[Msg]) ui.Widget {
	s := m.overview.settings
	changedName := func(value string) { editOverview(send, func(s *overviewSettings) { s.name = value }) }
	framework := selector.New([]selector.SelectItem{
		{Key: "gio", Label: "Go / Gio"}, {Key: "custom", Label: m.text("自定义")},
	}).Key("overview-framework").Label(m.text("框架")).SelectedKey(s.framework).
		Open(m.overview.frameworkOpen).FillWidth().OnChange(func(value string) {
		editOverview(send, func(s *overviewSettings) { s.framework = value })
	}).OnOpenChange(func(open bool) {
		send(demoMsg(func(m *demoState) { m.overview.frameworkOpen = open }))
	})
	date := datepicker.New("overview-date", s.date).FillWidth().OnChange(func(value time.Time) {
		editOverview(send, func(s *overviewSettings) { s.date = value })
	})
	fields := ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		frameworkField, dateField := field(m.text("框架"), framework), field(m.text("发布日期"), date)
		if gtx.Constraints.Max.X < gtx.Dp(400) {
			return layout.Column(frameworkField, dateField).Gap(12).Layout(ctx, gtx)
		}
		return layout.Row(layout.Expanded(frameworkField), layout.Expanded(dateField)).Gap(12).Layout(ctx, gtx)
	})
	return overviewCard(layout.Column(
		overviewHeading(m.text("项目设置")),
		field(m.text("项目名称"), input.New(s.name).Key("overview-name").Label(m.text("项目名称输入框")).
			FillWidth().OnChange(changedName)),
		fields,
		layout.Wrap(
			toggle.NewCheckbox(m.text("自动保存")).Key("overview-auto-save").Checked(s.autoSave).
				OnChange(func(value bool) { editOverview(send, func(s *overviewSettings) { s.autoSave = value }) }),
			toggle.NewSwitch(m.text("通知")).Key("overview-notifications").Checked(s.notifications).
				OnChange(func(value bool) { editOverview(send, func(s *overviewSettings) { s.notifications = value }) }),
		).Gap(18).LineGap(10),
		layout.Wrap(
			button.New(m.text("保存更改")).Key("overview-save").OnClick(func() {
				send(demoMsg(func(m *demoState) { m.overview.saved = m.overview.settings }))
				send(logMsg(m.text("所有更改已保存")))
			}),
			button.New(m.text("取消")).Key("overview-cancel").Variant(button.Outline).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.overview.settings = m.overview.saved }))
				send(logMsg(m.text("已恢复保存的设置")))
			}),
		).Gap(10).LineGap(10),
	).Gap(18))
}

func overviewRows(m demoState) []table.Row {
	projects := slices.Clone(m.overview.projects)
	if m.overview.sort.Column == "project" {
		slices.SortStableFunc(projects, func(a, b overviewProject) int {
			order := cmp.Compare(m.text(a.name), m.text(b.name))
			if m.overview.sort.Direction == table.SortDescending {
				return -order
			}
			return order
		})
	}
	start := (m.overview.page - 1) * 4
	end := min(start+4, len(projects))
	rows := []table.Row{}
	for _, project := range projects[start:end] {
		color := chip.ColorAccent
		switch project.status {
		case "已完成":
			color = chip.ColorSuccess
		case "审核中":
			color = chip.ColorWarning
		}
		rows = append(rows, table.Row{Key: project.key, Cells: []table.Cell{
			{Text: m.text(project.name)}, {Text: project.owner},
			{Content: chip.New(m.text(project.status)).Color(color).Variant(chip.VariantSoft).Size(chip.SizeSmall)},
			{Content: progress.Bar("overview-progress-"+project.key, project.progress).Range(0, 1).ShowValue()},
			{Text: project.date.Format("2006-01-02")},
		}})
	}
	return rows
}

func overviewTable(m demoState, send ui.Send[Msg]) ui.Widget {
	rows := overviewRows(m)
	visibleSelection := []string{}
	for _, row := range rows {
		if slices.Contains(m.overview.selected, row.Key) {
			visibleSelection = append(visibleSelection, row.Key)
		}
	}
	projects := table.New("overview-table", []table.Column{
		{Key: "project", Label: m.text("项目"), Weight: 2, MinWidth: 160, Sortable: true},
		{Key: "owner", Label: m.text("负责人"), Width: 120},
		{Key: "status", Label: m.text("状态"), Width: 140},
		{Key: "progress", Label: m.text("进度"), Weight: 1, MinWidth: 140},
		{Key: "updated", Label: m.text("更新日期"), Width: 120},
	}, rows).SelectionMode(table.SelectionMultiple).SelectedKeys(visibleSelection).
		ShowSelectionIndicator().Border(true).GridLines(true).RowHeight(36).HeaderHeight(36).
		MinWidth(760).SortDescriptor(m.overview.sort).OnSortChange(func(value table.SortDescriptor) {
		send(demoMsg(func(m *demoState) { m.overview.sort, m.overview.page = value, 1 }))
	}).OnSelectionChange(func(keys []string) {
		// The table reports selection for the current page; retain selections on other pages.
		send(demoMsg(func(m *demoState) {
			selected := slices.DeleteFunc(slices.Clone(m.overview.selected), func(key string) bool {
				return slices.ContainsFunc(rows, func(row table.Row) bool { return row.Key == key })
			})
			m.overview.selected = append(selected, keys...)
		}))
	}).OnRowActivate(func(key string) { send(logMsg(m.text("已选择：") + key)) })
	pages := max(1, (len(m.overview.projects)+3)/4)
	return overviewCard(layout.Column(
		layout.Wrap(overviewHeading(m.text("当前项目")),
			muted(fmt.Sprintf(m.text("%d 个项目 · 已选 %d 项"), len(m.overview.projects), len(m.overview.selected))),
			button.New(m.text("查看全部")).Key("overview-all").Variant(button.Ghost).Size(button.Small).
				OnClick(func() { send(selectMsg("table")) }),
		).Gap(16).LineGap(8),
		fixedHeight(248, projects),
		navwidget.NewPagination("overview-pages", m.overview.page, pages).OnChange(func(page int) {
			send(demoMsg(func(m *demoState) { m.overview.page = page }))
		}),
	).Gap(12))
}

func overviewCreateDialog(m demoState, send ui.Send[Msg]) ui.Widget {
	return modal.New(m.text("新建项目"), field(m.text("项目名称"),
		input.New(m.overview.newName).Key("overview-new-name").Label(m.text("新项目名称")).FillWidth().
			OnChange(func(value string) { send(demoMsg(func(m *demoState) { m.overview.newName = value })) }),
	)).Key("overview-create").Open(m.overview.creating).OnOpenChange(func(open bool) {
		send(demoMsg(func(m *demoState) { m.overview.creating = open }))
	}).Footer(button.New(m.text("创建项目")).Key("overview-create-submit").
		Disabled(strings.TrimSpace(m.overview.newName) == "").OnClick(func() {
		send(demoMsg(func(m *demoState) {
			s := &m.overview
			s.projects = append(slices.Clone(s.projects), overviewProject{
				key: fmt.Sprintf("project-%d", s.nextID), name: strings.TrimSpace(s.newName),
				owner: "Alex Chen", status: "进行中", date: s.settings.date,
			})
			s.nextID++
			s.creating, s.sort, s.page = false, table.SortDescriptor{}, (len(s.projects)+3)/4
		}))
		send(logMsg(m.text("已新建演示项目")))
	}))
}
