package progress

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
)

func TestMeterKeepsValueTextOverFormatter(t *testing.T) {
	meter := Meter(42).Key("storage").
		ValueFormatter(func(float64) string { return "formatted" }).
		ValueText("42 GB")
	bar := meter
	if !bar.hasValueText || bar.valueText != "42 GB" {
		t.Fatal("explicit meter value text was not retained")
	}
	if got := Meter(42).ValueFormatter(func(float64) string { return "formatted" }).resolved().valueText; got != "formatted" {
		t.Fatalf("formatted value = %q", got)
	}
}

func TestMeterFillUsesConcreteColor(t *testing.T) {
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	bar := Meter(42).Key("storage").Fill(want)
	fill := style.CascadePart(style.State{}, style.PartFill, bar.style)
	solid, ok := fill.Paint.Background.(style.SolidColor)
	if !ok || solid.Color != want {
		t.Fatalf("fill background = %#v, want %v", fill.Paint.Background, want)
	}
}

func TestMeterKeepsDefaultSemanticLabel(t *testing.T) {
	bar := Meter(42).Key("storage")
	if bar.semanticLabel != "" || bar.semanticKey != i18n.MeterDefault {
		t.Fatalf("default semantic label = %q key = %q", bar.semanticLabel, bar.semanticKey)
	}
	if got := bar.semanticDescription(); got != "Meter 42%" {
		t.Fatalf("default semantic description = %q", got)
	}
	chinese := core.NewWithCatalog(nil, nil, locale.LanguageChinese, i18n.Catalog{})
	if got := bar.semanticDescriptionFor(chinese); got != "仪表 42%" {
		t.Fatalf("Chinese semantic description = %q", got)
	}
	if got := Meter(42).Alt("Storage usage").semanticDescription(); got != "Storage usage 42%" {
		t.Fatalf("semantic description = %q", got)
	}
}
