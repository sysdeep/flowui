package datepicker

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// dateCellState carries the per-cell interaction state a day cell
// resolves against. It is a parameter packet rather than a replaceable-visual
// seam: the flags are all same-typed booleans, so a named struct keeps the call
// site readable and order-proof.
type dateCellState struct {
	Hovered  bool
	Pressed  bool
	Selected bool
	Today    bool
	Outside  bool
	Disabled bool
	Active   bool
	Range    bool
}

func datePickerCellContextKey(cell dateCellState) uint8 {
	var key uint8
	if cell.Hovered {
		key |= 1 << 0
	}
	if cell.Pressed {
		key |= 1 << 1
	}
	if cell.Selected {
		key |= 1 << 2
	}
	if cell.Today {
		key |= 1 << 3
	}
	if cell.Outside {
		key |= 1 << 4
	}
	if cell.Disabled {
		key |= 1 << 5
	}
	if cell.Active {
		key |= 1 << 6
	}
	if cell.Range {
		key |= 1 << 7
	}
	return key
}

func datePickerStyleStateKey(state style.State) uint16 {
	var key uint16
	flags := [...]bool{
		state.Hovered, state.Pressed, state.Focused, state.FocusVisible,
		state.Disabled, state.Selected, state.Checked, state.Indeterminate,
		state.ReadOnly, state.Invalid, state.Loading, state.Open,
		state.Expanded, state.Dragging, state.DropTarget, state.Warning,
	}
	for index, value := range flags {
		if value {
			key |= 1 << index
		}
	}
	return key
}

func datePickerResolvedCellKey(context dateCellState, state style.State) uint32 {
	return uint32(datePickerCellContextKey(context))<<16 | uint32(datePickerStyleStateKey(state))
}

func datePickerTextResolvedKey(state style.State, base color.NRGBA) uint64 {
	return uint64(datePickerStyleStateKey(state))<<32 |
		uint64(base.R)<<24 | uint64(base.G)<<16 | uint64(base.B)<<8 | uint64(base.A)
}

const datePickerResolvedPartCacheLimit = 256

// datePickerResolvedPartCache keeps transition-free calendar parts after the
// declaration cascade. Calendar frames revisit the same small set of state
// combinations, so retaining the resolved value avoids re-hashing the full
// inherited/custom style stack for every cell. Unsafe declarations and styles
// with transitions deliberately bypass this cache because they need per-use
// ownership or animation state.
type datePickerResolvedPartCache struct {
	theme                  *theme.Theme
	themeGeneration        uint64
	scopeHash              uint64
	variant                Variant
	enabled                bool
	cell                   map[uint32]style.ResolvedStyle
	focus                  map[uint16]style.ResolvedStyle
	rangeStyles            map[uint16]style.ResolvedStyle
	header                 map[uint16]style.ResolvedStyle
	nav                    map[uint16]style.ResolvedStyle
	segments               map[uint64]style.ResolvedStyle
	literals               map[uint64]style.ResolvedStyle
	weekday                [2]style.ResolvedStyle
	weekdayValid           [2]bool
	hits                   uint64
	misses                 uint64
	evictions              uint64
	themeInvalidation      uint64
	generationInvalidation uint64
	scopeInvalidation      uint64
	variantInvalidation    uint64
	unsafeBypass           uint64
}

func (c *datePickerResolvedPartCache) prepare(ctx *core.Context, custom style.Style, variants ...Variant) {
	activeTheme := core.ActiveTheme(ctx)
	generation := core.ThemeGeneration(ctx)
	enabled := datePickerStyleLayersCacheSafe(ctx, custom)
	variant := Primary
	if len(variants) > 0 {
		variant = variants[0]
	}
	scopeHash := uint64(0)
	if enabled {
		scopeHash = datePickerStyleScopeHash(ctx, custom)
	}
	if c.theme == activeTheme && c.themeGeneration == generation && c.scopeHash == scopeHash && c.enabled == enabled && c.variant == variant {
		if !enabled {
			c.unsafeBypass++
		}
		return
	}
	if c.theme != nil {
		switch {
		case c.theme != activeTheme:
			c.themeInvalidation++
		case c.themeGeneration != generation:
			c.generationInvalidation++
		case c.scopeHash != scopeHash:
			c.scopeInvalidation++
		case c.variant != variant:
			c.variantInvalidation++
		}
	}
	if !enabled {
		c.unsafeBypass++
	}
	c.theme = activeTheme
	c.themeGeneration = generation
	c.scopeHash = scopeHash
	c.variant = variant
	c.enabled = enabled
	clear(c.cell)
	clear(c.focus)
	clear(c.rangeStyles)
	clear(c.header)
	clear(c.nav)
	clear(c.segments)
	clear(c.literals)
	c.weekdayValid = [2]bool{}
}

func (c *datePickerResolvedPartCache) cellValue(key uint32) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.cell[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) focusValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.focus[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) rangeValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.rangeStyles[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) headerValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.header[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) navValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.nav[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) segmentValue(key uint64) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.segments[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) literalValue(key uint64) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.literals[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) weekdayValue(key uint8) (style.ResolvedStyle, bool) {
	if !c.enabled || key >= uint8(len(c.weekday)) || !c.weekdayValid[key] {
		if c.enabled {
			c.misses++
		}
		return style.ResolvedStyle{}, false
	}
	c.hits++
	return c.weekday[key], true
}

func (c *datePickerResolvedPartCache) storeCell(key uint32, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.cell) >= datePickerResolvedPartCacheLimit {
		clear(c.cell)
		c.evictions++
	}
	if c.cell == nil {
		c.cell = make(map[uint32]style.ResolvedStyle)
	}
	c.cell[key] = value
}

func (c *datePickerResolvedPartCache) storeFocus(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.focus) >= datePickerResolvedPartCacheLimit {
		clear(c.focus)
		c.evictions++
	}
	if c.focus == nil {
		c.focus = make(map[uint16]style.ResolvedStyle)
	}
	c.focus[key] = value
}

func (c *datePickerResolvedPartCache) storeRange(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.rangeStyles) >= datePickerResolvedPartCacheLimit {
		clear(c.rangeStyles)
		c.evictions++
	}
	if c.rangeStyles == nil {
		c.rangeStyles = make(map[uint16]style.ResolvedStyle)
	}
	c.rangeStyles[key] = value
}

func (c *datePickerResolvedPartCache) storeHeader(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.header) >= datePickerResolvedPartCacheLimit {
		clear(c.header)
		c.evictions++
	}
	if c.header == nil {
		c.header = make(map[uint16]style.ResolvedStyle)
	}
	c.header[key] = value
}

func (c *datePickerResolvedPartCache) storeNav(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.nav) >= datePickerResolvedPartCacheLimit {
		clear(c.nav)
		c.evictions++
	}
	if c.nav == nil {
		c.nav = make(map[uint16]style.ResolvedStyle)
	}
	c.nav[key] = value
}

func (c *datePickerResolvedPartCache) storeSegment(key uint64, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.segments) >= datePickerResolvedPartCacheLimit {
		clear(c.segments)
		c.evictions++
	}
	if c.segments == nil {
		c.segments = make(map[uint64]style.ResolvedStyle)
	}
	c.segments[key] = value
}

func (c *datePickerResolvedPartCache) storeLiteral(key uint64, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.literals) >= datePickerResolvedPartCacheLimit {
		clear(c.literals)
		c.evictions++
	}
	if c.literals == nil {
		c.literals = make(map[uint64]style.ResolvedStyle)
	}
	c.literals[key] = value
}

func (c *datePickerResolvedPartCache) storeWeekday(key uint8, value style.ResolvedStyle) {
	if !c.enabled || key >= uint8(len(c.weekday)) || len(value.Transitions) != 0 {
		return
	}
	c.weekday[key] = value
	c.weekdayValid[key] = true
}

func datePickerStyleLayersCacheSafe(ctx *core.Context, custom style.Style) bool {
	if custom.CacheUnsafe() {
		return false
	}
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	return true
}

func datePickerStyleScopeHash(ctx *core.Context, custom style.Style) uint64 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	hash := offset64
	mix := func(value uint64) {
		for range 8 {
			hash ^= value & 0xff
			hash *= prime64
			value >>= 8
		}
	}
	mix(uint64(len(core.ActiveInheritedStylesReadOnly(ctx))))
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	mix(uint64(len(core.ActiveStylesReadOnly(ctx))))
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	mix(custom.Hash64())
	return hash
}

// datePickerPartDefaultsCache stores the theme-derived defaults for the active
// variant plus the joined cell and range declarations. A calendar frame
// revisits the same small set of boolean combinations for every cell;
// rebuilding the part map for each one otherwise dominates the style resolver's
// allocation path. The instance style is never part of this cache.
type datePickerPartDefaultsCache struct {
	theme           *theme.Theme
	activeVariant   Variant
	defaultsCache   computed.VisualCache[Variant, datePickerDefaults]
	activeDefaults  datePickerDefaults
	activeBundleSet bool
	cell            [256]style.Style
	cellValid       [256]bool
	rangeStyle      [2]style.Style
	rangeValid      [2]bool
}

func (c *datePickerPartDefaultsCache) resetFor(activeTheme *theme.Theme) {
	if c.theme == activeTheme {
		return
	}
	c.theme = activeTheme
	c.activeVariant = Primary
	c.activeBundleSet = false
	for i := range c.cellValid {
		c.cellValid[i] = false
	}
	for i := range c.rangeValid {
		c.rangeValid[i] = false
	}
}

func (c *datePickerPartDefaultsCache) prepareVariant(activeTheme *theme.Theme, variant Variant) Variant {
	c.resetFor(activeTheme)
	key := variant
	if key > Underlined {
		key = Primary
	}
	if c.activeVariant != key {
		c.activeVariant = key
		c.activeBundleSet = false
		for i := range c.cellValid {
			c.cellValid[i] = false
		}
		for i := range c.rangeValid {
			c.rangeValid[i] = false
		}
	}
	return key
}

// defaultsFor resolves the complete defaults bundle needed by one layout frame.
// The cache key contains only inputs that change the defaults; the instance
// style is resolved as the last cascade layer at each part.
func (c *datePickerPartDefaultsCache) defaultsFor(activeTheme *theme.Theme, variant Variant) datePickerDefaults {
	key := c.prepareVariant(activeTheme, variant)
	defaults := c.defaultsCache.Resolve(activeTheme, key, func() datePickerDefaults {
		return buildDatePickerDefaults(activeTheme, key)
	})
	c.activeDefaults = defaults
	c.activeBundleSet = true
	return defaults
}

func (c *datePickerPartDefaultsCache) currentDefaults(activeTheme *theme.Theme) datePickerDefaults {
	c.resetFor(activeTheme)
	if c.activeBundleSet {
		return c.activeDefaults
	}
	return c.defaultsFor(activeTheme, c.activeVariant)
}

func (c *datePickerPartDefaultsCache) motion(activeTheme *theme.Theme) datePickerMotionDefaults {
	return c.currentDefaults(activeTheme).motion
}

func (c *datePickerPartDefaultsCache) cellDefaults(activeTheme *theme.Theme, base style.Style, context dateCellState) style.Style {
	c.resetFor(activeTheme)
	key := datePickerCellContextKey(context)
	if c.cellValid[key] {
		return c.cell[key]
	}
	declaration := dateCellDeclaration(activeTheme, context)
	value := style.JoinPart(base, style.Style{}.
		Part(PartCalendarCell, declaration), PartCalendarCell)
	c.cell[key] = value
	c.cellValid[key] = true
	return value
}

func (c *datePickerPartDefaultsCache) rangeDefaults(activeTheme *theme.Theme, base style.Style, outside bool) style.Style {
	c.resetFor(activeTheme)
	key := uint8(0)
	if outside {
		key = 1
	}
	if c.rangeValid[key] {
		return c.rangeStyle[key]
	}
	declaration := dateRangeDeclaration(activeTheme, outside)
	value := style.JoinPart(base, style.Style{}.
		Part(PartCalendarRange, declaration), PartCalendarRange)
	c.rangeStyle[key] = value
	c.rangeValid[key] = true
	return value
}
