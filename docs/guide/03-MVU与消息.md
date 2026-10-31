# 03 · MVU 与消息

## 四个角色

| 角色 | 签名（示意） | 职责 |
|------|--------------|------|
| Model | 任意 struct | 业务状态 |
| Msg | 封闭类型集 | 发生了什么 |
| Update | `func(*M, Msg) Cmd[Msg]` | 改 Model，可选副作用 |
| View | `func(*Context, M, Send[Msg]) Widget` | 只读渲染 |

## 消息分发

小型 `Update` 直接用 Go 的类型 `switch`。当消息处理变多时，可以用
`ui.Match` 把每种消息拆成类型化 handler；两种写法都支持，按复杂度选择。

```go
func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	return ui.Match(m, msg,
		ui.On(func(m *Model, _ Increment) ui.Cmd[Msg] {
			m.Count++
			return nil
		}),
		ui.On(func(m *Model, msg NameChanged) ui.Cmd[Msg] {
			m.Name = msg.Name
			return nil
		}),
	)
}
```

`ui.On` 的 Model、消息和 Cmd 类型由 handler 签名推断。`ui.Match` 按顺序执行
第一个匹配项；没有匹配时返回 `nil`。handler 可以放在不同领域的函数中组合，
但消息类型仍由应用定义。需要查看所有处理分支时，普通 `switch` 依然直接明了。

## 封闭消息集

用私有标记方法可以限制消息只在当前包内实现，便于维护应用自己的消息集合：

```go
type Msg interface{ msg() }

type NameChanged struct{ Name string }
type Submitted struct{}

func (NameChanged) msg() {}
func (Submitted) msg() {}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Name
	case Submitted:
		// ...
	}
	return nil
}
```

Go 不会检查类型 `switch` 是否穷尽；增加新消息后，未处理的消息会走默认路径。
子模块消息可以包装进父级消息（见下文）。

## View 规则

```go
func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	// ✓ 读 m
	// ✓ 回调里 send(...)
	// ✗ m.Count++ 或 *m = ...
	// ✗ 在 View 里 go 启动改状态的 goroutine
	return text.New(m.Name)
}
```

回调里只表达意图：

```go
button.New("保存").OnClick(func() {
	send(Submitted{})
})
```

## 单窗口：`ui.Run(Program)`

```go
ui.Run(ui.NewProgram(Model{}, Update, View),
	window.Title("App"),
	window.Size(800, 600),
)
```

`NewProgram` 适合固定初始 Model。`Update` 始终返回 `Cmd`；没有副作用时返回 `nil`。

## 完整 Program

```go
ui.Run(ui.Program[Model, Msg]{
	Init: func() (Model, ui.Cmd[Msg]) {
		return Model{}, loadInitial()
	},
	Update: func(m *Model, msg Msg) ui.Cmd[Msg] {
		// 先改完 Model，再返回 Cmd
		return nil
	},
	Subscriptions: func(m Model) []ui.Subscription[Msg] {
		return nil
	},
	View: View,
	// 可选：把窗口尺寸/焦点映射成消息
	// WindowStateMessage: func(ui.WindowState) Msg { ... },
	// 可选：把原生关窗请求映射成消息（需配合 ui.NewApplication 使用）
	// CloseRequestMessage: func() Msg { return CloseRequested{} },
},
	window.Title("App"),
	window.Size(800, 600),
)
```

| 字段 | 说明 |
|------|------|
| `Init` | 每窗口实例一次；可带初始 Cmd |
| `Update` | 改 Model 后返回 Cmd |
| `Subscriptions` | 按当前 Model 声明订阅集合 |
| `View` | 渲染 |
| `WindowStateMessage` | 原生窗口状态 → Msg |
| `CloseRequestMessage` | 原生关窗请求 → Msg；窗口保持打开，确认后调 `Application.Close` |

异步细节见 [08-命令与订阅](08-命令与订阅.md)。

## 嵌套模块

子包可有自己的 `Model` / `Msg` / `Update` / `View`，父级包装消息：

```go
type CounterMsg struct{ Value counter.Msg }

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case CounterMsg:
		cmd := counter.Update(&model.Counter, msg.Value)
		return ui.MapCmd(cmd, func(child counter.Msg) Msg {
			return CounterMsg{Value: child}
		})
	}
	return nil
}
```

View 侧把子 `send` 映射上去：

```go
layout.Key("counter", counter.View(model.Counter, func(msg counter.Msg) {
	send(CounterMsg{Value: msg})
}))
```

基础消息循环的可运行示例见仓库 `examples/counter`，在仓库根目录运行 `go run ./examples/counter`。

## Action 与 Cmd

| 类型 | 用途 |
|------|------|
| **Cmd** | MVU 异步副作用（网络、IO、延时） |
| **Action** | 菜单项、工具栏、快捷键配方（`NewAction`、`ActionMenuItem`…） |

不要混名：`Command` 旧说法已拆成 **Cmd** 与 **Action**。

## 常见选项

```go
ui.Run(ui.NewProgram(Model{}, Update, View),
	window.Title("标题"),
	window.Size(960, 640),
	window.MinSize(400, 300),
	window.CustomizeTheme(func(t *theme.Theme) {
		// 改主题
	}),
	window.Locale(window.LanguageChinese),
	window.OnError(func(err error) {
		// 记录 EffectError / RuntimePanicError 等
		log.Println(err)
	}),
)
```

## 下一步

- [04-布局](04-布局.md)
- [08-命令与订阅](08-命令与订阅.md)
