package main

import (
	"image"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/window"
)

func TestButtonDemoRedrawLifecycle(t *testing.T) {
	harness := uitest.New(image.Pt(1240, 860), theme.Light())
	model := initialModel()
	model.selected = "button"
	for _, step := range []struct {
		name    string
		loading bool
	}{
		{name: "idle"},
		{name: "loading", loading: true},
		{name: "stopped"},
	} {
		t.Run(step.name, func(t *testing.T) {
			model.demo.buttonLoading = step.loading
			var wake bool
			for range 6 {
				harness.Advance(time.Second)
				harness.Frame(View(harness.Context(), model, func(Msg) {}))
				_, wake = harness.Router().WakeupTime()
			}
			if wake != step.loading {
				t.Fatalf("redraw requested=%v, want %v", wake, step.loading)
			}
		})
	}
}

func TestLanguageSwitchPreservesDemo(t *testing.T) {
	model := initialModel()
	if model.selected != "overview" {
		t.Fatalf("default page = %q, want overview", model.selected)
	}
	if model.demo.language != window.LanguageEnglish || model.status != "Choose a component and try its interactions." {
		t.Fatalf("unexpected default language or status: %s / %q", model.demo.language, model.status)
	}
	model.selected, model.query, model.dark = "input", "input", true
	model.demo.name, model.demo.notes, model.demo.value = "我的项目", "User notes", 37
	model.demo.graph.Nodes[0].Position.X = 140
	renameExplorer(&model.demo, "main", "renamed.go")
	for _, language := range []window.Language{window.LanguageEnglish, window.LanguageChinese} {
		update(&model, languageMsg(language))
		if model.demo.language != language || model.selected != "input" || model.query != "input" || !model.dark {
			t.Fatalf("language switch changed navigation or theme: %+v", model)
		}
		if model.demo.name != "我的项目" || model.demo.notes != "User notes" || model.demo.value != 37 ||
			model.demo.graph.Nodes[0].Position.X != 140 || model.demo.treeItems[1].Label != "renamed.go" {
			t.Fatal("language switch discarded demo edits")
		}
	}
	update(&model, languageMsg(window.LanguageEnglish))
	version := model.demo.treeVersion
	update(&model, languageMsg(window.LanguageEnglish))
	if model.demo.language != window.LanguageEnglish || model.demo.treeVersion != version {
		t.Fatal("selecting the current language changed state")
	}
	update(&model, resetMsg{})
	if model.demo.language != window.LanguageEnglish || model.demo.notes != "Build desktop apps with Go." {
		t.Fatalf("reset did not retain English: %+v", model.demo)
	}
}

func TestLocalizedCatalog(t *testing.T) {
	for _, language := range []window.Language{window.LanguageChinese, window.LanguageEnglish} {
		m := initialDemo(language)
		sections, count := filteredSections("", m)
		if count != len(catalog) {
			t.Fatalf("%s: catalog contains %d of %d demos", language, count, len(catalog))
		}
		want := "概览"
		if language == window.LanguageEnglish {
			want = "Overview"
		}
		if got := sections[0].Items[0].Label; got != want {
			t.Fatalf("%s: first navigation label = %q, want %q", language, got, want)
		}
		for _, query := range []string{"button", "按钮", "primary", "主要操作"} {
			results, _ := filteredSections(query, m)
			found := false
			for _, section := range results {
				for _, item := range section.Items {
					found = found || item.Key == "button"
				}
			}
			if !found {
				t.Fatalf("%s: bilingual search %q missed Button", language, query)
			}
		}
		for _, original := range catalog {
			item := original.localized(m)
			if strings.Contains(item.title, " · ") {
				t.Fatalf("%s: bilingual title remains: %q", language, item.title)
			}
			if language == window.LanguageEnglish && strings.ContainsFunc(item.title+item.summary+item.group+item.code,
				func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
				t.Fatalf("%s: untranslated metadata or code for %s", language, item.key)
			}
		}
		if componentDocumentation("button", language) == nil {
			t.Fatalf("%s: missing Button documentation", language)
		}
	}
}

func TestShowcaseToolbar(t *testing.T) {
	for _, size := range []image.Point{image.Pt(1240, 860), image.Pt(640, 860), image.Pt(360, 640)} {
		t.Run(size.String(), func(t *testing.T) {
			harness := uitest.New(size, theme.Light())
			model := initialModel()
			root := func() ui.Widget {
				core.ReplaceLanguage(harness.Context(), locale.Language(model.demo.language))
				return View(harness.Context(), model, func(msg Msg) { update(&model, msg) })
			}
			frame := func() {
				harness.Advance(time.Second)
				harness.Frame(root())
				harness.Advance(time.Second)
				harness.Frame(root())
			}
			click := func(label string) {
				t.Helper()
				node, ok := uitest.SemanticByLabel(harness.Semantics(), label)
				if !ok {
					t.Fatalf("missing %q control: %v", label, uitest.SemanticLabels(harness.Semantics()))
				}
				bounds := node.Desc.Bounds
				if bounds.Empty() || !bounds.In(image.Rectangle{Max: size}) {
					t.Fatalf("%q is outside the viewport: %v", label, bounds)
				}
				harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
				frame()
			}
			frame()
			click("English")
			click("简体中文")
			click("简体中文")
			click("English")
			if model.demo.language != window.LanguageEnglish || !slices.Contains(uitest.SemanticLabels(harness.Semantics()), "Build something native.") {
				t.Fatal("English menu choice did not translate the demo")
			}
			click("English")
			harness.Key(key.NameEscape, 0)
			frame()
			if slices.Contains(uitest.SemanticLabels(harness.Semantics()), "简体中文") {
				t.Fatal("Escape did not dismiss the language menu")
			}
			if size.X < 820 {
				click("Component catalog")
				if !model.navOpen {
					t.Fatal("catalog button did not open navigation")
				}
			}
			click("Switch to dark")
			if !model.dark {
				t.Fatal("theme button did not select dark mode")
			}
			click("English")
			click("简体中文")
			if model.demo.language != window.LanguageChinese {
				t.Fatal("Chinese menu choice did not switch language")
			}
			click("切换浅色")
			if model.dark {
				t.Fatal("theme button did not select light mode")
			}
			model.demo.name = "Edited"
			click("重置演示")
			if model.demo.name != "FlowUI" || model.demo.language != window.LanguageChinese {
				t.Fatal("reset did not restore the demo in the current language")
			}
			if model.navOpen {
				click("返回演示")
				if model.navOpen {
					t.Fatal("back button did not close navigation")
				}
			}
		})
	}
}

func TestLocalizedPagesLayout(t *testing.T) {
	for _, language := range []window.Language{window.LanguageChinese, window.LanguageEnglish} {
		for _, size := range []image.Point{image.Pt(1240, 860), image.Pt(640, 860)} {
			for name, active := range map[string]theme.Theme{"light": theme.Light(), "dark": theme.Dark()} {
				t.Run(string(language)+"/"+name+"/"+size.String(), func(t *testing.T) {
					harness := uitest.NewWithConfig(uitest.Config{Size: size, Theme: &active, Language: language})
					model := initialModel()
					model.demo = initialDemo(language)
					model.dark = name == "dark"
					for _, item := range catalog {
						t.Run(item.key, func(t *testing.T) {
							update(&model, selectMsg(item.key))
							for range 2 {
								dimensions := harness.Frame(View(harness.Context(), model, func(msg Msg) { update(&model, msg) }))
								if dimensions.Size != size {
									t.Fatalf("page size = %v, want %v", dimensions.Size, size)
								}
								if language == window.LanguageEnglish {
									for _, label := range uitest.SemanticLabels(harness.Semantics()) {
										if label != "中文" && strings.ContainsFunc(label, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
											t.Fatalf("untranslated English page label: %q", label)
										}
									}
								}
							}
						})
					}
				})
			}
		}
	}
}

func TestShowcaseHeaderHeight(t *testing.T) {
	for _, width := range []int{1240, 640, 360} {
		for _, language := range []window.Language{window.LanguageChinese, window.LanguageEnglish} {
			harness := uitest.New(image.Pt(width, 100), theme.Light())
			model := initialModel()
			model.demo = initialDemo(language)
			dims := harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
				gtx.Constraints.Min.Y = 0
				return showcaseHeader(model, func(Msg) {}, width < 820).Layout(ctx, gtx)
			}))
			if dims.Size.X != width || dims.Size.Y > 36 {
				t.Fatalf("%s/%d: toolbar size = %v, want full width and height <= 36dp", language, width, dims.Size)
			}
		}
	}
}

func TestCheckboxDemoSelectAll(t *testing.T) {
	harness := uitest.New(image.Pt(640, 720), theme.Light())
	model := initialModel()
	model.demo = initialDemo(window.LanguageChinese)
	root := func() ui.Widget { return demoCheckbox(model.demo, func(msg Msg) { update(&model, msg) }) }
	frame := func() { harness.Frame(root()); harness.Frame(root()) }
	click := func(label string) {
		t.Helper()
		node, ok := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
			return node.Desc.Label == label && node.Desc.Class == semantic.CheckBox
		})
		if !ok {
			t.Fatalf("missing checkbox %q", label)
		}
		bounds := node.Desc.Bounds
		harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		frame()
	}
	frame()
	click("全选内容")
	if !slices.Equal(model.demo.checkboxKeys, []string{"docs", "examples", "tests"}) {
		t.Fatalf("select all = %v", model.demo.checkboxKeys)
	}
	click("示例")
	if !slices.Equal(model.demo.checkboxKeys, []string{"docs", "tests"}) {
		t.Fatalf("deselect example = %v", model.demo.checkboxKeys)
	}
	click("全选内容")
	click("全选内容")
	if len(model.demo.checkboxKeys) != 0 {
		t.Fatalf("clear all = %v", model.demo.checkboxKeys)
	}
}

func TestCustomComboDemoInput(t *testing.T) {
	harness := uitest.New(image.Pt(640, 640), theme.Light())
	model := initialModel()
	model.demo = initialDemo(window.LanguageChinese)
	root := func() ui.Widget { return demoCombo(model.demo, func(msg Msg) { update(&model, msg) }) }
	harness.Frame(root())
	tag := core.FieldFocusTag(harness.Context(), "combo-custom-input")
	if tag == nil {
		t.Fatal("custom combobox has no editor")
	}
	harness.Router().Source().Execute(key.FocusCmd{Tag: tag})
	harness.Frame(root())
	harness.Router().Queue(key.EditEvent{Range: key.Range{}, Text: "Zig"})
	harness.Frame(root())
	harness.Frame(root())
	if model.demo.customCombo != "Zig" || model.demo.customComboInput != "Zig" || !model.demo.customComboOpen {
		t.Fatalf("custom value=%q input=%q open=%v", model.demo.customCombo, model.demo.customComboInput, model.demo.customComboOpen)
	}
	if model.demo.combo != "go" || model.demo.comboInput != "Go" {
		t.Fatal("custom input changed the other combobox")
	}
	if _, ok := uitest.SemanticByLabel(harness.Semantics(), "没有匹配项，正在使用输入值"); !ok {
		t.Fatal("custom input did not display its empty-results message")
	}
}

func TestRequiredSelectDemo(t *testing.T) {
	harness := uitest.New(image.Pt(640, 720), theme.Light())
	model := initialModel()
	model.demo = initialDemo(window.LanguageChinese)
	root := func() ui.Widget { return demoSelect(model.demo, func(msg Msg) { update(&model, msg) }) }
	frame := func() { harness.Advance(time.Second); harness.Frame(root()); harness.Frame(root()) }
	click := func(label string) {
		t.Helper()
		node, ok := uitest.SemanticByLabel(harness.Semantics(), label)
		if !ok {
			t.Fatalf("missing select control %q", label)
		}
		bounds := node.Desc.Bounds
		harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		frame()
	}
	frame()
	click("请选择语言")
	if !model.demo.selectExtraOpen[1] {
		t.Fatal("required select did not open")
	}
	click("Rust")
	if model.demo.selectRequired != "rust" || model.demo.selectExtraOpen[1] {
		t.Fatalf("required selection = %q, open=%v", model.demo.selectRequired, model.demo.selectExtraOpen[1])
	}
	if uitest.HasSemanticLabel(harness.Semantics(), "此项为必选项") {
		t.Fatal("required select still displays an error after selection")
	}
}

func TestComboDemoOpensPopup(t *testing.T) {
	harness := uitest.New(image.Pt(640, 360), theme.Light())
	model := initialModel()
	root := func() ui.Widget {
		return demoCombo(model.demo, func(msg Msg) { update(&model, msg) })
	}

	harness.Frame(root())
	if node, ok := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
		return node.Desc.Class == semantic.Editor
	}); !ok || node.Desc.Bounds.Dx() <= 0 || node.Desc.Bounds.Dy() <= 0 {
		t.Fatalf("missing visible combobox editor: found=%v bounds=%v", ok, node.Desc.Bounds)
	}
	node, ok := uitest.SemanticByLabel(harness.Semantics(), "Open options")
	if !ok {
		t.Fatalf("missing combobox trigger: %v", uitest.SemanticLabels(harness.Semantics()))
	}
	bounds := node.Desc.Bounds
	harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	harness.Frame(root())
	harness.Frame(root())
	if !model.demo.comboOpen {
		t.Fatal("combobox demo did not open")
	}
	if _, ok := uitest.SemanticByLabel(harness.Semantics(), "Rust"); !ok {
		t.Fatalf("missing combobox option: %v", uitest.SemanticLabels(harness.Semantics()))
	}
}

func TestSelectDemoOpensPopup(t *testing.T) {
	harness := uitest.New(image.Pt(640, 360), theme.Light())
	model := initialModel()
	root := func() ui.Widget {
		return demoSelect(model.demo, func(msg Msg) { update(&model, msg) })
	}

	harness.Frame(root())
	node, ok := uitest.SemanticByLabel(harness.Semantics(), "Go")
	if !ok {
		t.Fatalf("missing select trigger: %v", uitest.SemanticLabels(harness.Semantics()))
	}
	bounds := node.Desc.Bounds
	harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	harness.Frame(root())
	harness.Frame(root())
	if !model.demo.selectOpen {
		t.Fatal("select demo did not open")
	}
	if _, ok := uitest.SemanticByLabel(harness.Semantics(), "Rust"); !ok {
		t.Fatalf("missing select option: %v", uitest.SemanticLabels(harness.Semantics()))
	}
}

func TestTabsDemoDocuments(t *testing.T) {
	harness := uitest.New(image.Pt(900, 2300), theme.Light())
	model := initialModel()
	root := func() ui.Widget { return demoTabs(model.demo, func(msg Msg) { update(&model, msg) }) }
	frame := func() { harness.Advance(time.Second); harness.Frame(root()); harness.Frame(root()) }
	click := func(label string) {
		t.Helper()
		node, ok := uitest.SemanticByLabel(harness.Semantics(), label)
		if !ok {
			t.Fatalf("missing tab control %q", label)
		}
		bounds := node.Desc.Bounds
		harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		frame()
	}
	frame()
	click("Add tab")
	if len(model.demo.navTabs) != 3 || model.demo.navTabSelected != "document-3" {
		t.Fatalf("add: tabs=%v selected=%q", model.demo.navTabs, model.demo.navTabSelected)
	}
	click("untitled-3.go")
	harness.Key(key.NameF2, 0)
	frame()
	if _, ok := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
		return node.Desc.Class == semantic.Editor
	}); !ok {
		t.Fatal("F2 did not open the document name editor")
	}
	harness.Router().Queue(key.EditEvent{Range: key.Range{Start: 0, End: len("untitled-3.go")}, Text: "renamed.go"})
	frame()
	harness.Key(key.NameReturn, 0)
	frame()
	if model.demo.navTabs[2].Label != "renamed.go" {
		t.Fatalf("renamed tab = %q", model.demo.navTabs[2].Label)
	}
	for _, label := range []string{"renamed.go", "README.md", "main.go"} {
		click(label)
		node, _ := uitest.SemanticByLabel(harness.Semantics(), label)
		bounds := node.Desc.Bounds
		harness.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse,
			Position: f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2)})
		frame()
		click("Close " + label)
	}
	if len(model.demo.navTabs) != 0 || model.demo.navTabSelected != "" {
		t.Fatalf("close all: tabs=%v selected=%q", model.demo.navTabs, model.demo.navTabSelected)
	}
	click("Add tab")
	if len(model.demo.navTabs) != 1 || model.demo.navTabSelected != "document-4" {
		t.Fatalf("reopen: tabs=%v selected=%q", model.demo.navTabs, model.demo.navTabSelected)
	}
}

func TestSidebarDemoCollapsedNavigation(t *testing.T) {
	harness := uitest.New(image.Pt(900, 1500), theme.Light())
	model := initialModel()
	model.demo = initialDemo(window.LanguageChinese)
	root := func() ui.Widget { return demoSidebar(model.demo, func(msg Msg) { update(&model, msg) }) }
	frame := func() { harness.Advance(time.Second); harness.Frame(root()); harness.Frame(root()) }
	click := func(label string) {
		t.Helper()
		node, ok := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
			return node.Desc.Label == label && node.Desc.Class != semantic.Unknown
		})
		if !ok {
			t.Fatalf("missing navigation control %q", label)
		}
		bounds := node.Desc.Bounds
		harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
		frame()
	}
	frame()
	click("API")
	if model.demo.sidebarSelected != "api" {
		t.Fatalf("nested selection = %q", model.demo.sidebarSelected)
	}
	click("折叠为图标导航")
	frame()
	if !model.demo.sidebarCollapsed {
		t.Fatal("collapse switch did not update navigation")
	}
	click("概览")
	if model.demo.sidebarSelected != "overview" {
		t.Fatalf("collapsed selection = %q", model.demo.sidebarSelected)
	}
	click("折叠为图标导航")
	frame()
	click("使用指南")
	if model.demo.sidebarSelected != "guide" {
		t.Fatalf("restored nested selection = %q", model.demo.sidebarSelected)
	}
}

func TestAdvancedTreeDemoRetry(t *testing.T) {
	harness := uitest.New(image.Pt(760, 900), theme.Light())
	model := initialModel()
	model.demo = initialDemo(window.LanguageChinese)
	root := func() ui.Widget { return demoAdvancedTree(model.demo, func(msg Msg) { update(&model, msg) }) }
	frame := func() { harness.Advance(time.Second); harness.Frame(root()); harness.Frame(root()) }
	frame()
	node, ok := uitest.SemanticByLabel(harness.Semantics(), "加载失败的目录")
	if !ok {
		t.Fatal("missing failed folder")
	}
	bounds := node.Desc.Bounds
	harness.Click(f32.Pt(float32(bounds.Min.X-18), float32(bounds.Min.Y+bounds.Max.Y)/2))
	frame()
	if !model.demo.treeRetryReady || !uitest.HasSemanticLabel(harness.Semantics(), "recovered.md") {
		t.Fatal("retry did not load the recovered child")
	}
	node, ok = uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
		return node.Desc.Label == "恢复失败状态" && node.Desc.Class == semantic.Button
	})
	if !ok {
		t.Fatal("missing restore button")
	}
	bounds = node.Desc.Bounds
	harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	frame()
	if model.demo.treeRetryReady || uitest.HasSemanticLabel(harness.Semantics(), "recovered.md") {
		t.Fatal("restore did not return the folder to its error state")
	}
}

func TestSupplementalDemoInteractions(t *testing.T) {
	for _, language := range []window.Language{window.LanguageEnglish, window.LanguageChinese} {
		t.Run(string(language), func(t *testing.T) {
			active := theme.Light()
			harness := uitest.NewWithConfig(uitest.Config{
				Size: image.Pt(760, 1200), Theme: &active, Language: language,
			})
			model := initialModel()
			model.demo = initialDemo(language)
			var demo func(demoState, ui.Send[Msg]) ui.Widget
			frame := func() {
				for range 3 {
					harness.Advance(time.Second)
					harness.Frame(demo(model.demo, func(msg Msg) { update(&model, msg) }))
				}
			}
			click := func(label string) {
				t.Helper()
				node, ok := uitest.FindSemantic(harness.Semantics(), func(node uitest.Semantic) bool {
					return node.Desc.Label == label && node.Desc.Class == semantic.Button
				})
				if !ok {
					t.Fatalf("missing button %q: %v", label, uitest.SemanticLabels(harness.Semantics()))
				}
				bounds := node.Desc.Bounds
				harness.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
				frame()
			}
			demo = demoNumericInput
			frame()
			click(model.text("设为最大值"))
			click(model.text("数量") + " +")
			if model.demo.quantity != 10 {
				t.Fatalf("increment exceeded maximum: %v", model.demo.quantity)
			}
			click(model.text("数量") + " −")
			if model.demo.quantity != 9 || model.demo.amount != 12.5 {
				t.Fatalf("quantity decrement changed unexpected state: %v / %v", model.demo.quantity, model.demo.amount)
			}
			click(model.text("设为最小值"))
			click(model.text("数量") + " −")
			if model.demo.quantity != 0 {
				t.Fatalf("decrement exceeded minimum: %v", model.demo.quantity)
			}

			demo = demoAlert
			frame()
			closeLabel := "关闭"
			if language == window.LanguageEnglish {
				closeLabel = "Close"
			}
			click(closeLabel)
			if !model.demo.alertHidden {
				t.Fatal("alert close did not hide the alert")
			}
			click(model.text("恢复提示"))
			if model.demo.alertHidden {
				t.Fatal("restore did not show the alert")
			}

			demo = demoModal
			frame()
			click(model.text("必须明确关闭"))
			harness.Key(key.NameEscape, 0)
			frame()
			harness.Click(f32.Pt(5, 5))
			frame()
			if model.demo.modalExample != "explicit" {
				t.Fatal("explicit dialog closed on Escape or backdrop")
			}
			click(model.text("我已了解"))
			if model.demo.modalExample != "" {
				t.Fatal("explicit dialog did not close on confirmation")
			}
			click(model.text("长内容弹窗"))
			harness.Key(key.NameEscape, 0)
			frame()
			if model.demo.modalExample != "" {
				t.Fatal("scrollable dialog did not close on Escape")
			}

			demo = demoToolbar
			frame()
			moreLabel := "更多"
			if language == window.LanguageEnglish {
				moreLabel = "More"
			}
			click(moreLabel)
			click(model.text("导出"))
			if model.status != model.text("已选择导出 JSON") {
				t.Fatalf("overflow command did not execute: %q", model.status)
			}
		})
	}
}
