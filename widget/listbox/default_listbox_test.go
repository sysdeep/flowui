package listbox

import (
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestDefaultListBoxSelectedBackground(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		defaults := buildListBoxDefaults(&active)
		for _, hovered := range []bool{false, true} {
			state := style.State{Selected: true, Hovered: hovered}
			item := defaults.item(ItemDefault)
			resolved := style.CascadePart(state, style.PartItem, item)
			got, ok := resolved.Paint.Background.(style.SolidColor)
			want := active.Semantic().Tone(theme.TonePrimary).Soft
			if hovered {
				want = active.Semantic().Tone(theme.TonePrimary).SoftHover
			}
			if !ok || got.Color != want {
				t.Fatalf("selected hovered=%v background = %#v, want %#v", hovered, resolved.Paint.Background, want)
			}
		}
	}
}

func TestDefaultListBoxItemVariantColors(t *testing.T) {
	for _, active := range []theme.Theme{theme.Light(), theme.Dark()} {
		defaults := buildListBoxDefaults(&active)
		sem := active.Semantic()
		for _, variant := range []ItemVariant{ItemDefault, ItemDanger} {
			item := defaults.item(variant)
			want := sem.Foreground
			if variant == ItemDanger {
				want = sem.Tone(theme.ToneDanger).Fill
			}
			for _, state := range []style.State{{}, {Disabled: true}} {
				resolved := style.CascadePart(state, style.PartItem, item)
				if resolved.Text == nil {
					t.Fatalf("variant=%v state=%#v text declaration missing", variant, state)
				}
				got, ok := resolved.Text.Color.(style.SolidColor)
				if !ok {
					t.Fatalf("variant=%v state=%#v text color = %#v", variant, state, resolved.Text.Color)
				}
				if state.Disabled {
					want = active.DisabledColor(want)
				}
				if got.Color != want {
					t.Fatalf("variant=%v state=%#v text color = %#v, want %#v", variant, state, got.Color, want)
				}
			}
		}
	}
}
