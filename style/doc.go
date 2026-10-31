// Package style contains FlowUI's immutable style declarations: layout, paint,
// text, and transition properties, plus state rules and named parts.
//
// It is a pure declaration mechanism and depends on no other FlowUI package.
// Themes supply values; widgets own their default appearance and express it
// through these declarations. Entry points that need the active frame context,
// inherited styles, or transition state live in
// github.com/qianniancn/flowui/style/resolve.
package style
