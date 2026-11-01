package navigation

import (
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type PaginationSize uint8

const (
	PaginationSizeMedium PaginationSize = iota
	PaginationSizeSmall
	PaginationSizeLarge
)

type PaginationWidget struct {
	style            style.Style
	key              string
	page             int
	total            int
	size             PaginationSize
	summary          core.Widget
	siblings         int
	boundaries       int
	showControls     bool
	disabled         bool
	previousLabel    string
	nextLabel        string
	hasPreviousLabel bool
	hasNextLabel     bool
	onChange         func(int)
	visualCache      *paginationVisualCache
}

func NewPagination(page, total int) PaginationWidget {
	total = max(total, 1)
	return PaginationWidget{
		key:          core.AutoKey("pagination"),
		page:         min(max(page, 1), total),
		total:        total,
		siblings:     1,
		boundaries:   1,
		showControls: true,
		visualCache:  &paginationVisualCache{},
	}
}

// Key sets a stable identity for pagination interaction state.
func (p PaginationWidget) Key(key string) PaginationWidget {
	p.key = key
	return p
}

func (p PaginationWidget) Size(size PaginationSize) PaginationWidget {
	p.size = size
	return p
}

func (p PaginationWidget) Summary(summary core.Widget) PaginationWidget {
	p.summary = summary
	return p
}

func (p PaginationWidget) Siblings(count int) PaginationWidget {
	p.siblings = max(count, 0)
	return p
}

func (p PaginationWidget) Boundaries(count int) PaginationWidget {
	p.boundaries = max(count, 0)
	return p
}

func (p PaginationWidget) ShowControls(show bool) PaginationWidget {
	p.showControls = show
	return p
}

func (p PaginationWidget) Labels(previous, next string) PaginationWidget {
	p.previousLabel = previous
	p.nextLabel = next
	p.hasPreviousLabel = true
	p.hasNextLabel = true
	return p
}

func (p PaginationWidget) Disabled(disabled bool) PaginationWidget {
	p.disabled = disabled
	return p
}

func (p PaginationWidget) OnChange(fn func(int)) PaginationWidget {
	p.onChange = fn
	return p
}

// paginationPageItems returns page numbers with zero representing an ellipsis.
func paginationPageItems(page, total, boundaries, siblings int) []int {
	if total <= 0 {
		return nil
	}
	// The page set is the union of at most five sorted ranges. Merging those
	// ranges directly avoids a map, a sort, and a second page slice on every
	// layout frame.
	type pageRange struct{ start, end int }
	ranges := [5]pageRange{}
	rangeCount := 0
	addRange := func(start, end int) {
		start, end = max(start, 1), min(end, total)
		if start <= end {
			ranges[rangeCount] = pageRange{start: start, end: end}
			rangeCount++
		}
	}
	addRange(1, boundaries)
	addRange(page-siblings, page+siblings)
	addRange(total-boundaries+1, total)
	addRange(1, 1)
	addRange(total, total)
	for index := 1; index < rangeCount; index++ {
		current := ranges[index]
		position := index
		for position > 0 && current.start < ranges[position-1].start {
			ranges[position] = ranges[position-1]
			position--
		}
		ranges[position] = current
	}

	// Keep the common compact range in one allocation without trusting the
	// caller-provided boundary counts for capacity arithmetic.
	capacity := min(total, 7) + 2
	items := make([]int, 0, capacity)
	previous := 0
	for index := 0; index < rangeCount; index++ {
		current := ranges[index]
		if current.start <= previous {
			if current.end <= previous {
				continue
			}
			current.start = previous + 1
		}
		if previous > 0 {
			gap := current.start - previous
			if gap == 2 {
				items = append(items, previous+1)
			} else if gap > 2 {
				items = append(items, 0)
			}
		}
		for value := current.start; value <= current.end; value++ {
			items = append(items, value)
		}
		previous = current.end
	}
	return items
}

// Style merges instance declarations into the widget. The declaration is
// resolved as the last cascade layer at layout time, so Style can sit anywhere
// in the chain.
func (w PaginationWidget) Style(values ...style.Style) PaginationWidget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}
