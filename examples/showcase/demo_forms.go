package main

import (
	"fmt"
	"image/color"
	"slices"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/colorpicker"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/divider"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/radiogroup"
	"github.com/qianniancn/flowui/widget/selector"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toggle"
)

func demoInput(m demoState, send ui.Send[Msg]) ui.Widget {
	nameChanged := setString(send, func(m *demoState) *string { return &m.name })
	return layout.Column(
		field(m.text("项目名称"), input.New(m.name).Key("input-name").Label(m.text("项目名称输入框")).Hint(m.text("输入项目名称")).
			FillWidth().OnChange(nameChanged).OnSubmit(func(value string) { send(logMsg(m.text("已提交：") + value)) })),
		field(m.text("邮箱"), input.New(m.email).Key("input-email").Label(m.text("邮箱")).Type(input.Email).FillWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.email }))),
		field(m.text("密码"), input.New(m.password).Key("input-password").Label(m.text("密码")).Type(input.Password).
			Hint(m.text("输入密码")).FillWidth().OnChange(setString(send, func(m *demoState) *string { return &m.password }))),
		field(m.text("多行输入"), input.NewTextArea(m.notes).Key("input-notes").Label(m.text("备注")).Rows(3).FillWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.notes }))),
		field(m.text("前后缀与操作"), input.Group(input.New(m.name).Key("input-group").OnChange(nameChanged)).
			Prefix(text.New(m.text("项目 / "))).Suffix(text.New(".go")).FillWidth().
			SuffixAction(input.GroupAction("clear-name", m.text("清空")).OnClick(func() { nameChanged("") }))),
		input.GroupTextArea(input.NewTextArea(m.notes).Key("notes-group").Rows(2).
			OnChange(setString(send, func(m *demoState) *string { return &m.notes }))).
			Prefix(text.New(m.text("备注"))).FillWidth(),
		input.New("name@example").Invalid(true).FillWidth(),
		input.New(m.text("只读内容")).ReadOnly(true).FillWidth(),
		input.New(m.text("禁用内容")).Disabled(true).FillWidth(),
		field(m.text("外观变体"), layout.Column(
			input.New(m.name).Key("input-primary").Label("Primary").Variant(input.Primary).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-secondary").Label("Secondary").Variant(input.Secondary).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-filled").Label("Filled").Variant(input.Filled).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-borderless").Label("Borderless").Variant(input.Borderless).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-underlined").Label("Underlined").Variant(input.Underlined).FillWidth().OnChange(nameChanged),
		).Gap(12)),
		field(m.text("尺寸对比"), layout.Column(
			input.New(m.name).Key("input-small").Label("Small").Size(input.Small).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-medium").Label("Medium").Size(input.Medium).FillWidth().OnChange(nameChanged),
			input.New(m.name).Key("input-large").Label("Large").Size(input.Large).FillWidth().OnChange(nameChanged),
		).Gap(12)),
	).Gap(16)
}

func field(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(14), child).Gap(6)
}

func demoCheckbox(m demoState, send ui.Send[Msg]) ui.Widget {
	options := []struct{ key, label string }{
		{key: "docs", label: "文档"}, {key: "examples", label: "示例"}, {key: "tests", label: "测试"},
	}
	children := []ui.Widget{
		toggle.NewCheckbox(m.text("全选内容")).Key("checkbox-all").Checked(len(m.checkboxKeys) == len(options)).
			Indeterminate(len(m.checkboxKeys) > 0 && len(m.checkboxKeys) < len(options)).OnChange(func(checked bool) {
			send(demoMsg(func(m *demoState) {
				m.checkboxKeys = []string{}
				if checked {
					for _, option := range options {
						m.checkboxKeys = append(m.checkboxKeys, option.key)
					}
				}
			}))
		}),
	}
	for _, option := range options {
		children = append(children, toggle.NewCheckbox(m.text(option.label)).Key("checkbox-option-"+option.key).
			Checked(slices.Contains(m.checkboxKeys, option.key)).OnChange(func(checked bool) {
			send(demoMsg(func(m *demoState) {
				keys := slices.Clone(m.checkboxKeys)
				if checked {
					keys = append(keys, option.key)
				} else {
					keys = slices.DeleteFunc(keys, func(key string) bool { return key == option.key })
				}
				m.checkboxKeys = keys
			}))
		}))
	}
	return layout.Column(
		toggle.NewCheckbox(m.text("接收项目更新")).Key("checkbox-main").Checked(m.checked).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.checked })),
		toggle.NewCheckbox(m.text("部分选中")).Key("checkbox-partial").Checked(false).Indeterminate(true),
		toggle.NewCheckbox(m.text("已选中 · 禁用")).Key("checkbox-disabled").Checked(true).Disabled(true),
		text.New(fmt.Sprintf(m.text("接收更新：%t"), m.checked)),
		field(m.text("全选与部分选中"), layout.Column(children...).Gap(12)),
		field(m.text("变体与校验"), layout.Column(
			toggle.NewCheckbox(m.text("次级外观")).Key("checkbox-secondary").Variant(toggle.CheckboxSecondary).
				Checked(m.checked).OnChange(setBool(send, func(m *demoState) *bool { return &m.checked })),
			toggle.NewCheckbox(m.text("只读选项")).Key("checkbox-readonly").Checked(true).ReadOnly(true),
			toggle.NewCheckbox(m.text("同意使用条款")).Key("checkbox-required").Checked(m.checked).Required(true).
				Invalid(!m.checked).ErrorMessage(m.text("请先同意使用条款")).
				Description(m.text("取消勾选可查看校验提示")).
				OnChange(setBool(send, func(m *demoState) *bool { return &m.checked })),
		).Gap(12)),
	).Gap(16)
}

func demoSwitch(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		toggle.NewSwitch(m.text("自动同步")).Key("switch-sync").Checked(m.enabled).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
		toggle.NewSwitch(m.text("不可用的设置")).Key("switch-disabled").Checked(false).Disabled(true),
		text.New(fmt.Sprintf(m.text("自动同步：%t"), m.enabled)),
		field(m.text("尺寸对比"), layout.Column(
			toggle.NewSwitch("Small").Key("switch-small").Size(toggle.SwitchSmall).Checked(m.enabled).
				OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
			toggle.NewSwitch("Medium").Key("switch-medium").Checked(m.enabled).
				OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
			toggle.NewSwitch("Large").Key("switch-large").Size(toggle.SwitchLarge).Checked(m.enabled).
				OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
		).Gap(12)),
		toggle.NewSwitch(m.text("标签在前")).Key("switch-before").LabelBefore().Checked(m.enabled).
			Description(m.text("修改任一开关会同步更新此页的可用开关")).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
		toggle.NewSwitch(m.text("已开启 · 禁用")).Key("switch-disabled-on").Checked(true).Disabled(true),
	).Gap(16)
}

func demoRadio(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []radiogroup.RadioItem{
		{Key: "desktop", Label: m.text("桌面应用")}, {Key: "web", Label: m.text("Web 应用")},
		{Key: "mobile", Label: m.text("移动应用"), Disabled: true},
	}
	changed := setString(send, func(m *demoState) *string { return &m.choice })
	return layout.Column(
		radiogroup.New(items).Key("radio-vertical").SelectedKey(m.choice).OnChange(changed),
		divider.New(),
		radiogroup.New(items).Key("radio-horizontal").SelectedKey(m.choice).Horizontal().OnChange(changed),
		text.New(m.text("当前选择：")+m.choice),
		field(m.text("次级外观"), radiogroup.New(items).Key("radio-secondary").Variant(radiogroup.RadioSecondary).
			SelectedKey(m.choice).OnChange(changed)),
		field(m.text("校验状态"), radiogroup.New(items).Key("radio-invalid").Invalid(true).
			SelectedKey(m.choice).OnChange(changed)),
		field(m.text("整组禁用"), radiogroup.New(items).Key("radio-disabled").Disabled(true).SelectedKey("desktop")),
	).Gap(20)
}

func languageItems(m demoState) []selector.SelectItem {
	return []selector.SelectItem{
		{Key: "go", Label: "Go", Description: m.text("原生桌面开发")},
		{Key: "rust", Label: "Rust", Description: m.text("系统开发")},
		{Key: "typescript", Label: "TypeScript", Description: m.text("Web 开发")},
		{Key: "python", Label: "Python", Description: m.text("数据分析")},
		{Key: "legacy", Label: m.text("旧版语言（禁用）"), Disabled: true},
	}
}

func demoSelect(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := setString(send, func(m *demoState) *string { return &m.combo })
	selectOpenChanged := setBool(send, func(m *demoState) *bool { return &m.selectOpen })
	sectionsOpenChanged := setBool(send, func(m *demoState) *bool { return &m.sectionsOpen })
	return layout.Column(
		field(m.text("单选"), selector.New(languageItems(m)).Key("select-single").SelectedKey(m.combo).
			Open(m.selectOpen).OnOpenChange(selectOpenChanged).FillWidth().OnChange(changed)),
		field(m.text("分组选项"), selector.NewSelectSections([]selector.SelectSection{
			{Title: m.text("编译型"), Items: languageItems(m)[:2]},
			{Title: m.text("脚本与应用"), Items: languageItems(m)[2:]},
		}).Key("select-sections").SelectedKey(m.combo).Open(m.sectionsOpen).
			OnOpenChange(sectionsOpenChanged).FillWidth().OnChange(changed)),
		field(m.text("次级外观"), selector.New(languageItems(m)).Key("select-secondary").Variant(selector.SelectSecondary).
			SelectedKey(m.combo).Open(m.selectExtraOpen[0]).FillWidth().OnChange(changed).
			OnOpenChange(func(open bool) { send(demoMsg(func(m *demoState) { m.selectExtraOpen[0] = open })) })),
		field(m.text("占位与校验"), selector.New(languageItems(m)).Key("select-invalid").SelectedKey(m.selectRequired).
			Placeholder(m.text("请选择语言")).Invalid(m.selectRequired == "").ErrorMessage(m.text("此项为必选项")).
			Open(m.selectExtraOpen[1]).FillWidth().OnChange(setString(send, func(m *demoState) *string { return &m.selectRequired })).
			OnOpenChange(func(open bool) { send(demoMsg(func(m *demoState) { m.selectExtraOpen[1] = open })) })),
		field(m.text("空列表"), selector.New(nil).Key("select-empty").Placeholder(m.text("暂无可选语言")).
			Open(m.selectExtraOpen[2]).FillWidth().
			OnOpenChange(func(open bool) { send(demoMsg(func(m *demoState) { m.selectExtraOpen[2] = open })) })),
		field(m.text("禁用状态"), selector.New(languageItems(m)).Key("select-disabled").SelectedKey(m.combo).
			Disabled(true).FillWidth()),
	).Gap(20)
}

func demoCombo(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []combobox.ComboboxItem{}
	for _, item := range languageItems(m) {
		items = append(items, combobox.ComboboxItem{
			Key: item.Key, Label: item.Label, Description: item.Description, Disabled: item.Disabled,
		})
	}
	return layout.Column(
		combobox.New(items).Key("combo-language").SelectedKey(m.combo).
			InputValue(m.comboInput).Hint(m.text("搜索语言，例如 Go")).Open(m.comboOpen).
			OnOpenChange(setBool(send, func(m *demoState) *bool { return &m.comboOpen })).FillWidth().
			OnInputChange(setString(send, func(m *demoState) *string { return &m.comboInput })).
			OnChange(func(value string) {
				send(demoMsg(func(m *demoState) {
					m.combo = value
					for _, item := range items {
						if item.Key == value {
							m.comboInput = item.Label
							break
						}
					}
				}))
			}),
		text.New(m.text("选中值：")+m.combo),
		field(m.text("允许自定义值"), combobox.New(items).Key("combo-custom").AllowCustomValue().Variant(input.Secondary).
			SelectedKey(m.customCombo).InputValue(m.customComboInput).Open(m.customComboOpen).
			Hint(m.text("输入内容可直接作为自定义值")).EmptyText(m.text("没有匹配项，正在使用输入值")).FillWidth().
			OnInputChange(setString(send, func(m *demoState) *string { return &m.customComboInput })).
			OnOpenChange(setBool(send, func(m *demoState) *bool { return &m.customComboOpen })).
			OnChange(func(value string) {
				send(demoMsg(func(m *demoState) {
					m.customCombo, m.customComboInput = value, value
					for _, item := range items {
						if item.Key == value {
							m.customComboInput = item.Label
						}
					}
				}))
			})),
		text.New(m.text("自定义值：")+m.customCombo),
		field(m.text("禁用状态"), combobox.New(items).Key("combo-disabled").SelectedKey("go").InputValue("Go").Disabled(true).FillWidth()),
	).Gap(16)
}

func demoDate(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := func(value time.Time) { send(demoMsg(func(m *demoState) { m.date = value })) }
	return layout.Column(
		field(m.text("日历选择"), datepicker.New("date-picker", m.date).Label(m.text("日期")).
			OnChange(changed)),
		field(m.text("可编辑日期"), datepicker.DateField("date-field", m.date).Label(m.text("日期字段")).OnChange(changed)),
		field(m.text("日期范围"), datepicker.DateRangePicker("date-range", m.dateRange).Label(m.text("计划周期")).
			OnChange(func(value datepicker.DateRange) {
				send(demoMsg(func(m *demoState) { m.dateRange = value }))
			})),
		text.New(m.text("所选日期：")+m.date.Format("2006-01-02")),
		field(m.text("日期边界"), datepicker.New("date-bounded", m.date).Label(m.text("仅限当前月份")).
			MinDate(time.Date(m.date.Year(), m.date.Month(), 1, 0, 0, 0, 0, m.date.Location())).
			MaxDate(time.Date(m.date.Year(), m.date.Month()+1, 0, 0, 0, 0, 0, m.date.Location())).OnChange(changed)),
		field(m.text("校验状态"), datepicker.New("date-required", m.dateRequired).Label(m.text("必填日期")).
			Required(true).Invalid(m.dateRequired.IsZero()).ErrorMessage(m.text("请选择日期")).
			OnChange(func(value time.Time) { send(demoMsg(func(m *demoState) { m.dateRequired = value })) })),
		field(m.text("禁用状态"), datepicker.New("date-disabled", m.date).Disabled(true)),
	).Gap(20)
}

func demoColor(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := func(value color.NRGBA) { send(demoMsg(func(m *demoState) { m.color = value })) }
	colors := []color.NRGBA{
		{R: 62, G: 99, B: 221, A: 255}, {R: 48, G: 164, B: 108, A: 255},
		{R: 229, G: 72, B: 77, A: 255}, {R: 245, G: 164, B: 36, A: 255},
	}
	return layout.Column(
		layout.Wrap(
			colorpicker.New("color-picker", m.color).ShowRGB(true).ShowHistory(true).OnChange(changed),
			colorpicker.ColorSwatch(m.color),
			text.New(fmt.Sprintf("#%02X%02X%02X", m.color.R, m.color.G, m.color.B)),
		).Gap(12).AlignMiddle(),
		layout.Wrap(
			colorpicker.ColorArea("color-area", m.color).OnChange(changed),
			colorpicker.ColorWheel("color-wheel", m.color).Size(180).OnChange(changed),
		).Gap(20).LineGap(16),
		colorpicker.ColorSlider("color-hue", m.color, colorpicker.ColorChannelHue).OnChange(changed),
		colorpicker.ColorSlider("color-alpha", m.color, colorpicker.ColorChannelAlpha).OnChange(changed),
		colorpicker.ColorField("color-field", m.color).OnChange(changed),
		colorpicker.ColorSwatchPicker("color-swatches", m.color, colors).OnChange(changed),
	).Gap(20)
}

func demoSlider(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		slider.New("slider-main", m.value).Label(m.text("音量")).Range(0, 100).Step(1).ShowValue().OnChange(setNumber(send)),
		slider.RangeSlider("slider-range", m.lower, m.upper).Label(m.text("价格范围")).Range(0, 100).
			OnRangeChange(func(lower, upper float64) {
				send(demoMsg(func(m *demoState) { m.lower, m.upper = lower, upper }))
			}),
		text.New(fmt.Sprintf(m.text("音量 %.0f · 范围 %.0f–%.0f"), m.value, m.lower, m.upper)),
		slider.New("slider-disabled", 40).Label(m.text("不可用")).Disabled(true),
		slider.New("slider-step", m.value).Label(m.text("步进与格式化")).Range(0, 100).Step(10).ShowValue().
			FormatValue(func(value float64) string { return fmt.Sprintf("%.0f%%", value) }).OnChange(setNumber(send)),
		field(m.text("垂直滑块"), layout.Row(
			layout.Column(slider.New("slider-vertical", m.value).Vertical().Range(0, 100).Step(1).
				OnChange(setNumber(send))).Width(16).Height(180),
			text.New(fmt.Sprintf("%.0f", m.value)),
		).Gap(16)),
	).Gap(24)
}
