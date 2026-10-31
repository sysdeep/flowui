// SPDX-License-Identifier: Unlicense OR MIT

package draw

import (
	"image"
	"image/color"
	"sync"

	"gioui.org/op/paint"
)

type softShadowCacheKey struct {
	w, h       int
	shape      shadowShape
	blur       int
	scale      int
	spread     int
	offX, offY int
	color      color.NRGBA
}

type softShadowStyleKey struct {
	shape      shadowShape
	blur       int
	spread     int
	offX, offY int
	color      color.NRGBA
}

func (k softShadowCacheKey) style() softShadowStyleKey {
	return softShadowStyleKey{
		shape:  k.shape,
		blur:   k.blur,
		spread: k.spread,
		offX:   k.offX,
		offY:   k.offY,
		color:  k.color,
	}
}

type softShadowCacheEntry struct {
	op         paint.ImageOp
	padX, padY int
	size       image.Point
	scale      int
	bytes      int
	lastUsed   uint64
}

type softShadowBuild struct {
	done  chan struct{}
	entry softShadowCacheEntry
}

const (
	softShadowCacheLimit    = 256
	softShadowCacheMaxBytes = 32 << 20
	softShadowVariantLimit  = 4
)

var softShadowCache = struct {
	sync.Mutex
	tick    uint64
	bytes   int
	entries map[softShadowCacheKey]softShadowCacheEntry
	builds  map[softShadowCacheKey]*softShadowBuild
}{
	entries: make(map[softShadowCacheKey]softShadowCacheEntry),
	builds:  make(map[softShadowCacheKey]*softShadowBuild),
}

func softShadowEntry(size image.Point, shape shadowShape, blur, spread, offX, offY int, col color.NRGBA) softShadowCacheEntry {
	if size.X <= 0 || size.Y <= 0 || blur < 0 || col.A == 0 {
		return softShadowCacheEntry{}
	}
	blur = min(blur, maxSoftShadowBlur)
	shape = normalizeShadowShape(image.Rectangle{Max: size}, shape)
	padX := ShadowRasterPadding(blur, spread, offX)
	padY := ShadowRasterPadding(blur, spread, offY)
	rasterSize, ok := shadowRasterSize(size, padX, padY)
	if !ok {
		return softShadowCacheEntry{}
	}
	scale := shadowRasterScale(blur, rasterSize.X, rasterSize.Y)
	key := softShadowCacheKey{
		w:      size.X,
		h:      size.Y,
		shape:  shape,
		blur:   blur,
		scale:  scale,
		spread: spread,
		offX:   offX, offY: offY, color: col,
	}

	softShadowCache.Lock()
	softShadowCache.tick++
	styleKey := key.style()
	if entry, ok := softShadowCache.entries[key]; ok {
		entry.lastUsed = softShadowCache.tick
		softShadowCache.entries[key] = entry
		softShadowCache.Unlock()
		return entry
	}
	if build, ok := softShadowCache.builds[key]; ok {
		softShadowCache.Unlock()
		<-build.done
		return build.entry
	}
	build := &softShadowBuild{done: make(chan struct{})}
	softShadowCache.builds[key] = build
	softShadowCache.Unlock()

	entry := buildSoftShadowEntry(size, shape, blur, spread, offX, offY, col)
	finishBuild := func(entry softShadowCacheEntry) softShadowCacheEntry {
		softShadowCache.Lock()
		delete(softShadowCache.builds, key)
		build.entry = entry
		close(build.done)
		softShadowCache.Unlock()
		return entry
	}
	if entry.op.Size() == (image.Point{}) {
		return finishBuild(entry)
	}
	if entry.bytes > softShadowCacheMaxBytes {
		return finishBuild(entry)
	}

	softShadowCache.Lock()
	entry.lastUsed = softShadowCache.tick
	for len(softShadowCache.entries) >= softShadowCacheLimit || softShadowCache.bytes+entry.bytes > softShadowCacheMaxBytes {
		evictOldestShadowEntryLocked()
	}
	softShadowCache.entries[key] = entry
	softShadowCache.bytes += entry.bytes
	trimShadowVariantsLocked(styleKey)
	delete(softShadowCache.builds, key)
	build.entry = entry
	close(build.done)
	softShadowCache.Unlock()
	return entry
}

func trimShadowVariantsLocked(style softShadowStyleKey) {
	for {
		count := 0
		var oldestKey softShadowCacheKey
		var oldestTick uint64
		first := true
		for key, entry := range softShadowCache.entries {
			if key.style() != style {
				continue
			}
			count++
			if first || entry.lastUsed < oldestTick {
				oldestKey = key
				oldestTick = entry.lastUsed
				first = false
			}
		}
		if count <= softShadowVariantLimit || first {
			return
		}
		removeShadowEntryLocked(oldestKey)
	}
}

func removeShadowEntryLocked(key softShadowCacheKey) {
	entry, ok := softShadowCache.entries[key]
	if !ok {
		return
	}
	softShadowCache.bytes -= entry.bytes
	if softShadowCache.bytes < 0 {
		softShadowCache.bytes = 0
	}
	delete(softShadowCache.entries, key)
}

func evictOldestShadowEntryLocked() {
	var oldestKey softShadowCacheKey
	var oldestTick uint64
	first := true
	for key, entry := range softShadowCache.entries {
		if first || entry.lastUsed < oldestTick {
			oldestKey = key
			oldestTick = entry.lastUsed
			first = false
		}
	}
	if !first {
		removeShadowEntryLocked(oldestKey)
	}
}
