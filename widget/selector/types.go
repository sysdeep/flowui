package selector

import (
	"github.com/qianniancn/flowui/widget/internal/fieldstyle"
	"github.com/qianniancn/flowui/widget/listbox"
)

type SelectItem = listbox.Item
type SelectSection = listbox.Section
type SelectVariant = fieldstyle.Variant

const (
	SelectPrimary   = fieldstyle.Primary
	SelectSecondary = fieldstyle.Secondary
)
