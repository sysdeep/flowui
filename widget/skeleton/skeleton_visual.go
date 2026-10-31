package skeleton

import (
	"time"

	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// skeletonDefaults is the component-owned appearance and geometry of a
// placeholder. The active theme only supplies tokens; sizing, layout, and the
// pulse animation stay with the widget.
type skeletonDefaults struct {
	root      style.Style
	highlight style.Style
	height    unit.Dp
	radius    unit.Dp
	period    time.Duration
	motion    theme.MotionTheme
}

func buildSkeletonDefaults(active *theme.Theme) skeletonDefaults {
	if active == nil {
		panic(theme.ErrMissingTheme)
	}
	semantic := active.Semantic()
	return skeletonDefaults{
		root: style.New().
			Background(style.Solid(semantic.Muted)).
			Opacity(1),
		highlight: style.New().
			Background(style.Solid(semantic.Border)),
		height: 16,
		radius: 6,
		period: 1600 * time.Millisecond,
		motion: active.Motion,
	}
}
