package button

import (
	"fmt"
	"reflect"
	"testing"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
)

func TestDefaultButtonSizes(t *testing.T) {
	active := theme.Light()
	for _, test := range []struct {
		name                             string
		size                             Size
		height, padding, radius, spinner unit.Dp
		fontSize                         unit.Sp
	}{
		{name: "small", size: Small, height: 24, padding: 8, radius: 3, spinner: 14, fontSize: 12},
		{name: "medium", size: Medium, height: 32, padding: 12, radius: 4, spinner: 16,
			fontSize: active.Typography.ControlSize},
		{name: "large", size: Large, height: 40, padding: 16, radius: 6, spinner: 18,
			fontSize: active.Typography.ControlSize + 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, iconOnly := range []bool{false, true} {
				value := Widget{size: test.size, iconOnly: iconOnly}
				defaults := value.defaultStyles(&active)
				root := style.Cascade(style.State{}, defaults.defaults, defaults.variant, defaults.size)
				padding := test.padding
				if iconOnly {
					padding = 0
					if root.Box.Width == nil || *root.Box.Width != test.height {
						t.Fatalf("icon-only width = %v, want %v", root.Box.Width, test.height)
					}
				}
				if *root.Box.Height != test.height || root.Box.Padding.Left != padding ||
					root.Box.Padding.Right != padding || *root.Paint.Radius != test.radius ||
					*root.Text.FontSize != test.fontSize || defaults.spinner != test.spinner {
					t.Fatalf("unexpected default geometry for icon-only=%v: %#v", iconOnly, root)
				}
			}
		})
	}
}

func TestDefaultButtonStates(t *testing.T) {
	for name, active := range map[string]theme.Theme{"light": theme.Light(), "dark": theme.Dark()} {
		t.Run(name, func(t *testing.T) {
			for variant := Primary; variant <= DangerOutline; variant++ {
				t.Run(fmt.Sprint(variant), func(t *testing.T) {
					defaults := Widget{variant: variant}.defaultStyles(&active)
					for _, state := range []style.State{
						{}, {Hovered: true}, {Pressed: true}, {FocusVisible: true},
						{Disabled: true}, {Loading: true},
					} {
						root := style.Cascade(state, defaults.defaults, defaults.variant, defaults.size)
						if root.Box == nil || root.Box.Height == nil || *root.Box.Height <= 0 {
							t.Fatalf("state %+v has no positive height", state)
						}
						for _, part := range []style.Part{style.PartLabel, style.PartIcon, style.PartIndicator} {
							resolved := style.CascadePart(state, part, defaults.defaults, defaults.variant, defaults.size)
							if resolved.Text == nil {
								t.Fatalf("part %v has no text style", part)
							}
							if _, ok := resolve.Color(resolved.Text.Color); !ok {
								t.Fatalf("part %v has no foreground", part)
							}
						}
						if state.FocusVisible && (root.Paint.Outline == nil || root.Paint.Outline.Width != 2 ||
							root.Paint.Outline.Offset != 2) {
							t.Fatal("keyboard focus must show the default outline")
						}
					}
				})
			}
		})
	}
}

func TestSecondaryButtonPressedStepsPastHover(t *testing.T) {
	active := theme.Light()
	visual := Widget{variant: Secondary}.defaultStyles(&active)

	rest := visual.variant.Resolve(style.State{})
	hover := visual.variant.Resolve(style.State{Hovered: true})
	pressed := visual.variant.Resolve(style.State{Hovered: true, Pressed: true})
	if hover.Paint.Background == pressed.Paint.Background {
		t.Fatalf("pressed background = %v, same as hover", hover.Paint.Background)
	}
	if rest.Paint.Background == hover.Paint.Background {
		t.Fatalf("hover background = %v, same as resting", hover.Paint.Background)
	}
}

func TestButtonDisabledOverridesInteraction(t *testing.T) {
	for _, mode := range []struct {
		name string
		new  func() theme.Theme
	}{
		{name: "light", new: theme.Light},
		{name: "dark", new: theme.Dark},
	} {
		t.Run(mode.name, func(t *testing.T) {
			active := mode.new()
			for _, test := range []struct {
				name    string
				variant Variant
			}{
				{name: "primary", variant: Primary},
				{name: "secondary", variant: Secondary},
				{name: "tertiary", variant: Tertiary},
				{name: "ghost", variant: Ghost},
				{name: "outline", variant: Outline},
				{name: "danger", variant: Danger},
				{name: "danger-soft", variant: DangerSoft},
				{name: "danger-outline", variant: DangerOutline},
			} {
				t.Run(test.name, func(t *testing.T) {
					visual := Widget{variant: test.variant}.defaultStyles(&active)
					for _, state := range []style.State{
						{Disabled: true}, {Disabled: true, Hovered: true}, {Disabled: true, Pressed: true},
					} {
						declaration := style.Cascade(state, visual.defaults, visual.variant, visual.size)
						fg, _ := resolve.Color(declaration.Text.Color)
						brush, _ := resolve.Brush(declaration.Paint.Background)
						bg := brush.ColorAt(.5)
						sem := active.Semantic()
						if fg != sem.DisabledForeground || bg != sem.DisabledBackground {
							t.Errorf("state %+v: disabled colours = %v on %v", state, fg, bg)
						}
					}
				})
			}
		})
	}
}

func TestDefaultVariantsRemainPaintable(t *testing.T) {
	for name, active := range map[string]theme.Theme{
		"light": theme.Light(), "dark": theme.Dark(),
		"seed": theme.New(theme.Seed{}),
	} {
		t.Run(name, func(t *testing.T) {
			for name, variant := range map[string]Variant{
				"primary": Primary, "secondary": Secondary, "tertiary": Tertiary,
				"ghost": Ghost, "outline": Outline, "danger": Danger,
				"danger-soft": DangerSoft, "danger-outline": DangerOutline,
			} {
				t.Run(name, func(t *testing.T) {
					visual := Widget{variant: variant}.defaultStyles(&active)
					rest := visual.variant.Resolve(style.State{})
					hover := visual.variant.Resolve(style.State{Hovered: true})
					pressed := visual.variant.Resolve(style.State{Hovered: true, Pressed: true})
					if rest.Text.Color == nil {
						t.Fatal("variant has no label color")
					}
					if rest.Paint.Background == hover.Paint.Background && rest.Text.Color == hover.Text.Color && rest.Paint.Border == hover.Paint.Border {
						t.Error("hover gives no visual feedback")
					}
					if hover.Paint.Background == pressed.Paint.Background && hover.Text.Color == pressed.Text.Color && hover.Paint.Border == pressed.Paint.Border {
						t.Error("press gives no visual feedback beyond hover")
					}
				})
			}
		})
	}
}

func TestUnknownButtonVariantFallsBackToPrimary(t *testing.T) {
	active := theme.Light()
	unknown := Widget{variant: Variant(255)}.defaultStyles(&active)
	primary := Widget{variant: Primary}.defaultStyles(&active)
	for _, state := range []style.State{{}, {Hovered: true}, {Pressed: true}, {Disabled: true}} {
		got := style.Cascade(state, unknown.defaults, unknown.variant, unknown.size)
		want := style.Cascade(state, primary.defaults, primary.variant, primary.size)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("state %+v: unknown variant style = %#v, want primary %#v", state, got, want)
		}
	}
}
