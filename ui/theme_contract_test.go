package ui

import (
	"testing"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/window"
)

func themeContractProgram() Program[int, struct{}] {
	return NewProgram(
		0,
		func(*int, struct{}) Cmd[struct{}] { return nil },
		func(*Context, int, Send[struct{}]) Widget { return nil },
	)
}

func TestNewWindowRequiresExplicitTheme(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("NewWindow accepted a missing explicit theme")
		}
	}()

	NewWindow("missing-theme", themeContractProgram())
}

func TestNewWindowAcceptsTokenTheme(t *testing.T) {
	spec := NewWindow("neutral", themeContractProgram(), window.Theme(theme.New(theme.Seed{})))
	if spec.Key() != "neutral" {
		t.Fatalf("window key = %q", spec.Key())
	}
}

func TestNewWindowAcceptsDefaultTheme(t *testing.T) {
	spec := NewWindow(
		"default-theme",
		themeContractProgram(),
		window.Theme(theme.Light()),
	)
	if spec.Key() != "default-theme" {
		t.Fatalf("window key = %q, want default-theme", spec.Key())
	}
}
