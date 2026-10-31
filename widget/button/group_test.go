package button

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func TestGroupCornerStyles(t *testing.T) {
	for _, test := range []struct {
		name        string
		orientation GroupOrientation
		position    buttonGroupPosition
		want        style.CornerRadii
	}{
		{name: "single", position: buttonGroupSingle,
			want: style.CornerRadii{TopLeft: 4, TopRight: 4, BottomLeft: 4, BottomRight: 4}},
		{name: "horizontal-start", position: buttonGroupStart,
			want: style.CornerRadii{TopLeft: 4, BottomLeft: 4}},
		{name: "horizontal-middle", position: buttonGroupMiddle},
		{name: "horizontal-end", position: buttonGroupEnd,
			want: style.CornerRadii{TopRight: 4, BottomRight: 4}},
		{name: "vertical-start", orientation: GroupVertical, position: buttonGroupStart,
			want: style.CornerRadii{TopLeft: 4, TopRight: 4}},
		{name: "vertical-middle", orientation: GroupVertical, position: buttonGroupMiddle},
		{name: "vertical-end", orientation: GroupVertical, position: buttonGroupEnd,
			want: style.CornerRadii{BottomLeft: 4, BottomRight: 4}},
	} {
		t.Run(test.name, func(t *testing.T) {
			group := buttonGroupItemStyle{grouped: true, orientation: test.orientation, position: test.position}
			resolved := group.cornerStyle(unit.Dp(4)).Resolve(style.State{})
			var radii style.CornerRadii
			if resolved.Paint.Radii != nil {
				radii = *resolved.Paint.Radii
			}
			if radii != test.want {
				t.Fatalf("corner radii = %+v, want %+v", radii, test.want)
			}
		})
	}
}

func TestGroupInheritsOptionsAndPreservesButtonOverrides(t *testing.T) {
	group := Group(
		New("First"),
		New("Second").Variant(Danger).Size(Large).Disabled(false),
	).Variant(Secondary).Size(Small).Disabled(true)

	first := group.prepareButton(group.buttons[0], 0, 2)
	second := group.prepareButton(group.buttons[1], 1, 2)
	if first.variant != Secondary || first.size != Small || !first.disabled {
		t.Fatalf("inherited button = %#v", first)
	}
	if second.variant != Danger || second.size != Large || second.disabled {
		t.Fatalf("overridden button = %#v", second)
	}
}

func TestGroupLayout(t *testing.T) {
	buttons := []Widget{
		New("A").IconOnly(),
		New("B").IconOnly(),
	}
	horizontal := Group(buttons...).Size(Small)
	if dims := horizontal.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(48, 24) {
		t.Fatalf("horizontal ButtonGroup size = %v, want (48,24)", dims.Size)
	}
	vertical := horizontal.Orientation(GroupVertical)
	if dims := vertical.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(24, 48) {
		t.Fatalf("vertical ButtonGroup size = %v, want (24,48)", dims.Size)
	}
	constrained := testLayoutContext()
	constrained.Constraints = layout.Exact(image.Pt(300, 200))
	if dims := vertical.Layout(newContext(nil), constrained); dims.Size != image.Pt(24, 48) {
		t.Fatalf("constrained vertical ButtonGroup size = %v, want intrinsic (24,48)", dims.Size)
	}
	fullWidth := Group(buttons...).FillWidth().Separators(true)
	if dims := fullWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 32) {
		t.Fatalf("full-width ButtonGroup size = %v, want (300,32)", dims.Size)
	}
	verticalFillWidth := fullWidth.Orientation(GroupVertical)
	if dims := verticalFillWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 64) {
		t.Fatalf("vertical full-width ButtonGroup size = %v, want (300,64)", dims.Size)
	}
}

func TestVerticalGroupEqualizesButtonWidths(t *testing.T) {
	narrow := &buttonGroupWidthProbe{size: image.Pt(16, 16)}
	wide := &buttonGroupWidthProbe{size: image.Pt(80, 16)}
	group := Group(
		New("").Content(narrow).Label("Narrow"),
		New("").Content(wide).Label("Wide"),
	).Orientation(GroupVertical)
	var router input.Router
	var operations op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 200)}, Source: router.Source(), Ops: &operations}
	if dims := group.Layout(newContext(nil), gtx); dims.Size != image.Pt(104, 64) {
		t.Fatalf("vertical ButtonGroup size = %v, want (104,64)", dims.Size)
	}
	router.Frame(&operations)
	buttons := buttonGroupSemanticButtons(router.AppendSemantics(nil))
	widths := make(map[string]int, 2)
	for _, button := range buttons {
		widths[button.Desc.Label] = button.Desc.Bounds.Dx()
	}
	if widths["Narrow"] != 104 || widths["Wide"] != 104 {
		t.Fatalf("vertical button semantic bounds = %#v, want two 104px buttons", buttons)
	}
	if narrow.layouts != 1 || wide.layouts != 1 {
		t.Fatalf("vertical button content layouts = %d/%d, want 1/1", narrow.layouts, wide.layouts)
	}
}

func TestVerticalGroupPassesDisabledContext(t *testing.T) {
	for _, button := range []Widget{
		New("").Content(&enabledProbeWidget{}).Disabled(true),
		New("").Content(&enabledProbeWidget{}).Loading(true),
	} {
		probe := button.child.(*enabledProbeWidget)
		Group(button).Orientation(GroupVertical).Layout(newContext(nil), testLayoutContext())
		if probe.enabled {
			t.Fatalf("%s button content was laid out as enabled", button.key)
		}
	}
}

func TestVerticalGroupUsesStyleDuringMeasurement(t *testing.T) {
	group := (Group(
		New("Styled").Loading(true),
		New("Default"),
	).Orientation(GroupVertical)).Style(style.Style{}.Padding(0, 48))
	defaultGroup := Group(
		New("Styled").Loading(true),
		New("Default"),
	).Orientation(GroupVertical)

	styledDims := group.Layout(newContext(nil), testLayoutContext())
	defaultDims := defaultGroup.Layout(newContext(nil), testLayoutContext())
	if styledDims.Size.X <= defaultDims.Size.X || styledDims.Size.Y != 64 {
		t.Fatalf("styled ButtonGroup size = %v, want wider than %v with height 64", styledDims.Size, defaultDims.Size)
	}
}

type buttonGroupWidthProbe struct {
	size    image.Point
	layouts int
}

func (p *buttonGroupWidthProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func buttonGroupSemanticButtons(nodes []input.SemanticNode) []input.SemanticNode {
	var buttons []input.SemanticNode
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			buttons = append(buttons, node)
		}
		buttons = append(buttons, buttonGroupSemanticButtons(node.Children)...)
	}
	return buttons
}
