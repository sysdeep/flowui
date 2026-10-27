package theme_test

import (
	"errors"
	"github.com/qianniancn/flowui/theme/feedbackstyle"
	"github.com/qianniancn/flowui/theme/fieldstyle"
	"github.com/qianniancn/flowui/theme/navigationstyle"
	"github.com/qianniancn/flowui/theme/surfacestyle"
	"testing"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/actionstyle"
	"github.com/qianniancn/flowui/theme/datastyle"
	inputstyle "github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/overlaystyle"
)

func TestSharedVisualRecipesRejectMissingTheme(t *testing.T) {
	tests := map[string]func(){
		"button":       func() { actionstyle.BuildButton(nil, theme.ButtonRequest{}, actionstyle.Config{}) },
		"badge":        func() { feedbackstyle.BuildBadge(nil, theme.BadgeRequest{}) },
		"checkbox":     func() { actionstyle.BuildCheckbox(nil, theme.CheckboxRequest{}) },
		"combo":        func() { fieldstyle.BuildCombo(nil, theme.ComboRequest{}) },
		"menu":         func() { navigationstyle.BuildMenu(nil, theme.MenuRequest{}) },
		"dropdown":     func() { navigationstyle.BuildDropdown(nil, theme.DropdownRequest{}) },
		"popover":      func() { overlaystyle.BuildPopover(nil, theme.PopoverRequest{}) },
		"tooltip":      func() { overlaystyle.BuildTooltip(nil, theme.TooltipRequest{}) },
		"modal":        func() { overlaystyle.BuildModal(nil, theme.ModalRequest{}) },
		"listbox":      func() { datastyle.BuildListBox(nil, theme.ListBoxRequest{}) },
		"sidebar":      func() { navigationstyle.BuildSidebar(nil, theme.SidebarRequest{}) },
		"table":        func() { datastyle.BuildTable(nil, theme.TableRequest{}) },
		"datepicker":   func() { fieldstyle.BuildDatePicker(nil, theme.DatePickerRequest{}) },
		"description":  func() { surfacestyle.BuildDescription(nil, theme.DescriptionRequest{}) },
		"input":        func() { inputstyle.BuildInput(nil, theme.InputRequest{}) },
		"input-action": func() { fieldstyle.BuildInputAction(nil, theme.InputActionRequest{}) },
		"label":        func() { surfacestyle.BuildLabel(nil, theme.LabelRequest{}) },
		"surface":      func() { surfacestyle.BuildSurface(nil, theme.SurfaceRequest{}) },
		"card":         func() { surfacestyle.BuildCard(nil, theme.CardRequest{}) },
		"statusbar":    func() { datastyle.BuildStatusBar(nil, theme.StatusBarRequest{}, datastyle.DefaultConfig()) },
		"radio":        func() { actionstyle.BuildRadioGroup(nil, theme.RadioGroupRequest{}) },
		"select":       func() { fieldstyle.BuildSelect(nil, theme.SelectRequest{}) },
		"slider":       func() { fieldstyle.BuildSlider(nil, theme.SliderRequest{}) },
		"switch":       func() { actionstyle.BuildSwitch(nil, theme.SwitchRequest{}) },
		"switch-group": func() { actionstyle.BuildSwitchGroup(nil, theme.SwitchGroupRequest{}) },
		"togglebutton": func() { actionstyle.BuildToggleButton(nil, theme.ToggleRequest{}) },
		"code":         func() { surfacestyle.BuildCode(nil, theme.CodeRequest{}) },
		"image":        func() { surfacestyle.BuildImage(nil, theme.ImageRequest{}) },
		"skeleton":     func() { feedbackstyle.BuildSkeleton(nil, theme.SkeletonRequest{}) },
		"segmented":    func() { navigationstyle.BuildSegmented(nil, theme.SegmentedRequest{}) },
	}
	for name, build := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				err, ok := recovered.(error)
				if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
					t.Fatalf("panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
				}
			}()
			build()
		})
	}
}
