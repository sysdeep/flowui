package tabs

import "gioui.org/unit"

// tabsGeometry is the component-owned default geometry consumed by layout.
type tabsGeometry struct {
	rootGap     unit.Dp
	listPadding unit.Dp
	listRadius  unit.Dp

	tabHeight        unit.Dp
	smallTabHeight   unit.Dp
	largeTabHeight   unit.Dp
	tabMinWidth      unit.Dp
	tabPaddingX      unit.Dp
	smallTabPaddingX unit.Dp
	largeTabPaddingX unit.Dp
	tabGap           unit.Dp

	iconSize unit.Dp
	iconGap  unit.Dp

	closeButtonSize unit.Dp
	closeButtonGap  unit.Dp

	extraContentGap unit.Dp

	indicatorRadius    unit.Dp
	indicatorLineWidth unit.Dp
	indicatorWidth     unit.Dp
	indicatorMinWidth  unit.Dp
	indicatorInset     unit.Dp

	focusRingWidth unit.Dp
	separatorWidth unit.Dp

	panelGap unit.Dp

	scrollButtonSize    unit.Dp
	scrollButtonInset   unit.Dp
	scrollShadowSize    unit.Dp
	scrollChevronSize   unit.Dp
	scrollChevronStroke unit.Dp
}

func tabsGeometryFor(variant Variant) tabsGeometry {
	geometry := tabsGeometry{
		rootGap: 8, listPadding: 4, listRadius: 4,

		tabHeight: 32, smallTabHeight: 24, largeTabHeight: 40,
		tabMinWidth: 80, tabPaddingX: 12,
		smallTabPaddingX: 12, largeTabPaddingX: 16, tabGap: 4,

		iconSize: 16, iconGap: 1,

		closeButtonSize: 16, closeButtonGap: 4,

		extraContentGap: 8,

		indicatorRadius: 3, indicatorLineWidth: 2,
		indicatorWidth: 0, indicatorMinWidth: 24, indicatorInset: 0,

		focusRingWidth: 2, separatorWidth: 1,

		panelGap: 16,

		scrollButtonSize: 16, scrollButtonInset: 4,
		scrollShadowSize: 64, scrollChevronSize: 10, scrollChevronStroke: 1.5,
	}
	if variant == Secondary {
		geometry.smallTabHeight = 32
		geometry.tabHeight = 40
		geometry.largeTabHeight = 48
	}
	return geometry
}

// tabsSizeStyle is the resolved geometry used by neutral layout.
type tabsSizeStyle struct {
	height   unit.Dp
	paddingX unit.Dp
}

func tabsSizeStyleFor(geometry tabsGeometry, size Size) tabsSizeStyle {
	result := tabsSizeStyle{height: geometry.tabHeight, paddingX: geometry.tabPaddingX}
	switch size {
	case Small:
		result.height, result.paddingX = geometry.smallTabHeight, geometry.smallTabPaddingX
	case Large:
		result.height, result.paddingX = geometry.largeTabHeight, geometry.largeTabPaddingX
	}
	return result
}
