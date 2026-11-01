// Package code renders a read-only, syntax-highlighted code block.
//
// The block is selectable: dragging across the text picks a rune range and the
// platform copy shortcut writes it to the clipboard. Multi-line blocks show a
// line-number gutter by default; LineNumbers overrides that. Highlighting is
// delegated to an internal tokenizer, so this package owns layout, selection,
// and theme colors only.
//
// Drawing is split from geometry on purpose. Gio's text shaper only shapes, so a
// caller that fills glyph outlines itself must reproduce the shaper's layout
// exactly. This widget measures the text for hit testing, then uses Gio labels
// to draw each already-wrapped span, including native bitmap glyphs.
package code

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/base"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/code/internal/highlight"
	"github.com/qianniancn/flowui/widget/text"
	"golang.org/x/image/math/fixed"
)

const stateSlot = "code"

// defaultHighlighter is shared by every block: creating a Chroma adapter per
// layout would discard its lexer cache.
var defaultHighlighter tokenizer = highlight.New()

type tokenizer interface {
	Tokenize(lang, source string) []highlight.Token
}

// Widget is a highlighted, selectable code block.
type Widget struct {
	key            string
	lang           string
	source         string
	highlighter    tokenizer
	size           unit.Sp
	sizeSet        bool
	lineNumbers    bool
	lineNumbersSet bool
	style          style.Style
}

// New creates a code block for source. lang selects a Chroma lexer; an empty or
// unknown language renders the source without coloring. A multi-line block
// shows line numbers by default; use LineNumbers to override that.
func New(lang, source string) Widget {
	return Widget{key: core.AutoKey("code"), lang: lang, source: expandTabs(normalizeNewlines(source)), size: 14}
}

// Key assigns the stable identity used by selection and copy state.
func (w Widget) Key(key string) Widget {
	w.key = key
	return w
}

func normalizeNewlines(source string) string {
	if !strings.ContainsRune(source, '\r') {
		return source
	}
	source = strings.ReplaceAll(source, "\r\n", "\n")
	return strings.ReplaceAll(source, "\r", "\n")
}

// expandTabs replaces tab characters with spaces on four-column stops. Gio's
// shaper has no tab glyph, so an unexpanded tab shapes as a missing-glyph box.
func expandTabs(source string) string {
	if !strings.ContainsRune(source, '\t') {
		return source
	}
	var builder strings.Builder
	builder.Grow(len(source) + 8)
	column := 0
	for _, r := range source {
		switch r {
		case '\n':
			column = 0
			builder.WriteRune(r)
		case '\t':
			spaces := 4 - column%4
			for range spaces {
				builder.WriteByte(' ')
			}
			column += spaces
		default:
			column++
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// Highlighter overrides the tokenizer. The default understands Go.
// Size overrides the monospace text size in sp. Without it the recipe's
// default applies.
func (w Widget) Size(sp unit.Sp) Widget {
	if sp > 0 {
		w.size = sp
		w.sizeSet = true
	}
	return w
}

// LineNumbers forces the line-number gutter on or off. Without it, a multi-line
// block shows numbers and a single-line block does not.
func (w Widget) LineNumbers(show bool) Widget {
	w.lineNumbers = show
	w.lineNumbersSet = true
	return w
}

// Style merges instance declarations into the block surface.
func (w Widget) Style(values ...style.Style) Widget {
	w.style = style.Join(w.style, style.Join(values...))
	return w
}

// mono is the code face. The weight is explicit: a zero weight lets the shaper
// pick the lightest matching face, which renders code strokes too thin.
func mono(active theme.Theme) font.Font {
	return font.Font{Typeface: active.Typography.MonoTypeface, Weight: font.Normal}
}

// lineHeight is the vertical rhythm of one source line.
func (w Widget) lineHeight() unit.Sp { return w.size * 3 / 2 }

// Layout renders the block inside a padded surface, with a header that names the
// language and offers a copy button.
func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	active := ctx.Theme()
	visual := buildCodeVisual(&active)
	if !w.sizeSet && visual.Metrics.TextSize > 0 {
		w.size = visual.Metrics.TextSize
	}
	owner := core.ClaimKey(ctx, core.KindCustom, w.key)
	behaviour := core.UseState[blockState](ctx, owner, stateSlot)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutContent(ctx, gtx, behaviour, active, visual.Metrics)
	})
	header := w.header(ctx, gtx, behaviour, active)
	return flowlayout.Box(flowlayout.Column(header, content).Gap(int(visual.Metrics.HeaderGap))).
		Style(visual.Root, w.style).Layout(ctx, gtx)
}

// copiedHold is how long the copy button reports the result.
const copiedHold = 1500 * time.Millisecond

// header shows the language tag and the copy action. The button is a click
// target of its own, so pressing it never starts a text selection.
func (w Widget) header(ctx *core.Context, gtx layout.Context, s *blockState, active theme.Theme) core.Widget {
	copied := !s.copiedAt.IsZero() && gtx.Now.Sub(s.copiedAt) < copiedHold
	if copied {
		ctx.Invalidate()
	}
	label := w.sourceName(ctx)
	ink := active.Semantic().MutedForeground
	if copied {
		label = ctx.Text(i18n.CodeCopied)
		ink = active.Colors.Tones[theme.ToneSuccess].Fill
	}
	children := []core.Widget{
		core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			if w.lang == "" {
				return layout.Dimensions{}
			}
			return text.New(w.lang).Size(11).Color(active.Semantic().MutedForeground).Layout(ctx, gtx)
		}),
		flowlayout.Expanded(flowlayout.Box(nil)),
		button.New(label).Key(core.DerivedKey(ctx, core.FullKey(ctx, w.key), "copy")).
			Content(core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
				return drawIcon(gtx, ink, copied)
			})).
			IconOnly().
			Variant(button.Ghost).
			Size(button.Small).
			OnClick(func() { s.copyRequested = true }),
	}
	return flowlayout.Row(children...).AlignMiddle()
}

// sourceName is the accessible name of the copy action.
func (w Widget) sourceName(ctx *core.Context) string {
	if w.lang == "" {
		return ctx.Text(i18n.CodeCopy)
	}
	return ctx.Textf(i18n.CodeCopySource, w.lang)
}

// blockState is the transient selection state of one block. It holds
// no application data: source and language stay on the Widget value.
type blockState struct {
	selectionState base.TextSelection

	// copyRequested is set by the header button and served by the next layout,
	// which is the first place that can reach the clipboard.
	copyRequested bool
	copiedAt      time.Time

	// measureKey identifies the inputs of the cached measurement. Re-shaping the
	// whole block every frame is expensive enough to starve a repainting window.
	measureKey    string
	measureSpans  [][]segment
	measureRows   []row
	measureTops   []int
	measureHeight int
	measureGutter int

	// tokenKey caches tokenization; Chroma lexers are much heavier than the
	// built-in Go scanner, so running them on every frame is not acceptable.
	tokenKey     string
	lineSegments [][]segment
}

func (s *blockState) selection() (int, int, bool) {
	return s.selectionState.Range()
}

// segment is a run of one source line sharing a token kind.
type segment struct {
	text string
	kind highlight.Kind
}

// glyphBox is one glyph's box on screen and the source runes it covers.
type glyphBox struct {
	rect        image.Rectangle
	start, stop int
}

// row is one rendered visual row: its vertical band, the rune range it covers
// without the trailing newline, its glyph boxes, and the source line it belongs
// to. first is true on the first visual row of a wrapped source line.
type row struct {
	top, bottom int
	start, stop int
	width       int
	line        int
	first       bool
	boxes       []glyphBox
}

func (w Widget) layoutContent(
	ctx *core.Context,
	gtx layout.Context,
	s *blockState,
	active theme.Theme,
	metrics codeMetrics,
) layout.Dimensions {
	width := gtx.Constraints.Max.X
	shaper := ctx.TextShaper()
	if shaper == nil || width <= 0 {
		return layout.Dimensions{}
	}
	highlighter := w.highlighter
	if highlighter == nil {
		highlighter = defaultHighlighter
	}
	// The header button cannot reach the clipboard itself; serve its request here.
	if s.copyRequested {
		s.copyRequested = false
		gtx.Execute(clipboard.WriteCmd{
			Type: "application/text",
			Data: io.NopCloser(strings.NewReader(w.source)),
		})
		s.copiedAt = gtx.Now
	}
	runes := []rune(w.source)
	tokenKey := fmt.Sprintf("%T:%p|%s|%s", highlighter, highlighter, w.lang, w.source)
	if s.tokenKey != tokenKey {
		s.tokenKey = tokenKey
		s.lineSegments = splitLines(highlighter.Tokenize(w.lang, w.source))
	}
	lines := s.lineSegments

	showNumbers := w.showLineNumbers(len(lines))
	lineNumberGap := gtx.Dp(metrics.LineNumberGap)
	if metrics.LineNumberGap <= 0 {
		lineNumberGap = gtx.Dp(12)
	}
	gutter := 0
	if showNumbers {
		digits := len(strconv.Itoa(max(len(lines), 1)))
		gutter = w.lineNumberGutter(gtx, shaper, active, digits, lineNumberGap)
		if gutter >= width/2 {
			gutter = 0
		}
	}
	contentWidth := width - gutter
	if contentWidth <= 0 {
		return layout.Dimensions{}
	}

	measureKey := fmt.Sprintf("%d|%v|%v|%s|%d", contentWidth, w.size, active.Typography.MonoTypeface, w.source, gutter)
	if s.measureKey != measureKey {
		s.measureKey = measureKey
		s.measureSpans, s.measureRows, s.measureTops, s.measureHeight =
			w.measure(gtx, shaper, lines, active, contentWidth)
	}
	s.measureGutter = gutter
	lineSpans, rows, tops, contentHeight := s.measureSpans, s.measureRows, s.measureTops, s.measureHeight
	if contentHeight <= 0 {
		return layout.Dimensions{}
	}

	viewport := image.Rect(0, 0, width, contentHeight)
	w.handleInput(ctx, gtx, s, rows, runes, viewport, gutter)

	restoreClip := clip.Rect(viewport).Push(gtx.Ops)
	defer restoreClip.Pop()
	// The block reads as one label; the glyphs are painted, not semantic.
	semantic.LabelOp(w.source).Add(gtx.Ops)

	if start, stop, ok := s.selection(); ok {
		paintSelection(gtx, rows, start, stop, active.Colors.Tones[theme.TonePrimary].Soft, gutter)
	}
	if gutter > 0 {
		w.drawLineNumbers(gtx, shaper, rows, tops, gutter, lineNumberGap, active)
	}
	w.drawLines(gtx, shaper, lineSpans, tops, active, gutter)
	return layout.Dimensions{Size: image.Pt(width, contentHeight)}
}

func (w Widget) showLineNumbers(lines int) bool {
	if w.lineNumbersSet {
		return w.lineNumbers
	}
	return lines > 1
}

func (w Widget) lineNumberGutter(
	gtx layout.Context,
	shaper *giotext.Shaper,
	active theme.Theme,
	digits int,
	gap int,
) int {
	if digits <= 0 {
		return 0
	}
	params := giotext.Parameters{
		Font:       mono(active),
		PxPerEm:    fixed.I(gtx.Sp(w.size)),
		MaxWidth:   1 << 20,
		Locale:     gtx.Locale,
		LineHeight: fixed.I(gtx.Sp(w.lineHeight())),
	}
	shaper.LayoutString(params, strings.Repeat("0", digits))
	width := 0
	for glyph, ok := shaper.NextGlyph(); ok; glyph, ok = shaper.NextGlyph() {
		width += glyph.Advance.Round()
	}
	return width + gap
}

// measure shapes the text to learn where each glyph lands. Lines wrap at the
// available width, breaking inside a token when needed, so one source line may
// cover several visual rows. The returned rows are visual rows: each carries the
// rune range it covers and its glyph rectangles, and lineSpans holds the text
// pieces of each row for the draw pass.
func (w Widget) measure(
	gtx layout.Context,
	shaper *giotext.Shaper,
	lines [][]segment,
	active theme.Theme,
	width int,
) (lineSpans [][]segment, rows []row, tops []int, contentHeight int) {
	pixels := gtx.Sp(w.size)
	lineHeight := gtx.Sp(w.lineHeight())
	if pixels <= 0 || lineHeight <= 0 || width <= 0 {
		return nil, nil, nil, 0
	}
	params := giotext.Parameters{
		Font:             mono(active),
		PxPerEm:          fixed.I(pixels),
		MaxWidth:         width,
		Locale:           gtx.Locale,
		WrapPolicy:       giotext.WrapGraphemes,
		LineHeight:       fixed.I(lineHeight),
		DisableSpaceTrim: true,
	}
	runeIndex := 0
	cursorY := 0
	startRow := func(top, line int, first bool) {
		if n := len(rows); n > 0 {
			rows[n-1].stop = runeIndex
		}
		rows = append(rows, row{top: top, bottom: top + lineHeight, start: runeIndex, line: line, first: first})
		lineSpans = append(lineSpans, nil)
		tops = append(tops, top)
	}
	for lineIndex, line := range lines {
		cursorX := 0
		top := cursorY
		startRow(top, lineIndex+1, true)
		for _, seg := range line {
			rest := []rune(seg.text)
			for len(rest) > 0 {
				// Only the first visual row of this attempt is consumed; the rest
				// of the token is re-shaped against the full width of the next row.
				params.MaxWidth = max(width-cursorX, pixels)
				shaper.LayoutString(params, string(rest))
				type metric struct {
					left, advance, count int
				}
				var glyphs []metric
				firstY := -1
				for glyph, ok := shaper.NextGlyph(); ok; glyph, ok = shaper.NextGlyph() {
					y := int(glyph.Y)
					if firstY < 0 {
						firstY = y
					}
					if y != firstY {
						break
					}
					count := 0
					if glyph.Flags&giotext.FlagClusterBreak != 0 {
						count = int(glyph.Runes)
					}
					glyphs = append(glyphs, metric{left: glyph.X.Round(), advance: glyph.Advance.Round(), count: count})
				}
				fitted := 0
				rowWidth := 0
				current := &rows[len(rows)-1]
				for _, glyph := range glyphs {
					left := cursorX + glyph.left
					if glyph.count > 0 && glyph.advance > 0 {
						current.boxes = append(current.boxes, glyphBox{
							rect:  image.Rect(left, top, left+glyph.advance, top+lineHeight),
							start: runeIndex,
							stop:  runeIndex + glyph.count,
						})
						current.width = max(current.width, left+glyph.advance)
					}
					runeIndex += glyph.count
					fitted += glyph.count
					rowWidth = max(rowWidth, left+glyph.advance)
				}
				if fitted == 0 {
					// The row cannot hold a single rune; stop instead of looping.
					break
				}
				lineSpans[len(lineSpans)-1] = append(lineSpans[len(lineSpans)-1],
					segment{text: string(rest[:fitted]), kind: seg.kind})
				rest = rest[fitted:]
				cursorX = rowWidth
				if len(rest) > 0 {
					cursorX = 0
					top += lineHeight
					startRow(top, lineIndex+1, false)
				}
			}
		}
		rows[len(rows)-1].stop = runeIndex
		runeIndex++ // the newline that separates this line from the next
		cursorY = rows[len(rows)-1].bottom
	}
	return lineSpans, rows, tops, cursorY
}

// drawLines paints already-wrapped spans without asking another text renderer
// to wrap them again. Label supports both vector and native bitmap glyphs.
func (w Widget) drawLines(
	gtx layout.Context,
	shaper *giotext.Shaper,
	lineSpans [][]segment,
	tops []int,
	active theme.Theme,
	gutter int,
) {
	face := mono(active)
	lineGtx := gtx
	lineGtx.Constraints.Min = image.Point{}
	lineGtx.Constraints.Max = image.Pt(1<<20, 1<<20)
	for index, line := range lineSpans {
		if index >= len(tops) {
			continue
		}
		x := gutter
		for _, seg := range line {
			if seg.text == "" {
				continue
			}
			material := op.Record(gtx.Ops)
			paint.ColorOp{Color: kindColor(seg.kind, active)}.Add(gtx.Ops)
			ink := material.Stop()
			offset := op.Offset(image.Pt(x, tops[index])).Push(gtx.Ops)
			widget.Label{}.Layout(lineGtx, shaper, face, w.size, seg.text, ink)
			offset.Pop()
			shaper.LayoutString(giotext.Parameters{
				Font: face, PxPerEm: fixed.I(gtx.Sp(w.size)),
				MaxWidth: 1 << 20, Locale: gtx.Locale, DisableSpaceTrim: true,
			}, seg.text)
			var advance fixed.Int26_6
			for glyph, ok := shaper.NextGlyph(); ok; glyph, ok = shaper.NextGlyph() {
				advance += glyph.Advance
			}
			x += advance.Round()
		}
	}
}

// drawLineNumbers paints the gutter numbers. Wrapped rows keep a blank gutter:
// only the first visual row of a source line is numbered.
func (w Widget) drawLineNumbers(
	gtx layout.Context,
	shaper *giotext.Shaper,
	rows []row,
	tops []int,
	gutter, gap int,
	active theme.Theme,
) {
	if gutter <= 0 {
		return
	}
	lineGtx := gtx
	lineGtx.Constraints.Min = image.Point{}
	lineGtx.Constraints.Max = image.Pt(1<<20, 1<<20)
	face := mono(active)
	ink := active.CodeColors().LineNumber
	for index, line := range rows {
		if !line.first || index >= len(tops) {
			continue
		}
		material := op.Record(gtx.Ops)
		paint.ColorOp{Color: ink}.Add(gtx.Ops)
		foreground := material.Stop()
		macro := op.Record(gtx.Ops)
		dims := widget.Label{}.Layout(lineGtx, shaper, face, w.size, strconv.Itoa(line.line), foreground)
		call := macro.Stop()
		x := max(gutter-gap-dims.Size.X, 0)
		offset := op.Offset(image.Pt(x, tops[index])).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
}

func (w Widget) handleInput(
	ctx *core.Context,
	gtx layout.Context,
	s *blockState,
	rows []row,
	runes []rune,
	viewport image.Rectangle,
	gutter int,
) {
	core.AddPointerArea(gtx, s, viewport, pointer.CursorText)
	s.selectionState.Update(ctx, gtx, s, runes, func(position f32.Point) int {
		position.X = max(position.X-float32(gutter), 0)
		return hitTest(rows, position)
	})
}

// hitTest maps a content-space position to a source rune index.
func hitTest(rows []row, position f32.Point) int {
	if len(rows) == 0 {
		return 0
	}
	y := int(position.Y)
	if y < rows[0].top {
		return rows[0].start
	}
	if y >= rows[len(rows)-1].bottom {
		return rows[len(rows)-1].stop
	}
	for _, line := range rows {
		if y < line.top || y >= line.bottom {
			continue
		}
		x := int(position.X)
		for _, box := range line.boxes {
			if x < box.rect.Min.X {
				return box.start
			}
			if x < box.rect.Max.X {
				if x < (box.rect.Min.X+box.rect.Max.X)/2 {
					return box.start
				}
				return box.stop
			}
		}
		if len(line.boxes) > 0 {
			return line.boxes[len(line.boxes)-1].stop
		}
		return line.start
	}
	return 0
}

// paintSelection draws one rectangle per selected line, extending to the line
// edges when the selection crosses the line break.
func paintSelection(gtx layout.Context, rows []row, start, stop int, fill color.NRGBA, gutter int) {
	for _, line := range rows {
		if stop <= line.start || start >= line.stop {
			continue
		}
		left := 0
		if start > line.start {
			left = xAt(line, start)
		}
		right := line.width
		if stop < line.stop {
			right = xAt(line, stop)
		}
		if right <= left {
			continue
		}
		paint.FillShape(gtx.Ops, fill,
			clip.Rect(image.Rect(left+gutter, line.top, right+gutter, line.bottom)).Op())
	}
}

// xAt returns the horizontal offset of a rune index inside one line.
func xAt(line row, index int) int {
	for _, box := range line.boxes {
		if index <= box.start {
			return box.rect.Min.X
		}
		if index < box.stop {
			return box.rect.Max.X
		}
	}
	if len(line.boxes) > 0 {
		return line.boxes[len(line.boxes)-1].rect.Max.X
	}
	return 0
}

// splitLines breaks tokens on newlines so highlighting survives across lines
// for block comments and raw strings.
func splitLines(tokens []highlight.Token) [][]segment {
	lines := [][]segment{nil}
	for _, token := range tokens {
		text := token.Text
		for {
			index := strings.IndexByte(text, '\n')
			if index < 0 {
				if text != "" {
					last := len(lines) - 1
					lines[last] = append(lines[last], segment{text: text, kind: token.Kind})
				}
				break
			}
			if index > 0 {
				last := len(lines) - 1
				lines[last] = append(lines[last], segment{text: text[:index], kind: token.Kind})
			}
			lines = append(lines, nil)
			text = text[index+1:]
		}
	}
	return lines
}

// kindColor maps a token kind onto the theme's code palette.
func kindColor(kind highlight.Kind, active theme.Theme) color.NRGBA {
	code := active.CodeColors()
	switch kind {
	case highlight.KindKeyword:
		return code.Keyword
	case highlight.KindType:
		return code.Type
	case highlight.KindFunction:
		return code.Function
	case highlight.KindString:
		return code.String
	case highlight.KindNumber:
		return code.Number
	case highlight.KindComment:
		return code.Comment
	case highlight.KindOperator:
		return code.Operator
	case highlight.KindPunctuation:
		return code.Punctuation
	case highlight.KindBuiltin:
		return code.Builtin
	}
	return code.Plain
}

// drawIcon paints the copy affordance: two overlapping sheets, replaced by a
// check mark once the copy lands. It is drawn rather than imported from an icon
// set so the component keeps no icon-pack dependency.
func drawIcon(gtx layout.Context, ink color.NRGBA, copied bool) layout.Dimensions {
	size := gtx.Dp(14)
	if size <= 0 {
		return layout.Dimensions{}
	}
	u := float32(size) / 16
	var path clip.Path
	path.Begin(gtx.Ops)
	if copied {
		path.MoveTo(f32.Pt(3*u, 8.5*u))
		path.LineTo(f32.Pt(6.5*u, 12*u))
		path.LineTo(f32.Pt(13*u, 4.5*u))
	} else {
		// Back sheet: only the edges the front sheet does not cover.
		path.MoveTo(f32.Pt(6*u, 2*u))
		path.LineTo(f32.Pt(14*u, 2*u))
		path.LineTo(f32.Pt(14*u, 10*u))
		path.LineTo(f32.Pt(11*u, 10*u))
		// Front sheet.
		path.MoveTo(f32.Pt(2*u, 6*u))
		path.LineTo(f32.Pt(10*u, 6*u))
		path.LineTo(f32.Pt(10*u, 14*u))
		path.LineTo(f32.Pt(2*u, 14*u))
		path.Close()
	}
	paint.FillShape(gtx.Ops, ink, clip.Stroke{Path: path.End(), Width: 1.5 * u}.Op())
	return layout.Dimensions{Size: image.Pt(size, size)}
}
