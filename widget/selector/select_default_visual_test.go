package selector

import (
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestSelectDefaultsFollowThemeTokens(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		defaults := buildSelectDefaults(&active, SelectPrimary)

		trigger := style.Cascade(style.State{}, defaults.trigger)
		if trigger.Box == nil || trigger.Box.Height == nil {
			t.Fatal("select trigger has no height")
		}
		if got := *trigger.Box.Height; got != selectDefaultHeight {
			t.Fatalf("select trigger height = %v, want %v", got, selectDefaultHeight)
		}

		panel := style.CascadePart(style.State{Open: true}, style.PartPanel, defaults.panel)
		if panel.Paint == nil || panel.Paint.Border == nil || panel.Paint.Border.Width == nil {
			t.Fatal("select panel has no border")
		}
		if got := *panel.Paint.Border.Width; got != selectDefaultPanelBorderWidth {
			t.Fatalf("select panel border = %v, want %v", got, selectDefaultPanelBorderWidth)
		}
		if len(panel.Paint.Shadows) == 0 || panel.Paint.Shadows[0].Profile == nil ||
			*panel.Paint.Shadows[0].Profile != style.ShadowMenu {
			t.Fatal("select panel does not use the menu shadow profile")
		}

		item := style.CascadePart(style.State{}, style.PartItem, defaults.item)
		if item.Box == nil || item.Box.MinHeight == nil {
			t.Fatal("select item has no minimum height")
		}
		if got := *item.Box.MinHeight; got != 32 {
			t.Fatalf("select item height = %v, want 32", got)
		}
	}
}
