package theme

import "errors"

// ErrMissingTheme indicates that a required theme was not supplied.
var ErrMissingTheme = errors.New("flowui: theme is missing")
