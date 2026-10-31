package datepicker

// datePickerResolvedCacheStats is a snapshot of the transition-free resolved
// part cache. It is intentionally limited to evidence used by tests and
// benchmarks; applications do not need to depend on cache internals.
type datePickerResolvedCacheStats struct {
	Hits                    uint64
	Misses                  uint64
	Evictions               uint64
	ThemeInvalidations      uint64
	GenerationInvalidations uint64
	ScopeInvalidations      uint64
	VariantInvalidations    uint64
	UnsafeBypasses          uint64
}

func (c *datePickerResolvedPartCache) stats() datePickerResolvedCacheStats {
	if c == nil {
		return datePickerResolvedCacheStats{}
	}
	return datePickerResolvedCacheStats{
		Hits:                    c.hits,
		Misses:                  c.misses,
		Evictions:               c.evictions,
		ThemeInvalidations:      c.themeInvalidation,
		GenerationInvalidations: c.generationInvalidation,
		ScopeInvalidations:      c.scopeInvalidation,
		VariantInvalidations:    c.variantInvalidation,
		UnsafeBypasses:          c.unsafeBypass,
	}
}
