package testtheme

import "github.com/qianniancn/flowui/theme"

func (behaviorProvider) Icon(active *theme.Theme, request theme.IconRequest) theme.IconVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.IconVisual{
		Size:  active.Components.Icon.Size,
		Color: request.Foreground,
	}
}
