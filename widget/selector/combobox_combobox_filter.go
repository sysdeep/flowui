package selector

import "strings"

func (c ComboboxWidget) selectedLabel() (string, bool) {
	for _, item := range c.items {
		if item.Key == c.selectedKey {
			return item.Label, true
		}
	}
	return "", false
}

func comboboxComboBoxVisibleItems(items []ComboboxItem, query, selectedLabel string) []int {
	query = strings.TrimSpace(query)
	if query == "" || strings.EqualFold(query, selectedLabel) {
		visible := make([]int, len(items))
		for i := range items {
			visible[i] = i
		}
		return visible
	}
	query = strings.ToLower(query)
	visible := make([]int, 0, len(items))
	for i, item := range items {
		if comboboxComboBoxMatches(item, query) {
			visible = append(visible, i)
		}
	}
	return visible
}

func comboboxComboBoxMatches(item ComboboxItem, query string) bool {
	return strings.Contains(strings.ToLower(item.Label), query) ||
		strings.Contains(strings.ToLower(item.Description), query) ||
		strings.Contains(strings.ToLower(item.Key), query)
}
