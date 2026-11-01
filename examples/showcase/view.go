package main

import (
	"fmt"
	"strings"

	"gioui.org/font"
	giolayout "gioui.org/layout"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	widgetcode "github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/popover"
	"github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/window"
)

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		compact := gtx.Constraints.Max.X < gtx.Dp(unit.Dp(820))
		var body ui.Widget = detail(m, send, compact)
		if compact && m.navOpen {
			body = navigation(m, send, true)
		}
		if !compact {
			body = layout.Row(navigation(m, send, false), layout.Expanded(body))
		}
		root := layout.Stack(
			layout.Stacked(layout.Column(
				showcaseHeader(m, send, compact),
				divider.New(),
				layout.Expanded(body),
				statusbar.New(text.New(m.status).MaxLines(1).Truncator("…"), text.New("FlowUI · Go / Gio")),
			)),
			layout.Overlay(toast.New(toastItems(m.demo)).Key("showcase-toasts").
				OnClose(func(key string) { removeToast(send, key) }).
				OnAction(func(key string) {
					removeToast(send, key)
					send(logMsg(m.text("已撤销演示操作")))
				})),
		)
		return root.Layout(ctx, gtx)
	})
}

func navigation(m Model, send ui.Send[Msg], compact bool) ui.Widget {
	sections, count := filteredSections(m.query, m.demo)
	header := []ui.Widget{
		muted(m.text("组件展示")),
		input.New(m.query).Key("catalog-search").Label(m.text("搜索组件")).
			Hint(m.text("搜索名称、用途或包名")).FillWidth().
			OnChange(func(value string) { send(searchMsg(value)) }),
		muted(fmt.Sprintf(m.text("%d / %d 个演示"), count, len(catalog))),
	}
	nav := sidebar.NewSections(m.selected, sections).Key("catalog").
		Width(280).Padding(16).ItemHeight(34).ItemGap(2).
		Header(layout.Column(header...).Gap(12)).
		EmptyText(m.text("没有匹配组件，请换个关键词")).
		OnChange(func(key string) { send(selectMsg(key)) })
	if compact {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			width := int(gtx.Metric.PxToDp(gtx.Constraints.Max.X))
			return nav.Width(max(1, width)).Layout(ctx, gtx)
		})
	}
	return nav
}

func detail(m Model, send ui.Send[Msg], compact bool) ui.Widget {
	if m.selected == "overview" {
		return layout.Scroll(layout.Box(demoOverview(m.demo, send)).
			Style(style.New().FillWidth().Padding(24))).Key("overview-scroll").Vertical()
	}
	item, ok := findEntry(m.selected)
	if !ok {
		item = catalog[0]
	}
	item = item.localized(m.demo)
	padding := unit.Dp(28)
	if compact {
		padding = 16
	}
	path := "github.com/qianniancn/flowui/layout"
	if item.packages != "" {
		path = "github.com/qianniancn/flowui/widget/" + strings.Fields(item.packages)[0]
	}
	if item.importPath != "" {
		path = item.importPath
	}
	sections := []ui.Widget{
		layout.Column(
			muted(item.group),
			text.New(item.title).Size(28).Weight(font.Bold),
			text.New(item.summary).Size(14),
		).Gap(10),
		section(m.text("交互演示"), item.demo(m.demo, send)),
	}
	if item.code != "" {
		sections = append(sections, section(m.text("基本用法"), widgetcode.New("go", item.code).Key("code-"+item.key)))
	}
	if documentation := componentDocumentation(item.key, m.demo.language); documentation != nil {
		sections = append(sections, documentation)
	}
	sections = append(sections, layout.Column(
		muted(path),
		muted(m.text("完整示例：examples/showcase/")+item.source),
		muted(m.text("修改控件后可在底部查看反馈；“重置演示”恢复初始值。")),
	).Gap(8))
	content := layout.Column(sections...).Gap(24)
	return layout.Scroll(
		layout.Box(content).Style(style.New().FillWidth().MaxWidth(1080).Padding(padding)),
	).Key("detail-scroll-" + item.key).Vertical()
}

func section(title string, body ui.Widget) ui.Widget {
	return surface.NewCard(layout.Column(
		text.New(title).Size(14).Weight(font.Bold),
		body,
	).Gap(20))
}

func muted(value string) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		active := ctx.Theme()
		return text.New(value).Size(12).Color(active.Semantic().MutedForeground).Layout(ctx, gtx)
	})
}

func fixedHeight(height unit.Dp, child ui.Widget) ui.Widget {
	return layout.Box(child).Style(style.New().FillWidth().Height(height))
}

func showcaseHeader(m Model, send ui.Send[Msg], compact bool) ui.Widget {
	brand := text.New("FlowUI").Size(18).Weight(font.Bold).MaxLines(1).Truncator("…")
	var identity ui.Widget = layout.Row(
		layout.Box(brand).Style(style.New().Width(252)),
		muted(m.text("组件 / ")+m.currentTitle()),
	).Gap(16).AlignMiddle()
	if compact {
		label, glyph := m.text("组件目录"), lucide.Menu
		if m.navOpen {
			label, glyph = m.text("返回演示"), lucide.ArrowLeft
		}
		identity = layout.Row(
			headerAction("catalog-toggle", label, glyph, func() { send(navMsg(!m.navOpen)) }),
			layout.Expanded(brand),
		).Gap(8).AlignMiddle()
	}
	themeLabel, themeIcon := m.text("切换深色"), lucide.Moon
	if m.dark {
		themeLabel, themeIcon = m.text("切换浅色"), lucide.Sun
	}
	return layout.Row(
		layout.Expanded(identity),
		layout.Row(
			languageMenu(m, send),
			headerAction("theme-toggle", themeLabel, themeIcon, func() { send(themeMsg{}) }),
			headerAction("demo-reset", m.text("重置演示"), lucide.RotateCcw, func() { send(resetMsg{}) }),
		).Gap(4).AlignMiddle(),
	).Gap(16).AlignMiddle().FillWidth().Padding(4, 16)
}

func (m Model) currentTitle() string {
	item, ok := findEntry(m.selected)
	if !ok {
		item = catalog[0]
	}
	return item.localized(m.demo).title
}

func headerAction(key, label string, glyph []byte, onClick func()) ui.Widget {
	trigger := button.New(label).Key(key).Content(icon.New(glyph).Size(18)).
		IconOnly().Size(button.Small).Variant(button.Ghost).OnClick(onClick)
	return tooltip.New(trigger, label).Key(key + "-hint").Placement(popover.Bottom).Offset(8)
}

func languageMenu(m Model, send ui.Send[Msg]) ui.Widget {
	label := "简体中文"
	if m.demo.language == window.LanguageEnglish {
		label = "English"
	}
	trigger := button.New(label).Key("language-trigger").Variant(button.Ghost).Size(button.Small).
		Content(layout.Row(icon.New(lucide.Languages).Size(16), text.New(label), icon.New(lucide.ChevronDown).Size(12)).Gap(8).AlignMiddle())
	items := []dropdown.Item{
		{Key: string(window.LanguageChinese), Label: "简体中文", IndicatorType: dropdown.IndicatorCheckmark,
			OnAction: func() { send(languageMsg(window.LanguageChinese)) }},
		{Key: string(window.LanguageEnglish), Label: "English", IndicatorType: dropdown.IndicatorCheckmark,
			OnAction: func() { send(languageMsg(window.LanguageEnglish)) }},
	}
	return dropdown.New(trigger, items).Key("showcase-language").
		Menu(menu.New(items).Key("showcase-language-menu").Width(176).Compact(true).
			SelectionMode(menu.SelectionSingle).SelectedKey(string(m.demo.language))).
		Placement(popover.BottomEnd).Offset(8)
}
