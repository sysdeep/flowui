package listbox

import (
	"github.com/qianniancn/flowui/internal/core"
)

type Item struct {
	Key         string
	Label       string
	Description string
	Leading     core.Widget
	Trailing    core.Widget
	Indicator   func(selected bool) core.Widget
	Disabled    bool
	Variant     ItemVariant
}

type Section struct {
	Title string
	Items []Item
}

type ItemVariant int

const (
	ItemDefault ItemVariant = iota
	ItemDanger
)

type SelectionMode int

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionNone
)
