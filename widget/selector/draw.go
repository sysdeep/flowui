package selector

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// selectPartWidth resolves a part's slot width, falling back to the theme
// metric when the declaration does not set one.
func selectPartWidth(gtx layout.Context, resolved style.ResolvedStyle, fallback unit.Dp) int {
	if resolved.Box != nil && resolved.Box.Width != nil {
		return max(gtx.Dp(*resolved.Box.Width), 0)
	}
	return max(gtx.Dp(fallback), 0)
}

// selectPartSize resolves a square glyph size from a part, falling back to the
// theme metric when the declaration does not set one.
func selectPartSize(gtx layout.Context, resolved style.ResolvedStyle, fallback unit.Dp) int {
	size := fallback
	if resolved.Box != nil {
		switch {
		case resolved.Box.Width != nil:
			size = *resolved.Box.Width
		case resolved.Box.Height != nil:
			size = *resolved.Box.Height
		}
	}
	return max(gtx.Dp(size), 1)
}

// drawSelectChevron paints the trigger glyph directly. rotation is expressed in
// radians.
// DrawChevron paints the Select indicator glyph. Select-like controls should
// use this instead of substituting a text arrow, which has different metrics
// and renders inconsistently across fonts and platforms.
func DrawChevron(gtx layout.Context, size int, stroke unit.Dp, foreground color.NRGBA, rotation float32) {
	if size <= 0 {
		return
	}
	center := f32.Pt(float32(size)/2, float32(size)/2)
	transform := op.Affine(f32.AffineId().Rotate(center, rotation)).Push(gtx.Ops)
	defer transform.Pop()

	halfWidth := float32(size) * 0.24
	halfHeight := float32(size) * 0.14
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-halfWidth, center.Y-halfHeight))
	path.LineTo(f32.Pt(center.X, center.Y+halfHeight))
	path.LineTo(f32.Pt(center.X+halfWidth, center.Y-halfHeight))
	strokeStack := clip.Stroke{
		Path:  path.End(),
		Width: float32(max(gtx.Dp(stroke), 1)),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	strokeStack.Pop()
}
