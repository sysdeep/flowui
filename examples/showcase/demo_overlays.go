package main

import (
	"fmt"
	"image"
	"slices"
	"time"

	giolayout "gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alert"
	"github.com/qianniancn/flowui/widget/alertdialog"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/popover"
	"github.com/qianniancn/flowui/widget/portal"
	"github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/toggle"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func actionItems(m demoState, send ui.Send[Msg]) []menu.Item {
	return []menu.Item{
		{Key: "new", Label: m.text("新建项目"), Shortcut: "Ctrl+N", OnAction: report(send, m.text("已新建演示项目"))},
		{Key: "save", Label: m.text("保存"), Shortcut: "Ctrl+S", OnAction: report(send, m.text("已保存演示项目"))},
		{Key: "export", Label: m.text("导出"), Children: []menu.Item{
			{Key: "png", Label: m.text("PNG 图片"), OnAction: report(send, m.text("已选择导出 PNG"))},
			{Key: "json", Label: m.text("JSON 数据"), OnAction: report(send, m.text("已选择导出 JSON"))},
		}},
		menu.Separator(),
		{Key: "disabled", Label: m.text("暂不可用"), Disabled: true},
		menu.Separator(),
		{Key: "delete", Label: m.text("删除项目"), Shortcut: "Del", Variant: menu.ItemDanger,
			OnAction: report(send, m.text("已删除演示项目"))},
	}
}

func demoMenu(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		menu.New("inline-menu", actionItems(m, send)),
		field(m.text("勾选与互斥菜单"), menu.New("settings-menu", []menu.Item{
			{Key: "sync", Label: m.text("自动同步"), Kind: menu.ItemCheckbox, Checked: m.enabled,
				OnAction: func() { send(demoMsg(func(m *demoState) { m.enabled = !m.enabled })) }},
			menu.Separator(),
			{Key: "day", Label: m.text("日"), Kind: menu.ItemRadio, Checked: m.segment == "day", RadioGroup: "period",
				OnAction: func() { send(demoMsg(func(m *demoState) { m.segment = "day" })) }},
			{Key: "week", Label: m.text("周"), Kind: menu.ItemRadio, Checked: m.segment == "week", RadioGroup: "period",
				OnAction: func() { send(demoMsg(func(m *demoState) { m.segment = "week" })) }},
		}).Compact(true)),
		menu.ContextMenu("context-demo", button.New(m.text("在此右键打开菜单")).Variant(button.Secondary),
			menu.New("context-actions", actionItems(m, send))),
	).Gap(24)
}

func demoDropdown(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		dropdown.New("dropdown-demo", button.New(m.text("更多操作")), actionItems(m, send)),
		dropdown.Button("dropdown-split", button.New(m.text("发布")), actionItems(m, send)).
			OnClick(report(send, m.text("已发布演示项目"))),
		dropdown.New("dropdown-hover", button.New(m.text("悬停展开")).Variant(button.Secondary), actionItems(m, send)).
			TriggerMode(dropdown.TriggerHover),
		dropdown.New("dropdown-disabled", button.New(m.text("不可用")).Variant(button.Secondary), actionItems(m, send)).Disabled(true),
	).Gap(16).LineGap(16)
}

func demoMenubar(m demoState, send ui.Send[Msg]) ui.Widget {
	return menubar.New("menubar-demo", []menubar.Item{
		menubar.NewMenu("file", m.text("文件"), actionItems(m, send)),
		menubar.NewMenu("edit", m.text("编辑"), []menu.Item{
			{Key: "undo", Label: m.text("撤销"), OnAction: report(send, m.text("撤销操作"))},
			{Key: "copy", Label: m.text("复制"), OnAction: report(send, m.text("复制操作"))},
		}),
		menubar.NewMenu("unavailable", m.text("暂不可用"), nil).Disabled(true),
	})
}

func demoModal(m demoState, send ui.Send[Msg]) ui.Widget {
	openChanged := setBool(send, func(m *demoState) *bool { return &m.dialog })
	closeExample := func() { send(demoMsg(func(m *demoState) { m.modalExample = "" })) }
	paragraphs := make([]ui.Widget, 0, 16)
	for index := range 16 {
		paragraphs = append(paragraphs, text.New(fmt.Sprintf(m.text("第 %d 节 · 滚动正文查看内容，底部操作始终保留。"), index+1)))
	}
	return layout.Column(
		button.New(m.text("编辑项目")).OnClick(func() { openChanged(true) }),
		modal.New(m.text("编辑项目"), layout.Column(
			text.New(m.text("修改项目名称，然后保存。")),
			input.New(m.name).Key("modal-name").Label(m.text("编辑项目名称")).FillWidth().
				OnChange(setString(send, func(m *demoState) *string { return &m.name })),
		).Gap(16)).Key("demo-dialog").Open(m.dialog).OnOpenChange(openChanged).Footer(layout.Wrap(
			button.New(m.text("取消")).Variant(button.Secondary).OnClick(func() { openChanged(false) }),
			button.New(m.text("保存")).OnClick(func() {
				openChanged(false)
				send(logMsg(m.text("已保存项目：") + m.name))
			}),
		).Gap(8)),
		layout.Wrap(
			button.New(m.text("长内容弹窗")).Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.modalExample = "scroll" }))
			}),
			button.New(m.text("必须明确关闭")).Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.modalExample = "explicit" }))
			}),
		).Gap(12).LineGap(12),
		modal.New(m.text("长内容弹窗"), layout.Column(paragraphs...).Gap(24)).Key("modal-scroll").
			Open(m.modalExample == "scroll").Size(modal.Large).Scroll(modal.ScrollInside).
			OnOpenChange(func(open bool) {
				if !open {
					closeExample()
				}
			}).
			Footer(button.New(m.text("完成阅读")).OnClick(closeExample)),
		modal.New(m.text("必须明确关闭"), text.New(m.text("点击遮罩或按 Esc 不会关闭，请使用下方按钮。"))).
			Key("modal-explicit").Open(m.modalExample == "explicit").Size(modal.Small).
			Dismissable(false).KeyboardDismissDisabled(true).CloseButton(false).
			Footer(button.New(m.text("我已了解")).OnClick(closeExample)),
	).Gap(12)
}

func demoAlertDialog(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := setBool(send, func(m *demoState) *bool { return &m.confirm })
	closeSuccess := func() { send(demoMsg(func(m *demoState) { m.modalExample = "" })) }
	return layout.Column(
		button.New(m.text("删除演示项目")).Variant(button.Danger).OnClick(func() { changed(true) }),
		alertdialog.New(m.text("删除项目？"), m.text("此处仅演示确认流程，不涉及磁盘文件。这个操作需要明确确认。")).Key("demo-confirm").Open(m.confirm).
			Status(alertdialog.StatusDanger).OnOpenChange(changed).Footer(layout.Wrap(
			button.New(m.text("保留项目")).Variant(button.Secondary).OnClick(func() { changed(false) }),
			button.New(m.text("确认删除")).Variant(button.Danger).OnClick(func() {
				changed(false)
				send(logMsg(m.text("已确认删除演示项目")))
			}),
		).Gap(8)),
		button.New(m.text("成功确认弹窗")).Variant(button.Secondary).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.modalExample = "success" }))
		}),
		alertdialog.New(m.text("保存成功"), m.text("项目的最新更改已同步。")).Key("confirm-success").
			Open(m.modalExample == "success").Status(alertdialog.StatusSuccess).Size(alertdialog.SizeSmall).
			OnOpenChange(func(open bool) {
				if !open {
					closeSuccess()
				}
			}).
			Footer(button.New(m.text("我已了解")).OnClick(closeSuccess)),
	).Gap(12)
}

func demoPopover(m demoState, send ui.Send[Msg]) ui.Widget {
	openChanged := setBool(send, func(m *demoState) *bool { return &m.popup })
	placements := []ui.Widget{}
	for _, item := range []struct {
		key       string
		label     string
		placement popover.Placement
	}{
		{"top", "上方", popover.Top}, {"bottom", "下方", popover.Bottom},
		{"left", "左侧", popover.Left}, {"right", "右侧", popover.Right},
	} {
		placements = append(placements, popover.New(button.New(m.text(item.label)).Variant(button.Secondary).
			OnClick(func() {
				send(demoMsg(func(m *demoState) {
					if m.popoverPlacement == item.key {
						m.popoverPlacement = ""
					} else {
						m.popoverPlacement = item.key
					}
				}))
			}), text.New(m.text("点击外部或按 Esc 关闭"))).Key("popover-"+item.key).
			Heading(m.text(item.label)).Placement(item.placement).Arrow(true).Open(m.popoverPlacement == item.key).
			OnOpenChange(func(open bool) {
				if !open {
					send(demoMsg(func(m *demoState) { m.popoverPlacement = "" }))
				}
			}))
	}
	return layout.Column(popover.New(button.New(m.text("项目信息")).
		OnClick(func() { openChanged(!m.popup) }), layout.Column(text.New(m.text("FlowUI · 本地项目")),
		button.New(m.text("查看详情")).OnClick(report(send, m.text("查看项目详情"))),
	).Gap(16)).Key("popover-demo").Open(m.popup).Heading(m.text("项目信息")).
		OnOpenChange(openChanged),
		field(m.text("四向箭头浮层 · 空间不足时自动翻转"), layout.Wrap(placements...).Gap(12).LineGap(12)),
	).Gap(20)
}

func demoTooltip(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			tooltip.New("tooltip-demo", button.New(m.text("悬停或聚焦我")).
				OnClick(report(send, m.text("已保存当前文档"))), m.text("保存当前文档（Ctrl+S）")),
			tooltip.New("tooltip-bottom", button.New(m.text("下方提示")).Variant(button.Secondary), m.text("保存当前文档（Ctrl+S）")).
				Placement(popover.Bottom).Arrow(true),
			tooltip.New("tooltip-focus", button.New(m.text("聚焦时提示")).Variant(button.Secondary), m.text("保存当前文档（Ctrl+S）")).
				Trigger(tooltip.Focus),
		).Gap(12).LineGap(12),
		muted(m.text("使用 Tab 聚焦按钮，查看焦点触发的提示。")),
	).Gap(16)
}

func demoPortal(m demoState, send ui.Send[Msg]) ui.Widget {
	trigger := button.New(m.text("切换根浮层")).OnClick(func() {
		send(demoMsg(func(m *demoState) { m.portal = !m.portal }))
	})
	return portal.New("portal-demo", m.portal, trigger, func(anchor image.Rectangle, _ bool) ui.Widget {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			position := image.Pt(anchor.Min.X, anchor.Max.Y+8)
			defer op.Offset(position).Push(gtx.Ops).Pop()
			gtx.Constraints.Min = image.Point{}
			gtx.Constraints.Max.X = max(1, gtx.Constraints.Max.X-position.X)
			gtx.Constraints.Max.Y = max(1, gtx.Constraints.Max.Y-position.Y)
			return surface.NewCard(text.New(m.text("这段内容绘制在窗口根浮层。"))).Layout(ctx, gtx)
		})
	}).Passive(true)
}

func demoAlert(m demoState, send ui.Send[Msg]) ui.Widget {
	var dismissible ui.Widget = button.New(m.text("恢复提示")).Variant(button.Secondary).
		OnClick(func() { send(demoMsg(func(m *demoState) { m.alertHidden = false })) })
	if !m.alertHidden {
		dismissible = alert.New(m.text("可关闭的提示"), m.text("关闭后可通过按钮恢复。")).Key("dismissible-alert").
			Closable(true).OnClose(func() { send(demoMsg(func(m *demoState) { m.alertHidden = true })) })
	}
	return layout.Column(
		alert.New(m.text("项目提示"), m.text("所有演示都使用本地数据。")),
		alert.New(m.text("保存成功"), m.text("项目的最新更改已同步。")).Status(alert.Success),
		alert.New(m.text("即将达到限额"), m.text("请检查项目使用量。")).Status(alert.Warning),
		alert.New(m.text("无法连接"), m.text("请检查网络设置后重试。")).Status(alert.Danger),
		dismissible,
		alert.New(m.text("发现可用更新"), m.text("带操作按钮的提示。")).Action(
			button.New(m.text("查看更新")).Variant(button.Secondary).OnClick(report(send, m.text("已查看更新")))),
	).Gap(12)
}

func demoToast(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			button.New(m.text("显示成功消息")).OnClick(func() { addToast(send, false) }),
			button.New(m.text("显示可撤销消息")).Variant(button.Secondary).
				OnClick(func() { addToast(send, true) }),
		).Gap(12).LineGap(12),
		layout.Wrap(
			button.New(m.text("显示普通消息")).Variant(button.Secondary).OnClick(func() { addVariantToast(send, toast.Default) }),
			button.New(m.text("显示警告消息")).Variant(button.Secondary).OnClick(func() { addVariantToast(send, toast.Warning) }),
			button.New(m.text("显示错误消息")).Variant(button.Danger).OnClick(func() { addVariantToast(send, toast.Danger) }),
		).Gap(12).LineGap(12),
		muted(m.text("消息出现在窗口角落，4 秒后消失，也可以手动关闭。")),
	).Gap(16)
}

func addToast(send ui.Send[Msg], action bool) {
	send(demoMsg(func(m *demoState) {
		m.toastID++
		m.toasts = append(m.toasts, toastNotice{key: fmt.Sprintf("toast-%d", m.toastID), action: action, variant: toast.Success})
	}))
}

func addVariantToast(send ui.Send[Msg], variant toast.Variant) {
	send(demoMsg(func(m *demoState) {
		m.toastID++
		m.toasts = append(m.toasts, toastNotice{key: fmt.Sprintf("toast-%d", m.toastID), variant: variant})
	}))
}

func toastItems(m demoState) []toast.Item {
	items := make([]toast.Item, 0, len(m.toasts))
	for _, notification := range m.toasts {
		title, description := "保存成功", "演示更改已保存"
		switch notification.variant {
		case toast.Default:
			title, description = "项目提示", "这是一条普通通知。"
		case toast.Warning:
			title, description = "即将达到限额", "请检查项目使用量。"
		case toast.Danger:
			title, description = "保存失败", "演示错误消息，请稍后重试。"
		}
		item := toast.NewItem(notification.key, m.text(title)).
			Description(m.text(description)).Variant(notification.variant).Timeout(4 * time.Second)
		if notification.action {
			item = item.Action(m.text("撤销"))
		}
		items = append(items, item)
	}
	return items
}

func removeToast(send ui.Send[Msg], key string) {
	send(demoMsg(func(m *demoState) {
		m.toasts = slices.DeleteFunc(slices.Clone(m.toasts), func(item toastNotice) bool { return item.key == key })
	}))
}

func demoProgress(m demoState, send ui.Send[Msg]) ui.Widget {
	var pending ui.Widget = muted(m.text("开启后展示无法确定完成时间的任务。"))
	if m.progressPending {
		pending = layout.Column(
			progress.Bar("progress-pending", 0).Label(m.text("正在处理")).Indeterminate(),
			progress.Circle("circle-pending", 0).Indeterminate(),
		).Gap(12)
	}
	return layout.Column(
		slider.New("progress-control", m.value).Label(m.text("调整进度")).ShowValue().OnChange(setNumber(send)),
		progress.Bar("progress-bar", m.value).Label(m.text("上传进度")).ShowValue(),
		progress.Circle("progress-circle", m.value),
		progress.Meter("progress-meter", m.value).Label(m.text("存储使用量")).ShowValue(),
		field(m.text("语义颜色与尺寸"), layout.Column(
			progress.Bar("progress-success", 100).Color(progress.BarSuccess).Size(progress.BarSmall).ShowValue(),
			progress.Bar("progress-warning", m.value).Color(progress.BarWarning).ShowValue(),
			progress.Bar("progress-danger", m.value).Color(progress.BarDanger).Size(progress.BarLarge).ShowValue(),
			progress.Bar("progress-empty", 0).ShowValue(),
			progress.Bar("progress-disabled", 40).Disabled(true).ShowValue(),
		).Gap(12)),
		progress.Bar("progress-files", m.value*5).Range(0, 500).Label(m.text("自定义范围与数值格式")).ShowValue().
			ValueFormatter(func(value float64) string { return fmt.Sprintf("%.0f / 500", value) }),
		toggle.NewSwitch(m.text("演示不确定进度")).Key("progress-pending-switch").Checked(m.progressPending).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.progressPending })),
		pending,
	).Gap(24)
}

func demoSpinner(m demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(layout.Wrap(
		spinner.New().Size(spinner.Small), spinner.New(), spinner.New().Size(spinner.Large),
		spinner.New().Size(spinner.ExtraLarge), text.New(m.text("正在加载…")),
	).Gap(24).AlignMiddle(),
		field(m.text("语义颜色"), layout.Wrap(
			spinner.New().Color(spinner.Current).Label(m.text("默认")),
			spinner.New().Color(spinner.Success).Label(m.text("成功")),
			spinner.New().Color(spinner.Warning).Label(m.text("警告")),
			spinner.New().Color(spinner.Danger).Label(m.text("错误")),
		).Gap(24).AlignMiddle()),
	).Gap(24)
}
