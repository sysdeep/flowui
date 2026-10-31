package ui_test

import (
	"context"
	"testing"

	"github.com/qianniancn/flowui/ui"
)

type matchModel struct {
	count int
	name  string
}

type matchMessage any

type matchIncrement struct{ amount int }
type matchRename struct{ name string }
type matchIgnored struct{}

func TestMatchDispatchesFirstMatchingHandler(t *testing.T) {
	model := matchModel{}
	called := 0
	command := ui.Match(&model, matchMessage(matchIncrement{amount: 2}),
		ui.On(func(model *matchModel, msg matchIncrement) ui.Cmd[matchMessage] {
			model.count += msg.amount
			called++
			return func(_ context.Context, send ui.Send[matchMessage]) error {
				send(matchRename{name: "from command"})
				return nil
			}
		}),
		ui.On(func(model *matchModel, _ matchIncrement) ui.Cmd[matchMessage] {
			model.count = 100
			called++
			return nil
		}),
	)
	if command == nil || model.count != 2 || called != 1 {
		t.Fatalf("match result: model=%+v called=%d command=%v", model, called, command)
	}
	command(context.Background(), func(msg matchMessage) {
		if value, ok := msg.(matchRename); !ok || value.name != "from command" {
			t.Fatalf("command message = %#v, want rename", msg)
		}
	})
}

func TestMatchReturnsNilForUnmatchedMessage(t *testing.T) {
	model := matchModel{}
	command := ui.Match(&model, matchMessage(matchIgnored{}),
		ui.On(func(model *matchModel, msg matchIncrement) ui.Cmd[matchMessage] {
			model.count += msg.amount
			return nil
		}),
		ui.On(func(model *matchModel, msg matchRename) ui.Cmd[matchMessage] {
			model.name = msg.name
			return nil
		}),
	)
	if command != nil || model != (matchModel{}) {
		t.Fatalf("unmatched result: model=%+v command=%v", model, command)
	}
}
