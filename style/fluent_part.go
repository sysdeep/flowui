package style

import "maps"

// Part augments the root style with a named component part.
func (s Style) Part(part Part, override Style) Style {
	if part == PartRoot {
		return merge(s, override)
	}
	result := s
	result.parts = make(map[Part]Style, len(s.parts)+1)
	maps.Copy(result.parts, s.parts)
	result.parts[part] = merge(result.parts[part], override)
	return result
}
