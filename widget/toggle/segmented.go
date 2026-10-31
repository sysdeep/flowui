package toggle

import (
	"image"

	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

type SegmentedItem struct {
	Key      string
	Label    string
	Disabled bool
}

type SegmentedWidget struct {
	key       string
	selected  string
	items     []SegmentedItem
	onChange  func(string)
	disabled  bool
	fullWidth bool
	style     style.Style
}

// NewSegmented creates a segmented control. It starts with nothing selected;
// SelectedKey sets the controlled value.
func NewSegmented(items []SegmentedItem) SegmentedWidget {
	return SegmentedWidget{key: core.AutoKey("segmented"), items: append([]SegmentedItem(nil), items...)}
}

// Key sets an explicit identity for the control state.
func (s SegmentedWidget) Key(key string) SegmentedWidget {
	s.key = key
	return s
}

func (s SegmentedWidget) SelectedKey(key string) SegmentedWidget {
	s.selected = key
	return s
}

func (s SegmentedWidget) OnChange(fn func(string)) SegmentedWidget {
	s.onChange = fn
	return s
}

func (s SegmentedWidget) Disabled(disabled bool) SegmentedWidget {
	s.disabled = disabled
	return s
}

func (s SegmentedWidget) FillWidth() SegmentedWidget {
	s.fullWidth = true
	return s
}

// Style merges instance declarations into the track. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (s SegmentedWidget) Style(values ...style.Style) SegmentedWidget {
	s.style = style.Join(s.style, style.Join(values...))
	return s
}

func (s SegmentedWidget) Layout(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
	owner := core.ClaimKey(ctx, core.KindCustom, s.key)
	gtx.Constraints.Min = image.Point{}
	if s.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	children := make([]core.Widget, 0, len(s.items))
	for _, item := range s.items {
		selected := item.Key == s.selected
		if item.Key == "" {
			panic("flowui: segmented item requires a key")
		}
		choice := NewToggleButton(item.Label).Key(core.DerivedKey(ctx, owner, item.Key)).
			Selected(selected).Variant(ToggleButtonGhost).Disabled(s.disabled || item.Disabled)
		if !selected {
			choice = choice.OnChange(func(next bool) {
				if next && s.onChange != nil {
					s.onChange(item.Key)
				}
			})
		}
		var child core.Widget = choice
		if s.fullWidth {
			child = flowlayout.Expanded(choice.Style(style.New().FillWidth()))
		}
		children = append(children, child)
	}
	active := core.ActiveTheme(ctx)
	defaults := buildSegmentedDefaults(active)
	return flowlayout.Box(flowlayout.Row(children...).Gap(int(defaults.gap))).
		Style(defaults.root, s.style).Layout(ctx, gtx)
}

var _ core.Widget = SegmentedWidget{}
