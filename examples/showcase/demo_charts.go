package main

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/barchart"
	"github.com/qianniancn/flowui/widget/candlestick"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/gantt"
	"github.com/qianniancn/flowui/widget/heatmap"
	"github.com/qianniancn/flowui/widget/linechart"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/piechart"
)

func chartSelection(m demoState, send ui.Send[Msg]) func(chart.Selection) {
	return func(value chart.Selection) {
		label := value.Label
		if len(value.Items) > 0 {
			item := value.Items[0]
			label += fmt.Sprintf(" · %s：%.2f", item.SeriesLabel, item.Y)
		}
		send(logMsg(m.text("图表选择：") + label))
	}
}

func demoLine(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(linechart.New([]linechart.Series{
		linechart.Values("visits", m.text("访问量"), []float64{28, 42, 36, 58, 64, 52, 78}).Smooth(true).Area(true),
		linechart.Values("signups", m.text("注册量"), []float64{12, 18, 16, 28, 32, 26, 40}).Smooth(true),
	}).Key("line-demo").Categories([]string{m.text("周一"), m.text("周二"), m.text("周三"), m.text("周四"), m.text("周五"), m.text("周六"), m.text("周日")}).
		Height(300).Legend(true).OnDataClick(chartSelection(m, send)),
		field(m.text("阶梯线、虚线与缺失数据"), linechart.New([]linechart.Series{
			linechart.Values("step", m.text("计划"), []float64{20, 20, 45, 45, 60, 60, 80}).Step(linechart.StepEnd),
			linechart.Values("gaps", m.text("实际"), []float64{15, 28, math.NaN(), 38, 52, 48, 70}).
				LineStyle(linechart.LineDashed).ShowPoints(true),
		}).Key("line-steps").Height(260).Legend(true).OnDataClick(chartSelection(m, send))),
		field(m.text("空数据"), linechart.New(nil).Key("line-empty").Height(120).EmptyText(m.text("暂无数据"))),
	).Gap(24)
}

func demoBar(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(barchart.New([]barchart.Series{
		barchart.Values("planned", m.text("计划"), []float64{40, 60, 48, 72}),
		barchart.Values("actual", m.text("实际"), []float64{36, 64, 44, 68}),
	}).Key("bar-demo").Categories([]string{m.text("第一周"), m.text("第二周"), m.text("第三周"), m.text("第四周")}).
		Height(300).Legend(true).OnDataClick(chartSelection(m, send)),
		field(m.text("横向堆叠与数据标签"), barchart.New([]barchart.Series{
			barchart.Values("desktop", m.text("桌面"), []float64{40, 60, 48, 72}).Stack("platform").ShowLabels(true),
			barchart.Values("web", "Web", []float64{20, 32, 28, 36}).Stack("platform").ShowLabels(true),
		}).Key("bar-stacked").Categories([]string{"Go", "Rust", "TypeScript", "Python"}).Orientation(barchart.Horizontal).
			Height(280).Legend(true).OnDataClick(chartSelection(m, send))),
	).Gap(24)
}

func demoPie(m demoState, send ui.Send[Msg]) ui.Widget {
	data := []piechart.Data{
		piechart.Slice("desktop", m.text("桌面"), 56), piechart.Slice("web", "Web", 28),
		piechart.Slice("mobile", m.text("移动端"), 16),
	}
	return layout.Column(
		field(m.text("圆环图"), piechart.New(data).Key("pie-demo").InnerRadius(0.3).Height(300).
			Legend(true).OnDataClick(chartSelection(m, send))),
		field(m.text("实心饼图与扇区间距"), piechart.New(data).Key("pie-solid").PadAngle(3).Height(260).
			Legend(true).Labels(true).OnDataClick(chartSelection(m, send))),
		field(m.text("玫瑰图"), piechart.New(data).Key("pie-rose").RoseType(piechart.RoseArea).Height(260).
			Legend(true).OnDataClick(chartSelection(m, send))),
	).Gap(24)
}

func demoCandle(m demoState, send ui.Send[Msg]) ui.Widget {
	data := []candlestick.Candle{}
	for index := range 24 {
		open := 30 + float64(index)*0.6 + float64(index%3)*2
		close := open + float64(index%5) - 2
		data = append(data, candlestick.OHLC(open, close, min(open, close)-2, max(open, close)+3))
	}
	times := make([]time.Time, len(data))
	for index := range times {
		times[index] = time.Date(2026, time.October, index+1, 0, 0, 0, 0, time.Local)
	}
	return candlestick.New(data).Key("candle-demo").Times(times).FormatTime(func(value time.Time) string {
		return value.Format("01-02")
	}).YAxis(m.text("价格")).FormatY(func(value float64) string { return fmt.Sprintf("$%.0f", value) }).
		Crosshair(true).Height(300).OnDataClick(chartSelection(m, send))
}

func demoHeatmap(m demoState, send ui.Send[Msg]) ui.Widget {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 4, 0)
	values := []heatmap.CalendarValue{}
	for day := range 122 {
		values = append(values, heatmap.CalendarValue{Date: start.AddDate(0, 0, day), Value: float64(day * 7 % 13)})
	}
	return heatmap.New(start, end, values).Key("heatmap-demo").Height(240).OnDataClick(chartSelection(m, send))
}

func demoGantt(m demoState, send ui.Send[Msg]) ui.Widget {
	start := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local)
	return gantt.New([]gantt.Task{
		gantt.NewTask("design", m.text("设计"), start, start.AddDate(0, 0, 4)).Progress(1),
		gantt.NewTask("build", m.text("开发"), start.AddDate(0, 0, 3), start.AddDate(0, 0, 12)).Progress(0.65).DependsOn("design"),
		gantt.NewTask("test", m.text("测试"), start.AddDate(0, 0, 10), start.AddDate(0, 0, 16)).Progress(0.2).DependsOn("build"),
		gantt.NewTask("release", m.text("发布"), start.AddDate(0, 0, 16), start.AddDate(0, 0, 18)).DependsOn("test"),
		gantt.NewMilestone("launch", m.text("发布里程碑"), start.AddDate(0, 0, 18)).DependsOn("release"),
	}).Key("gantt-demo").Height(300).Legend(true).OnTaskClick(func(value gantt.Selection) {
		send(logMsg(fmt.Sprintf(m.text("%s · %s 至 %s · 进度 %.0f%%"),
			value.Label,
			value.Start.Format("01-02"),
			value.End.Format("01-02"),
			value.Progress*100,
		)))
	})
}

func initialGraph() nodegraph.Graph {
	return nodegraph.Graph{
		Nodes: []nodegraph.Node{
			nodegraph.NewNode("source", "数据源", nodegraph.Point{X: 20, Y: 40}).
				Outputs(nodegraph.NewPort("out", "数据")),
			nodegraph.NewNode("transform", "转换", nodegraph.Point{X: 250, Y: 120}).
				Inputs(nodegraph.NewPort("in", "输入")).Outputs(nodegraph.NewPort("out", "输出")),
			nodegraph.NewNode("result", "预览", nodegraph.Point{X: 480, Y: 40}).
				Inputs(nodegraph.NewPort("in", "结果")),
		},
		Edges: []nodegraph.Edge{
			nodegraph.NewEdge("source-transform", nodegraph.NewEndpoint("source", "out"), nodegraph.NewEndpoint("transform", "in")),
			nodegraph.NewEdge("transform-result", nodegraph.NewEndpoint("transform", "out"), nodegraph.NewEndpoint("result", "in")),
		},
	}
}

func demoGraph(m demoState, send ui.Send[Msg]) ui.Widget {
	graph := m.graph
	graph.Nodes = slices.Clone(graph.Nodes)
	for index := range graph.Nodes {
		node := &graph.Nodes[index]
		node.Title = m.text(node.Title)
		node.InputPorts = slices.Clone(node.InputPorts)
		node.OutputPorts = slices.Clone(node.OutputPorts)
		for port := range node.InputPorts {
			node.InputPorts[port].Label = m.text(node.InputPorts[port].Label)
		}
		for port := range node.OutputPorts {
			node.OutputPorts[port].Label = m.text(node.OutputPorts[port].Label)
		}
	}
	return nodegraph.New(graph).Key("graph-demo").Height(360).Minimap(true).
		OnNodesChange(func(changes []nodegraph.NodeChange) {
			send(demoMsg(func(m *demoState) { m.graph.Nodes = nodegraph.ApplyNodeChanges(m.graph.Nodes, changes) }))
		}).OnEdgesChange(func(changes []nodegraph.EdgeChange) {
		send(demoMsg(func(m *demoState) { m.graph.Edges = nodegraph.ApplyEdgeChanges(m.graph.Edges, changes) }))
	}).OnConnect(func(connection nodegraph.Connection) {
		send(demoMsg(func(m *demoState) {
			for _, edge := range m.graph.Edges {
				if edge.Source == connection.Source && edge.Target == connection.Target {
					return
				}
			}
			m.edgeID++
			id := fmt.Sprintf("edge-%d", m.edgeID)
			m.graph.Edges = append(m.graph.Edges, nodegraph.NewEdge(id, connection.Source, connection.Target))
		}))
	})
}
