package style

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/pointer"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

// Style is an immutable style declaration. Its fluent methods return an
// independent value; only a resolver exposes its computed properties.
type Style struct {
	box   *BoxStyle
	paint *PaintStyle
	text  *TextStyle
	trans *TransformStyle

	transitions []Transition

	// rules are state table rows, ordered without evaluating predicates.
	// CacheUnsafe accounts for custom easing in their Set declarations.
	rules []Rule

	// dim scales the alpha of every color this declaration resolves to. It is a
	// scalar rather than a closure so the table above stays hashable, and it is
	// applied once at the resolution boundary, not per rule row, so a row that
	// asks for it dims the merged result rather than only its own fields.
	dim *float32

	parts map[Part]Style
}

// ResolvedStyle is a computed snapshot used for layout and drawing.
//
// Resolvers may return cache-owned data when the snapshot has no transitions;
// such values must be treated as read-only. Use the runtime CloneResolvedStyle
// helper before mutating a snapshot whose ownership is not otherwise known.
type ResolvedStyle struct {
	Box         *BoxStyle
	Paint       *PaintStyle
	Text        *TextStyle
	Trans       *TransformStyle
	Transitions []Transition
}

// Part identifies a named element inside a compound component. Applications
// may define additional names for custom components.
type Part string

const (
	PartRoot        Part = ""
	PartContent     Part = "content"
	PartLabel       Part = "label"
	PartDescription Part = "description"
	PartIcon        Part = "icon"
	PartTrack       Part = "track"
	PartFill        Part = "fill"
	PartThumb       Part = "thumb"
	PartIndicator   Part = "indicator"
	PartPanel       Part = "panel"
	PartItem        Part = "item"
	PartBackdrop    Part = "backdrop"
	PartPlaceholder Part = "placeholder"
	PartSelection   Part = "selection"
	PartPrefix      Part = "prefix"
	PartSuffix      Part = "suffix"
)

type BoxStyle struct {
	Width       *unit.Dp
	Height      *unit.Dp
	MinWidth    *unit.Dp
	MaxWidth    *unit.Dp
	MinHeight   *unit.Dp
	MaxHeight   *unit.Dp
	FillWidth   *bool
	FillHeight  *bool
	AspectRatio *float32
	Padding     *Insets
	Margin      *Insets
	Overflow    *Overflow
	Cursor      *pointer.Cursor

	paddingMask uint8
	marginMask  uint8
}

type Insets struct {
	Top, Right, Bottom, Left unit.Dp
}

type Overflow uint8

const (
	OverflowVisible Overflow = iota
	OverflowHidden
)

type PaintSource interface {
	isPaintSource()
}

// ColorSource is a concrete color value. The same value can be used by
// backgrounds, text, and borders.
type ColorSource interface {
	PaintSource
	isColorSource()
}

type SolidColor struct {
	Color color.NRGBA
}

func (SolidColor) isPaintSource() {}
func (SolidColor) isColorSource() {}

type GradientStop struct {
	Offset float32
	Color  ColorSource
}

type Gradient struct {
	AngleDegrees float32
	Stops        []GradientStop
}

func (Gradient) isPaintSource() {}

type PaintStyle struct {
	Background PaintSource
	Border     *BorderStyle
	// BorderBottom is an optional bottom-only border. It is kept separate from
	// Border so a component can underline a control without painting the other
	// three sides.
	BorderBottom *BorderStyle
	Radius       *unit.Dp
	Radii        *CornerRadii
	Shadows      []Shadow
	Outline      *OutlineStyle
	// VisualOutset is an optional minimum visual overflow reservation. It is
	// combined with the shadow and outline extents when a clipping container
	// lays out this box.
	VisualOutset *Insets
	Opacity      *float32

	radiusMask      uint8
	backgroundSet   bool
	shadowsSet      bool
	visualOutsetSet bool
}

type CornerRadii struct {
	TopLeft, TopRight, BottomRight, BottomLeft unit.Dp
}

type Shadow struct {
	OffsetX, OffsetY unit.Dp
	Blur, Spread     unit.Dp
	Color            ColorSource
	Profile          *ShadowProfile
	// Transient shadows are painted but do not reserve visual overflow in
	// clipping containers. Use this for focus and validation decoration that
	// would otherwise change viewport layout when the state changes.
	Transient bool
}

type ShadowProfile uint8

const (
	ShadowSurface ShadowProfile = iota
	ShadowOverlay
	ShadowMenu
)

type OutlineStyle struct {
	Width  unit.Dp
	Offset unit.Dp
	Color  ColorSource
}

type BorderStyle struct {
	Color ColorSource
	Width *unit.Dp
	// LineStyle selects the stroke pattern. Nil and LineSolid both paint a
	// solid border; the pointer keeps "unset" distinguishable from an
	// explicit choice, as every other field in this package does.
	LineStyle *LineStyle
}

// LineStyle selects a stroke pattern for borders.
type LineStyle uint8

const (
	LineSolid LineStyle = iota
	LineDashed
	LineDotted
)

// DashPattern returns the on/off lengths in Dp for a patterned LineStyle.
// A solid line returns zero, which lets callers skip the segment split.
func (l LineStyle) DashPattern() (on, off unit.Dp) {
	switch l {
	case LineDashed:
		return 4, 3
	case LineDotted:
		return 1, 2
	default:
		return 0, 0
	}
}

type TextStyle struct {
	Color           ColorSource
	FontSize        *unit.Sp
	FontWeight      *int
	Typeface        *font.Typeface
	FontStyle       *font.Style
	LineHeight      *unit.Sp
	LineHeightScale *float32
	MaxLines        *int
	Align           *TextAlign
	Wrap            *giotext.WrapPolicy
	Truncator       *string
	// Underline draws a rule under each text line. It is a style property
	// rather than a widget option because link buttons differ from text
	// buttons by this alone.
	Underline *bool
}

type TextAlign uint8

const (
	TextAlignStart TextAlign = iota
	TextAlignCenter
	TextAlignEnd
)

type TransformStyle struct {
	TranslateX *unit.Dp
	TranslateY *unit.Dp
	ScaleX     *float32
	ScaleY     *float32
	Rotate     *float32
}

// PropertyID identifies a Style property that may participate in Transition.
// IsEmpty reports whether this declaration contains no style values, rules,
// transitions, or part declarations.
func (s Style) IsEmpty() bool {
	return s.box == nil && s.paint == nil && s.text == nil && s.trans == nil &&
		len(s.transitions) == 0 && len(s.rules) == 0 && s.dim == nil && len(s.parts) == 0
}

// New creates an empty style declaration for fluent configuration.
//
// Style values are immutable; each fluent method returns a configured copy.
func New() Style {
	return Style{}
}
