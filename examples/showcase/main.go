// Showcase is an interactive catalog of FlowUI's public components.
package main

import (
	"image/color"
	"time"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tree"
	"github.com/qianniancn/flowui/widget/workbench"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	selected string
	query    string
	dark     bool
	navOpen  bool
	status   string
	demo     demoState
}

type demoState struct {
	overview                      overviewState
	language                      window.Language
	name, email, password, notes  string
	choice, combo, comboInput     string
	selectOpen, sectionsOpen      bool
	comboOpen                     bool
	customCombo, customComboInput string
	customComboOpen               bool
	selectExtraOpen               [3]bool
	selectRequired                string
	checkboxKeys                  []string
	checked, enabled, bold        bool
	expanded                      bool
	value, lower, upper           float64
	page, clicks                  int
	date                          time.Time
	dateRequired                  time.Time
	dateRange                     datepicker.DateRange
	color                         color.NRGBA
	selected, tab, panel          string
	tableSelected                 []string
	tableRowSelected              string
	tableSort                     table.SortDescriptor
	tableNameWidth                int
	tableTypeWidth                int
	tableLoadCount                int
	tableEmpty                    bool
	keys, expandedKeys            []string
	dialog, confirm, popup        bool
	portal                        bool
	toasts                        []toastNotice
	toastID                       int
	graph                         nodegraph.Graph
	edgeID                        int
	workbench                     workbench.State
	amount                        float64
	segment, location             string
	loading, palette              bool
	buttonLoading                 bool
	virtualSelected               string
	virtualSort                   table.SortDescriptor
	treeItems                     []tree.SimpleItem
	treeSelected, treeFilter      string
	treeExpanded, treeChecked     []string
	treeHalfChecked               []string
	treeVersion                   uint64
	animationExpanded             bool
	listGrouped                   string
	listOptional                  string
	navTabs                       []tabs.TabItem
	navTabSelected                string
	navTabID                      int
	navPage                       int
	sidebarSelected               string
	sidebarOpen                   []string
	sidebarCollapsed              bool
	sidebarHover                  bool
	timelineReverse               bool
	timelineLoading               bool
	collapseMultiple              []string
	treeMultiSelected             []string
	treeStrictChecked             []string
	treeRetryReady                bool
	virtualLogCount               int
	virtualTableSelected          []string
	quantity                      float64
	modalExample                  string
	alertHidden                   bool
	progressPending               bool
	popoverPlacement              string
	splitRatio                    float32
}

type toastNotice struct {
	key     string
	action  bool
	variant toast.Variant
}

func initialDemo(language window.Language) demoState {
	day := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.Local)
	m := demoState{
		overview: initialOverview(),
		language: language,
		name:     "FlowUI", email: "hello@flowui.dev", notes: "用 Go 构建桌面应用。",
		choice: "desktop", combo: "go", comboInput: "Go", checked: true, enabled: true,
		checkboxKeys: []string{"docs"},
		expanded:     true, value: 64, lower: 20, upper: 80, page: 1,
		date: day, dateRange: datepicker.DateRange{Start: day, End: day.AddDate(0, 0, 6)},
		color:    color.NRGBA{R: 62, G: 99, B: 221, A: 255},
		selected: "flowui", tab: "overview", panel: "preview", tableSelected: []string{"flowui", "docs"}, tableRowSelected: "flowui",
		tableNameWidth: 180, tableTypeWidth: 120,
		keys: []string{"go"}, expandedKeys: []string{"src", "widgets"},
		toasts: []toastNotice{}, workbench: initialWorkbench(),
		amount: 12.5, segment: "day", location: "components", loading: true,
		treeExpanded: []string{"src"},
		treeChecked:  []string{}, treeHalfChecked: []string{},
		listGrouped: "flowui", listOptional: "go", navPage: 50,
		navTabs: []tabs.TabItem{
			{Key: "document-1", Label: "main.go", Closable: true, Editable: true},
			{Key: "document-2", Label: "README.md", Closable: true, Editable: true},
		},
		navTabSelected: "document-1", navTabID: 2,
		sidebarSelected: "flowui", sidebarOpen: []string{"projects", "resources"},
		collapseMultiple:  []string{"go", "theme"},
		treeMultiSelected: []string{"main.go"}, treeStrictChecked: []string{"main.go"},
		virtualLogCount: 20, virtualTableSelected: []string{"grid-00001"},
		quantity: 3, splitRatio: 0.5,
	}
	m.notes = m.text(m.notes)
	m.graph = initialGraph()
	m.treeItems = initialExplorer()
	return m
}

func initialModel() Model {
	demo := initialDemo(window.LanguageEnglish)
	return Model{selected: "overview", status: demo.text("选择组件，试试它的交互。"), demo: demo}
}

type Msg interface{ msg() }
type searchMsg string
type selectMsg string
type logMsg string
type themeMsg struct{}
type languageMsg window.Language
type navMsg bool
type resetMsg struct{}
type demoMsg func(*demoState)

func (searchMsg) msg()   {}
func (selectMsg) msg()   {}
func (logMsg) msg()      {}
func (themeMsg) msg()    {}
func (languageMsg) msg() {}
func (navMsg) msg()      {}
func (resetMsg) msg()    {}
func (demoMsg) msg()     {}

func update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case searchMsg:
		m.query = string(msg)
	case selectMsg:
		if _, ok := findEntry(string(msg)); !ok {
			return nil
		}
		m.selected = string(msg)
		m.navOpen = false
		m.demo.dialog, m.demo.confirm, m.demo.popup, m.demo.portal = false, false, false, false
		m.demo.selectOpen, m.demo.sectionsOpen, m.demo.comboOpen, m.demo.customComboOpen = false, false, false, false
		m.demo.selectExtraOpen = [3]bool{}
		m.demo.palette = false
		m.demo.modalExample, m.demo.popoverPlacement = "", ""
		m.demo.overview.creating, m.demo.overview.frameworkOpen = false, false
		m.status = m.text("正在展示 ") + m.selected
	case themeMsg:
		m.dark = !m.dark
	case languageMsg:
		if m.demo.language == window.Language(msg) {
			return nil
		}
		m.demo.language = window.Language(msg)
		m.demo.treeVersion++
		m.status = m.text("正在展示 ") + m.selected
	case navMsg:
		m.navOpen = bool(msg)
	case logMsg:
		m.status = string(msg)
	case demoMsg:
		msg(&m.demo)
	case resetMsg:
		m.demo = initialDemo(m.demo.language)
		m.status = m.text("演示状态已重置")
	}
	return nil
}

func main() {
	app := ui.NewApplication()
	program := ui.NewProgram(initialModel(), func(m *Model, msg Msg) ui.Cmd[Msg] {
		cmd := update(m, msg)
		if _, ok := msg.(themeMsg); ok {
			active := theme.Light()
			if m.dark {
				active = theme.Dark()
			}
			app.SetTheme("main", active)
		}
		if _, ok := msg.(languageMsg); ok {
			app.SetLanguage("main", m.demo.language)
		}
		return cmd
	}, View)
	app.Run(ui.NewWindow("main", program,
		window.Title("FlowUI Showcase"),
		window.Size(1240, 860),
		window.Theme(theme.Light()),
		window.Locale(window.LanguageEnglish),
	))
}

func report(send ui.Send[Msg], action string) func() {
	return func() { send(logMsg(action)) }
}

func setString(send ui.Send[Msg], target func(*demoState) *string) func(string) {
	return func(value string) {
		send(demoMsg(func(m *demoState) { *target(m) = value }))
	}
}

func setBool(send ui.Send[Msg], target func(*demoState) *bool) func(bool) {
	return func(value bool) {
		send(demoMsg(func(m *demoState) { *target(m) = value }))
	}
}

func setNumber(send ui.Send[Msg]) func(float64) {
	return func(value float64) { send(demoMsg(func(m *demoState) { m.value = value })) }
}

func selection(m demoState, send ui.Send[Msg]) func(string) {
	return func(value string) {
		send(demoMsg(func(m *demoState) { m.selected = value }))
		send(logMsg(m.text("已选择：") + value))
	}
}
